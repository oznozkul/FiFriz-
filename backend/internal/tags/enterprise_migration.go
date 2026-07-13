package tags

import "gorm.io/gorm"

func MigrateEnterprise(db *gorm.DB) error {
	if err := db.AutoMigrate(&TagAlias{}); err != nil {
		return err
	}

	columns := []struct {
		name string
		sql  string
	}{
		{"status", `ALTER TABLE tags ADD COLUMN status text NOT NULL DEFAULT 'active'`},
		{"is_system", `ALTER TABLE tags ADD COLUMN is_system boolean NOT NULL DEFAULT false`},
		{"icon", `ALTER TABLE tags ADD COLUMN icon text`},
	}

	for _, column := range columns {
		if !db.Migrator().HasColumn(&Tag{}, column.name) {
			if err := db.Exec(column.sql).Error; err != nil {
				return err
			}
		}
	}

	return db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_tenant_name_active
		ON tags (tenant_id, lower(name))
		WHERE deleted_at IS NULL
	`).Error
}
