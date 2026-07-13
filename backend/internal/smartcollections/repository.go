package smartcollections

import (
	"context"
	"time"

	"fifriz/internal/documents"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const largeFileThresholdBytes int64 = 50 * 1024 * 1024

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Count(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, accessUserID *uuid.UUID, key string, now time.Time) (int64, error) {
	if key == "large_files" {
		return r.countLargeFiles(ctx, tenantID, userID, accessUserID)
	}

	query, err := r.scopedQuery(ctx, tenantID, userID, accessUserID, key, now)
	if err != nil {
		return 0, err
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) List(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, accessUserID *uuid.UUID, key string, now time.Time) ([]documents.Document, error) {
	if key == "large_files" {
		return r.listLargeFiles(ctx, tenantID, userID, accessUserID)
	}

	query, err := r.scopedQuery(ctx, tenantID, userID, accessUserID, key, now)
	if err != nil {
		return nil, err
	}

	order := "documents.created_at DESC"
	switch key {
	case "favorites":
		order = "documents.favorited_at DESC NULLS LAST, documents.created_at DESC"
	}

	var items []documents.Document
	if err := query.Preload("Text").Order(order).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) baseQuery(ctx context.Context, tenantID uuid.UUID, accessUserID *uuid.UUID) *gorm.DB {
	query := r.db.WithContext(ctx).
		Model(&documents.Document{}).
		Select("documents.*, folders.name AS folder_name").
		Joins("LEFT JOIN folders ON folders.id = documents.folder_id AND folders.tenant_id = documents.tenant_id AND folders.deleted_at IS NULL").
		Where("documents.tenant_id = ?", tenantID)
	if accessUserID != nil {
		query = query.Where(permissions.MemberDocumentScopeSQL("documents"), *accessUserID, *accessUserID)
	}
	return query
}

func (r *Repository) scopedQuery(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, accessUserID *uuid.UUID, key string, now time.Time) (*gorm.DB, error) {
	if !IsValidKey(key) {
		return nil, ErrInvalidCollectionKey
	}

	query := r.baseQuery(ctx, tenantID, accessUserID)

	switch key {
	case "recent_7_days":
		query = query.Where("documents.created_at >= ?", now.AddDate(0, 0, -7))
	case "uploaded_this_month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)
		query = query.Where("documents.created_at >= ? AND documents.created_at < ?", start, end)
	case "favorites":
		query = query.Where("documents.is_favorite = ?", true)
	case "shared_with_me":
		query = query.Where(`
			documents.id IN (
				SELECT document_id FROM document_shares
				WHERE tenant_id = documents.tenant_id AND shared_with_user_id = ? AND deleted_at IS NULL
			) AND documents.uploaded_by_user_id != ?`, userID, userID)
	case "pending_approval":
		query = query.Where("documents.approval_status = ?", documents.DocumentApprovalStatusPending)
	case "failed_processing":
		query = query.Where("documents.status = ?", documents.DocumentStatusFailed)
	case "pdf_documents":
		query = query.Where("LOWER(documents.file_extension) = ?", "pdf")
	case "word_documents":
		query = query.Where("LOWER(documents.file_extension) IN ?", []string{"doc", "docx"})
	case "excel_documents":
		query = query.Where("LOWER(documents.file_extension) IN ?", []string{"xls", "xlsx"})
	case "email_imported":
		query = query.Where("documents.source_type IN ?", []string{"imap", "gmail", "outlook"})
	case "cloud_imported":
		query = query.Where("documents.source_type IN ?", []string{"google_drive", "onedrive", "dropbox"})
	case "invoices":
		query = query.Where("documents.ai_document_type = ?", "invoice")
	case "contracts":
		query = query.Where("documents.ai_document_type = ?", "contract")
	case "identity_documents":
		query = query.Where("documents.ai_document_type IN ?", []string{"identity_document", "passport"})
	case "insurance_policies":
		query = query.Where("documents.ai_document_type = ?", "insurance_policy")
	default:
		return nil, ErrInvalidCollectionKey
	}

	return query, nil
}

func (r *Repository) countLargeFiles(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, accessUserID *uuid.UUID) (int64, error) {
	query := r.baseQuery(ctx, tenantID, accessUserID).Where("documents.file_size >= ?", largeFileThresholdBytes)
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return count, nil
	}

	var total int64
	if err := r.baseQuery(ctx, tenantID, accessUserID).Count(&total).Error; err != nil {
		return 0, err
	}
	if total > 20 {
		return 20, nil
	}
	return total, nil
}

func (r *Repository) listLargeFiles(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, accessUserID *uuid.UUID) ([]documents.Document, error) {
	_ = userID

	query := r.baseQuery(ctx, tenantID, accessUserID).
		Where("documents.file_size >= ?", largeFileThresholdBytes).
		Order("documents.file_size DESC").
		Limit(100)

	var items []documents.Document
	if err := query.Preload("Text").Find(&items).Error; err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return items, nil
	}

	fallback := r.baseQuery(ctx, tenantID, accessUserID).
		Order("documents.file_size DESC").
		Limit(20)

	if err := fallback.Preload("Text").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
