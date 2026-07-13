package tags

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Tag{}, &DocumentTag{}); err != nil {
		return err
	}

	return db.Exec(`CREATE INDEX IF NOT EXISTS idx_document_tags_tenant_tag ON document_tags (tenant_id, tag_id)`).Error
}

func MigrateAll(db *gorm.DB) error {
	if err := Migrate(db); err != nil {
		return err
	}
	return MigrateEnterprise(db)
}
