package plugins

import (
	"context"

	"fifriz/internal/audit"
	"fifriz/internal/integrations"
	"fifriz/internal/integrations/registry"
	"fifriz/internal/notifications"
)

type outlookPlugin struct {
	inner integrations.OutlookProvider
	deps  Dependencies
}

func newOutlookPlugin(deps Dependencies) registry.Provider {
	return &outlookPlugin{
		inner: integrations.NewOutlookProvider(deps.Microsoft, deps.Credentials, deps.Repository),
		deps:  deps,
	}
}

func (p *outlookPlugin) Name() string        { return integrations.ProviderOutlook }
func (p *outlookPlugin) DisplayName() string { return "Outlook" }
func (p *outlookPlugin) Description() string {
	return "Outlook hesabinizdan e-posta ve ekleri otomatik ice aktarin."
}
func (p *outlookPlugin) Available() bool { return p.deps.Microsoft.IsConfigured() }

func (p *outlookPlugin) Validate(ctx context.Context, account any) error {
	return noopValidate(ctx, account)
}

func (p *outlookPlugin) Connect(ctx context.Context, account any) error {
	return noopConnect(ctx, account)
}

func (p *outlookPlugin) Disconnect(ctx context.Context, account any) error {
	return noopDisconnect(ctx, account)
}

func (p *outlookPlugin) Sync(ctx context.Context, account any, env registry.SyncEnvironment) (*registry.SyncResult, error) {
	item, err := accountFromAny(account)
	if err != nil {
		return nil, err
	}
	result, err := p.inner.Sync(ctx, item, syncEnvironment(env))
	return toSyncResult(result), err
}

func (p *outlookPlugin) Health(ctx context.Context, account any, env registry.SyncEnvironment) (*registry.HealthResult, error) {
	item, err := accountFromAny(account)
	if err != nil {
		return nil, err
	}
	return healthFromError(p.inner.TestConnection(ctx, item, syncEnvironment(env))), nil
}

func (p *outlookPlugin) Capabilities() registry.ProviderCapabilities {
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

func (p *outlookPlugin) ConnectedAuditAction() string          { return audit.ActionIntegrationOutlookConnected }
func (p *outlookPlugin) SyncStartedAuditAction() string      { return audit.ActionIntegrationOutlookSyncStarted }
func (p *outlookPlugin) SyncCompletedAuditAction() string    { return audit.ActionIntegrationOutlookSyncCompleted }
func (p *outlookPlugin) SyncFailedAuditAction() string       { return audit.ActionIntegrationOutlookSyncFailed }
func (p *outlookPlugin) SyncCompletedNotificationType() string { return notifications.TypeOutlookSyncCompleted }
func (p *outlookPlugin) SyncFailedNotificationType() string  { return notifications.TypeOutlookSyncFailed }

func (p *outlookPlugin) BeginOAuth(ctx context.Context, tenantID, userID any) (string, error) {
	tid, uid, err := tenantUserFromAny(tenantID, userID)
	if err != nil {
		return "", err
	}
	return integrations.BeginOutlookOAuth(ctx, p.deps.Microsoft, p.deps.StateSecret, tid, uid)
}

func (p *outlookPlugin) CompleteOAuth(ctx context.Context, code, state string) (any, error) {
	return integrations.CompleteOutlookOAuth(ctx, p.deps.Repository, p.deps.Credentials, p.deps.Microsoft, p.deps.StateSecret, code, state)
}

func (p *outlookPlugin) OAuthCallbackKey() string { return integrations.ProviderOutlook }

var _ registry.OAuthProvider = (*outlookPlugin)(nil)
var _ registry.Provider = (*outlookPlugin)(nil)
