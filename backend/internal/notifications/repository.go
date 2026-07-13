package notifications

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, notification *Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]Notification, int64, error) {
	query := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("tenant_id = ? AND user_id = ?", filter.TenantID, filter.UserID)

	if filter.UnreadOnly {
		query = query.Where("is_read = ?", false)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var items []Notification
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, totalCount, err
}

func (r *Repository) CountUnread(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("tenant_id = ? AND user_id = ? AND is_read = ?", tenantID, userID, false).
		Count(&count).Error
	return count, err
}

func (r *Repository) GetByID(ctx context.Context, tenantID, userID, id uuid.UUID) (*Notification, error) {
	var notification Notification
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ? AND id = ?", tenantID, userID, id).
		First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *Repository) MarkRead(ctx context.Context, tenantID, userID, id uuid.UUID, readAt time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("tenant_id = ? AND user_id = ? AND id = ?", tenantID, userID, id).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": readAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID, readAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("tenant_id = ? AND user_id = ? AND is_read = ?", tenantID, userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": readAt,
		}).Error
}

func toNotificationResponse(notification Notification) NotificationResponse {
	var metadata map[string]interface{}
	if len(notification.Metadata) > 0 {
		_ = json.Unmarshal(notification.Metadata, &metadata)
	}

	return NotificationResponse{
		ID:         notification.ID,
		Type:       notification.Type,
		Title:      notification.Title,
		Message:    notification.Message,
		EntityType: notification.EntityType,
		EntityID:   notification.EntityID,
		IsRead:     notification.IsRead,
		ReadAt:     notification.ReadAt,
		Metadata:   metadata,
		CreatedAt:  notification.CreatedAt,
	}
}
