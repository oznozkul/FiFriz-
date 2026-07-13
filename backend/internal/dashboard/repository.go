package dashboard

import (
	"context"
	"time"

	"fifriz/internal/documents"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

type recentDocumentRow struct {
	ID               uuid.UUID `gorm:"column:id"`
	OriginalFileName string    `gorm:"column:original_file_name"`
	FileExtension    string    `gorm:"column:file_extension"`
	FileSize         int64     `gorm:"column:file_size"`
	Status           string    `gorm:"column:status"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	FolderName       *string   `gorm:"column:folder_name"`
	IsFavorite       bool      `gorm:"column:is_favorite"`
}

func (r *Repository) GetSummary(ctx context.Context, tenantID, userID uuid.UUID, tenantWide bool) (*SummaryResponse, error) {
	summary := &SummaryResponse{
		RecentDocuments:   []RecentDocument{},
		RecentActivities:  []RecentActivity{},
		DocumentTypeStats: []DocumentTypeStat{},
		TopFolders:        []TopFolder{},
	}

	activeDocuments := r.db.WithContext(ctx).Table("documents").Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if !tenantWide {
		activeDocuments = activeDocuments.Where(permissions.MemberDocumentScopeSQL("documents"), userID, userID)
	}

	if err := activeDocuments.Session(&gorm.Session{}).Count(&summary.TotalDocuments).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Table("folders").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&summary.TotalFolders).Error; err != nil {
		return nil, err
	}

	if err := activeDocuments.Session(&gorm.Session{}).Where("is_favorite = ?", true).Count(&summary.FavoriteDocuments).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Unscoped().Table("documents").
		Where("tenant_id = ? AND deleted_at IS NOT NULL", tenantID).
		Scopes(func(db *gorm.DB) *gorm.DB {
			if tenantWide {
				return db
			}
			return db.Where(permissions.MemberDocumentScopeSQL("documents"), userID, userID)
		}).
		Count(&summary.TrashedDocuments).Error; err != nil {
		return nil, err
	}

	var totalStorage *int64
	if err := activeDocuments.Session(&gorm.Session{}).Select("COALESCE(SUM(file_size), 0)").Scan(&totalStorage).Error; err != nil {
		return nil, err
	}
	if totalStorage != nil {
		summary.TotalStorageBytes = *totalStorage
	}

	monthStart := time.Date(time.Now().UTC().Year(), time.Now().UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	if err := activeDocuments.Session(&gorm.Session{}).Where("created_at >= ?", monthStart).Count(&summary.DocumentsUploadedThisMonth).Error; err != nil {
		return nil, err
	}

	var recentRows []recentDocumentRow
	recentQuery := r.db.WithContext(ctx).
		Table("documents d").
		Select(`
			d.id,
			d.original_file_name,
			d.file_extension,
			d.file_size,
			d.status,
			d.created_at,
			d.is_favorite,
			folders.name AS folder_name
		`).
		Joins("LEFT JOIN folders ON folders.id = d.folder_id AND folders.tenant_id = d.tenant_id AND folders.deleted_at IS NULL").
		Where("d.tenant_id = ? AND d.deleted_at IS NULL", tenantID)
	if !tenantWide {
		recentQuery = recentQuery.Where(permissions.MemberDocumentScopeSQL("d"), userID, userID)
	}
	if err := recentQuery.
		Order("d.created_at DESC").
		Limit(5).
		Scan(&recentRows).Error; err != nil {
		return nil, err
	}

	for _, row := range recentRows {
		summary.RecentDocuments = append(summary.RecentDocuments, RecentDocument{
			ID:               row.ID,
			OriginalFileName: row.OriginalFileName,
			FileExtension:    row.FileExtension,
			FileSize:         row.FileSize,
			Status:           row.Status,
			CreatedAt:        row.CreatedAt,
			FolderName:       row.FolderName,
			IsFavorite:       row.IsFavorite,
		})
	}

	activityQuery := r.db.WithContext(ctx).
		Table("audit_logs").
		Select("id, action, entity_type, description, created_at").
		Where("tenant_id = ?", tenantID)
	if !tenantWide {
		activityQuery = activityQuery.Where("user_id = ?", userID)
	}
	if err := activityQuery.
		Order("created_at DESC").
		Limit(5).
		Scan(&summary.RecentActivities).Error; err != nil {
		return nil, err
	}

	typeStatsQuery := r.db.WithContext(ctx).
		Table("documents").
		Select("LOWER(file_extension) AS extension, COUNT(*) AS count").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if !tenantWide {
		typeStatsQuery = typeStatsQuery.Where(permissions.MemberDocumentScopeSQL("documents"), userID, userID)
	}
	if err := typeStatsQuery.
		Group("LOWER(file_extension)").
		Order("count DESC, extension ASC").
		Scan(&summary.DocumentTypeStats).Error; err != nil {
		return nil, err
	}

	topFoldersJoin := "LEFT JOIN documents d ON d.folder_id = f.id AND d.tenant_id = f.tenant_id AND d.deleted_at IS NULL"
	topFoldersQuery := r.db.WithContext(ctx).
		Table("folders f").
		Select(`
			f.id AS folder_id,
			f.name AS folder_name,
			COUNT(d.id) AS document_count
		`)
	if !tenantWide {
		topFoldersJoin += " AND (" + permissions.MemberDocumentScopeSQL("d") + ")"
		topFoldersQuery = topFoldersQuery.Joins(topFoldersJoin, userID, userID)
	} else {
		topFoldersQuery = topFoldersQuery.Joins(topFoldersJoin)
	}
	if err := topFoldersQuery.
		Where("f.tenant_id = ? AND f.deleted_at IS NULL", tenantID).
		Group("f.id, f.name").
		Order("document_count DESC, f.name ASC").
		Limit(5).
		Scan(&summary.TopFolders).Error; err != nil {
		return nil, err
	}

	if err := r.loadImportantDatesSummary(ctx, tenantID, userID, tenantWide, summary); err != nil {
		return nil, err
	}

	return summary, nil
}

type importantDateRow struct {
	ID               uuid.UUID `gorm:"column:id"`
	DocumentID       uuid.UUID `gorm:"column:document_id"`
	Title            string    `gorm:"column:title"`
	Date             time.Time `gorm:"column:date"`
	Type             *string   `gorm:"column:type"`
	IsCompleted      bool      `gorm:"column:is_completed"`
	DocumentName     string    `gorm:"column:document_name"`
	AssignedToUserID *uuid.UUID `gorm:"column:assigned_to_user_id"`
}

func (r *Repository) importantDatesBaseQuery(ctx context.Context, tenantID uuid.UUID) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("document_important_dates AS did").
		Joins("INNER JOIN documents d ON d.id = did.document_id AND d.tenant_id = did.tenant_id AND d.deleted_at IS NULL").
		Where("did.tenant_id = ? AND did.deleted_at IS NULL", tenantID)
}

func (r *Repository) applyImportantDateScope(query *gorm.DB, userID uuid.UUID, tenantWide bool) *gorm.DB {
	if tenantWide {
		return query
	}
	return query.Where(
		"(did.created_by_user_id = ? OR did.assigned_to_user_id = ? OR "+permissions.MemberDocumentScopeSQL("d")+")",
		userID, userID, userID, userID,
	)
}

func (r *Repository) loadImportantDatesSummary(ctx context.Context, tenantID, userID uuid.UUID, tenantWide bool, summary *SummaryResponse) error {
	today := time.Now().UTC().Truncate(24 * time.Hour).Format("2006-01-02")

	base := r.applyImportantDateScope(r.importantDatesBaseQuery(ctx, tenantID), userID, tenantWide).
		Where("did.is_completed = ? AND did.status = ?", false, documents.ImportantDateStatusActive)

	if err := base.Session(&gorm.Session{}).Where("did.date >= ?", today).Count(&summary.UpcomingImportantDates).Error; err != nil {
		return err
	}
	if err := base.Session(&gorm.Session{}).Where("did.date < ?", today).Count(&summary.OverdueImportantDates).Error; err != nil {
		return err
	}
	if err := base.Session(&gorm.Session{}).Where("did.assigned_to_user_id = ?", userID).Count(&summary.AssignedImportantDates).Error; err != nil {
		return err
	}

	var rows []importantDateRow
	if err := r.applyImportantDateScope(r.importantDatesBaseQuery(ctx, tenantID), userID, tenantWide).
		Select("did.id, did.document_id, did.title, did.date, did.type, did.is_completed, did.assigned_to_user_id, d.original_file_name AS document_name").
		Where("did.is_completed = ? AND did.status = ? AND did.date >= ?", false, documents.ImportantDateStatusActive, today).
		Order("did.date ASC").
		Limit(5).
		Scan(&rows).Error; err != nil {
		return err
	}

	summary.UpcomingImportantDateItems = make([]ImportantDateItem, 0, len(rows))
	for _, row := range rows {
		summary.UpcomingImportantDateItems = append(summary.UpcomingImportantDateItems, ImportantDateItem{
			ID:               row.ID,
			DocumentID:       row.DocumentID,
			Title:            row.Title,
			Date:             row.Date.Format("2006-01-02"),
			Type:             row.Type,
			IsCompleted:      row.IsCompleted,
			DocumentName:     row.DocumentName,
			AssignedToUserID: row.AssignedToUserID,
		})
	}

	return nil
}
