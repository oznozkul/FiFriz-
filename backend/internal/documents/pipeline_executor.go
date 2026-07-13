package documents

import (
	"context"
	"errors"
	"strings"

	"fifriz/internal/ai/pipeline"
	domain "fifriz/internal/events/events"
	"fifriz/internal/auth"
	"fifriz/internal/jobs"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PipelineExecutor implements pipeline.DocumentExecutor using the documents service.
type PipelineExecutor struct {
	service *Service
}

func NewPipelineExecutor(service *Service) *PipelineExecutor {
	return &PipelineExecutor{service: service}
}

func (e *PipelineExecutor) ExtractText(ctx context.Context, job *pipeline.PipelineJob) error {
	document, err := e.loadDocument(ctx, job)
	if err != nil {
		return err
	}

	document.Status = DocumentStatusProcessing
	if err := e.service.repository.Update(ctx, document); err != nil {
		return err
	}

	location, err := documentStorageLocation(e.service.storage, document)
	if err != nil {
		return err
	}

	extension := normalizeExtension("." + document.FileExtension)
	var fullText string
	err = e.service.storage.WithLocalFile(ctx, location, extension, func(filePath string) error {
		extracted, extractErr := extractDocumentText(document.FileExtension, filePath)
		if extractErr != nil {
			return extractErr
		}
		fullText = extracted
		return nil
	})
	if err != nil {
		return err
	}

	fullText = strings.TrimSpace(fullText)
	if fullText == "" {
		_ = e.service.repository.DeleteTextByDocumentID(ctx, document.ID)
		_ = e.service.repository.DeleteChunksByDocumentID(ctx, document.ID)
		return nil
	}

	documentText := &DocumentText{
		DocumentID: document.ID,
		FullText:   fullText,
	}
	if err := e.service.repository.UpsertText(ctx, documentText); err != nil {
		return err
	}
	_ = e.service.repository.DeleteChunksByDocumentID(ctx, document.ID)
	return nil
}

func (e *PipelineExecutor) GenerateChunks(ctx context.Context, job *pipeline.PipelineJob) error {
	documentText, err := e.service.repository.GetTextByDocumentID(ctx, job.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	fullText := strings.TrimSpace(documentText.FullText)
	if fullText == "" {
		return nil
	}

	chunks := buildDocumentChunks(job.DocumentID, fullText)
	return e.service.repository.ReplaceChunks(ctx, job.DocumentID, chunks)
}

func (e *PipelineExecutor) GenerateEmbeddings(ctx context.Context, job *pipeline.PipelineJob) error {
	chunks, err := e.service.repository.ListChunksByDocumentID(ctx, job.DocumentID)
	if err != nil {
		return err
	}
	if len(chunks) == 0 {
		return nil
	}

	chunks = e.service.embedChunks(ctx, chunks)
	return e.service.repository.ReplaceChunks(ctx, job.DocumentID, chunks)
}

func (e *PipelineExecutor) GenerateSummary(ctx context.Context, job *pipeline.PipelineJob) error {
	if e.service.summaryChatProvider == nil {
		return nil
	}
	_, err := e.service.GenerateSummaryForRequest(ctx, SummaryGenerateRequest{
		TenantID:   job.TenantID,
		DocumentID: job.DocumentID,
		UserID:     job.UserID,
	})
	if errors.Is(err, ErrSummaryProviderDisabled) || errors.Is(err, ErrSummaryNoContent) {
		return nil
	}
	return err
}

func (e *PipelineExecutor) ClassifyDocument(ctx context.Context, job *pipeline.PipelineJob) error {
	if e.service.summaryChatProvider == nil {
		return nil
	}
	_, err := e.service.GenerateClassificationForRequest(ctx, ClassificationGenerateRequest{
		TenantID:   job.TenantID,
		DocumentID: job.DocumentID,
		UserID:     job.UserID,
	})
	if errors.Is(err, ErrClassificationProviderDisabled) || errors.Is(err, ErrClassificationNoContent) || errors.Is(err, ErrClassificationParseFailed) {
		return nil
	}
	return err
}

func (e *PipelineExecutor) ExtractMetadata(ctx context.Context, job *pipeline.PipelineJob) error {
	if e.service.summaryChatProvider == nil {
		return nil
	}
	_, err := e.service.GenerateAIMetadataForRequest(ctx, AIMetadataGenerateRequest{
		TenantID:   job.TenantID,
		DocumentID: job.DocumentID,
		UserID:     job.UserID,
	})
	if errors.Is(err, ErrMetadataProviderDisabled) || errors.Is(err, ErrMetadataNoContent) || errors.Is(err, ErrMetadataParseFailed) {
		return nil
	}
	return err
}

func (e *PipelineExecutor) ExtractImportantDates(ctx context.Context, job *pipeline.PipelineJob) error {
	if e.service.summaryChatProvider == nil {
		return nil
	}
	_, err := e.service.GenerateAIImportantDatesForRequest(ctx, AIImportantDateGenerateRequest{
		TenantID:   job.TenantID,
		DocumentID: job.DocumentID,
		UserID:     job.UserID,
	})
	if errors.Is(err, ErrAIImportantDatesProviderDisabled) || errors.Is(err, ErrAIImportantDatesNoContent) || errors.Is(err, ErrAIImportantDatesParseFailed) {
		return nil
	}
	return err
}

func (e *PipelineExecutor) SuggestTags(ctx context.Context, job *pipeline.PipelineJob) error {
	if e.service.summaryChatProvider == nil {
		return nil
	}
	_, err := e.service.GenerateSuggestedTagsForRequest(ctx, SuggestedTagGenerateRequest{
		TenantID:   job.TenantID,
		DocumentID: job.DocumentID,
		UserID:     job.UserID,
	})
	if errors.Is(err, ErrSuggestedTagsProviderDisabled) || errors.Is(err, ErrSuggestedTagsNoContent) {
		return nil
	}
	return err
}

func (e *PipelineExecutor) ExtractEntities(ctx context.Context, job *pipeline.PipelineJob) error {
	if e.service.summaryChatProvider == nil {
		return nil
	}
	_, err := e.service.GenerateEntitiesForRequest(ctx, EntityGenerateRequest{
		TenantID:   job.TenantID,
		DocumentID: job.DocumentID,
		UserID:     job.UserID,
	})
	if errors.Is(err, ErrAIEntitiesProviderDisabled) || errors.Is(err, ErrAIEntitiesNoContent) || errors.Is(err, ErrAIEntitiesParseFailed) {
		return nil
	}
	return err
}

func (e *PipelineExecutor) LinkRelatedDocuments(ctx context.Context, job *pipeline.PipelineJob) error {
	accessUserID := job.UserID
	var accessFilter *uuid.UUID
	if accessUserID != uuid.Nil {
		accessFilter = &accessUserID
	}
	_, err := e.service.GenerateRelatedSuggestionsForRequest(ctx, RelatedSuggestionGenerateRequest{
		TenantID:         job.TenantID,
		SourceDocumentID: job.DocumentID,
		UserID:           job.UserID,
		AccessUserID:     accessFilter,
	})
	if errors.Is(err, ErrAIRelatedDocumentsNoEmbeddings) {
		return nil
	}
	return err
}

func (e *PipelineExecutor) FinalizePipeline(ctx context.Context, job *pipeline.PipelineJob) error {
	document, err := e.loadDocument(ctx, job)
	if err != nil {
		return err
	}

	document.Status = DocumentStatusReady
	if err := e.service.repository.Update(ctx, document); err != nil {
		return err
	}

	recipientID := job.UserID
	if recipientID == uuid.Nil {
		recipientID = document.UploadedByUserID
	}
	e.service.publisher.Publish(ctx, domain.NewDocumentProcessed(job.TenantID, document.ID, recipientID, document.OriginalFileName, job.Trigger))
	return nil
}

func (e *PipelineExecutor) IsTextExtractionSupported(documentID, tenantID uuid.UUID) (bool, error) {
	document, err := e.service.repository.GetByID(context.Background(), tenantID, documentID)
	if err != nil {
		return false, err
	}
	return isTextExtractionSupported(document.FileExtension), nil
}

func (e *PipelineExecutor) loadDocument(ctx context.Context, job *pipeline.PipelineJob) (*Document, error) {
	return e.service.repository.GetByID(ctx, job.TenantID, job.DocumentID)
}

// ProcessPipelineStepJob executes one queued pipeline step.
func (s *Service) ProcessPipelineStepJob(ctx context.Context, payload jobs.PipelineStepPayload) error {
	if s.pipelineRunner == nil {
		return nil
	}
	job := pipeline.NewJob(payload.TenantID, payload.DocumentID, payload.UserID, payload.Trigger, payload.UploadID)
	return s.pipelineRunner.RunStep(ctx, job, payload.StepName)
}

// GetAIProgress returns pipeline step progress for a document.
func (s *Service) GetAIProgress(ctx context.Context, documentID string) ([]pipeline.StepProgress, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if s.pipelineRunner == nil {
		return []pipeline.StepProgress{}, nil
	}

	parsedDocumentID, err := uuid.Parse(strings.TrimSpace(documentID))
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, parsedDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	return s.pipelineRunner.GetProgress(ctx, authContext.TenantID, parsedDocumentID)
}

func (s *Service) SetPipelineRunner(runner *pipeline.Runner) {
	s.pipelineRunner = runner
}
