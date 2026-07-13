package sharing

import (
	"context"
	"errors"
	"strings"

	"fifriz/internal/auth"
	"fifriz/internal/documents"
	domain "fifriz/internal/events/events"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidInternalShareID   = errors.New("invalid internal share id")
	ErrInternalShareNotFound    = errors.New("internal share not found")
	ErrDuplicateInternalShare   = errors.New("duplicate internal share")
	ErrCannotShareWithSelf      = errors.New("cannot share with self")
	ErrShareUserNotInTenant     = errors.New("share user not in tenant")
	ErrInvalidInternalPermission = errors.New("invalid internal share permission")
)

func (s *Service) ListInternalShares(ctx context.Context, documentID string) ([]InternalShareResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, err := uuid.Parse(documentID)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, parsedDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	canManage, err := s.access.CanManageInternalShares(ctx, authContext, document)
	if err != nil {
		return nil, err
	}
	if !canManage {
		return nil, ErrForbidden
	}

	shares, err := s.repository.ListInternalSharesByDocument(ctx, authContext.TenantID, parsedDocumentID)
	if err != nil {
		return nil, err
	}

	return s.toInternalShareResponses(ctx, shares)
}

func (s *Service) CreateInternalShare(ctx context.Context, documentID string, input CreateInternalShareRequest) (*InternalShareResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, err := uuid.Parse(documentID)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	sharedWithUserID, err := uuid.Parse(strings.TrimSpace(input.SharedWithUserID))
	if err != nil {
		return nil, errors.New("invalid shared with user id")
	}

	permission := InternalSharePermission(strings.TrimSpace(string(input.Permission)))
	if !isValidInternalSharePermission(permission) {
		return nil, ErrInvalidInternalPermission
	}

	if sharedWithUserID == authContext.UserID {
		return nil, ErrCannotShareWithSelf
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, parsedDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	canCreate, err := s.access.CanCreateInternalShare(ctx, authContext, document)
	if err != nil {
		return nil, err
	}
	if !canCreate {
		return nil, ErrForbidden
	}

	if _, err := s.authRepository.GetMembership(ctx, sharedWithUserID, authContext.TenantID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShareUserNotInTenant
		}
		return nil, err
	}

	exists, err := s.repository.HasActiveInternalShare(ctx, authContext.TenantID, parsedDocumentID, sharedWithUserID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateInternalShare
	}

	share := &DocumentShare{
		TenantID:         authContext.TenantID,
		DocumentID:       document.ID,
		SharedWithUserID: sharedWithUserID,
		SharedByUserID:   authContext.UserID,
		ShareType:        ShareTypeInternalUser,
		Permission:       permission,
	}

	if err := s.repository.CreateInternalShare(ctx, share); err != nil {
		return nil, err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewDocumentShared(
		authContext.TenantID,
		authContext.UserID,
		document.ID,
		sharedWithUserID,
		document.OriginalFileName,
	))

	responses, err := s.toInternalShareResponses(ctx, []DocumentShare{*share})
	if err != nil || len(responses) == 0 {
		return nil, err
	}
	return &responses[0], nil
}

func (s *Service) UpdateInternalShare(ctx context.Context, documentID, shareID string, input UpdateInternalShareRequest) (*InternalShareResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, err := uuid.Parse(documentID)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	parsedShareID, err := uuid.Parse(shareID)
	if err != nil {
		return nil, ErrInvalidInternalShareID
	}

	permission := InternalSharePermission(strings.TrimSpace(string(input.Permission)))
	if !isValidInternalSharePermission(permission) {
		return nil, ErrInvalidInternalPermission
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, parsedDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	share, err := s.repository.GetInternalShareByID(ctx, authContext.TenantID, parsedShareID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInternalShareNotFound
		}
		return nil, err
	}
	if share.DocumentID != parsedDocumentID {
		return nil, ErrInternalShareNotFound
	}

	if !canModifyInternalShare(ctx, s, authContext, document, share) {
		return nil, ErrForbidden
	}

	share.Permission = permission
	if err := s.repository.UpdateInternalShare(ctx, share); err != nil {
		return nil, err
	}

	responses, err := s.toInternalShareResponses(ctx, []DocumentShare{*share})
	if err != nil || len(responses) == 0 {
		return nil, err
	}
	return &responses[0], nil
}

func (s *Service) DeleteInternalShare(ctx context.Context, documentID, shareID string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	parsedDocumentID, err := uuid.Parse(documentID)
	if err != nil {
		return ErrInvalidDocumentID
	}

	parsedShareID, err := uuid.Parse(shareID)
	if err != nil {
		return ErrInvalidInternalShareID
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, parsedDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDocumentNotFound
		}
		return err
	}

	share, err := s.repository.GetInternalShareByID(ctx, authContext.TenantID, parsedShareID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInternalShareNotFound
		}
		return err
	}
	if share.DocumentID != parsedDocumentID {
		return ErrInternalShareNotFound
	}

	if !canRemoveInternalShare(ctx, s, authContext, document, share) {
		return ErrForbidden
	}

	if err := s.repository.SoftDeleteInternalShare(ctx, authContext.TenantID, parsedShareID); err != nil {
		return err
	}

	return nil
}

func (s *Service) ListSharedWithMe(ctx context.Context) ([]SharedWithMeResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	shares, err := s.repository.ListInternalSharesForUser(ctx, authContext.TenantID, authContext.UserID)
	if err != nil {
		return nil, err
	}

	responses := make([]SharedWithMeResponse, 0, len(shares))
	for _, share := range shares {
		document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, share.DocumentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}

		sharedByUser, err := s.authRepository.GetUserByID(ctx, share.SharedByUserID)
		if err != nil {
			return nil, err
		}

		responses = append(responses, SharedWithMeResponse{
			ShareID:          share.ID.String(),
			DocumentID:       share.DocumentID.String(),
			DocumentName:     document.OriginalFileName,
			SharedByUserID:   share.SharedByUserID.String(),
			SharedByFullName: sharedByUser.FullName,
			Permission:       share.Permission,
			CreatedAt:        share.CreatedAt,
		})
	}

	return responses, nil
}

func canModifyInternalShare(ctx context.Context, service *Service, authContext auth.Context, document *documents.Document, share *DocumentShare) bool {
	if authContext.Role == auth.TenantRoleOwner || authContext.Role == auth.TenantRoleAdmin {
		return true
	}
	if share.SharedByUserID == authContext.UserID {
		return true
	}
	if document.UploadedByUserID == authContext.UserID {
		return false
	}
	allowed, err := service.access.CanManageInternalShares(ctx, authContext, document)
	return err == nil && allowed
}

func canRemoveInternalShare(ctx context.Context, service *Service, authContext auth.Context, document *documents.Document, share *DocumentShare) bool {
	return canModifyInternalShare(ctx, service, authContext, document, share)
}

func (s *Service) toInternalShareResponses(ctx context.Context, shares []DocumentShare) ([]InternalShareResponse, error) {
	responses := make([]InternalShareResponse, 0, len(shares))
	for _, share := range shares {
		sharedWithUser, err := s.authRepository.GetUserByID(ctx, share.SharedWithUserID)
		if err != nil {
			return nil, err
		}
		sharedByUser, err := s.authRepository.GetUserByID(ctx, share.SharedByUserID)
		if err != nil {
			return nil, err
		}

		responses = append(responses, InternalShareResponse{
			ID:                 share.ID.String(),
			DocumentID:         share.DocumentID.String(),
			SharedWithUserID:   share.SharedWithUserID.String(),
			SharedWithFullName: sharedWithUser.FullName,
			SharedWithEmail:    sharedWithUser.Email,
			SharedByUserID:     share.SharedByUserID.String(),
			SharedByFullName:   sharedByUser.FullName,
			Permission:         share.Permission,
			CreatedAt:          share.CreatedAt,
			UpdatedAt:          share.UpdatedAt,
		})
	}
	return responses, nil
}
