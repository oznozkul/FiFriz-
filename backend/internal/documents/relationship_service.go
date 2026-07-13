package documents

import (
	"context"
	"errors"
	"strings"

	"fifriz/internal/applog"
	"fifriz/internal/auth"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrDocumentRelationshipNotFound = errors.New("document relationship not found")
	ErrInvalidRelationshipID        = errors.New("invalid relationship id")
	ErrInvalidRelationshipType      = errors.New("invalid relationship type")
	ErrSelfDocumentRelationship     = errors.New("document cannot be related to itself")
	ErrDuplicateDocumentRelationship = errors.New("document relationship already exists")
)

func (s *Service) ListRelationships(ctx context.Context, documentIDRaw string) (*DocumentRelationshipsResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	rows, err := s.repository.ListRelationshipsByDocument(ctx, authContext.TenantID, documentID)
	if err != nil {
		return nil, err
	}

	response := buildRelationshipsResponse(documentID, rows)
	return &response, nil
}

func (s *Service) CreateRelationship(ctx context.Context, documentIDRaw string, input CreateDocumentRelationshipRequest) (*DocumentRelationshipEdge, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	sourceDocumentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	targetDocumentID, err := uuid.Parse(strings.TrimSpace(input.TargetDocumentID))
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	relationshipType := strings.TrimSpace(input.RelationshipType)
	if !IsAllowedRelationshipType(relationshipType) {
		return nil, ErrInvalidRelationshipType
	}

	if sourceDocumentID == targetDocumentID {
		return nil, ErrSelfDocumentRelationship
	}

	sourceDocument, err := s.repository.GetByID(ctx, authContext.TenantID, sourceDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireEditAccess(ctx, authContext, sourceDocument); err != nil {
		return nil, err
	}

	targetDocument, err := s.repository.GetByID(ctx, authContext.TenantID, targetDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, targetDocument); err != nil {
		return nil, err
	}

	exists, err := s.repository.RelationshipExists(ctx, authContext.TenantID, sourceDocumentID, targetDocumentID, relationshipType)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateDocumentRelationship
	}

	relationship := &DocumentRelationship{
		TenantID:         authContext.TenantID,
		SourceDocumentID: sourceDocumentID,
		TargetDocumentID: targetDocumentID,
		RelationshipType: relationshipType,
		CreatedByUserID:  authContext.UserID,
	}
	if err := s.repository.CreateRelationship(ctx, relationship); err != nil {
		return nil, err
	}

	rows, err := s.repository.ListRelationshipsByDocument(ctx, authContext.TenantID, sourceDocumentID)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		if row.ID == relationship.ID {
			response := buildRelationshipsResponse(sourceDocumentID, []RelationshipListRow{row})
			if len(response.Items) > 0 {
				item := response.Items[0]
				return &item, nil
			}
		}
	}

	return &DocumentRelationshipEdge{
		ID:               relationship.ID,
		RelationshipType: relationship.RelationshipType,
		Direction:        "outgoing",
		SourceDocumentID: relationship.SourceDocumentID,
		TargetDocumentID: relationship.TargetDocumentID,
		RelatedDocument: RelatedDocumentNode{
			ID:               targetDocument.ID,
			OriginalFileName: targetDocument.OriginalFileName,
			FileExtension:    targetDocument.FileExtension,
			MimeType:         targetDocument.MimeType,
			FolderID:         targetDocument.FolderID,
			FolderName:       targetDocument.FolderName,
		},
		CreatedByUserID: relationship.CreatedByUserID,
		CreatedAt:       relationship.CreatedAt,
	}, nil
}

func (s *Service) DeleteRelationship(ctx context.Context, relationshipIDRaw string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	relationshipID, err := uuid.Parse(relationshipIDRaw)
	if err != nil {
		return ErrInvalidRelationshipID
	}

	relationship, err := s.repository.GetRelationshipByID(ctx, authContext.TenantID, relationshipID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDocumentRelationshipNotFound
		}
		return err
	}

	sourceDocument, err := s.repository.GetByID(ctx, authContext.TenantID, relationship.SourceDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDocumentRelationshipNotFound
		}
		return err
	}
	if err := s.requireViewAccess(ctx, authContext, sourceDocument); err != nil {
		return err
	}
	if err := s.requireManageRelationshipAccess(ctx, authContext, sourceDocument, relationship.CreatedByUserID); err != nil {
		return err
	}

	return s.repository.DeleteRelationship(ctx, authContext.TenantID, relationshipID)
}

func (s *Service) requireManageRelationshipAccess(ctx context.Context, authContext auth.Context, sourceDocument *Document, createdByUserID uuid.UUID) error {
	subject := permissions.SubjectFromAuth(authContext)
	if permissions.CanManageDocumentComment(subject, createdByUserID) {
		return nil
	}
	if err := s.requireEditAccess(ctx, authContext, sourceDocument); err == nil {
		return nil
	}
	applog.PermissionDenied("manage_document_relationship", sourceDocument.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
	return ErrForbidden
}
