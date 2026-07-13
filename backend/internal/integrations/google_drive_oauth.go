package integrations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"fifriz/internal/applog"
	"fifriz/internal/audit"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GoogleDriveOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

func (c GoogleDriveOAuthConfig) IsConfigured() bool {
	return strings.TrimSpace(c.ClientID) != "" &&
		strings.TrimSpace(c.ClientSecret) != "" &&
		strings.TrimSpace(c.RedirectURL) != ""
}

func (c GoogleDriveOAuthConfig) oauthConfig() *oauth2.Config {
	scopes := c.Scopes
	if len(scopes) == 0 {
		scopes = []string{drive.DriveReadonlyScope}
	}
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}
}

type GoogleDriveProvider struct {
	googleDrive GoogleDriveOAuthConfig
	credentials *CredentialService
	repository  *Repository
}

func NewGoogleDriveProvider(googleDrive GoogleDriveOAuthConfig, credentials *CredentialService, repository *Repository) GoogleDriveProvider {
	return GoogleDriveProvider{
		googleDrive: googleDrive,
		credentials: credentials,
		repository:  repository,
	}
}

func (GoogleDriveProvider) Name() string { return ProviderGoogleDrive }

func (p GoogleDriveProvider) TestConnection(ctx context.Context, account IntegrationAccount, syncCtx SyncContext) error {
	accountCopy := account
	_, err := p.driveService(ctx, &accountCopy, syncCtx)
	return userFriendlyGoogleDriveError(err)
}

func (p GoogleDriveProvider) Sync(ctx context.Context, account IntegrationAccount, syncCtx SyncContext) (*SyncProviderResult, error) {
	settings, err := ParseGoogleDriveSettings(account.Settings)
	if err != nil {
		return nil, err
	}

	service, err := p.driveService(ctx, &account, syncCtx)
	if err != nil {
		return nil, userFriendlyGoogleDriveError(err)
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

		pageToken := ""
		for {
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			if result.TotalItems >= maxFiles {
				break
			}

			listCall := service.Files.List().
				Q(fmt.Sprintf("'%s' in parents and trashed = false", folderID)).
				Fields("nextPageToken, files(id, name, mimeType, modifiedTime, parents, webViewLink, size)").
				PageSize(100)
			if pageToken != "" {
				listCall = listCall.PageToken(pageToken)
			}

			fileList, err := listCall.Context(ctx).Do()
			if err != nil {
				return nil, userFriendlyGoogleDriveError(err)
			}
			if len(fileList.Files) == 0 {
				break
			}

			for _, file := range fileList.Files {
				if file == nil || strings.TrimSpace(file.Id) == "" {
					continue
				}
				if file.MimeType == GoogleDriveFolderMimeType {
					if settings.IncludeSubfolders {
						folderQueue = append(folderQueue, file.Id)
					}
					continue
				}
				if isGoogleAppsNativeFile(file.MimeType) {
					continue
				}
				if !settings.IsExtensionAllowed(file.Name, file.MimeType) {
					continue
				}
				modifiedAt, parseErr := time.Parse(time.RFC3339, file.ModifiedTime)
				if parseErr == nil && modifiedAt.Before(settings.SyncFromDate) {
					continue
				}

				result.TotalItems++
				if result.TotalItems > maxFiles {
					break
				}

				if err := p.processDriveFile(ctx, service, account, syncCtx, settings, file, result); err != nil {
					applog.Warn("google drive file processing failed: integration_id=%s file_id=%s error=%v", account.ID, file.Id, err)
					result.FailedItems++
				}
			}

			if fileList.NextPageToken == "" || result.TotalItems >= maxFiles {
				break
			}
			pageToken = fileList.NextPageToken
		}
	}

	return result, nil
}

func (p GoogleDriveProvider) processDriveFile(
	ctx context.Context,
	service *drive.Service,
	account IntegrationAccount,
	syncCtx SyncContext,
	settings GoogleDriveSettings,
	file *drive.File,
	result *SyncProviderResult,
) error {
	content, err := downloadDriveFile(ctx, service, file.Id)
	if err != nil || len(content) == 0 {
		return err
	}

	filename := sanitizeIMAPFilename(file.Name)
	if filename == "" {
		filename = file.Id
	}
	extension := imapExtensionFromFilename(filename)
	if extension == "" {
		extension = extensionFromDriveMimeType(file.MimeType)
	}
	if extension == "" {
		return errors.New("unsupported drive file extension")
	}

	parentFolderID := settings.RootFolderID()
	if len(file.Parents) > 0 && strings.TrimSpace(file.Parents[0]) != "" {
		parentFolderID = file.Parents[0]
	}

	checksum := ChecksumHex(content)
	metadata := map[string]any{
		"driveFileId":  file.Id,
		"fileName":     file.Name,
		"mimeType":     file.MimeType,
		"modifiedTime": file.ModifiedTime,
		"webViewLink":  file.WebViewLink,
		"folderId":     parentFolderID,
	}

	outcome := syncCtx.Importer.ImportFile(ctx, ImportFileInput{
		Account:           &account,
		ExternalID:        file.Id,
		ExternalType:      ExternalTypeDriveFile,
		OriginalFileName:  filename,
		FileExtension:     extension,
		MimeType:          mimeTypeForIMAPExtension(extension),
		Content:           content,
		SourceMetadata:    metadata,
		Checksum:          checksum,
		AuditAction:       audit.ActionIntegrationGoogleDriveFileImported,
		TriggeredByUserID: syncCtx.TriggeredByUserID,
	})
	applyGoogleDriveImportOutcome(result, outcome)
	return nil
}

func downloadDriveFile(ctx context.Context, service *drive.Service, fileID string) ([]byte, error) {
	if service == nil {
		return nil, errors.New("drive service is not configured")
	}
	response, err := service.Files.Get(fileID).Context(ctx).Download()
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}

func (p GoogleDriveProvider) driveService(ctx context.Context, account *IntegrationAccount, syncCtx SyncContext) (*drive.Service, error) {
	if !p.googleDrive.IsConfigured() {
		return nil, errors.New("Google Drive OAuth is not configured")
	}
	if syncCtx.Credentials == nil {
		return nil, errors.New("credential service is not configured")
	}

	tokens, err := syncCtx.Credentials.GetOAuthTokens(account.EncryptedCredentials)
	if err != nil {
		return nil, err
	}

	oauthCfg := p.googleDrive.oauthConfig()
	oauthToken := &oauth2.Token{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		Expiry:       tokens.Expiry,
	}

	tokenSource := oauthCfg.TokenSource(ctx, oauthToken)
	freshToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("google drive token refresh failed: %w", err)
	}

	if freshToken.AccessToken != tokens.AccessToken || !freshToken.Expiry.Equal(tokens.Expiry) {
		if err := PersistGoogleDriveTokens(ctx, p.repository, syncCtx.Credentials, account, freshToken); err != nil {
			applog.Warn("google drive token persist failed: integration_id=%s error=%v", account.ID, err)
		}
	}

	client := oauth2.NewClient(ctx, oauth2.ReuseTokenSource(freshToken, tokenSource))
	return drive.NewService(ctx, option.WithHTTPClient(client))
}

func applyGoogleDriveImportOutcome(result *SyncProviderResult, outcome ImportOutcome) {
	switch {
	case outcome.Imported:
		result.ImportedItems++
	case outcome.Skipped:
		result.SkippedItems++
	case outcome.Failed:
		result.FailedItems++
	}
}

func userFriendlyGoogleDriveError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "token refresh failed"), strings.Contains(message, "invalid_grant"), strings.Contains(message, "unauthorized"):
		return fmt.Errorf("Google Drive oturumu sona erdi. Hesabi yeniden baglayin")
	case strings.Contains(message, "oauth is not configured"):
		return fmt.Errorf("Google Drive OAuth yapilandirmasi eksik")
	default:
		return fmt.Errorf("Google Drive senkronizasyonu basarisiz: %s", sanitizeGoogleDriveErrorMessage(err.Error()))
	}
}

func sanitizeGoogleDriveErrorMessage(message string) string {
	lower := strings.ToLower(message)
	if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "auth") {
		return "baglanti hatasi"
	}
	if len(message) > 200 {
		return message[:200]
	}
	return message
}

func BeginGoogleDriveOAuth(ctx context.Context, googleDrive GoogleDriveOAuthConfig, stateSecret string, tenantID, userID uuid.UUID) (string, error) {
	if !googleDrive.IsConfigured() {
		return "", errors.New("Google Drive OAuth is not configured")
	}
	payload, err := newOAuthState(tenantID, userID)
	if err != nil {
		return "", err
	}
	state, err := signOAuthState(stateSecret, payload)
	if err != nil {
		return "", err
	}
	authURL := googleDrive.oauthConfig().AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	_ = ctx
	return authURL, nil
}

func CompleteGoogleDriveOAuth(
	ctx context.Context,
	repository *Repository,
	credentials *CredentialService,
	googleDrive GoogleDriveOAuthConfig,
	stateSecret string,
	code string,
	state string,
) (*IntegrationAccount, error) {
	payload, err := verifyOAuthState(stateSecret, state)
	if err != nil {
		return nil, err
	}
	if !googleDrive.IsConfigured() {
		return nil, errors.New("Google Drive OAuth is not configured")
	}

	token, err := googleDrive.oauthConfig().Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google drive token exchange failed: %w", err)
	}

	client := googleDrive.oauthConfig().Client(ctx, token)
	email, err := fetchGoogleDriveUserEmail(client)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(email) == "" {
		return nil, errors.New("google drive email address could not be resolved")
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

	settings, err := encodeJSONMap(googleDriveSettingsToMap(defaultGoogleDriveSettings()))
	if err != nil {
		return nil, err
	}

	displayName := fmt.Sprintf("Google Drive - %s", email)
	account, err := repository.FindByProviderEmail(ctx, payload.TenantID, payload.UserID, ProviderGoogleDrive, email)
	if err != nil {
		return nil, err
	}

	if account == nil {
		account = &IntegrationAccount{
			TenantID:             payload.TenantID,
			UserID:               payload.UserID,
			Provider:             ProviderGoogleDrive,
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

type googleDriveAboutResponse struct {
	User struct {
		EmailAddress string `json:"emailAddress"`
	} `json:"user"`
}

func fetchGoogleDriveUserEmail(client *http.Client) (string, error) {
	if client == nil {
		return "", errors.New("http client is not configured")
	}
	request, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/drive/v3/about?fields=user", nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return "", fmt.Errorf("google drive about request failed")
	}
	var about googleDriveAboutResponse
	if err := json.NewDecoder(response.Body).Decode(&about); err != nil {
		return "", err
	}
	return strings.TrimSpace(about.User.EmailAddress), nil
}

func PersistGoogleDriveTokens(ctx context.Context, repository *Repository, credentials *CredentialService, account *IntegrationAccount, token *oauth2.Token) error {
	if account == nil || token == nil {
		return errors.New("invalid google drive token persistence input")
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
