package dashboard

import (
	"context"

	"fifriz/internal/auth"
	"fifriz/internal/cache"
	"fifriz/internal/permissions"
)

type Service struct {
	repository *Repository
	cache      *cache.Service
}

func NewService(repository *Repository, cacheService *cache.Service) *Service {
	return &Service{
		repository: repository,
		cache:      cacheService,
	}
}

func (s *Service) GetSummary(ctx context.Context) (*SummaryResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	return cache.GetOrSetJSON(ctx, s.cache, cache.DashboardKey(authContext.TenantID, authContext.UserID), cache.TTLDashboard, func() (*SummaryResponse, error) {
		return s.repository.GetSummary(ctx, authContext.TenantID, authContext.UserID, permissions.HasTenantWideAccess(permissions.Role(authContext.Role)))
	})
}
