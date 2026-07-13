package plugins

import (
	"context"
	"errors"
	"fmt"

	"fifriz/internal/integrations"
	"fifriz/internal/integrations/registry"

	"github.com/google/uuid"
)

// Dependencies wires shared integration services into plugins.
type Dependencies struct {
	Repository  *integrations.Repository
	Importer    *integrations.ImportService
	Credentials *integrations.CredentialService
	Google      integrations.GoogleOAuthConfig
	GoogleDrive integrations.GoogleDriveOAuthConfig
	Microsoft   integrations.MicrosoftOAuthConfig
	OneDrive    integrations.OneDriveOAuthConfig
	StateSecret string
}

// RegisterAll registers built-in integration providers.
func RegisterAll(reg *registry.Registry, deps Dependencies) error {
	if reg == nil {
		return errors.New("registry is required")
	}

	registrations := []registry.Provider{
		newGmailPlugin(deps),
		newOutlookPlugin(deps),
		newIMAPPlugin(deps),
		newGoogleDrivePlugin(deps),
		newOneDrivePlugin(deps),
	}

	for _, provider := range registrations {
		if err := reg.Register(provider); err != nil {
			return err
		}
	}
	return nil
}

func syncEnvironment(env registry.SyncEnvironment) integrations.SyncContext {
	if ctx, ok := env.Runtime.(integrations.SyncContext); ok {
		return ctx
	}
	return integrations.SyncContext{}
}

func toSyncResult(result *integrations.SyncProviderResult) *registry.SyncResult {
	if result == nil {
		return &registry.SyncResult{}
	}
	return &registry.SyncResult{
		TotalItems:    result.TotalItems,
		ImportedItems: result.ImportedItems,
		SkippedItems:  result.SkippedItems,
		FailedItems:   result.FailedItems,
		Metadata:      result.Metadata,
	}
}

func accountFromAny(value any) (integrations.IntegrationAccount, error) {
	account, ok := value.(integrations.IntegrationAccount)
	if !ok {
		return integrations.IntegrationAccount{}, fmt.Errorf("invalid integration account")
	}
	return account, nil
}

func tenantUserFromAny(tenantID, userID any) (uuid.UUID, uuid.UUID, error) {
	tid, ok := tenantID.(uuid.UUID)
	if !ok {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid tenant id")
	}
	uid, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid user id")
	}
	return tid, uid, nil
}

func healthFromError(err error) *registry.HealthResult {
	if err == nil {
		return &registry.HealthResult{Healthy: true, Message: "ok"}
	}
	return &registry.HealthResult{Healthy: false, Message: err.Error()}
}

func noopValidate(context.Context, any) error { return nil }
func noopConnect(context.Context, any) error  { return nil }
func noopDisconnect(context.Context, any) error {
	return nil
}
