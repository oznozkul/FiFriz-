package notifications

import (
	"time"

	"github.com/google/uuid"
)

type NotificationResponse struct {
	ID         uuid.UUID              `json:"id"`
	Type       string                 `json:"type"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	EntityType *string                `json:"entityType"`
	EntityID   *uuid.UUID             `json:"entityId"`
	IsRead     bool                   `json:"isRead"`
	ReadAt     *time.Time             `json:"readAt"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"createdAt"`
}

type NotificationListResponse struct {
	Items      []NotificationResponse `json:"items"`
	TotalCount int64                  `json:"totalCount"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"pageSize"`
}

type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

type CreateInput struct {
	TenantID    uuid.UUID
	UserID      uuid.UUID
	ActorUserID uuid.UUID
	Type        string
	Title       string
	Message     string
	EntityType  *string
	EntityID    *uuid.UUID
	Metadata    map[string]interface{}
}

type ListFilter struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	UnreadOnly bool
	Page       int
	PageSize   int
}
