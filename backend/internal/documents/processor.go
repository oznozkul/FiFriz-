package documents

import (
	"context"
	"errors"
	"fmt"

	"fifriz/internal/applog"
	domain "fifriz/internal/events/events"
	"fifriz/internal/jobs"
	"fifriz/internal/ai/pipeline"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *Service) ProcessDocumentJob(ctx context.Context, payload jobs.ProcessDocumentPayload) error {
	if s.repository == nil {
		return fmt.Errorf("document repository is not configured")
	}

	document, err := s.repository.GetByID(ctx, payload.TenantID, payload.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("document not found: %s", payload.DocumentID)
		}
		return err
	}

	if s.pipelineRunner != nil {
		document.Status = DocumentStatusProcessing
		if err := s.repository.Update(ctx, document); err != nil {
			return err
		}
		applog.Info("document pipeline started: document_id=%s trigger=%s", document.ID, payload.Trigger)
		job := pipeline.NewJob(payload.TenantID, payload.DocumentID, payload.TriggeredByUserID, payload.Trigger, "")
		if err := s.pipelineRunner.Start(ctx, job); err != nil {
			document.Status = DocumentStatusFailed
			_ = s.repository.Update(ctx, document)
			recipientID := payload.TriggeredByUserID
			if recipientID == uuid.Nil {
				recipientID = document.UploadedByUserID
			}
			s.publisher.Publish(ctx, domain.NewDocumentProcessingFailed(payload.TenantID, document.ID, recipientID, document.OriginalFileName, payload.Trigger))
			return err
		}
		return nil
	}

	return s.processDocumentLegacy(ctx, payload, document)
}

func (s *Service) scheduleDocumentProcessing(ctx context.Context, tenantID, documentID, userID uuid.UUID, trigger string) {
	if err := jobs.EnqueueProcessDocument(ctx, s.jobQueue, tenantID, documentID, userID, trigger); err != nil {
		applog.Warn("document processing enqueue failed: document_id=%s trigger=%s error=%v", documentID, trigger, err)
	}
}

// ScheduleProcessing enqueues document processing for integrations and other callers.
func (s *Service) ScheduleProcessing(ctx context.Context, tenantID, documentID, userID uuid.UUID, trigger string) {
	s.scheduleDocumentProcessing(ctx, tenantID, documentID, userID, trigger)
}

// RecoverStuckProcessing marks documents stuck in processing as processed when text already exists.
func (s *Service) RecoverStuckProcessing(ctx context.Context) {
	if s.repository == nil {
		return
	}
	count, err := s.repository.RecoverStuckProcessingWithText(ctx)
	if err != nil {
		applog.Warn("recover stuck processing failed: error=%v", err)
		return
	}
	if count > 0 {
		applog.Info("recovered stuck processing documents: count=%d", count)
	}
}

func (s *Service) processDocumentLegacy(ctx context.Context, payload jobs.ProcessDocumentPayload, document *Document) error {
	document.Status = DocumentStatusProcessing
	if err := s.repository.Update(ctx, document); err != nil {
		return err
	}

	applog.Info("document processing started: document_id=%s trigger=%s", document.ID, payload.Trigger)

	var processErr error
	if isTextExtractionSupported(document.FileExtension) {
		processErr = s.extractAndSaveDocumentText(ctx, document)
	} else {
		_ = s.repository.DeleteTextByDocumentID(ctx, document.ID)
		_ = s.repository.DeleteChunksByDocumentID(ctx, document.ID)
	}

	if processErr != nil {
		document.Status = DocumentStatusFailed
		applog.Warn("document processing failed: document_id=%s error=%v", document.ID, processErr)
	} else {
		document.Status = DocumentStatusReady
		applog.Info("document processing finished: document_id=%s", document.ID)
	}

	if err := s.repository.Update(ctx, document); err != nil {
		return err
	}

	recipientID := payload.TriggeredByUserID
	if recipientID == uuid.Nil {
		recipientID = document.UploadedByUserID
	}
	if processErr != nil {
		s.publisher.Publish(ctx, domain.NewDocumentProcessingFailed(payload.TenantID, document.ID, recipientID, document.OriginalFileName, payload.Trigger))
	} else {
		s.publisher.Publish(ctx, domain.NewDocumentProcessed(payload.TenantID, document.ID, recipientID, document.OriginalFileName, payload.Trigger))
	}

	return processErr
}
