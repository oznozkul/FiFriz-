package integrations

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"fifriz/internal/applog"
	"fifriz/internal/audit"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

func (c GoogleOAuthConfig) IsConfigured() bool {
	return strings.TrimSpace(c.ClientID) != "" &&
		strings.TrimSpace(c.ClientSecret) != "" &&
		strings.TrimSpace(c.RedirectURL) != ""
}

func (c GoogleOAuthConfig) oauthConfig() *oauth2.Config {
	scopes := c.Scopes
	if len(scopes) == 0 {
		scopes = []string{gmail.GmailReadonlyScope}
	}
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}
}

type GmailProvider struct {
	google      GoogleOAuthConfig
	credentials *CredentialService
	repository  *Repository
}

func NewGmailProvider(google GoogleOAuthConfig, credentials *CredentialService, repository *Repository) GmailProvider {
	return GmailProvider{
		google:      google,
		credentials: credentials,
		repository:  repository,
	}
}

func (GmailProvider) Name() string { return ProviderGmail }

func (p GmailProvider) TestConnection(ctx context.Context, account IntegrationAccount, syncCtx SyncContext) error {
	accountCopy := account
	_, err := p.gmailService(ctx, &accountCopy, syncCtx)
	return userFriendlyGmailError(err)
}

func (p GmailProvider) Sync(ctx context.Context, account IntegrationAccount, syncCtx SyncContext) (*SyncProviderResult, error) {
	settings, err := ParseGmailSettings(account.Settings)
	if err != nil {
		return nil, err
	}

	service, err := p.gmailService(ctx, &account, syncCtx)
	if err != nil {
		return nil, userFriendlyGmailError(err)
	}

	query := settings.BuildQuery()
	maxEmails := EffectiveMaxEmailsPerSync(settings.MaxEmailsPerSync)

	messageRefs := make([]*gmail.Message, 0, maxEmails)
	pageToken := ""
	for len(messageRefs) < maxEmails {
		pageSize := int64(MailSyncPageSize)
		remaining := int64(maxEmails - len(messageRefs))
		if remaining < pageSize {
			pageSize = remaining
		}

		listCall := service.Users.Messages.List("me").MaxResults(pageSize)
		if query != "" {
			listCall = listCall.Q(query)
		}
		if pageToken != "" {
			listCall = listCall.PageToken(pageToken)
		}

		listResponse, err := listCall.Do()
		if err != nil {
			return nil, userFriendlyGmailError(err)
		}
		if len(listResponse.Messages) == 0 {
			break
		}
		messageRefs = append(messageRefs, listResponse.Messages...)
		if listResponse.NextPageToken == "" {
			break
		}
		pageToken = listResponse.NextPageToken
	}

	result := &SyncProviderResult{
		Metadata: map[string]any{
			"query":       query,
			"emailCount":  len(messageRefs),
			"maxEmails":   maxEmails,
		},
	}
	if len(messageRefs) == 0 {
		return result, nil
	}

	result.TotalItems = len(messageRefs)
	for _, messageRef := range messageRefs {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		if messageRef == nil || strings.TrimSpace(messageRef.Id) == "" {
			continue
		}
		if err := p.processGmailMessage(ctx, service, account, syncCtx, settings, messageRef.Id, result); err != nil {
			applog.Warn("gmail message processing failed: integration_id=%s message_id=%s error=%v", account.ID, messageRef.Id, err)
			result.FailedItems++
		}
	}

	return result, nil
}

func (p GmailProvider) gmailService(ctx context.Context, account *IntegrationAccount, syncCtx SyncContext) (*gmail.Service, error) {
	if !p.google.IsConfigured() {
		return nil, errors.New("Gmail OAuth is not configured")
	}
	if syncCtx.Credentials == nil {
		return nil, errors.New("credential service is not configured")
	}

	tokens, err := syncCtx.Credentials.GetOAuthTokens(account.EncryptedCredentials)
	if err != nil {
		return nil, err
	}

	oauthCfg := p.google.oauthConfig()
	oauthToken := &oauth2.Token{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		Expiry:       tokens.Expiry,
	}

	tokenSource := oauthCfg.TokenSource(ctx, oauthToken)
	freshToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("gmail token refresh failed: %w", err)
	}

	if freshToken.AccessToken != tokens.AccessToken || !freshToken.Expiry.Equal(tokens.Expiry) {
		if err := PersistGmailTokens(ctx, p.repository, syncCtx.Credentials, account, freshToken); err != nil {
			applog.Warn("gmail token persist failed: integration_id=%s error=%v", account.ID, err)
		}
	}

	client := oauth2.NewClient(ctx, oauth2.ReuseTokenSource(freshToken, tokenSource))
	return gmail.NewService(ctx, option.WithHTTPClient(client))
}

func (p GmailProvider) processGmailMessage(
	ctx context.Context,
	service *gmail.Service,
	account IntegrationAccount,
	syncCtx SyncContext,
	settings GmailSettings,
	messageID string,
	result *SyncProviderResult,
) error {
	message, err := service.Users.Messages.Get("me", messageID).Format("full").Do()
	if err != nil {
		return err
	}

	headers := gmailHeaders(message)
	subject := decodeMIMEHeaderValue(headers["subject"])
	if subject == "" {
		subject = "email"
	}

	baseMetadata := map[string]any{
		"from":           headers["from"],
		"to":             headers["to"],
		"subject":        subject,
		"date":           headers["date"],
		"messageId":      headers["message-id"],
		"gmailMessageId": messageID,
	}

	if settings.ImportEmailBody {
		bodyText := extractGmailBodyText(message.Payload)
		if strings.TrimSpace(bodyText) != "" {
			externalID := messageID + ":body"
			outcome := syncCtx.Importer.ImportFile(ctx, ImportFileInput{
				Account:           &account,
				ExternalID:        externalID,
				ExternalType:      ExternalTypeEmailBody,
				OriginalFileName:  safeGmailBodyFilename(subject),
				FileExtension:     "txt",
				MimeType:          "text/plain",
				Content:           []byte(bodyText),
				SourceMetadata:    baseMetadata,
				AuditAction:       audit.ActionIntegrationGmailEmailImported,
				TriggeredByUserID: syncCtx.TriggeredByUserID,
			})
			applyGmailImportOutcome(result, outcome)
		}
	}

	if settings.ImportAttachments {
		attachments := collectGmailAttachments(message.Payload)
		for _, attachment := range attachments {
			if !settings.IsExtensionAllowed(attachment.filename) {
				result.SkippedItems++
				continue
			}
			content, err := downloadGmailAttachment(ctx, service, messageID, attachment)
			if err != nil || len(content) == 0 {
				result.FailedItems++
				continue
			}

			checksum := ChecksumHex(content)
			externalID := fmt.Sprintf("%s:attachment:%s:%s", messageID, attachment.attachmentID, checksum[:16])
			attachmentMetadata := copyGmailMetadata(baseMetadata)
			attachmentMetadata["attachmentFileName"] = attachment.filename

			outcome := syncCtx.Importer.ImportFile(ctx, ImportFileInput{
				Account:           &account,
				ExternalID:        externalID,
				ExternalType:      ExternalTypeEmailAttachment,
				OriginalFileName:  sanitizeIMAPFilename(attachment.filename),
				FileExtension:     imapExtensionFromFilename(attachment.filename),
				MimeType:          mimeTypeForIMAPExtension(imapExtensionFromFilename(attachment.filename)),
				Content:           content,
				SourceMetadata:    attachmentMetadata,
				Checksum:          checksum,
				AuditAction:       audit.ActionIntegrationGmailAttachmentImported,
				TriggeredByUserID: syncCtx.TriggeredByUserID,
			})
			applyGmailImportOutcome(result, outcome)
		}
	}

	return nil
}

type gmailAttachmentPart struct {
	filename     string
	attachmentID string
	mimeType     string
	size         int64
}

func collectGmailAttachments(part *gmail.MessagePart) []gmailAttachmentPart {
	if part == nil {
		return nil
	}
	results := make([]gmailAttachmentPart, 0)
	var walk func(*gmail.MessagePart)
	walk = func(current *gmail.MessagePart) {
		if current == nil {
			return
		}
		if current.Filename != "" && current.Body != nil && current.Body.AttachmentId != "" {
			results = append(results, gmailAttachmentPart{
				filename:     current.Filename,
				attachmentID: current.Body.AttachmentId,
				mimeType:     current.MimeType,
				size:         current.Body.Size,
			})
		}
		for _, child := range current.Parts {
			walk(child)
		}
	}
	walk(part)
	return results
}

func downloadGmailAttachment(ctx context.Context, service *gmail.Service, messageID string, attachment gmailAttachmentPart) ([]byte, error) {
	if service == nil {
		return nil, errors.New("gmail service is not configured")
	}
	response, err := service.Users.Messages.Attachments.Get("me", messageID, attachment.attachmentID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return decodeGmailBodyData(response.Data)
}

func extractGmailBodyText(part *gmail.MessagePart) string {
	if part == nil {
		return ""
	}
	plain, html := "", ""
	var walk func(*gmail.MessagePart)
	walk = func(current *gmail.MessagePart) {
		if current == nil {
			return
		}
		mimeType := strings.ToLower(current.MimeType)
		if current.Body != nil && current.Body.Data != "" {
			decoded, err := decodeGmailBodyData(current.Body.Data)
			if err == nil {
				text := strings.TrimSpace(string(decoded))
				if strings.HasPrefix(mimeType, "text/plain") && text != "" {
					plain = text
				} else if strings.HasPrefix(mimeType, "text/html") && text != "" {
					html = text
				}
			}
		}
		for _, child := range current.Parts {
			walk(child)
		}
	}
	walk(part)
	if plain != "" {
		return plain
	}
	return htmlToPlainText(html)
}

func decodeGmailBodyData(data string) ([]byte, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return nil, nil
	}
	return base64.RawURLEncoding.DecodeString(data)
}

func gmailHeaders(message *gmail.Message) map[string]string {
	headers := make(map[string]string)
	if message == nil || message.Payload == nil {
		return headers
	}
	for _, header := range message.Payload.Headers {
		key := strings.ToLower(strings.TrimSpace(header.Name))
		if key == "" {
			continue
		}
		headers[key] = strings.TrimSpace(header.Value)
	}
	return headers
}

func safeGmailBodyFilename(subject string) string {
	base := sanitizeIMAPFilename(subject)
	if base == "" {
		base = "email"
	}
	if !strings.HasSuffix(strings.ToLower(base), ".eml.txt") {
		if strings.HasSuffix(strings.ToLower(base), ".txt") {
			base = strings.TrimSuffix(base, ".txt") + ".eml.txt"
		} else {
			base += ".eml.txt"
		}
	}
	return base
}

func copyGmailMetadata(source map[string]any) map[string]any {
	target := make(map[string]any, len(source))
	for key, value := range source {
		target[key] = value
	}
	return target
}

func applyGmailImportOutcome(result *SyncProviderResult, outcome ImportOutcome) {
	switch {
	case outcome.Imported:
		result.ImportedItems++
	case outcome.Skipped:
		result.SkippedItems++
	case outcome.Failed:
		result.FailedItems++
	}
}

func userFriendlyGmailError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "token refresh failed"), strings.Contains(message, "invalid_grant"), strings.Contains(message, "unauthorized"):
		return fmt.Errorf("Gmail oturumu sona erdi. Hesabi yeniden baglayin")
	case strings.Contains(message, "oauth is not configured"):
		return fmt.Errorf("Gmail OAuth yapilandirmasi eksik")
	default:
		return fmt.Errorf("Gmail senkronizasyonu basarisiz: %s", sanitizeGmailErrorMessage(err.Error()))
	}
}

func sanitizeGmailErrorMessage(message string) string {
	lower := strings.ToLower(message)
	if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "auth") {
		return "baglanti hatasi"
	}
	if len(message) > 200 {
		return message[:200]
	}
	return message
}

// BeginGmailOAuth builds the Google consent URL.
func BeginGmailOAuth(ctx context.Context, google GoogleOAuthConfig, stateSecret string, tenantID, userID uuid.UUID) (string, error) {
	if !google.IsConfigured() {
		return "", errors.New("Gmail OAuth is not configured")
	}
	payload, err := newOAuthState(tenantID, userID)
	if err != nil {
		return "", err
	}
	state, err := signOAuthState(stateSecret, payload)
	if err != nil {
		return "", err
	}
	authURL := google.oauthConfig().AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	_ = ctx
	return authURL, nil
}

// CompleteGmailOAuth exchanges the auth code and upserts the integration account.
func CompleteGmailOAuth(
	ctx context.Context,
	repository *Repository,
	credentials *CredentialService,
	google GoogleOAuthConfig,
	stateSecret string,
	code string,
	state string,
) (*IntegrationAccount, error) {
	payload, err := verifyOAuthState(stateSecret, state)
	if err != nil {
		return nil, err
	}
	if !google.IsConfigured() {
		return nil, errors.New("Gmail OAuth is not configured")
	}

	token, err := google.oauthConfig().Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("gmail token exchange failed: %w", err)
	}

	client := google.oauthConfig().Client(ctx, token)
	profile, err := fetchGmailProfile(client)
	if err != nil {
		return nil, err
	}
	email := strings.TrimSpace(profile.EmailAddress)
	if email == "" {
		return nil, errors.New("gmail email address could not be resolved")
	}

	encrypted, err := credentials.StoreOAuthTokens(OAuthTokens{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	})
	if err != nil {
		return nil, err
	}

	settings, err := encodeJSONMap(gmailSettingsToMap(defaultGmailSettings()))
	if err != nil {
		return nil, err
	}

	displayName := fmt.Sprintf("Gmail - %s", email)
	account, err := repository.FindByProviderEmail(ctx, payload.TenantID, payload.UserID, ProviderGmail, email)
	if err != nil {
		return nil, err
	}

	if account == nil {
		account = &IntegrationAccount{
			TenantID:             payload.TenantID,
			UserID:               payload.UserID,
			Provider:             ProviderGmail,
			DisplayName:          displayName,
			EmailAddress:         &email,
			Status:               StatusConnected,
			Settings:             settings,
			EncryptedCredentials: encrypted,
			IsActive:             true,
		}
		if err := repository.Create(ctx, account); err != nil {
			return nil, err
		}
		return account, nil
	}

	account.DisplayName = displayName
	account.EmailAddress = &email
	account.Status = StatusConnected
	account.EncryptedCredentials = encrypted
	account.LastError = nil
	account.IsActive = true
	if len(account.Settings) == 0 {
		account.Settings = settings
	}
	if err := repository.Update(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func gmailSettingsToMap(settings GmailSettings) map[string]any {
	return map[string]any{
		"query":             settings.Query,
		"syncFromDate":      settings.SyncFromDate.Format("2006-01-02"),
		"importAttachments": settings.ImportAttachments,
		"importEmailBody":   settings.ImportEmailBody,
		"allowedExtensions": settings.AllowedExtensions,
		"maxEmailsPerSync":  settings.MaxEmailsPerSync,
	}
}

type gmailProfileResponse struct {
	EmailAddress string `json:"emailAddress"`
}

func fetchGmailProfile(client *http.Client) (*gmailProfileResponse, error) {
	if client == nil {
		return nil, errors.New("http client is not configured")
	}
	request, err := http.NewRequest(http.MethodGet, "https://gmail.googleapis.com/gmail/v1/users/me/profile", nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("gmail profile request failed")
	}
	var profile gmailProfileResponse
	if err := json.NewDecoder(response.Body).Decode(&profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// PersistGmailTokens saves refreshed tokens for an account.
func PersistGmailTokens(ctx context.Context, repository *Repository, credentials *CredentialService, account *IntegrationAccount, token *oauth2.Token) error {
	if account == nil || token == nil {
		return errors.New("invalid gmail token persistence input")
	}
	encrypted, err := credentials.StoreOAuthTokens(OAuthTokens{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	})
	if err != nil {
		return err
	}
	account.EncryptedCredentials = encrypted
	return repository.Update(ctx, account)
}
