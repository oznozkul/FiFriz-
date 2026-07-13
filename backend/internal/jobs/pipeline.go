package jobs

import (
	"context"

	"github.com/google/uuid"
)

type PipelineStepPayload struct {
	TenantID   uuid.UUID
	DocumentID uuid.UUID
	UserID     uuid.UUID
	StepName   string
	Trigger    string
	UploadID   string
}

func (j Job) PipelineStepPayload() PipelineStepPayload {
	return PipelineStepPayload{
		TenantID:   j.TenantID,
		DocumentID: j.DocumentID,
		UserID:     j.TriggeredByUserID,
		StepName:   j.PipelineStepName,
		Trigger:    j.Trigger,
	}
}

func NewPipelineStepJob(tenantID, documentID, userID uuid.UUID, stepName, trigger string) Job {
	return Job{
		Type:              TypePipelineStep,
		TenantID:          tenantID,
		DocumentID:        documentID,
		TriggeredByUserID: userID,
		Trigger:           trigger,
		PipelineStepName:  stepName,
	}
}

func EnqueuePipelineStep(ctx context.Context, enqueuer Enqueuer, tenantID, documentID, userID uuid.UUID, stepName, trigger string) error {
	if enqueuer == nil {
		return nil
	}
	return enqueuer.Enqueue(ctx, NewPipelineStepJob(tenantID, documentID, userID, stepName, trigger))
}
