package documents

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommentListRow struct {
	DocumentComment
	UserFullName string `gorm:"column:user_full_name"`
}

func (r *Repository) ListCommentsByDocument(ctx context.Context, tenantID, documentID uuid.UUID) ([]CommentListRow, error) {
	var rows []CommentListRow
	err := r.db.WithContext(ctx).
		Table("document_comments AS dc").
		Select("dc.*, u.full_name AS user_full_name").
		Joins("LEFT JOIN users u ON u.id = dc.user_id").
		Where("dc.tenant_id = ? AND dc.document_id = ? AND dc.deleted_at IS NULL", tenantID, documentID).
		Order("dc.created_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) GetCommentByID(ctx context.Context, tenantID, documentID, commentID uuid.UUID) (*DocumentComment, error) {
	var comment DocumentComment
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ? AND id = ?", tenantID, documentID, commentID).
		First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *Repository) CreateComment(ctx context.Context, comment *DocumentComment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *Repository) UpdateComment(ctx context.Context, comment *DocumentComment) error {
	return r.db.WithContext(ctx).Save(comment).Error
}

func (r *Repository) SoftDeleteComment(ctx context.Context, tenantID, documentID, commentID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ? AND id = ?", tenantID, documentID, commentID).
		Delete(&DocumentComment{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) GetCommentWithAuthor(ctx context.Context, tenantID, documentID, commentID uuid.UUID) (*CommentListRow, error) {
	var row CommentListRow
	err := r.db.WithContext(ctx).
		Table("document_comments AS dc").
		Select("dc.*, u.full_name AS user_full_name").
		Joins("LEFT JOIN users u ON u.id = dc.user_id").
		Where("dc.tenant_id = ? AND dc.document_id = ? AND dc.id = ? AND dc.deleted_at IS NULL", tenantID, documentID, commentID).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repository) ParentCommentExists(ctx context.Context, tenantID, documentID, parentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&DocumentComment{}).
		Where("tenant_id = ? AND document_id = ? AND id = ? AND deleted_at IS NULL", tenantID, documentID, parentID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func EnsureDocumentCommentIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_document_comments_document_created
		ON document_comments (document_id, created_at)
		WHERE deleted_at IS NULL
	`).Error; err != nil {
		return err
	}
	return nil
}

func validateParentComment(ctx context.Context, r *Repository, tenantID, documentID uuid.UUID, parentID *uuid.UUID) error {
	if parentID == nil {
		return nil
	}
	if *parentID == uuid.Nil {
		return ErrInvalidParentComment
	}
	exists, err := r.ParentCommentExists(ctx, tenantID, documentID, *parentID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrInvalidParentComment
	}
	return nil
}

var ErrInvalidParentComment = errors.New("invalid parent comment")
