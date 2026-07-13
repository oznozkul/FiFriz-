package documents

import (
	"context"
	"strings"
	"time"

	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func importantDateSelectColumns() string {
	return `
		did.*,
		creator.full_name AS created_by_user_name,
		assignee.full_name AS assigned_to_user_name,
		d.original_file_name AS document_name
	`
}

func importantDateBaseQuery(db *gorm.DB, ctx context.Context, tenantID uuid.UUID) *gorm.DB {
	return db.WithContext(ctx).
		Table("document_important_dates AS did").
		Select(importantDateSelectColumns()).
		Joins("INNER JOIN documents d ON d.id = did.document_id AND d.tenant_id = did.tenant_id AND d.deleted_at IS NULL").
		Joins("LEFT JOIN users creator ON creator.id = did.created_by_user_id").
		Joins("LEFT JOIN users assignee ON assignee.id = did.assigned_to_user_id").
		Where("did.tenant_id = ? AND did.deleted_at IS NULL", tenantID)
}

func scopeActiveImportantDates(query *gorm.DB) *gorm.DB {
	return query.Where("did.status = ?", ImportantDateStatusActive)
}

func scopeSuggestedImportantDates(query *gorm.DB) *gorm.DB {
	return query.Where("did.status = ?", ImportantDateStatusSuggested)
}

func (r *Repository) applyImportantDateAccessScope(query *gorm.DB, userID uuid.UUID, tenantWide bool) *gorm.DB {
	if tenantWide {
		return query
	}
	return query.Where(
		"(did.created_by_user_id = ? OR did.assigned_to_user_id = ? OR "+permissions.MemberDocumentScopeSQL("d")+")",
		userID, userID, userID, userID,
	)
}

func (r *Repository) ListImportantDatesByDocument(ctx context.Context, tenantID, documentID uuid.UUID) ([]ImportantDateListRow, error) {
	var rows []ImportantDateListRow
	err := scopeActiveImportantDates(importantDateBaseQuery(r.db, ctx, tenantID)).
		Where("did.document_id = ?", documentID).
		Order("did.date ASC, did.created_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) ListSuggestedImportantDatesByDocument(ctx context.Context, tenantID, documentID uuid.UUID) ([]ImportantDateListRow, error) {
	var rows []ImportantDateListRow
	err := scopeSuggestedImportantDates(importantDateBaseQuery(r.db, ctx, tenantID)).
		Where("did.document_id = ?", documentID).
		Order("did.date ASC, did.created_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) DeleteSuggestedImportantDatesByDocument(ctx context.Context, tenantID, documentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ? AND status = ? AND source = ?", tenantID, documentID, ImportantDateStatusSuggested, ImportantDateSourceAI).
		Delete(&DocumentImportantDate{}).Error
}

func (r *Repository) HasSimilarImportantDate(ctx context.Context, tenantID, documentID uuid.UUID, title string, date time.Time) (bool, error) {
	normalizedTitle := normalizeImportantDateLookupTitle(title)
	if normalizedTitle == "" {
		return false, nil
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&DocumentImportantDate{}).
		Where("tenant_id = ? AND document_id = ? AND deleted_at IS NULL", tenantID, documentID).
		Where("date = ?", date.Format("2006-01-02")).
		Where("LOWER(TRIM(title)) = ?", normalizedTitle).
		Where("status IN ?", []string{ImportantDateStatusActive, ImportantDateStatusSuggested}).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func normalizeImportantDateLookupTitle(title string) string {
	return strings.ToLower(strings.TrimSpace(title))
}

func (r *Repository) GetImportantDateByID(ctx context.Context, tenantID, documentID, dateID uuid.UUID) (*DocumentImportantDate, error) {
	var item DocumentImportantDate
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ? AND id = ?", tenantID, documentID, dateID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) GetImportantDateWithDetails(ctx context.Context, tenantID, documentID, dateID uuid.UUID) (*ImportantDateListRow, error) {
	var row ImportantDateListRow
	err := importantDateBaseQuery(r.db, ctx, tenantID).
		Where("did.document_id = ? AND did.id = ?", documentID, dateID).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repository) CreateImportantDate(ctx context.Context, item *DocumentImportantDate) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *Repository) UpdateImportantDate(ctx context.Context, item *DocumentImportantDate) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *Repository) SoftDeleteImportantDate(ctx context.Context, tenantID, documentID, dateID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ? AND id = ?", tenantID, documentID, dateID).
		Delete(&DocumentImportantDate{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) ListImportantDates(ctx context.Context, tenantID, userID uuid.UUID, tenantWide bool, query ImportantDateListQuery) ([]ImportantDateListRow, error) {
	dbQuery := scopeActiveImportantDates(r.applyImportantDateAccessScope(importantDateBaseQuery(r.db, ctx, tenantID), userID, tenantWide))

	if query.Completed != nil {
		dbQuery = dbQuery.Where("did.is_completed = ?", *query.Completed)
	}
	if query.AssignedToMe {
		dbQuery = dbQuery.Where("did.assigned_to_user_id = ?", userID)
	}
	if query.Type != nil && *query.Type != "" {
		dbQuery = dbQuery.Where("did.type = ?", *query.Type)
	}
	if query.From != nil {
		dbQuery = dbQuery.Where("did.date >= ?", query.From.Format("2006-01-02"))
	}
	if query.To != nil {
		dbQuery = dbQuery.Where("did.date <= ?", query.To.Format("2006-01-02"))
	}
	if query.UpcomingOnly {
		today := time.Now().UTC().Truncate(24 * time.Hour)
		dbQuery = dbQuery.Where("did.is_completed = ? AND did.date >= ?", false, today.Format("2006-01-02"))
	}

	var rows []ImportantDateListRow
	err := dbQuery.Order("did.date ASC, did.created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *Repository) GetDueReminders(ctx context.Context, reminderDate time.Time) ([]ImportantDateListRow, error) {
	targetDate := reminderDate.UTC().Truncate(24 * time.Hour).Format("2006-01-02")
	var rows []ImportantDateListRow
	err := r.db.WithContext(ctx).
		Table("document_important_dates AS did").
		Select(importantDateSelectColumns()).
		Joins("INNER JOIN documents d ON d.id = did.document_id AND d.tenant_id = did.tenant_id AND d.deleted_at IS NULL").
		Joins("LEFT JOIN users creator ON creator.id = did.created_by_user_id").
		Joins("LEFT JOIN users assignee ON assignee.id = did.assigned_to_user_id").
		Where("did.deleted_at IS NULL AND did.is_completed = ? AND did.status = ? AND did.reminder_days_before IS NOT NULL", false, ImportantDateStatusActive).
		Where("(did.date - (did.reminder_days_before || ' days')::interval)::date = ?::date", targetDate).
		Order("did.tenant_id ASC, did.date ASC").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) CountImportantDates(ctx context.Context, tenantID, userID uuid.UUID, tenantWide bool, upcomingOnly, overdueOnly, assignedToMe bool) (int64, error) {
	query := r.applyImportantDateAccessScope(
		r.db.WithContext(ctx).Table("document_important_dates AS did").
			Joins("INNER JOIN documents d ON d.id = did.document_id AND d.tenant_id = did.tenant_id AND d.deleted_at IS NULL").
			Where("did.tenant_id = ? AND did.deleted_at IS NULL AND did.is_completed = ? AND did.status = ?", tenantID, false, ImportantDateStatusActive),
		userID,
		tenantWide,
	)

	today := time.Now().UTC().Truncate(24 * time.Hour).Format("2006-01-02")
	if upcomingOnly {
		query = query.Where("did.date >= ?", today)
	}
	if overdueOnly {
		query = query.Where("did.date < ?", today)
	}
	if assignedToMe {
		query = query.Where("did.assigned_to_user_id = ?", userID)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

func EnsureDocumentImportantDateIndexes(db *gorm.DB) error {
	statements := []string{
		`CREATE INDEX IF NOT EXISTS idx_document_important_dates_document_date
		 ON document_important_dates (document_id, date)
		 WHERE deleted_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_document_important_dates_tenant_date
		 ON document_important_dates (tenant_id, date)
		 WHERE deleted_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_document_important_dates_status
		 ON document_important_dates (tenant_id, document_id, status)
		 WHERE deleted_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_document_important_dates_assigned
		 ON document_important_dates (tenant_id, assigned_to_user_id, date)
		 WHERE deleted_at IS NULL AND assigned_to_user_id IS NOT NULL`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) TenantMemberExists(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("tenant_members").
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
