package permissions

import (
	"testing"

	"github.com/google/uuid"
)

func TestCanViewDocument(t *testing.T) {
	tenantID := uuid.New()
	ownerID := uuid.New()
	memberID := uuid.New()

	subject := Subject{TenantID: tenantID, UserID: memberID, Role: RoleMember}
	doc := DocumentFromFields(tenantID, ownerID, InternalShareView)

	if !CanViewDocument(subject, doc) {
		t.Fatal("shared view permission should allow view")
	}

	doc.SharedPermission = ""
	if CanViewDocument(subject, doc) {
		t.Fatal("missing share should deny view for member")
	}

	admin := Subject{TenantID: tenantID, UserID: uuid.New(), Role: RoleAdmin}
	if !CanViewDocument(admin, doc) {
		t.Fatal("admin should view all tenant documents")
	}
}

func TestCanDeleteDocument(t *testing.T) {
	tenantID := uuid.New()
	ownerID := uuid.New()
	otherTenant := uuid.New()

	uploader := Subject{TenantID: tenantID, UserID: ownerID, Role: RoleMember}
	doc := DocumentFromFields(tenantID, ownerID, "")

	if !CanDeleteDocument(uploader, doc) {
		t.Fatal("uploader should delete own document")
	}

	sharedMember := Subject{TenantID: tenantID, UserID: uuid.New(), Role: RoleMember}
	sharedDoc := DocumentFromFields(tenantID, ownerID, InternalShareEdit)
	if CanDeleteDocument(sharedMember, sharedDoc) {
		t.Fatal("shared edit permission must not allow delete")
	}

	if CanDeleteDocument(uploader, DocumentFromFields(otherTenant, ownerID, "")) {
		t.Fatal("cross-tenant delete must be denied")
	}
}

func TestCanShareDocument(t *testing.T) {
	tenantID := uuid.New()
	ownerID := uuid.New()

	manager := Subject{TenantID: tenantID, UserID: uuid.New(), Role: RoleMember}
	doc := DocumentFromFields(tenantID, ownerID, InternalShareManage)

	if !CanShareDocument(manager, doc) {
		t.Fatal("manage share permission should allow sharing")
	}

	doc.SharedPermission = InternalShareEdit
	if CanShareDocument(manager, doc) {
		t.Fatal("edit permission must not allow sharing management")
	}
}
