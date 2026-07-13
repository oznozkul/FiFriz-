package domain

import "github.com/google/uuid"

const (
	CommentCreatedName = "comment.created"
	CommentDeletedName = "comment.deleted"
)

type CommentCreated struct {
	Base
	DocumentID       uuid.UUID
	CommentID        uuid.UUID
	ParentID         *uuid.UUID
	OriginalFileName string
	DocumentOwnerID  uuid.UUID
	ParentAuthorID   *uuid.UUID
}

func NewCommentCreated(tenantID, actorID, documentID, commentID, documentOwnerID uuid.UUID, parentID, parentAuthorID *uuid.UUID, fileName string) CommentCreated {
	base := NewBase(CommentCreatedName, tenantID, &actorID)
	return CommentCreated{
		Base:             base,
		DocumentID:       documentID,
		CommentID:        commentID,
		ParentID:         parentID,
		OriginalFileName: fileName,
		DocumentOwnerID:  documentOwnerID,
		ParentAuthorID:   parentAuthorID,
	}
}

type CommentDeleted struct {
	Base
	DocumentID uuid.UUID
	CommentID  uuid.UUID
}

func NewCommentDeleted(tenantID, actorID, documentID, commentID uuid.UUID) CommentDeleted {
	base := NewBase(CommentDeletedName, tenantID, &actorID)
	return CommentDeleted{Base: base, DocumentID: documentID, CommentID: commentID}
}
