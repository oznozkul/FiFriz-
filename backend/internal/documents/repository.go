package documents

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

type DocumentSearchResult struct {
	DocumentID       uuid.UUID `gorm:"column:document_id"`
	Title            string    `gorm:"column:title"`
	OriginalFileName string    `gorm:"column:original_file_name"`
	FileExtension    string    `gorm:"column:file_extension"`
	MatchedSnippet   string    `gorm:"column:matched_snippet"`
	FullText         string    `gorm:"column:full_text"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

type SemanticSearchResult struct {
	DocumentID    uuid.UUID `gorm:"column:document_id"`
	DocumentTitle string    `gorm:"column:document_title"`
	ChunkID       uuid.UUID `gorm:"column:chunk_id"`
	ChunkIndex    int       `gorm:"column:chunk_index"`
	Content       string    `gorm:"column:content"`
	Similarity    float64   `gorm:"column:similarity"`
}

type ChunkSearchFilter struct {
	DocumentID          *uuid.UUID
	FolderID            *uuid.UUID
	TagIDs              []uuid.UUID
	DocumentIDs         []uuid.UUID
	RestrictToDocuments bool
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, tenantID uuid.UUID, accessUserID *uuid.UUID) ([]Document, error) {
	var documents []Document
	query := r.db.WithContext(ctx).
		Model(&Document{}).
		Select("documents.*, folders.name AS folder_name").
		Joins("LEFT JOIN folders ON folders.id = documents.folder_id AND folders.tenant_id = documents.tenant_id AND folders.deleted_at IS NULL").
		Preload("Text").
		Where("documents.tenant_id = ?", tenantID).
		Order("documents.created_at DESC")
	if accessUserID != nil {
		query = query.Scopes(memberDocumentScope(*accessUserID))
	}
	err := query.Find(&documents).Error
	return documents, err
}

func (r *Repository) ListByFolder(ctx context.Context, tenantID uuid.UUID, folderID *uuid.UUID, accessUserID *uuid.UUID) ([]Document, error) {
	var documents []Document
	query := r.db.WithContext(ctx).
		Model(&Document{}).
		Select("documents.*, folders.name AS folder_name").
		Joins("LEFT JOIN folders ON folders.id = documents.folder_id AND folders.tenant_id = documents.tenant_id AND folders.deleted_at IS NULL").
		Preload("Text").
		Where("documents.tenant_id = ?", tenantID).
		Order("documents.created_at DESC")
	if accessUserID != nil {
		query = query.Scopes(memberDocumentScope(*accessUserID))
	}
	if folderID == nil {
		query = query.Where("documents.folder_id IS NULL")
	} else {
		query = query.Where("documents.folder_id = ?", *folderID)
	}
	err := query.Find(&documents).Error
	return documents, err
}

func (r *Repository) ListDeleted(ctx context.Context, tenantID uuid.UUID, accessUserID *uuid.UUID) ([]Document, error) {
	var documents []Document
	query := r.db.WithContext(ctx).Unscoped().
		Model(&Document{}).
		Select("documents.*, folders.name AS folder_name").
		Joins("LEFT JOIN folders ON folders.id = documents.folder_id AND folders.tenant_id = documents.tenant_id").
		Preload("Text").
		Where("documents.tenant_id = ? AND documents.deleted_at IS NOT NULL", tenantID).
		Order("documents.deleted_at DESC")
	if accessUserID != nil {
		query = query.Scopes(memberDocumentScope(*accessUserID))
	}
	err := query.Find(&documents).Error
	return documents, err
}

func (r *Repository) ListFavorites(ctx context.Context, tenantID uuid.UUID, accessUserID *uuid.UUID) ([]Document, error) {
	var documents []Document
	query := r.db.WithContext(ctx).
		Model(&Document{}).
		Select("documents.*, folders.name AS folder_name").
		Joins("LEFT JOIN folders ON folders.id = documents.folder_id AND folders.tenant_id = documents.tenant_id AND folders.deleted_at IS NULL").
		Preload("Text").
		Where("documents.tenant_id = ? AND documents.is_favorite = ?", tenantID, true).
		Order("documents.favorited_at DESC NULLS LAST, documents.created_at DESC")
	if accessUserID != nil {
		query = query.Scopes(memberDocumentScope(*accessUserID))
	}
	err := query.Find(&documents).Error
	return documents, err
}

func (r *Repository) ListByTag(ctx context.Context, tenantID uuid.UUID, tagID uuid.UUID, folderID *uuid.UUID, accessUserID *uuid.UUID) ([]Document, error) {
	var documents []Document
	query := r.db.WithContext(ctx).
		Model(&Document{}).
		Select("documents.*, folders.name AS folder_name").
		Joins("INNER JOIN document_tags ON document_tags.document_id = documents.id AND document_tags.tenant_id = documents.tenant_id").
		Joins("LEFT JOIN folders ON folders.id = documents.folder_id AND folders.tenant_id = documents.tenant_id AND folders.deleted_at IS NULL").
		Preload("Text").
		Where("documents.tenant_id = ? AND document_tags.tag_id = ?", tenantID, tagID).
		Order("documents.created_at DESC")
	if accessUserID != nil {
		query = query.Scopes(memberDocumentScope(*accessUserID))
	}
	if folderID != nil {
		query = query.Where("documents.folder_id = ?", *folderID)
	}
	err := query.Find(&documents).Error
	return documents, err
}

func (r *Repository) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*Document, error) {
	var document Document
	err := r.db.WithContext(ctx).
		Model(&Document{}).
		Select("documents.*, folders.name AS folder_name").
		Joins("LEFT JOIN folders ON folders.id = documents.folder_id AND folders.tenant_id = documents.tenant_id AND folders.deleted_at IS NULL").
		Preload("Text").
		First(&document, "documents.id = ? AND documents.tenant_id = ?", id, tenantID).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *Repository) FolderExists(ctx context.Context, tenantID uuid.UUID, folderID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("folders").
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", folderID, tenantID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) Create(ctx context.Context, document *Document) error {
	return r.db.WithContext(ctx).Create(document).Error
}

func (r *Repository) Update(ctx context.Context, document *Document) error {
	return r.db.WithContext(ctx).Save(document).Error
}

func (r *Repository) Delete(ctx context.Context, document *Document) error {
	return r.db.WithContext(ctx).Delete(document).Error
}

func (r *Repository) GetDeletedByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*Document, error) {
	var document Document
	err := r.db.WithContext(ctx).Unscoped().Where("id = ? AND tenant_id = ? AND deleted_at IS NOT NULL", id, tenantID).First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *Repository) Restore(ctx context.Context, document *Document) error {
	return r.db.WithContext(ctx).Unscoped().Model(document).Update("deleted_at", nil).Error
}

func (r *Repository) PermanentlyDelete(ctx context.Context, document *Document) error {
	return r.db.WithContext(ctx).Unscoped().Delete(document).Error
}

func (r *Repository) UpsertText(ctx context.Context, documentText *DocumentText) error {
	return r.db.WithContext(ctx).Where("document_id = ?", documentText.DocumentID).Assign(DocumentText{
		FullText: documentText.FullText,
	}).FirstOrCreate(documentText).Error
}

func (r *Repository) SaveTextAndChunks(ctx context.Context, documentText *DocumentText, chunks []DocumentChunk) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("document_id = ?", documentText.DocumentID).Assign(DocumentText{
			FullText: documentText.FullText,
		}).FirstOrCreate(documentText).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Where("document_id = ?", documentText.DocumentID).Delete(&DocumentChunk{}).Error; err != nil {
			return err
		}
		if len(chunks) == 0 {
			return nil
		}
		return tx.Create(&chunks).Error
	})
}

func (r *Repository) SaveTextChunksAndStatus(
	ctx context.Context,
	tenantID, documentID uuid.UUID,
	status DocumentStatus,
	documentText *DocumentText,
	chunks []DocumentChunk,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Document{}).
			Where("id = ? AND tenant_id = ?", documentID, tenantID).
			Update("status", status).Error; err != nil {
			return err
		}

		if documentText == nil {
			return nil
		}

		if err := tx.Where("document_id = ?", documentText.DocumentID).Assign(DocumentText{
			FullText: documentText.FullText,
		}).FirstOrCreate(documentText).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Where("document_id = ?", documentText.DocumentID).Delete(&DocumentChunk{}).Error; err != nil {
			return err
		}
		if len(chunks) == 0 {
			return nil
		}
		return tx.Create(&chunks).Error
	})
}

func (r *Repository) RecoverStuckProcessingWithText(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Exec(`
		UPDATE documents
		SET status = ?, updated_at = NOW()
		WHERE status = ?
		  AND deleted_at IS NULL
		  AND EXISTS (
		    SELECT 1 FROM document_texts dt WHERE dt.document_id = documents.id
		  )`, DocumentStatusReady, DocumentStatusProcessing)
	return result.RowsAffected, result.Error
}

func (r *Repository) GetTextByDocumentID(ctx context.Context, documentID uuid.UUID) (*DocumentText, error) {
	var documentText DocumentText
	err := r.db.WithContext(ctx).First(&documentText, "document_id = ?", documentID).Error
	if err != nil {
		return nil, err
	}
	return &documentText, nil
}

func (r *Repository) DeleteTextByDocumentID(ctx context.Context, documentID uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Where("document_id = ?", documentID).Delete(&DocumentText{}).Error
}

func (r *Repository) DeleteChunksByDocumentID(ctx context.Context, documentID uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Where("document_id = ?", documentID).Delete(&DocumentChunk{}).Error
}

func (r *Repository) ReplaceChunks(ctx context.Context, documentID uuid.UUID, chunks []DocumentChunk) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("document_id = ?", documentID).Delete(&DocumentChunk{}).Error; err != nil {
			return err
		}

		if len(chunks) == 0 {
			return nil
		}

		return tx.Create(&chunks).Error
	})
}

func (r *Repository) ListChunksByDocumentID(ctx context.Context, documentID uuid.UUID) ([]DocumentChunk, error) {
	var chunks []DocumentChunk
	err := r.db.WithContext(ctx).
		Select("id", "document_id", "chunk_index", "content", "token_estimate", "created_at", "updated_at").
		Where("document_id = ?", documentID).
		Order("chunk_index ASC").
		Find(&chunks).Error
	return chunks, err
}

func (r *Repository) SearchText(ctx context.Context, tenantID uuid.UUID, query string, accessUserID *uuid.UUID) ([]DocumentSearchResult, error) {
	var results []DocumentSearchResult
	memberFilter := ""
	args := []interface{}{query, tenantID}
	if accessUserID != nil {
		memberFilter = " AND " + memberDocumentAccessSQL("d")
		args = append(args, *accessUserID, *accessUserID)
	}

	err := r.db.WithContext(ctx).Raw(`
		WITH search_query AS (
			SELECT to_tsquery('simple', ?) AS query
		)
		SELECT
			d.id AS document_id,
			d.original_file_name AS title,
			d.original_file_name AS original_file_name,
			d.file_extension AS file_extension,
			d.created_at AS created_at,
			dt.full_text AS full_text,
			ts_headline(
				'simple',
				translate(lower(coalesce(dt.full_text, '')), 'çğıöşüâîû', 'cgiosuaiu'),
				search_query.query,
				'StartSel='''', StopSel='''', MaxWords=18, MinWords=6, ShortWord=3, MaxFragments=2, FragmentDelimiter='' ... '''
			) AS matched_snippet
		FROM document_texts dt
		INNER JOIN documents d ON d.id = dt.document_id
		CROSS JOIN search_query
		WHERE d.deleted_at IS NULL
			AND d.tenant_id = ?`+memberFilter+`
			AND to_tsvector('simple', translate(lower(coalesce(dt.full_text, '')), 'çğıöşüâîû', 'cgiosuaiu')) @@ search_query.query
		ORDER BY ts_rank_cd(to_tsvector('simple', translate(lower(coalesce(dt.full_text, '')), 'çğıöşüâîû', 'cgiosuaiu')), search_query.query) DESC, d.created_at DESC
		LIMIT 50
	`, args...).Scan(&results).Error
	return results, err
}

func (r *Repository) SearchChunksByEmbedding(ctx context.Context, tenantID uuid.UUID, embeddingValues []float32, limit int, accessUserID *uuid.UUID) ([]SemanticSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	queryVector := pgvector.NewVector(embeddingValues)

	memberFilter := ""
	args := []interface{}{queryVector, tenantID}
	if accessUserID != nil {
		memberFilter = " AND " + memberDocumentAccessSQL("d")
		args = append(args, *accessUserID, *accessUserID)
	}
	args = append(args, queryVector, limit)

	var results []SemanticSearchResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			dc.document_id AS document_id,
			dc.id AS chunk_id,
			dc.chunk_index AS chunk_index,
			dc.content AS content,
			1 - (dc.embedding <=> ?) AS similarity
		FROM document_chunks dc
		INNER JOIN documents d ON d.id = dc.document_id
		WHERE dc.embedding IS NOT NULL
			AND d.deleted_at IS NULL
			AND d.tenant_id = ?`+memberFilter+`
		ORDER BY dc.embedding <=> ?
		LIMIT ?
	`, args...).Scan(&results).Error
	return results, err
}

func (r *Repository) SearchChunksByEmbeddingFiltered(
	ctx context.Context,
	tenantID uuid.UUID,
	embeddingValues []float32,
	limit int,
	accessUserID *uuid.UUID,
	filter ChunkSearchFilter,
) ([]SemanticSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if filter.RestrictToDocuments && len(filter.DocumentIDs) == 0 {
		return nil, nil
	}

	queryVector := pgvector.NewVector(embeddingValues)

	memberFilter := ""
	args := []interface{}{queryVector, tenantID}
	if accessUserID != nil {
		memberFilter = " AND " + memberDocumentAccessSQL("d")
		args = append(args, *accessUserID, *accessUserID)
	}

	extraFilter := ""
	if filter.DocumentID != nil {
		extraFilter += " AND d.id = ?"
		args = append(args, *filter.DocumentID)
	}
	if filter.FolderID != nil {
		extraFilter += " AND d.folder_id = ?"
		args = append(args, *filter.FolderID)
	}
	if len(filter.TagIDs) > 0 {
		extraFilter += `
			AND EXISTS (
				SELECT 1 FROM document_tags dt
				WHERE dt.tenant_id = d.tenant_id
					AND dt.document_id = d.id
					AND dt.tag_id IN ?
			)`
		args = append(args, filter.TagIDs)
	}
	if filter.RestrictToDocuments {
		extraFilter += " AND d.id IN ?"
		args = append(args, filter.DocumentIDs)
	}

	args = append(args, queryVector, limit)

	var results []SemanticSearchResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			dc.document_id AS document_id,
			d.original_file_name AS document_title,
			dc.id AS chunk_id,
			dc.chunk_index AS chunk_index,
			dc.content AS content,
			1 - (dc.embedding <=> ?) AS similarity
		FROM document_chunks dc
		INNER JOIN documents d ON d.id = dc.document_id
		WHERE dc.embedding IS NOT NULL
			AND d.deleted_at IS NULL
			AND d.tenant_id = ?`+memberFilter+extraFilter+`
		ORDER BY dc.embedding <=> ?
		LIMIT ?
	`, args...).Scan(&results).Error
	return results, err
}

func (r *Repository) ListLocalStorageDocuments(ctx context.Context, tenantID uuid.UUID) ([]Document, error) {
	var documents []Document
	err := r.db.WithContext(ctx).
		Model(&Document{}).
		Where("tenant_id = ?", tenantID).
		Where("deleted_at IS NULL").
		Where("storage_provider IS NULL OR storage_provider = '' OR storage_provider = ?", "local").
		Find(&documents).Error
	return documents, err
}
