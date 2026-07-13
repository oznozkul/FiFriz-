package domain

import "github.com/google/uuid"

const (
	FolderCreatedName           = "folder.created"
	FolderUpdatedName           = "folder.updated"
	FolderDeletedName           = "folder.deleted"
	FolderRenamedName           = "folder.renamed"
	FolderMovedName             = "folder.moved"
	FolderLockedName            = "folder.locked"
	FolderUnlockedName          = "folder.unlocked"
	FolderPermissionChangedName = "folder.permission_changed"
)

type FolderCreated struct {
	Base
	FolderID uuid.UUID
	Name     string
}

func NewFolderCreated(tenantID, actorID, folderID uuid.UUID, name string) FolderCreated {
	base := NewBase(FolderCreatedName, tenantID, &actorID)
	return FolderCreated{Base: base, FolderID: folderID, Name: name}
}

type FolderUpdated struct {
	Base
	FolderID uuid.UUID
	Name     string
}

func NewFolderUpdated(tenantID, actorID, folderID uuid.UUID, name string) FolderUpdated {
	base := NewBase(FolderUpdatedName, tenantID, &actorID)
	return FolderUpdated{Base: base, FolderID: folderID, Name: name}
}

type FolderDeleted struct {
	Base
	FolderID uuid.UUID
	Name     string
}

func NewFolderDeleted(tenantID, actorID, folderID uuid.UUID, name string) FolderDeleted {
	base := NewBase(FolderDeletedName, tenantID, &actorID)
	return FolderDeleted{Base: base, FolderID: folderID, Name: name}
}

type FolderRenamed struct {
	Base
	FolderID uuid.UUID
	Name     string
	OldName  string
}

func NewFolderRenamed(tenantID, actorID, folderID uuid.UUID, oldName, newName string) FolderRenamed {
	base := NewBase(FolderRenamedName, tenantID, &actorID)
	return FolderRenamed{Base: base, FolderID: folderID, Name: newName, OldName: oldName}
}

type FolderMoved struct {
	Base
	FolderID      uuid.UUID
	Name          string
	OldParentID   string
	NewParentID   string
}

func NewFolderMoved(tenantID, actorID, folderID uuid.UUID, name, oldParentID, newParentID string) FolderMoved {
	base := NewBase(FolderMovedName, tenantID, &actorID)
	return FolderMoved{Base: base, FolderID: folderID, Name: name, OldParentID: oldParentID, NewParentID: newParentID}
}

type FolderLocked struct {
	Base
	FolderID uuid.UUID
	Name     string
}

func NewFolderLocked(tenantID, actorID, folderID uuid.UUID, name string) FolderLocked {
	base := NewBase(FolderLockedName, tenantID, &actorID)
	return FolderLocked{Base: base, FolderID: folderID, Name: name}
}

type FolderUnlocked struct {
	Base
	FolderID uuid.UUID
	Name     string
}

func NewFolderUnlocked(tenantID, actorID, folderID uuid.UUID, name string) FolderUnlocked {
	base := NewBase(FolderUnlockedName, tenantID, &actorID)
	return FolderUnlocked{Base: base, FolderID: folderID, Name: name}
}

type FolderPermissionChanged struct {
	Base
	FolderID            uuid.UUID
	Name                string
	InheritPermissions  bool
	PermissionCount     int
}

func NewFolderPermissionChanged(tenantID, actorID, folderID uuid.UUID, name string, inherit bool, count int) FolderPermissionChanged {
	base := NewBase(FolderPermissionChangedName, tenantID, &actorID)
	return FolderPermissionChanged{
		Base:               base,
		FolderID:           folderID,
		Name:               name,
		InheritPermissions: inherit,
		PermissionCount:    count,
	}
}
