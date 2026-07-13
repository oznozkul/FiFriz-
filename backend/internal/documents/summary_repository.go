package documents

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrDocumentSummaryNotFound = errors.New("document summary not found")

func (r *Repository) GetSummaryByDocumentID(ctx context.Context, tenantID, documentID uuid.UUID) (*DocumentSummary, error) {
	var summary DocumentSummary
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ?", tenantID, documentID).
		First(&summary).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDocumentSummaryNotFound
	}
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *Repository) SaveDocumentSummary(ctx context.Context, summary *DocumentSummary) error {
	var existing DocumentSummary
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ?", summary.TenantID, summary.DocumentID).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(summary).Error
	}
	if err != nil {
		return err
	}

	existing.GeneratedByUserID = summary.GeneratedByUserID
	existing.Summary = summary.Summary
	existing.KeyPoints = summary.KeyPoints
	existing.SourceChunkIDs = summary.SourceChunkIDs
	existing.Model = summary.Model
	return r.db.WithContext(ctx).Save(&existing).Error
}
