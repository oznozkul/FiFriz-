package permissions

import (
	"context"

	"github.com/google/uuid"
)

type SharePermissionLookup interface {
	GetInternalSharePermission(ctx context.Context, tenantID, documentID, userID uuid.UUID) (InternalSharePermission, error)
}

type Checker struct {
	lookup SharePermissionLookup
}

func NewChecker(lookup SharePermissionLookup) *Checker {
	return &Checker{lookup: lookup}
}

func (c *Checker) BuildDocumentResource(ctx context.Context, subject Subject, tenantID, documentID, ownerUserID uuid.UUID) (DocumentResource, error) {
	doc := DocumentFromFields(tenantID, ownerUserID, "")
	if !SameTenant(subject, tenantID) {
		return doc, nil
	}
	if IsPrivilegedDocumentUser(subject, doc) {
		return doc, nil
	}
	if c == nil || c.lookup == nil {
		return doc, nil
	}
	permission, err := c.lookup.GetInternalSharePermission(ctx, tenantID, documentID, subject.UserID)
	if err != nil {
		return DocumentResource{}, err
	}
	doc.SharedPermission = permission
	return doc, nil
}

func (c *Checker) CanViewDocument(ctx context.Context, subject Subject, tenantID, documentID, ownerUserID uuid.UUID) (bool, error) {
	doc, err := c.BuildDocumentResource(ctx, subject, tenantID, documentID, ownerUserID)
	if err != nil {
		return false, err
	}
	return CanViewDocument(subject, doc), nil
}

func (c *Checker) CanDownloadDocument(ctx context.Context, subject Subject, tenantID, documentID, ownerUserID uuid.UUID) (bool, error) {
	doc, err := c.BuildDocumentResource(ctx, subject, tenantID, documentID, ownerUserID)
	if err != nil {
		return false, err
	}
	return CanDownloadDocument(subject, doc), nil
}

func (c *Checker) CanEditDocument(ctx context.Context, subject Subject, tenantID, documentID, ownerUserID uuid.UUID) (bool, error) {
	doc, err := c.BuildDocumentResource(ctx, subject, tenantID, documentID, ownerUserID)
	if err != nil {
		return false, err
	}
	return CanEditDocument(subject, doc), nil
}

func (c *Checker) CanDeleteDocument(ctx context.Context, subject Subject, tenantID, documentID, ownerUserID uuid.UUID) (bool, error) {
	doc, err := c.BuildDocumentResource(ctx, subject, tenantID, documentID, ownerUserID)
	if err != nil {
		return false, err
	}
	return CanDeleteDocument(subject, doc), nil
}

func (c *Checker) CanUploadVersion(ctx context.Context, subject Subject, tenantID, documentID, ownerUserID uuid.UUID) (bool, error) {
	doc, err := c.BuildDocumentResource(ctx, subject, tenantID, documentID, ownerUserID)
	if err != nil {
		return false, err
	}
	return CanUploadDocumentVersion(subject, doc), nil
}

func (c *Checker) CanRestoreVersion(ctx context.Context, subject Subject, tenantID, documentID, ownerUserID uuid.UUID) (bool, error) {
	doc, err := c.BuildDocumentResource(ctx, subject, tenantID, documentID, ownerUserID)
	if err != nil {
		return false, err
	}
	return CanRestoreDocumentVersion(subject, doc), nil
}

func (c *Checker) CanShareDocument(ctx context.Context, subject Subject, tenantID, documentID, ownerUserID uuid.UUID) (bool, error) {
	doc, err := c.BuildDocumentResource(ctx, subject, tenantID, documentID, ownerUserID)
	if err != nil {
		return false, err
	}
	return CanShareDocument(subject, doc), nil
}

func (c *Checker) IsSharedDocumentAccess(ctx context.Context, subject Subject, tenantID, documentID, ownerUserID uuid.UUID) (bool, error) {
	doc, err := c.BuildDocumentResource(ctx, subject, tenantID, documentID, ownerUserID)
	if err != nil {
		return false, err
	}
	return IsSharedDocumentAccess(subject, doc), nil
}
