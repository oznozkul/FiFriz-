package navigation

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"fifriz/internal/cache"

	"github.com/google/uuid"
)

type RecentStore struct {
	cache *cache.Service
}

func NewRecentStore(cacheService *cache.Service) *RecentStore {
	return &RecentStore{cache: cacheService}
}

func (s *RecentStore) List(ctx context.Context, tenantID, userID uuid.UUID) ([]RecentItem, error) {
	if s == nil || s.cache == nil {
		return []RecentItem{}, nil
	}

	key := cache.RecentItemsKey(tenantID, userID)
	raw, ok, err := s.cache.GetRaw(ctx, key)
	if err != nil || !ok {
		return []RecentItem{}, nil
	}

	var items []RecentItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return []RecentItem{}, nil
	}
	return items, nil
}

func (s *RecentStore) Record(ctx context.Context, tenantID, userID uuid.UUID, item RecentItem) error {
	if s == nil || s.cache == nil {
		return nil
	}

	item.EntityType = strings.TrimSpace(item.EntityType)
	item.EntityID = strings.TrimSpace(item.EntityID)
	item.Label = strings.TrimSpace(item.Label)
	item.Path = strings.TrimSpace(item.Path)
	if item.EntityType == "" || item.EntityID == "" || item.Label == "" {
		return nil
	}
	if item.OpenedAt.IsZero() {
		item.OpenedAt = time.Now().UTC()
	}

	key := cache.RecentItemsKey(tenantID, userID)
	items, _ := s.List(ctx, tenantID, userID)

	filtered := make([]RecentItem, 0, len(items)+1)
	for _, existing := range items {
		if existing.EntityType == item.EntityType && existing.EntityID == item.EntityID {
			continue
		}
		filtered = append(filtered, existing)
	}
	filtered = append([]RecentItem{item}, filtered...)
	if len(filtered) > MaxRecentItems {
		filtered = filtered[:MaxRecentItems]
	}

	raw, err := json.Marshal(filtered)
	if err != nil {
		return err
	}
	return s.cache.SetRaw(ctx, key, raw, cache.TTLNavigationRecent)
}

func (s *RecentStore) InvalidateReadCache(ctx context.Context, tenantID, userID uuid.UUID) {
	if s == nil || s.cache == nil {
		return
	}
	_ = s.cache.Delete(ctx, cache.RecentListKey(tenantID, userID))
}
