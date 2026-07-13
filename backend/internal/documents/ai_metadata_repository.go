package documents

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrDocumentAIMetadataNotFound = errors.New("document ai metadata not found")

func (r *Repository) GetAIMetadataByDocumentID(ctx context.Context, tenantID, documentID uuid.UUID) (*DocumentAIMetadata, error) {
	var record DocumentAIMetadata
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ?", tenantID, documentID).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDocumentAIMetadataNotFound
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) SaveAIMetadata(ctx context.Context, record *DocumentAIMetadata) error {
	var existing DocumentAIMetadata
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ?", record.TenantID, record.DocumentID).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(record).Error
	}
	if err != nil {
		return err
	}

	existing.ExtractedBy = record.ExtractedBy
	existing.DocumentType = record.DocumentType
	existing.Metadata = record.Metadata
	existing.Confidence = record.Confidence
	existing.Model = record.Model
	return r.db.WithContext(ctx).Save(&existing).Error
}
