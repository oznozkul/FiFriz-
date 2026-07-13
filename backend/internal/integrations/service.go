package integrations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"fifriz/internal/applog"
	"fifriz/internal/auth"
	"fifriz/internal/events"
	domain "fifriz/internal/events/events"
	"fifriz/internal/integrations/registry"
	"fifriz/internal/jobs"
	"fifriz/internal/notifications"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrIntegrationNotFound     = errors.New("integration not found")
	ErrInvalidIntegrationID    = errors.New("invalid integration id")
	ErrInvalidIntegrationInput = errors.New("invalid integration input")
	ErrInvalidProvider         = errors.New("invalid provider")
	ErrIMAPPasswordRequired    = errors.New("IMAP password is required")
	ErrIMAPEmailRequired       = errors.New("IMAP email address is required")
	ErrIntegrationConnection = errors.New("integration connection failed")
)

type Service struct {
	repository    *Repository
	importer      *ImportService
	credentials   *CredentialService
	registry      *registry.Registry
	stateSecret   string
	appURL        string
	publisher     events.Publisher
	notifications notifications.Creator
	jobQueue      jobs.Enqueuer
	syncInterval  time.Duration
}

func NewService(
	repository *Repository,
	importer *ImportService,
	providerRegistry *registry.Registry,
	credentials *CredentialService,
	publisher events.Publisher,
	notificationCreator notifications.Creator,
	jobQueue jobs.Enqueuer,
	stateSecret string,
	appURL string,
	syncInterval time.Duration,
) *Service {
	if credentials == nil {
		credentials = NewCredentialService()
	}
	return &Service{
		repository:    repository,
		importer:      importer,
		credentials:   credentials,
		registry:      providerRegistry,
		stateSecret:   stateSecret,
		appURL:        appURL,
		publisher:     publisher,
		notifications: notificationCreator,
		jobQueue:      jobQueue,
		syncInterval:  syncInterval,
	}
}

func (s *Service) List(ctx context.Context) ([]IntegrationResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	items, err := s.repository.ListByUser(ctx, authContext.TenantID, authContext.UserID)
	if err != nil {
		return nil, err
	}

	responses := make([]IntegrationResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toIntegrationResponse(item))
	}
	return responses, nil
}

func (s *Service) Create(ctx context.Context, input CreateIntegrationRequest) (*IntegrationResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	provider := strings.TrimSpace(input.Provider)
	if !s.isRegisteredProvider(provider) {
		return nil, ErrInvalidIntegrationInput
	}
	plugin, ok := s.registry.Get(provider)
	if !ok {
		return nil, ErrInvalidProvider
	}
	if plugin.Capabilities().SupportsOAuth {
		return nil, fmt.Errorf("%w: use OAuth connect endpoint", ErrInvalidIntegrationInput)
	}
	if provider == ProviderIMAP {
		return s.createIMAP(ctx, authContext, input)
	}

	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		return nil, ErrInvalidIntegrationInput
	}

	settingsMap := SanitizeSettings(input.Settings)
	settings, err := encodeJSONMap(settingsMap)
	if err != nil {
		return nil, err
	}

	status := StatusDisconnected
	if IsMailProvider(provider) {
		status = StatusConnected
	}

	var emailAddress *string
	if input.EmailAddress != nil {
		if trimmed := strings.TrimSpace(*input.EmailAddress); trimmed != "" {
			emailAddress = &trimmed
		}
	}

	account := &IntegrationAccount{
		TenantID:     authContext.TenantID,
		UserID:       authContext.UserID,
		Provider:     provider,
		DisplayName:  displayName,
		EmailAddress: emailAddress,
		Status:       status,
		Settings:     settings,
		IsActive:     true,
	}

	if err := s.repository.Create(ctx, account); err != nil {
		return nil, err
	}

	if plugin.Capabilities().SupportsScheduling {
		_ = s.enqueueSyncForAccount(ctx, account, SyncTypeInitialHistory, false)
	}

	response := toIntegrationResponse(*account)
	return &response, nil
}

func (s *Service) createIMAP(ctx context.Context, authContext auth.Context, input CreateIntegrationRequest) (*IntegrationResponse, error) {
	email := ""
	if input.EmailAddress != nil {
		email = strings.TrimSpace(*input.EmailAddress)
	}
	if email == "" {
		return nil, ErrIMAPEmailRequired
	}
	if input.Credentials == nil || strings.TrimSpace(input.Credentials.Password) == "" {
		return nil, ErrIMAPPasswordRequired
	}

	discovered, err := DiscoverIMAPSettings(email, input.Credentials.Password)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrIntegrationConnection, err.Error())
	}
	if err := applyIMAPSyncFromDate(&discovered, SanitizeSettings(input.Settings)); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidIntegrationInput, err.Error())
	}

	settings, err := encodeJSONMap(IMAPSettingsToMap(discovered))
	if err != nil {
		return nil, err
	}

	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		displayName = email
	}

	encrypted, err := s.credentials.StorePassword(input.Credentials.Password)
	if err != nil {
		return nil, err
	}

	account := &IntegrationAccount{
		TenantID:             authContext.TenantID,
		UserID:               authContext.UserID,
		Provider:             ProviderIMAP,
		DisplayName:          displayName,
		EmailAddress:         &email,
		Status:               StatusConnected,
		Settings:             settings,
		EncryptedCredentials: encrypted,
		IsActive:             true,
	}

	if err := s.repository.Create(ctx, account); err != nil {
		return nil, err
	}

	_ = s.enqueueSyncForAccount(ctx, account, SyncTypeInitialHistory, false)

	response := toIntegrationResponse(*account)
	return &response, nil
}

func (s *Service) GetByID(ctx context.Context, idRaw string) (*IntegrationResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idRaw)
	if err != nil {
		return nil, ErrInvalidIntegrationID
	}

	account, err := s.repository.GetByID(ctx, authContext.TenantID, authContext.UserID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIntegrationNotFound
		}
		return nil, err
	}

	importedCount, err := s.repository.CountImportedByAccount(ctx, authContext.TenantID, account.ID)
	if err != nil {
		return nil, err
	}

	response := toIntegrationResponseWithCount(*account, importedCount)
	s.publisher.Publish(ctx, domain.NewIntegrationOpened(
		authContext.TenantID,
		authContext.UserID,
		account.ID,
		account.DisplayName,
		"/integrations",
	))
	return &response, nil
}

func (s *Service) Update(ctx context.Context, idRaw string, input UpdateIntegrationRequest) (*IntegrationResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idRaw)
	if err != nil {
		return nil, ErrInvalidIntegrationID
	}

	account, err := s.repository.GetByID(ctx, authContext.TenantID, authContext.UserID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIntegrationNotFound
		}
		return nil, err
	}

	if input.DisplayName != nil {
		if trimmed := strings.TrimSpace(*input.DisplayName); trimmed != "" {
			account.DisplayName = trimmed
		}
	}
	if input.Settings != nil {
		if account.Provider == ProviderIMAP {
			merged, mergeErr := mergeIMAPUserSettings(account.Settings, SanitizeSettings(input.Settings))
			if mergeErr != nil {
				return nil, fmt.Errorf("%w: %s", ErrInvalidIntegrationInput, mergeErr.Error())
			}
			account.Settings = merged
		} else {
			settings, encodeErr := encodeJSONMap(SanitizeSettings(input.Settings))
			if encodeErr != nil {
				return nil, encodeErr
			}
			account.Settings = settings
			if account.Provider == ProviderGmail {
				if _, parseErr := ParseGmailSettings(settings); parseErr != nil {
					return nil, ErrInvalidIntegrationInput
				}
			}
			if account.Provider == ProviderOutlook {
				if _, parseErr := ParseOutlookSettings(settings); parseErr != nil {
					return nil, ErrInvalidIntegrationInput
				}
			}
			if account.Provider == ProviderGoogleDrive {
				if _, parseErr := ParseGoogleDriveSettings(settings); parseErr != nil {
					return nil, ErrInvalidIntegrationInput
				}
			}
			if account.Provider == ProviderOneDrive {
				if _, parseErr := ParseOneDriveSettings(settings); parseErr != nil {
					return nil, ErrInvalidIntegrationInput
				}
			}
		}
	}
	if input.IsActive != nil {
		account.IsActive = *input.IsActive
	}
	if input.Credentials != nil && strings.TrimSpace(input.Credentials.Password) != "" {
		password := strings.TrimSpace(input.Credentials.Password)
		if account.Provider == ProviderIMAP && account.EmailAddress != nil && strings.TrimSpace(*account.EmailAddress) != "" {
			preservedSyncFrom, hasPreservedSyncFrom := preservedIMAPSyncFromDate(account.Settings)
			discovered, discoverErr := DiscoverIMAPSettings(*account.EmailAddress, password)
			if discoverErr != nil {
				return nil, fmt.Errorf("%w: %s", ErrIntegrationConnection, discoverErr.Error())
			}
			if hasPreservedSyncFrom {
				discovered.SyncFromDate = preservedSyncFrom
			}
			settings, encodeErr := encodeJSONMap(IMAPSettingsToMap(discovered))
			if encodeErr != nil {
				return nil, encodeErr
			}
			account.Settings = settings
		}
		encrypted, credErr := s.credentials.StorePassword(password)
		if credErr != nil {
			return nil, credErr
		}
		account.EncryptedCredentials = encrypted
	}

	if input.Settings != nil || (input.Credentials != nil && strings.TrimSpace(input.Credentials.Password) != "") {
		account.LastError = nil
		if account.Status == StatusSyncFailed {
			account.Status = StatusConnected
		}
	}

	if account.Provider == ProviderIMAP {
		if len(account.EncryptedCredentials) == 0 {
			return nil, ErrIMAPPasswordRequired
		}
		if _, err := ParseIMAPSettings(account.Settings); err != nil {
			payload := decodeJSONMap(account.Settings)
			if stringValue(payload["syncFromDate"]) == "" {
				return nil, ErrInvalidIntegrationInput
			}
		}
	}

	if err := s.repository.Update(ctx, account); err != nil {
		return nil, err
	}

	importedCount, _ := s.repository.CountImportedByAccount(ctx, authContext.TenantID, account.ID)
	response := toIntegrationResponseWithCount(*account, importedCount)
	return &response, nil
}

func (s *Service) Delete(ctx context.Context, idRaw string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(idRaw)
	if err != nil {
		return ErrInvalidIntegrationID
	}

	if _, err := s.repository.GetByID(ctx, authContext.TenantID, authContext.UserID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrIntegrationNotFound
		}
		return err
	}

	if err := s.repository.SoftDelete(ctx, authContext.TenantID, authContext.UserID, id); err != nil {
		return err
	}

	return nil
}

func (s *Service) QueueSync(ctx context.Context, idRaw string) (*SyncQueuedResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idRaw)
	if err != nil {
		return nil, ErrInvalidIntegrationID
	}

	account, err := s.repository.GetByID(ctx, authContext.TenantID, authContext.UserID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIntegrationNotFound
		}
		return nil, err
	}

	if !account.IsActive {
		return nil, ErrInvalidIntegrationInput
	}

	if err := s.enqueueSyncForAccount(ctx, account, SyncTypeManual, true); err != nil {
		return nil, err
	}

	return &SyncQueuedResponse{Message: "Sync queued"}, nil
}

func (s *Service) ListSyncLogs(ctx context.Context, idRaw string) ([]SyncLogResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idRaw)
	if err != nil {
		return nil, ErrInvalidIntegrationID
	}

	if _, err := s.repository.GetByID(ctx, authContext.TenantID, authContext.UserID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIntegrationNotFound
		}
		return nil, err
	}

	logs, err := s.repository.ListSyncLogs(ctx, authContext.TenantID, authContext.UserID, id)
	if err != nil {
		return nil, err
	}

	responses := make([]SyncLogResponse, 0, len(logs))
	for _, item := range logs {
		responses = append(responses, toSyncLogResponse(item))
	}
	return responses, nil
}

func (s *Service) ProcessSyncIntegrationJob(ctx context.Context, payload jobs.SyncIntegrationPayload) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			applog.Warn("sync_integration panic recovered: integration_id=%s panic=%v", payload.IntegrationAccountID, recovered)
			err = fmt.Errorf("sync worker panic")
		}
	}()

	account, err := s.repository.GetByIDForWorker(ctx, payload.TenantID, payload.IntegrationAccountID)
	if err != nil {
		return err
	}

	syncType := payload.SyncType
	if syncType == "" {
		syncType = SyncTypeManual
	}

	now := time.Now().UTC()
	syncLog := &IntegrationSyncLog{
		TenantID:             payload.TenantID,
		IntegrationAccountID: account.ID,
		SyncType:             syncType,
		Status:               SyncStatusRunning,
		StartedAt:            &now,
	}
	if err := s.repository.CreateSyncLog(ctx, syncLog); err != nil {
		return err
	}

	account.Status = StatusSyncing
	account.LastError = nil
	if err := s.repository.Update(ctx, account); err != nil {
		return err
	}


	if err := s.prepareAccountSync(ctx, account); err != nil {
		failMessage := err.Error()
		_ = s.repository.FailSync(ctx, payload.TenantID, account.ID, syncLog.ID, failMessage)
		s.publishSyncFailed(ctx, account, payload.TriggeredByUserID, failMessage)
		return fmt.Errorf("%s", failMessage)
	}

	plugin, err := s.pluginForAccount(account)
	if err != nil {
		failMessage := "integration provider is not supported"
		_ = s.repository.FailSync(ctx, payload.TenantID, account.ID, syncLog.ID, failMessage)
		s.publishSyncFailed(ctx, account, payload.TriggeredByUserID, failMessage)
		return fmt.Errorf("%s", failMessage)
	}

	syncCtx := s.buildSyncContext(payload.TriggeredByUserID)
	result, syncErr := plugin.Sync(ctx, *account, registry.SyncEnvironment{Runtime: syncCtx})
	if syncErr != nil {
		failMessage := syncErr.Error()
		_ = s.repository.FailSync(ctx, payload.TenantID, account.ID, syncLog.ID, failMessage)
		s.publishSyncFailed(ctx, account, payload.TriggeredByUserID, failMessage)
		return syncErr
	}

	if result == nil {
		result = &registry.SyncResult{}
	}

	completion := SyncCompletion{
		TotalItems:    result.TotalItems,
		ImportedItems: result.ImportedItems,
		SkippedItems:  result.SkippedItems,
		FailedItems:   result.FailedItems,
	}
	if len(result.Metadata) > 0 {
		metadata, encodeErr := json.Marshal(result.Metadata)
		if encodeErr == nil {
			completion.Metadata = metadata
		}
	}

	if err := s.repository.CompleteSync(ctx, payload.TenantID, account.ID, syncLog.ID, completion); err != nil {
		failMessage := err.Error()
		_ = s.repository.FailSync(ctx, payload.TenantID, account.ID, syncLog.ID, failMessage)
		s.publishSyncFailed(ctx, account, payload.TriggeredByUserID, failMessage)
		return err
	}

	s.publishSyncCompleted(ctx, account, payload.TriggeredByUserID, result)
	return nil
}

func (s *Service) GmailConnectURL(ctx context.Context) (string, error) {
	return s.ProviderConnectURL(ctx, ProviderGmail)
}

func (s *Service) CompleteGmailOAuth(ctx context.Context, code, state string) (*IntegrationAccount, error) {
	return s.CompleteProviderOAuth(ctx, ProviderGmail, code, state)
}

func (s *Service) OutlookConnectURL(ctx context.Context) (string, error) {
	return s.ProviderConnectURL(ctx, ProviderOutlook)
}

func (s *Service) CompleteOutlookOAuth(ctx context.Context, code, state string) (*IntegrationAccount, error) {
	return s.CompleteProviderOAuth(ctx, ProviderOutlook, code, state)
}

func (s *Service) GoogleDriveConnectURL(ctx context.Context) (string, error) {
	return s.ProviderConnectURL(ctx, ProviderGoogleDrive)
}

func (s *Service) CompleteGoogleDriveOAuth(ctx context.Context, code, state string) (*IntegrationAccount, error) {
	return s.CompleteProviderOAuth(ctx, ProviderGoogleDrive, code, state)
}

func (s *Service) OneDriveConnectURL(ctx context.Context) (string, error) {
	return s.ProviderConnectURL(ctx, ProviderOneDrive)
}

func (s *Service) CompleteOneDriveOAuth(ctx context.Context, code, state string) (*IntegrationAccount, error) {
	return s.CompleteProviderOAuth(ctx, ProviderOneDrive, code, state)
}

func (s *Service) FrontendIntegrationsURL() string {
	return strings.TrimRight(strings.TrimSpace(s.appURL), "/") + "/integrations"
}

func (s *Service) buildSyncContext(userID uuid.UUID) SyncContext {
	return SyncContext{
		Importer:          s.importer,
		Credentials:       s.credentials,
		TriggeredByUserID: userID,
	}
}

func (s *Service) publishSyncCompleted(ctx context.Context, account *IntegrationAccount, userID uuid.UUID, result *registry.SyncResult) {
	if account == nil {
		return
	}
	imported := 0
	if result != nil {
		imported = result.ImportedItems
	}
	s.publisher.PublishAfterCommit(ctx, domain.NewIntegrationSyncCompleted(
		account.TenantID,
		userID,
		account.ID,
		account.UserID,
		account.Provider,
		account.DisplayName,
		imported,
	))
}

func (s *Service) publishSyncFailed(ctx context.Context, account *IntegrationAccount, userID uuid.UUID, message string) {
	if account == nil {
		return
	}
	s.publisher.Publish(ctx, domain.NewIntegrationSyncFailed(
		account.TenantID,
		userID,
		account.ID,
		account.UserID,
		account.Provider,
		account.DisplayName,
		message,
	))
}
