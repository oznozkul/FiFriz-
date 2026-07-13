package domain

import "github.com/google/uuid"

const (
	DocumentOpenedName        = "document.opened"
	FolderOpenedName          = "folder.opened"
	SmartCollectionOpenedName = "smart_collection.opened"
	IntegrationOpenedName     = "integration.opened"
	SettingsPageOpenedName    = "settings.page.opened"
	PinnedCreatedName         = "navigation.pinned.created"
	PinnedDeletedName         = "navigation.pinned.deleted"
)

type DocumentOpened struct {
	Base
	DocumentID       uuid.UUID
	OriginalFileName string
	Path             string
}

func NewDocumentOpened(tenantID, actorID, documentID uuid.UUID, fileName, path string) DocumentOpened {
	base := NewBase(DocumentOpenedName, tenantID, &actorID)
	return DocumentOpened{Base: base, DocumentID: documentID, OriginalFileName: fileName, Path: path}
}

type FolderOpened struct {
	Base
	FolderID uuid.UUID
	Name     string
	Path     string
}

func NewFolderOpened(tenantID, actorID, folderID uuid.UUID, name, path string) FolderOpened {
	base := NewBase(FolderOpenedName, tenantID, &actorID)
	return FolderOpened{Base: base, FolderID: folderID, Name: name, Path: path}
}

type SmartCollectionOpened struct {
	Base
	CollectionKey   string
	CollectionTitle string
	Path            string
}

func NewSmartCollectionOpened(tenantID, actorID uuid.UUID, key, title, path string) SmartCollectionOpened {
	base := NewBase(SmartCollectionOpenedName, tenantID, &actorID)
	return SmartCollectionOpened{Base: base, CollectionKey: key, CollectionTitle: title, Path: path}
}

type IntegrationOpened struct {
	Base
	IntegrationID uuid.UUID
	DisplayName   string
	Path          string
}

func NewIntegrationOpened(tenantID, actorID, integrationID uuid.UUID, displayName, path string) IntegrationOpened {
	base := NewBase(IntegrationOpenedName, tenantID, &actorID)
	return IntegrationOpened{Base: base, IntegrationID: integrationID, DisplayName: displayName, Path: path}
}

type SettingsPageOpened struct {
	Base
	PageKey string
	Label   string
	Path    string
}

func NewSettingsPageOpened(tenantID, actorID uuid.UUID, pageKey, label, path string) SettingsPageOpened {
	base := NewBase(SettingsPageOpenedName, tenantID, &actorID)
	return SettingsPageOpened{Base: base, PageKey: pageKey, Label: label, Path: path}
}

type PinnedCreated struct {
	Base
	PinID      uuid.UUID
	EntityType string
	EntityID   string
	Label      string
	Path       string
}

func NewPinnedCreated(tenantID, actorID, pinID uuid.UUID, entityType, entityID, label, path string) PinnedCreated {
	base := NewBase(PinnedCreatedName, tenantID, &actorID)
	return PinnedCreated{Base: base, PinID: pinID, EntityType: entityType, EntityID: entityID, Label: label, Path: path}
}

type PinnedDeleted struct {
	Base
	PinID      uuid.UUID
	EntityType string
	EntityID   string
}

func NewPinnedDeleted(tenantID, actorID, pinID uuid.UUID, entityType, entityID string) PinnedDeleted {
	base := NewBase(PinnedDeletedName, tenantID, &actorID)
	return PinnedDeleted{Base: base, PinID: pinID, EntityType: entityType, EntityID: entityID}
}
