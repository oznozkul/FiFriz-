package approvals

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&ApprovalWorkflow{}, &ApprovalStep{}); err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_approval_workflows_one_pending_per_document
		ON approval_workflows (tenant_id, document_id)
		WHERE status = 'pending'
	`).Error; err != nil {
		return err
	}

	return db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_approval_steps_workflow_order
		ON approval_steps (workflow_id, step_order)
	`).Error
}
