package database

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"fifriz/internal/ai"
	"fifriz/internal/approvals"
	"fifriz/internal/audit"
	"fifriz/internal/auth"
	"fifriz/internal/config"
	"fifriz/internal/documents"
	"fifriz/internal/folders"
	"fifriz/internal/integrations"
	"fifriz/internal/navigation"
	"fifriz/internal/notifications"
	"fifriz/internal/sharing"
	"fifriz/internal/tags"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var validDatabaseName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil && isMissingDatabaseError(err) {
		if ensureErr := ensureDatabaseExists(cfg.DSN); ensureErr != nil {
			return nil, fmt.Errorf("ensure database: %w (original: %v)", ensureErr, err)
		}
		db, err = gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	}
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

func ConnectWithRetry(cfg config.DatabaseConfig, attempts int, delay time.Duration) (*gorm.DB, error) {
	if attempts < 1 {
		attempts = 1
	}
	if delay <= 0 {
		delay = time.Second
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		db, err := Connect(cfg)
		if err == nil {
			return db, nil
		}
		lastErr = err
		if attempt < attempts {
			log.Printf("database connect attempt %d/%d failed: %v", attempt, attempts, err)
			time.Sleep(delay)
		}
	}
	return nil, fmt.Errorf("database connect failed after %d attempts: %w", attempts, lastErr)
}

func isMissingDatabaseError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "does not exist") && strings.Contains(message, "database")
}

func ensureDatabaseExists(dsn string) error {
	dbName := databaseNameFromDSN(dsn)
	if dbName == "" || dbName == "postgres" {
		return fmt.Errorf("could not determine database name from DSN")
	}
	if !validDatabaseName.MatchString(dbName) {
		return fmt.Errorf("invalid database name %q", dbName)
	}

	adminDSN := replaceDatabaseInDSN(dsn, "postgres")
	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := adminDB.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	var exists bool
	if err := adminDB.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)", dbName).Scan(&exists).Error; err != nil {
		return err
	}
	if exists {
		return nil
	}

	return adminDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, dbName)).Error
}

func databaseNameFromDSN(dsn string) string {
	for _, part := range strings.Fields(strings.ReplaceAll(dsn, "\n", " ")) {
		if strings.HasPrefix(part, "dbname=") {
			return strings.TrimPrefix(part, "dbname=")
		}
	}
	return ""
}

func replaceDatabaseInDSN(dsn, dbName string) string {
	parts := strings.Fields(strings.ReplaceAll(dsn, "\n", " "))
	for index, part := range parts {
		if strings.HasPrefix(part, "dbname=") {
			parts[index] = "dbname=" + dbName
			return strings.Join(parts, " ")
		}
	}
	return dsn + " dbname=" + dbName
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return err
	}

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS vector`).Error; err != nil {
		return err
	}

	legacyTenantID, legacyUserID, err := auth.Migrate(db)
	if err != nil {
		return err
	}

	if err := folders.MigrateAll(db); err != nil {
		return err
	}

	if err := audit.Migrate(db); err != nil {
		return err
	}

	if err := tags.MigrateAll(db); err != nil {
		return err
	}

	if err := sharing.Migrate(db); err != nil {
		return err
	}

	if err := notifications.Migrate(db); err != nil {
		return err
	}

	if err := approvals.Migrate(db); err != nil {
		return err
	}

	if err := integrations.Migrate(db); err != nil {
		return err
	}

	if err := navigation.Migrate(db); err != nil {
		return err
	}

	if err := ai.Migrate(db); err != nil {
		return err
	}

	return documents.Migrate(db, legacyTenantID, legacyUserID)
}
