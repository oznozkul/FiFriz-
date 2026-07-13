package folders

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Folder{}); err != nil {
		return err
	}

	if !db.Migrator().HasColumn(&Folder{}, "is_system") {
		if err := db.Exec(`ALTER TABLE folders ADD COLUMN is_system boolean NOT NULL DEFAULT false`).Error; err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(&Folder{}, "icon") {
		if err := db.Exec(`ALTER TABLE folders ADD COLUMN icon text`).Error; err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(&Folder{}, "color") {
		if err := db.Exec(`ALTER TABLE folders ADD COLUMN color text`).Error; err != nil {
			return err
		}
	}

	return db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_folders_tenant_parent_name_active
		ON folders (tenant_id, COALESCE(parent_id, '00000000-0000-0000-0000-000000000000'::uuid), lower(name))
		WHERE deleted_at IS NULL
	`).Error
}

func MigrateAll(db *gorm.DB) error {
	if err := Migrate(db); err != nil {
		return err
	}
	return MigrateEnterprise(db)
}
