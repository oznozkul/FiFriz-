package worker

import (
	"context"
	"errors"
	"log"
	"time"

	"fifriz/internal/jobs"

	"github.com/redis/go-redis/v9"
)

type DocumentProcessor interface {
	ProcessDocumentJob(ctx context.Context, payload jobs.ProcessDocumentPayload) error
}

type DocumentCompressor interface {
	ProcessCompressDocumentJob(ctx context.Context, payload jobs.CompressDocumentPayload) error
}

type IntegrationSyncProcessor interface {
	ProcessSyncIntegrationJob(ctx context.Context, payload jobs.SyncIntegrationPayload) error
}

type Worker struct {
	queue                jobs.Dequeuer
	documentProcessor    DocumentProcessor
	documentCompressor   DocumentCompressor
	integrationProcessor IntegrationSyncProcessor
	pipelineProcessor    PipelineStepProcessor
}

type PipelineStepProcessor interface {
	ProcessPipelineStepJob(ctx context.Context, payload jobs.PipelineStepPayload) error
}

func New(queue jobs.Dequeuer, documentProcessor DocumentProcessor, integrationProcessor IntegrationSyncProcessor, documentCompressor DocumentCompressor, pipelineProcessor PipelineStepProcessor) *Worker {
	return &Worker{
		queue:                queue,
		documentProcessor:    documentProcessor,
		documentCompressor:   documentCompressor,
		integrationProcessor: integrationProcessor,
		pipelineProcessor:    pipelineProcessor,
	}
}

func (w *Worker) Run(ctx context.Context) {
	if w == nil || w.queue == nil {
		log.Printf("worker disabled: missing queue")
		return
	}

	log.Printf("background worker started")
	for {
		if ctx.Err() != nil {
			return
		}

		job, err := w.queue.Dequeue(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			if errors.Is(err, redis.Nil) {
				continue
			}
			log.Printf("worker dequeue error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		switch job.Type {
		case jobs.TypeProcessDocument:
			if w.documentProcessor == nil {
				log.Printf("worker skipped process_document: missing document processor")
				continue
			}
			payload := job.ProcessDocumentPayload()
			if err := w.documentProcessor.ProcessDocumentJob(ctx, payload); err != nil {
				log.Printf("worker process_document failed: document_id=%s error=%v", payload.DocumentID, err)
			}
		case jobs.TypeCompressDocument:
			if w.documentCompressor == nil {
				log.Printf("worker skipped compress_document: missing document compressor")
				continue
			}
			payload := job.CompressDocumentPayload()
			if err := w.documentCompressor.ProcessCompressDocumentJob(ctx, payload); err != nil {
				log.Printf("worker compress_document failed: document_id=%s error=%v", payload.DocumentID, err)
			}
		case jobs.TypePipelineStep:
			if w.pipelineProcessor == nil {
				log.Printf("worker skipped pipeline_step: missing pipeline processor")
				continue
			}
			payload := job.PipelineStepPayload()
			if err := w.pipelineProcessor.ProcessPipelineStepJob(ctx, payload); err != nil {
				log.Printf("worker pipeline_step failed: document_id=%s step=%s error=%v", payload.DocumentID, payload.StepName, err)
			}
		case jobs.TypeSyncIntegration:
			if w.integrationProcessor == nil {
				log.Printf("worker skipped sync_integration: missing integration processor")
				continue
			}
			payload := job.SyncIntegrationPayload()
			if err := w.integrationProcessor.ProcessSyncIntegrationJob(ctx, payload); err != nil {
				log.Printf("worker sync_integration failed: integration_id=%s error=%v", payload.IntegrationAccountID, err)
			}
		default:
			log.Printf("worker skipped unknown job type: %s", job.Type)
		}
	}
}
