package sharing

import (
	"context"
	"fmt"

	"fifriz/internal/auth"
	"fifriz/internal/documents"
	"fifriz/internal/notifications"
)

func (s *Service) notifyDocumentSharedWithUser(ctx context.Context, authContext auth.Context, document *documents.Document, share *DocumentShare) {
	if s.notifications == nil {
		return
	}

	sharerName := "Bir kullanici"
	if user, err := s.authRepository.GetUserByID(ctx, authContext.UserID); err == nil && user.FullName != "" {
		sharerName = user.FullName
	}

	entityType := notifications.EntityTypeDocument
	entityID := document.ID
	notifications.Notify(s.notifications, ctx, notifications.CreateInput{
		TenantID:    authContext.TenantID,
		UserID:      share.SharedWithUserID,
		ActorUserID: authContext.UserID,
		Type:        notifications.TypeDocumentSharedWithYou,
		Title:       "Belge sizinle paylasildi",
		Message:     fmt.Sprintf("%s size '%s' belgesini paylasildi.", sharerName, document.OriginalFileName),
		EntityType:  &entityType,
		EntityID:    &entityID,
		Metadata: map[string]interface{}{
			"documentId":       document.ID.String(),
			"sharedWithUserId": share.SharedWithUserID.String(),
			"permission":       string(share.Permission),
		},
	})
}
