package navigation

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&PinnedItem{}); err != nil {
		return err
	}

	return db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_pins_tenant_user_entity
		ON pinned_items (tenant_id, user_id, entity_type, entity_id)
	`).Error
}
