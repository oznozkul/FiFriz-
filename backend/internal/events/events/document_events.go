package domain

import "github.com/google/uuid"

const (
	DocumentCreatedName         = "document.created"
	DocumentUpdatedName         = "document.updated"
	DocumentDeletedName         = "document.deleted"
	DocumentRestoredName        = "document.restored"
	DocumentMovedName           = "document.moved"
	DocumentSharedName          = "document.shared"
	DocumentVersionCreatedName  = "document.version.created"
	DocumentVersionRestoredName = "document.version.restored"
	DocumentSearchedName        = "document.searched"
	FavoriteAddedName           = "document.favorite.added"
	FavoriteRemovedName           = "document.favorite.removed"
	DocumentProcessedName         = "document.processed"
	DocumentProcessingFailedName  = "document.processing.failed"
)

type DocumentCreated struct {
	Base
	DocumentID       uuid.UUID
	OriginalFileName string
	FileExtension    string
	FolderID         *uuid.UUID
}

func NewDocumentCreated(tenantID, actorID, documentID uuid.UUID, fileName, ext string, folderID *uuid.UUID) DocumentCreated {
	base := NewBase(DocumentCreatedName, tenantID, &actorID)
	return DocumentCreated{
		Base:             base,
		DocumentID:       documentID,
		OriginalFileName: fileName,
		FileExtension:    ext,
		FolderID:         folderID,
	}
}

type DocumentUpdated struct {
	Base
	DocumentID       uuid.UUID
	OriginalFileName string
}

func NewDocumentUpdated(tenantID, actorID, documentID uuid.UUID, fileName string) DocumentUpdated {
	base := NewBase(DocumentUpdatedName, tenantID, &actorID)
	return DocumentUpdated{Base: base, DocumentID: documentID, OriginalFileName: fileName}
}

type DocumentDeleted struct {
	Base
	DocumentID       uuid.UUID
	OriginalFileName string
	Permanent        bool
}

func NewDocumentDeleted(tenantID, actorID, documentID uuid.UUID, fileName string, permanent bool) DocumentDeleted {
	base := NewBase(DocumentDeletedName, tenantID, &actorID)
	return DocumentDeleted{Base: base, DocumentID: documentID, OriginalFileName: fileName, Permanent: permanent}
}

type DocumentRestored struct {
	Base
	DocumentID       uuid.UUID
	OriginalFileName string
}

func NewDocumentRestored(tenantID, actorID, documentID uuid.UUID, fileName string) DocumentRestored {
	base := NewBase(DocumentRestoredName, tenantID, &actorID)
	return DocumentRestored{Base: base, DocumentID: documentID, OriginalFileName: fileName}
}

type DocumentMoved struct {
	Base
	DocumentID uuid.UUID
	FolderID   *uuid.UUID
}

func NewDocumentMoved(tenantID, actorID, documentID uuid.UUID, folderID *uuid.UUID) DocumentMoved {
	base := NewBase(DocumentMovedName, tenantID, &actorID)
	return DocumentMoved{Base: base, DocumentID: documentID, FolderID: folderID}
}

type DocumentShared struct {
	Base
	DocumentID       uuid.UUID
	OriginalFileName string
	SharedWithUserID uuid.UUID
}

func NewDocumentShared(tenantID, actorID, documentID, sharedWith uuid.UUID, fileName string) DocumentShared {
	base := NewBase(DocumentSharedName, tenantID, &actorID)
	return DocumentShared{Base: base, DocumentID: documentID, OriginalFileName: fileName, SharedWithUserID: sharedWith}
}

type DocumentVersionCreated struct {
	Base
	DocumentID       uuid.UUID
	VersionID        uuid.UUID
	VersionNumber    int
	OriginalFileName string
	DocumentOwnerID  uuid.UUID
}

func NewDocumentVersionCreated(tenantID, actorID, documentID, versionID, ownerID uuid.UUID, versionNumber int, fileName string) DocumentVersionCreated {
	base := NewBase(DocumentVersionCreatedName, tenantID, &actorID)
	return DocumentVersionCreated{
		Base:             base,
		DocumentID:       documentID,
		VersionID:        versionID,
		VersionNumber:    versionNumber,
		OriginalFileName: fileName,
		DocumentOwnerID:  ownerID,
	}
}

type DocumentVersionRestored struct {
	Base
	DocumentID       uuid.UUID
	VersionID        uuid.UUID
	VersionNumber    int
	OriginalFileName string
	DocumentOwnerID  uuid.UUID
}

func NewDocumentVersionRestored(tenantID, actorID, documentID, versionID, ownerID uuid.UUID, versionNumber int, fileName string) DocumentVersionRestored {
	base := NewBase(DocumentVersionRestoredName, tenantID, &actorID)
	return DocumentVersionRestored{
		Base:             base,
		DocumentID:       documentID,
		VersionID:        versionID,
		VersionNumber:    versionNumber,
		OriginalFileName: fileName,
		DocumentOwnerID:  ownerID,
	}
}

type DocumentSearched struct {
	Base
	Query       string
	ResultCount int
}

func NewDocumentSearched(tenantID, actorID uuid.UUID, query string, resultCount int) DocumentSearched {
	base := NewBase(DocumentSearchedName, tenantID, &actorID)
	return DocumentSearched{Base: base, Query: query, ResultCount: resultCount}
}

type FavoriteAdded struct {
	Base
	DocumentID       uuid.UUID
	OriginalFileName string
}

func NewFavoriteAdded(tenantID, actorID, documentID uuid.UUID, fileName string) FavoriteAdded {
	base := NewBase(FavoriteAddedName, tenantID, &actorID)
	return FavoriteAdded{Base: base, DocumentID: documentID, OriginalFileName: fileName}
}

type FavoriteRemoved struct {
	Base
	DocumentID       uuid.UUID
	OriginalFileName string
}

func NewFavoriteRemoved(tenantID, actorID, documentID uuid.UUID, fileName string) FavoriteRemoved {
	base := NewBase(FavoriteRemovedName, tenantID, &actorID)
	return FavoriteRemoved{Base: base, DocumentID: documentID, OriginalFileName: fileName}
}

type DocumentProcessed struct {
	Base
	DocumentID       uuid.UUID
	OriginalFileName string
	RecipientUserID  uuid.UUID
	Trigger          string
}

func NewDocumentProcessed(tenantID, documentID, recipientUserID uuid.UUID, fileName, trigger string) DocumentProcessed {
	base := NewBase(DocumentProcessedName, tenantID, nil)
	return DocumentProcessed{Base: base, DocumentID: documentID, OriginalFileName: fileName, RecipientUserID: recipientUserID, Trigger: trigger}
}

type DocumentProcessingFailed struct {
	Base
	DocumentID       uuid.UUID
	OriginalFileName string
	RecipientUserID  uuid.UUID
	Trigger          string
}

func NewDocumentProcessingFailed(tenantID, documentID, recipientUserID uuid.UUID, fileName, trigger string) DocumentProcessingFailed {
	base := NewBase(DocumentProcessingFailedName, tenantID, nil)
	return DocumentProcessingFailed{Base: base, DocumentID: documentID, OriginalFileName: fileName, RecipientUserID: recipientUserID, Trigger: trigger}
}
