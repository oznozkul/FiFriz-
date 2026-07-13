package integrations

import (
	"context"
	"fmt"
	"strings"

	"fifriz/internal/auth"
	"fifriz/internal/integrations/registry"

	"github.com/google/uuid"
)

func (s *Service) ListProviders() []registry.ProviderDescriptor {
	if s.registry == nil {
		return nil
	}
	return s.registry.List()
}

func (s *Service) CheckHealth(ctx context.Context, idRaw string) (*registry.HealthResult, error) {
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
		return nil, ErrIntegrationNotFound
	}

	plugin, ok := s.registry.Get(account.Provider)
	if !ok {
		return nil, ErrInvalidProvider
	}

	env := registry.SyncEnvironment{Runtime: s.buildSyncContext(authContext.UserID)}
	return plugin.Health(ctx, *account, env)
}

func (s *Service) ProviderConnectURL(ctx context.Context, providerName string) (string, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return "", err
	}

	providerName = strings.TrimSpace(providerName)
	oauth, ok := s.registry.AsOAuth(providerName)
	if !ok {
		return "", fmt.Errorf("%w: provider does not support OAuth", ErrInvalidProvider)
	}
	if plugin, exists := s.registry.Get(providerName); exists && !plugin.Available() {
		return "", fmt.Errorf("%w: provider is not configured", ErrInvalidIntegrationInput)
	}
	return oauth.BeginOAuth(ctx, authContext.TenantID, authContext.UserID)
}

func (s *Service) CompleteProviderOAuth(ctx context.Context, providerName, code, state string) (*IntegrationAccount, error) {
	providerName = strings.TrimSpace(providerName)
	oauth, ok := s.registry.AsOAuth(providerName)
	if !ok {
		return nil, ErrInvalidProvider
	}

	accountAny, err := oauth.CompleteOAuth(ctx, code, state)
	if err != nil {
		return nil, err
	}

	account, ok := accountAny.(*IntegrationAccount)
	if !ok {
		return nil, fmt.Errorf("invalid oauth account result")
	}

	_ = s.enqueueSyncForAccount(ctx, account, SyncTypeInitialHistory, false)
	return account, nil
}

func (s *Service) isRegisteredProvider(name string) bool {
	if name == ProviderManualUpload {
		return true
	}
	return s.registry != nil && s.registry.Has(name)
}

func (s *Service) pluginForAccount(account *IntegrationAccount) (registry.Provider, error) {
	if account == nil {
		return nil, ErrInvalidIntegrationInput
	}
	plugin, ok := s.registry.Get(account.Provider)
	if !ok {
		return nil, ErrInvalidProvider
	}
	return plugin, nil
}

func (s *Service) prepareAccountSync(ctx context.Context, account *IntegrationAccount) error {
	if prepare, ok := s.registry.AsPrepareSync(account.Provider); ok {
		return prepare.PrepareSync(ctx, *account)
	}
	return nil
}
