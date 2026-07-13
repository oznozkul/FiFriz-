package pipeline

import (
	"context"

	"github.com/google/uuid"
)

// Step name constants — queue job types map 1:1 to these names.
const (
	StepExtractText          = "extract_text"
	StepChunkDocument        = "chunk_document"
	StepGenerateEmbedding    = "generate_embedding"
	StepGenerateSummary      = "generate_summary"
	StepClassifyDocument     = "classify_document"
	StepExtractMetadata      = "extract_metadata"
	StepExtractDates         = "extract_dates"
	StepGenerateTags         = "generate_tags"
	StepExtractEntities      = "extract_entities"
	StepLinkRelatedDocuments = "link_related_documents"
	StepCompleted            = "completed"
)

// DefaultStepOrder is the canonical pipeline sequence.
var DefaultStepOrder = []string{
	StepExtractText,
	StepChunkDocument,
	StepGenerateEmbedding,
	StepGenerateSummary,
	StepClassifyDocument,
	StepExtractMetadata,
	StepExtractDates,
	StepGenerateTags,
	StepExtractEntities,
	StepLinkRelatedDocuments,
	StepCompleted,
}

// Job status values stored in document_ai_jobs.
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusRetrying  = "retrying"
	StatusSkipped   = "skipped"
)

// PipelineJob carries tenant-scoped work for a single document through the pipeline.
type PipelineJob struct {
	TenantID   uuid.UUID
	DocumentID uuid.UUID
	UserID     uuid.UUID
	UploadID   string
	Trigger    string
	Metadata   map[string]interface{}
	Context    map[string]interface{}
}

// PipelineStep is implemented by every AI processing stage.
type PipelineStep interface {
	Name() string
	Execute(ctx context.Context, job *PipelineJob) error
	CanRun(ctx context.Context, job *PipelineJob) bool
}

// DocumentExecutor performs document-scoped work for pipeline steps.
// Implemented by documents.Service to avoid import cycles.
type DocumentExecutor interface {
	ExtractText(ctx context.Context, job *PipelineJob) error
	GenerateChunks(ctx context.Context, job *PipelineJob) error
	GenerateEmbeddings(ctx context.Context, job *PipelineJob) error
	GenerateSummary(ctx context.Context, job *PipelineJob) error
	ClassifyDocument(ctx context.Context, job *PipelineJob) error
	ExtractMetadata(ctx context.Context, job *PipelineJob) error
	ExtractImportantDates(ctx context.Context, job *PipelineJob) error
	SuggestTags(ctx context.Context, job *PipelineJob) error
	ExtractEntities(ctx context.Context, job *PipelineJob) error
	LinkRelatedDocuments(ctx context.Context, job *PipelineJob) error
	FinalizePipeline(ctx context.Context, job *PipelineJob) error
	IsTextExtractionSupported(documentID uuid.UUID, tenantID uuid.UUID) (bool, error)
}

// JobTracker persists per-step status for dashboard and retry logic.
type JobTracker interface {
	EnsureSteps(ctx context.Context, tenantID, documentID uuid.UUID, stepNames []string) error
	MarkRunning(ctx context.Context, tenantID, documentID uuid.UUID, stepName string) error
	MarkCompleted(ctx context.Context, tenantID, documentID uuid.UUID, stepName string) error
	MarkFailed(ctx context.Context, tenantID, documentID uuid.UUID, stepName string, err error) error
	MarkRetrying(ctx context.Context, tenantID, documentID uuid.UUID, stepName string, err error) error
	MarkSkipped(ctx context.Context, tenantID, documentID uuid.UUID, stepName string) error
	GetProgress(ctx context.Context, tenantID, documentID uuid.UUID) ([]StepProgress, error)
	RetryCount(ctx context.Context, tenantID, documentID uuid.UUID, stepName string) (int, error)
}

// StepEnqueuer schedules individual pipeline step jobs.
type StepEnqueuer interface {
	EnqueuePipelineStep(ctx context.Context, job *PipelineJob, stepName string) error
}

// StepProgress is returned by the progress API.
type StepProgress struct {
	StepName    string  `json:"stepName"`
	Status      string  `json:"status"`
	StartedAt   *string `json:"startedAt,omitempty"`
	CompletedAt *string `json:"completedAt,omitempty"`
	Error       *string `json:"error,omitempty"`
	RetryCount  int     `json:"retryCount"`
}

// EventPublisher publishes pipeline domain events.
type EventPublisher interface {
	PublishStepCompleted(ctx context.Context, job *PipelineJob, stepName string)
	PublishPipelineCompleted(ctx context.Context, job *PipelineJob)
}
