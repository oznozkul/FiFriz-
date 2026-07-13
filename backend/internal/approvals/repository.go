package approvals

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) HasPendingWorkflow(ctx context.Context, tenantID, documentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ApprovalWorkflow{}).
		Where("tenant_id = ? AND document_id = ? AND status = ?", tenantID, documentID, WorkflowStatusPending).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) CreateWorkflowWithSteps(ctx context.Context, workflow *ApprovalWorkflow, steps []ApprovalStep, documentApprovalStatus string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&ApprovalWorkflow{}).
			Where("tenant_id = ? AND document_id = ? AND status = ?", workflow.TenantID, workflow.DocumentID, WorkflowStatusPending).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrActiveWorkflowExists
		}

		if err := tx.Create(workflow).Error; err != nil {
			return err
		}
		for i := range steps {
			steps[i].WorkflowID = workflow.ID
			steps[i].TenantID = workflow.TenantID
		}
		if len(steps) > 0 {
			if err := tx.Create(&steps).Error; err != nil {
				return err
			}
		}

		return tx.Table("documents").
			Where("tenant_id = ? AND id = ?", workflow.TenantID, workflow.DocumentID).
			Update("approval_status", documentApprovalStatus).Error
	})
}

func (r *Repository) ListByDocument(ctx context.Context, tenantID, documentID uuid.UUID) ([]WorkflowListRow, error) {
	var rows []WorkflowListRow
	err := r.db.WithContext(ctx).
		Table("approval_workflows AS aw").
		Select("aw.*, d.original_file_name AS document_name, u.full_name AS created_by_full_name").
		Joins("JOIN documents d ON d.id = aw.document_id AND d.tenant_id = aw.tenant_id").
		Joins("LEFT JOIN users u ON u.id = aw.created_by_user_id").
		Where("aw.tenant_id = ? AND aw.document_id = ?", tenantID, documentID).
		Order("aw.created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) GetWorkflowByID(ctx context.Context, tenantID, workflowID uuid.UUID) (*ApprovalWorkflow, error) {
	var workflow ApprovalWorkflow
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, workflowID).
		First(&workflow).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *Repository) ListStepsByWorkflow(ctx context.Context, tenantID, workflowID uuid.UUID) ([]StepListRow, error) {
	var rows []StepListRow
	err := r.db.WithContext(ctx).
		Table("approval_steps AS s").
		Select("s.*, u.full_name AS approver_full_name").
		Joins("LEFT JOIN users u ON u.id = s.approver_user_id").
		Where("s.tenant_id = ? AND s.workflow_id = ?", tenantID, workflowID).
		Order("s.step_order ASC").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) GetCurrentStep(ctx context.Context, tenantID, workflowID uuid.UUID, stepOrder int) (*ApprovalStep, error) {
	var step ApprovalStep
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND workflow_id = ? AND step_order = ?", tenantID, workflowID, stepOrder).
		First(&step).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

func (r *Repository) ApproveStepAdvance(ctx context.Context, workflow *ApprovalWorkflow, step *ApprovalStep, note *string, documentApprovalStatus string, completed bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		stepUpdates := map[string]interface{}{
			"status":     StepStatusApproved,
			"decided_at": now,
		}
		if note != nil {
			stepUpdates["note"] = *note
		}
		if err := tx.Model(&ApprovalStep{}).
			Where("id = ? AND tenant_id = ? AND workflow_id = ?", step.ID, workflow.TenantID, workflow.ID).
			Updates(stepUpdates).Error; err != nil {
			return err
		}

		workflowUpdates := map[string]interface{}{"updated_at": now}
		if completed {
			workflowUpdates["status"] = WorkflowStatusApproved
			workflowUpdates["completed_at"] = now
		} else {
			workflowUpdates["current_step_order"] = workflow.CurrentStepOrder + 1
		}
		if err := tx.Model(&ApprovalWorkflow{}).
			Where("id = ? AND tenant_id = ?", workflow.ID, workflow.TenantID).
			Updates(workflowUpdates).Error; err != nil {
			return err
		}

		return tx.Table("documents").
			Where("tenant_id = ? AND id = ?", workflow.TenantID, workflow.DocumentID).
			Update("approval_status", documentApprovalStatus).Error
	})
}

func (r *Repository) RejectWorkflow(ctx context.Context, workflow *ApprovalWorkflow, step *ApprovalStep, note *string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		stepUpdates := map[string]interface{}{
			"status":     StepStatusRejected,
			"decided_at": now,
		}
		if note != nil {
			stepUpdates["note"] = *note
		}
		if err := tx.Model(&ApprovalStep{}).
			Where("id = ? AND tenant_id = ? AND workflow_id = ?", step.ID, workflow.TenantID, workflow.ID).
			Updates(stepUpdates).Error; err != nil {
			return err
		}

		if err := tx.Model(&ApprovalWorkflow{}).
			Where("id = ? AND tenant_id = ?", workflow.ID, workflow.TenantID).
			Updates(map[string]interface{}{
				"status":       WorkflowStatusRejected,
				"completed_at": now,
				"updated_at":   now,
			}).Error; err != nil {
			return err
		}

		return tx.Table("documents").
			Where("tenant_id = ? AND id = ?", workflow.TenantID, workflow.DocumentID).
			Update("approval_status", "rejected").Error
	})
}

func (r *Repository) CancelWorkflow(ctx context.Context, workflow *ApprovalWorkflow) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		if err := tx.Model(&ApprovalWorkflow{}).
			Where("id = ? AND tenant_id = ?", workflow.ID, workflow.TenantID).
			Updates(map[string]interface{}{
				"status":       WorkflowStatusCancelled,
				"completed_at": now,
				"updated_at":   now,
			}).Error; err != nil {
			return err
		}

		return tx.Table("documents").
			Where("tenant_id = ? AND id = ?", workflow.TenantID, workflow.DocumentID).
			Update("approval_status", "none").Error
	})
}

func (r *Repository) ListPendingForApprover(ctx context.Context, tenantID, userID uuid.UUID) ([]PendingApprovalResponse, error) {
	var rows []PendingApprovalResponse
	err := r.db.WithContext(ctx).
		Table("approval_steps AS s").
		Select(`
			aw.id AS workflow_id,
			s.id AS step_id,
			aw.document_id AS document_id,
			d.original_file_name AS document_name,
			aw.title AS workflow_title,
			s.step_order AS step_order,
			aw.created_by_user_id AS created_by_user_id,
			u.full_name AS created_by_full_name,
			aw.created_at AS created_at
		`).
		Joins("JOIN approval_workflows aw ON aw.id = s.workflow_id AND aw.tenant_id = s.tenant_id").
		Joins("JOIN documents d ON d.id = aw.document_id AND d.tenant_id = aw.tenant_id").
		Joins("LEFT JOIN users u ON u.id = aw.created_by_user_id").
		Where("s.tenant_id = ? AND s.approver_user_id = ? AND aw.status = ? AND s.step_order = aw.current_step_order AND s.status = ?",
			tenantID, userID, WorkflowStatusPending, StepStatusPending).
		Order("aw.created_at DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) IsWorkflowApprover(ctx context.Context, tenantID, workflowID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ApprovalStep{}).
		Where("tenant_id = ? AND workflow_id = ? AND approver_user_id = ?", tenantID, workflowID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) GetMaxStepOrder(ctx context.Context, tenantID, workflowID uuid.UUID) (int, error) {
	var maxOrder *int
	err := r.db.WithContext(ctx).
		Model(&ApprovalStep{}).
		Select("MAX(step_order)").
		Where("tenant_id = ? AND workflow_id = ?", tenantID, workflowID).
		Scan(&maxOrder).Error
	if err != nil {
		return 0, err
	}
	if maxOrder == nil {
		return 0, nil
	}
	return *maxOrder, nil
}

var ErrActiveWorkflowExists = errors.New("active workflow already exists for document")
