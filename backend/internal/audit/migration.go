package audit

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&AuditLog{}); err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_created_at
		ON audit_logs (tenant_id, created_at DESC)
	`).Error; err != nil {
		return err
	}

	return nil
}
