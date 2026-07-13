package tags

import (
	"context"
	"errors"
	"strings"

	"fifriz/internal/auth"
	"fifriz/internal/cache"
	"fifriz/internal/events"
	domain "fifriz/internal/events/events"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTagNotFound         = errors.New("tag not found")
	ErrInvalidTagID        = errors.New("invalid tag id")
	ErrInvalidTagData      = errors.New("invalid tag data")
	ErrTagNameConflict     = errors.New("tag name conflict")
	ErrTagAliasConflict    = errors.New("tag alias conflict")
	ErrTagNameTooLong      = errors.New("tag name too long")
	ErrSystemTagProtected  = errors.New("system tag cannot be modified")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidTagIDs       = errors.New("invalid tag ids")
)

type Service struct {
	repository *Repository
	publisher  events.Publisher
	cache      *cache.Service
}

func NewService(repository *Repository, publisher events.Publisher, cacheService *cache.Service) *Service {
	return &Service{
		repository: repository,
		publisher:  publisher,
		cache:      cacheService,
	}
}

func (s *Service) List(ctx context.Context) ([]TagResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	isAdmin := permissions.CanManageTags(permissions.Role(authContext.Role))

	return cache.GetOrSetJSON(ctx, s.cache, cache.TagsKey(authContext.TenantID), cache.TTLTags, func() ([]TagResponse, error) {
		items, err := s.repository.List(ctx, authContext.TenantID)
		if err != nil {
			return nil, err
		}
		counts, err := s.repository.TagUsageCounts(ctx, authContext.TenantID)
		if err != nil {
			return nil, err
		}

		responses := make([]TagResponse, 0, len(items))
		for _, item := range items {
			if !IsTagVisible(&item, isAdmin) {
				continue
			}
			responses = append(responses, toTagResponse(item, counts[item.ID]))
		}
		return responses, nil
	})
}

func (s *Service) GetByID(ctx context.Context, id string) (*TagResponse, error) {
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
	if !IsTagVisible(tag, permissions.CanManageTags(permissions.Role(authContext.Role))) {
		return nil, ErrTagNotFound
	}

	count, _ := s.repository.CountDocumentsByTag(ctx, authContext.TenantID, tagID)
	response := toTagResponse(*tag, count)
	return &response, nil
}

func (s *Service) Create(ctx context.Context, input CreateTagRequest) (*TagResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if !permissions.CanCreateTags(permissions.Role(authContext.Role)) {
		return nil, ErrForbidden
	}

	name, err := ValidateTagName(input.Name)
	if err != nil {
		return nil, err
	}

	exists, err := s.repository.ExistsByName(ctx, authContext.TenantID, name, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTagNameConflict
	}

	tag := &Tag{
		TenantID:        authContext.TenantID,
		Name:            name,
		Status:          StatusActive,
		Color:           normalizeColor(input.Color),
		Icon:            normalizeIcon(input.Icon),
		CreatedByUserID: authContext.UserID,
	}

	if err := s.repository.Create(ctx, tag); err != nil {
		return nil, err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewTagCreated(authContext.TenantID, authContext.UserID, tag.ID, tag.Name))

	response := toTagResponse(*tag, 0)
	return &response, nil
}

func (s *Service) Update(ctx context.Context, id string, input UpdateTagRequest) (*TagResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if !permissions.CanManageTags(permissions.Role(authContext.Role)) {
		return nil, ErrForbidden
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
	if tag.IsSystem {
		return nil, ErrSystemTagProtected
	}

	name, err := ValidateTagName(input.Name)
	if err != nil {
		return nil, err
	}

	exists, err := s.repository.ExistsByName(ctx, authContext.TenantID, name, &tag.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTagNameConflict
	}

	oldStatus := tag.Status
	tag.Name = name
	tag.Color = normalizeColor(input.Color)
	tag.Icon = normalizeIcon(input.Icon)
	if strings.TrimSpace(input.Status) != "" {
		tag.Status = NormalizeStatus(input.Status)
	}

	if err := s.repository.Update(ctx, tag); err != nil {
		return nil, err
	}

	if oldStatus != tag.Status && tag.Status == StatusArchived {
		s.publisher.PublishAfterCommit(ctx, domain.NewTagArchived(authContext.TenantID, authContext.UserID, tag.ID, tag.Name))
	} else {
		s.publisher.PublishAfterCommit(ctx, domain.NewTagUpdated(authContext.TenantID, authContext.UserID, tag.ID, tag.Name))
	}

	count, _ := s.repository.CountDocumentsByTag(ctx, authContext.TenantID, tagID)
	response := toTagResponse(*tag, count)
	return &response, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}
	if !permissions.CanManageTags(permissions.Role(authContext.Role)) {
		return ErrForbidden
	}

	tagID, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidTagID
	}

	tag, err := s.repository.GetByID(ctx, authContext.TenantID, tagID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTagNotFound
		}
		return err
	}
	if tag.IsSystem {
		return ErrSystemTagProtected
	}

	if err := s.repository.DeleteByTagID(ctx, authContext.TenantID, tag.ID); err != nil {
		return err
	}
	if err := s.repository.Delete(ctx, tag); err != nil {
		return err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewTagDeleted(authContext.TenantID, authContext.UserID, tag.ID, tag.Name))
	return nil
}

func (s *Service) MapTagsForDocuments(ctx context.Context, tenantID uuid.UUID, documentIDs []uuid.UUID) (map[uuid.UUID][]TagBrief, error) {
	if s == nil || s.repository == nil || len(documentIDs) == 0 {
		return map[uuid.UUID][]TagBrief{}, nil
	}
	return s.repository.ListTagsByDocumentIDs(ctx, tenantID, documentIDs)
}

func (s *Service) ReplaceDocumentTags(ctx context.Context, tenantID uuid.UUID, documentID uuid.UUID, tagIDs []uuid.UUID) error {
	if s == nil || s.repository == nil {
		return nil
	}
	if err := s.repository.ValidateTagIDs(ctx, tenantID, tagIDs); err != nil {
		return err
	}
	return s.repository.ReplaceDocumentTags(ctx, tenantID, documentID, tagIDs)
}

func (s *Service) FindOrCreateTag(ctx context.Context, tenantID, userID uuid.UUID, name string) (*Tag, error) {
	return s.ResolveTagByNameOrAlias(ctx, tenantID, userID, name)
}

func (s *Service) AppendDocumentTag(ctx context.Context, tenantID, documentID, tagID uuid.UUID) error {
	if s == nil || s.repository == nil {
		return nil
	}
	if err := s.repository.ValidateTagIDs(ctx, tenantID, []uuid.UUID{tagID}); err != nil {
		return err
	}
	return s.repository.AppendDocumentTag(ctx, tenantID, documentID, tagID)
}

func (s *Service) ListTagNames(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	return s.ListKnownTagTerms(ctx, tenantID)
}

func (s *Service) ParseTagIDs(rawIDs []string) ([]uuid.UUID, error) {
	if len(rawIDs) == 0 {
		return nil, nil
	}

	unique := make(map[uuid.UUID]struct{}, len(rawIDs))
	tagIDs := make([]uuid.UUID, 0, len(rawIDs))
	for _, raw := range rawIDs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		tagID, err := uuid.Parse(raw)
		if err != nil {
			return nil, ErrInvalidTagID
		}
		if _, exists := unique[tagID]; exists {
			continue
		}
		unique[tagID] = struct{}{}
		tagIDs = append(tagIDs, tagID)
	}
	return tagIDs, nil
}

func normalizeColor(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func normalizeIcon(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
