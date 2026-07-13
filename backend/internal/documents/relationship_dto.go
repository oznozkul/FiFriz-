package documents

import (
	"time"

	"github.com/google/uuid"
)

type RelatedDocumentNode struct {
	ID               uuid.UUID  `json:"id"`
	OriginalFileName string     `json:"originalFileName"`
	FileExtension    string     `json:"fileExtension"`
	MimeType         string     `json:"mimeType"`
	FolderID         *uuid.UUID `json:"folderId"`
	FolderName       string     `json:"folderName"`
}

type DocumentRelationshipEdge struct {
	ID               uuid.UUID           `json:"id"`
	RelationshipType string              `json:"relationshipType"`
	Direction        string              `json:"direction"`
	SourceDocumentID uuid.UUID           `json:"sourceDocumentId"`
	TargetDocumentID uuid.UUID           `json:"targetDocumentId"`
	RelatedDocument  RelatedDocumentNode `json:"relatedDocument"`
	CreatedByUserID  uuid.UUID           `json:"createdByUserId"`
	CreatedAt        time.Time           `json:"createdAt"`
}

type DocumentRelationshipGraphSummary struct {
	TotalCount    int            `json:"totalCount"`
	OutgoingCount int            `json:"outgoingCount"`
	IncomingCount int            `json:"incomingCount"`
	ByType        map[string]int `json:"byType"`
}

type DocumentRelationshipsResponse struct {
	DocumentID uuid.UUID                        `json:"documentId"`
	Items      []DocumentRelationshipEdge       `json:"items"`
	Summary    DocumentRelationshipGraphSummary `json:"summary"`
}

type CreateDocumentRelationshipRequest struct {
	TargetDocumentID string `json:"targetDocumentId"`
	RelationshipType string `json:"relationshipType"`
}

func buildRelationshipsResponse(documentID uuid.UUID, rows []RelationshipListRow) DocumentRelationshipsResponse {
	items := make([]DocumentRelationshipEdge, 0, len(rows))
	summary := DocumentRelationshipGraphSummary{
		ByType: make(map[string]int),
	}

	for _, row := range rows {
		direction := "incoming"
		relatedID := row.SourceDocumentID
		relatedName := row.SourceOriginalFileName
		relatedExt := row.SourceFileExtension
		relatedMime := row.SourceMimeType
		relatedFolderID := row.SourceFolderID
		relatedFolderName := row.SourceFolderName

		if row.SourceDocumentID == documentID {
			direction = "outgoing"
			relatedID = row.TargetDocumentID
			relatedName = row.TargetOriginalFileName
			relatedExt = row.TargetFileExtension
			relatedMime = row.TargetMimeType
			relatedFolderID = row.TargetFolderID
			relatedFolderName = row.TargetFolderName
			summary.OutgoingCount++
		} else {
			summary.IncomingCount++
		}

		summary.ByType[row.RelationshipType]++

		items = append(items, DocumentRelationshipEdge{
			ID:               row.ID,
			RelationshipType: row.RelationshipType,
			Direction:        direction,
			SourceDocumentID: row.SourceDocumentID,
			TargetDocumentID: row.TargetDocumentID,
			RelatedDocument: RelatedDocumentNode{
				ID:               relatedID,
				OriginalFileName: relatedName,
				FileExtension:    relatedExt,
				MimeType:         relatedMime,
				FolderID:         relatedFolderID,
				FolderName:       relatedFolderName,
			},
			CreatedByUserID: row.CreatedByUserID,
			CreatedAt:       row.CreatedAt,
		})
	}

	summary.TotalCount = len(items)
	return DocumentRelationshipsResponse{
		DocumentID: documentID,
		Items:      items,
		Summary:    summary,
	}
}
