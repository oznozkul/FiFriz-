package audit

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, logEntry *AuditLog) error {
	return r.db.WithContext(ctx).Create(logEntry).Error
}

func (r *Repository) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*auditLogRow, error) {
	var row auditLogRow
	err := r.db.WithContext(ctx).
		Table("audit_logs al").
		Select(`
			al.id,
			al.tenant_id,
			al.user_id,
			COALESCE(u.email, '') AS user_email,
			COALESCE(u.full_name, '') AS user_full_name,
			al.action,
			al.entity_type,
			al.entity_id,
			al.description,
			al.metadata,
			al.ip_address,
			al.user_agent,
			al.created_at
		`).
		Joins("LEFT JOIN users u ON u.id = al.user_id").
		Where("al.id = ? AND al.tenant_id = ?", id, tenantID).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]auditLogRow, int64, error) {
	query := r.db.WithContext(ctx).
		Table("audit_logs al").
		Select(`
			al.id,
			al.tenant_id,
			al.user_id,
			COALESCE(u.email, '') AS user_email,
			COALESCE(u.full_name, '') AS user_full_name,
			al.action,
			al.entity_type,
			al.entity_id,
			al.description,
			al.metadata,
			al.ip_address,
			al.user_agent,
			al.created_at
		`).
		Joins("LEFT JOIN users u ON u.id = al.user_id").
		Where("al.tenant_id = ?", filter.TenantID)

	if filter.Action != "" {
		query = query.Where("al.action = ?", filter.Action)
	}
	if filter.EntityType != "" {
		query = query.Where("al.entity_type = ?", filter.EntityType)
	}
	if filter.UserID != nil {
		query = query.Where("al.user_id = ?", *filter.UserID)
	}
	if filter.StartDate != nil {
		query = query.Where("al.created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("al.created_at <= ?", *filter.EndDate)
	}

	var totalCount int64
	if err := query.Session(&gorm.Session{}).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var rows []auditLogRow
	err := query.
		Order("al.created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows).Error
	return rows, totalCount, err
}
