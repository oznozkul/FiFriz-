package sharing

import "gorm.io/gorm"

func MigrateEnterprise(db *gorm.DB) error {
	columns := []struct {
		name string
		sql  string
	}{
		{"share_type", `ALTER TABLE document_share_links ADD COLUMN share_type text NOT NULL DEFAULT 'public_link'`},
		{"status", `ALTER TABLE document_share_links ADD COLUMN status text NOT NULL DEFAULT 'active'`},
		{"expiration_preset", `ALTER TABLE document_share_links ADD COLUMN expiration_preset text`},
		{"allow_print", `ALTER TABLE document_share_links ADD COLUMN allow_print boolean NOT NULL DEFAULT false`},
		{"allow_copy_text", `ALTER TABLE document_share_links ADD COLUMN allow_copy_text boolean NOT NULL DEFAULT false`},
		{"allowed_domains", `ALTER TABLE document_share_links ADD COLUMN allowed_domains text`},
		{"watermark_enabled", `ALTER TABLE document_share_links ADD COLUMN watermark_enabled boolean NOT NULL DEFAULT false`},
		{"watermark_template", `ALTER TABLE document_share_links ADD COLUMN watermark_template text`},
		{"download_count", `ALTER TABLE document_share_links ADD COLUMN download_count integer NOT NULL DEFAULT 0`},
		{"last_opened_at", `ALTER TABLE document_share_links ADD COLUMN last_opened_at timestamptz`},
		{"last_downloaded_at", `ALTER TABLE document_share_links ADD COLUMN last_downloaded_at timestamptz`},
		{"last_ip_address", `ALTER TABLE document_share_links ADD COLUMN last_ip_address text`},
		{"last_user_agent", `ALTER TABLE document_share_links ADD COLUMN last_user_agent text`},
		{"invited_email", `ALTER TABLE document_share_links ADD COLUMN invited_email text`},
	}

	for _, column := range columns {
		if !db.Migrator().HasColumn(&DocumentShareLink{}, column.name) {
			if err := db.Exec(column.sql).Error; err != nil {
				return err
			}
		}
	}

	if !db.Migrator().HasColumn(&DocumentShare{}, "share_type") {
		if err := db.Exec(`ALTER TABLE document_shares ADD COLUMN share_type text NOT NULL DEFAULT 'internal_user'`).Error; err != nil {
			return err
		}
	}

	return db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_document_share_links_tenant_status
		ON document_share_links (tenant_id, status)
		WHERE deleted_at IS NULL
	`).Error
}
