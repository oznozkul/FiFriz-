package approvals

import (
	"time"

	"github.com/google/uuid"
)

type CreateWorkflowStepInput struct {
	ApproverUserID uuid.UUID `json:"approverUserId"`
	StepOrder      int       `json:"stepOrder"`
}

type CreateWorkflowRequest struct {
	Title       string                    `json:"title"`
	Description *string                   `json:"description"`
	Steps       []CreateWorkflowStepInput `json:"steps"`
}

type DecideWorkflowRequest struct {
	Note *string `json:"note"`
}

type WorkflowStepResponse struct {
	ID                 uuid.UUID  `json:"id"`
	StepOrder          int        `json:"stepOrder"`
	ApproverUserID     uuid.UUID  `json:"approverUserId"`
	ApproverFullName   string     `json:"approverFullName"`
	Status             string     `json:"status"`
	Note               *string    `json:"note"`
	DecidedAt          *time.Time `json:"decidedAt"`
	IsCurrent          bool       `json:"isCurrent"`
	CreatedAt          time.Time  `json:"createdAt"`
}

type WorkflowResponse struct {
	ID                  uuid.UUID              `json:"id"`
	DocumentID          uuid.UUID              `json:"documentId"`
	DocumentName        string                 `json:"documentName"`
	CreatedByUserID     uuid.UUID              `json:"createdByUserId"`
	CreatedByFullName   string                 `json:"createdByFullName"`
	Title               string                 `json:"title"`
	Description         *string                `json:"description"`
	Status              string                 `json:"status"`
	CurrentStepOrder    int                    `json:"currentStepOrder"`
	Steps               []WorkflowStepResponse `json:"steps"`
	CreatedAt           time.Time              `json:"createdAt"`
	UpdatedAt           time.Time              `json:"updatedAt"`
	CompletedAt         *time.Time             `json:"completedAt"`
}

type PendingApprovalResponse struct {
	WorkflowID        uuid.UUID `json:"workflowId"`
	StepID            uuid.UUID `json:"stepId"`
	DocumentID        uuid.UUID `json:"documentId"`
	DocumentName      string    `json:"documentName"`
	WorkflowTitle     string    `json:"workflowTitle"`
	StepOrder         int       `json:"stepOrder"`
	CreatedByUserID   uuid.UUID `json:"createdByUserId"`
	CreatedByFullName string    `json:"createdByFullName"`
	CreatedAt         time.Time `json:"createdAt"`
}

type StepListRow struct {
	ApprovalStep
	ApproverFullName string `gorm:"column:approver_full_name"`
}

type WorkflowListRow struct {
	ApprovalWorkflow
	DocumentName      string `gorm:"column:document_name"`
	CreatedByFullName string `gorm:"column:created_by_full_name"`
}
