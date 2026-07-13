package ai

import (
	"context"
	"time"

	"fifriz/internal/documents"
	"fifriz/internal/smartcollections"

	"github.com/google/uuid"
)

type SmartCollectionAdapter struct {
	repository *smartcollections.Repository
}

func NewSmartCollectionAdapter(repository *smartcollections.Repository) *SmartCollectionAdapter {
	return &SmartCollectionAdapter{repository: repository}
}

func (a *SmartCollectionAdapter) ListDocuments(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	accessUserID *uuid.UUID,
	key string,
	now time.Time,
) ([]documents.Document, error) {
	return a.repository.List(ctx, tenantID, userID, accessUserID, key, now)
}
