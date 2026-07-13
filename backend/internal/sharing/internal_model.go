package sharing

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InternalSharePermission string

const (
	InternalSharePermissionView     InternalSharePermission = "view"
	InternalSharePermissionDownload InternalSharePermission = "download"
	InternalSharePermissionEdit     InternalSharePermission = "edit"
	InternalSharePermissionManage   InternalSharePermission = "manage"
)

type DocumentShare struct {
	ID               uuid.UUID               `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID         uuid.UUID               `gorm:"type:uuid;not null;index" json:"tenant_id"`
	DocumentID       uuid.UUID               `gorm:"type:uuid;not null;index" json:"document_id"`
	SharedWithUserID uuid.UUID               `gorm:"type:uuid;not null;index" json:"shared_with_user_id"`
	SharedByUserID   uuid.UUID               `gorm:"type:uuid;not null;index" json:"shared_by_user_id"`
	ShareType        string                  `gorm:"size:32;not null;default:internal_user" json:"share_type"`
	Permission       InternalSharePermission `gorm:"type:text;not null" json:"permission"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
	DeletedAt        gorm.DeletedAt          `gorm:"index" json:"-"`
}

func (s *DocumentShare) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.ShareType == "" {
		s.ShareType = ShareTypeInternalUser
	}
	return nil
}

func (s *DocumentShare) TableName() string {
	return "document_shares"
}

func isValidInternalSharePermission(value InternalSharePermission) bool {
	switch value {
	case InternalSharePermissionView,
		InternalSharePermissionDownload,
		InternalSharePermissionEdit,
		InternalSharePermissionManage:
		return true
	default:
		return false
	}
}

func internalPermissionAllowsView(permission InternalSharePermission) bool {
	return isValidInternalSharePermission(permission)
}

func internalPermissionAllowsDownload(permission InternalSharePermission) bool {
	switch permission {
	case InternalSharePermissionDownload, InternalSharePermissionEdit, InternalSharePermissionManage:
		return true
	default:
		return false
	}
}

func internalPermissionAllowsEdit(permission InternalSharePermission) bool {
	switch permission {
	case InternalSharePermissionEdit, InternalSharePermissionManage:
		return true
	default:
		return false
	}
}

func internalPermissionAllowsManageSharing(permission InternalSharePermission) bool {
	return permission == InternalSharePermissionManage
}
