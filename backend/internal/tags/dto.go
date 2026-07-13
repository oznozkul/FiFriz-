package tags

import (
	"time"

	"github.com/google/uuid"
)

type TagResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	LifecycleStatus string    `json:"lifecycleStatus"`
	IsSystem        bool      `json:"isSystem"`
	Color           *string   `json:"color"`
	Icon            *string   `json:"icon"`
	DocumentCount   int64     `json:"documentCount"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type TagBrief struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color *string   `json:"color"`
	Icon  *string   `json:"icon"`
}

type CreateTagRequest struct {
	Name  string  `json:"name"`
	Color string  `json:"color"`
	Icon  *string `json:"icon"`
}

type UpdateTagRequest struct {
	Name   string  `json:"name"`
	Color  string  `json:"color"`
	Status string  `json:"status"`
	Icon   *string `json:"icon"`
}

type SetDocumentTagsRequest struct {
	TagIDs []string `json:"tagIds"`
}

func toTagResponse(tag Tag, documentCount int64) TagResponse {
	return TagResponse{
		ID:              tag.ID,
		Name:            tag.Name,
		Status:          NormalizeStatus(tag.Status),
		LifecycleStatus: EffectiveLifecycleStatus(&tag),
		IsSystem:        tag.IsSystem,
		Color:           tag.Color,
		Icon:            tag.Icon,
		DocumentCount:   documentCount,
		CreatedAt:       tag.CreatedAt,
		UpdatedAt:       tag.UpdatedAt,
	}
}

func toTagBrief(tag Tag) TagBrief {
	return TagBrief{
		ID:    tag.ID,
		Name:  tag.Name,
		Color: tag.Color,
		Icon:  tag.Icon,
	}
}
