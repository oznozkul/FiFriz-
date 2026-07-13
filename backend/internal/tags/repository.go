package tags

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, tenantID uuid.UUID) ([]Tag, error) {
	var items []Tag
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("name ASC, created_at ASC").
		Find(&items).Error
	return items, err
}

func (r *Repository) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*Tag, error) {
	var tag Tag
	err := r.db.WithContext(ctx).First(&tag, "id = ? AND tenant_id = ?", id, tenantID).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *Repository) Create(ctx context.Context, tag *Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *Repository) Update(ctx context.Context, tag *Tag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

func (r *Repository) Delete(ctx context.Context, tag *Tag) error {
	return r.db.WithContext(ctx).Delete(tag).Error
}

func (r *Repository) ExistsByName(ctx context.Context, tenantID uuid.UUID, name string, exceptID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&Tag{}).
		Where("tenant_id = ? AND LOWER(name) = LOWER(?)", tenantID, name)
	if exceptID != nil {
		query = query.Where("id <> ?", *exceptID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

type documentTagRow struct {
	DocumentID uuid.UUID `gorm:"column:document_id"`
	ID         uuid.UUID `gorm:"column:id"`
	Name       string    `gorm:"column:name"`
	Color      *string   `gorm:"column:color"`
	Icon       *string   `gorm:"column:icon"`
}

func (r *Repository) ListTagsByDocumentIDs(ctx context.Context, tenantID uuid.UUID, documentIDs []uuid.UUID) (map[uuid.UUID][]TagBrief, error) {
	result := make(map[uuid.UUID][]TagBrief)
	if len(documentIDs) == 0 {
		return result, nil
	}

	var rows []documentTagRow
	err := r.db.WithContext(ctx).
		Table("document_tags dt").
		Select("dt.document_id, t.id, t.name, t.color, t.icon").
		Joins("INNER JOIN tags t ON t.id = dt.tag_id AND t.tenant_id = dt.tenant_id AND t.deleted_at IS NULL").
		Where("dt.tenant_id = ? AND dt.document_id IN ?", tenantID, documentIDs).
		Order("t.name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.DocumentID] = append(result[row.DocumentID], TagBrief{
			ID:    row.ID,
			Name:  row.Name,
			Color: row.Color,
			Icon:  row.Icon,
		})
	}

	return result, nil
}

func (r *Repository) ReplaceDocumentTags(ctx context.Context, tenantID uuid.UUID, documentID uuid.UUID, tagIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ? AND document_id = ?", tenantID, documentID).Delete(&DocumentTag{}).Error; err != nil {
			return err
		}

		if len(tagIDs) == 0 {
			return nil
		}

		entries := make([]DocumentTag, 0, len(tagIDs))
		for _, tagID := range tagIDs {
			entries = append(entries, DocumentTag{
				TenantID:   tenantID,
				DocumentID: documentID,
				TagID:      tagID,
			})
		}

		return tx.Create(&entries).Error
	})
}

func (r *Repository) ValidateTagIDs(ctx context.Context, tenantID uuid.UUID, tagIDs []uuid.UUID) error {
	if len(tagIDs) == 0 {
		return nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&Tag{}).
		Where("tenant_id = ? AND id IN ?", tenantID, tagIDs).
		Count(&count).Error; err != nil {
		return err
	}
	if int(count) != len(tagIDs) {
		return ErrTagNotFound
	}
	return nil
}

func (r *Repository) DeleteByTagID(ctx context.Context, tenantID uuid.UUID, tagID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND tag_id = ?", tenantID, tagID).
		Delete(&DocumentTag{}).Error
}

func (r *Repository) FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*Tag, error) {
	name = normalizeTagName(name)
	if name == "" {
		return nil, ErrInvalidTagData
	}

	var tag Tag
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND LOWER(name) = LOWER(?)", tenantID, name).
		First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *Repository) FindOrCreateTag(ctx context.Context, tenantID, userID uuid.UUID, name string) (*Tag, error) {
	name, err := ValidateTagName(name)
	if err != nil {
		return nil, err
	}

	existing, err := r.FindByName(ctx, tenantID, name)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	tag := &Tag{
		TenantID:        tenantID,
		Name:            name,
		CreatedByUserID: userID,
	}
	if createErr := r.Create(ctx, tag); createErr != nil {
		if retry, retryErr := r.FindByName(ctx, tenantID, name); retryErr == nil {
			return retry, nil
		}
		return nil, createErr
	}
	return tag, nil
}

func (r *Repository) AppendDocumentTag(ctx context.Context, tenantID, documentID, tagID uuid.UUID) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&DocumentTag{}).
		Where("tenant_id = ? AND document_id = ? AND tag_id = ?", tenantID, documentID, tagID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return r.db.WithContext(ctx).Create(&DocumentTag{
		TenantID:   tenantID,
		DocumentID: documentID,
		TagID:      tagID,
	}).Error
}

func (r *Repository) ListTagNames(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).Model(&Tag{}).
		Where("tenant_id = ?", tenantID).
		Order("name ASC").
		Pluck("name", &names).Error
	return names, err
}

func normalizeTagName(value string) string {
	return strings.TrimSpace(value)
}
