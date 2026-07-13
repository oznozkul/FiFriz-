package folders

import (
	"context"
	"errors"
	"time"

	"fifriz/internal/auth"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FolderMetadataResponse struct {
	FolderID            uuid.UUID  `json:"folderId"`
	DocumentCount       int64      `json:"documentCount"`
	SubFolderCount      int64      `json:"subFolderCount"`
	TotalSize           int64      `json:"totalSize"`
	StorageUsage        int64      `json:"storageUsage"`
	LastActivityAt      *time.Time `json:"lastActivityAt"`
	LastModifiedByUserID *uuid.UUID `json:"lastModifiedByUserId"`
	AIIndexed           bool       `json:"aiIndexed"`
}

func (s *Service) GetFolderMetadata(ctx context.Context, id string) (*FolderMetadataResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
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

	allFolders, err := s.repository.List(ctx, authContext.TenantID)
	if err != nil {
		return nil, err
	}
	scopeIDs := CollectFolderAndDescendantIDs(allFolders, folderID)

	meta, err := s.repository.GetFolderMetadata(ctx, authContext.TenantID, scopeIDs)
	if err != nil {
		return nil, err
	}

	lastActivity := meta.LastActivityAt
	if folder.LastActivityAt != nil && (lastActivity == nil || folder.LastActivityAt.After(*lastActivity)) {
		lastActivity = folder.LastActivityAt
	}

	return &FolderMetadataResponse{
		FolderID:             folder.ID,
		DocumentCount:        meta.DocumentCount,
		SubFolderCount:       meta.SubFolderCount,
		TotalSize:            meta.TotalSize,
		StorageUsage:         meta.StorageUsage,
		LastActivityAt:       lastActivity,
		LastModifiedByUserID: folder.LastModifiedByUserID,
		AIIndexed:            meta.AIIndexed,
	}, nil
}

func (s *Service) GetBreadcrumb(ctx context.Context, id string) ([]FolderResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	folderID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidFolderID
	}

	allFolders, err := s.repository.List(ctx, authContext.TenantID)
	if err != nil {
		return nil, err
	}

	path := BuildBreadcrumb(allFolders, folderID)
	if len(path) == 0 {
		return nil, ErrFolderNotFound
	}

	responses := make([]FolderResponse, 0, len(path))
	for _, item := range path {
		if !IsFolderVisible(&item, permissions.CanManageFolder(permissions.Role(authContext.Role))) {
			continue
		}
		responses = append(responses, toFolderResponse(item))
	}
	if len(responses) == 0 {
		return nil, ErrFolderNotFound
	}
	return responses, nil
}

func (s *Service) SearchFolders(ctx context.Context, query string) ([]FolderSearchResult, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.repository.SearchFolders(ctx, authContext.TenantID, query, 20)
}
