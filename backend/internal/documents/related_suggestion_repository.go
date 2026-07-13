package documents

import (
	"context"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type RelatedSuggestionListRow struct {
	DocumentRelatedSuggestion
	TargetOriginalFileName string     `gorm:"column:target_original_file_name"`
	TargetFileExtension    string     `gorm:"column:target_file_extension"`
	TargetMimeType         string     `gorm:"column:target_mime_type"`
	TargetFolderID         *uuid.UUID `gorm:"column:target_folder_id"`
	TargetFolderName       string     `gorm:"column:target_folder_name"`
}

type DocumentSimilarityScore struct {
	DocumentID uuid.UUID `gorm:"column:document_id"`
	Similarity float64   `gorm:"column:similarity"`
}

func (r *Repository) ListSuggestedRelatedDocuments(ctx context.Context, tenantID, sourceDocumentID uuid.UUID) ([]RelatedSuggestionListRow, error) {
	var rows []RelatedSuggestionListRow
	err := r.db.WithContext(ctx).
		Table("document_related_suggestions AS drs").
		Select(`
			drs.*,
			td.original_file_name AS target_original_file_name,
			td.file_extension AS target_file_extension,
			td.mime_type AS target_mime_type,
			td.folder_id AS target_folder_id,
			tf.name AS target_folder_name
		`).
		Joins("INNER JOIN documents td ON td.id = drs.target_document_id AND td.tenant_id = drs.tenant_id AND td.deleted_at IS NULL").
		Joins("LEFT JOIN folders tf ON tf.id = td.folder_id AND tf.tenant_id = drs.tenant_id AND tf.deleted_at IS NULL").
		Where("drs.tenant_id = ? AND drs.source_document_id = ? AND drs.status = ?", tenantID, sourceDocumentID, RelatedSuggestionStatusSuggested).
		Order("drs.similarity DESC NULLS LAST, drs.created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) GetRelatedSuggestionByID(ctx context.Context, tenantID, sourceDocumentID, suggestionID uuid.UUID) (*DocumentRelatedSuggestion, error) {
	var item DocumentRelatedSuggestion
	err := r.db.WithContext(ctx).
		First(&item, "id = ? AND tenant_id = ? AND source_document_id = ?", suggestionID, tenantID, sourceDocumentID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) DeleteSuggestedRelatedSuggestions(ctx context.Context, tenantID, sourceDocumentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND source_document_id = ? AND status = ?", tenantID, sourceDocumentID, RelatedSuggestionStatusSuggested).
		Delete(&DocumentRelatedSuggestion{}).Error
}

func (r *Repository) HasRelatedSuggestionPair(ctx context.Context, tenantID, sourceDocumentID, targetDocumentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&DocumentRelatedSuggestion{}).
		Where("tenant_id = ? AND source_document_id = ? AND target_document_id = ?", tenantID, sourceDocumentID, targetDocumentID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) CreateRelatedSuggestions(ctx context.Context, items []DocumentRelatedSuggestion) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&items).Error
}

func (r *Repository) UpdateRelatedSuggestion(ctx context.Context, item *DocumentRelatedSuggestion) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *Repository) SearchSimilarDocumentsByEmbedding(
	ctx context.Context,
	tenantID uuid.UUID,
	embeddingValues []float32,
	excludeDocumentID uuid.UUID,
	limit int,
	accessUserID *uuid.UUID,
) ([]DocumentSimilarityScore, error) {
	if limit <= 0 {
		limit = 20
	}

	queryVector := pgvector.NewVector(embeddingValues)
	memberFilter := ""
	args := []interface{}{queryVector, tenantID, excludeDocumentID}
	if accessUserID != nil {
		memberFilter = " AND " + memberDocumentAccessSQL("d")
		args = append(args, *accessUserID, *accessUserID)
	}
	args = append(args, limit)

	var results []DocumentSimilarityScore
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			dc.document_id AS document_id,
			MAX(1 - (dc.embedding <=> ?)) AS similarity
		FROM document_chunks dc
		INNER JOIN documents d ON d.id = dc.document_id
		WHERE dc.embedding IS NOT NULL
			AND d.deleted_at IS NULL
			AND d.tenant_id = ?
			AND d.id <> ?`+memberFilter+`
		GROUP BY dc.document_id
		ORDER BY similarity DESC
		LIMIT ?
	`, args...).Scan(&results).Error
	return results, err
}

func (r *Repository) CountSharedEntities(ctx context.Context, tenantID, sourceDocumentID, targetDocumentID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT LOWER(e1.value) || ':' || e1.entity_type)
		FROM document_entities e1
		INNER JOIN document_entities e2
			ON e1.tenant_id = e2.tenant_id
			AND e1.entity_type = e2.entity_type
			AND LOWER(TRIM(e1.value)) = LOWER(TRIM(e2.value))
		WHERE e1.tenant_id = ?
			AND e1.document_id = ?
			AND e2.document_id = ?
			AND e1.deleted_at IS NULL
			AND e2.deleted_at IS NULL
	`, tenantID, sourceDocumentID, targetDocumentID).Scan(&count).Error
	return count, err
}

func (r *Repository) HasAnyRelationshipBetweenDocuments(ctx context.Context, tenantID, sourceDocumentID, targetDocumentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&DocumentRelationship{}).
		Where(
			"tenant_id = ? AND ((source_document_id = ? AND target_document_id = ?) OR (source_document_id = ? AND target_document_id = ?))",
			tenantID, sourceDocumentID, targetDocumentID, targetDocumentID, sourceDocumentID,
		).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func EnsureDocumentRelatedSuggestionIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_document_related_suggestions_unique_pair
		ON document_related_suggestions (tenant_id, source_document_id, target_document_id)
	`).Error; err != nil {
		return err
	}
	return db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_document_related_suggestions_status
		ON document_related_suggestions (tenant_id, source_document_id, status)
	`).Error
}
