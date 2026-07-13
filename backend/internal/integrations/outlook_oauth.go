package integrations

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"fifriz/internal/applog"
	"fifriz/internal/audit"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

type MicrosoftOAuthConfig struct {
	ClientID     string
	ClientSecret string
	Tenant       string
	RedirectURL  string
	Scopes       []string
}

func (c MicrosoftOAuthConfig) IsConfigured() bool {
	return strings.TrimSpace(c.ClientID) != "" &&
		strings.TrimSpace(c.ClientSecret) != "" &&
		strings.TrimSpace(c.RedirectURL) != ""
}

func (c MicrosoftOAuthConfig) tenantID() string {
	tenant := strings.TrimSpace(c.Tenant)
	if tenant == "" {
		return "common"
	}
	return tenant
}

func (c MicrosoftOAuthConfig) oauthConfig() *oauth2.Config {
	scopes := c.Scopes
	if len(scopes) == 0 {
		scopes = []string{"offline_access", "User.Read", "Mail.Read"}
	}
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURL,
		Scopes:       scopes,
		Endpoint:     microsoft.AzureADEndpoint(c.tenantID()),
	}
}

type OutlookProvider struct {
	microsoft   MicrosoftOAuthConfig
	credentials *CredentialService
	repository  *Repository
}

func NewOutlookProvider(microsoft MicrosoftOAuthConfig, credentials *CredentialService, repository *Repository) OutlookProvider {
	return OutlookProvider{
		microsoft:   microsoft,
		credentials: credentials,
		repository:  repository,
	}
}

func (OutlookProvider) Name() string { return ProviderOutlook }

func (p OutlookProvider) TestConnection(ctx context.Context, account IntegrationAccount, syncCtx SyncContext) error {
	accountCopy := account
	_, err := p.graphClient(ctx, &accountCopy, syncCtx)
	return userFriendlyOutlookError(err)
}

func (p OutlookProvider) Sync(ctx context.Context, account IntegrationAccount, syncCtx SyncContext) (*SyncProviderResult, error) {
	settings, err := ParseOutlookSettings(account.Settings)
	if err != nil {
		return nil, err
	}

	client, err := p.graphClient(ctx, &account, syncCtx)
	if err != nil {
		return nil, userFriendlyOutlookError(err)
	}

	messages, err := p.listMessages(ctx, client, settings)
	if err != nil {
		return nil, userFriendlyOutlookError(err)
	}

	result := &SyncProviderResult{
		Metadata: map[string]any{
			"folder":     settings.Folder,
			"emailCount": len(messages),
		},
	}
	if len(messages) == 0 {
		return result, nil
	}

	result.TotalItems = len(messages)
	for _, message := range messages {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		if strings.TrimSpace(message.ID) == "" {
			continue
		}
		if err := p.processOutlookMessage(ctx, client, account, syncCtx, settings, message.ID, result); err != nil {
			applog.Warn("outlook message processing failed: integration_id=%s message_id=%s error=%v", account.ID, message.ID, err)
			result.FailedItems++
		}
	}

	return result, nil
}

func (p OutlookProvider) graphClient(ctx context.Context, account *IntegrationAccount, syncCtx SyncContext) (*http.Client, error) {
	if !p.microsoft.IsConfigured() {
		return nil, errors.New("Outlook OAuth is not configured")
	}
	if syncCtx.Credentials == nil {
		return nil, errors.New("credential service is not configured")
	}

	tokens, err := syncCtx.Credentials.GetOAuthTokens(account.EncryptedCredentials)
	if err != nil {
		return nil, err
	}

	oauthCfg := p.microsoft.oauthConfig()
	oauthToken := &oauth2.Token{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		Expiry:       tokens.Expiry,
	}

	tokenSource := oauthCfg.TokenSource(ctx, oauthToken)
	freshToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("outlook token refresh failed: %w", err)
	}

	if freshToken.AccessToken != tokens.AccessToken || !freshToken.Expiry.Equal(tokens.Expiry) {
		if err := PersistOutlookTokens(ctx, p.repository, syncCtx.Credentials, account, freshToken); err != nil {
			applog.Warn("outlook token persist failed: integration_id=%s error=%v", account.ID, err)
		}
	}

	return oauth2.NewClient(ctx, oauth2.ReuseTokenSource(freshToken, tokenSource)), nil
}

type outlookMessageRef struct {
	ID string `json:"id"`
}

type outlookMessageListResponse struct {
	Value         []outlookMessageRef `json:"value"`
	NextLink      string              `json:"@odata.nextLink"`
	ODataNextLink string              `json:"odata.nextLink"`
}

type outlookEmailAddress struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type outlookRecipient struct {
	EmailAddress outlookEmailAddress `json:"emailAddress"`
}

type outlookBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type outlookMessage struct {
	ID                   string             `json:"id"`
	Subject              string             `json:"subject"`
	ReceivedDateTime     string             `json:"receivedDateTime"`
	InternetMessageID    string             `json:"internetMessageId"`
	HasAttachments       bool               `json:"hasAttachments"`
	From                 outlookRecipient   `json:"from"`
	ToRecipients         []outlookRecipient `json:"toRecipients"`
	Body                 outlookBody        `json:"body"`
	Attachments          []outlookAttachment `json:"attachments"`
}

type outlookAttachment struct {
	ODataType    string `json:"@odata.type"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	ContentType  string `json:"contentType"`
	ContentBytes string `json:"contentBytes"`
	Size         int    `json:"size"`
}

func (p OutlookProvider) listMessages(ctx context.Context, client *http.Client, settings OutlookSettings) ([]outlookMessageRef, error) {
	folder := settings.MailFolderPath()
	filter := settings.BuildFilter()
	max := EffectiveMaxEmailsPerSync(settings.MaxEmailsPerSync)
	pageSize := MailSyncPageSize
	if max < pageSize {
		pageSize = max
	}

	messages := make([]outlookMessageRef, 0, max)
	nextURL := fmt.Sprintf(
		"https://graph.microsoft.com/v1.0/me/mailFolders/%s/messages?$top=%d&$orderby=receivedDateTime%%20asc&$select=id",
		url.PathEscape(folder),
		pageSize,
	)
	if filter != "" {
		nextURL += "&$filter=" + url.QueryEscape(filter)
	}

	for nextURL != "" && len(messages) < max {
		var page outlookMessageListResponse
		if err := p.graphGET(ctx, client, nextURL, &page); err != nil {
			return nil, err
		}
		messages = append(messages, page.Value...)
		if len(messages) >= max {
			messages = messages[:max]
			break
		}
		nextURL = page.NextLink
		if nextURL == "" {
			nextURL = page.ODataNextLink
		}
	}

	return messages, nil
}

func (p OutlookProvider) processOutlookMessage(
	ctx context.Context,
	client *http.Client,
	account IntegrationAccount,
	syncCtx SyncContext,
	settings OutlookSettings,
	messageID string,
	result *SyncProviderResult,
) error {
	var message outlookMessage
	messageURL := fmt.Sprintf(
		"https://graph.microsoft.com/v1.0/me/messages/%s?$select=subject,body,from,toRecipients,receivedDateTime,internetMessageId,hasAttachments",
		url.PathEscape(messageID),
	)
	if err := p.graphGET(ctx, client, messageURL, &message); err != nil {
		return err
	}

	subject := decodeMIMEHeaderValue(message.Subject)
	if subject == "" {
		subject = "email"
	}

	baseMetadata := map[string]any{
		"from":               outlookAddressString(message.From),
		"to":                 outlookRecipientsString(message.ToRecipients),
		"subject":            subject,
		"receivedDateTime":   message.ReceivedDateTime,
		"messageId":          message.InternetMessageID,
		"outlookMessageId":   messageID,
	}

	if settings.ImportEmailBody {
		bodyText := extractOutlookBodyText(message.Body)
		if strings.TrimSpace(bodyText) != "" {
			externalID := messageID + ":body"
			outcome := syncCtx.Importer.ImportFile(ctx, ImportFileInput{
				Account:           &account,
				ExternalID:        externalID,
				ExternalType:      ExternalTypeEmailBody,
				OriginalFileName:  safeOutlookBodyFilename(subject),
				FileExtension:     "txt",
				MimeType:          "text/plain",
				Content:           []byte(bodyText),
				SourceMetadata:    baseMetadata,
				AuditAction:       audit.ActionIntegrationOutlookEmailImported,
				TriggeredByUserID: syncCtx.TriggeredByUserID,
			})
			applyOutlookImportOutcome(result, outcome)
		}
	}

	if settings.ImportAttachments && message.HasAttachments {
		attachments, err := p.listAttachments(ctx, client, messageID)
		if err != nil {
			return err
		}
		for _, attachment := range attachments {
			if attachment.ODataType != "#microsoft.graph.fileAttachment" {
				result.SkippedItems++
				continue
			}
			filename := decodeMIMEHeaderValue(attachment.Name)
			if !settings.IsExtensionAllowed(filename) {
				result.SkippedItems++
				continue
			}
			content, err := base64.StdEncoding.DecodeString(strings.TrimSpace(attachment.ContentBytes))
			if err != nil || len(content) == 0 {
				if content, err = p.downloadAttachment(ctx, client, messageID, attachment.ID); err != nil || len(content) == 0 {
					result.FailedItems++
					continue
				}
			}

			checksum := ChecksumHex(content)
			externalID := fmt.Sprintf("%s:attachment:%s", messageID, attachment.ID)
			attachmentMetadata := copyOutlookMetadata(baseMetadata)
			attachmentMetadata["attachmentFileName"] = filename

			outcome := syncCtx.Importer.ImportFile(ctx, ImportFileInput{
				Account:           &account,
				ExternalID:        externalID,
				ExternalType:      ExternalTypeEmailAttachment,
				OriginalFileName:  sanitizeIMAPFilename(filename),
				FileExtension:     imapExtensionFromFilename(filename),
				MimeType:          mimeTypeForIMAPExtension(imapExtensionFromFilename(filename)),
				Content:           content,
				SourceMetadata:    attachmentMetadata,
				Checksum:          checksum,
				AuditAction:       audit.ActionIntegrationOutlookAttachmentImported,
				TriggeredByUserID: syncCtx.TriggeredByUserID,
			})
			applyOutlookImportOutcome(result, outcome)
		}
	}

	return nil
}

func (p OutlookProvider) listAttachments(ctx context.Context, client *http.Client, messageID string) ([]outlookAttachment, error) {
	var response struct {
		Value []outlookAttachment `json:"value"`
	}
	attachmentsURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/messages/%s/attachments", url.PathEscape(messageID))
	if err := p.graphGET(ctx, client, attachmentsURL, &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

func (p OutlookProvider) downloadAttachment(ctx context.Context, client *http.Client, messageID, attachmentID string) ([]byte, error) {
	var attachment outlookAttachment
	attachmentURL := fmt.Sprintf(
		"https://graph.microsoft.com/v1.0/me/messages/%s/attachments/%s",
		url.PathEscape(messageID),
		url.PathEscape(attachmentID),
	)
	if err := p.graphGET(ctx, client, attachmentURL, &attachment); err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(strings.TrimSpace(attachment.ContentBytes))
}

func (p OutlookProvider) graphGET(ctx context.Context, client *http.Client, requestURL string, target any) error {
	if client == nil {
		return errors.New("http client is not configured")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		_, _ = io.ReadAll(io.LimitReader(response.Body, 512))
		return fmt.Errorf("microsoft graph request failed: status=%d", response.StatusCode)
	}
	return json.NewDecoder(response.Body).Decode(target)
}

func extractOutlookBodyText(body outlookBody) string {
	content := strings.TrimSpace(body.Content)
	if content == "" {
		return ""
	}
	if strings.EqualFold(strings.TrimSpace(body.ContentType), "html") {
		return htmlToPlainText(content)
	}
	return content
}

func outlookAddressString(recipient outlookRecipient) string {
	if strings.TrimSpace(recipient.EmailAddress.Name) != "" {
		return fmt.Sprintf("%s <%s>", recipient.EmailAddress.Name, recipient.EmailAddress.Address)
	}
	return strings.TrimSpace(recipient.EmailAddress.Address)
}

func outlookRecipientsString(recipients []outlookRecipient) string {
	parts := make([]string, 0, len(recipients))
	for _, recipient := range recipients {
		if value := outlookAddressString(recipient); value != "" {
			parts = append(parts, value)
		}
	}
	return strings.Join(parts, ", ")
}

func safeOutlookBodyFilename(subject string) string {
	return safeGmailBodyFilename(subject)
}

func copyOutlookMetadata(source map[string]any) map[string]any {
	return copyGmailMetadata(source)
}

func applyOutlookImportOutcome(result *SyncProviderResult, outcome ImportOutcome) {
	applyGmailImportOutcome(result, outcome)
}

func userFriendlyOutlookError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "token refresh failed"), strings.Contains(message, "invalid_grant"), strings.Contains(message, "unauthorized"):
		return fmt.Errorf("Outlook oturumu sona erdi. Hesabi yeniden baglayin")
	case strings.Contains(message, "oauth is not configured"):
		return fmt.Errorf("Outlook OAuth yapilandirmasi eksik")
	default:
		return fmt.Errorf("Outlook senkronizasyonu basarisiz: %s", sanitizeOutlookErrorMessage(err.Error()))
	}
}

func sanitizeOutlookErrorMessage(message string) string {
	lower := strings.ToLower(message)
	if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "auth") {
		return "baglanti hatasi"
	}
	if len(message) > 200 {
		return message[:200]
	}
	return message
}

func BeginOutlookOAuth(ctx context.Context, microsoft MicrosoftOAuthConfig, stateSecret string, tenantID, userID uuid.UUID) (string, error) {
	if !microsoft.IsConfigured() {
		return "", errors.New("Outlook OAuth is not configured")
	}
	payload, err := newOAuthState(tenantID, userID)
	if err != nil {
		return "", err
	}
	state, err := signOAuthState(stateSecret, payload)
	if err != nil {
		return "", err
	}
	authURL := microsoft.oauthConfig().AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
	_ = ctx
	return authURL, nil
}

func CompleteOutlookOAuth(
	ctx context.Context,
	repository *Repository,
	credentials *CredentialService,
	microsoft MicrosoftOAuthConfig,
	stateSecret string,
	code string,
	state string,
) (*IntegrationAccount, error) {
	payload, err := verifyOAuthState(stateSecret, state)
	if err != nil {
		return nil, err
	}
	if !microsoft.IsConfigured() {
		return nil, errors.New("Outlook OAuth is not configured")
	}

	token, err := microsoft.oauthConfig().Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("outlook token exchange failed: %w", err)
	}

	client := microsoft.oauthConfig().Client(ctx, token)
	profile, err := fetchOutlookProfile(client)
	if err != nil {
		return nil, err
	}
	email := strings.TrimSpace(profile.Mail)
	if email == "" {
		email = strings.TrimSpace(profile.UserPrincipalName)
	}
	if email == "" {
		return nil, errors.New("outlook email address could not be resolved")
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

	settings, err := encodeJSONMap(outlookSettingsToMap(defaultOutlookSettings()))
	if err != nil {
		return nil, err
	}

	displayName := fmt.Sprintf("Outlook - %s", email)
	account, err := repository.FindByProviderEmail(ctx, payload.TenantID, payload.UserID, ProviderOutlook, email)
	if err != nil {
		return nil, err
	}

	if account == nil {
		account = &IntegrationAccount{
			TenantID:             payload.TenantID,
			UserID:               payload.UserID,
			Provider:             ProviderOutlook,
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

type outlookProfileResponse struct {
	Mail              string `json:"mail"`
	UserPrincipalName string `json:"userPrincipalName"`
	DisplayName       string `json:"displayName"`
}

func fetchOutlookProfile(client *http.Client) (*outlookProfileResponse, error) {
	if client == nil {
		return nil, errors.New("http client is not configured")
	}
	request, err := http.NewRequest(http.MethodGet, "https://graph.microsoft.com/v1.0/me?$select=mail,userPrincipalName,displayName", nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("outlook profile request failed")
	}
	var profile outlookProfileResponse
	if err := json.NewDecoder(response.Body).Decode(&profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func PersistOutlookTokens(ctx context.Context, repository *Repository, credentials *CredentialService, account *IntegrationAccount, token *oauth2.Token) error {
	if account == nil || token == nil {
		return errors.New("invalid outlook token persistence input")
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
