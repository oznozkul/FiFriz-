package listeners

import (
	"context"
	"fmt"

	"fifriz/internal/audit"
	"fifriz/internal/events"
	domain "fifriz/internal/events/events"
)

type AuditListener struct {
	audit audit.Recorder
}

func NewAuditListener(recorder audit.Recorder) *AuditListener {
	return &AuditListener{audit: recorder}
}

func (l *AuditListener) Register(bus events.EventBus) {
	if l == nil || bus == nil {
		return
	}
	for _, name := range allAuditEventNames() {
		bus.Subscribe(name, l.Handle)
	}
}

func (l *AuditListener) Handle(ctx context.Context, event events.Event) {
	if l == nil || event == nil {
		return
	}

	input := audit.CreateLogInput{
		TenantID: event.TenantID(),
		UserID:   event.ActorUserID(),
	}

	switch e := event.(type) {
	case domain.DocumentCreated:
		input.Action = audit.ActionDocumentUploaded
		input.EntityType = audit.EntityTypeDocument
		id := e.DocumentID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Belge yuklendi: %s", e.OriginalFileName)
	case domain.DocumentUpdated:
		input.Action = audit.ActionDocumentReprocessed
		input.EntityType = audit.EntityTypeDocument
		id := e.DocumentID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Belge guncellendi: %s", e.OriginalFileName)
	case domain.DocumentDeleted:
		if e.Permanent {
			input.Action = audit.ActionDocumentDeleted
		} else {
			input.Action = audit.ActionDocumentTrashed
		}
		input.EntityType = audit.EntityTypeDocument
		id := e.DocumentID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Belge silindi: %s", e.OriginalFileName)
	case domain.DocumentRestored:
		input.Action = audit.ActionDocumentRestored
		input.EntityType = audit.EntityTypeDocument
		id := e.DocumentID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Belge geri yuklendi: %s", e.OriginalFileName)
	case domain.DocumentMoved:
		input.Action = audit.ActionDocumentTagsUpdated
		input.EntityType = audit.EntityTypeDocument
		id := e.DocumentID
		input.EntityID = &id
		input.Description = "Belge tasindi"
	case domain.DocumentShared:
		input.Action = audit.ActionDocumentInternalShared
		input.EntityType = audit.EntityTypeDocumentShare
		id := e.DocumentID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Belge paylasildi: %s", e.OriginalFileName)
	case domain.DocumentVersionCreated:
		input.Action = audit.ActionDocumentVersionCreated
		input.EntityType = audit.EntityTypeDocumentVersion
		id := e.VersionID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Belge versiyonu olusturuldu: %s", e.OriginalFileName)
	case domain.DocumentVersionRestored:
		input.Action = audit.ActionDocumentVersionRestored
		input.EntityType = audit.EntityTypeDocumentVersion
		id := e.VersionID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Belge versiyonu geri yuklendi: %s", e.OriginalFileName)
	case domain.DocumentSearched:
		input.Action = audit.ActionDocumentSearched
		input.EntityType = audit.EntityTypeSearch
		input.Description = fmt.Sprintf("Belge aramasi yapildi: %s", e.Query)
		input.Metadata = map[string]interface{}{"query": e.Query, "resultCount": e.ResultCount}
	case domain.FavoriteAdded:
		input.Action = audit.ActionDocumentFavorited
		input.EntityType = audit.EntityTypeDocument
		id := e.DocumentID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Belge favorilere eklendi: %s", e.OriginalFileName)
	case domain.FavoriteRemoved:
		input.Action = audit.ActionDocumentUnfavorited
		input.EntityType = audit.EntityTypeDocument
		id := e.DocumentID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Belge favorilerden cikarildi: %s", e.OriginalFileName)
	case domain.FolderCreated:
		input.Action = audit.ActionFolderCreated
		input.EntityType = audit.EntityTypeFolder
		id := e.FolderID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Klasor olusturuldu: %s", e.Name)
	case domain.FolderUpdated:
		input.Action = audit.ActionFolderUpdated
		input.EntityType = audit.EntityTypeFolder
		id := e.FolderID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Klasor guncellendi: %s", e.Name)
	case domain.FolderRenamed:
		input.Action = audit.ActionFolderRenamed
		input.EntityType = audit.EntityTypeFolder
		id := e.FolderID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Klasor yeniden adlandirildi: %s", e.Name)
		input.Metadata = map[string]interface{}{"oldName": e.OldName, "newName": e.Name}
	case domain.FolderMoved:
		input.Action = audit.ActionFolderMoved
		input.EntityType = audit.EntityTypeFolder
		id := e.FolderID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Klasor tasindi: %s", e.Name)
		input.Metadata = map[string]interface{}{"oldParentId": e.OldParentID, "newParentId": e.NewParentID}
	case domain.FolderLocked:
		input.Action = audit.ActionFolderLocked
		input.EntityType = audit.EntityTypeFolder
		id := e.FolderID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Klasor kilitlendi: %s", e.Name)
	case domain.FolderUnlocked:
		input.Action = audit.ActionFolderUnlocked
		input.EntityType = audit.EntityTypeFolder
		id := e.FolderID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Klasor kilidi acildi: %s", e.Name)
	case domain.FolderPermissionChanged:
		input.Action = audit.ActionFolderPermissionChanged
		input.EntityType = audit.EntityTypeFolder
		id := e.FolderID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Klasor izinleri guncellendi: %s", e.Name)
		input.Metadata = map[string]interface{}{
			"inheritPermissions": e.InheritPermissions,
			"permissionCount":    e.PermissionCount,
		}
	case domain.FolderDeleted:
		input.Action = audit.ActionFolderDeleted
		input.EntityType = audit.EntityTypeFolder
		id := e.FolderID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Klasor silindi: %s", e.Name)
	case domain.TagCreated:
		input.Action = audit.ActionTagCreated
		input.EntityType = audit.EntityTypeTag
		id := e.TagID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Etiket olusturuldu: %s", e.Name)
	case domain.TagUpdated:
		input.Action = audit.ActionTagUpdated
		input.EntityType = audit.EntityTypeTag
		id := e.TagID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Etiket guncellendi: %s", e.Name)
	case domain.TagArchived:
		input.Action = audit.ActionTagArchived
		input.EntityType = audit.EntityTypeTag
		id := e.TagID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Etiket arsivlendi: %s", e.Name)
	case domain.TagMerged:
		input.Action = audit.ActionTagMerged
		input.EntityType = audit.EntityTypeTag
		id := e.SourceTagID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Etiket birlestirildi: %s -> %s", e.SourceName, e.TargetName)
		input.Metadata = map[string]interface{}{
			"sourceTagId": e.SourceTagID.String(),
			"targetTagId": e.TargetTagID.String(),
		}
	case domain.TagAliasCreated:
		input.Action = audit.ActionTagAliasCreated
		input.EntityType = audit.EntityTypeTag
		id := e.TagID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Etiket alias eklendi: %s -> %s", e.Alias, e.Name)
	case domain.TagAliasDeleted:
		input.Action = audit.ActionTagAliasDeleted
		input.EntityType = audit.EntityTypeTag
		id := e.TagID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Etiket alias silindi: %s -> %s", e.Alias, e.Name)
	case domain.TagDeleted:
		input.Action = audit.ActionTagDeleted
		input.EntityType = audit.EntityTypeTag
		id := e.TagID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Etiket silindi: %s", e.Name)
	case domain.TagAssigned:
		input.Action = audit.ActionDocumentTagsUpdated
		input.EntityType = audit.EntityTypeDocument
		id := e.DocumentID
		input.EntityID = &id
		input.Description = "Belge etiketleri guncellendi"
	case domain.CommentCreated:
		input.Action = audit.ActionDocumentCommentCreated
		input.EntityType = audit.EntityTypeDocumentComment
		id := e.CommentID
		input.EntityID = &id
		input.Description = "Belge yorumu eklendi"
	case domain.CommentDeleted:
		input.Action = audit.ActionDocumentCommentDeleted
		input.EntityType = audit.EntityTypeDocumentComment
		id := e.CommentID
		input.EntityID = &id
		input.Description = "Belge yorumu silindi"
	case domain.WorkflowCreated:
		input.Action = audit.ActionWorkflowCreated
		input.EntityType = audit.EntityTypeWorkflow
		id := e.WorkflowID
		input.EntityID = &id
		input.Description = "Onay sureci olusturuldu"
	case domain.WorkflowApproved:
		input.Action = audit.ActionWorkflowStepApproved
		input.EntityType = audit.EntityTypeWorkflow
		id := e.WorkflowID
		input.EntityID = &id
		input.Description = "Onay adimi onaylandi"
	case domain.WorkflowRejected:
		input.Action = audit.ActionWorkflowStepRejected
		input.EntityType = audit.EntityTypeWorkflow
		id := e.WorkflowID
		input.EntityID = &id
		input.Description = "Onay adimi reddedildi"
	case domain.IntegrationSyncCompleted:
		input.Action = audit.ActionIntegrationSyncCompleted
		input.EntityType = audit.EntityTypeIntegration
		id := e.IntegrationAccountID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Entegrasyon senkronu tamamlandi: %s", e.Provider)
	case domain.UserRegistered:
		input.Action = audit.ActionUserRegistered
		input.EntityType = audit.EntityTypeUser
		id := e.UserID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Kullanici kaydi olusturuldu: %s", e.Email)
	case domain.UserLoggedIn:
		input.Action = audit.ActionUserLoggedIn
		input.EntityType = audit.EntityTypeAuth
		id := e.UserID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Kullanici giris yapti: %s", e.Email)
	case domain.PasswordResetRequested:
		input.Action = audit.ActionPasswordResetRequested
		input.EntityType = audit.EntityTypeAuth
		id := e.UserID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Sifre sifirlama talebi: %s", e.Email)
	case domain.PasswordResetCompleted:
		input.Action = audit.ActionPasswordResetCompleted
		input.EntityType = audit.EntityTypeAuth
		id := e.UserID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Sifre sifirlama tamamlandi: %s", e.Email)
	case domain.ProfileUpdated:
		input.Action = audit.ActionProfileUpdated
		input.EntityType = audit.EntityTypeUser
		id := e.UserID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Profil guncellendi: %s", e.FullName)
	case domain.PasswordChanged:
		input.Action = audit.ActionPasswordChanged
		input.EntityType = audit.EntityTypeAuth
		id := e.UserID
		input.EntityID = &id
		input.Description = fmt.Sprintf("Sifre degistirildi: %s", e.Email)
	default:
		return
	}

	audit.Log(ctx, l.audit, input)
}

func allAuditEventNames() []string {
	return []string{
		domain.DocumentCreatedName,
		domain.DocumentUpdatedName,
		domain.DocumentDeletedName,
		domain.DocumentRestoredName,
		domain.DocumentMovedName,
		domain.DocumentSharedName,
		domain.DocumentVersionCreatedName,
		domain.DocumentVersionRestoredName,
		domain.DocumentSearchedName,
		domain.FavoriteAddedName,
		domain.FavoriteRemovedName,
		domain.FolderCreatedName,
		domain.FolderUpdatedName,
		domain.FolderRenamedName,
		domain.FolderMovedName,
		domain.FolderLockedName,
		domain.FolderUnlockedName,
		domain.FolderPermissionChangedName,
		domain.FolderDeletedName,
		domain.TagCreatedName,
		domain.TagUpdatedName,
		domain.TagArchivedName,
		domain.TagMergedName,
		domain.TagAliasCreatedName,
		domain.TagAliasDeletedName,
		domain.TagDeletedName,
		domain.TagAssignedName,
		domain.CommentCreatedName,
		domain.CommentDeletedName,
		domain.WorkflowCreatedName,
		domain.WorkflowApprovedName,
		domain.WorkflowRejectedName,
		domain.IntegrationSyncCompletedName,
		domain.UserRegisteredName,
		domain.UserLoggedInName,
		domain.PasswordResetRequestedName,
		domain.PasswordResetCompletedName,
		domain.ProfileUpdatedName,
		domain.PasswordChangedName,
	}
}
