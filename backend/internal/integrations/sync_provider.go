package integrations

import (
	"context"

	"github.com/google/uuid"
)

// SyncContext carries runtime dependencies for legacy sync provider implementations.
type SyncContext struct {
	Importer          *ImportService
	Credentials       *CredentialService
	TriggeredByUserID uuid.UUID
}

// SyncProviderResult is the legacy sync result type used by provider implementations.
type SyncProviderResult struct {
	TotalItems    int
	ImportedItems int
	SkippedItems  int
	FailedItems   int
	Metadata      map[string]any
}

// SyncProvider is the legacy sync contract implemented by mail and drive providers.
type SyncProvider interface {
	Name() string
	TestConnection(ctx context.Context, account IntegrationAccount, syncCtx SyncContext) error
	Sync(ctx context.Context, account IntegrationAccount, syncCtx SyncContext) (*SyncProviderResult, error)
}
