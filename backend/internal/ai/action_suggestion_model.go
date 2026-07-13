package ai

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ActionSuggestionStatusPending  = "pending"
	ActionSuggestionStatusAccepted = "accepted"
	ActionSuggestionStatusRejected = "rejected"

	ActionSuggestionTypeAddTag             = "add_tag"
	ActionSuggestionTypeMoveFolder         = "move_folder"
	ActionSuggestionTypeCreateReminder     = "create_reminder"
	ActionSuggestionTypeStartApproval      = "start_approval"
	ActionSuggestionTypeLinkRelated        = "link_related"
	ActionSuggestionTypeRegenerateMetadata = "regenerate_metadata"
	ActionSuggestionTypeGenerateSummary    = "generate_summary"
	ActionSuggestionTypeFavoriteDocument   = "favorite_document"
)

type AIActionSuggestion struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_ai_action_suggestions_tenant_user" json:"tenantId"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index:idx_ai_action_suggestions_tenant_user" json:"userId"`
	DocumentID     *uuid.UUID `gorm:"type:uuid;index" json:"documentId,omitempty"`
	SuggestionType string     `gorm:"type:text;not null;index" json:"suggestionType"`
	Title          string     `gorm:"type:text;not null" json:"title"`
	Description    *string    `gorm:"type:text" json:"description,omitempty"`
	ActionPayload  []byte     `gorm:"type:jsonb;not null" json:"actionPayload"`
	Status         string     `gorm:"type:text;not null;index" json:"status"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (s *AIActionSuggestion) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Status == "" {
		s.Status = ActionSuggestionStatusPending
	}
	return nil
}

func (AIActionSuggestion) TableName() string {
	return "ai_action_suggestions"
}
