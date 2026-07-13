package tools

import (
	"context"

	"fifriz/internal/approvals"
	"fifriz/internal/auth"
	"fifriz/internal/documents"
	"fifriz/internal/tags"

	"github.com/google/uuid"
)

type DocumentAccess interface {
	CanView(ctx context.Context, authContext auth.Context, document *documents.Document) (bool, error)
	CanEdit(ctx context.Context, authContext auth.Context, document *documents.Document) (bool, error)
}

type DocumentReader interface {
	GetByID(ctx context.Context, tenantID, documentID uuid.UUID) (*documents.Document, error)
}

type Dependencies struct {
	Documents  *documents.Service
	DocumentReader DocumentReader
	Tags       *tags.Service
	Approvals  *approvals.Service
	Access     DocumentAccess
}
