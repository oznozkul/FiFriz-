package search

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FolderDocumentCounts(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]int, error) {
	type row struct {
		FolderID uuid.UUID
		Count    int
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("documents").
		Select("folder_id AS folder_id, COUNT(*) AS count").
		Where("tenant_id = ? AND deleted_at IS NULL AND folder_id IS NOT NULL", tenantID).
		Group("folder_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]int, len(rows))
	for _, item := range rows {
		result[item.FolderID] = item.Count
	}
	return result, nil
}

func (r *Repository) TagUsageCounts(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]int, error) {
	type row struct {
		TagID uuid.UUID
		Count int
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("document_tags dt").
		Select("dt.tag_id AS tag_id, COUNT(*) AS count").
		Joins("INNER JOIN documents d ON d.id = dt.document_id AND d.tenant_id = dt.tenant_id AND d.deleted_at IS NULL").
		Where("dt.tenant_id = ?", tenantID).
		Group("dt.tag_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]int, len(rows))
	for _, item := range rows {
		result[item.TagID] = item.Count
	}
	return result, nil
}

type documentFolderMeta struct {
	FolderID   *uuid.UUID
	UpdatedAt  time.Time
	FolderName string
}

func (r *Repository) DocumentFolderMeta(ctx context.Context, tenantID uuid.UUID, documentIDs []uuid.UUID) (map[uuid.UUID]documentFolderMeta, error) {
	result := make(map[uuid.UUID]documentFolderMeta)
	if len(documentIDs) == 0 {
		return result, nil
	}
	var rows []struct {
		ID         uuid.UUID
		FolderID   *uuid.UUID
		UpdatedAt  time.Time
		FolderName *string
	}
	err := r.db.WithContext(ctx).
		Table("documents d").
		Select("d.id, d.folder_id, d.updated_at, f.name AS folder_name").
		Joins("LEFT JOIN folders f ON f.id = d.folder_id AND f.tenant_id = d.tenant_id").
		Where("d.tenant_id = ? AND d.id IN ?", tenantID, documentIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		folderName := ""
		if row.FolderName != nil {
			folderName = *row.FolderName
		}
		result[row.ID] = documentFolderMeta{
			FolderID:   row.FolderID,
			UpdatedAt:  row.UpdatedAt,
			FolderName: folderName,
		}
	}
	return result, nil
}
