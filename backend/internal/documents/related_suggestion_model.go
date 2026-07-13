package documents

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RelatedSuggestionStatusSuggested = "suggested"
	RelatedSuggestionStatusAccepted  = "accepted"
	RelatedSuggestionStatusRejected  = "rejected"
	RelatedSuggestionGeneratedByAI   = "ai"
)

type DocumentRelatedSuggestion struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID         uuid.UUID `gorm:"type:uuid;not null;index:idx_document_related_suggestions_source" json:"tenantId"`
	SourceDocumentID uuid.UUID `gorm:"type:uuid;not null;index:idx_document_related_suggestions_source" json:"sourceDocumentId"`
	TargetDocumentID uuid.UUID `gorm:"type:uuid;not null;index" json:"targetDocumentId"`
	RelationshipType *string   `gorm:"type:text" json:"relationshipType,omitempty"`
	Reason           *string   `gorm:"type:text" json:"reason,omitempty"`
	Similarity       *float64  `json:"similarity,omitempty"`
	Status           string    `gorm:"type:text;not null;index" json:"status"`
	GeneratedBy      string    `gorm:"type:text;not null;default:ai" json:"generatedBy"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func (s *DocumentRelatedSuggestion) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.GeneratedBy == "" {
		s.GeneratedBy = RelatedSuggestionGeneratedByAI
	}
	if s.Status == "" {
		s.Status = RelatedSuggestionStatusSuggested
	}
	return nil
}

func (DocumentRelatedSuggestion) TableName() string {
	return "document_related_suggestions"
}
