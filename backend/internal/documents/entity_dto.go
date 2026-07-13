package documents

import "time"

type DocumentEntityResponse struct {
	ID          string    `json:"id"`
	DocumentID  string    `json:"documentId"`
	EntityType  string    `json:"entityType"`
	Value       string    `json:"value"`
	Label       *string   `json:"label,omitempty"`
	Confidence  *float64  `json:"confidence,omitempty"`
	SourceText  *string   `json:"sourceText,omitempty"`
	PageNumber  *int      `json:"pageNumber,omitempty"`
	ExtractedBy string    `json:"extractedBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toDocumentEntityResponse(item DocumentEntity) DocumentEntityResponse {
	return DocumentEntityResponse{
		ID:          item.ID.String(),
		DocumentID:  item.DocumentID.String(),
		EntityType:  item.EntityType,
		Value:       item.Value,
		Label:       item.Label,
		Confidence:  item.Confidence,
		SourceText:  item.SourceText,
		PageNumber:  item.PageNumber,
		ExtractedBy: item.ExtractedBy,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func toDocumentEntityResponses(items []DocumentEntity) []DocumentEntityResponse {
	responses := make([]DocumentEntityResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toDocumentEntityResponse(item))
	}
	return responses
}
