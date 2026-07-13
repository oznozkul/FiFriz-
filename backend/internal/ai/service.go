package ai

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"fifriz/internal/audit"
	"fifriz/internal/auth"
	"fifriz/internal/folders"
	"fifriz/internal/tags"

	"github.com/google/uuid"
)

var (
	ErrInvalidSessionID = errors.New("invalid session id")
	ErrInvalidScopeType = errors.New("invalid scope type")
)

type Service struct {
	rag      *RAGService
	sessions *SessionRepository
	audit    audit.Recorder
	folders  *folders.Repository
	tags     *tags.Repository
}

type ServiceConfig struct {
	RAG      *RAGService
	Sessions *SessionRepository
	Audit    audit.Recorder
	Folders  *folders.Repository
	Tags     *tags.Repository
}

func NewService(cfg ServiceConfig) *Service {
	return &Service{
		rag:      cfg.RAG,
		sessions: cfg.Sessions,
		audit:    cfg.Audit,
		folders:  cfg.Folders,
		tags:     cfg.Tags,
	}
}

func (s *Service) Chat(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	question := strings.TrimSpace(request.Question)
	if question == "" {
		return nil, ErrInvalidQuestion
	}

	session, isNew, err := s.resolveSession(ctx, authContext, request, question)
	if err != nil {
		return nil, err
	}

	userID := authContext.UserID
	if err := s.sessions.CreateMessage(ctx, &ChatMessage{
		TenantID:  authContext.TenantID,
		SessionID: session.ID,
		UserID:    &userID,
		Role:      MessageRoleUser,
		Content:   question,
	}); err != nil {
		return nil, err
	}
	s.logMessageSent(ctx, authContext, session.ID)

	ragRequest := requestWithSessionScope(request, session)
	result, err := s.rag.generateRAGAnswer(ctx, ragRequest)
	if err != nil {
		return nil, err
	}

	sourcesJSON, err := json.Marshal(result.Sources)
	if err != nil {
		return nil, err
	}
	if err := s.sessions.CreateMessage(ctx, &ChatMessage{
		TenantID:  authContext.TenantID,
		SessionID: session.ID,
		Role:      MessageRoleAssistant,
		Content:   result.Answer,
		Sources:   sourcesJSON,
	}); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	session.UpdatedAt = now
	if isNew && session.Title == "" {
		session.Title = sessionTitleFromQuestion(question)
	}
	if err := s.sessions.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	result.SessionID = session.ID
	return result, nil
}

func (s *Service) CreateSession(ctx context.Context, request CreateSessionRequest) (*ChatSessionSummaryResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	scope, err := parseSessionScope(request.ScopeType, request.ScopeID)
	if err != nil {
		return nil, err
	}
	if err := s.validateSessionScopeAccess(ctx, authContext, scope); err != nil {
		return nil, err
	}

	title := "Yeni sohbet"
	if request.Title != nil {
		if trimmed := strings.TrimSpace(*request.Title); trimmed != "" {
			title = trimmed
		}
	}

	session := &ChatSession{
		TenantID:  authContext.TenantID,
		UserID:    authContext.UserID,
		Title:     title,
		ScopeType: scope.ScopeType,
		ScopeID:   scope.ScopeID,
	}
	if err := s.sessions.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	s.logSessionCreated(ctx, authContext, session.ID)

	response := toSessionSummary(session)
	return &response, nil
}

func (s *Service) ListSessions(ctx context.Context, scopeType *string, scopeID *string) ([]ChatSessionSummaryResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedScopeType, parsedScopeID, err := parseListScopeFilter(scopeType, scopeID)
	if err != nil {
		return nil, err
	}

	sessions, err := s.sessions.ListSessions(ctx, authContext.TenantID, authContext.UserID, parsedScopeType, parsedScopeID)
	if err != nil {
		return nil, err
	}

	responses := make([]ChatSessionSummaryResponse, 0, len(sessions))
	for index := range sessions {
		responses = append(responses, toSessionSummary(&sessions[index]))
	}
	return responses, nil
}

func (s *Service) GetSession(ctx context.Context, sessionID string) (*ChatSessionDetailResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedSessionID, err := uuid.Parse(strings.TrimSpace(sessionID))
	if err != nil {
		return nil, ErrInvalidSessionID
	}

	session, err := s.sessions.GetSession(ctx, authContext.TenantID, authContext.UserID, parsedSessionID)
	if err != nil {
		return nil, err
	}
	if err := s.validateSessionScopeAccess(ctx, authContext, scopeFromSession(session)); err != nil {
		return nil, err
	}

	messages, err := s.sessions.ListMessages(ctx, authContext.TenantID, authContext.UserID, parsedSessionID)
	if err != nil {
		return nil, err
	}

	return toSessionDetail(session, messages)
}

func (s *Service) DeleteSession(ctx context.Context, sessionID string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	parsedSessionID, err := uuid.Parse(strings.TrimSpace(sessionID))
	if err != nil {
		return ErrInvalidSessionID
	}

	session, err := s.sessions.GetSession(ctx, authContext.TenantID, authContext.UserID, parsedSessionID)
	if err != nil {
		return err
	}
	if err := s.validateSessionScopeAccess(ctx, authContext, scopeFromSession(session)); err != nil {
		return err
	}

	if err := s.sessions.DeleteSession(ctx, authContext.TenantID, authContext.UserID, parsedSessionID); err != nil {
		return err
	}

	s.logSessionDeleted(ctx, authContext, parsedSessionID)
	return nil
}

func (s *Service) SendSessionMessage(ctx context.Context, sessionID string, request ChatRequest) (*ChatResponse, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, ErrInvalidSessionID
	}
	request.SessionID = &sessionID
	return s.Chat(ctx, request)
}

func (s *Service) resolveSession(ctx context.Context, authContext auth.Context, request ChatRequest, question string) (*ChatSession, bool, error) {
	if request.SessionID != nil && strings.TrimSpace(*request.SessionID) != "" {
		parsedSessionID, err := uuid.Parse(strings.TrimSpace(*request.SessionID))
		if err != nil {
			return nil, false, ErrInvalidSessionID
		}

		session, err := s.sessions.GetSession(ctx, authContext.TenantID, authContext.UserID, parsedSessionID)
		if err != nil {
			return nil, false, err
		}
		if err := s.validateSessionScopeAccess(ctx, authContext, scopeFromSession(session)); err != nil {
			return nil, false, err
		}
		return session, false, nil
	}

	scope := deriveSessionScope(request)
	if err := s.validateSessionScopeAccess(ctx, authContext, scope); err != nil {
		return nil, false, err
	}

	session := &ChatSession{
		TenantID:  authContext.TenantID,
		UserID:    authContext.UserID,
		Title:     sessionTitleFromQuestion(question),
		ScopeType: scope.ScopeType,
		ScopeID:   scope.ScopeID,
	}
	if err := s.sessions.CreateSession(ctx, session); err != nil {
		return nil, false, err
	}

	s.logSessionCreated(ctx, authContext, session.ID)
	return session, true, nil
}

func requestWithSessionScope(request ChatRequest, session *ChatSession) ChatRequest {
	merged := request
	if session.ScopeType == nil {
		return merged
	}

	switch *session.ScopeType {
	case ScopeTypeDocument:
		if session.ScopeID != nil && (merged.DocumentID == nil || strings.TrimSpace(*merged.DocumentID) == "") {
			id := session.ScopeID.String()
			merged.DocumentID = &id
		}
	case ScopeTypeFolder:
		if session.ScopeID != nil && (merged.FolderID == nil || strings.TrimSpace(*merged.FolderID) == "") {
			id := session.ScopeID.String()
			merged.FolderID = &id
		}
	case ScopeTypeTag:
		if session.ScopeID != nil && len(merged.TagIDs) == 0 {
			merged.TagIDs = []string{session.ScopeID.String()}
		}
	}

	return merged
}

func parseSessionScope(scopeType *string, scopeID *string) (sessionScope, error) {
	if scopeType == nil || strings.TrimSpace(*scopeType) == "" {
		all := ScopeTypeAll
		return sessionScope{ScopeType: &all, ScopeID: nil}, nil
	}

	normalized := strings.TrimSpace(*scopeType)
	if !isValidScopeType(normalized) {
		return sessionScope{}, ErrInvalidScopeType
	}

	scope := sessionScope{ScopeType: &normalized, ScopeID: nil}
	if scopeID != nil && strings.TrimSpace(*scopeID) != "" {
		parsedScopeID, err := uuid.Parse(strings.TrimSpace(*scopeID))
		if err != nil {
			return sessionScope{}, ErrInvalidScopeType
		}
		scope.ScopeID = &parsedScopeID
	}
	return scope, nil
}

func parseListScopeFilter(scopeType *string, scopeID *string) (*string, *uuid.UUID, error) {
	if scopeType == nil || strings.TrimSpace(*scopeType) == "" {
		return nil, nil, nil
	}

	normalized := strings.TrimSpace(*scopeType)
	if !isValidScopeType(normalized) {
		return nil, nil, ErrInvalidScopeType
	}

	var parsedScopeID *uuid.UUID
	if scopeID != nil && strings.TrimSpace(*scopeID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*scopeID))
		if err != nil {
			return nil, nil, ErrInvalidScopeType
		}
		parsedScopeID = &id
	}

	return &normalized, parsedScopeID, nil
}

func isValidScopeType(value string) bool {
	switch value {
	case ScopeTypeAll, ScopeTypeDocument, ScopeTypeFolder, ScopeTypeTag, ScopeTypeCollection, ScopeTypeWorkspace, ScopeTypeAgent:
		return true
	default:
		return false
	}
}

func toSessionSummary(session *ChatSession) ChatSessionSummaryResponse {
	return ChatSessionSummaryResponse{
		ID:        session.ID,
		Title:     session.Title,
		ScopeType: session.ScopeType,
		ScopeID:   session.ScopeID,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}
}

func toSessionDetail(session *ChatSession, messages []ChatMessage) (*ChatSessionDetailResponse, error) {
	response := &ChatSessionDetailResponse{
		ID:        session.ID,
		Title:     session.Title,
		ScopeType: session.ScopeType,
		ScopeID:   session.ScopeID,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
		Messages:  make([]ChatMessageResponse, 0, len(messages)),
	}

	for index := range messages {
		item, err := toMessageResponse(&messages[index])
		if err != nil {
			return nil, err
		}
		response.Messages = append(response.Messages, item)
	}

	return response, nil
}

func toMessageResponse(message *ChatMessage) (ChatMessageResponse, error) {
	response := ChatMessageResponse{
		ID:        message.ID,
		Role:      message.Role,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}

	if len(message.Sources) > 0 {
		var sources []ChatSource
		if err := json.Unmarshal(message.Sources, &sources); err != nil {
			return response, err
		}
		response.Sources = sources
	}

	return response, nil
}

func (s *Service) logSessionCreated(ctx context.Context, authContext auth.Context, sessionID uuid.UUID) {
	userID := authContext.UserID
	audit.LogWithActor(ctx, s.audit, authContext.TenantID, &userID, audit.CreateLogInput{
		TenantID:    authContext.TenantID,
		Action:      audit.ActionAIChatSessionCreated,
		EntityType:  audit.EntityTypeAIChatSession,
		EntityID:    &sessionID,
		Description: "AI chat session created",
	})
}

func (s *Service) logMessageSent(ctx context.Context, authContext auth.Context, sessionID uuid.UUID) {
	userID := authContext.UserID
	audit.LogWithActor(ctx, s.audit, authContext.TenantID, &userID, audit.CreateLogInput{
		TenantID:    authContext.TenantID,
		Action:      audit.ActionAIChatMessageSent,
		EntityType:  audit.EntityTypeAIChatSession,
		EntityID:    &sessionID,
		Description: "AI chat message sent",
	})
}

func (s *Service) logSessionDeleted(ctx context.Context, authContext auth.Context, sessionID uuid.UUID) {
	userID := authContext.UserID
	audit.LogWithActor(ctx, s.audit, authContext.TenantID, &userID, audit.CreateLogInput{
		TenantID:    authContext.TenantID,
		Action:      audit.ActionAIChatSessionDeleted,
		EntityType:  audit.EntityTypeAIChatSession,
		EntityID:    &sessionID,
		Description: "AI chat session deleted",
	})
}
