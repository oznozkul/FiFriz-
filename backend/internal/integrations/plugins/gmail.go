package plugins

import (
	"context"

	"fifriz/internal/audit"
	"fifriz/internal/integrations"
	"fifriz/internal/integrations/registry"
	"fifriz/internal/notifications"
)

type gmailPlugin struct {
	inner integrations.GmailProvider
	deps  Dependencies
}

func newGmailPlugin(deps Dependencies) registry.Provider {
	return &gmailPlugin{
		inner: integrations.NewGmailProvider(deps.Google, deps.Credentials, deps.Repository),
		deps:  deps,
	}
}

func (p *gmailPlugin) Name() string        { return integrations.ProviderGmail }
func (p *gmailPlugin) DisplayName() string { return "Gmail" }
func (p *gmailPlugin) Description() string {
	return "Gmail hesabinizdan e-posta ve ekleri otomatik ice aktarin."
}
func (p *gmailPlugin) Available() bool { return p.deps.Google.IsConfigured() }

func (p *gmailPlugin) Validate(ctx context.Context, account any) error {
	return noopValidate(ctx, account)
}

func (p *gmailPlugin) Connect(ctx context.Context, account any) error {
	return noopConnect(ctx, account)
}

func (p *gmailPlugin) Disconnect(ctx context.Context, account any) error {
	return noopDisconnect(ctx, account)
}

func (p *gmailPlugin) Sync(ctx context.Context, account any, env registry.SyncEnvironment) (*registry.SyncResult, error) {
	item, err := accountFromAny(account)
	if err != nil {
		return nil, err
	}
	result, err := p.inner.Sync(ctx, item, syncEnvironment(env))
	return toSyncResult(result), err
}

func (p *gmailPlugin) Health(ctx context.Context, account any, env registry.SyncEnvironment) (*registry.HealthResult, error) {
	item, err := accountFromAny(account)
	if err != nil {
		return nil, err
	}
	return healthFromError(p.inner.TestConnection(ctx, item, syncEnvironment(env))), nil
}

func (p *gmailPlugin) Capabilities() registry.ProviderCapabilities {
	return registry.ProviderCapabilities{
		SupportsOAuth:           true,
		SupportsAttachments:     true,
		SupportsHistoryImport:   true,
		SupportsIncrementalSync: true,
		SupportsScheduling:      true,
		SupportsSearch:          true,
		SupportsLabels:          true,
		SupportsMetadata:        true,
	}
}

func (p *gmailPlugin) ConnectedAuditAction() string          { return audit.ActionIntegrationGmailConnected }
func (p *gmailPlugin) SyncStartedAuditAction() string      { return audit.ActionIntegrationGmailSyncStarted }
func (p *gmailPlugin) SyncCompletedAuditAction() string    { return audit.ActionIntegrationGmailSyncCompleted }
func (p *gmailPlugin) SyncFailedAuditAction() string       { return audit.ActionIntegrationGmailSyncFailed }
func (p *gmailPlugin) SyncCompletedNotificationType() string { return notifications.TypeGmailSyncCompleted }
func (p *gmailPlugin) SyncFailedNotificationType() string  { return notifications.TypeGmailSyncFailed }

func (p *gmailPlugin) BeginOAuth(ctx context.Context, tenantID, userID any) (string, error) {
	tid, uid, err := tenantUserFromAny(tenantID, userID)
	if err != nil {
		return "", err
	}
	return integrations.BeginGmailOAuth(ctx, p.deps.Google, p.deps.StateSecret, tid, uid)
}

func (p *gmailPlugin) CompleteOAuth(ctx context.Context, code, state string) (any, error) {
	return integrations.CompleteGmailOAuth(ctx, p.deps.Repository, p.deps.Credentials, p.deps.Google, p.deps.StateSecret, code, state)
}

func (p *gmailPlugin) OAuthCallbackKey() string { return integrations.ProviderGmail }

var _ registry.OAuthProvider = (*gmailPlugin)(nil)
var _ registry.Provider = (*gmailPlugin)(nil)
