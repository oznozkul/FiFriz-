package pipeline

import (
	"context"

	"fifriz/internal/jobs"
)

// QueueEnqueuer schedules pipeline steps on the shared job queue.
type QueueEnqueuer struct {
	Queue jobs.Enqueuer
}

func NewQueueEnqueuer(queue jobs.Enqueuer) QueueEnqueuer {
	return QueueEnqueuer{Queue: queue}
}

func (q QueueEnqueuer) EnqueuePipelineStep(ctx context.Context, job *PipelineJob, stepName string) error {
	if job == nil {
		return ValidateJob(job)
	}
	return jobs.EnqueuePipelineStep(ctx, q.Queue, job.TenantID, job.DocumentID, job.UserID, stepName, job.Trigger)
}
