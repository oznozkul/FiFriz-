package notifications

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Type       string     `gorm:"type:text;not null;index" json:"type"`
	Title      string     `gorm:"type:text;not null" json:"title"`
	Message    string     `gorm:"type:text;not null" json:"message"`
	EntityType *string    `gorm:"type:text;index" json:"entity_type"`
	EntityID   *uuid.UUID `gorm:"type:uuid;index" json:"entity_id"`
	IsRead     bool       `gorm:"not null;default:false;index" json:"is_read"`
	ReadAt     *time.Time `gorm:"type:timestamptz" json:"read_at"`
	Metadata   []byte     `gorm:"type:jsonb" json:"metadata"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (n *Notification) BeforeCreate(_ *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

func (Notification) TableName() string {
	return "notifications"
}
