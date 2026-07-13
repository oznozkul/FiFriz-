package smartcollections

import (
	"context"
	"time"

	"fifriz/internal/auth"
	"fifriz/internal/documents"
	"fifriz/internal/events"
	domain "fifriz/internal/events/events"

	"github.com/google/uuid"
)

type DocumentMapper interface {
	ToResponses(ctx context.Context, tenantID uuid.UUID, items []documents.Document) ([]documents.DocumentResponse, error)
}

type Service struct {
	repository *Repository
	documents  DocumentMapper
	publisher  events.Publisher
}

func NewService(repository *Repository, documentMapper DocumentMapper, publisher events.Publisher) *Service {
	return &Service{
		repository: repository,
		documents:  documentMapper,
		publisher:  publisher,
	}
}

func (s *Service) ListCollections(ctx context.Context) ([]CollectionResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	accessUserID := memberAccessUserID(authContext)
	responses := make([]CollectionResponse, 0, len(definitions))

	for _, definition := range definitions {
		count, err := s.repository.Count(ctx, authContext.TenantID, authContext.UserID, accessUserID, definition.Key, now)
		if err != nil {
			return nil, err
		}
		responses = append(responses, CollectionResponse{
			Key:         definition.Key,
			Title:       definition.Title,
			Description: definition.Description,
			Icon:        definition.Icon,
			Count:       count,
		})
	}

	return responses, nil
}

func (s *Service) ListDocuments(ctx context.Context, key string) ([]documents.DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if !IsValidKey(key) {
		return nil, ErrInvalidCollectionKey
	}

	definition, ok := GetDefinition(key)
	if !ok {
		return nil, ErrInvalidCollectionKey
	}

	now := time.Now().UTC()
	accessUserID := memberAccessUserID(authContext)
	items, err := s.repository.List(ctx, authContext.TenantID, authContext.UserID, accessUserID, key, now)
	if err != nil {
		return nil, err
	}

	responses, err := s.documents.ToResponses(ctx, authContext.TenantID, items)
	if err != nil {
		return nil, err
	}

	s.publisher.Publish(ctx, domain.NewSmartCollectionOpened(
		authContext.TenantID,
		authContext.UserID,
		definition.Key,
		definition.Title,
		"/collections/"+definition.Key,
	))

	return responses, nil
}

func memberAccessUserID(authContext auth.Context) *uuid.UUID {
	if documents.HasTenantWideDocumentAccess(authContext.Role) {
		return nil
	}
	userID := authContext.UserID
	return &userID
}
