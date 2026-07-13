package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"fifriz/internal/ai/tools"
	"fifriz/internal/approvals"
	"fifriz/internal/audit"
	"fifriz/internal/auth"
	"fifriz/internal/documents"
	"fifriz/internal/folders"
	"fifriz/internal/tags"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrInvalidAgentMessage = errors.New("message is required")

type AgentChatService struct {
	rag          *RAGService
	sessions     *SessionRepository
	chat         ChatProvider
	toolRegistry *tools.Registry
	suggestions  *ActionSuggestionRepository
	documents    *documents.Repository
	folders      *folders.Repository
	tags         *tags.Repository
	approvals    *approvals.Repository
	shareAccess  *documents.ShareAccessService
	audit        audit.Recorder
	context      agentContextBuilder
}

type AgentChatServiceConfig struct {
	RAG          *RAGService
	Sessions     *SessionRepository
	Chat         ChatProvider
	ToolRegistry *tools.Registry
	Suggestions  *ActionSuggestionRepository
	Documents    *documents.Repository
	Folders      *folders.Repository
	Tags         *tags.Repository
	Approvals    *approvals.Repository
	ShareAccess  *documents.ShareAccessService
	Audit        audit.Recorder
}

func NewAgentChatService(cfg AgentChatServiceConfig) *AgentChatService {
	return &AgentChatService{
		rag:          cfg.RAG,
		sessions:     cfg.Sessions,
		chat:         cfg.Chat,
		toolRegistry: cfg.ToolRegistry,
		suggestions:  cfg.Suggestions,
		documents:    cfg.Documents,
		folders:      cfg.Folders,
		tags:         cfg.Tags,
		approvals:    cfg.Approvals,
		shareAccess:  cfg.ShareAccess,
		audit:        cfg.Audit,
		context: agentContextBuilder{
			documents: cfg.Documents,
			folders:   cfg.Folders,
			tags:      cfg.Tags,
			approvals: cfg.Approvals,
		},
	}
}

func (s *AgentChatService) Chat(ctx context.Context, request AgentChatRequest) (*AgentChatResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	message := strings.TrimSpace(request.Message)
	if message == "" {
		return nil, ErrInvalidAgentMessage
	}
	if s.chat == nil {
		return nil, ErrChatProviderDisabled
	}
	if s.toolRegistry == nil {
		return nil, ErrChatProviderDisabled
	}

	if err := s.validateRequestScope(ctx, authContext, request); err != nil {
		return nil, err
	}

	session, isNew, err := s.resolveSession(ctx, authContext, request, message)
	if err != nil {
		return nil, err
	}

	userID := authContext.UserID
	if err := s.sessions.CreateMessage(ctx, &ChatMessage{
		TenantID:  authContext.TenantID,
		SessionID: session.ID,
		UserID:    &userID,
		Role:      MessageRoleUser,
		Content:   message,
	}); err != nil {
		return nil, err
	}
	s.logAgentMessageSent(ctx, authContext, session.ID, request)

	contextText, err := s.context.build(ctx, authContext, request)
	if err != nil {
		return nil, err
	}

	ragRequest := ChatRequest{
		Question:   message,
		DocumentID: request.DocumentID,
		FolderID:   request.FolderID,
	}
	ragResult, err := s.rag.generateRAGAnswer(ctx, ragRequest)
	if err != nil {
		return nil, err
	}

	chunkTexts := make([]string, 0, len(ragResult.Sources))
	for _, source := range ragResult.Sources {
		chunkTexts = append(chunkTexts, fmt.Sprintf("Belge: %s (documentId: %s) | Parça %d\n%s",
			source.DocumentTitle, source.DocumentID, source.ChunkIndex, source.ContentPreview))
	}

	completion, err := s.chat.Complete(ctx, ChatCompletionRequest{
		SystemPrompt: agentSystemPrompt(s.toolRegistry),
		UserPrompt:   buildAgentUserPrompt(message, contextText, chunkTexts),
	})
	if err != nil {
		return nil, err
	}

	parsed, err := parseAgentLLMResponse(completion.Content)
	if err != nil {
		return nil, err
	}
	answer := strings.TrimSpace(parsed.Answer)
	if answer == "" {
		answer = strings.TrimSpace(completion.Content)
	}
	if answer == "" {
		answer = noContextAnswer
	}

	suggestedActions, err := s.createSuggestedActions(ctx, authContext, request, parsed.Actions)
	if err != nil {
		return nil, err
	}

	sourcesJSON, err := json.Marshal(ragResult.Sources)
	if err != nil {
		return nil, err
	}
	if err := s.sessions.CreateMessage(ctx, &ChatMessage{
		TenantID:  authContext.TenantID,
		SessionID: session.ID,
		Role:      MessageRoleAssistant,
		Content:   answer,
		Sources:   sourcesJSON,
	}); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	session.UpdatedAt = now
	if isNew && session.Title == "" {
		session.Title = sessionTitleFromQuestion(message)
	}
	if err := s.sessions.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	return &AgentChatResponse{
		SessionID:        session.ID,
		Answer:           answer,
		Sources:          ragResult.Sources,
		SuggestedActions: suggestedActions,
	}, nil
}

func (s *AgentChatService) validateRequestScope(ctx context.Context, authContext auth.Context, request AgentChatRequest) error {
	if request.DocumentID != nil && strings.TrimSpace(*request.DocumentID) != "" {
		documentID, err := uuid.Parse(strings.TrimSpace(*request.DocumentID))
		if err != nil {
			return ErrInvalidDocumentFilter
		}
		document, err := s.documents.GetByID(ctx, authContext.TenantID, documentID)
		if err != nil {
			return err
		}
		return requireViewAccess(ctx, s.shareAccess, authContext, document)
	}
	if request.FolderID != nil && strings.TrimSpace(*request.FolderID) != "" {
		folderID, err := uuid.Parse(strings.TrimSpace(*request.FolderID))
		if err != nil {
			return ErrInvalidFolderFilter
		}
		_, err = s.folders.GetByID(ctx, authContext.TenantID, folderID)
		return err
	}
	return nil
}

func (s *AgentChatService) resolveSession(ctx context.Context, authContext auth.Context, request AgentChatRequest, message string) (*ChatSession, bool, error) {
	if request.SessionID != nil && strings.TrimSpace(*request.SessionID) != "" {
		parsedSessionID, err := uuid.Parse(strings.TrimSpace(*request.SessionID))
		if err != nil {
			return nil, false, ErrInvalidSessionID
		}
		session, err := s.sessions.GetSession(ctx, authContext.TenantID, authContext.UserID, parsedSessionID)
		if err != nil {
			return nil, false, err
		}
		if err := s.validateAgentSessionScope(ctx, authContext, scopeFromSession(session)); err != nil {
			return nil, false, err
		}
		return session, false, nil
	}

	scope := deriveAgentSessionScope(request)
	if err := s.validateAgentSessionScope(ctx, authContext, scope); err != nil {
		return nil, false, err
	}

	session := &ChatSession{
		TenantID:  authContext.TenantID,
		UserID:    authContext.UserID,
		Title:     sessionTitleFromQuestion(message),
		ScopeType: scope.ScopeType,
		ScopeID:   scope.ScopeID,
	}
	if err := s.sessions.CreateSession(ctx, session); err != nil {
		return nil, false, err
	}

	userID := authContext.UserID
	audit.LogWithActor(ctx, s.audit, authContext.TenantID, &userID, audit.CreateLogInput{
		TenantID:    authContext.TenantID,
		Action:      audit.ActionAIChatSessionCreated,
		EntityType:  audit.EntityTypeAIChatSession,
		EntityID:    &session.ID,
		Description: "AI agent chat session created",
		Metadata: map[string]interface{}{
			"scopeType": scope.ScopeType,
		},
	})

	return session, true, nil
}

func deriveAgentSessionScope(request AgentChatRequest) sessionScope {
	if request.DocumentID != nil && strings.TrimSpace(*request.DocumentID) != "" {
		documentID, err := uuid.Parse(strings.TrimSpace(*request.DocumentID))
		if err == nil {
			scopeType := ScopeTypeAgent
			return sessionScope{ScopeType: &scopeType, ScopeID: &documentID}
		}
	}
	if request.FolderID != nil && strings.TrimSpace(*request.FolderID) != "" {
		folderID, err := uuid.Parse(strings.TrimSpace(*request.FolderID))
		if err == nil {
			scopeType := ScopeTypeWorkspace
			return sessionScope{ScopeType: &scopeType, ScopeID: &folderID}
		}
	}
	scopeType := ScopeTypeWorkspace
	return sessionScope{ScopeType: &scopeType, ScopeID: nil}
}

func (s *AgentChatService) validateAgentSessionScope(ctx context.Context, authContext auth.Context, scope sessionScope) error {
	if scope.ScopeType == nil {
		return nil
	}
	switch *scope.ScopeType {
	case ScopeTypeWorkspace, ScopeTypeAll:
		if scope.ScopeID == nil || s.folders == nil {
			return nil
		}
		_, err := s.folders.GetByID(ctx, authContext.TenantID, *scope.ScopeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return folders.ErrFolderNotFound
			}
			return err
		}
		return nil
	case ScopeTypeAgent:
		if scope.ScopeID == nil {
			return nil
		}
		document, err := s.documents.GetByID(ctx, authContext.TenantID, *scope.ScopeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return documents.ErrForbidden
			}
			return err
		}
		return requireViewAccess(ctx, s.shareAccess, authContext, document)
	default:
		return nil
	}
}

func (s *AgentChatService) createSuggestedActions(
	ctx context.Context,
	authContext auth.Context,
	request AgentChatRequest,
	actions []agentLLMAction,
) ([]AgentSuggestedAction, error) {
	if len(actions) == 0 || s.suggestions == nil {
		return nil, nil
	}

	result := make([]AgentSuggestedAction, 0, len(actions))
	for _, action := range actions {
		toolName := strings.TrimSpace(action.ToolName)
		if toolName == "" {
			continue
		}
		tool, ok := s.toolRegistry.Get(toolName)
		if !ok {
			continue
		}

		input := action.Input
		if input == nil {
			input = map[string]interface{}{}
		}
		s.enrichActionInput(request, input)

		if err := tool.CanExecute(ctx, authContext, input); err != nil {
			continue
		}

		title := strings.TrimSpace(action.Title)
		if title == "" {
			title = tool.Description()
		}
		var description *string
		if trimmed := strings.TrimSpace(action.Description); trimmed != "" {
			description = &trimmed
		}

		extras := map[string]interface{}{}
		if panel := panelForTool(toolName); panel != "" {
			extras["panel"] = panel
		}
		extras["confirmMessage"] = toolConfirmationText(toolName, mergePayloadForConfirm(input, extras))
		payload := withToolPayload(toolName, input, extras)

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		var documentID *uuid.UUID
		if rawDocumentID, ok := input["documentId"].(string); ok && strings.TrimSpace(rawDocumentID) != "" {
			parsed, parseErr := uuid.Parse(strings.TrimSpace(rawDocumentID))
			if parseErr == nil {
				documentID = &parsed
			}
		}

		item := AIActionSuggestion{
			TenantID:       authContext.TenantID,
			UserID:         authContext.UserID,
			DocumentID:     documentID,
			SuggestionType: suggestionTypeForTool(toolName),
			Title:          title,
			Description:    description,
			ActionPayload:  payloadJSON,
			Status:         ActionSuggestionStatusPending,
		}
		if err := s.suggestions.CreateBatch(ctx, []AIActionSuggestion{item}); err != nil {
			return nil, err
		}

		s.logActionSuggested(ctx, authContext, item.ID, toolName, documentID, input)

		result = append(result, AgentSuggestedAction{
			SuggestionID:         item.ID,
			Title:                title,
			Description:          description,
			ToolName:             toolName,
			ActionPayload:        payload,
			RequiresConfirmation: true,
		})
	}

	return result, nil
}

func (s *AgentChatService) enrichActionInput(request AgentChatRequest, input map[string]interface{}) {
	if _, ok := input["documentId"]; !ok && request.DocumentID != nil && strings.TrimSpace(*request.DocumentID) != "" {
		input["documentId"] = strings.TrimSpace(*request.DocumentID)
	}
	if _, ok := input["folderId"]; !ok && request.FolderID != nil && strings.TrimSpace(*request.FolderID) != "" {
		input["folderId"] = strings.TrimSpace(*request.FolderID)
	}
}

func mergePayloadForConfirm(input map[string]interface{}, extras map[string]interface{}) map[string]interface{} {
	merged := map[string]interface{}{}
	for key, value := range input {
		merged[key] = value
	}
	for key, value := range extras {
		merged[key] = value
	}
	return merged
}

func (s *AgentChatService) logAgentMessageSent(ctx context.Context, authContext auth.Context, sessionID uuid.UUID, request AgentChatRequest) {
	userID := authContext.UserID
	metadata := map[string]interface{}{
		"sessionId": sessionID.String(),
	}
	if request.DocumentID != nil {
		metadata["documentId"] = strings.TrimSpace(*request.DocumentID)
	}
	if request.FolderID != nil {
		metadata["folderId"] = strings.TrimSpace(*request.FolderID)
	}
	audit.LogWithActor(ctx, s.audit, authContext.TenantID, &userID, audit.CreateLogInput{
		TenantID:    authContext.TenantID,
		Action:      audit.ActionAIAgentMessageSent,
		EntityType:  audit.EntityTypeAIChatSession,
		EntityID:    &sessionID,
		Description: "AI agent message sent",
		Metadata:    metadata,
	})
}

func (s *AgentChatService) logActionSuggested(
	ctx context.Context,
	authContext auth.Context,
	suggestionID uuid.UUID,
	toolName string,
	documentID *uuid.UUID,
	input map[string]interface{},
) {
	userID := authContext.UserID
	metadata := map[string]interface{}{
		"toolName":     toolName,
		"suggestionId": suggestionID.String(),
		"input":        input,
	}
	if documentID != nil {
		metadata["documentId"] = documentID.String()
	}
	audit.LogWithActor(ctx, s.audit, authContext.TenantID, &userID, audit.CreateLogInput{
		TenantID:    authContext.TenantID,
		Action:      audit.ActionAIAgentActionSuggested,
		EntityType:  audit.EntityTypeAIChatSession,
		EntityID:    &suggestionID,
		Description: "AI agent action suggested",
		Metadata:    metadata,
	})
}
