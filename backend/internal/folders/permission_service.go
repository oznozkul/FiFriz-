package folders

import (
	"context"
	"errors"
	"strings"

	"fifriz/internal/auth"
	domain "fifriz/internal/events/events"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FolderPermissionEntry struct {
	UserID     string `json:"userId"`
	Permission string `json:"permission"`
}

type UpdateFolderPermissionsRequest struct {
	InheritPermissions *bool                   `json:"inheritPermissions"`
	Permissions        []FolderPermissionEntry `json:"permissions"`
}

type FolderPermissionsResponse struct {
	FolderID           uuid.UUID               `json:"folderId"`
	InheritPermissions bool                    `json:"inheritPermissions"`
	Permissions        []FolderPermissionEntry `json:"permissions"`
}

func (s *Service) GetFolderPermissions(ctx context.Context, id string) (*FolderPermissionsResponse, error) {
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

	folder, err := s.repository.GetByID(ctx, authContext.TenantID, folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFolderNotFound
		}
		return nil, err
	}

	items, err := s.repository.ListFolderPermissions(ctx, authContext.TenantID, folderID)
	if err != nil {
		return nil, err
	}

	entries := make([]FolderPermissionEntry, 0, len(items))
	for _, item := range items {
		entries = append(entries, FolderPermissionEntry{
			UserID:     item.UserID.String(),
			Permission: item.Permission,
		})
	}

	return &FolderPermissionsResponse{
		FolderID:           folder.ID,
		InheritPermissions: folder.InheritPermissions,
		Permissions:        entries,
	}, nil
}

func (s *Service) UpdateFolderPermissions(ctx context.Context, id string, input UpdateFolderPermissionsRequest) (*FolderPermissionsResponse, error) {
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
	if err := s.requireFolderUnlocked(folder); err != nil {
		return nil, err
	}

	if input.InheritPermissions != nil {
		folder.InheritPermissions = *input.InheritPermissions
	}
	folder.LastModifiedByUserID = &authContext.UserID

	if err := s.repository.SetInheritPermissions(ctx, folder); err != nil {
		return nil, err
	}

	records := make([]FolderPermission, 0, len(input.Permissions))
	for _, entry := range input.Permissions {
		userID, parseErr := uuid.Parse(strings.TrimSpace(entry.UserID))
		if parseErr != nil {
			return nil, ErrInvalidFolderData
		}
		permission := strings.TrimSpace(strings.ToLower(entry.Permission))
		if !IsValidFolderPermission(permission) {
			return nil, ErrInvalidFolderData
		}
		records = append(records, FolderPermission{
			TenantID:   authContext.TenantID,
			FolderID:   folderID,
			UserID:     userID,
			Permission: permission,
		})
	}

	if err := s.repository.ReplaceFolderPermissions(ctx, authContext.TenantID, folderID, records); err != nil {
		return nil, err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewFolderPermissionChanged(
		authContext.TenantID,
		authContext.UserID,
		folder.ID,
		folder.Name,
		folder.InheritPermissions,
		len(records),
	))

	return s.GetFolderPermissions(ctx, id)
}
