package integrations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fifriz/internal/applog"
	"fifriz/internal/audit"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

type OneDriveOAuthConfig struct {
	ClientID     string
	ClientSecret string
	Tenant       string
	RedirectURL  string
	Scopes       []string
}

func (c OneDriveOAuthConfig) IsConfigured() bool {
	return strings.TrimSpace(c.ClientID) != "" &&
		strings.TrimSpace(c.ClientSecret) != "" &&
		strings.TrimSpace(c.RedirectURL) != ""
}

func (c OneDriveOAuthConfig) tenantID() string {
	tenant := strings.TrimSpace(c.Tenant)
	if tenant == "" {
		return "common"
	}
	return tenant
}

func (c OneDriveOAuthConfig) oauthConfig() *oauth2.Config {
	scopes := c.Scopes
	if len(scopes) == 0 {
		scopes = []string{"offline_access", "User.Read", "Files.Read.All"}
	}
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURL,
		Scopes:       scopes,
		Endpoint:     microsoft.AzureADEndpoint(c.tenantID()),
	}
}

type OneDriveProvider struct {
	onedrive    OneDriveOAuthConfig
	credentials *CredentialService
	repository  *Repository
}

func NewOneDriveProvider(onedrive OneDriveOAuthConfig, credentials *CredentialService, repository *Repository) OneDriveProvider {
	return OneDriveProvider{
		onedrive:    onedrive,
		credentials: credentials,
		repository:  repository,
	}
}

func (OneDriveProvider) Name() string { return ProviderOneDrive }

func (p OneDriveProvider) TestConnection(ctx context.Context, account IntegrationAccount, syncCtx SyncContext) error {
	accountCopy := account
	client, err := p.graphClient(ctx, &accountCopy, syncCtx)
	if err != nil {
		return userFriendlyOneDriveError(err)
	}
	var payload struct {
		ID string `json:"id"`
	}
	return userFriendlyOneDriveError(oneDriveGraphGET(ctx, client, "https://graph.microsoft.com/v1.0/me/drive/root?$select=id", &payload))
}

func (p OneDriveProvider) Sync(ctx context.Context, account IntegrationAccount, syncCtx SyncContext) (*SyncProviderResult, error) {
	settings, err := ParseOneDriveSettings(account.Settings)
	if err != nil {
		return nil, err
	}

	client, err := p.graphClient(ctx, &account, syncCtx)
	if err != nil {
		return nil, userFriendlyOneDriveError(err)
	}

	maxFiles := EffectiveMaxFilesPerSync(settings.MaxFilesPerSync)
	result := &SyncProviderResult{
		Metadata: map[string]any{
			"rootFolderId":      settings.RootFolderID(),
			"includeSubfolders": settings.IncludeSubfolders,
			"maxFiles":          maxFiles,
		},
	}

	rootFolderID := settings.RootFolderID()
	folderQueue := []string{rootFolderID}
	visitedFolders := map[string]bool{}

	for len(folderQueue) > 0 && result.TotalItems < maxFiles {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		folderID := folderQueue[0]
		folderQueue = folderQueue[1:]
		if visitedFolders[folderID] {
			continue
		}
		visitedFolders[folderID] = true

		nextURL := oneDriveChildrenURL(folderID)
		for nextURL != "" {
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			if result.TotalItems >= maxFiles {
				break
			}

			var page oneDriveItemListResponse
			if err := oneDriveGraphGET(ctx, client, nextURL, &page); err != nil {
				return nil, userFriendlyOneDriveError(err)
			}
			if len(page.Value) == 0 {
				break
			}

			for _, item := range page.Value {
				if strings.TrimSpace(item.ID) == "" {
					continue
				}
				if item.Folder != nil {
					if settings.IncludeSubfolders {
						folderQueue = append(folderQueue, item.ID)
					}
					continue
				}
				if item.File == nil {
					continue
				}
				if !settings.IsExtensionAllowed(item.Name, item.File.MimeType) {
					continue
				}
				modifiedAt, parseErr := time.Parse(time.RFC3339, item.LastModifiedDateTime)
				if parseErr == nil && modifiedAt.Before(settings.SyncFromDate) {
					continue
				}

				result.TotalItems++
				if result.TotalItems > maxFiles {
					break
				}

				if err := p.processDriveItem(ctx, client, account, syncCtx, settings, item, result); err != nil {
					applog.Warn("onedrive file processing failed: integration_id=%s item_id=%s error=%v", account.ID, item.ID, err)
					result.FailedItems++
				}
			}

			nextURL = page.NextLink
			if nextURL == "" {
				nextURL = page.ODataNextLink
			}
		}
	}

	return result, nil
}

type oneDriveItemListResponse struct {
	Value         []oneDriveItem `json:"value"`
	NextLink      string         `json:"@odata.nextLink"`
	ODataNextLink string         `json:"odata.nextLink"`
}

type oneDriveItem struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	LastModifiedDateTime string `json:"lastModifiedDateTime"`
	WebURL               string `json:"webUrl"`
	File                 *struct {
		MimeType string `json:"mimeType"`
	} `json:"file"`
	Folder *struct {
		ChildCount int `json:"childCount"`
	} `json:"folder"`
	ParentReference *struct {
		ID string `json:"id"`
	} `json:"parentReference"`
}

func oneDriveChildrenURL(folderID string) string {
	selectQuery := url.QueryEscape("id,name,file,folder,lastModifiedDateTime,webUrl,parentReference")
	if folderID == "" || folderID == "root" {
		return fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/root/children?$top=100&$select=%s", selectQuery)
	}
	return fmt.Sprintf(
		"https://graph.microsoft.com/v1.0/me/drive/items/%s/children?$top=100&$select=%s",
		url.PathEscape(folderID),
		selectQuery,
	)
}

func (p OneDriveProvider) processDriveItem(
	ctx context.Context,
	client *http.Client,
	account IntegrationAccount,
	syncCtx SyncContext,
	settings OneDriveSettings,
	item oneDriveItem,
	result *SyncProviderResult,
) error {
	content, err := downloadOneDriveItem(ctx, client, item.ID)
	if err != nil || len(content) == 0 {
		return err
	}

	filename := sanitizeIMAPFilename(item.Name)
	if filename == "" {
		filename = item.ID
	}
	extension := imapExtensionFromFilename(filename)
	if extension == "" && item.File != nil {
		extension = extensionFromGraphMimeType(item.File.MimeType)
	}
	if extension == "" {
		return errors.New("unsupported onedrive file extension")
	}

	parentFolderID := settings.RootFolderID()
	if item.ParentReference != nil && strings.TrimSpace(item.ParentReference.ID) != "" {
		parentFolderID = item.ParentReference.ID
	}

	mimeType := ""
	if item.File != nil {
		mimeType = item.File.MimeType
	}

	metadata := map[string]any{
		"fileName":     item.Name,
		"mimeType":     mimeType,
		"modifiedTime": item.LastModifiedDateTime,
		"webUrl":       item.WebURL,
		"folderId":     parentFolderID,
		"driveItemId":  item.ID,
	}

	outcome := syncCtx.Importer.ImportFile(ctx, ImportFileInput{
		Account:           &account,
		ExternalID:        item.ID,
		ExternalType:      ExternalTypeDriveFile,
		OriginalFileName:  filename,
		FileExtension:     extension,
		MimeType:          mimeTypeForIMAPExtension(extension),
		Content:           content,
		SourceMetadata:    metadata,
		Checksum:          ChecksumHex(content),
		AuditAction:       audit.ActionIntegrationOneDriveFileImported,
		TriggeredByUserID: syncCtx.TriggeredByUserID,
	})
	applyOneDriveImportOutcome(result, outcome)
	return nil
}

func downloadOneDriveItem(ctx context.Context, client *http.Client, itemID string) ([]byte, error) {
	if client == nil {
		return nil, errors.New("http client is not configured")
	}
	requestURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s/content", url.PathEscape(itemID))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("onedrive download failed: status=%d", response.StatusCode)
	}
	return io.ReadAll(response.Body)
}

func (p OneDriveProvider) graphClient(ctx context.Context, account *IntegrationAccount, syncCtx SyncContext) (*http.Client, error) {
	if !p.onedrive.IsConfigured() {
		return nil, errors.New("OneDrive OAuth is not configured")
	}
	if syncCtx.Credentials == nil {
		return nil, errors.New("credential service is not configured")
	}

	tokens, err := syncCtx.Credentials.GetOAuthTokens(account.EncryptedCredentials)
	if err != nil {
		return nil, err
	}

	oauthCfg := p.onedrive.oauthConfig()
	oauthToken := &oauth2.Token{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		Expiry:       tokens.Expiry,
	}

	tokenSource := oauthCfg.TokenSource(ctx, oauthToken)
	freshToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("onedrive token refresh failed: %w", err)
	}

	if freshToken.AccessToken != tokens.AccessToken || !freshToken.Expiry.Equal(tokens.Expiry) {
		if err := PersistOneDriveTokens(ctx, p.repository, syncCtx.Credentials, account, freshToken); err != nil {
			applog.Warn("onedrive token persist failed: integration_id=%s error=%v", account.ID, err)
		}
	}

	return oauth2.NewClient(ctx, oauth2.ReuseTokenSource(freshToken, tokenSource)), nil
}

func oneDriveGraphGET(ctx context.Context, client *http.Client, requestURL string, target any) error {
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

func applyOneDriveImportOutcome(result *SyncProviderResult, outcome ImportOutcome) {
	switch {
	case outcome.Imported:
		result.ImportedItems++
	case outcome.Skipped:
		result.SkippedItems++
	case outcome.Failed:
		result.FailedItems++
	}
}

func userFriendlyOneDriveError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "token refresh failed"), strings.Contains(message, "invalid_grant"), strings.Contains(message, "unauthorized"):
		return fmt.Errorf("OneDrive oturumu sona erdi. Hesabi yeniden baglayin")
	case strings.Contains(message, "oauth is not configured"):
		return fmt.Errorf("OneDrive OAuth yapilandirmasi eksik")
	default:
		return fmt.Errorf("OneDrive senkronizasyonu basarisiz: %s", sanitizeOneDriveErrorMessage(err.Error()))
	}
}

func sanitizeOneDriveErrorMessage(message string) string {
	lower := strings.ToLower(message)
	if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "auth") {
		return "baglanti hatasi"
	}
	if len(message) > 200 {
		return message[:200]
	}
	return message
}

func BeginOneDriveOAuth(ctx context.Context, onedrive OneDriveOAuthConfig, stateSecret string, tenantID, userID uuid.UUID) (string, error) {
	if !onedrive.IsConfigured() {
		return "", errors.New("OneDrive OAuth is not configured")
	}
	payload, err := newOAuthState(tenantID, userID)
	if err != nil {
		return "", err
	}
	state, err := signOAuthState(stateSecret, payload)
	if err != nil {
		return "", err
	}
	authURL := onedrive.oauthConfig().AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
	_ = ctx
	return authURL, nil
}

func CompleteOneDriveOAuth(
	ctx context.Context,
	repository *Repository,
	credentials *CredentialService,
	onedrive OneDriveOAuthConfig,
	stateSecret string,
	code string,
	state string,
) (*IntegrationAccount, error) {
	payload, err := verifyOAuthState(stateSecret, state)
	if err != nil {
		return nil, err
	}
	if !onedrive.IsConfigured() {
		return nil, errors.New("OneDrive OAuth is not configured")
	}

	token, err := onedrive.oauthConfig().Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("onedrive token exchange failed: %w", err)
	}

	client := onedrive.oauthConfig().Client(ctx, token)
	profile, err := fetchOutlookProfile(client)
	if err != nil {
		return nil, err
	}
	email := strings.TrimSpace(profile.Mail)
	if email == "" {
		email = strings.TrimSpace(profile.UserPrincipalName)
	}
	if email == "" {
		return nil, errors.New("onedrive email address could not be resolved")
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

	settings, err := encodeJSONMap(oneDriveSettingsToMap(defaultOneDriveSettings()))
	if err != nil {
		return nil, err
	}

	displayName := fmt.Sprintf("OneDrive - %s", email)
	account, err := repository.FindByProviderEmail(ctx, payload.TenantID, payload.UserID, ProviderOneDrive, email)
	if err != nil {
		return nil, err
	}

	if account == nil {
		account = &IntegrationAccount{
			TenantID:             payload.TenantID,
			UserID:               payload.UserID,
			Provider:             ProviderOneDrive,
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

func PersistOneDriveTokens(ctx context.Context, repository *Repository, credentials *CredentialService, account *IntegrationAccount, token *oauth2.Token) error {
	if account == nil || token == nil {
		return errors.New("invalid onedrive token persistence input")
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
