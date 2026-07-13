package documents

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentComment struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	DocumentID uuid.UUID      `gorm:"type:uuid;not null;index" json:"document_id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	ParentID   *uuid.UUID     `gorm:"type:uuid;index" json:"parent_id"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	IsEdited   bool           `gorm:"not null;default:false" json:"is_edited"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *DocumentComment) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (c *DocumentComment) TableName() string {
	return "document_comments"
}
