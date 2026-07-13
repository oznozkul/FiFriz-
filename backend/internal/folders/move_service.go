package folders

import (
	"context"
	"errors"

	"fifriz/internal/auth"
	domain "fifriz/internal/events/events"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
)

var (
	ErrFolderMoveCycle       = errors.New("folder cannot be moved into its own descendant")
	ErrFolderMoveIntoSelf    = errors.New("folder cannot be moved into itself")
	ErrFolderMoveCrossTenant = errors.New("folder cannot be moved across tenants")
)

func (s *Service) MoveFolder(ctx context.Context, id string, input MoveFolderRequest) (*FolderResponse, error) {
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

	allFolders, err := s.repository.List(ctx, authContext.TenantID)
	if err != nil {
		return nil, err
	}

	var folder *Folder
	for index := range allFolders {
		if allFolders[index].ID == folderID {
			folder = &allFolders[index]
			break
		}
	}
	if folder == nil {
		return nil, ErrFolderNotFound
	}
	if folder.IsSystem {
		return nil, ErrSystemFolderProtected
	}
	if err := s.requireFolderUnlocked(folder); err != nil {
		return nil, err
	}

	newParentID, err := s.parseParentID(ctx, authContext.TenantID, input.ParentID)
	if err != nil {
		return nil, err
	}
	if newParentID != nil && *newParentID == folderID {
		return nil, ErrFolderMoveIntoSelf
	}
	if newParentID != nil && IsDescendant(allFolders, folderID, *newParentID) {
		return nil, ErrFolderMoveCycle
	}

	exists, err := s.repository.ExistsByName(ctx, authContext.TenantID, newParentID, folder.Name, &folder.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrFolderNameConflict
	}

	oldParentID := folder.ParentID
	folder.ParentID = newParentID
	folder.LastModifiedByUserID = &authContext.UserID
	touchActivity(folder)

	if err := s.repository.UpdateParent(ctx, folder); err != nil {
		return nil, err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewFolderMoved(
		authContext.TenantID,
		authContext.UserID,
		folder.ID,
		folder.Name,
		parentIDString(oldParentID),
		parentIDString(newParentID),
	))

	response := toFolderResponse(*folder)
	return &response, nil
}

func parentIDString(parentID *uuid.UUID) string {
	if parentID == nil {
		return ""
	}
	return parentID.String()
}
