package notifications

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Notification{}); err != nil {
		return err
	}

	return db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
		ON notifications (tenant_id, user_id, created_at DESC)
		WHERE is_read = false
	`).Error
}
