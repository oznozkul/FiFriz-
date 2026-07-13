package listeners

import (
	"fifriz/internal/audit"
	"fifriz/internal/cache"
	"fifriz/internal/events"
	"fifriz/internal/extensions"
	"fifriz/internal/jobs"
	domain "fifriz/internal/events/events"
	"fifriz/internal/navigation"
	"fifriz/internal/notifications"
)

type Dependencies struct {
	Audit         audit.Recorder
	Notifications notifications.Creator
	Cache         *cache.Service
	Queue         jobs.Enqueuer
	Extensions    *extensions.Registry
}

type Runtime struct {
	Bus         events.EventBus
	Publisher   events.Publisher
	Cache       *cache.Service
	Invalidator *cache.Invalidator
}

func Bootstrap(deps Dependencies) *Runtime {
	bus := events.NewBus()
	publisher := events.NewPublisher(bus)
	invalidator := cache.NewInvalidator(deps.Cache)

	NewAuditListener(deps.Audit).Register(bus)
	NewNotificationListener(deps.Notifications).Register(bus)
	NewCacheListener(invalidator).Register(bus)
	NewSearchListener().Register(bus)
	NewCompressionListener(deps.Queue).Register(bus)
	NewAILogicListener().Register(bus)
	NewWebhookListener().Register(bus)

	if deps.Extensions != nil {
		extensions.NewExtensionHandler(deps.Extensions).Register(bus, allDomainEventNames()...)
	}

	return &Runtime{
		Bus:         bus,
		Publisher:   publisher,
		Cache:       deps.Cache,
		Invalidator: invalidator,
	}
}

func BootstrapNavigation(bus events.EventBus, invalidator *cache.Invalidator, recentStore *navigation.RecentStore) {
	NewRecentListener(recentStore, invalidator).Register(bus)
	NewNavigationCacheListener(invalidator).Register(bus)
}

func allDomainEventNames() []string {
	return []string{
		domain.DocumentCreatedName,
		domain.DocumentUpdatedName,
		domain.DocumentDeletedName,
		domain.DocumentRestoredName,
		domain.DocumentMovedName,
		domain.DocumentSharedName,
		domain.DocumentVersionCreatedName,
		domain.DocumentVersionRestoredName,
		domain.DocumentSearchedName,
		domain.FavoriteAddedName,
		domain.FavoriteRemovedName,
		domain.FolderCreatedName,
		domain.FolderUpdatedName,
		domain.FolderDeletedName,
		domain.TagCreatedName,
		domain.TagDeletedName,
		domain.TagAssignedName,
		domain.CommentCreatedName,
		domain.CommentDeletedName,
		domain.WorkflowCreatedName,
		domain.WorkflowApprovedName,
		domain.WorkflowRejectedName,
		domain.NotificationCreatedName,
		domain.IntegrationSyncCompletedName,
		domain.IntegrationSyncFailedName,
		domain.DocumentProcessedName,
		domain.DocumentProcessingFailedName,
		domain.DocumentOpenedName,
		domain.FolderOpenedName,
		domain.SmartCollectionOpenedName,
		domain.IntegrationOpenedName,
		domain.SettingsPageOpenedName,
		domain.PinnedCreatedName,
		domain.PinnedDeletedName,
	}
}
