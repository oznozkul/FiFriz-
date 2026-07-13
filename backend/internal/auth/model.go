package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantType string

const (
	TenantTypePersonal     TenantType = "personal"
	TenantTypeOrganization TenantType = "organization"
)

type TenantRole string

const (
	TenantRoleOwner  TenantRole = "owner"
	TenantRoleAdmin  TenantRole = "admin"
	TenantRoleMember TenantRole = "member"
)

type Tenant struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string     `gorm:"size:255;not null" json:"name"`
	Type      TenantType `gorm:"size:40;not null;index" json:"type"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (t *Tenant) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	FullName     string    `gorm:"size:255;not null" json:"full_name"`
	Email        string    `gorm:"size:255;not null;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type TenantMember struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID  uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_tenant_members_tenant_user;index" json:"tenant_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_tenant_members_tenant_user;index" json:"user_id"`
	Role      TenantRole `gorm:"size:40;not null" json:"role"`
	Tenant    Tenant     `gorm:"constraint:OnDelete:CASCADE" json:"tenant"`
	User      User       `gorm:"constraint:OnDelete:CASCADE" json:"user"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (m *TenantMember) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
