package plugins

import (
	"context"

	"fifriz/internal/audit"
	"fifriz/internal/integrations"
	"fifriz/internal/integrations/registry"
	"fifriz/internal/notifications"
)

type imapPlugin struct {
	inner integrations.IMAPProvider
	deps  Dependencies
}

func newIMAPPlugin(deps Dependencies) registry.Provider {
	return &imapPlugin{
		inner: integrations.IMAPProvider{},
		deps:  deps,
	}
}

func (p *imapPlugin) Name() string        { return integrations.ProviderIMAP }
func (p *imapPlugin) DisplayName() string { return "IMAP Mail" }
func (p *imapPlugin) Description() string {
	return "IMAP sunucunuz uzerinden e-posta ve ekleri ice aktarin."
}
func (p *imapPlugin) Available() bool { return true }

func (p *imapPlugin) Validate(ctx context.Context, account any) error {
	return noopValidate(ctx, account)
}

func (p *imapPlugin) Connect(ctx context.Context, account any) error {
	return noopConnect(ctx, account)
}

func (p *imapPlugin) Disconnect(ctx context.Context, account any) error {
	return noopDisconnect(ctx, account)
}

func (p *imapPlugin) Sync(ctx context.Context, account any, env registry.SyncEnvironment) (*registry.SyncResult, error) {
	item, err := accountFromAny(account)
	if err != nil {
		return nil, err
	}
	result, err := p.inner.Sync(ctx, item, syncEnvironment(env))
	return toSyncResult(result), err
}

func (p *imapPlugin) Health(ctx context.Context, account any, env registry.SyncEnvironment) (*registry.HealthResult, error) {
	item, err := accountFromAny(account)
	if err != nil {
		return nil, err
	}
	return healthFromError(p.inner.TestConnection(ctx, item, syncEnvironment(env))), nil
}

func (p *imapPlugin) Capabilities() registry.ProviderCapabilities {
	return registry.ProviderCapabilities{
		SupportsPassword:      true,
		SupportsAttachments:   true,
		SupportsHistoryImport: true,
		SupportsScheduling:    true,
		SupportsMetadata:      true,
	}
}

func (p *imapPlugin) ConnectedAuditAction() string          { return audit.ActionIntegrationIMAPConnected }
func (p *imapPlugin) SyncStartedAuditAction() string      { return audit.ActionIntegrationIMAPSyncStarted }
func (p *imapPlugin) SyncCompletedAuditAction() string    { return audit.ActionIntegrationIMAPSyncCompleted }
func (p *imapPlugin) SyncFailedAuditAction() string       { return audit.ActionIntegrationIMAPSyncFailed }
func (p *imapPlugin) SyncCompletedNotificationType() string { return notifications.TypeIntegrationSyncCompleted }
func (p *imapPlugin) SyncFailedNotificationType() string  { return notifications.TypeIntegrationSyncFailed }

func (p *imapPlugin) PrepareSync(ctx context.Context, account any) error {
	item, err := accountFromAny(account)
	if err != nil {
		return err
	}
	return integrations.EnsureIMAPAccountSettings(ctx, p.deps.Repository, p.deps.Credentials, &item)
}

var _ registry.PrepareSyncProvider = (*imapPlugin)(nil)
var _ registry.Provider = (*imapPlugin)(nil)
