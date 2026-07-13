package integrations

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IntegrationAccount struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID             uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID               uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Provider             string         `gorm:"type:text;not null;index" json:"provider"`
	DisplayName          string         `gorm:"type:text;not null" json:"display_name"`
	EmailAddress         *string        `gorm:"type:text" json:"email_address"`
	Status               string         `gorm:"type:text;not null;index" json:"status"`
	Settings             []byte         `gorm:"type:jsonb" json:"settings"`
	EncryptedCredentials []byte         `gorm:"type:jsonb" json:"encrypted_credentials"`
	LastSyncAt           *time.Time     `gorm:"type:timestamptz" json:"last_sync_at"`
	LastError            *string        `gorm:"type:text" json:"last_error"`
	IsActive             bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (IntegrationAccount) TableName() string {
	return "integration_accounts"
}

func (a *IntegrationAccount) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.Status == "" {
		a.Status = StatusDisconnected
	}
	return nil
}

type IntegrationSyncLog struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID             uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	IntegrationAccountID uuid.UUID  `gorm:"type:uuid;not null;index" json:"integration_account_id"`
	SyncType             string     `gorm:"type:text;not null" json:"sync_type"`
	Status               string     `gorm:"type:text;not null;index" json:"status"`
	StartedAt            *time.Time `gorm:"type:timestamptz" json:"started_at"`
	FinishedAt           *time.Time `gorm:"type:timestamptz" json:"finished_at"`
	TotalItems           int        `gorm:"not null;default:0" json:"total_items"`
	ImportedItems        int        `gorm:"not null;default:0" json:"imported_items"`
	SkippedItems         int        `gorm:"not null;default:0" json:"skipped_items"`
	FailedItems          int        `gorm:"not null;default:0" json:"failed_items"`
	ErrorMessage         *string    `gorm:"type:text" json:"error_message"`
	Metadata             []byte     `gorm:"type:jsonb" json:"metadata"`
	CreatedAt            time.Time  `json:"created_at"`
}

func (IntegrationSyncLog) TableName() string {
	return "integration_sync_logs"
}

func (l *IntegrationSyncLog) BeforeCreate(_ *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

type ImportedExternalItem struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID             uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	IntegrationAccountID uuid.UUID `gorm:"type:uuid;not null;index" json:"integration_account_id"`
	ExternalID           string    `gorm:"type:text;not null" json:"external_id"`
	ExternalType         string    `gorm:"type:text;not null" json:"external_type"`
	DocumentID           *uuid.UUID `gorm:"type:uuid;index" json:"document_id"`
	Checksum             *string   `gorm:"type:text" json:"checksum"`
	ImportedAt           time.Time `gorm:"type:timestamptz;not null" json:"imported_at"`
	Metadata             []byte    `gorm:"type:jsonb" json:"metadata"`
}

func (ImportedExternalItem) TableName() string {
	return "imported_external_items"
}

func (i *ImportedExternalItem) BeforeCreate(_ *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	if i.ImportedAt.IsZero() {
		i.ImportedAt = time.Now().UTC()
	}
	return nil
}
