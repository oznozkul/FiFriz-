package integrations

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&IntegrationAccount{}, &IntegrationSyncLog{}, &ImportedExternalItem{}); err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_imported_external_items_unique
		ON imported_external_items (tenant_id, integration_account_id, external_id, external_type)
	`).Error; err != nil {
		return err
	}

	return db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_integration_sync_logs_account_created
		ON integration_sync_logs (integration_account_id, created_at DESC)
	`).Error
}
