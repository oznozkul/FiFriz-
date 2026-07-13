package cache

import (
	"context"

	"github.com/google/uuid"
)

type Invalidator struct {
	service *Service
}

func NewInvalidator(service *Service) *Invalidator {
	return &Invalidator{service: service}
}

func (i *Invalidator) InvalidateDashboard(ctx context.Context, tenantID uuid.UUID) {
	_ = i.service.DeleteByPattern(ctx, TenantModulePattern(tenantID, "dashboard"))
}

func (i *Invalidator) InvalidateDocuments(ctx context.Context, tenantID uuid.UUID) {
	_ = i.service.DeleteByPattern(ctx, TenantModulePattern(tenantID, "documents"))
}

func (i *Invalidator) InvalidateDocument(ctx context.Context, tenantID, documentID uuid.UUID) {
	_ = i.service.Delete(ctx, DocumentDetailKey(tenantID, documentID))
	_ = i.service.Delete(ctx, DocumentWorkspaceKey(tenantID, documentID))
	_ = i.service.DeleteByPattern(ctx, TenantModulePattern(tenantID, "documents"))
}

func (i *Invalidator) InvalidateFolders(ctx context.Context, tenantID uuid.UUID) {
	_ = i.service.Delete(ctx, FolderTreeKey(tenantID))
	_ = i.service.DeleteByPattern(ctx, TenantModulePattern(tenantID, "folders"))
}

func (i *Invalidator) InvalidateTags(ctx context.Context, tenantID uuid.UUID) {
	_ = i.service.Delete(ctx, TagsKey(tenantID))
}

func (i *Invalidator) InvalidateCollections(ctx context.Context, tenantID uuid.UUID) {
	_ = i.service.DeleteByPattern(ctx, TenantModulePattern(tenantID, "collections"))
}

func (i *Invalidator) InvalidateNotifications(ctx context.Context, tenantID, userID uuid.UUID) {
	_ = i.service.Delete(ctx, NotificationsKey(tenantID, userID))
}

func (i *Invalidator) InvalidateGlobalSearch(ctx context.Context, tenantID uuid.UUID) {
	_ = i.service.DeleteByPattern(ctx, TenantModulePattern(tenantID, "search"))
}

func (i *Invalidator) InvalidateWorkspaceInsights(ctx context.Context, tenantID uuid.UUID) {
	_ = i.service.DeleteByPattern(ctx, TenantModulePattern(tenantID, "ai"))
}

func (i *Invalidator) InvalidateNavigationPins(ctx context.Context, tenantID, userID uuid.UUID) {
	_ = i.service.Delete(ctx, PinsListKey(tenantID, userID))
}

func (i *Invalidator) InvalidateNavigationRecentList(ctx context.Context, tenantID, userID uuid.UUID) {
	_ = i.service.Delete(ctx, RecentListKey(tenantID, userID))
}
