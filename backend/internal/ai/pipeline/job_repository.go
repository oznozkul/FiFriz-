package pipeline

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DocumentAIJob tracks per-step pipeline execution status.
type DocumentAIJob struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_document_ai_jobs_tenant_document_step,priority:1" json:"tenantId"`
	DocumentID  uuid.UUID  `gorm:"type:uuid;not null;index:idx_document_ai_jobs_tenant_document_step,priority:2" json:"documentId"`
	StepName    string     `gorm:"type:text;not null;index:idx_document_ai_jobs_tenant_document_step,priority:3" json:"stepName"`
	Status      string     `gorm:"type:text;not null;index" json:"status"`
	StartedAt   *time.Time `gorm:"type:timestamptz" json:"startedAt"`
	CompletedAt *time.Time `gorm:"type:timestamptz" json:"completedAt"`
	Error       *string    `gorm:"type:text" json:"error"`
	RetryCount  int        `gorm:"not null;default:0" json:"retryCount"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func (j *DocumentAIJob) BeforeCreate(_ *gorm.DB) error {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return nil
}

func (DocumentAIJob) TableName() string {
	return "document_ai_jobs"
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&DocumentAIJob{})
}

// JobRepository implements JobTracker.
type JobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) EnsureSteps(ctx context.Context, tenantID, documentID uuid.UUID, stepNames []string) error {
	for _, stepName := range stepNames {
		var existing DocumentAIJob
		err := r.db.WithContext(ctx).
			Where("tenant_id = ? AND document_id = ? AND step_name = ?", tenantID, documentID, stepName).
			First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := r.db.WithContext(ctx).Create(&DocumentAIJob{
			TenantID:   tenantID,
			DocumentID: documentID,
			StepName:   stepName,
			Status:     StatusPending,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *JobRepository) MarkRunning(ctx context.Context, tenantID, documentID uuid.UUID, stepName string) error {
	now := time.Now().UTC()
	return r.updateStep(ctx, tenantID, documentID, stepName, map[string]interface{}{
		"status":     StatusRunning,
		"started_at": now,
		"error":      nil,
	})
}

func (r *JobRepository) MarkCompleted(ctx context.Context, tenantID, documentID uuid.UUID, stepName string) error {
	now := time.Now().UTC()
	return r.updateStep(ctx, tenantID, documentID, stepName, map[string]interface{}{
		"status":       StatusCompleted,
		"completed_at": now,
		"error":        nil,
	})
}

func (r *JobRepository) MarkFailed(ctx context.Context, tenantID, documentID uuid.UUID, stepName string, stepErr error) error {
	message := ""
	if stepErr != nil {
		message = stepErr.Error()
	}
	return r.updateStep(ctx, tenantID, documentID, stepName, map[string]interface{}{
		"status": StatusFailed,
		"error":  message,
	})
}

func (r *JobRepository) MarkRetrying(ctx context.Context, tenantID, documentID uuid.UUID, stepName string, stepErr error) error {
	message := ""
	if stepErr != nil {
		message = stepErr.Error()
	}
	return r.db.WithContext(ctx).Model(&DocumentAIJob{}).
		Where("tenant_id = ? AND document_id = ? AND step_name = ?", tenantID, documentID, stepName).
		Updates(map[string]interface{}{
			"status":      StatusRetrying,
			"error":       message,
			"retry_count": gorm.Expr("retry_count + 1"),
		}).Error
}

func (r *JobRepository) MarkSkipped(ctx context.Context, tenantID, documentID uuid.UUID, stepName string) error {
	now := time.Now().UTC()
	return r.updateStep(ctx, tenantID, documentID, stepName, map[string]interface{}{
		"status":       StatusSkipped,
		"completed_at": now,
		"error":        nil,
	})
}

func (r *JobRepository) GetProgress(ctx context.Context, tenantID, documentID uuid.UUID) ([]StepProgress, error) {
	var rows []DocumentAIJob
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ?", tenantID, documentID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	progress := make([]StepProgress, 0, len(rows))
	for _, row := range rows {
		item := StepProgress{
			StepName:   row.StepName,
			Status:     row.Status,
			RetryCount: row.RetryCount,
			StartedAt:  formatTimePtr(row.StartedAt),
			CompletedAt: formatTimePtr(row.CompletedAt),
		}
		if row.Error != nil && *row.Error != "" {
			item.Error = row.Error
		}
		progress = append(progress, item)
	}
	return progress, nil
}

func (r *JobRepository) RetryCount(ctx context.Context, tenantID, documentID uuid.UUID, stepName string) (int, error) {
	var row DocumentAIJob
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND document_id = ? AND step_name = ?", tenantID, documentID, stepName).
		First(&row).Error
	if err != nil {
		return 0, err
	}
	return row.RetryCount, nil
}

func (r *JobRepository) updateStep(ctx context.Context, tenantID, documentID uuid.UUID, stepName string, values map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&DocumentAIJob{}).
		Where("tenant_id = ? AND document_id = ? AND step_name = ?", tenantID, documentID, stepName).
		Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("pipeline step record not found: %s", stepName)
	}
	return nil
}
