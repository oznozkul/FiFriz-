package pipeline

import (
	"context"
	"fmt"
	"time"

	"fifriz/internal/events"
	domain "fifriz/internal/events/events"

	"github.com/google/uuid"
)

// Publisher adapts the application event bus to pipeline events.
type Publisher struct {
	bus events.Publisher
}

func NewEventPublisher(bus events.Publisher) *Publisher {
	return &Publisher{bus: bus}
}

func (p *Publisher) PublishStepCompleted(ctx context.Context, job *PipelineJob, stepName string) {
	if p == nil || p.bus == nil || job == nil {
		return
	}
	switch stepName {
	case StepGenerateEmbedding:
		p.bus.Publish(ctx, domain.NewDocumentEmbeddingCompleted(job.TenantID, job.DocumentID, job.UserID))
	case StepGenerateSummary:
		p.bus.Publish(ctx, domain.NewDocumentSummaryCompleted(job.TenantID, job.DocumentID, job.UserID))
	case StepExtractMetadata:
		p.bus.Publish(ctx, domain.NewDocumentMetadataCompleted(job.TenantID, job.DocumentID, job.UserID))
	case StepGenerateTags:
		p.bus.Publish(ctx, domain.NewDocumentTagsGenerated(job.TenantID, job.DocumentID, job.UserID))
	case StepExtractEntities:
		p.bus.Publish(ctx, domain.NewDocumentEntitiesExtracted(job.TenantID, job.DocumentID, job.UserID))
	case StepClassifyDocument:
		p.bus.Publish(ctx, domain.NewDocumentClassificationCompleted(job.TenantID, job.DocumentID, job.UserID))
	}
}

func (p *Publisher) PublishPipelineCompleted(ctx context.Context, job *PipelineJob) {
	if p == nil || p.bus == nil || job == nil {
		return
	}
	p.bus.Publish(ctx, domain.NewDocumentPipelineCompleted(job.TenantID, job.DocumentID, job.UserID))
}

// NewJob builds a pipeline job from upload/processing metadata.
func NewJob(tenantID, documentID, userID uuid.UUID, trigger, uploadID string) *PipelineJob {
	return &PipelineJob{
		TenantID:   tenantID,
		DocumentID: documentID,
		UserID:     userID,
		UploadID:   uploadID,
		Trigger:    trigger,
		Metadata:   map[string]interface{}{},
		Context:    map[string]interface{}{},
	}
}

// ValidateJob ensures required identifiers are present.
func ValidateJob(job *PipelineJob) error {
	if job == nil {
		return fmt.Errorf("pipeline job is nil")
	}
	if job.TenantID == uuid.Nil {
		return fmt.Errorf("pipeline job tenant id is required")
	}
	if job.DocumentID == uuid.Nil {
		return fmt.Errorf("pipeline job document id is required")
	}
	return nil
}

func formatTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}
