package documents

import (
	"context"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB, legacyTenantID string, legacyUserID string) error {
	if db.Migrator().HasTable(&Document{}) {
		if !db.Migrator().HasColumn(&Document{}, "tenant_id") {
			if err := db.Exec(`ALTER TABLE documents ADD COLUMN tenant_id uuid`).Error; err != nil {
				return err
			}
		}
		if !db.Migrator().HasColumn(&Document{}, "uploaded_by_user_id") {
			if err := db.Exec(`ALTER TABLE documents ADD COLUMN uploaded_by_user_id uuid`).Error; err != nil {
				return err
			}
		}
		if !db.Migrator().HasColumn(&Document{}, "folder_id") {
			if err := db.Exec(`ALTER TABLE documents ADD COLUMN folder_id uuid`).Error; err != nil {
				return err
			}
		}
		if err := db.Exec(`UPDATE documents SET tenant_id = ? WHERE tenant_id IS NULL`, legacyTenantID).Error; err != nil {
			return err
		}
		if err := db.Exec(`UPDATE documents SET uploaded_by_user_id = ? WHERE uploaded_by_user_id IS NULL`, legacyUserID).Error; err != nil {
			return err
		}
	}

	if db.Migrator().HasTable(&Document{}) {
		if !db.Migrator().HasColumn(&Document{}, "approval_status") {
			if err := db.Exec(`ALTER TABLE documents ADD COLUMN approval_status text NOT NULL DEFAULT 'none'`).Error; err != nil {
				return err
			}
		}

		sourceColumns := []struct {
			name       string
			definition string
		}{
			{"source_type", `text NOT NULL DEFAULT 'manual_upload'`},
			{"source_integration_account_id", `uuid`},
			{"source_external_id", `text`},
			{"source_metadata", `jsonb`},
		}
		for _, column := range sourceColumns {
			if !db.Migrator().HasColumn(&Document{}, column.name) {
				if err := db.Exec(`ALTER TABLE documents ADD COLUMN ` + column.name + ` ` + column.definition).Error; err != nil {
					return err
				}
			}
		}
	}

	if db.Migrator().HasTable(&Document{}) {
		classificationColumns := []struct {
			name       string
			definition string
		}{
			{"ai_document_type", `text`},
			{"ai_document_type_confidence", `double precision`},
			{"ai_classified_at", `timestamptz`},
			{"ai_classification_reason", `text`},
		}
		for _, column := range classificationColumns {
			if !db.Migrator().HasColumn(&Document{}, column.name) {
				if err := db.Exec(`ALTER TABLE documents ADD COLUMN ` + column.name + ` ` + column.definition).Error; err != nil {
					return err
				}
			}
		}
	}

	if err := EnsureDocumentClassificationIndexes(db); err != nil {
		return err
	}

	if err := db.AutoMigrate(&Document{}, &DocumentText{}, &DocumentChunk{}, &DocumentVersion{}, &DocumentComment{}, &DocumentImportantDate{}, &DocumentRelationship{}, &DocumentSummary{}, &DocumentSuggestedTag{}, &DocumentAIMetadata{}, &DocumentEntity{}, &DocumentRelatedSuggestion{}); err != nil {
		return err
	}

	if err := EnsureDocumentRelatedSuggestionIndexes(db); err != nil {
		return err
	}

	if err := EnsureDocumentEntityIndexes(db); err != nil {
		return err
	}

	if err := EnsureDocumentVersionIndexes(db); err != nil {
		return err
	}

	if err := EnsureDocumentCommentIndexes(db); err != nil {
		return err
	}

	if err := EnsureDocumentImportantDateIndexes(db); err != nil {
		return err
	}

	if db.Migrator().HasTable(&DocumentImportantDate{}) {
		if err := db.Exec(`
			UPDATE document_important_dates
			SET source = 'manual', status = 'active'
			WHERE source IS NULL OR source = '' OR status IS NULL OR status = ''
		`).Error; err != nil {
			return err
		}
	}

	if err := EnsureDocumentRelationshipIndexes(db); err != nil {
		return err
	}

	if err := documentsBackfillVersions(db); err != nil {
		return err
	}

	if db.Migrator().HasTable(&Document{}) {
		if err := db.Exec(`ALTER TABLE documents ALTER COLUMN tenant_id SET NOT NULL`).Error; err != nil {
			return err
		}
		if err := db.Exec(`ALTER TABLE documents ALTER COLUMN uploaded_by_user_id SET NOT NULL`).Error; err != nil {
			return err
		}
	}

	if db.Migrator().HasTable(&Document{}) {
		if !db.Migrator().HasColumn(&Document{}, "is_favorite") {
			if err := db.Exec(`ALTER TABLE documents ADD COLUMN is_favorite boolean NOT NULL DEFAULT false`).Error; err != nil {
				return err
			}
		}
		if !db.Migrator().HasColumn(&Document{}, "favorited_at") {
			if err := db.Exec(`ALTER TABLE documents ADD COLUMN favorited_at timestamptz`).Error; err != nil {
				return err
			}
		}
	}

	if db.Migrator().HasTable(&Document{}) {
		storageColumns := []struct {
			name       string
			definition string
		}{
			{"storage_provider", `text NOT NULL DEFAULT 'local'`},
			{"storage_key", `text`},
			{"bucket_name", `text`},
		}
		for _, column := range storageColumns {
			if !db.Migrator().HasColumn(&Document{}, column.name) {
				if err := db.Exec(`ALTER TABLE documents ADD COLUMN ` + column.name + ` ` + column.definition).Error; err != nil {
					return err
				}
			}
		}
	}

	if db.Migrator().HasTable(&DocumentVersion{}) {
		versionStorageColumns := []struct {
			name       string
			definition string
		}{
			{"storage_provider", `text NOT NULL DEFAULT 'local'`},
			{"storage_key", `text`},
			{"bucket_name", `text`},
		}
		for _, column := range versionStorageColumns {
			if !db.Migrator().HasColumn(&DocumentVersion{}, column.name) {
				if err := db.Exec(`ALTER TABLE document_versions ADD COLUMN ` + column.name + ` ` + column.definition).Error; err != nil {
					return err
				}
			}
		}
	}

	if db.Migrator().HasTable(&Document{}) {
		if err := db.Exec(`
		UPDATE documents
		SET storage_provider = 'local'
		WHERE storage_provider IS NULL OR storage_provider = ''
	`).Error; err != nil {
			return err
		}
		if err := db.Exec(`
		UPDATE documents
		SET storage_key = stored_file_name
		WHERE (storage_key IS NULL OR storage_key = '') AND stored_file_name <> ''
	`).Error; err != nil {
			return err
		}
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_document_texts_full_text_fts
		ON document_texts
		USING GIN (to_tsvector('simple', coalesce(full_text, '')))
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_document_texts_full_text_normalized_fts
		ON document_texts
		USING GIN (
			to_tsvector(
				'simple',
				translate(lower(coalesce(full_text, '')), 'çğıöşüâîû', 'cgiosuaiu')
			)
		)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_document_chunks_document_id_chunk_index
		ON document_chunks (document_id, chunk_index)
	`).Error; err != nil {
		return err
	}

	if !db.Migrator().HasColumn(&DocumentChunk{}, "embedding") {
		if err := db.Exec(`ALTER TABLE document_chunks ADD COLUMN embedding vector(1024)`).Error; err != nil {
			return err
		}
	} else if err := db.Exec(`ALTER TABLE document_chunks ALTER COLUMN embedding TYPE vector(1024)`).Error; err != nil {
		return err
	}

	legacyColumns := []string{
		"name",
		"category",
		"folder",
		"size",
		"file_type",
		"original_name",
	}

	for _, column := range legacyColumns {
		if db.Migrator().HasColumn(&Document{}, column) {
			if err := db.Migrator().DropColumn(&Document{}, column); err != nil {
				return err
			}
		}
	}

	if err := ensureEnterpriseDocumentColumns(db); err != nil {
		return err
	}

	return nil
}

func documentsBackfillVersions(db *gorm.DB) error {
	repository := NewRepository(db)
	return repository.BackfillMissingVersions(context.Background())
}
