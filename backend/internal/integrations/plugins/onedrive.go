package plugins

import (
	"context"

	"fifriz/internal/audit"
	"fifriz/internal/integrations"
	"fifriz/internal/integrations/registry"
	"fifriz/internal/notifications"
)

type oneDrivePlugin struct {
	inner integrations.OneDriveProvider
	deps  Dependencies
}

func newOneDrivePlugin(deps Dependencies) registry.Provider {
	return &oneDrivePlugin{
		inner: integrations.NewOneDriveProvider(deps.OneDrive, deps.Credentials, deps.Repository),
		deps:  deps,
	}
}

func (p *oneDrivePlugin) Name() string        { return integrations.ProviderOneDrive }
func (p *oneDrivePlugin) DisplayName() string { return "OneDrive" }
func (p *oneDrivePlugin) Description() string {
	return "OneDrive ve Microsoft 365 dosyalarinizi belge kutunuza aktarin."
}
func (p *oneDrivePlugin) Available() bool { return p.deps.OneDrive.IsConfigured() }

func (p *oneDrivePlugin) Validate(ctx context.Context, account any) error {
	return noopValidate(ctx, account)
}

func (p *oneDrivePlugin) Connect(ctx context.Context, account any) error {
	return noopConnect(ctx, account)
}

func (p *oneDrivePlugin) Disconnect(ctx context.Context, account any) error {
	return noopDisconnect(ctx, account)
}

func (p *oneDrivePlugin) Sync(ctx context.Context, account any, env registry.SyncEnvironment) (*registry.SyncResult, error) {
	item, err := accountFromAny(account)
	if err != nil {
		return nil, err
	}
	result, err := p.inner.Sync(ctx, item, syncEnvironment(env))
	return toSyncResult(result), err
}

func (p *oneDrivePlugin) Health(ctx context.Context, account any, env registry.SyncEnvironment) (*registry.HealthResult, error) {
	item, err := accountFromAny(account)
	if err != nil {
		return nil, err
	}
	return healthFromError(p.inner.TestConnection(ctx, item, syncEnvironment(env))), nil
}

func (p *oneDrivePlugin) Capabilities() registry.ProviderCapabilities {
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

func (p *oneDrivePlugin) ConnectedAuditAction() string           { return audit.ActionIntegrationOneDriveConnected }
func (p *oneDrivePlugin) SyncStartedAuditAction() string         { return audit.ActionIntegrationOneDriveSyncStarted }
func (p *oneDrivePlugin) SyncCompletedAuditAction() string       { return audit.ActionIntegrationOneDriveSyncCompleted }
func (p *oneDrivePlugin) SyncFailedAuditAction() string          { return audit.ActionIntegrationOneDriveSyncFailed }
func (p *oneDrivePlugin) SyncCompletedNotificationType() string { return notifications.TypeOneDriveSyncCompleted }
func (p *oneDrivePlugin) SyncFailedNotificationType() string    { return notifications.TypeOneDriveSyncFailed }

func (p *oneDrivePlugin) BeginOAuth(ctx context.Context, tenantID, userID any) (string, error) {
	tid, uid, err := tenantUserFromAny(tenantID, userID)
	if err != nil {
		return "", err
	}
	return integrations.BeginOneDriveOAuth(ctx, p.deps.OneDrive, p.deps.StateSecret, tid, uid)
}

func (p *oneDrivePlugin) CompleteOAuth(ctx context.Context, code, state string) (any, error) {
	return integrations.CompleteOneDriveOAuth(ctx, p.deps.Repository, p.deps.Credentials, p.deps.OneDrive, p.deps.StateSecret, code, state)
}

func (p *oneDrivePlugin) OAuthCallbackKey() string { return integrations.ProviderOneDrive }

var _ registry.OAuthProvider = (*oneDrivePlugin)(nil)
var _ registry.Provider = (*oneDrivePlugin)(nil)
