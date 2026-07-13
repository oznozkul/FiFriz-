package pipeline

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
)

// RetryPolicy controls per-step retries without failing the whole pipeline.
type RetryPolicy struct {
	MaxRetries int
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{MaxRetries: 2}
}

// Runner orchestrates pipeline execution via the registry and queue.
type Runner struct {
	registry *Registry
	tracker  JobTracker
	enqueuer StepEnqueuer
	events   EventPublisher
	retry    RetryPolicy
}

type RunnerConfig struct {
	Registry *Registry
	Tracker  JobTracker
	Enqueuer StepEnqueuer
	Events   EventPublisher
	Retry    RetryPolicy
}

func NewRunner(cfg RunnerConfig) *Runner {
	retry := cfg.Retry
	if retry.MaxRetries <= 0 {
		retry = DefaultRetryPolicy()
	}
	return &Runner{
		registry: cfg.Registry,
		tracker:  cfg.Tracker,
		enqueuer: cfg.Enqueuer,
		events:   cfg.Events,
		retry:    retry,
	}
}

// Start initializes step records and enqueues the first pipeline step.
func (r *Runner) Start(ctx context.Context, job *PipelineJob) error {
	if err := ValidateJob(job); err != nil {
		return err
	}
	if r.registry == nil || r.tracker == nil || r.enqueuer == nil {
		return fmt.Errorf("pipeline runner is not fully configured")
	}

	stepNames := r.registry.StepNames()
	if err := r.tracker.EnsureSteps(ctx, job.TenantID, job.DocumentID, stepNames); err != nil {
		return err
	}

	first := stepNames[0]
	return r.enqueuer.EnqueuePipelineStep(ctx, job, first)
}

// RunStep executes a single queued pipeline step and schedules the next one.
func (r *Runner) RunStep(ctx context.Context, job *PipelineJob, stepName string) error {
	if err := ValidateJob(job); err != nil {
		return err
	}

	step, ok := r.registry.Get(stepName)
	if !ok {
		return fmt.Errorf("unknown pipeline step: %s", stepName)
	}

	if !step.CanRun(ctx, job) {
		if err := r.tracker.MarkSkipped(ctx, job.TenantID, job.DocumentID, stepName); err != nil {
			log.Printf("pipeline skip mark failed: step=%s document_id=%s error=%v", stepName, job.DocumentID, err)
		}
		return r.scheduleNext(ctx, job, stepName)
	}

	if err := r.tracker.MarkRunning(ctx, job.TenantID, job.DocumentID, stepName); err != nil {
		return err
	}

	execErr := step.Execute(ctx, job)
	if execErr != nil {
		return r.handleStepFailure(ctx, job, stepName, execErr)
	}

	if err := r.tracker.MarkCompleted(ctx, job.TenantID, job.DocumentID, stepName); err != nil {
		return err
	}

	if r.events != nil {
		r.events.PublishStepCompleted(ctx, job, stepName)
	}

	if stepName == StepCompleted && r.events != nil {
		r.events.PublishPipelineCompleted(ctx, job)
	}

	return r.scheduleNext(ctx, job, stepName)
}

func (r *Runner) handleStepFailure(ctx context.Context, job *PipelineJob, stepName string, execErr error) error {
	retries, err := r.tracker.RetryCount(ctx, job.TenantID, job.DocumentID, stepName)
	if err != nil {
		retries = 0
	}

	if retries < r.retry.MaxRetries {
		if markErr := r.tracker.MarkRetrying(ctx, job.TenantID, job.DocumentID, stepName, execErr); markErr != nil {
			log.Printf("pipeline retry mark failed: step=%s error=%v", stepName, markErr)
		}
		log.Printf("pipeline step retry: step=%s document_id=%s attempt=%d error=%v", stepName, job.DocumentID, retries+1, execErr)
		return r.enqueuer.EnqueuePipelineStep(ctx, job, stepName)
	}

	if markErr := r.tracker.MarkFailed(ctx, job.TenantID, job.DocumentID, stepName, execErr); markErr != nil {
		log.Printf("pipeline failed mark error: step=%s error=%v", stepName, markErr)
	}
	log.Printf("pipeline step failed (continuing): step=%s document_id=%s error=%v", stepName, job.DocumentID, execErr)
	return r.scheduleNext(ctx, job, stepName)
}

func (r *Runner) scheduleNext(ctx context.Context, job *PipelineJob, currentStep string) error {
	nextStep, ok := r.registry.NextStep(currentStep)
	if !ok {
		return nil
	}
	return r.enqueuer.EnqueuePipelineStep(ctx, job, nextStep)
}

// GetProgress returns pipeline status for a document.
func (r *Runner) GetProgress(ctx context.Context, tenantID, documentID uuid.UUID) ([]StepProgress, error) {
	if r.tracker == nil {
		return nil, fmt.Errorf("pipeline tracker is not configured")
	}
	return r.tracker.GetProgress(ctx, tenantID, documentID)
}
