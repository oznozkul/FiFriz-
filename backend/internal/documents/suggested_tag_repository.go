package documents

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrSuggestedTagNotFound          = errors.New("suggested tag not found")
	ErrSuggestedTagInvalidStatus     = errors.New("suggested tag is not pending")
	ErrSuggestedTagsProviderDisabled = errors.New("tag suggestion provider is disabled")
	ErrSuggestedTagsNoContent        = errors.New("document has no extractable content for tag suggestions")
)

func (r *Repository) ListSuggestedTagsByDocument(ctx context.Context, tenantID, documentID uuid.UUID, statuses ...string) ([]DocumentSuggestedTag, error) {
	query := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ?", tenantID, documentID).
		Order("created_at ASC, tag_name ASC")

	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}

	var items []DocumentSuggestedTag
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) GetSuggestedTagByID(ctx context.Context, tenantID, documentID, suggestedTagID uuid.UUID) (*DocumentSuggestedTag, error) {
	var item DocumentSuggestedTag
	err := r.db.WithContext(ctx).
		First(&item, "id = ? AND tenant_id = ? AND document_id = ?", suggestedTagID, tenantID, documentID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) DeletePendingSuggestedTags(ctx context.Context, tenantID, documentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ? AND status = ?", tenantID, documentID, SuggestedTagStatusPending).
		Delete(&DocumentSuggestedTag{}).Error
}

func (r *Repository) CreateSuggestedTags(ctx context.Context, items []DocumentSuggestedTag) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&items).Error
}

func (r *Repository) UpdateSuggestedTagStatus(ctx context.Context, tenantID, documentID, suggestedTagID uuid.UUID, status string) error {
	result := r.db.WithContext(ctx).
		Model(&DocumentSuggestedTag{}).
		Where("id = ? AND tenant_id = ? AND document_id = ?", suggestedTagID, tenantID, documentID).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) HasSuggestedTagName(ctx context.Context, tenantID, documentID uuid.UUID, tagName string, statuses ...string) (bool, error) {
	tagName = strings.TrimSpace(tagName)
	if tagName == "" {
		return false, nil
	}

	query := r.db.WithContext(ctx).Model(&DocumentSuggestedTag{}).
		Where("tenant_id = ? AND document_id = ? AND LOWER(tag_name) = LOWER(?)", tenantID, documentID, tagName)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
