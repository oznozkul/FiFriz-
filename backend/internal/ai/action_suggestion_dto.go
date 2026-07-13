package ai

import (
	"encoding/json"
	"time"

	"fifriz/internal/ai/tools"

	"github.com/google/uuid"
)

type ActionSuggestionResponse struct {
	ID             uuid.UUID              `json:"id"`
	DocumentID     *uuid.UUID             `json:"documentId,omitempty"`
	SuggestionType string                 `json:"suggestionType"`
	Title          string                 `json:"title"`
	Description    *string                `json:"description,omitempty"`
	ActionPayload  map[string]interface{} `json:"actionPayload"`
	Status         string                 `json:"status"`
	CreatedAt      string                 `json:"createdAt"`
	UpdatedAt      string                 `json:"updatedAt"`
}

type ActionSuggestionAcceptResponse struct {
	Suggestion ActionSuggestionResponse  `json:"suggestion"`
	Executed   bool                      `json:"executed"`
	Execution  *tools.ExecutionResult    `json:"execution,omitempty"`
}

func toActionSuggestionResponse(item AIActionSuggestion) (ActionSuggestionResponse, error) {
	payload := map[string]interface{}{}
	if len(item.ActionPayload) > 0 {
		if err := json.Unmarshal(item.ActionPayload, &payload); err != nil {
			return ActionSuggestionResponse{}, err
		}
	}
	return ActionSuggestionResponse{
		ID:             item.ID,
		DocumentID:     item.DocumentID,
		SuggestionType: item.SuggestionType,
		Title:          item.Title,
		Description:    item.Description,
		ActionPayload:  payload,
		Status:         item.Status,
		CreatedAt:      item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      item.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func toActionSuggestionResponses(items []AIActionSuggestion) ([]ActionSuggestionResponse, error) {
	result := make([]ActionSuggestionResponse, 0, len(items))
	for _, item := range items {
		response, err := toActionSuggestionResponse(item)
		if err != nil {
			return nil, err
		}
		result = append(result, response)
	}
	return result, nil
}
