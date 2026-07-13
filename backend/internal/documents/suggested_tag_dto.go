package documents

import (
	"time"

	"fifriz/internal/tags"

	"github.com/google/uuid"
)

type DocumentSuggestedTagResponse struct {
	ID          uuid.UUID `json:"id"`
	DocumentID  uuid.UUID `json:"documentId"`
	TagName     string    `json:"tagName"`
	Confidence  *float64  `json:"confidence,omitempty"`
	Status      string    `json:"status"`
	GeneratedBy string    `json:"generatedBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SuggestedTagGenerateRequest struct {
	TenantID   uuid.UUID
	DocumentID uuid.UUID
	UserID     uuid.UUID
}

type AcceptSuggestedTagsResult struct {
	Accepted []DocumentSuggestedTagResponse `json:"accepted"`
	Tags     []tags.TagBrief                `json:"tags"`
}
