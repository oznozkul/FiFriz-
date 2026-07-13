package documents

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

func (r *Repository) SearchTextRanked(ctx context.Context, tenantID uuid.UUID, tsQuery string, accessUserID *uuid.UUID) ([]HybridSearchRankedResult, error) {
	if strings.TrimSpace(tsQuery) == "" {
		return nil, nil
	}

	memberFilter := ""
	args := []interface{}{tsQuery, tenantID}
	if accessUserID != nil {
		memberFilter = " AND " + memberDocumentAccessSQL("d")
		args = append(args, *accessUserID, *accessUserID)
	}

	var results []HybridSearchRankedResult
	err := r.db.WithContext(ctx).Raw(`
		WITH search_query AS (
			SELECT to_tsquery('simple', ?) AS query
		)
		SELECT
			d.id AS document_id,
			ts_rank_cd(
				to_tsvector('simple', translate(lower(coalesce(dt.full_text, '')), 'çğıöşüâîû', 'cgiosuaiu')),
				search_query.query
			) AS score,
			ts_headline(
				'simple',
				translate(lower(coalesce(dt.full_text, '')), 'çğıöşüâîû', 'cgiosuaiu'),
				search_query.query,
				'StartSel='''', StopSel='''', MaxWords=18, MinWords=6, ShortWord=3, MaxFragments=2, FragmentDelimiter='' ... '''
			) AS matched_snippet,
			dt.full_text AS full_text
		FROM document_texts dt
		INNER JOIN documents d ON d.id = dt.document_id
		CROSS JOIN search_query
		WHERE d.deleted_at IS NULL
			AND d.tenant_id = ?`+memberFilter+`
			AND search_query.query <> ''::tsquery
			AND to_tsvector('simple', translate(lower(coalesce(dt.full_text, '')), 'çğıöşüâîû', 'cgiosuaiu')) @@ search_query.query
		ORDER BY score DESC, d.created_at DESC
		LIMIT 50
	`, args...).Scan(&results).Error
	return results, err
}

func (r *Repository) SearchDocumentsByEmbeddingGrouped(
	ctx context.Context,
	tenantID uuid.UUID,
	embeddingValues []float32,
	limit int,
	accessUserID *uuid.UUID,
) ([]DocumentSimilarityScore, error) {
	if limit <= 0 {
		limit = 30
	}

	queryVector := pgvector.NewVector(embeddingValues)
	memberFilter := ""
	args := []interface{}{queryVector, tenantID}
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
			AND d.tenant_id = ?`+memberFilter+`
		GROUP BY dc.document_id
		ORDER BY similarity DESC
		LIMIT ?
	`, args...).Scan(&results).Error
	return results, err
}

func (r *Repository) SearchSemanticChunks(
	ctx context.Context,
	tenantID uuid.UUID,
	embeddingValues []float32,
	limit int,
	accessUserID *uuid.UUID,
) ([]HybridSearchSemanticChunk, error) {
	if limit <= 0 {
		limit = 40
	}

	queryVector := pgvector.NewVector(embeddingValues)
	memberFilter := ""
	args := []interface{}{queryVector, tenantID}
	if accessUserID != nil {
		memberFilter = " AND " + memberDocumentAccessSQL("d")
		args = append(args, *accessUserID, *accessUserID)
	}
	args = append(args, queryVector, limit)

	var results []HybridSearchSemanticChunk
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			dc.document_id AS document_id,
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

func (r *Repository) SearchMetadataRanked(ctx context.Context, tenantID uuid.UUID, tsQuery string, accessUserID *uuid.UUID) ([]HybridSearchRankedResult, error) {
	if strings.TrimSpace(tsQuery) == "" {
		return nil, nil
	}

	memberFilter := ""
	args := []interface{}{tsQuery, tenantID}
	if accessUserID != nil {
		memberFilter = " AND " + memberDocumentAccessSQL("d")
		args = append(args, *accessUserID, *accessUserID)
	}

	var results []HybridSearchRankedResult
	err := r.db.WithContext(ctx).Raw(`
		WITH search_query AS (
			SELECT to_tsquery('simple', ?) AS query
		)
		SELECT
			d.id AS document_id,
			ts_rank_cd(
				to_tsvector('simple', translate(lower(coalesce(m.metadata::text, '') || ' ' || coalesce(m.document_type, '')), 'çğıöşüâîû', 'cgiosuaiu')),
				search_query.query
			) AS score,
			'' AS matched_snippet,
			'' AS full_text
		FROM document_ai_metadata m
		INNER JOIN documents d ON d.id = m.document_id AND d.tenant_id = m.tenant_id
		CROSS JOIN search_query
		WHERE d.deleted_at IS NULL
			AND m.tenant_id = ?`+memberFilter+`
			AND search_query.query <> ''::tsquery
			AND to_tsvector('simple', translate(lower(coalesce(m.metadata::text, '') || ' ' || coalesce(m.document_type, '')), 'çğıöşüâîû', 'cgiosuaiu')) @@ search_query.query
		ORDER BY score DESC
		LIMIT 30
	`, args...).Scan(&results).Error
	return results, err
}

func (r *Repository) SearchEntitiesRanked(ctx context.Context, tenantID uuid.UUID, terms []string, accessUserID *uuid.UUID) ([]HybridSearchEntityHit, error) {
	if len(terms) == 0 {
		return nil, nil
	}

	needle := terms[0]
	pattern := "%" + needle + "%"
	memberFilter := ""
	args := []interface{}{needle, pattern, tenantID, pattern, pattern}
	if accessUserID != nil {
		memberFilter = " AND " + memberDocumentAccessSQL("d")
		args = append(args, *accessUserID, *accessUserID)
	}

	var results []HybridSearchEntityHit
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			e.document_id AS document_id,
			e.value AS value,
			e.label AS label,
			e.entity_type AS entity_type,
			CASE
				WHEN LOWER(e.value) = LOWER(?) THEN 1.0
				WHEN LOWER(e.value) LIKE LOWER(?) THEN 0.85
				ELSE 0.7
			END AS score
		FROM document_entities e
		INNER JOIN documents d ON d.id = e.document_id AND d.tenant_id = e.tenant_id
		WHERE e.tenant_id = ?
			AND e.deleted_at IS NULL
			AND d.deleted_at IS NULL`+memberFilter+`
			AND (
				LOWER(e.value) LIKE LOWER(?)
				OR LOWER(COALESCE(e.label, '')) LIKE LOWER(?)
			)
		ORDER BY score DESC
		LIMIT 40
	`, args...).Scan(&results).Error
	return results, err
}

func (r *Repository) ListHybridSearchDocumentContexts(ctx context.Context, tenantID uuid.UUID, documentIDs []uuid.UUID) ([]HybridSearchDocumentContext, error) {
	if len(documentIDs) == 0 {
		return nil, nil
	}

	var rows []HybridSearchDocumentContext
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			d.id AS document_id,
			d.original_file_name AS original_file_name,
			d.file_extension AS file_extension,
			COALESCE(f.name, '') AS folder_name,
			d.source_type AS source_type,
			d.ai_document_type AS ai_document_type,
			d.is_favorite AS is_favorite,
			d.created_at AS created_at
		FROM documents d
		LEFT JOIN folders f ON f.id = d.folder_id AND f.tenant_id = d.tenant_id AND f.deleted_at IS NULL
		WHERE d.tenant_id = ? AND d.id IN ? AND d.deleted_at IS NULL
	`, tenantID, documentIDs).Scan(&rows).Error
	return rows, err
}

func buildFullTextSearchQuery(query string) string {
	terms := extractNormalizedSearchTerms(query)
	return strings.Join(terms, " & ")
}

func buildHybridEntitySearchPattern(query string) []string {
	terms := extractNormalizedSearchTerms(query)
	if len(terms) == 0 {
		return nil
	}
	if len(terms) == 1 {
		return []string{terms[0]}
	}
	return []string{strings.Join(terms, " ")}
}

func normalizeHybridScore(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func fieldMatchScore(haystack, needle string) float64 {
	haystack = strings.TrimSpace(strings.ToLower(haystack))
	needle = strings.TrimSpace(strings.ToLower(needle))
	if haystack == "" || needle == "" {
		return 0
	}
	if haystack == needle {
		return 1
	}
	if strings.Contains(haystack, needle) {
		return 0.85
	}
	for _, term := range strings.Fields(needle) {
		if len(term) >= 2 && strings.Contains(haystack, term) {
			return 0.65
		}
	}
	return 0
}

func entityDedupKey(entityType, value string) string {
	return fmt.Sprintf("%s\x00%s", entityType, strings.ToLower(strings.TrimSpace(value)))
}
