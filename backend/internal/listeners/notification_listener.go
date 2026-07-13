package listeners

import (
	"context"
	"fmt"

	"fifriz/internal/events"
	domain "fifriz/internal/events/events"
	"fifriz/internal/notifications"

	"github.com/google/uuid"
)

type NotificationListener struct {
	notifications notifications.Creator
}

func NewNotificationListener(creator notifications.Creator) *NotificationListener {
	return &NotificationListener{notifications: creator}
}

func (l *NotificationListener) Register(bus events.EventBus) {
	if l == nil || bus == nil {
		return
	}
	bus.Subscribe(domain.CommentCreatedName, l.Handle)
	bus.Subscribe(domain.DocumentVersionCreatedName, l.Handle)
	bus.Subscribe(domain.DocumentVersionRestoredName, l.Handle)
	bus.Subscribe(domain.WorkflowCreatedName, l.Handle)
	bus.Subscribe(domain.WorkflowApprovedName, l.Handle)
	bus.Subscribe(domain.WorkflowRejectedName, l.Handle)
	bus.Subscribe(domain.DocumentSharedName, l.Handle)
	bus.Subscribe(domain.DocumentProcessedName, l.Handle)
	bus.Subscribe(domain.DocumentProcessingFailedName, l.Handle)
	bus.Subscribe(domain.IntegrationSyncCompletedName, l.Handle)
	bus.Subscribe(domain.IntegrationSyncFailedName, l.Handle)
	bus.Subscribe(domain.PasswordChangedName, l.Handle)
	bus.Subscribe(domain.PasswordResetCompletedName, l.Handle)
}

func (l *NotificationListener) Handle(ctx context.Context, event events.Event) {
	if l == nil || l.notifications == nil || event == nil {
		return
	}

	switch e := event.(type) {
	case domain.CommentCreated:
		entityType := notifications.EntityTypeDocument
		entityID := e.DocumentID
		metadata := map[string]interface{}{
			"documentId": e.DocumentID.String(),
			"commentId":  e.CommentID.String(),
		}
		if e.ParentID != nil {
			metadata["parentId"] = e.ParentID.String()
			if e.ParentAuthorID != nil {
				notifications.Notify(l.notifications, ctx, notifications.CreateInput{
					TenantID:    e.TenantID(),
					UserID:      *e.ParentAuthorID,
					ActorUserID: *e.ActorUserID(),
					Type:        notifications.TypeDocumentCommentReplied,
					Title:       "Yorumunuza cevap verildi",
					Message:     fmt.Sprintf("%s belgesindeki yorumunuza yeni bir cevap eklendi.", e.OriginalFileName),
					EntityType:  &entityType,
					EntityID:    &entityID,
					Metadata:    metadata,
				})
			}
			return
		}
		notifications.Notify(l.notifications, ctx, notifications.CreateInput{
			TenantID:    e.TenantID(),
			UserID:      e.DocumentOwnerID,
			ActorUserID: *e.ActorUserID(),
			Type:        notifications.TypeDocumentCommentAdded,
			Title:       "Belgeye yeni yorum",
			Message:     fmt.Sprintf("%s belgesine yeni bir yorum eklendi.", e.OriginalFileName),
			EntityType:  &entityType,
			EntityID:    &entityID,
			Metadata:    metadata,
		})
	case domain.DocumentVersionCreated:
		entityType := notifications.EntityTypeDocument
		entityID := e.DocumentID
		notifications.Notify(l.notifications, ctx, notifications.CreateInput{
			TenantID:    e.TenantID(),
			UserID:      e.DocumentOwnerID,
			ActorUserID: *e.ActorUserID(),
			Type:        notifications.TypeDocumentVersionUploaded,
			Title:       "Yeni belge versiyonu",
			Message:     fmt.Sprintf("%s belgesine yeni bir versiyon (v%d) yuklendi.", e.OriginalFileName, e.VersionNumber),
			EntityType:  &entityType,
			EntityID:    &entityID,
			Metadata: map[string]interface{}{
				"documentId":    e.DocumentID.String(),
				"versionId":     e.VersionID.String(),
				"versionNumber": e.VersionNumber,
			},
		})
	case domain.DocumentVersionRestored:
		entityType := notifications.EntityTypeDocument
		entityID := e.DocumentID
		notifications.Notify(l.notifications, ctx, notifications.CreateInput{
			TenantID:    e.TenantID(),
			UserID:      e.DocumentOwnerID,
			ActorUserID: *e.ActorUserID(),
			Type:        notifications.TypeDocumentVersionRestored,
			Title:       "Belge versiyonu geri yuklendi",
			Message:     fmt.Sprintf("%s belgesinde v%d versiyonu aktif hale getirildi.", e.OriginalFileName, e.VersionNumber),
			EntityType:  &entityType,
			EntityID:    &entityID,
			Metadata: map[string]interface{}{
				"documentId":    e.DocumentID.String(),
				"versionId":     e.VersionID.String(),
				"versionNumber": e.VersionNumber,
			},
		})
	case domain.WorkflowCreated:
		if e.OwnerUserID == *e.ActorUserID() {
			return
		}
		entityType := notifications.EntityTypeWorkflow
		entityID := e.WorkflowID
		notifications.Notify(l.notifications, ctx, notifications.CreateInput{
			TenantID:    e.TenantID(),
			UserID:      e.OwnerUserID,
			ActorUserID: *e.ActorUserID(),
			Type:        notifications.TypeWorkflowCreated,
			Title:       "Onay sureci baslatildi",
			Message:     fmt.Sprintf("%s belgesi icin onay sureci olusturuldu: %s", e.OriginalFileName, e.WorkflowTitle),
			EntityType:  &entityType,
			EntityID:    &entityID,
			Metadata: map[string]interface{}{
				"workflowId": e.WorkflowID.String(),
				"documentId": e.DocumentID.String(),
			},
		})
	case domain.WorkflowApproved:
		entityType := notifications.EntityTypeWorkflow
		entityID := e.WorkflowID
		notifications.Notify(l.notifications, ctx, notifications.CreateInput{
			TenantID:   e.TenantID(),
			UserID:     e.RecipientUserID,
			Type:       notifications.TypeWorkflowApproved,
			Title:      "Onay sureci onaylandi",
			Message:    fmt.Sprintf("%s belgesi icin onay sureci onaylandi: %s", e.OriginalFileName, e.WorkflowTitle),
			EntityType: &entityType,
			EntityID:   &entityID,
			Metadata: map[string]interface{}{
				"workflowId": e.WorkflowID.String(),
				"documentId": e.DocumentID.String(),
			},
		})
	case domain.WorkflowRejected:
		entityType := notifications.EntityTypeWorkflow
		entityID := e.WorkflowID
		notifications.Notify(l.notifications, ctx, notifications.CreateInput{
			TenantID:   e.TenantID(),
			UserID:     e.RecipientUserID,
			Type:       notifications.TypeWorkflowRejected,
			Title:      "Onay sureci reddedildi",
			Message:    fmt.Sprintf("%s belgesi icin onay sureci reddedildi: %s", e.OriginalFileName, e.WorkflowTitle),
			EntityType: &entityType,
			EntityID:   &entityID,
			Metadata: map[string]interface{}{
				"workflowId": e.WorkflowID.String(),
				"documentId": e.DocumentID.String(),
			},
		})
	case domain.DocumentShared:
		entityType := notifications.EntityTypeDocument
		entityID := e.DocumentID
		actorID := uuid.Nil
		if e.ActorUserID() != nil {
			actorID = *e.ActorUserID()
		}
		notifications.Notify(l.notifications, ctx, notifications.CreateInput{
			TenantID:    e.TenantID(),
			UserID:      e.SharedWithUserID,
			ActorUserID: actorID,
			Type:        notifications.TypeDocumentSharedWithYou,
			Title:       "Belge sizinle paylasildi",
			Message:     fmt.Sprintf("%s belgesi sizinle paylasildi.", e.OriginalFileName),
			EntityType:  &entityType,
			EntityID:    &entityID,
			Metadata: map[string]interface{}{
				"documentId": e.DocumentID.String(),
			},
		})
	case domain.DocumentProcessed:
		entityType := notifications.EntityTypeDocument
		entityID := e.DocumentID
		notifications.Notify(l.notifications, ctx, notifications.CreateInput{
			TenantID:   e.TenantID(),
			UserID:     e.RecipientUserID,
			Type:       notifications.TypeDocumentProcessed,
			Title:      "Belge isleme tamamlandi",
			Message:    fmt.Sprintf("%s belgesi islendi ve arama icin hazir.", e.OriginalFileName),
			EntityType: &entityType,
			EntityID:   &entityID,
			Metadata: map[string]interface{}{
				"documentId": e.DocumentID.String(),
				"trigger":    e.Trigger,
			},
		})
	case domain.DocumentProcessingFailed:
		entityType := notifications.EntityTypeDocument
		entityID := e.DocumentID
		notifications.Notify(l.notifications, ctx, notifications.CreateInput{
			TenantID:   e.TenantID(),
			UserID:     e.RecipientUserID,
			Type:       notifications.TypeDocumentProcessingFailed,
			Title:      "Belge isleme basarisiz",
			Message:    fmt.Sprintf("%s belgesi islenirken bir hata olustu.", e.OriginalFileName),
			EntityType: &entityType,
			EntityID:   &entityID,
			Metadata: map[string]interface{}{
				"documentId": e.DocumentID.String(),
				"trigger":    e.Trigger,
			},
		})
	case domain.IntegrationSyncCompleted:
		entityType := notifications.EntityTypeIntegration
		entityID := e.IntegrationAccountID
		message := fmt.Sprintf("%s entegrasyonu basariyla senkronize edildi.", e.DisplayName)
		if e.ImportedCount > 0 {
			message = fmt.Sprintf("%s senkronizasyonu tamamlandi. %d oge ice aktarildi.", e.DisplayName, e.ImportedCount)
		}
		notifications.Notify(l.notifications, ctx, notifications.CreateInput{
			TenantID:   e.TenantID(),
			UserID:     e.TargetUserID,
			Type:       notifications.TypeIntegrationSyncCompleted,
			Title:      "Entegrasyon senkronizasyonu tamamlandi",
			Message:    message,
			EntityType: &entityType,
			EntityID:   &entityID,
			Metadata: map[string]interface{}{
				"provider": e.Provider,
			},
		})
	case domain.IntegrationSyncFailed:
		entityType := notifications.EntityTypeIntegration
		entityID := e.IntegrationAccountID
		notifications.Notify(l.notifications, ctx, notifications.CreateInput{
			TenantID:   e.TenantID(),
			UserID:     e.TargetUserID,
			Type:       notifications.TypeIntegrationSyncFailed,
			Title:      "Entegrasyon senkronizasyonu basarisiz",
			Message:    e.Message,
			EntityType: &entityType,
			EntityID:   &entityID,
			Metadata: map[string]interface{}{
				"provider": e.Provider,
				"error":    e.Message,
			},
		})
	case domain.PasswordChanged:
		notifications.Notify(l.notifications, ctx, notifications.CreateInput{
			TenantID:    e.TenantID(),
			UserID:      e.UserID,
			ActorUserID: *e.ActorUserID(),
			Type:        notifications.TypePasswordChanged,
			Title:       "Sifreniz degistirildi",
			Message:     "Hesap sifreniz basariyla guncellendi.",
		})
	case domain.PasswordResetCompleted:
		notifications.Notify(l.notifications, ctx, notifications.CreateInput{
			TenantID:    e.TenantID(),
			UserID:      e.UserID,
			ActorUserID: *e.ActorUserID(),
			Type:        notifications.TypePasswordChanged,
			Title:       "Sifreniz degistirildi",
			Message:     "Hesap sifreniz sifirlama ile guncellendi.",
		})
	}
}
