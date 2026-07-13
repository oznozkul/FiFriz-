package tags

import (
	"context"
	"errors"
	"time"

	"fifriz/internal/auth"
	domain "fifriz/internal/events/events"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateTagAliasRequest struct {
	Alias string `json:"alias"`
}

type TagAliasResponse struct {
	ID        uuid.UUID `json:"id"`
	TagID     uuid.UUID `json:"tagId"`
	Alias     string    `json:"alias"`
	CreatedAt string    `json:"createdAt"`
}

func (s *Service) ListTagAliases(ctx context.Context, tagID string) ([]TagAliasResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedTagID, err := uuid.Parse(tagID)
	if err != nil {
		return nil, ErrInvalidTagID
	}
	if _, err := s.repository.GetByID(ctx, authContext.TenantID, parsedTagID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	items, err := s.repository.ListAliases(ctx, authContext.TenantID, parsedTagID)
	if err != nil {
		return nil, err
	}

	responses := make([]TagAliasResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, TagAliasResponse{
			ID:        item.ID,
			TagID:     item.TagID,
			Alias:     item.Alias,
			CreatedAt: item.CreatedAt.Format(time.RFC3339),
		})
	}
	return responses, nil
}

func (s *Service) CreateTagAlias(ctx context.Context, tagID string, input CreateTagAliasRequest) (*TagAliasResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if !permissions.CanManageTags(permissions.Role(authContext.Role)) {
		return nil, ErrForbidden
	}

	parsedTagID, err := uuid.Parse(tagID)
	if err != nil {
		return nil, ErrInvalidTagID
	}
	tag, err := s.repository.GetByID(ctx, authContext.TenantID, parsedTagID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	alias, err := ValidateAlias(input.Alias)
	if err != nil {
		return nil, err
	}
	if exists, err := s.repository.ExistsByName(ctx, authContext.TenantID, alias, nil); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrTagNameConflict
	}
	if exists, err := s.repository.AliasExists(ctx, authContext.TenantID, alias, nil); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrTagAliasConflict
	}

	record := &TagAlias{
		TenantID: authContext.TenantID,
		TagID:    tag.ID,
		Alias:    alias,
	}
	if err := s.repository.CreateAlias(ctx, record); err != nil {
		return nil, err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewTagAliasCreated(authContext.TenantID, authContext.UserID, tag.ID, tag.Name, alias))

	return &TagAliasResponse{
		ID:        record.ID,
		TagID:     record.TagID,
		Alias:     record.Alias,
		CreatedAt: record.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) DeleteTagAlias(ctx context.Context, tagID, aliasID string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}
	if !permissions.CanManageTags(permissions.Role(authContext.Role)) {
		return ErrForbidden
	}

	parsedTagID, err := uuid.Parse(tagID)
	if err != nil {
		return ErrInvalidTagID
	}
	parsedAliasID, err := uuid.Parse(aliasID)
	if err != nil {
		return ErrInvalidTagID
	}

	tag, err := s.repository.GetByID(ctx, authContext.TenantID, parsedTagID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTagNotFound
		}
		return err
	}

	aliases, err := s.repository.ListAliases(ctx, authContext.TenantID, parsedTagID)
	if err != nil {
		return err
	}
	var deletedAlias string
	for _, item := range aliases {
		if item.ID == parsedAliasID {
			deletedAlias = item.Alias
			break
		}
	}
	if deletedAlias == "" {
		return ErrTagNotFound
	}

	if err := s.repository.DeleteAlias(ctx, authContext.TenantID, parsedAliasID); err != nil {
		return err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewTagAliasDeleted(authContext.TenantID, authContext.UserID, tag.ID, tag.Name, deletedAlias))
	return nil
}
