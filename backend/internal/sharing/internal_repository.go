package sharing

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *Repository) CreateInternalShare(ctx context.Context, share *DocumentShare) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&DocumentShare{}).
			Where("tenant_id = ? AND document_id = ? AND shared_with_user_id = ?", share.TenantID, share.DocumentID, share.SharedWithUserID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrDuplicateInternalShare
		}
		return tx.Create(share).Error
	})
}

func (r *Repository) GetInternalShareByID(ctx context.Context, tenantID, shareID uuid.UUID) (*DocumentShare, error) {
	var share DocumentShare
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, shareID).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (r *Repository) GetActiveInternalShareByDocumentAndUser(ctx context.Context, tenantID, documentID, userID uuid.UUID) (*DocumentShare, error) {
	var share DocumentShare
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ? AND shared_with_user_id = ?", tenantID, documentID, userID).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (r *Repository) ListInternalSharesByDocument(ctx context.Context, tenantID, documentID uuid.UUID) ([]DocumentShare, error) {
	var shares []DocumentShare
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ?", tenantID, documentID).
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

func (r *Repository) ListInternalSharesForUser(ctx context.Context, tenantID, userID uuid.UUID) ([]DocumentShare, error) {
	var shares []DocumentShare
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND shared_with_user_id = ?", tenantID, userID).
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

func (r *Repository) UpdateInternalShare(ctx context.Context, share *DocumentShare) error {
	return r.db.WithContext(ctx).Save(share).Error
}

func (r *Repository) SoftDeleteInternalShare(ctx context.Context, tenantID, shareID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, shareID).
		Delete(&DocumentShare{}).Error
}

func (r *Repository) HasActiveInternalShare(ctx context.Context, tenantID, documentID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&DocumentShare{}).
		Where("tenant_id = ? AND document_id = ? AND shared_with_user_id = ?", tenantID, documentID, userID).
		Count(&count).Error
	return count > 0, err
}

func EnsureInternalShareUniqueIndex(db *gorm.DB) error {
	return db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_document_shares_unique_active
		ON document_shares (document_id, shared_with_user_id)
		WHERE deleted_at IS NULL
	`).Error
}
