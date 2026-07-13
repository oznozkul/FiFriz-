package documents

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentAIMetadata struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_document_ai_metadata_tenant_document" json:"tenantId"`
	DocumentID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_document_ai_metadata_tenant_document" json:"documentId"`
	ExtractedBy  string    `gorm:"type:text;not null;default:ai" json:"extractedBy"`
	DocumentType *string   `gorm:"type:text" json:"documentType,omitempty"`
	Metadata     []byte    `gorm:"type:jsonb;not null" json:"metadata"`
	Confidence   *float64  `json:"confidence,omitempty"`
	Model        *string   `gorm:"type:text" json:"model,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (m *DocumentAIMetadata) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.ExtractedBy == "" {
		m.ExtractedBy = "ai"
	}
	return nil
}

func (DocumentAIMetadata) TableName() string {
	return "document_ai_metadata"
}

type AIMetadataGenerateRequest struct {
	TenantID   uuid.UUID
	DocumentID uuid.UUID
	UserID     uuid.UUID
}
