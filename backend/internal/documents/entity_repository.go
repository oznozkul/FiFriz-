package documents

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *Repository) ListEntitiesByDocument(ctx context.Context, tenantID, documentID uuid.UUID) ([]DocumentEntity, error) {
	var items []DocumentEntity
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ?", tenantID, documentID).
		Order("entity_type ASC, value ASC, created_at ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) GetEntityByID(ctx context.Context, tenantID, documentID, entityID uuid.UUID) (*DocumentEntity, error) {
	var item DocumentEntity
	err := r.db.WithContext(ctx).
		First(&item, "id = ? AND tenant_id = ? AND document_id = ?", entityID, tenantID, documentID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) DeleteAIEntitiesByDocument(ctx context.Context, tenantID, documentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ? AND extracted_by = ?", tenantID, documentID, EntityExtractedByAI).
		Delete(&DocumentEntity{}).Error
}

func (r *Repository) CreateEntities(ctx context.Context, items []DocumentEntity) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&items).Error
}

func (r *Repository) SoftDeleteEntity(ctx context.Context, tenantID, documentID, entityID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND document_id = ?", entityID, tenantID, documentID).
		Delete(&DocumentEntity{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) HasEntity(ctx context.Context, tenantID, documentID uuid.UUID, entityType, value string) (bool, error) {
	entityType = strings.TrimSpace(entityType)
	value = normalizeEntityValue(value)
	if entityType == "" || value == "" {
		return false, nil
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&DocumentEntity{}).
		Where(
			"tenant_id = ? AND document_id = ? AND entity_type = ? AND LOWER(TRIM(value)) = ?",
			tenantID, documentID, entityType, value,
		).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func EnsureDocumentEntityIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_document_entities_type
		ON document_entities (tenant_id, document_id, entity_type)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}
	return db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_document_entities_value_lookup
		ON document_entities (tenant_id, entity_type, value)
		WHERE deleted_at IS NULL
	`).Error
}
