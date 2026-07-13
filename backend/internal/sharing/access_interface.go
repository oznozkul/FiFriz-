package sharing

import (
	"context"

	"fifriz/internal/auth"
	"fifriz/internal/documents"
)

type DocumentShareAccess interface {
	CanView(ctx context.Context, authContext auth.Context, document *documents.Document) (bool, error)
	CanDownload(ctx context.Context, authContext auth.Context, document *documents.Document) (bool, error)
	CanEdit(ctx context.Context, authContext auth.Context, document *documents.Document) (bool, error)
	CanManageInternalShares(ctx context.Context, authContext auth.Context, document *documents.Document) (bool, error)
	CanCreateInternalShare(ctx context.Context, authContext auth.Context, document *documents.Document) (bool, error)
	IsSharedAccess(ctx context.Context, authContext auth.Context, document *documents.Document) (bool, error)
}
