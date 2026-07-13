package tags

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TagStatsResponse struct {
	TagID            uuid.UUID          `json:"tagId"`
	DocumentCount    int64              `json:"documentCount"`
	LastUsedAt       *time.Time         `json:"lastUsedAt"`
	CreatedAt        time.Time          `json:"createdAt"`
	MostUsedFolders  []TagFolderUsage   `json:"mostUsedFolders"`
	RelatedTags      []TagBrief         `json:"relatedTags"`
}

type TagFolderUsage struct {
	FolderID   *uuid.UUID `json:"folderId"`
	FolderName string     `json:"folderName"`
	Count      int64      `json:"count"`
}

type TagSearchResult struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	LifecycleStatus string    `json:"lifecycleStatus"`
	Color           *string   `json:"color"`
	Icon            *string   `json:"icon"`
	DocumentCount   int64     `json:"documentCount"`
	AliasMatch      bool      `json:"aliasMatch"`
}

func (r *Repository) CountDocumentsByTag(ctx context.Context, tenantID, tagID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&DocumentTag{}).
		Joins("INNER JOIN documents d ON d.id = document_tags.document_id AND d.tenant_id = document_tags.tenant_id AND d.deleted_at IS NULL").
		Where("document_tags.tenant_id = ? AND document_tags.tag_id = ?", tenantID, tagID).
		Count(&count).Error
	return count, err
}

func (r *Repository) LastUsedAt(ctx context.Context, tenantID, tagID uuid.UUID) (*time.Time, error) {
	var lastUsed *time.Time
	err := r.db.WithContext(ctx).Model(&DocumentTag{}).
		Select("MAX(created_at)").
		Where("tenant_id = ? AND tag_id = ?", tenantID, tagID).
		Scan(&lastUsed).Error
	return lastUsed, err
}

func (r *Repository) TagUsageCounts(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]int64, error) {
	type row struct {
		TagID uuid.UUID `gorm:"column:tag_id"`
		Count int64     `gorm:"column:count"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Table("document_tags dt").
		Select("dt.tag_id, COUNT(*) AS count").
		Joins("INNER JOIN documents d ON d.id = dt.document_id AND d.tenant_id = dt.tenant_id AND d.deleted_at IS NULL").
		Where("dt.tenant_id = ?", tenantID).
		Group("dt.tag_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]int64, len(rows))
	for _, item := range rows {
		result[item.TagID] = item.Count
	}
	return result, nil
}

func (r *Repository) MostUsedFoldersForTag(ctx context.Context, tenantID, tagID uuid.UUID, limit int) ([]TagFolderUsage, error) {
	if limit <= 0 {
		limit = 5
	}
	var rows []TagFolderUsage
	err := r.db.WithContext(ctx).Table("document_tags dt").
		Select(`
			d.folder_id AS folder_id,
			COALESCE(f.name, 'Kök') AS folder_name,
			COUNT(*) AS count
		`).
		Joins("INNER JOIN documents d ON d.id = dt.document_id AND d.tenant_id = dt.tenant_id AND d.deleted_at IS NULL").
		Joins("LEFT JOIN folders f ON f.id = d.folder_id AND f.tenant_id = d.tenant_id AND f.deleted_at IS NULL").
		Where("dt.tenant_id = ? AND dt.tag_id = ?", tenantID, tagID).
		Group("d.folder_id, f.name").
		Order("count DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) RelatedTags(ctx context.Context, tenantID, tagID uuid.UUID, limit int) ([]TagBrief, error) {
	if limit <= 0 {
		limit = 5
	}
	var rows []TagBrief
	err := r.db.WithContext(ctx).Table("document_tags dt1").
		Select("DISTINCT t.id, t.name, t.color").
		Joins("INNER JOIN document_tags dt2 ON dt2.document_id = dt1.document_id AND dt2.tenant_id = dt1.tenant_id AND dt2.tag_id <> dt1.tag_id").
		Joins("INNER JOIN tags t ON t.id = dt2.tag_id AND t.tenant_id = dt2.tenant_id AND t.deleted_at IS NULL").
		Where("dt1.tenant_id = ? AND dt1.tag_id = ? AND t.status = ?", tenantID, tagID, StatusActive).
		Order("t.name ASC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) MergeDocumentTags(ctx context.Context, tenantID, sourceTagID, targetTagID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sourceLinks []DocumentTag
		if err := tx.Where("tenant_id = ? AND tag_id = ?", tenantID, sourceTagID).Find(&sourceLinks).Error; err != nil {
			return err
		}

		for _, link := range sourceLinks {
			var existing int64
			if err := tx.Model(&DocumentTag{}).
				Where("tenant_id = ? AND document_id = ? AND tag_id = ?", tenantID, link.DocumentID, targetTagID).
				Count(&existing).Error; err != nil {
				return err
			}
			if existing > 0 {
				if err := tx.Delete(&DocumentTag{}, "tenant_id = ? AND document_id = ? AND tag_id = ?", tenantID, link.DocumentID, sourceTagID).Error; err != nil {
					return err
				}
				continue
			}
			if err := tx.Model(&DocumentTag{}).
				Where("id = ?", link.ID).
				Update("tag_id", targetTagID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) ListAliases(ctx context.Context, tenantID, tagID uuid.UUID) ([]TagAlias, error) {
	var items []TagAlias
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND tag_id = ?", tenantID, tagID).
		Order("alias ASC").
		Find(&items).Error
	return items, err
}

func (r *Repository) FindTagByAlias(ctx context.Context, tenantID uuid.UUID, alias string) (*Tag, error) {
	alias = strings.TrimSpace(alias)
	var row TagAlias
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND LOWER(alias) = LOWER(?)", tenantID, alias).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, tenantID, row.TagID)
}

func (r *Repository) CreateAlias(ctx context.Context, alias *TagAlias) error {
	return r.db.WithContext(ctx).Create(alias).Error
}

func (r *Repository) DeleteAlias(ctx context.Context, tenantID, aliasID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, aliasID).
		Delete(&TagAlias{}).Error
}

func (r *Repository) AliasExists(ctx context.Context, tenantID uuid.UUID, alias string, exceptID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&TagAlias{}).
		Where("tenant_id = ? AND LOWER(alias) = LOWER(?)", tenantID, alias)
	if exceptID != nil {
		query = query.Where("id <> ?", *exceptID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) ListAllAliases(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID][]string, error) {
	var items []TagAlias
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&items).Error; err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID][]string)
	for _, item := range items {
		result[item.TagID] = append(result[item.TagID], item.Alias)
	}
	return result, nil
}

func (r *Repository) ListKnownTagTerms(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	tags, err := r.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	aliasesByTag, err := r.ListAllAliases(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	terms := make([]string, 0)
	seen := make(map[string]struct{})
	for _, tag := range tags {
		key := strings.ToLower(tag.Name)
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			terms = append(terms, tag.Name)
		}
		for _, alias := range aliasesByTag[tag.ID] {
			aliasKey := strings.ToLower(alias)
			if _, ok := seen[aliasKey]; ok {
				continue
			}
			seen[aliasKey] = struct{}{}
			terms = append(terms, alias)
		}
	}
	return terms, nil
}

func (r *Repository) SearchTags(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]TagSearchResult, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return []TagSearchResult{}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	items, err := r.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	counts, _ := r.TagUsageCounts(ctx, tenantID)
	aliasesByTag, _ := r.ListAllAliases(ctx, tenantID)

	results := make([]TagSearchResult, 0)
	for _, tag := range items {
		nameMatch := strings.Contains(strings.ToLower(tag.Name), query)
		statusMatch := strings.Contains(strings.ToLower(tag.Status), query)
		colorMatch := tag.Color != nil && strings.Contains(strings.ToLower(*tag.Color), query)
		aliasMatch := false
		for _, alias := range aliasesByTag[tag.ID] {
			if strings.Contains(strings.ToLower(alias), query) {
				aliasMatch = true
				break
			}
		}
		countMatch := false
		if count := counts[tag.ID]; count > 0 {
			countMatch = strings.Contains(strings.TrimSpace(query), "") // usage filter via numeric query optional
			_ = countMatch
		}

		if !nameMatch && !statusMatch && !colorMatch && !aliasMatch {
			continue
		}

		results = append(results, TagSearchResult{
			ID:              tag.ID,
			Name:            tag.Name,
			Status:          tag.Status,
			LifecycleStatus: EffectiveLifecycleStatus(&tag),
			Color:           tag.Color,
			Icon:            tag.Icon,
			DocumentCount:   counts[tag.ID],
			AliasMatch:      aliasMatch,
		})
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}
