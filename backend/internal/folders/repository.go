package folders

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

func (r *Repository) List(ctx context.Context, tenantID uuid.UUID) ([]Folder, error) {
	var folders []Folder
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("parent_id NULLS FIRST, name ASC, created_at ASC").
		Find(&folders).Error
	return folders, err
}

func (r *Repository) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*Folder, error) {
	var folder Folder
	err := r.db.WithContext(ctx).First(&folder, "id = ? AND tenant_id = ?", id, tenantID).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *Repository) ExistsByName(ctx context.Context, tenantID uuid.UUID, parentID *uuid.UUID, name string, exceptID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&Folder{}).Where("tenant_id = ? AND name = ?", tenantID, name)
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	if exceptID != nil {
		query = query.Where("id <> ?", *exceptID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) Create(ctx context.Context, folder *Folder) error {
	return r.db.WithContext(ctx).Create(folder).Error
}

func (r *Repository) CreateBatch(ctx context.Context, folders []Folder) error {
	if len(folders) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&folders).Error
}

func (r *Repository) HasSystemFolders(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Folder{}).
		Where("tenant_id = ? AND is_system = ?", tenantID, true).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) Update(ctx context.Context, folder *Folder) error {
	return r.db.WithContext(ctx).Save(folder).Error
}

func (r *Repository) Delete(ctx context.Context, folder *Folder) error {
	return r.db.WithContext(ctx).Delete(folder).Error
}
