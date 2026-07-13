package folders

import (
	"context"
	"errors"
	"time"

	"fifriz/internal/auth"
	domain "fifriz/internal/events/events"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrFolderLocked = errors.New("folder is locked")

func (s *Service) requireFolderUnlocked(folder *Folder) error {
	if IsFolderLocked(folder) {
		return ErrFolderLocked
	}
	return nil
}

func canManageFolderLock(authContext auth.Context) bool {
	return authContext.Role == auth.TenantRoleOwner || authContext.Role == auth.TenantRoleAdmin
}

func (s *Service) LockFolder(ctx context.Context, id string) (*FolderResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if !canManageFolderLock(authContext) {
		return nil, ErrForbidden
	}

	folderID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidFolderID
	}

	folder, err := s.repository.GetByID(ctx, authContext.TenantID, folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFolderNotFound
		}
		return nil, err
	}
	if folder.IsSystem {
		return nil, ErrSystemFolderProtected
	}

	now := time.Now()
	folder.IsLocked = true
	folder.LockedAt = &now
	folder.LockedByUserID = &authContext.UserID
	folder.LastModifiedByUserID = &authContext.UserID
	touchActivity(folder)

	if err := s.repository.Update(ctx, folder); err != nil {
		return nil, err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewFolderLocked(authContext.TenantID, authContext.UserID, folder.ID, folder.Name))
	response := toFolderResponse(*folder)
	return &response, nil
}

func (s *Service) UnlockFolder(ctx context.Context, id string) (*FolderResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if !canManageFolderLock(authContext) {
		return nil, ErrForbidden
	}

	folderID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidFolderID
	}

	folder, err := s.repository.GetByID(ctx, authContext.TenantID, folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFolderNotFound
		}
		return nil, err
	}

	folder.IsLocked = false
	folder.LockedAt = nil
	folder.LockedByUserID = nil
	folder.LastModifiedByUserID = &authContext.UserID
	touchActivity(folder)

	if err := s.repository.Update(ctx, folder); err != nil {
		return nil, err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewFolderUnlocked(authContext.TenantID, authContext.UserID, folder.ID, folder.Name))
	response := toFolderResponse(*folder)
	return &response, nil
}
