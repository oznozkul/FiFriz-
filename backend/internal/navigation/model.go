package navigation

import (
	"time"

	"github.com/google/uuid"
)

type PinnedItem struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index:idx_pins_tenant_user"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index:idx_pins_tenant_user"`
	EntityType string    `gorm:"type:varchar(64);not null"`
	EntityID   string    `gorm:"type:varchar(128);not null"`
	Label      string    `gorm:"type:varchar(255);not null"`
	Path       string    `gorm:"type:varchar(512);not null"`
	CreatedAt  time.Time
}

func (PinnedItem) TableName() string {
	return "pinned_items"
}

type RecentItem struct {
	EntityType string    `json:"entityType"`
	EntityID   string    `json:"entityId"`
	Label      string    `json:"label"`
	Path       string    `json:"path"`
	OpenedAt   time.Time `json:"openedAt"`
}
