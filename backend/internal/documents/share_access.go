package documents

import (
	"context"
	"errors"

	"fifriz/internal/auth"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShareAccessService struct {
	checker *permissions.Checker
}

func NewShareAccessService(repository *Repository) *ShareAccessService {
	return &ShareAccessService{
		checker: permissions.NewChecker(repository),
	}
}

func HasTenantWideDocumentAccess(role auth.TenantRole) bool {
	return permissions.HasTenantWideAccess(permissions.Role(role))
}

func (a *ShareAccessService) CanView(ctx context.Context, authContext auth.Context, document *Document) (bool, error) {
	return a.checker.CanViewDocument(ctx, permissions.SubjectFromAuth(authContext), document.TenantID, document.ID, document.UploadedByUserID)
}

func (a *ShareAccessService) CanDownload(ctx context.Context, authContext auth.Context, document *Document) (bool, error) {
	return a.checker.CanDownloadDocument(ctx, permissions.SubjectFromAuth(authContext), document.TenantID, document.ID, document.UploadedByUserID)
}

func (a *ShareAccessService) CanEdit(ctx context.Context, authContext auth.Context, document *Document) (bool, error) {
	return a.checker.CanEditDocument(ctx, permissions.SubjectFromAuth(authContext), document.TenantID, document.ID, document.UploadedByUserID)
}

func (a *ShareAccessService) CanManageInternalShares(ctx context.Context, authContext auth.Context, document *Document) (bool, error) {
	return a.checker.CanShareDocument(ctx, permissions.SubjectFromAuth(authContext), document.TenantID, document.ID, document.UploadedByUserID)
}

func (a *ShareAccessService) CanCreateInternalShare(ctx context.Context, authContext auth.Context, document *Document) (bool, error) {
	return a.CanManageInternalShares(ctx, authContext, document)
}

func (a *ShareAccessService) IsSharedAccess(ctx context.Context, authContext auth.Context, document *Document) (bool, error) {
	return a.checker.IsSharedDocumentAccess(ctx, permissions.SubjectFromAuth(authContext), document.TenantID, document.ID, document.UploadedByUserID)
}

func (a *ShareAccessService) CanRestoreVersion(ctx context.Context, authContext auth.Context, document *Document) (bool, error) {
	return a.checker.CanRestoreVersion(ctx, permissions.SubjectFromAuth(authContext), document.TenantID, document.ID, document.UploadedByUserID)
}

func (a *ShareAccessService) CanUploadVersion(ctx context.Context, authContext auth.Context, document *Document) (bool, error) {
	return a.checker.CanUploadVersion(ctx, permissions.SubjectFromAuth(authContext), document.TenantID, document.ID, document.UploadedByUserID)
}

func memberDocumentScope(userID uuid.UUID) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(permissions.MemberDocumentScopeSQL("documents"), userID, userID)
	}
}

func memberDocumentAccessSQL(alias string) string {
	return permissions.MemberDocumentScopeSQL(alias)
}

func (r *Repository) GetInternalSharePermission(ctx context.Context, tenantID, documentID, userID uuid.UUID) (permissions.InternalSharePermission, error) {
	var row struct {
		Permission string
	}
	err := r.db.WithContext(ctx).
		Table("document_shares").
		Select("permission").
		Where("tenant_id = ? AND document_id = ? AND shared_with_user_id = ? AND deleted_at IS NULL", tenantID, documentID, userID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return permissions.InternalSharePermission(row.Permission), nil
}
