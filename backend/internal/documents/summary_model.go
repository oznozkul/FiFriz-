package documents

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentSummary struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_document_summaries_tenant_document" json:"tenantId"`
	DocumentID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_document_summaries_tenant_document" json:"documentId"`
	GeneratedByUserID uuid.UUID `gorm:"type:uuid;not null;index" json:"generatedByUserId"`
	Summary           string    `gorm:"type:text;not null" json:"summary"`
	KeyPoints         []byte    `gorm:"type:jsonb" json:"keyPoints"`
	SourceChunkIDs    []byte    `gorm:"type:jsonb" json:"sourceChunkIds"`
	Model             *string   `gorm:"type:text" json:"model"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func (s *DocumentSummary) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (DocumentSummary) TableName() string {
	return "document_summaries"
}

// SummaryGenerateRequest is the input for sync or future queued summary generation.
type SummaryGenerateRequest struct {
	TenantID   uuid.UUID
	DocumentID uuid.UUID
	UserID     uuid.UUID
}

// SummaryGenerator generates document summaries; sync today, queue-friendly tomorrow.
type SummaryGenerator interface {
	GenerateSummary(ctx context.Context, request SummaryGenerateRequest) (*DocumentSummaryResponse, error)
}
