package cache

import (
	"fmt"

	"github.com/google/uuid"
)

func TenantModuleResource(tenantID uuid.UUID, module, resource string) string {
	return fmt.Sprintf("tenant:%s:module:%s:resource:%s", tenantID.String(), module, resource)
}

func TenantUserResource(tenantID, userID uuid.UUID, module, resource string) string {
	return fmt.Sprintf("tenant:%s:user:%s:module:%s:resource:%s", tenantID.String(), userID.String(), module, resource)
}

func TenantPattern(tenantID uuid.UUID) string {
	return fmt.Sprintf("tenant:%s:*", tenantID.String())
}

func TenantModulePattern(tenantID uuid.UUID, module string) string {
	return fmt.Sprintf("tenant:%s:module:%s:*", tenantID.String(), module)
}

func DashboardKey(tenantID, userID uuid.UUID) string {
	return TenantUserResource(tenantID, userID, "dashboard", "summary")
}

func FolderTreeKey(tenantID uuid.UUID) string {
	return TenantModuleResource(tenantID, "folders", "tree")
}

func DocumentDetailKey(tenantID, documentID uuid.UUID) string {
	return TenantModuleResource(tenantID, "documents", "detail:"+documentID.String())
}

func DocumentWorkspaceKey(tenantID, documentID uuid.UUID) string {
	return TenantModuleResource(tenantID, "documents", "workspace:"+documentID.String())
}

func DocumentListKey(tenantID, userID uuid.UUID, suffix string) string {
	return TenantUserResource(tenantID, userID, "documents", "list:"+suffix)
}

func TagsKey(tenantID uuid.UUID) string {
	return TenantModuleResource(tenantID, "tags", "list")
}

func NotificationsKey(tenantID, userID uuid.UUID) string {
	return TenantUserResource(tenantID, userID, "notifications", "list")
}

func SmartCollectionsKey(tenantID, userID uuid.UUID) string {
	return TenantUserResource(tenantID, userID, "collections", "list")
}

func GlobalSearchKey(tenantID, userID uuid.UUID, queryHash string) string {
	return TenantUserResource(tenantID, userID, "search", "global:"+queryHash)
}

func HybridSearchKey(tenantID, userID uuid.UUID, queryHash string) string {
	return TenantUserResource(tenantID, userID, "search", "hybrid:"+queryHash)
}

func WorkspaceInsightsKey(tenantID, userID uuid.UUID) string {
	return TenantUserResource(tenantID, userID, "ai", "workspace-insights")
}

func PinsListKey(tenantID, userID uuid.UUID) string {
	return TenantUserResource(tenantID, userID, "navigation", "pins")
}

func RecentItemsKey(tenantID, userID uuid.UUID) string {
	return TenantUserResource(tenantID, userID, "navigation", "recent:items")
}

func RecentListKey(tenantID, userID uuid.UUID) string {
	return TenantUserResource(tenantID, userID, "navigation", "recent:list")
}
