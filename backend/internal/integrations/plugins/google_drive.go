package plugins

import (
	"context"

	"fifriz/internal/audit"
	"fifriz/internal/integrations"
	"fifriz/internal/integrations/registry"
	"fifriz/internal/notifications"
)

type googleDrivePlugin struct {
	inner integrations.GoogleDriveProvider
	deps  Dependencies
}

func newGoogleDrivePlugin(deps Dependencies) registry.Provider {
	return &googleDrivePlugin{
		inner: integrations.NewGoogleDriveProvider(deps.GoogleDrive, deps.Credentials, deps.Repository),
		deps:  deps,
	}
}

func (p *googleDrivePlugin) Name() string        { return integrations.ProviderGoogleDrive }
func (p *googleDrivePlugin) DisplayName() string { return "Google Drive" }
func (p *googleDrivePlugin) Description() string {
	return "Google Drive dosyalarinizi belge kutunuza aktarin."
}
func (p *googleDrivePlugin) Available() bool { return p.deps.GoogleDrive.IsConfigured() }

func (p *googleDrivePlugin) Validate(ctx context.Context, account any) error {
	return noopValidate(ctx, account)
}

func (p *googleDrivePlugin) Connect(ctx context.Context, account any) error {
	return noopConnect(ctx, account)
}

func (p *googleDrivePlugin) Disconnect(ctx context.Context, account any) error {
	return noopDisconnect(ctx, account)
}

func (p *googleDrivePlugin) Sync(ctx context.Context, account any, env registry.SyncEnvironment) (*registry.SyncResult, error) {
	item, err := accountFromAny(account)
	if err != nil {
		return nil, err
	}
	result, err := p.inner.Sync(ctx, item, syncEnvironment(env))
	return toSyncResult(result), err
}

func (p *googleDrivePlugin) Health(ctx context.Context, account any, env registry.SyncEnvironment) (*registry.HealthResult, error) {
	item, err := accountFromAny(account)
	if err != nil {
		return nil, err
	}
	return healthFromError(p.inner.TestConnection(ctx, item, syncEnvironment(env))), nil
}

func (p *googleDrivePlugin) Capabilities() registry.ProviderCapabilities {
	return registry.ProviderCapabilities{
		SupportsOAuth:           true,
		SupportsAttachments:     true,
		SupportsFolders:         true,
		SupportsHistoryImport:   true,
		SupportsIncrementalSync: true,
		SupportsScheduling:      true,
		SupportsMetadata:        true,
	}
}

func (p *googleDrivePlugin) ConnectedAuditAction() string          { return audit.ActionIntegrationGoogleDriveConnected }
func (p *googleDrivePlugin) SyncStartedAuditAction() string      { return audit.ActionIntegrationGoogleDriveSyncStarted }
func (p *googleDrivePlugin) SyncCompletedAuditAction() string    { return audit.ActionIntegrationGoogleDriveSyncCompleted }
func (p *googleDrivePlugin) SyncFailedAuditAction() string       { return audit.ActionIntegrationGoogleDriveSyncFailed }
func (p *googleDrivePlugin) SyncCompletedNotificationType() string { return notifications.TypeGoogleDriveSyncCompleted }
func (p *googleDrivePlugin) SyncFailedNotificationType() string  { return notifications.TypeGoogleDriveSyncFailed }

func (p *googleDrivePlugin) BeginOAuth(ctx context.Context, tenantID, userID any) (string, error) {
	tid, uid, err := tenantUserFromAny(tenantID, userID)
	if err != nil {
		return "", err
	}
	return integrations.BeginGoogleDriveOAuth(ctx, p.deps.GoogleDrive, p.deps.StateSecret, tid, uid)
}

func (p *googleDrivePlugin) CompleteOAuth(ctx context.Context, code, state string) (any, error) {
	return integrations.CompleteGoogleDriveOAuth(ctx, p.deps.Repository, p.deps.Credentials, p.deps.GoogleDrive, p.deps.StateSecret, code, state)
}

func (p *googleDrivePlugin) OAuthCallbackKey() string { return integrations.ProviderGoogleDrive }

var _ registry.OAuthProvider = (*googleDrivePlugin)(nil)
var _ registry.Provider = (*googleDrivePlugin)(nil)
