package navigation

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrPinNotFound = errors.New("pin not found")
var ErrInvalidPinData = errors.New("invalid pin data")
var ErrPinAlreadyExists = errors.New("pin already exists")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]PinnedItem, error) {
	var items []PinnedItem
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *Repository) GetByID(ctx context.Context, tenantID, userID, pinID uuid.UUID) (*PinnedItem, error) {
	var item PinnedItem
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND user_id = ?", pinID, tenantID, userID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) FindByEntity(ctx context.Context, tenantID, userID uuid.UUID, entityType, entityID string) (*PinnedItem, error) {
	var item PinnedItem
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ? AND entity_type = ? AND entity_id = ?", tenantID, userID, entityType, entityID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item *PinnedItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *Repository) Delete(ctx context.Context, tenantID, userID, pinID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND user_id = ?", pinID, tenantID, userID).
		Delete(&PinnedItem{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
