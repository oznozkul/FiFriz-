package navigation

import (
	"context"
	"errors"
	"strings"

	"fifriz/internal/auth"
	"fifriz/internal/cache"
	"fifriz/internal/events"
	domain "fifriz/internal/events/events"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	repository  *Repository
	recentStore *RecentStore
	cache       *cache.Service
	invalidator *cache.Invalidator
	publisher   events.Publisher
}

func NewService(
	repository *Repository,
	recentStore *RecentStore,
	cacheService *cache.Service,
	invalidator *cache.Invalidator,
	publisher events.Publisher,
) *Service {
	return &Service{
		repository:  repository,
		recentStore: recentStore,
		cache:       cacheService,
		invalidator: invalidator,
		publisher:   publisher,
	}
}

func (s *Service) ListPins(ctx context.Context) ([]PinResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	return cache.GetOrSetJSON(ctx, s.cache, cache.PinsListKey(authContext.TenantID, authContext.UserID), cache.TTLNavigationPins, func() ([]PinResponse, error) {
		items, err := s.repository.ListByUser(ctx, authContext.TenantID, authContext.UserID)
		if err != nil {
			return nil, err
		}
		responses := make([]PinResponse, 0, len(items))
		for _, item := range items {
			responses = append(responses, toPinResponse(item))
		}
		return responses, nil
	})
}

func (s *Service) CreatePin(ctx context.Context, input CreatePinRequest) (*PinResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	entityType := strings.TrimSpace(input.EntityType)
	entityID := strings.TrimSpace(input.EntityID)
	label := strings.TrimSpace(input.Label)
	path := strings.TrimSpace(input.Path)
	if entityType == "" || entityID == "" || label == "" || path == "" {
		return nil, ErrInvalidPinData
	}

	if existing, err := s.repository.FindByEntity(ctx, authContext.TenantID, authContext.UserID, entityType, entityID); err == nil && existing != nil {
		response := toPinResponse(*existing)
		return &response, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	item := &PinnedItem{
		ID:         uuid.New(),
		TenantID:   authContext.TenantID,
		UserID:     authContext.UserID,
		EntityType: entityType,
		EntityID:   entityID,
		Label:      label,
		Path:       path,
	}
	if err := s.repository.Create(ctx, item); err != nil {
		return nil, err
	}

	s.publisher.Publish(ctx, domain.NewPinnedCreated(
		authContext.TenantID,
		authContext.UserID,
		item.ID,
		item.EntityType,
		item.EntityID,
		item.Label,
		item.Path,
	))

	response := toPinResponse(*item)
	return &response, nil
}

func (s *Service) DeletePin(ctx context.Context, pinIDRaw string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	pinID, err := uuid.Parse(pinIDRaw)
	if err != nil {
		return ErrInvalidPinData
	}

	item, err := s.repository.GetByID(ctx, authContext.TenantID, authContext.UserID, pinID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPinNotFound
		}
		return err
	}

	if err := s.repository.Delete(ctx, authContext.TenantID, authContext.UserID, pinID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPinNotFound
		}
		return err
	}

	s.publisher.Publish(ctx, domain.NewPinnedDeleted(
		authContext.TenantID,
		authContext.UserID,
		item.ID,
		item.EntityType,
		item.EntityID,
	))
	return nil
}

func (s *Service) ListRecent(ctx context.Context) (*RecentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	return cache.GetOrSetJSON(ctx, s.cache, cache.RecentListKey(authContext.TenantID, authContext.UserID), cache.TTLNavigationRecent, func() (*RecentResponse, error) {
		items, err := s.recentStore.List(ctx, authContext.TenantID, authContext.UserID)
		if err != nil {
			return nil, err
		}
		return groupRecentItems(items), nil
	})
}

func (s *Service) TrackVisit(ctx context.Context, input TrackVisitRequest) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	entityType := strings.TrimSpace(input.EntityType)
	entityID := strings.TrimSpace(input.EntityID)
	label := strings.TrimSpace(input.Label)
	path := strings.TrimSpace(input.Path)
	if entityType == "" || entityID == "" || label == "" || path == "" {
		return ErrInvalidPinData
	}

	switch entityType {
	case EntitySettingsPage:
		s.publisher.Publish(ctx, domain.NewSettingsPageOpened(authContext.TenantID, authContext.UserID, entityID, label, path))
	case EntitySmartCollection:
		s.publisher.Publish(ctx, domain.NewSmartCollectionOpened(authContext.TenantID, authContext.UserID, entityID, label, path))
	case EntityIntegration:
		integrationID, parseErr := uuid.Parse(entityID)
		if parseErr != nil {
			return ErrInvalidPinData
		}
		s.publisher.Publish(ctx, domain.NewIntegrationOpened(authContext.TenantID, authContext.UserID, integrationID, label, path))
	default:
		return ErrInvalidPinData
	}
	return nil
}

func groupRecentItems(items []RecentItem) *RecentResponse {
	response := &RecentResponse{
		Documents:        []RecentItem{},
		Folders:          []RecentItem{},
		SmartCollections: []RecentItem{},
		Integrations:     []RecentItem{},
		SettingsPages:    []RecentItem{},
	}
	for _, item := range items {
		switch item.EntityType {
		case EntityDocument:
			response.Documents = append(response.Documents, item)
		case EntityFolder:
			response.Folders = append(response.Folders, item)
		case EntitySmartCollection:
			response.SmartCollections = append(response.SmartCollections, item)
		case EntityIntegration:
			response.Integrations = append(response.Integrations, item)
		case EntitySettingsPage:
			response.SettingsPages = append(response.SettingsPages, item)
		}
	}
	return response
}

func toPinResponse(item PinnedItem) PinResponse {
	return PinResponse{
		ID:         item.ID.String(),
		EntityType: item.EntityType,
		EntityID:   item.EntityID,
		Label:      item.Label,
		Path:       item.Path,
		CreatedAt:  item.CreatedAt,
	}
}
