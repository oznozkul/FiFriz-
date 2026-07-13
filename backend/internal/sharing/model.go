package sharing

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SharePermission string

const (
	SharePermissionView     SharePermission = "view"
	SharePermissionDownload SharePermission = "download"
)

type DocumentShareLink struct {
	ID                 uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID           uuid.UUID       `gorm:"type:uuid;not null;index" json:"tenant_id"`
	DocumentID         uuid.UUID       `gorm:"type:uuid;not null;index" json:"document_id"`
	CreatedByUserID    uuid.UUID       `gorm:"type:uuid;not null;index" json:"created_by_user_id"`
	ShareType          string          `gorm:"size:32;not null;default:public_link;index" json:"share_type"`
	Status             string          `gorm:"size:32;not null;default:active;index" json:"status"`
	Token              string          `gorm:"type:text;not null;uniqueIndex" json:"token"`
	Permission         SharePermission `gorm:"type:text;not null" json:"permission"`
	PasswordHash       *string         `gorm:"type:text" json:"-"`
	ExpiresAt          *time.Time      `gorm:"type:timestamptz" json:"expires_at"`
	ExpirationPreset   string          `gorm:"size:32" json:"expiration_preset"`
	MaxViewCount       *int            `json:"max_view_count"`
	ViewCount          int             `gorm:"not null;default:0" json:"view_count"`
	DownloadCount      int             `gorm:"not null;default:0" json:"download_count"`
	AllowDownload      bool            `gorm:"not null;default:true" json:"allow_download"`
	AllowPrint         bool            `gorm:"not null;default:false" json:"allow_print"`
	AllowCopyText      bool            `gorm:"not null;default:false" json:"allow_copy_text"`
	AllowedDomains     string          `gorm:"type:text" json:"allowed_domains"`
	WatermarkEnabled   bool            `gorm:"not null;default:false" json:"watermark_enabled"`
	WatermarkTemplate  *string         `gorm:"type:text" json:"watermark_template"`
	InvitedEmail       *string         `gorm:"size:320" json:"invited_email"`
	LastOpenedAt       *time.Time      `gorm:"type:timestamptz" json:"last_opened_at"`
	LastDownloadedAt   *time.Time      `gorm:"type:timestamptz" json:"last_downloaded_at"`
	LastIPAddress      *string         `gorm:"type:text" json:"last_ip_address"`
	LastUserAgent      *string         `gorm:"type:text" json:"last_user_agent"`
	IsActive           bool            `gorm:"not null;default:true;index" json:"is_active"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	DeletedAt          gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (l *DocumentShareLink) BeforeCreate(_ *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	if l.ShareType == "" {
		l.ShareType = ShareTypePublicLink
	}
	if l.Status == "" {
		l.Status = ShareStatusActive
	}
	return nil
}

func (l *DocumentShareLink) TableName() string {
	return "document_share_links"
}
