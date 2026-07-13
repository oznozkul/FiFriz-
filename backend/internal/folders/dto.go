package folders

import (
	"time"

	"github.com/google/uuid"
)

type CreateFolderRequest struct {
	Name        string  `json:"name"`
	ParentID    string  `json:"parentId"`
	Description string  `json:"description"`
	Icon        *string `json:"icon"`
	Color       *string `json:"color"`
}

type UpdateFolderRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Icon        *string `json:"icon"`
	Color       *string `json:"color"`
}

type MoveFolderRequest struct {
	ParentID string `json:"parentId"`
}

type FolderResponse struct {
	ID                   uuid.UUID        `json:"id"`
	TenantID             uuid.UUID        `json:"tenant_id"`
	ParentID             *uuid.UUID       `json:"parentId"`
	Name                 string           `json:"name"`
	Description          string           `json:"description"`
	Status               string           `json:"status"`
	LifecycleStatus      string           `json:"lifecycleStatus"`
	IsSystem             bool             `json:"isSystem"`
	IsLocked             bool             `json:"isLocked"`
	LockedAt             *time.Time       `json:"lockedAt"`
	LockedByUserID       *uuid.UUID       `json:"lockedByUserId"`
	InheritPermissions   bool             `json:"inheritPermissions"`
	Icon                 *string          `json:"icon"`
	Color                *string          `json:"color"`
	LastActivityAt       *time.Time       `json:"lastActivityAt"`
	LastModifiedByUserID *uuid.UUID       `json:"lastModifiedByUserId"`
	CreatedByUserID      uuid.UUID        `json:"created_by_user_id"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
	Children             []FolderResponse `json:"children,omitempty"`
}

func toFolderResponse(folder Folder) FolderResponse {
	return FolderResponse{
		ID:                   folder.ID,
		TenantID:             folder.TenantID,
		ParentID:             folder.ParentID,
		Name:                 folder.Name,
		Description:          folder.Description,
		Status:               NormalizeStatus(folder.Status),
		LifecycleStatus:      EffectiveLifecycleStatus(&folder),
		IsSystem:             folder.IsSystem,
		IsLocked:             folder.IsLocked,
		LockedAt:             folder.LockedAt,
		LockedByUserID:       folder.LockedByUserID,
		InheritPermissions:   folder.InheritPermissions,
		Icon:                 folder.Icon,
		Color:                folder.Color,
		LastActivityAt:       folder.LastActivityAt,
		LastModifiedByUserID: folder.LastModifiedByUserID,
		CreatedByUserID:      folder.CreatedByUserID,
		CreatedAt:            folder.CreatedAt,
		UpdatedAt:            folder.UpdatedAt,
	}
}

func toFolderTree(items []Folder, isAdmin bool) []FolderResponse {
	visible := make([]Folder, 0, len(items))
	for _, item := range items {
		if IsFolderVisible(&item, isAdmin) {
			visible = append(visible, item)
		}
	}

	responses := make([]FolderResponse, 0, len(visible))
	indexByID := make(map[uuid.UUID]int, len(visible))

	for _, item := range visible {
		response := toFolderResponse(item)
		response.Children = []FolderResponse{}
		indexByID[response.ID] = len(responses)
		responses = append(responses, response)
	}

	roots := make([]FolderResponse, 0)
	for index := range responses {
		parentID := responses[index].ParentID
		if parentID == nil {
			roots = append(roots, responses[index])
			continue
		}

		parentIndex, ok := indexByID[*parentID]
		if !ok {
			roots = append(roots, responses[index])
			continue
		}
		responses[parentIndex].Children = append(responses[parentIndex].Children, responses[index])
	}

	return rebuildTree(roots, responses)
}

func rebuildTree(nodes []FolderResponse, all []FolderResponse) []FolderResponse {
	byParent := make(map[uuid.UUID][]FolderResponse)
	for _, node := range all {
		if node.ParentID != nil {
			byParent[*node.ParentID] = append(byParent[*node.ParentID], node)
		}
	}

	var attach func([]FolderResponse) []FolderResponse
	attach = func(current []FolderResponse) []FolderResponse {
		for index := range current {
			current[index].Children = attach(byParent[current[index].ID])
			if len(current[index].Children) == 0 {
				current[index].Children = nil
			}
		}
		return current
	}

	return attach(nodes)
}
