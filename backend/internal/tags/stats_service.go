package tags

import (
	"context"
	"errors"

	"fifriz/internal/auth"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *Service) GetTagStats(ctx context.Context, id string) (*TagStatsResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	tagID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidTagID
	}

	tag, err := s.repository.GetByID(ctx, authContext.TenantID, tagID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	docCount, err := s.repository.CountDocumentsByTag(ctx, authContext.TenantID, tagID)
	if err != nil {
		return nil, err
	}
	lastUsed, err := s.repository.LastUsedAt(ctx, authContext.TenantID, tagID)
	if err != nil {
		return nil, err
	}
	folders, err := s.repository.MostUsedFoldersForTag(ctx, authContext.TenantID, tagID, 5)
	if err != nil {
		return nil, err
	}
	related, err := s.repository.RelatedTags(ctx, authContext.TenantID, tagID, 5)
	if err != nil {
		return nil, err
	}

	return &TagStatsResponse{
		TagID:           tag.ID,
		DocumentCount:   docCount,
		LastUsedAt:      lastUsed,
		CreatedAt:       tag.CreatedAt,
		MostUsedFolders: folders,
		RelatedTags:     related,
	}, nil
}

func (s *Service) SearchTags(ctx context.Context, query string) ([]TagSearchResult, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.repository.SearchTags(ctx, authContext.TenantID, query, 20)
}

func (s *Service) ResolveTagByNameOrAlias(ctx context.Context, tenantID, userID uuid.UUID, name string) (*Tag, error) {
	validated, err := ValidateTagName(name)
	if err != nil {
		return nil, err
	}

	existing, err := s.repository.FindByName(ctx, tenantID, validated)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	byAlias, err := s.repository.FindTagByAlias(ctx, tenantID, validated)
	if err == nil {
		return byAlias, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return s.repository.FindOrCreateTag(ctx, tenantID, userID, validated)
}

func (s *Service) ListKnownTagTerms(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	if s == nil || s.repository == nil {
		return []string{}, nil
	}
	return s.repository.ListKnownTagTerms(ctx, tenantID)
}
