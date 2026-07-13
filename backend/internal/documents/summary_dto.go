package documents

import (
	"time"

	"github.com/google/uuid"
)

type DocumentSummaryResponse struct {
	DocumentID uuid.UUID `json:"documentId"`
	Summary    string    `json:"summary"`
	KeyPoints  []string  `json:"keyPoints"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Model      *string   `json:"model,omitempty"`
}
