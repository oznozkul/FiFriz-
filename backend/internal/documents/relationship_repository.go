package documents

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RelationshipListRow struct {
	DocumentRelationship
	SourceOriginalFileName string     `gorm:"column:source_original_file_name"`
	SourceFileExtension    string     `gorm:"column:source_file_extension"`
	SourceMimeType         string     `gorm:"column:source_mime_type"`
	SourceFolderID         *uuid.UUID `gorm:"column:source_folder_id"`
	SourceFolderName       string     `gorm:"column:source_folder_name"`
	TargetOriginalFileName string     `gorm:"column:target_original_file_name"`
	TargetFileExtension    string     `gorm:"column:target_file_extension"`
	TargetMimeType         string     `gorm:"column:target_mime_type"`
	TargetFolderID         *uuid.UUID `gorm:"column:target_folder_id"`
	TargetFolderName       string     `gorm:"column:target_folder_name"`
}

func (r *Repository) ListRelationshipsByDocument(ctx context.Context, tenantID, documentID uuid.UUID) ([]RelationshipListRow, error) {
	var rows []RelationshipListRow
	err := r.db.WithContext(ctx).
		Table("document_relationships AS dr").
		Select(`
			dr.*,
			sd.original_file_name AS source_original_file_name,
			sd.file_extension AS source_file_extension,
			sd.mime_type AS source_mime_type,
			sd.folder_id AS source_folder_id,
			sf.name AS source_folder_name,
			td.original_file_name AS target_original_file_name,
			td.file_extension AS target_file_extension,
			td.mime_type AS target_mime_type,
			td.folder_id AS target_folder_id,
			tf.name AS target_folder_name
		`).
		Joins("INNER JOIN documents sd ON sd.id = dr.source_document_id AND sd.tenant_id = dr.tenant_id AND sd.deleted_at IS NULL").
		Joins("INNER JOIN documents td ON td.id = dr.target_document_id AND td.tenant_id = dr.tenant_id AND td.deleted_at IS NULL").
		Joins("LEFT JOIN folders sf ON sf.id = sd.folder_id AND sf.tenant_id = dr.tenant_id AND sf.deleted_at IS NULL").
		Joins("LEFT JOIN folders tf ON tf.id = td.folder_id AND tf.tenant_id = dr.tenant_id AND tf.deleted_at IS NULL").
		Where("dr.tenant_id = ? AND (dr.source_document_id = ? OR dr.target_document_id = ?)", tenantID, documentID, documentID).
		Order("dr.created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) GetRelationshipByID(ctx context.Context, tenantID, relationshipID uuid.UUID) (*DocumentRelationship, error) {
	var relationship DocumentRelationship
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, relationshipID).
		First(&relationship).Error
	if err != nil {
		return nil, err
	}
	return &relationship, nil
}

func (r *Repository) RelationshipExists(ctx context.Context, tenantID, sourceDocumentID, targetDocumentID uuid.UUID, relationshipType string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&DocumentRelationship{}).
		Where(
			"tenant_id = ? AND source_document_id = ? AND target_document_id = ? AND relationship_type = ?",
			tenantID, sourceDocumentID, targetDocumentID, relationshipType,
		).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) CreateRelationship(ctx context.Context, relationship *DocumentRelationship) error {
	return r.db.WithContext(ctx).Create(relationship).Error
}

func (r *Repository) DeleteRelationship(ctx context.Context, tenantID, relationshipID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, relationshipID).
		Delete(&DocumentRelationship{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func EnsureDocumentRelationshipIndexes(db *gorm.DB) error {
	return db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_document_relationships_unique_edge
		ON document_relationships (tenant_id, source_document_id, target_document_id, relationship_type)
	`).Error
}
