package documents

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MetadataParty struct {
	Name string `json:"name"`
	Role string `json:"role,omitempty"`
}

type MetadataDate struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type MetadataAmount struct {
	Label    string   `json:"label"`
	Value    float64  `json:"value"`
	Currency string   `json:"currency,omitempty"`
}

type MetadataIdentifier struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type ExtractedDocumentMetadata struct {
	DocumentType string               `json:"documentType,omitempty"`
	Title        string               `json:"title,omitempty"`
	Parties      []MetadataParty      `json:"parties"`
	Dates        []MetadataDate       `json:"dates"`
	Amounts      []MetadataAmount     `json:"amounts"`
	Identifiers  []MetadataIdentifier `json:"identifiers"`
	Addresses    []string             `json:"addresses"`
	Keywords     []string             `json:"keywords"`
	Summary      string               `json:"summary,omitempty"`
	Confidence   *float64             `json:"confidence,omitempty"`
}

type DocumentAIMetadataResponse struct {
	DocumentID   uuid.UUID                 `json:"documentId"`
	DocumentType *string                   `json:"documentType,omitempty"`
	Metadata     ExtractedDocumentMetadata `json:"metadata"`
	Confidence   *float64                  `json:"confidence,omitempty"`
	Model        *string                   `json:"model,omitempty"`
	ExtractedBy  string                    `json:"extractedBy"`
	CreatedAt    time.Time                 `json:"createdAt"`
	UpdatedAt    time.Time                 `json:"updatedAt"`
	RawJSON      json.RawMessage           `json:"rawJson,omitempty"`
}
