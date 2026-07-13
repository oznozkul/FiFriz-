package integrations

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"strings"

	"fifriz/internal/applog"
	"fifriz/internal/audit"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

type IMAPProvider struct{}

func (IMAPProvider) Name() string { return ProviderIMAP }

func (IMAPProvider) TestConnection(_ context.Context, account IntegrationAccount, syncCtx SyncContext) error {
	settings, err := ParseIMAPSettings(account.Settings)
	if err != nil {
		return err
	}
	password, err := syncCtx.Credentials.GetPassword(account.EncryptedCredentials)
	if err != nil {
		return fmt.Errorf("IMAP password is not configured")
	}
	c, err := connectIMAP(settings, password)
	if err != nil {
		return userFriendlyIMAPError(err)
	}
	defer c.Logout()
	return nil
}

func (IMAPProvider) Sync(ctx context.Context, account IntegrationAccount, syncCtx SyncContext) (*SyncProviderResult, error) {
	settings, err := ParseIMAPSettings(account.Settings)
	if err != nil {
		return nil, err
	}
	password, err := syncCtx.Credentials.GetPassword(account.EncryptedCredentials)
	if err != nil {
		return nil, fmt.Errorf("IMAP password is not configured")
	}

	c, err := connectIMAP(settings, password)
	if err != nil {
		return nil, userFriendlyIMAPError(err)
	}
	defer func() {
		_ = c.Logout()
	}()

	if _, err := c.Select(settings.Mailbox, false); err != nil {
		return nil, userFriendlyIMAPError(fmt.Errorf("mailbox select failed: %w", err))
	}

	criteria := imap.NewSearchCriteria()
	criteria.Since = settings.SyncFromDate
	uids, err := c.UidSearch(criteria)
	if err != nil {
		return nil, userFriendlyIMAPError(fmt.Errorf("mail search failed: %w", err))
	}

	maxEmails := EffectiveMaxEmailsPerSync(settings.MaxEmailsPerSync)
	sort.Slice(uids, func(i, j int) bool { return uids[i] < uids[j] })
	if len(uids) > maxEmails {
		uids = uids[:maxEmails]
	}

	result := &SyncProviderResult{
		Metadata: map[string]any{
			"mailbox":    settings.Mailbox,
			"emailCount": len(uids),
		},
	}
	result.TotalItems = len(uids)
	if len(uids) == 0 {
		return result, nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchUid, imap.FetchEnvelope, section.FetchItem()}

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(seqset, items, messages)
	}()

	for msg := range messages {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		if err := processIMAPMessage(ctx, account, syncCtx, settings, msg, section, result); err != nil {
			applog.Warn("imap message processing failed: integration_id=%s uid=%d error=%v", account.ID, msg.Uid, err)
			result.FailedItems++
		}
	}

	if err := <-done; err != nil {
		return result, userFriendlyIMAPError(fmt.Errorf("mail fetch failed: %w", err))
	}

	return result, nil
}

func connectIMAP(settings IMAPSettings, password string) (*client.Client, error) {
	host := NormalizeIMAPHost(settings.Host)
	if host == "" {
		return nil, fmt.Errorf("IMAP host is required")
	}
	address := fmt.Sprintf("%s:%d", host, settings.Port)
	tlsConfig := imapTLSConfig(settings, host)
	mode := settings.ConnectionMode()

	var c *client.Client
	var err error

	switch mode {
	case "none":
		c, err = client.Dial(address)
	case "starttls":
		c, err = client.Dial(address)
		if err == nil {
			if startErr := c.StartTLS(tlsConfig); startErr != nil {
				_ = c.Logout()
				err = startErr
				c = nil
			}
		}
	default: // ssl
		c, err = client.DialTLS(address, tlsConfig)
	}
	if err != nil {
		return nil, err
	}
	if err := c.Login(settings.Username, password); err != nil {
		_ = c.Logout()
		return nil, err
	}
	return c, nil
}

func imapTLSConfig(settings IMAPSettings, host string) *tls.Config {
	serverName := resolveIMAPTLSServerName(host, settings.Username, settings.TLSServerName)
	return &tls.Config{
		ServerName:         serverName,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: settings.InsecureSkipTLS,
	}
}

func resolveIMAPTLSServerName(host, username, configured string) string {
	if configured = NormalizeIMAPHost(configured); configured != "" {
		return configured
	}
	if net.ParseIP(host) == nil {
		return host
	}
	if domain := emailDomain(username); domain != "" {
		return "mail." + domain
	}
	return host
}

func emailDomain(username string) string {
	at := strings.LastIndex(username, "@")
	if at < 0 || at == len(username)-1 {
		return ""
	}
	return strings.TrimSpace(username[at+1:])
}

type attachmentCandidate struct {
	filename string
	content  []byte
}

func processIMAPMessage(
	ctx context.Context,
	account IntegrationAccount,
	syncCtx SyncContext,
	settings IMAPSettings,
	msg *imap.Message,
	section *imap.BodySectionName,
	result *SyncProviderResult,
) error {
	if msg == nil {
		return fmt.Errorf("empty message")
	}

	rawReader := msg.GetBody(section)
	if rawReader == nil {
		return fmt.Errorf("message body is empty")
	}
	rawData, err := io.ReadAll(rawReader)
	if err != nil {
		return err
	}

	mailReader, err := mail.CreateReader(bytes.NewReader(rawData))
	if err != nil {
		return err
	}

	header := mailReader.Header
	messageID := strings.TrimSpace(header.Get("Message-Id"))
	if messageID == "" {
		messageID = fmt.Sprintf("uid:%d", msg.Uid)
	}

	subject, _ := header.Subject()
	subject = strings.TrimSpace(subject)
	if subject == "" {
		subject = decodeMIMEHeaderValue(header.Get("Subject"))
	}
	if subject == "" {
		subject = "email"
	}

	baseMetadata := map[string]any{
		"from":      header.Get("From"),
		"to":        header.Get("To"),
		"subject":   subject,
		"date":      header.Get("Date"),
		"messageId": messageID,
		"mailbox":   settings.Mailbox,
	}

	var bodyBuilder strings.Builder
	attachments := make([]attachmentCandidate, 0)

	for {
		part, partErr := mailReader.NextPart()
		if partErr == io.EOF {
			break
		}
		if partErr != nil {
			return partErr
		}

		switch typed := part.Header.(type) {
		case *mail.AttachmentHeader:
			filename, _ := typed.Filename()
			filename = sanitizeIMAPFilename(filename)
			if filename == "" {
				continue
			}
			content, readErr := io.ReadAll(part.Body)
			if readErr != nil || len(content) == 0 {
				continue
			}
			attachments = append(attachments, attachmentCandidate{filename: filename, content: content})
		case *mail.InlineHeader:
			contentType := strings.ToLower(part.Header.Get("Content-Type"))
			if strings.HasPrefix(contentType, "text/plain") {
				content, readErr := io.ReadAll(part.Body)
				if readErr == nil && len(strings.TrimSpace(string(content))) > 0 {
					if bodyBuilder.Len() > 0 {
						bodyBuilder.WriteString("\n\n")
					}
					bodyBuilder.Write(content)
				}
			}
		}
	}

	if settings.ImportEmailBody {
		bodyText := strings.TrimSpace(bodyBuilder.String())
		if bodyText != "" {
			externalID := messageID + ":body"
			outcome := syncCtx.Importer.ImportFile(ctx, ImportFileInput{
				Account:           &account,
				ExternalID:        externalID,
				ExternalType:      ExternalTypeEmailBody,
				OriginalFileName:  safeEmailBodyFilename(subject),
				FileExtension:     "txt",
				MimeType:          "text/plain",
				Content:           []byte(bodyText),
				SourceMetadata:    baseMetadata,
				AuditAction:       audit.ActionIntegrationIMAPEmailImported,
				TriggeredByUserID: syncCtx.TriggeredByUserID,
			})
			applyImportOutcome(result, outcome)
		}
	}

	if settings.ImportAttachments {
		for _, attachment := range attachments {
			if !settings.IsExtensionAllowed(attachment.filename) {
				result.SkippedItems++
				continue
			}

			checksum := ChecksumHex(attachment.content)
			externalID := fmt.Sprintf("%s:attachment:%s:%s", messageID, sanitizeIMAPFilename(attachment.filename), checksum[:16])
			attachmentMetadata := copyIMAPMetadata(baseMetadata)
			attachmentMetadata["attachmentFileName"] = attachment.filename

			outcome := syncCtx.Importer.ImportFile(ctx, ImportFileInput{
				Account:           &account,
				ExternalID:        externalID,
				ExternalType:      ExternalTypeEmailAttachment,
				OriginalFileName:  sanitizeIMAPFilename(attachment.filename),
				FileExtension:     imapExtensionFromFilename(attachment.filename),
				MimeType:          mimeTypeForIMAPExtension(imapExtensionFromFilename(attachment.filename)),
				Content:           attachment.content,
				SourceMetadata:    attachmentMetadata,
				Checksum:          checksum,
				AuditAction:       audit.ActionIntegrationIMAPAttachmentImported,
				TriggeredByUserID: syncCtx.TriggeredByUserID,
			})
			applyImportOutcome(result, outcome)
		}
	}

	return nil
}

func applyImportOutcome(result *SyncProviderResult, outcome ImportOutcome) {
	switch {
	case outcome.Imported:
		result.ImportedItems++
	case outcome.Skipped:
		result.SkippedItems++
	case outcome.Failed:
		result.FailedItems++
	}
}

func imapStringValue(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func imapExtensionFromFilename(filename string) string {
	filename = strings.TrimSpace(filename)
	if idx := strings.LastIndex(filename, "."); idx >= 0 && idx < len(filename)-1 {
		return strings.ToLower(filename[idx+1:])
	}
	return ""
}

func mimeTypeForIMAPExtension(extension string) string {
	switch strings.ToLower(extension) {
	case "pdf":
		return "application/pdf"
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

func copyIMAPMetadata(source map[string]any) map[string]any {
	target := make(map[string]any, len(source))
	for key, value := range source {
		target[key] = value
	}
	return target
}

func userFriendlyIMAPError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "authentication") || strings.Contains(message, "login") || strings.Contains(message, "invalid credentials"):
		return fmt.Errorf("IMAP giris basarisiz. Kullanici adi ve sifreyi kontrol edin")
	case strings.Contains(message, "connection refused"),
		strings.Contains(message, "actively refused"),
		strings.Contains(message, "refused"),
		strings.Contains(message, "no such host"),
		strings.Contains(message, "dial tcp"):
		return fmt.Errorf("IMAP sunucusuna baglanilamadi. Host dogru mu? SSL icin port 993, STARTTLS icin port 143 deneyin")
	case strings.Contains(message, "timeout"), errors.Is(err, net.ErrClosed):
		return fmt.Errorf("IMAP baglantisi zaman asimina ugradi")
	case strings.Contains(message, "x509"), strings.Contains(message, "certificate"):
		return fmt.Errorf("IMAP sertifika dogrulamasi basarisiz. IP yerine mail sunucu adini kullanin (or. mail.alanadi.com) veya TLS sunucu adi alanina sertifikadaki adi girin")
	case strings.Contains(message, "tls"), strings.Contains(message, "starttls"), strings.Contains(message, "handshake"):
		return fmt.Errorf("IMAP TLS baglantisi kurulamadi. Baglanti turunu degistirin (993 SSL <-> 143 STARTTLS) veya TLS sunucu adini kontrol edin")
	default:
		return fmt.Errorf("IMAP senkronizasyonu basarisiz: %s", sanitizeIMAPErrorMessage(err.Error()))
	}
}

func sanitizeIMAPErrorMessage(message string) string {
	lower := strings.ToLower(message)
	if strings.Contains(lower, "password") || strings.Contains(lower, "auth") {
		return "connection error"
	}
	if len(message) > 200 {
		return message[:200]
	}
	return message
}
