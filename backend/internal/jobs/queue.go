package jobs

import (
	"context"

	"github.com/google/uuid"
)

type Enqueuer interface {
	Enqueue(ctx context.Context, job Job) error
}

type Dequeuer interface {
	Dequeue(ctx context.Context) (Job, error)
}

type Queue interface {
	Enqueuer
	Dequeuer
}

func EnqueueProcessDocument(ctx context.Context, enqueuer Enqueuer, tenantID, documentID, triggeredByUserID uuid.UUID, trigger string) error {
	if enqueuer == nil {
		return nil
	}
	return enqueuer.Enqueue(ctx, NewProcessDocumentJob(tenantID, documentID, triggeredByUserID, trigger))
}

func EnqueueSyncIntegration(ctx context.Context, enqueuer Enqueuer, tenantID, integrationAccountID, userID uuid.UUID, syncType string) error {
	if enqueuer == nil {
		return nil
	}
	return enqueuer.Enqueue(ctx, NewSyncIntegrationJob(tenantID, integrationAccountID, userID, syncType))
}

func EnqueueCompressDocument(ctx context.Context, enqueuer Enqueuer, tenantID, documentID uuid.UUID) error {
	if enqueuer == nil {
		return nil
	}
	return enqueuer.Enqueue(ctx, NewCompressDocumentJob(tenantID, documentID))
}
