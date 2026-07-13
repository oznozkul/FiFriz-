package permissions

import (
	"github.com/google/uuid"
)

type InternalSharePermission string

const (
	InternalShareView     InternalSharePermission = "view"
	InternalShareDownload InternalSharePermission = "download"
	InternalShareEdit     InternalSharePermission = "edit"
	InternalShareManage   InternalSharePermission = "manage"
)

type DocumentResource struct {
	TenantID         uuid.UUID
	OwnerUserID      uuid.UUID
	SharedPermission InternalSharePermission
}

func DocumentFromFields(tenantID, ownerUserID uuid.UUID, sharedPermission InternalSharePermission) DocumentResource {
	return DocumentResource{
		TenantID:         tenantID,
		OwnerUserID:      ownerUserID,
		SharedPermission: sharedPermission,
	}
}

func IsDocumentOwner(subject Subject, doc DocumentResource) bool {
	return subject.UserID == doc.OwnerUserID
}

func IsPrivilegedDocumentUser(subject Subject, doc DocumentResource) bool {
	return SameTenant(subject, doc.TenantID) &&
		(HasTenantWideAccess(subject.Role) || IsDocumentOwner(subject, doc))
}

func shareAllowsView(permission InternalSharePermission) bool {
	switch permission {
	case InternalShareView, InternalShareDownload, InternalShareEdit, InternalShareManage:
		return true
	default:
		return false
	}
}

func shareAllowsDownload(permission InternalSharePermission) bool {
	switch permission {
	case InternalShareDownload, InternalShareEdit, InternalShareManage:
		return true
	default:
		return false
	}
}

func shareAllowsEdit(permission InternalSharePermission) bool {
	switch permission {
	case InternalShareEdit, InternalShareManage:
		return true
	default:
		return false
	}
}

func shareAllowsManageSharing(permission InternalSharePermission) bool {
	return permission == InternalShareManage
}

func CanViewDocument(subject Subject, doc DocumentResource) bool {
	if !SameTenant(subject, doc.TenantID) {
		return false
	}
	if IsPrivilegedDocumentUser(subject, doc) {
		return true
	}
	return shareAllowsView(doc.SharedPermission)
}

func CanDownloadDocument(subject Subject, doc DocumentResource) bool {
	if !SameTenant(subject, doc.TenantID) {
		return false
	}
	if IsPrivilegedDocumentUser(subject, doc) {
		return true
	}
	return shareAllowsDownload(doc.SharedPermission)
}

func CanEditDocument(subject Subject, doc DocumentResource) bool {
	if !SameTenant(subject, doc.TenantID) {
		return false
	}
	if IsPrivilegedDocumentUser(subject, doc) {
		return true
	}
	return shareAllowsEdit(doc.SharedPermission)
}

func CanDeleteDocument(subject Subject, doc DocumentResource) bool {
	if !SameTenant(subject, doc.TenantID) {
		return false
	}
	return HasTenantWideAccess(subject.Role) || IsDocumentOwner(subject, doc)
}

func CanShareDocument(subject Subject, doc DocumentResource) bool {
	if !SameTenant(subject, doc.TenantID) {
		return false
	}
	if IsPrivilegedDocumentUser(subject, doc) {
		return true
	}
	return shareAllowsManageSharing(doc.SharedPermission)
}

func CanUploadDocumentVersion(subject Subject, doc DocumentResource) bool {
	return CanEditDocument(subject, doc)
}

func CanRestoreDocumentVersion(subject Subject, doc DocumentResource) bool {
	return CanShareDocument(subject, doc)
}

func CanManageDocumentComment(subject Subject, commentAuthorID uuid.UUID) bool {
	return subject.UserID == commentAuthorID || HasTenantWideAccess(subject.Role)
}

func CanDeleteDocumentImportantDate(subject Subject, createdByUserID uuid.UUID) bool {
	return subject.UserID == createdByUserID || HasTenantWideAccess(subject.Role)
}

func CanCompleteDocumentImportantDate(subject Subject, createdByUserID uuid.UUID, assignedToUserID *uuid.UUID) bool {
	if HasTenantWideAccess(subject.Role) {
		return true
	}
	if subject.UserID == createdByUserID {
		return true
	}
	if assignedToUserID != nil && subject.UserID == *assignedToUserID {
		return true
	}
	return false
}

func CanUpdateDocumentImportantDate(subject Subject, createdByUserID uuid.UUID, canEditDocument bool) bool {
	if canEditDocument {
		return true
	}
	return subject.UserID == createdByUserID || HasTenantWideAccess(subject.Role)
}

func IsSharedDocumentAccess(subject Subject, doc DocumentResource) bool {
	if IsPrivilegedDocumentUser(subject, doc) {
		return false
	}
	return doc.SharedPermission != ""
}

func MemberDocumentScopeSQL(alias string) string {
	return "(" + alias + ".uploaded_by_user_id = ? OR " + alias + ".id IN (SELECT document_id FROM document_shares WHERE tenant_id = " + alias + ".tenant_id AND shared_with_user_id = ? AND deleted_at IS NULL))"
}
