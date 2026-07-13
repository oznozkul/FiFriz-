package documents

import "time"

type DocumentRelatedSuggestionResponse struct {
	ID               string              `json:"id"`
	SourceDocumentID string              `json:"sourceDocumentId"`
	TargetDocumentID string              `json:"targetDocumentId"`
	RelationshipType *string             `json:"relationshipType,omitempty"`
	Reason           *string             `json:"reason,omitempty"`
	Similarity       *float64            `json:"similarity,omitempty"`
	Status           string              `json:"status"`
	GeneratedBy      string              `json:"generatedBy"`
	TargetDocument   RelatedDocumentNode `json:"targetDocument"`
	CreatedAt        time.Time           `json:"createdAt"`
	UpdatedAt        time.Time           `json:"updatedAt"`
}

func toRelatedSuggestionResponse(row RelatedSuggestionListRow) DocumentRelatedSuggestionResponse {
	return DocumentRelatedSuggestionResponse{
		ID:               row.ID.String(),
		SourceDocumentID: row.SourceDocumentID.String(),
		TargetDocumentID: row.TargetDocumentID.String(),
		RelationshipType: row.RelationshipType,
		Reason:           row.Reason,
		Similarity:       row.Similarity,
		Status:           row.Status,
		GeneratedBy:      row.GeneratedBy,
		TargetDocument: RelatedDocumentNode{
			ID:               row.TargetDocumentID,
			OriginalFileName: row.TargetOriginalFileName,
			FileExtension:    row.TargetFileExtension,
			MimeType:         row.TargetMimeType,
			FolderID:         row.TargetFolderID,
			FolderName:       row.TargetFolderName,
		},
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toRelatedSuggestionResponses(rows []RelatedSuggestionListRow) []DocumentRelatedSuggestionResponse {
	responses := make([]DocumentRelatedSuggestionResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, toRelatedSuggestionResponse(row))
	}
	return responses
}
