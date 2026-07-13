package documents

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantDocumentPolicy stores per-tenant upload overrides.
type TenantDocumentPolicy struct {
	TenantID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"tenant_id"`
	MaxFileSizeBytes *int64    `json:"max_file_size_bytes"`
	AllowedMimeTypes string    `gorm:"type:text" json:"allowed_mime_types"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (TenantDocumentPolicy) TableName() string {
	return "tenant_document_policies"
}

func ensureEnterpriseDocumentColumns(db *gorm.DB) error {
	if !db.Migrator().HasTable(&Document{}) {
		return nil
	}

	columns := []struct {
		name       string
		definition string
	}{
		{"sha256_checksum", `varchar(64) NOT NULL DEFAULT ''`},
		{"compressed_size", `bigint`},
		{"content_encoding", `varchar(32) NOT NULL DEFAULT ''`},
		{"storage_status", `varchar(32) NOT NULL DEFAULT 'available'`},
		{"is_locked", `boolean NOT NULL DEFAULT false`},
		{"locked_at", `timestamptz`},
		{"locked_by_user_id", `uuid`},
		{"archived_at", `timestamptz`},
		{"expires_at", `timestamptz`},
		{"retention_locked_at", `timestamptz`},
	}

	for _, column := range columns {
		if !db.Migrator().HasColumn(&Document{}, column.name) {
			if err := db.Exec(`ALTER TABLE documents ADD COLUMN ` + column.name + ` ` + column.definition).Error; err != nil {
				return err
			}
		}
	}

	if err := db.Exec(`UPDATE documents SET status = 'uploading' WHERE status = 'uploaded'`).Error; err != nil {
		return err
	}
	if err := db.Exec(`UPDATE documents SET status = 'ready' WHERE status = 'processed'`).Error; err != nil {
		return err
	}
	if err := db.Exec(`ALTER TABLE documents ALTER COLUMN status SET DEFAULT 'uploading'`).Error; err != nil {
		return err
	}

	return db.AutoMigrate(&TenantDocumentPolicy{})
}
