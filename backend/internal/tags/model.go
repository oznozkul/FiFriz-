package tags

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tag struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID        uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_tags_tenant_name" json:"tenant_id"`
	Name            string         `gorm:"size:120;not null;uniqueIndex:idx_tags_tenant_name" json:"name"`
	Status          string         `gorm:"size:32;not null;default:active;index" json:"status"`
	IsSystem        bool           `gorm:"not null;default:false;index" json:"is_system"`
	Color           *string        `gorm:"size:32" json:"color"`
	Icon            *string        `gorm:"size:80" json:"icon"`
	CreatedByUserID uuid.UUID      `gorm:"type:uuid;not null;index" json:"created_by_user_id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type TagAlias struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_tag_aliases_unique" json:"tenant_id"`
	TagID     uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_tag_aliases_unique" json:"tag_id"`
	Alias     string    `gorm:"size:120;not null;uniqueIndex:idx_tag_aliases_unique" json:"alias"`
	CreatedAt time.Time `json:"created_at"`
}

func (TagAlias) TableName() string {
	return "tag_aliases"
}

func (t *Tag) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Status == "" {
		t.Status = StatusActive
	}
	return nil
}

func (a *TagAlias) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type DocumentTag struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_document_tags_unique" json:"tenant_id"`
	DocumentID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_document_tags_unique" json:"document_id"`
	TagID      uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_document_tags_unique" json:"tag_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (d *DocumentTag) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
