package folders

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FolderMetadata struct {
	DocumentCount   int64      `json:"documentCount"`
	SubFolderCount  int64      `json:"subFolderCount"`
	TotalSize       int64      `json:"totalSize"`
	StorageUsage    int64      `json:"storageUsage"`
	LastActivityAt  *time.Time `json:"lastActivityAt"`
	LastModifiedBy  *uuid.UUID `json:"lastModifiedByUserId"`
	AIIndexed       bool       `json:"aiIndexed"`
}

type FolderSearchResult struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ParentID        *uuid.UUID `json:"parentId"`
	Status          string    `json:"status"`
	LifecycleStatus string    `json:"lifecycleStatus"`
	DocumentCount   int64     `json:"documentCount"`
	SubFolderCount  int64     `json:"subFolderCount"`
	AIClassMatch    bool      `json:"aiClassMatch"`
}

func (r *Repository) ListIncludingDeleted(ctx context.Context, tenantID uuid.UUID) ([]Folder, error) {
	var folders []Folder
	err := r.db.WithContext(ctx).Unscoped().
		Where("tenant_id = ?", tenantID).
		Order("parent_id NULLS FIRST, name ASC, created_at ASC").
		Find(&folders).Error
	return folders, err
}

func (r *Repository) GetByIDUnscoped(ctx context.Context, tenantID, id uuid.UUID) (*Folder, error) {
	var folder Folder
	err := r.db.WithContext(ctx).Unscoped().First(&folder, "id = ? AND tenant_id = ?", id, tenantID).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *Repository) SoftDeleteIDs(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id IN ?", tenantID, ids).
		Delete(&Folder{}).Error
}

func (r *Repository) RestoreIDs(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Unscoped().
		Model(&Folder{}).
		Where("tenant_id = ? AND id IN ?", tenantID, ids).
		Update("deleted_at", nil).Error
}

func (r *Repository) UpdateParent(ctx context.Context, folder *Folder) error {
	return r.db.WithContext(ctx).Model(folder).Updates(map[string]interface{}{
		"parent_id":               folder.ParentID,
		"last_modified_by_user_id": folder.LastModifiedByUserID,
		"last_activity_at":       folder.LastActivityAt,
		"updated_at":              folder.UpdatedAt,
	}).Error
}

func (r *Repository) CountChildren(ctx context.Context, tenantID, folderID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Folder{}).
		Where("tenant_id = ? AND parent_id = ?", tenantID, folderID).
		Count(&count).Error
	return count, err
}

func (r *Repository) GetFolderMetadata(ctx context.Context, tenantID uuid.UUID, folderIDs []uuid.UUID) (*FolderMetadata, error) {
	meta := &FolderMetadata{}
	if len(folderIDs) == 0 {
		return meta, nil
	}

	if err := r.db.WithContext(ctx).Table("documents").
		Where("tenant_id = ? AND folder_id IN ? AND deleted_at IS NULL", tenantID, folderIDs).
		Count(&meta.DocumentCount).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Model(&Folder{}).
		Where("tenant_id = ? AND parent_id IN ?", tenantID, folderIDs).
		Count(&meta.SubFolderCount).Error; err != nil {
		return nil, err
	}

	var totalSize *int64
	if err := r.db.WithContext(ctx).Table("documents").
		Select("COALESCE(SUM(file_size), 0)").
		Where("tenant_id = ? AND folder_id IN ? AND deleted_at IS NULL", tenantID, folderIDs).
		Scan(&totalSize).Error; err != nil {
		return nil, err
	}
	if totalSize != nil {
		meta.TotalSize = *totalSize
		meta.StorageUsage = *totalSize
	}

	var lastDocActivity *time.Time
	if err := r.db.WithContext(ctx).Table("documents").
		Select("MAX(updated_at)").
		Where("tenant_id = ? AND folder_id IN ? AND deleted_at IS NULL", tenantID, folderIDs).
		Scan(&lastDocActivity).Error; err != nil {
		return nil, err
	}
	meta.LastActivityAt = lastDocActivity

	var indexedCount int64
	if err := r.db.WithContext(ctx).Table("documents AS d").
		Joins("INNER JOIN document_chunks AS c ON c.document_id = d.id AND c.tenant_id = d.tenant_id").
		Where("d.tenant_id = ? AND d.folder_id IN ? AND d.deleted_at IS NULL", tenantID, folderIDs).
		Distinct("d.id").
		Count(&indexedCount).Error; err != nil {
		return nil, err
	}
	meta.AIIndexed = meta.DocumentCount > 0 && indexedCount >= meta.DocumentCount

	return meta, nil
}

func (r *Repository) SearchFolders(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]FolderSearchResult, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return []FolderSearchResult{}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	allFolders, err := r.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	type docAgg struct {
		FolderID      uuid.UUID `gorm:"column:folder_id"`
		DocumentCount int64     `gorm:"column:document_count"`
	}
	var aggregates []docAgg
	if err := r.db.WithContext(ctx).Table("documents").
		Select("folder_id, COUNT(*) AS document_count").
		Where("tenant_id = ? AND deleted_at IS NULL AND folder_id IS NOT NULL", tenantID).
		Group("folder_id").
		Scan(&aggregates).Error; err != nil {
		return nil, err
	}

	aggByFolder := make(map[uuid.UUID]docAgg, len(aggregates))
	for _, item := range aggregates {
		aggByFolder[item.FolderID] = item
	}

	aiMatchByFolder := make(map[uuid.UUID]bool)
	var aiRows []struct {
		FolderID uuid.UUID `gorm:"column:folder_id"`
	}
	if err := r.db.WithContext(ctx).Table("documents").
		Select("DISTINCT folder_id").
		Where("tenant_id = ? AND deleted_at IS NULL AND folder_id IS NOT NULL AND LOWER(COALESCE(ai_document_type, '')) LIKE ?", tenantID, "%"+query+"%").
		Scan(&aiRows).Error; err != nil {
		return nil, err
	}
	for _, row := range aiRows {
		aiMatchByFolder[row.FolderID] = true
	}

	childCounts := make(map[uuid.UUID]int64)
	for _, folder := range allFolders {
		if folder.ParentID != nil {
			childCounts[*folder.ParentID]++
		}
	}

	results := make([]FolderSearchResult, 0)
	for _, folder := range allFolders {
		nameMatch := strings.Contains(strings.ToLower(folder.Name), query)
		descMatch := strings.Contains(strings.ToLower(folder.Description), query)
		agg := aggByFolder[folder.ID]
		aiMatch := aiMatchByFolder[folder.ID]
		metaMatch := strings.Contains(strings.TrimSpace(folder.Description), query)

		if !nameMatch && !descMatch && !aiMatch && !metaMatch {
			continue
		}

		results = append(results, FolderSearchResult{
			ID:              folder.ID,
			Name:            folder.Name,
			Description:     folder.Description,
			ParentID:        folder.ParentID,
			Status:          folder.Status,
			LifecycleStatus: EffectiveLifecycleStatus(&folder),
			DocumentCount:   agg.DocumentCount,
			SubFolderCount:  childCounts[folder.ID],
			AIClassMatch:    aiMatch,
		})
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (r *Repository) ListFolderPermissions(ctx context.Context, tenantID, folderID uuid.UUID) ([]FolderPermission, error) {
	var items []FolderPermission
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND folder_id = ?", tenantID, folderID).
		Order("created_at ASC").
		Find(&items).Error
	return items, err
}

func (r *Repository) ReplaceFolderPermissions(ctx context.Context, tenantID, folderID uuid.UUID, permissions []FolderPermission) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ? AND folder_id = ?", tenantID, folderID).Delete(&FolderPermission{}).Error; err != nil {
			return err
		}
		if len(permissions) == 0 {
			return nil
		}
		return tx.Create(&permissions).Error
	})
}

func (r *Repository) SetInheritPermissions(ctx context.Context, folder *Folder) error {
	return r.db.WithContext(ctx).Model(folder).Updates(map[string]interface{}{
		"inherit_permissions":      folder.InheritPermissions,
		"last_modified_by_user_id": folder.LastModifiedByUserID,
		"updated_at":               folder.UpdatedAt,
	}).Error
}
