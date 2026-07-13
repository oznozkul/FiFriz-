package ai

import "github.com/google/uuid"

const (
	InsightTypeUpcomingDates        = "upcoming_dates"
	InsightTypePendingApprovals     = "pending_approvals"
	InsightTypeRecentImportant      = "recent_important_documents"
	InsightTypeClassifiedDocuments  = "classified_documents"
	InsightTypeMissingMetadata      = "missing_metadata"
	InsightTypeFailedProcessing     = "failed_processing"
	InsightTypeRelatedSuggestions   = "related_suggestions"
	InsightTypePopularTags          = "popular_tags"
	InsightTypeActiveFolders          = "active_folders"

	InsightPriorityHigh   = "high"
	InsightPriorityMedium = "medium"
	InsightPriorityLow    = "low"
)

type WorkspaceInsightRelatedDocument struct {
	DocumentID       uuid.UUID `json:"documentId"`
	OriginalFileName string    `json:"originalFileName"`
	FileExtension    string    `json:"fileExtension,omitempty"`
}

type WorkspaceInsight struct {
	Title            string                            `json:"title"`
	Description      string                            `json:"description"`
	Type             string                            `json:"type"`
	Priority         string                            `json:"priority"`
	RelatedDocuments []WorkspaceInsightRelatedDocument `json:"relatedDocuments"`
	ActionLabel      string                            `json:"actionLabel"`
	ActionURL        string                            `json:"actionUrl"`
}
