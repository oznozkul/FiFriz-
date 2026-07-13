package ai

import (
	"time"

	"github.com/google/uuid"
)

type ChatRequest struct {
	Question      string   `json:"question"`
	SessionID     *string  `json:"sessionId"`
	DocumentID    *string  `json:"documentId"`
	FolderID      *string  `json:"folderId"`
	TagIDs        []string `json:"tagIds"`
	CollectionKey *string  `json:"collectionKey"`
}

type CreateSessionRequest struct {
	Title     *string `json:"title"`
	ScopeType *string `json:"scopeType"`
	ScopeID   *string `json:"scopeId"`
}

type ChatSource struct {
	DocumentID     uuid.UUID `json:"documentId"`
	DocumentTitle  string    `json:"documentTitle"`
	ChunkID        uuid.UUID `json:"chunkId"`
	ChunkIndex     int       `json:"chunkIndex"`
	Similarity     float64   `json:"similarity"`
	ContentPreview string    `json:"contentPreview"`
}

type ChatResponse struct {
	SessionID uuid.UUID    `json:"sessionId"`
	Answer    string       `json:"answer"`
	Sources   []ChatSource `json:"sources"`
}

type ChatSessionSummaryResponse struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	ScopeType *string    `json:"scopeType"`
	ScopeID   *uuid.UUID `json:"scopeId"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type ChatMessageResponse struct {
	ID        uuid.UUID    `json:"id"`
	Role      string       `json:"role"`
	Content   string       `json:"content"`
	Sources   []ChatSource `json:"sources,omitempty"`
	CreatedAt time.Time    `json:"createdAt"`
}

type ChatSessionDetailResponse struct {
	ID        uuid.UUID               `json:"id"`
	Title     string                  `json:"title"`
	ScopeType *string                 `json:"scopeType"`
	ScopeID   *uuid.UUID              `json:"scopeId"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`
	Messages  []ChatMessageResponse   `json:"messages"`
}

type AgentChatRequest struct {
	Message    string  `json:"message"`
	SessionID  *string `json:"sessionId"`
	DocumentID *string `json:"documentId"`
	FolderID   *string `json:"folderId"`
}

type AgentSuggestedAction struct {
	SuggestionID         uuid.UUID              `json:"suggestionId"`
	Title                string                 `json:"title"`
	Description          *string                `json:"description,omitempty"`
	ToolName             string                 `json:"toolName"`
	ActionPayload        map[string]interface{} `json:"actionPayload"`
	RequiresConfirmation bool                   `json:"requiresConfirmation"`
}

type AgentChatResponse struct {
	SessionID        uuid.UUID              `json:"sessionId"`
	Answer           string                 `json:"answer"`
	Sources          []ChatSource           `json:"sources,omitempty"`
	SuggestedActions []AgentSuggestedAction `json:"suggestedActions,omitempty"`
}
