package documents

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type DocumentImportantDateResponse struct {
	ID                   uuid.UUID  `json:"id"`
	DocumentID           uuid.UUID  `json:"documentId"`
	Title                string     `json:"title"`
	Date                 string     `json:"date"`
	ReminderDaysBefore   *int       `json:"reminderDaysBefore"`
	Note                 *string    `json:"note"`
	Type                 *string    `json:"type"`
	IsCompleted          bool       `json:"isCompleted"`
	CompletedAt          *time.Time `json:"completedAt"`
	CreatedByUserID      uuid.UUID  `json:"createdByUserId"`
	CreatedByUserName    string     `json:"createdByUserName"`
	AssignedToUserID     *uuid.UUID `json:"assignedToUserId"`
	AssignedToUserName   *string    `json:"assignedToUserName"`
	DocumentName         string     `json:"documentName,omitempty"`
	Source               string     `json:"source"`
	Status               string     `json:"status"`
	Confidence           *float64   `json:"confidence,omitempty"`
	SourceText           *string    `json:"sourceText,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type CreateDocumentImportantDateRequest struct {
	Title              string     `json:"title"`
	Date               string     `json:"date"`
	ReminderDaysBefore *int       `json:"reminderDaysBefore"`
	Note               *string    `json:"note"`
	Type               *string    `json:"type"`
	AssignedToUserID   *uuid.UUID `json:"assignedToUserId"`
}

type UpdateDocumentImportantDateRequest struct {
	Title              string     `json:"title"`
	Date               string     `json:"date"`
	ReminderDaysBefore *int       `json:"reminderDaysBefore"`
	Note               *string    `json:"note"`
	Type               *string    `json:"type"`
	AssignedToUserID   *uuid.UUID `json:"assignedToUserId"`
}

type ImportantDateListQuery struct {
	UpcomingOnly bool
	Completed    *bool
	AssignedToMe bool
	From         *time.Time
	To           *time.Time
	Type         *string
}

type ImportantDateListRow struct {
	DocumentImportantDate
	CreatedByUserName  string  `gorm:"column:created_by_user_name"`
	AssignedToUserName *string `gorm:"column:assigned_to_user_name"`
	DocumentName       string  `gorm:"column:document_name"`
}

func toDocumentImportantDateResponse(row ImportantDateListRow) DocumentImportantDateResponse {
	source := row.Source
	if strings.TrimSpace(source) == "" {
		source = ImportantDateSourceManual
	}
	status := row.Status
	if strings.TrimSpace(status) == "" {
		status = ImportantDateStatusActive
	}

	return DocumentImportantDateResponse{
		ID:                 row.ID,
		DocumentID:         row.DocumentID,
		Title:              row.Title,
		Date:               row.Date.Format("2006-01-02"),
		ReminderDaysBefore: row.ReminderDaysBefore,
		Note:               row.Note,
		Type:               row.Type,
		IsCompleted:        row.IsCompleted,
		CompletedAt:        row.CompletedAt,
		CreatedByUserID:    row.CreatedByUserID,
		CreatedByUserName:  row.CreatedByUserName,
		AssignedToUserID:   row.AssignedToUserID,
		AssignedToUserName: row.AssignedToUserName,
		DocumentName:       row.DocumentName,
		Source:             source,
		Status:             status,
		Confidence:         row.Confidence,
		SourceText:         row.SourceText,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

func toImportantDateResponses(rows []ImportantDateListRow) []DocumentImportantDateResponse {
	result := make([]DocumentImportantDateResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDocumentImportantDateResponse(row))
	}
	return result
}
