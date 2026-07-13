package documents

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type ClassificationGenerateRequest struct {
	TenantID   uuid.UUID
	DocumentID uuid.UUID
	UserID     uuid.UUID
}

type DocumentClassificationResponse struct {
	DocumentID   uuid.UUID  `json:"documentId"`
	DocumentType string     `json:"documentType"`
	Confidence   *float64   `json:"confidence,omitempty"`
	Reason       *string    `json:"reason,omitempty"`
	ClassifiedAt *time.Time `json:"classifiedAt,omitempty"`
}

type extractedDocumentClassification struct {
	DocumentType string   `json:"documentType"`
	Confidence   *float64 `json:"confidence"`
	Reason       string   `json:"reason"`
}

func toDocumentClassificationResponse(document *Document) *DocumentClassificationResponse {
	if document == nil || document.AIDocumentType == nil || strings.TrimSpace(*document.AIDocumentType) == "" {
		return nil
	}

	return &DocumentClassificationResponse{
		DocumentID:   document.ID,
		DocumentType: strings.TrimSpace(*document.AIDocumentType),
		Confidence:   document.AIDocumentTypeConfidence,
		Reason:       document.AIClassificationReason,
		ClassifiedAt: document.AIClassifiedAt,
	}
}
