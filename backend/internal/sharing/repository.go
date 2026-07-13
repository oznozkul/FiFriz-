package sharing

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, link *DocumentShareLink) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(link).Error
	})
}

func (r *Repository) ListActiveByDocument(ctx context.Context, tenantID, documentID uuid.UUID) ([]DocumentShareLink, error) {
	var items []DocumentShareLink
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ? AND is_active = ?", tenantID, documentID, true).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *Repository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*DocumentShareLink, error) {
	var link DocumentShareLink
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *Repository) GetByToken(ctx context.Context, token string) (*DocumentShareLink, error) {
	var link DocumentShareLink
	err := r.db.WithContext(ctx).
		Where("token = ?", token).
		First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *Repository) Deactivate(ctx context.Context, tenantID, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&DocumentShareLink{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{
			"is_active": false,
		}).Error
}

func (r *Repository) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&DocumentShareLink{}).Error
}

func (r *Repository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&DocumentShareLink{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
