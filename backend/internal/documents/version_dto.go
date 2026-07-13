package documents

import (
	"time"

	"github.com/google/uuid"
)

type DocumentVersionResponse struct {
	ID                 uuid.UUID `json:"id"`
	DocumentID         uuid.UUID `json:"documentId"`
	VersionNumber      int       `json:"versionNumber"`
	OriginalFileName   string    `json:"originalFileName"`
	FileExtension      string    `json:"fileExtension"`
	MimeType           string    `json:"mimeType"`
	FileSize           int64     `json:"fileSize"`
	UploadedByUserID   uuid.UUID `json:"uploadedByUserId"`
	UploadedByFullName string    `json:"uploadedByFullName"`
	ChangeNote         *string   `json:"changeNote"`
	IsCurrent          bool      `json:"isCurrent"`
	CreatedAt          time.Time `json:"createdAt"`
}

func toDocumentVersionResponse(row VersionListRow) DocumentVersionResponse {
	return DocumentVersionResponse{
		ID:                 row.ID,
		DocumentID:         row.DocumentID,
		VersionNumber:      row.VersionNumber,
		OriginalFileName:   row.OriginalFileName,
		FileExtension:      row.FileExtension,
		MimeType:           row.MimeType,
		FileSize:           row.FileSize,
		UploadedByUserID:   row.UploadedByUserID,
		UploadedByFullName: row.UploadedByFullName,
		ChangeNote:         row.ChangeNote,
		IsCurrent:          row.IsCurrent,
		CreatedAt:          row.CreatedAt,
	}
}
