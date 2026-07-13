package registry

import (
	"context"
	"errors"
)

var ErrProviderNotFound = errors.New("integration provider not found")

// ProviderCapabilities describes what an integration plugin supports.
type ProviderCapabilities struct {
	SupportsOAuth             bool `json:"supportsOAuth"`
	SupportsPassword          bool `json:"supportsPassword"`
	SupportsWebhook           bool `json:"supportsWebhook"`
	SupportsRealtime          bool `json:"supportsRealtime"`
	SupportsAttachments       bool `json:"supportsAttachments"`
	SupportsFolders           bool `json:"supportsFolders"`
	SupportsHistoryImport       bool `json:"supportsHistoryImport"`
	SupportsIncrementalSync   bool `json:"supportsIncrementalSync"`
	SupportsScheduling          bool `json:"supportsScheduling"`
	SupportsSearch              bool `json:"supportsSearch"`
	SupportsLabels              bool `json:"supportsLabels"`
	SupportsMetadata            bool `json:"supportsMetadata"`
}

// ProviderDescriptor is returned by the providers list API.
type ProviderDescriptor struct {
	Name         string               `json:"name"`
	DisplayName  string               `json:"displayName"`
	Description  string               `json:"description"`
	Available    bool                 `json:"available"`
	Capabilities ProviderCapabilities `json:"capabilities"`
}

// SyncResult is the outcome of a provider sync run.
type SyncResult struct {
	TotalItems    int
	ImportedItems int
	SkippedItems  int
	FailedItems   int
	Metadata      map[string]any
}

// HealthResult is the outcome of a provider health check.
type HealthResult struct {
	Healthy bool              `json:"healthy"`
	Message string            `json:"message"`
	Details map[string]any    `json:"details,omitempty"`
}

// SyncEnvironment carries runtime dependencies for sync and health checks.
// Concrete wiring lives in the integrations package; plugins receive this via adapters.
type SyncEnvironment struct {
	Runtime any
}

// Provider is the plugin SDK contract for integration providers.
type Provider interface {
	Name() string
	DisplayName() string
	Description() string
	Available() bool

	Validate(ctx context.Context, account any) error
	Connect(ctx context.Context, account any) error
	Disconnect(ctx context.Context, account any) error
	Sync(ctx context.Context, account any, env SyncEnvironment) (*SyncResult, error)
	Health(ctx context.Context, account any, env SyncEnvironment) (*HealthResult, error)

	Capabilities() ProviderCapabilities

	ConnectedAuditAction() string
	SyncStartedAuditAction() string
	SyncCompletedAuditAction() string
	SyncFailedAuditAction() string
	SyncCompletedNotificationType() string
	SyncFailedNotificationType() string
}

// OAuthProvider extends Provider with OAuth connect flows.
type OAuthProvider interface {
	Provider
	BeginOAuth(ctx context.Context, tenantID, userID any) (authURL string, err error)
	CompleteOAuth(ctx context.Context, code, state string) (account any, err error)
	OAuthCallbackKey() string
}

// PasswordCreateProvider supports credential-based account creation (e.g. IMAP).
type PasswordCreateProvider interface {
	Provider
	CreateWithPassword(ctx context.Context, input any) (account any, err error)
}

// PrepareSyncProvider can run pre-sync setup (e.g. IMAP settings discovery).
type PrepareSyncProvider interface {
	Provider
	PrepareSync(ctx context.Context, account any) error
}
