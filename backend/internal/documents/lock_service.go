package documents

import (
	"context"
	"errors"
	"time"

	"fifriz/internal/auth"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *Service) requireDocumentUnlocked(document *Document) error {
	if document != nil && document.IsLocked {
		return ErrDocumentLocked
	}
	return nil
}

func canManageDocumentLock(authContext auth.Context) bool {
	return authContext.Role == auth.TenantRoleOwner || authContext.Role == auth.TenantRoleAdmin
}

func (s *Service) LockDocument(ctx context.Context, id string) (*DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if !canManageDocumentLock(authContext) {
		return nil, ErrForbidden
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	now := time.Now()
	document.IsLocked = true
	document.LockedAt = &now
	document.LockedByUserID = &authContext.UserID

	if err := s.repository.Update(ctx, document); err != nil {
		return nil, err
	}

	return s.toResponsePtr(ctx, authContext.TenantID, *document)
}

func (s *Service) UnlockDocument(ctx context.Context, id string) (*DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if !canManageDocumentLock(authContext) {
		return nil, ErrForbidden
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	document.IsLocked = false
	document.LockedAt = nil
	document.LockedByUserID = nil

	if err := s.repository.Update(ctx, document); err != nil {
		return nil, err
	}

	return s.toResponsePtr(ctx, authContext.TenantID, *document)
}
