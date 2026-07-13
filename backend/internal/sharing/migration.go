package sharing

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&DocumentShareLink{}, &DocumentShare{}); err != nil {
		return err
	}
	if err := EnsureInternalShareUniqueIndex(db); err != nil {
		return err
	}
	return MigrateEnterprise(db)
}
