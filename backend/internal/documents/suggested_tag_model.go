package documents

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	SuggestedTagStatusPending  = "pending"
	SuggestedTagStatusAccepted = "accepted"
	SuggestedTagStatusRejected = "rejected"
)

type DocumentSuggestedTag struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index:idx_document_suggested_tags_tenant_document" json:"tenantId"`
	DocumentID  uuid.UUID `gorm:"type:uuid;not null;index:idx_document_suggested_tags_tenant_document" json:"documentId"`
	TagName     string    `gorm:"type:text;not null" json:"tagName"`
	Confidence  *float64  `json:"confidence,omitempty"`
	Status      string    `gorm:"type:text;not null;index" json:"status"`
	GeneratedBy string    `gorm:"type:text;not null;default:ai" json:"generatedBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (s *DocumentSuggestedTag) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.GeneratedBy == "" {
		s.GeneratedBy = "ai"
	}
	return nil
}

func (DocumentSuggestedTag) TableName() string {
	return "document_suggested_tags"
}
