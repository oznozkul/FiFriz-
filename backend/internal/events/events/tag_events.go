package domain

import "github.com/google/uuid"

const (
	TagCreatedName       = "tag.created"
	TagUpdatedName       = "tag.updated"
	TagDeletedName       = "tag.deleted"
	TagArchivedName      = "tag.archived"
	TagMergedName        = "tag.merged"
	TagAliasCreatedName  = "tag.alias.created"
	TagAliasDeletedName  = "tag.alias.deleted"
	TagAssignedName      = "tag.assigned"
)

type TagCreated struct {
	Base
	TagID uuid.UUID
	Name  string
}

func NewTagCreated(tenantID, actorID, tagID uuid.UUID, name string) TagCreated {
	base := NewBase(TagCreatedName, tenantID, &actorID)
	return TagCreated{Base: base, TagID: tagID, Name: name}
}

type TagUpdated struct {
	Base
	TagID uuid.UUID
	Name  string
}

func NewTagUpdated(tenantID, actorID, tagID uuid.UUID, name string) TagUpdated {
	base := NewBase(TagUpdatedName, tenantID, &actorID)
	return TagUpdated{Base: base, TagID: tagID, Name: name}
}

type TagDeleted struct {
	Base
	TagID uuid.UUID
	Name  string
}

func NewTagDeleted(tenantID, actorID, tagID uuid.UUID, name string) TagDeleted {
	base := NewBase(TagDeletedName, tenantID, &actorID)
	return TagDeleted{Base: base, TagID: tagID, Name: name}
}

type TagArchived struct {
	Base
	TagID uuid.UUID
	Name  string
}

func NewTagArchived(tenantID, actorID, tagID uuid.UUID, name string) TagArchived {
	base := NewBase(TagArchivedName, tenantID, &actorID)
	return TagArchived{Base: base, TagID: tagID, Name: name}
}

type TagMerged struct {
	Base
	SourceTagID uuid.UUID
	TargetTagID uuid.UUID
	SourceName  string
	TargetName  string
}

func NewTagMerged(tenantID, actorID, sourceTagID, targetTagID uuid.UUID, sourceName, targetName string) TagMerged {
	base := NewBase(TagMergedName, tenantID, &actorID)
	return TagMerged{
		Base:        base,
		SourceTagID: sourceTagID,
		TargetTagID: targetTagID,
		SourceName:  sourceName,
		TargetName:  targetName,
	}
}

type TagAliasCreated struct {
	Base
	TagID uuid.UUID
	Name  string
	Alias string
}

func NewTagAliasCreated(tenantID, actorID, tagID uuid.UUID, name, alias string) TagAliasCreated {
	base := NewBase(TagAliasCreatedName, tenantID, &actorID)
	return TagAliasCreated{Base: base, TagID: tagID, Name: name, Alias: alias}
}

type TagAliasDeleted struct {
	Base
	TagID uuid.UUID
	Name  string
	Alias string
}

func NewTagAliasDeleted(tenantID, actorID, tagID uuid.UUID, name, alias string) TagAliasDeleted {
	base := NewBase(TagAliasDeletedName, tenantID, &actorID)
	return TagAliasDeleted{Base: base, TagID: tagID, Name: name, Alias: alias}
}

type TagAssigned struct {
	Base
	DocumentID uuid.UUID
	TagIDs     []uuid.UUID
}

func NewTagAssigned(tenantID, actorID, documentID uuid.UUID, tagIDs []uuid.UUID) TagAssigned {
	base := NewBase(TagAssignedName, tenantID, &actorID)
	return TagAssigned{Base: base, DocumentID: documentID, TagIDs: tagIDs}
}
