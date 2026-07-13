package ai

import (
	"context"
	"time"

	"fifriz/internal/documents"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	upcomingInsightDays       = 30
	recentImportantDays       = 14
	relatedSuggestionMinScore = 0.70
	insightDocumentLimit      = 5
)

type insightDocumentRow struct {
	DocumentID       uuid.UUID `gorm:"column:document_id"`
	OriginalFileName string    `gorm:"column:original_file_name"`
	FileExtension    string    `gorm:"column:file_extension"`
}

type insightTagRow struct {
	TagID      uuid.UUID `gorm:"column:tag_id"`
	TagName    string    `gorm:"column:tag_name"`
	UsageCount int       `gorm:"column:usage_count"`
}

type insightFolderRow struct {
	FolderID      uuid.UUID `gorm:"column:folder_id"`
	FolderName    string    `gorm:"column:folder_name"`
	DocumentCount int       `gorm:"column:document_count"`
}

type insightClassificationCount struct {
	DocumentType string `gorm:"column:ai_document_type"`
	Count        int    `gorm:"column:count"`
}

type insightRelatedPairRow struct {
	SourceDocumentID       uuid.UUID `gorm:"column:source_document_id"`
	TargetDocumentID       uuid.UUID `gorm:"column:target_document_id"`
	SourceOriginalFileName string    `gorm:"column:source_original_file_name"`
	TargetOriginalFileName string    `gorm:"column:target_original_file_name"`
	SourceFileExtension    string    `gorm:"column:source_file_extension"`
	TargetFileExtension    string    `gorm:"column:target_file_extension"`
	Similarity             float64   `gorm:"column:similarity"`
}

type WorkspaceInsightsRepository struct {
	db *gorm.DB
}

func NewWorkspaceInsightsRepository(db *gorm.DB) *WorkspaceInsightsRepository {
	return &WorkspaceInsightsRepository{db: db}
}

func (r *WorkspaceInsightsRepository) accessibleDocumentsQuery(ctx context.Context, tenantID, userID uuid.UUID, tenantWide bool) *gorm.DB {
	query := r.db.WithContext(ctx).
		Table("documents d").
		Where("d.tenant_id = ? AND d.deleted_at IS NULL", tenantID)
	if !tenantWide {
		query = query.Where(permissions.MemberDocumentScopeSQL("d"), userID, userID)
	}
	return query
}

func (r *WorkspaceInsightsRepository) CountUpcomingImportantDates(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
	withinDays int,
) (int64, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	until := today.AddDate(0, 0, withinDays)

	var count int64
	query := r.db.WithContext(ctx).
		Table("document_important_dates AS did").
		Joins("INNER JOIN documents d ON d.id = did.document_id AND d.tenant_id = did.tenant_id AND d.deleted_at IS NULL").
		Where("did.tenant_id = ? AND did.deleted_at IS NULL AND did.is_completed = ? AND did.status = ?", tenantID, false, documents.ImportantDateStatusActive).
		Where("did.date >= ? AND did.date <= ?", today.Format("2006-01-02"), until.Format("2006-01-02"))

	if !tenantWide {
		query = query.Where(
			"(did.created_by_user_id = ? OR did.assigned_to_user_id = ? OR "+permissions.MemberDocumentScopeSQL("d")+")",
			userID, userID, userID, userID,
		)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *WorkspaceInsightsRepository) ListUpcomingImportantDateDocuments(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
	withinDays int,
	limit int,
) ([]insightDocumentRow, error) {
	if limit <= 0 {
		limit = insightDocumentLimit
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	until := today.AddDate(0, 0, withinDays)

	var rows []insightDocumentRow
	query := r.db.WithContext(ctx).
		Table("document_important_dates AS did").
		Select("DISTINCT d.id AS document_id, d.original_file_name, d.file_extension").
		Joins("INNER JOIN documents d ON d.id = did.document_id AND d.tenant_id = did.tenant_id AND d.deleted_at IS NULL").
		Where("did.tenant_id = ? AND did.deleted_at IS NULL AND did.is_completed = ? AND did.status = ?", tenantID, false, documents.ImportantDateStatusActive).
		Where("did.date >= ? AND did.date <= ?", today.Format("2006-01-02"), until.Format("2006-01-02"))

	if !tenantWide {
		query = query.Where(
			"(did.created_by_user_id = ? OR did.assigned_to_user_id = ? OR "+permissions.MemberDocumentScopeSQL("d")+")",
			userID, userID, userID, userID,
		)
	}

	err := query.Order("did.date ASC").Limit(limit).Scan(&rows).Error
	return rows, err
}

func (r *WorkspaceInsightsRepository) CountContractUpcomingDates(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
	withinDays int,
) (int64, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	until := today.AddDate(0, 0, withinDays)

	var count int64
	query := r.db.WithContext(ctx).
		Table("document_important_dates AS did").
		Select("COUNT(DISTINCT d.id)").
		Joins("INNER JOIN documents d ON d.id = did.document_id AND d.tenant_id = did.tenant_id AND d.deleted_at IS NULL").
		Where("did.tenant_id = ? AND did.deleted_at IS NULL AND did.is_completed = ? AND did.status = ?", tenantID, false, documents.ImportantDateStatusActive).
		Where("did.date >= ? AND did.date <= ?", today.Format("2006-01-02"), until.Format("2006-01-02")).
		Where("d.ai_document_type = ?", "contract")

	if !tenantWide {
		query = query.Where(
			"(did.created_by_user_id = ? OR did.assigned_to_user_id = ? OR "+permissions.MemberDocumentScopeSQL("d")+")",
			userID, userID, userID, userID,
		)
	}

	err := query.Scan(&count).Error
	return count, err
}

func (r *WorkspaceInsightsRepository) ListRecentImportantDocuments(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
	limit int,
) ([]insightDocumentRow, error) {
	if limit <= 0 {
		limit = insightDocumentLimit
	}

	since := time.Now().UTC().AddDate(0, 0, -recentImportantDays)
	var rows []insightDocumentRow
	err := r.accessibleDocumentsQuery(ctx, tenantID, userID, tenantWide).
		Select("d.id AS document_id, d.original_file_name, d.file_extension").
		Where("d.created_at >= ?", since).
		Where("d.ai_document_type IN ?", []string{"contract", "invoice", "insurance_policy", "court_decision", "warranty_document"}).
		Order("d.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *WorkspaceInsightsRepository) ListClassificationCounts(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
) ([]insightClassificationCount, error) {
	var rows []insightClassificationCount
	err := r.accessibleDocumentsQuery(ctx, tenantID, userID, tenantWide).
		Select("d.ai_document_type, COUNT(*) AS count").
		Where("d.ai_document_type IN ?", []string{"contract", "invoice", "insurance_policy"}).
		Group("d.ai_document_type").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *WorkspaceInsightsRepository) CountMissingMetadata(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
) (int64, error) {
	var count int64
	err := r.accessibleDocumentsQuery(ctx, tenantID, userID, tenantWide).
		Joins("LEFT JOIN document_ai_metadata m ON m.document_id = d.id AND m.tenant_id = d.tenant_id").
		Where("d.status IN ?", []string{string(documents.DocumentStatusReady), string(documents.DocumentStatusProcessed)}).
		Where("m.id IS NULL").
		Count(&count).Error
	return count, err
}

func (r *WorkspaceInsightsRepository) ListMissingMetadataDocuments(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
	limit int,
) ([]insightDocumentRow, error) {
	if limit <= 0 {
		limit = insightDocumentLimit
	}

	var rows []insightDocumentRow
	err := r.accessibleDocumentsQuery(ctx, tenantID, userID, tenantWide).
		Select("d.id AS document_id, d.original_file_name, d.file_extension").
		Joins("LEFT JOIN document_ai_metadata m ON m.document_id = d.id AND m.tenant_id = d.tenant_id").
		Where("d.status IN ?", []string{string(documents.DocumentStatusReady), string(documents.DocumentStatusProcessed)}).
		Where("m.id IS NULL").
		Order("d.updated_at DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *WorkspaceInsightsRepository) CountFailedDocuments(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
) (int64, error) {
	var count int64
	err := r.accessibleDocumentsQuery(ctx, tenantID, userID, tenantWide).
		Where("d.status = ?", documents.DocumentStatusFailed).
		Count(&count).Error
	return count, err
}

func (r *WorkspaceInsightsRepository) ListFailedDocuments(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
	limit int,
) ([]insightDocumentRow, error) {
	if limit <= 0 {
		limit = insightDocumentLimit
	}

	var rows []insightDocumentRow
	err := r.accessibleDocumentsQuery(ctx, tenantID, userID, tenantWide).
		Select("d.id AS document_id, d.original_file_name, d.file_extension").
		Where("d.status = ?", documents.DocumentStatusFailed).
		Order("d.updated_at DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *WorkspaceInsightsRepository) ListHighSimilaritySuggestions(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
	limit int,
) ([]insightRelatedPairRow, error) {
	if limit <= 0 {
		limit = insightDocumentLimit
	}

	query := r.db.WithContext(ctx).
		Table("document_related_suggestions AS drs").
		Select(`
			drs.source_document_id,
			drs.target_document_id,
			sd.original_file_name AS source_original_file_name,
			td.original_file_name AS target_original_file_name,
			sd.file_extension AS source_file_extension,
			td.file_extension AS target_file_extension,
			drs.similarity
		`).
		Joins("INNER JOIN documents sd ON sd.id = drs.source_document_id AND sd.tenant_id = drs.tenant_id AND sd.deleted_at IS NULL").
		Joins("INNER JOIN documents td ON td.id = drs.target_document_id AND td.tenant_id = drs.tenant_id AND td.deleted_at IS NULL").
		Where("drs.tenant_id = ? AND drs.status = ? AND drs.similarity >= ?", tenantID, documents.RelatedSuggestionStatusSuggested, relatedSuggestionMinScore)

	if !tenantWide {
		query = query.Where(permissions.MemberDocumentScopeSQL("sd"), userID, userID)
	}

	var rows []insightRelatedPairRow
	err := query.Order("drs.similarity DESC, drs.created_at DESC").Limit(limit).Scan(&rows).Error
	return rows, err
}

func (r *WorkspaceInsightsRepository) CountHighSimilaritySuggestions(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
) (int64, error) {
	query := r.db.WithContext(ctx).
		Table("document_related_suggestions AS drs").
		Joins("INNER JOIN documents sd ON sd.id = drs.source_document_id AND sd.tenant_id = drs.tenant_id AND sd.deleted_at IS NULL").
		Joins("INNER JOIN documents td ON td.id = drs.target_document_id AND td.tenant_id = drs.tenant_id AND td.deleted_at IS NULL").
		Where("drs.tenant_id = ? AND drs.status = ? AND drs.similarity >= ?", tenantID, documents.RelatedSuggestionStatusSuggested, relatedSuggestionMinScore)

	if !tenantWide {
		query = query.Where(permissions.MemberDocumentScopeSQL("sd"), userID, userID)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

func (r *WorkspaceInsightsRepository) ListTopTags(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
	limit int,
) ([]insightTagRow, error) {
	if limit <= 0 {
		limit = 5
	}

	query := r.db.WithContext(ctx).
		Table("tags t").
		Select("t.id AS tag_id, t.name AS tag_name, COUNT(dt.document_id) AS usage_count").
		Joins("INNER JOIN document_tags dt ON dt.tag_id = t.id AND dt.tenant_id = t.tenant_id").
		Joins("INNER JOIN documents d ON d.id = dt.document_id AND d.tenant_id = dt.tenant_id AND d.deleted_at IS NULL").
		Where("t.tenant_id = ? AND t.deleted_at IS NULL", tenantID)

	if !tenantWide {
		query = query.Where(permissions.MemberDocumentScopeSQL("d"), userID, userID)
	}

	var rows []insightTagRow
	err := query.Group("t.id, t.name").Order("usage_count DESC, t.name ASC").Limit(limit).Scan(&rows).Error
	return rows, err
}

func (r *WorkspaceInsightsRepository) ListActiveFolders(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
	limit int,
) ([]insightFolderRow, error) {
	if limit <= 0 {
		limit = 5
	}

	query := r.db.WithContext(ctx).
		Table("folders f").
		Select("f.id AS folder_id, f.name AS folder_name, COUNT(d.id) AS document_count").
		Joins("INNER JOIN documents d ON d.folder_id = f.id AND d.tenant_id = f.tenant_id AND d.deleted_at IS NULL").
		Where("f.tenant_id = ? AND f.deleted_at IS NULL", tenantID)

	if !tenantWide {
		query = query.Where(permissions.MemberDocumentScopeSQL("d"), userID, userID)
	}

	var rows []insightFolderRow
	err := query.Group("f.id, f.name").Order("document_count DESC, f.name ASC").Limit(limit).Scan(&rows).Error
	return rows, err
}
