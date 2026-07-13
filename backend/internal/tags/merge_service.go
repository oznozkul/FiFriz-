package tags

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

type MergeTagRequest struct {
	TargetTagID string `json:"targetTagId"`
}

func (s *Service) MergeTag(ctx context.Context, id string, input MergeTagRequest) (*TagResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if !permissions.CanManageTags(permissions.Role(authContext.Role)) {
		return nil, ErrForbidden
	}

	sourceID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidTagID
	}
	targetID, err := uuid.Parse(strings.TrimSpace(input.TargetTagID))
	if err != nil {
		return nil, ErrInvalidTagID
	}
	if sourceID == targetID {
		return nil, ErrInvalidTagData
	}

	source, err := s.repository.GetByID(ctx, authContext.TenantID, sourceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}
	if source.IsSystem {
		return nil, ErrSystemTagProtected
	}

	target, err := s.repository.GetByID(ctx, authContext.TenantID, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	if err := s.repository.MergeDocumentTags(ctx, authContext.TenantID, sourceID, targetID); err != nil {
		return nil, err
	}

	source.Status = StatusArchived
	if err := s.repository.Update(ctx, source); err != nil {
		return nil, err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewTagMerged(
		authContext.TenantID,
		authContext.UserID,
		source.ID,
		target.ID,
		source.Name,
		target.Name,
	))

	count, _ := s.repository.CountDocumentsByTag(ctx, authContext.TenantID, target.ID)
	response := toTagResponse(*target, count)
	return &response, nil
}
