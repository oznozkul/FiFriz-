package documents

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VersionListRow struct {
	DocumentVersion
	UploadedByFullName string `gorm:"column:uploaded_by_full_name"`
}

func (r *Repository) CreateWithInitialVersion(ctx context.Context, document *Document) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(document).Error; err != nil {
			return err
		}
		version := versionFromDocument(document, 1, true, nil, document.UploadedByUserID)
		return tx.Create(version).Error
	})
}

func (r *Repository) ListVersionsByDocument(ctx context.Context, tenantID, documentID uuid.UUID) ([]VersionListRow, error) {
	var rows []VersionListRow
	err := r.db.WithContext(ctx).
		Table("document_versions dv").
		Select("dv.*, u.full_name AS uploaded_by_full_name").
		Joins("LEFT JOIN users u ON u.id = dv.uploaded_by_user_id").
		Where("dv.tenant_id = ? AND dv.document_id = ?", tenantID, documentID).
		Order("dv.version_number DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) GetVersionByID(ctx context.Context, tenantID, documentID, versionID uuid.UUID) (*DocumentVersion, error) {
	var version DocumentVersion
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ? AND id = ?", tenantID, documentID, versionID).
		First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *Repository) GetMaxVersionNumber(ctx context.Context, tenantID, documentID uuid.UUID) (int, error) {
	var maxNumber *int
	err := r.db.WithContext(ctx).
		Model(&DocumentVersion{}).
		Select("MAX(version_number)").
		Where("tenant_id = ? AND document_id = ?", tenantID, documentID).
		Scan(&maxNumber).Error
	if err != nil {
		return 0, err
	}
	if maxNumber == nil {
		return 0, nil
	}
	return *maxNumber, nil
}

func (r *Repository) CreateVersionAndUpdateDocument(ctx context.Context, document *Document, version *DocumentVersion) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&DocumentVersion{}).
			Where("tenant_id = ? AND document_id = ? AND is_current = ?", document.TenantID, document.ID, true).
			Update("is_current", false).Error; err != nil {
			return err
		}
		if err := tx.Create(version).Error; err != nil {
			return err
		}
		return tx.Save(document).Error
	})
}

func (r *Repository) RestoreVersionAndUpdateDocument(ctx context.Context, document *Document, version *DocumentVersion) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&DocumentVersion{}).
			Where("tenant_id = ? AND document_id = ? AND is_current = ?", document.TenantID, document.ID, true).
			Update("is_current", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&DocumentVersion{}).
			Where("id = ? AND tenant_id = ? AND document_id = ?", version.ID, document.TenantID, document.ID).
			Update("is_current", true).Error; err != nil {
			return err
		}
		return tx.Save(document).Error
	})
}

func (r *Repository) BackfillMissingVersions(ctx context.Context) error {
	type documentRow struct {
		ID               uuid.UUID
		TenantID         uuid.UUID
		UploadedByUserID uuid.UUID
		OriginalFileName string
		StoredFileName   string
		FilePath         string
		FileExtension    string
		MimeType         string
		FileSize         int64
		CreatedAt        time.Time
	}

	var documents []documentRow
	if err := r.db.WithContext(ctx).
		Table("documents").
		Select("id, tenant_id, uploaded_by_user_id, original_file_name, stored_file_name, file_path, file_extension, mime_type, file_size, created_at").
		Where("deleted_at IS NULL").
		Where("NOT EXISTS (SELECT 1 FROM document_versions dv WHERE dv.document_id = documents.id)").
		Find(&documents).Error; err != nil {
		return err
	}

	for _, row := range documents {
		version := &DocumentVersion{
			TenantID:         row.TenantID,
			DocumentID:       row.ID,
			VersionNumber:    1,
			OriginalFileName: row.OriginalFileName,
			StoredFileName:   row.StoredFileName,
			FilePath:         row.FilePath,
			FileExtension:    row.FileExtension,
			MimeType:         row.MimeType,
			FileSize:         row.FileSize,
			UploadedByUserID: row.UploadedByUserID,
			IsCurrent:        true,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.CreatedAt,
		}
		if err := r.db.WithContext(ctx).Create(version).Error; err != nil {
			return err
		}
	}
	return nil
}

func EnsureDocumentVersionIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_document_versions_document_version_number
		ON document_versions (document_id, version_number)
	`).Error; err != nil {
		return err
	}
	return db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_document_versions_one_current
		ON document_versions (document_id)
		WHERE is_current = true
	`).Error
}
