package documents

import "gorm.io/gorm"

func EnsureDocumentClassificationIndexes(db *gorm.DB) error {
	return db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_documents_ai_document_type
		ON documents (tenant_id, ai_document_type)
		WHERE deleted_at IS NULL AND ai_document_type IS NOT NULL
	`).Error
}
