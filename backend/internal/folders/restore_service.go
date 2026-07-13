package folders

import (
	"context"
	"errors"

	"fifriz/internal/auth"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *Service) RestoreFolder(ctx context.Context, id string) (*FolderResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if !permissions.CanManageFolder(permissions.Role(authContext.Role)) {
		return nil, ErrForbidden
	}

	folderID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidFolderID
	}

	folder, err := s.repository.GetByIDUnscoped(ctx, authContext.TenantID, folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFolderNotFound
		}
		return nil, err
	}
	if !folder.DeletedAt.Valid {
		response := toFolderResponse(*folder)
		return &response, nil
	}

	allDeleted, err := s.repository.ListIncludingDeleted(ctx, authContext.TenantID)
	if err != nil {
		return nil, err
	}

	restoreIDs := CollectFolderAndDescendantIDs(allDeleted, folderID)
	// only restore descendants that are also soft-deleted
	filtered := make([]uuid.UUID, 0, len(restoreIDs))
	for _, candidateID := range restoreIDs {
		for _, item := range allDeleted {
			if item.ID == candidateID && item.DeletedAt.Valid {
				filtered = append(filtered, candidateID)
				break
			}
		}
	}

	if err := s.repository.RestoreIDs(ctx, authContext.TenantID, filtered); err != nil {
		return nil, err
	}

	restored, err := s.repository.GetByID(ctx, authContext.TenantID, folderID)
	if err != nil {
		return nil, err
	}
	response := toFolderResponse(*restored)
	return &response, nil
}
