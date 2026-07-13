package documents

import (
	"context"

	"fifriz/internal/audit"
	"fifriz/internal/auth"
)

func (s *Service) logDocumentFileAccess(ctx context.Context, authContext auth.Context, document *Document, accessMode string) {
	if s.auditRecorder == nil || document == nil {
		return
	}

	action := audit.ActionDocumentPreviewed
	description := "Belge önizlendi"
	if accessMode == "download" {
		action = audit.ActionDocumentDownloaded
		description = "Belge indirildi"
	}

	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, authContext.TenantID, &userID, action, document.ID, description, map[string]interface{}{
		"accessMode": accessMode,
		"mimeType":   document.MimeType,
	})
}
