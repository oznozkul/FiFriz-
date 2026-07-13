package folders

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Folder struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID             uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_folders_tenant_parent_name" json:"tenant_id"`
	ParentID             *uuid.UUID     `gorm:"type:uuid;index;uniqueIndex:idx_folders_tenant_parent_name" json:"parent_id"`
	Name                 string         `gorm:"size:160;not null;uniqueIndex:idx_folders_tenant_parent_name" json:"name"`
	Description          string         `gorm:"size:500" json:"description"`
	Status               string         `gorm:"size:32;not null;default:active;index" json:"status"`
	IsSystem             bool           `gorm:"not null;default:false;index" json:"is_system"`
	IsLocked             bool           `gorm:"not null;default:false;index" json:"is_locked"`
	LockedAt             *time.Time     `gorm:"type:timestamptz" json:"locked_at"`
	LockedByUserID       *uuid.UUID     `gorm:"type:uuid;index" json:"locked_by_user_id"`
	InheritPermissions   bool           `gorm:"not null;default:true" json:"inherit_permissions"`
	Icon                 *string        `gorm:"size:80" json:"icon"`
	Color                *string        `gorm:"size:40" json:"color"`
	LastActivityAt       *time.Time     `gorm:"type:timestamptz" json:"last_activity_at"`
	LastModifiedByUserID *uuid.UUID     `gorm:"type:uuid;index" json:"last_modified_by_user_id"`
	CreatedByUserID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"created_by_user_id"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

type FolderPermission struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_folder_permissions_unique" json:"tenant_id"`
	FolderID   uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_folder_permissions_unique" json:"folder_id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_folder_permissions_unique" json:"user_id"`
	Permission string    `gorm:"size:32;not null" json:"permission"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (FolderPermission) TableName() string {
	return "folder_permissions"
}

func (f *Folder) BeforeCreate(_ *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	if f.Status == "" {
		f.Status = StatusActive
	}
	return nil
}

func (p *FolderPermission) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

const (
	FolderPermissionView   = "view"
	FolderPermissionEdit   = "edit"
	FolderPermissionManage = "manage"
)

func IsValidFolderPermission(value string) bool {
	switch value {
	case FolderPermissionView, FolderPermissionEdit, FolderPermissionManage:
		return true
	default:
		return false
	}
}
