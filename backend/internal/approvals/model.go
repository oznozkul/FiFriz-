package approvals

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApprovalWorkflow struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	DocumentID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"document_id"`
	CreatedByUserID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"created_by_user_id"`
	Title            string     `gorm:"type:text;not null" json:"title"`
	Description      *string    `gorm:"type:text" json:"description"`
	Status           string     `gorm:"type:text;not null;index" json:"status"`
	CurrentStepOrder int        `gorm:"not null;default:1" json:"current_step_order"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	CompletedAt      *time.Time `gorm:"type:timestamptz" json:"completed_at"`
}

func (w *ApprovalWorkflow) BeforeCreate(_ *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

func (ApprovalWorkflow) TableName() string {
	return "approval_workflows"
}

type ApprovalStep struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	WorkflowID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"workflow_id"`
	StepOrder      int        `gorm:"not null;index" json:"step_order"`
	ApproverUserID uuid.UUID  `gorm:"type:uuid;not null;index" json:"approver_user_id"`
	Status         string     `gorm:"type:text;not null" json:"status"`
	Note           *string    `gorm:"type:text" json:"note"`
	DecidedAt      *time.Time `gorm:"type:timestamptz" json:"decided_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (s *ApprovalStep) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (ApprovalStep) TableName() string {
	return "approval_steps"
}
