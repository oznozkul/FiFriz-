package ai

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"fifriz/internal/auth"
	"fifriz/internal/documents"
	"fifriz/internal/folders"
	"fifriz/internal/tags"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const sessionTitleMaxRunes = 60

type sessionScope struct {
	ScopeType *string
	ScopeID   *uuid.UUID
}

func deriveSessionScope(request ChatRequest) sessionScope {
	if request.DocumentID != nil && strings.TrimSpace(*request.DocumentID) != "" {
		documentID, err := uuid.Parse(strings.TrimSpace(*request.DocumentID))
		if err == nil {
			scopeType := ScopeTypeDocument
			return sessionScope{ScopeType: &scopeType, ScopeID: &documentID}
		}
	}

	if request.FolderID != nil && strings.TrimSpace(*request.FolderID) != "" {
		folderID, err := uuid.Parse(strings.TrimSpace(*request.FolderID))
		if err == nil {
			scopeType := ScopeTypeFolder
			return sessionScope{ScopeType: &scopeType, ScopeID: &folderID}
		}
	}

	if len(request.TagIDs) > 0 {
		scopeType := ScopeTypeTag
		if len(request.TagIDs) == 1 {
			tagID, err := uuid.Parse(strings.TrimSpace(request.TagIDs[0]))
			if err == nil {
				return sessionScope{ScopeType: &scopeType, ScopeID: &tagID}
			}
		}
		return sessionScope{ScopeType: &scopeType, ScopeID: nil}
	}

	if request.CollectionKey != nil && strings.TrimSpace(*request.CollectionKey) != "" {
		scopeType := ScopeTypeCollection
		return sessionScope{ScopeType: &scopeType, ScopeID: nil}
	}

	scopeType := ScopeTypeAll
	return sessionScope{ScopeType: &scopeType, ScopeID: nil}
}

func sessionTitleFromQuestion(question string) string {
	question = strings.Join(strings.Fields(strings.TrimSpace(question)), " ")
	if question == "" {
		return "Yeni sohbet"
	}
	if utf8.RuneCountInString(question) <= sessionTitleMaxRunes {
		return question
	}
	runes := []rune(question)
	return string(runes[:sessionTitleMaxRunes-3]) + "..."
}

func (s *Service) validateSessionScopeAccess(ctx context.Context, authContext auth.Context, scope sessionScope) error {
	if scope.ScopeType == nil {
		return nil
	}

	switch *scope.ScopeType {
	case ScopeTypeAll, ScopeTypeCollection, ScopeTypeWorkspace:
		return nil
	case ScopeTypeAgent:
		if scope.ScopeID == nil {
			return nil
		}
		document, err := s.rag.documents.GetByID(ctx, authContext.TenantID, *scope.ScopeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return documents.ErrForbidden
			}
			return err
		}
		return requireViewAccess(ctx, s.rag.shareAccess, authContext, document)
	case ScopeTypeDocument:
		if scope.ScopeID == nil {
			return nil
		}
		document, err := s.rag.documents.GetByID(ctx, authContext.TenantID, *scope.ScopeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return documents.ErrForbidden
			}
			return err
		}
		return requireViewAccess(ctx, s.rag.shareAccess, authContext, document)
	case ScopeTypeFolder:
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
	case ScopeTypeTag:
		if scope.ScopeID == nil || s.tags == nil {
			return nil
		}
		_, err := s.tags.GetByID(ctx, authContext.TenantID, *scope.ScopeID)
		if err != nil {
			if errors.Is(err, tags.ErrTagNotFound) {
				return tags.ErrTagNotFound
			}
			return err
		}
		return nil
	default:
		return nil
	}
}

func scopeFromSession(session *ChatSession) sessionScope {
	return sessionScope{
		ScopeType: session.ScopeType,
		ScopeID:   session.ScopeID,
	}
}
