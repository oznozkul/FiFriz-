package ai

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrActionSuggestionNotFound = errors.New("action suggestion not found")

type ActionSuggestionRepository struct {
	db *gorm.DB
}

func NewActionSuggestionRepository(db *gorm.DB) *ActionSuggestionRepository {
	return &ActionSuggestionRepository{db: db}
}

func (r *ActionSuggestionRepository) ListPendingByUser(ctx context.Context, tenantID, userID uuid.UUID, documentID *uuid.UUID, limit int) ([]AIActionSuggestion, error) {
	if limit <= 0 {
		limit = 50
	}
	query := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ? AND status = ?", tenantID, userID, ActionSuggestionStatusPending)
	if documentID != nil {
		query = query.Where("document_id = ?", *documentID)
	}
	var items []AIActionSuggestion
	err := query.Order("created_at DESC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *ActionSuggestionRepository) GetByID(ctx context.Context, tenantID, userID, id uuid.UUID) (*AIActionSuggestion, error) {
	var item AIActionSuggestion
	err := r.db.WithContext(ctx).
		First(&item, "id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrActionSuggestionNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *ActionSuggestionRepository) HasPendingByType(ctx context.Context, tenantID, userID uuid.UUID, documentID *uuid.UUID, suggestionType string) (bool, error) {
	query := r.db.WithContext(ctx).Model(&AIActionSuggestion{}).
		Where("tenant_id = ? AND user_id = ? AND suggestion_type = ? AND status = ?", tenantID, userID, suggestionType, ActionSuggestionStatusPending)
	if documentID != nil {
		query = query.Where("document_id = ?", *documentID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *ActionSuggestionRepository) DeletePendingByDocument(ctx context.Context, tenantID, userID, documentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ? AND document_id = ? AND status = ?", tenantID, userID, documentID, ActionSuggestionStatusPending).
		Delete(&AIActionSuggestion{}).Error
}

func (r *ActionSuggestionRepository) CreateBatch(ctx context.Context, items []AIActionSuggestion) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&items).Error
}

func (r *ActionSuggestionRepository) UpdateStatus(ctx context.Context, tenantID, userID, id uuid.UUID, status string) error {
	result := r.db.WithContext(ctx).Model(&AIActionSuggestion{}).
		Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrActionSuggestionNotFound
	}
	return nil
}
