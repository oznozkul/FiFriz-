package listeners

import (
	"context"

	"fifriz/internal/cache"
	"fifriz/internal/events"
	domain "fifriz/internal/events/events"
)

type CacheListener struct {
	invalidator *cache.Invalidator
}

func NewCacheListener(invalidator *cache.Invalidator) *CacheListener {
	return &CacheListener{invalidator: invalidator}
}

func (l *CacheListener) Register(bus events.EventBus) {
	if l == nil || bus == nil {
		return
	}
	for _, name := range []string{
		domain.DocumentCreatedName,
		domain.DocumentUpdatedName,
		domain.DocumentDeletedName,
		domain.DocumentRestoredName,
		domain.DocumentMovedName,
		domain.DocumentSharedName,
		domain.DocumentVersionCreatedName,
		domain.DocumentVersionRestoredName,
		domain.FavoriteAddedName,
		domain.FavoriteRemovedName,
		domain.TagAssignedName,
		domain.FolderCreatedName,
		domain.FolderUpdatedName,
		domain.FolderRenamedName,
		domain.FolderMovedName,
		domain.FolderLockedName,
		domain.FolderUnlockedName,
		domain.FolderPermissionChangedName,
		domain.FolderDeletedName,
		domain.TagCreatedName,
		domain.TagUpdatedName,
		domain.TagArchivedName,
		domain.TagMergedName,
		domain.TagAliasCreatedName,
		domain.TagAliasDeletedName,
		domain.TagDeletedName,
		domain.WorkflowApprovedName,
		domain.WorkflowRejectedName,
		domain.WorkflowCreatedName,
		domain.DocumentMetadataCompletedName,
		domain.DocumentEntitiesExtractedName,
		domain.DocumentClassificationCompletedName,
		domain.DocumentPipelineCompletedName,
	} {
		bus.Subscribe(name, l.Handle)
	}
}

func (l *CacheListener) Handle(ctx context.Context, event events.Event) {
	if l == nil || l.invalidator == nil || event == nil {
		return
	}
	tenantID := event.TenantID()

	switch event.(type) {
	case domain.DocumentCreated:
		l.invalidator.InvalidateDashboard(ctx, tenantID)
		l.invalidator.InvalidateDocuments(ctx, tenantID)
		l.invalidator.InvalidateCollections(ctx, tenantID)
		l.invalidator.InvalidateGlobalSearch(ctx, tenantID)
		l.invalidator.InvalidateWorkspaceInsights(ctx, tenantID)
	case domain.DocumentUpdated, domain.DocumentRestored, domain.FavoriteAdded, domain.FavoriteRemoved:
		l.invalidator.InvalidateDocuments(ctx, tenantID)
		l.invalidator.InvalidateDashboard(ctx, tenantID)
		l.invalidator.InvalidateGlobalSearch(ctx, tenantID)
		l.invalidator.InvalidateWorkspaceInsights(ctx, tenantID)
	case domain.DocumentDeleted, domain.DocumentMoved, domain.DocumentShared:
		l.invalidator.InvalidateDocuments(ctx, tenantID)
		l.invalidator.InvalidateDashboard(ctx, tenantID)
		l.invalidator.InvalidateCollections(ctx, tenantID)
		l.invalidator.InvalidateWorkspaceInsights(ctx, tenantID)
	case domain.DocumentVersionCreated, domain.DocumentVersionRestored:
		if e, ok := event.(domain.DocumentVersionCreated); ok {
			l.invalidator.InvalidateDocument(ctx, tenantID, e.DocumentID)
		}
		if e, ok := event.(domain.DocumentVersionRestored); ok {
			l.invalidator.InvalidateDocument(ctx, tenantID, e.DocumentID)
		}
		l.invalidator.InvalidateDocuments(ctx, tenantID)
	case domain.FolderCreated, domain.FolderUpdated, domain.FolderRenamed, domain.FolderMoved,
		domain.FolderLocked, domain.FolderUnlocked, domain.FolderPermissionChanged, domain.FolderDeleted:
		l.invalidator.InvalidateFolders(ctx, tenantID)
		l.invalidator.InvalidateDashboard(ctx, tenantID)
		l.invalidator.InvalidateGlobalSearch(ctx, tenantID)
		l.invalidator.InvalidateWorkspaceInsights(ctx, tenantID)
	case domain.TagCreated, domain.TagUpdated, domain.TagArchived, domain.TagMerged,
		domain.TagAliasCreated, domain.TagAliasDeleted, domain.TagDeleted:
		l.invalidator.InvalidateTags(ctx, tenantID)
		l.invalidator.InvalidateGlobalSearch(ctx, tenantID)
		l.invalidator.InvalidateWorkspaceInsights(ctx, tenantID)
	case domain.TagAssigned:
		l.invalidator.InvalidateTags(ctx, tenantID)
		l.invalidator.InvalidateDocuments(ctx, tenantID)
		l.invalidator.InvalidateDashboard(ctx, tenantID)
		l.invalidator.InvalidateCollections(ctx, tenantID)
		l.invalidator.InvalidateGlobalSearch(ctx, tenantID)
		l.invalidator.InvalidateWorkspaceInsights(ctx, tenantID)
	case domain.DocumentMetadataCompleted, domain.DocumentEntitiesExtracted, domain.DocumentClassificationCompleted, domain.DocumentPipelineCompleted:
		l.invalidator.InvalidateDocuments(ctx, tenantID)
		l.invalidator.InvalidateGlobalSearch(ctx, tenantID)
		l.invalidator.InvalidateWorkspaceInsights(ctx, tenantID)
	case domain.WorkflowCreated, domain.WorkflowApproved, domain.WorkflowRejected:
		l.invalidator.InvalidateDashboard(ctx, tenantID)
		l.invalidator.InvalidateWorkspaceInsights(ctx, tenantID)
		if e, ok := event.(domain.WorkflowApproved); ok {
			l.invalidator.InvalidateDocument(ctx, tenantID, e.DocumentID)
		}
		if e, ok := event.(domain.WorkflowRejected); ok {
			l.invalidator.InvalidateDocument(ctx, tenantID, e.DocumentID)
		}
		if e, ok := event.(domain.WorkflowCreated); ok {
			l.invalidator.InvalidateDocument(ctx, tenantID, e.DocumentID)
		}
	}
}
