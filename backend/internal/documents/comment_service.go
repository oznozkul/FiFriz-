package documents

import (
	"context"
	"errors"
	"strings"

	"fifriz/internal/applog"
	"fifriz/internal/auth"
	"fifriz/internal/permissions"
	domain "fifriz/internal/events/events"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrDocumentCommentNotFound = errors.New("document comment not found")
	ErrInvalidCommentID        = errors.New("invalid comment id")
	ErrEmptyCommentContent     = errors.New("empty comment content")
)

func (s *Service) ListComments(ctx context.Context, documentIDRaw string) ([]DocumentCommentResponse, error) {
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

	rows, err := s.repository.ListCommentsByDocument(ctx, authContext.TenantID, documentID)
	if err != nil {
		return nil, err
	}

	return buildCommentThread(rows), nil
}

func (s *Service) CreateComment(ctx context.Context, documentIDRaw string, input CreateDocumentCommentRequest) (*DocumentCommentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrEmptyCommentContent
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
	if err := validateParentComment(ctx, s.repository, authContext.TenantID, documentID, input.ParentID); err != nil {
		return nil, err
	}

	var parentComment *DocumentComment
	if input.ParentID != nil {
		parentComment, err = s.repository.GetCommentByID(ctx, authContext.TenantID, documentID, *input.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrInvalidParentComment
			}
			return nil, err
		}
	}

	comment := &DocumentComment{
		TenantID:   authContext.TenantID,
		DocumentID: documentID,
		UserID:     authContext.UserID,
		ParentID:   input.ParentID,
		Content:    content,
	}
	if err := s.repository.CreateComment(ctx, comment); err != nil {
		return nil, err
	}

	var parentAuthorID *uuid.UUID
	if parentComment != nil {
		parentAuthorID = &parentComment.UserID
	}
	s.publisher.PublishAfterCommit(ctx, domain.NewCommentCreated(
		authContext.TenantID,
		authContext.UserID,
		documentID,
		comment.ID,
		document.UploadedByUserID,
		input.ParentID,
		parentAuthorID,
		document.OriginalFileName,
	))

	row, err := s.repository.GetCommentWithAuthor(ctx, authContext.TenantID, documentID, comment.ID)
	if err != nil {
		response := toDocumentCommentResponse(CommentListRow{DocumentComment: *comment})
		return &response, nil
	}

	response := toDocumentCommentResponse(*row)
	return &response, nil
}

func (s *Service) UpdateComment(ctx context.Context, documentIDRaw, commentIDRaw string, input UpdateDocumentCommentRequest) (*DocumentCommentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrEmptyCommentContent
	}

	documentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	commentID, err := uuid.Parse(commentIDRaw)
	if err != nil {
		return nil, ErrInvalidCommentID
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

	comment, err := s.repository.GetCommentByID(ctx, authContext.TenantID, documentID, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentCommentNotFound
		}
		return nil, err
	}
	if err := s.requireManageCommentAccess(authContext, comment); err != nil {
		return nil, err
	}

	comment.Content = content
	comment.IsEdited = true
	if err := s.repository.UpdateComment(ctx, comment); err != nil {
		return nil, err
	}

	row, err := s.repository.GetCommentWithAuthor(ctx, authContext.TenantID, documentID, comment.ID)
	if err != nil {
		response := toDocumentCommentResponse(CommentListRow{DocumentComment: *comment})
		return &response, nil
	}

	response := toDocumentCommentResponse(*row)
	return &response, nil
}

func (s *Service) DeleteComment(ctx context.Context, documentIDRaw, commentIDRaw string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	documentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return ErrInvalidDocumentID
	}

	commentID, err := uuid.Parse(commentIDRaw)
	if err != nil {
		return ErrInvalidCommentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDocumentNotFound
		}
		return err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return err
	}

	comment, err := s.repository.GetCommentByID(ctx, authContext.TenantID, documentID, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDocumentCommentNotFound
		}
		return err
	}
	if err := s.requireManageCommentAccess(authContext, comment); err != nil {
		return err
	}

	if err := s.repository.SoftDeleteComment(ctx, authContext.TenantID, documentID, commentID); err != nil {
		return err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewCommentDeleted(authContext.TenantID, authContext.UserID, documentID, commentID))
	return nil
}

func (s *Service) requireManageCommentAccess(authContext auth.Context, comment *DocumentComment) error {
	subject := permissions.SubjectFromAuth(authContext)
	if !permissions.CanManageDocumentComment(subject, comment.UserID) {
		applog.PermissionDenied("manage_document_comment", comment.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
		return ErrForbidden
	}
	return nil
}
