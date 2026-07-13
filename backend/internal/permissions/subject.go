package permissions

import "github.com/google/uuid"

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type Subject struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Role     Role
}

func HasTenantWideAccess(role Role) bool {
	return role == RoleOwner || role == RoleAdmin
}

func SameTenant(subject Subject, resourceTenantID uuid.UUID) bool {
	return subject.TenantID == resourceTenantID && subject.TenantID != uuid.Nil
}
