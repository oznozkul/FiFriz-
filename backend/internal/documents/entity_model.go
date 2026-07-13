package documents

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const EntityExtractedByAI = "ai"

type DocumentEntity struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_document_entities_tenant_document" json:"tenantId"`
	DocumentID  uuid.UUID      `gorm:"type:uuid;not null;index:idx_document_entities_tenant_document" json:"documentId"`
	EntityType  string         `gorm:"type:text;not null;index" json:"entityType"`
	Value       string         `gorm:"type:text;not null" json:"value"`
	Label       *string        `gorm:"type:text" json:"label,omitempty"`
	Confidence  *float64       `json:"confidence,omitempty"`
	SourceText  *string        `gorm:"type:text" json:"sourceText,omitempty"`
	PageNumber  *int           `json:"pageNumber,omitempty"`
	ExtractedBy string         `gorm:"type:text;not null;default:ai" json:"extractedBy"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (e *DocumentEntity) BeforeCreate(_ *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.ExtractedBy == "" {
		e.ExtractedBy = EntityExtractedByAI
	}
	return nil
}

func (DocumentEntity) TableName() string {
	return "document_entities"
}
