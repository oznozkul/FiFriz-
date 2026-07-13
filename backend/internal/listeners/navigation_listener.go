package listeners

import (
	"context"
	"time"

	"fifriz/internal/cache"
	"fifriz/internal/events"
	domain "fifriz/internal/events/events"
	"fifriz/internal/navigation"
)

type RecentListener struct {
	store       *navigation.RecentStore
	invalidator *cache.Invalidator
}

func NewRecentListener(store *navigation.RecentStore, invalidator *cache.Invalidator) *RecentListener {
	return &RecentListener{store: store, invalidator: invalidator}
}

func (l *RecentListener) Register(bus events.EventBus) {
	if l == nil || bus == nil {
		return
	}
	for _, name := range []string{
		domain.DocumentOpenedName,
		domain.FolderOpenedName,
		domain.SmartCollectionOpenedName,
		domain.IntegrationOpenedName,
		domain.SettingsPageOpenedName,
	} {
		bus.Subscribe(name, l.Handle)
	}
}

func (l *RecentListener) Handle(ctx context.Context, event events.Event) {
	if l == nil || l.store == nil || event == nil {
		return
	}

	actorID := event.ActorUserID()
	if actorID == nil {
		return
	}

	tenantID := event.TenantID()
	userID := *actorID
	now := time.Now().UTC()

	var item navigation.RecentItem

	switch e := event.(type) {
	case domain.DocumentOpened:
		item = navigation.RecentItem{
			EntityType: navigation.EntityDocument,
			EntityID:   e.DocumentID.String(),
			Label:      e.OriginalFileName,
			Path:       e.Path,
			OpenedAt:   now,
		}
	case domain.FolderOpened:
		item = navigation.RecentItem{
			EntityType: navigation.EntityFolder,
			EntityID:   e.FolderID.String(),
			Label:      e.Name,
			Path:       e.Path,
			OpenedAt:   now,
		}
	case domain.SmartCollectionOpened:
		item = navigation.RecentItem{
			EntityType: navigation.EntitySmartCollection,
			EntityID:   e.CollectionKey,
			Label:      e.CollectionTitle,
			Path:       e.Path,
			OpenedAt:   now,
		}
	case domain.IntegrationOpened:
		item = navigation.RecentItem{
			EntityType: navigation.EntityIntegration,
			EntityID:   e.IntegrationID.String(),
			Label:      e.DisplayName,
			Path:       e.Path,
			OpenedAt:   now,
		}
	case domain.SettingsPageOpened:
		item = navigation.RecentItem{
			EntityType: navigation.EntitySettingsPage,
			EntityID:   e.PageKey,
			Label:      e.Label,
			Path:       e.Path,
			OpenedAt:   now,
		}
	default:
		return
	}

	_ = l.store.Record(ctx, tenantID, userID, item)
	if l.invalidator != nil {
		l.invalidator.InvalidateNavigationRecentList(ctx, tenantID, userID)
	}
}

type NavigationCacheListener struct {
	invalidator *cache.Invalidator
}

func NewNavigationCacheListener(invalidator *cache.Invalidator) *NavigationCacheListener {
	return &NavigationCacheListener{invalidator: invalidator}
}

func (l *NavigationCacheListener) Register(bus events.EventBus) {
	if l == nil || bus == nil {
		return
	}
	for _, name := range []string{
		domain.PinnedCreatedName,
		domain.PinnedDeletedName,
	} {
		bus.Subscribe(name, l.Handle)
	}
}

func (l *NavigationCacheListener) Handle(ctx context.Context, event events.Event) {
	if l == nil || l.invalidator == nil || event == nil {
		return
	}
	actorID := event.ActorUserID()
	if actorID == nil {
		return
	}
	l.invalidator.InvalidateNavigationPins(ctx, event.TenantID(), *actorID)
}
