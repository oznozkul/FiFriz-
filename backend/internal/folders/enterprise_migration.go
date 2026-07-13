package folders

import "gorm.io/gorm"

func MigrateEnterprise(db *gorm.DB) error {
	if err := db.AutoMigrate(&FolderPermission{}); err != nil {
		return err
	}

	columns := []struct {
		name string
		sql  string
	}{
		{"status", `ALTER TABLE folders ADD COLUMN status text NOT NULL DEFAULT 'active'`},
		{"is_locked", `ALTER TABLE folders ADD COLUMN is_locked boolean NOT NULL DEFAULT false`},
		{"locked_at", `ALTER TABLE folders ADD COLUMN locked_at timestamptz`},
		{"locked_by_user_id", `ALTER TABLE folders ADD COLUMN locked_by_user_id uuid`},
		{"inherit_permissions", `ALTER TABLE folders ADD COLUMN inherit_permissions boolean NOT NULL DEFAULT true`},
		{"last_activity_at", `ALTER TABLE folders ADD COLUMN last_activity_at timestamptz`},
		{"last_modified_by_user_id", `ALTER TABLE folders ADD COLUMN last_modified_by_user_id uuid`},
	}

	for _, column := range columns {
		if !db.Migrator().HasColumn(&Folder{}, column.name) {
			if err := db.Exec(column.sql).Error; err != nil {
				return err
			}
		}
	}

	return db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_folders_tenant_status_active
		ON folders (tenant_id, status)
		WHERE deleted_at IS NULL
	`).Error
}
