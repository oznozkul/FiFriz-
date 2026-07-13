package domain

import "github.com/google/uuid"

const (
	DocumentEmbeddingCompletedName = "document.embedding.completed"
	DocumentSummaryCompletedName     = "document.summary.completed"
	DocumentMetadataCompletedName    = "document.metadata.completed"
	DocumentTagsGeneratedName        = "document.tags.generated"
	DocumentEntitiesExtractedName    = "document.entities.extracted"
	DocumentClassificationCompletedName = "document.classification.completed"
	DocumentPipelineCompletedName    = "document.pipeline.completed"
)

type DocumentEmbeddingCompleted struct {
	Base
	DocumentID uuid.UUID
	UserID     uuid.UUID
}

func NewDocumentEmbeddingCompleted(tenantID, documentID, userID uuid.UUID) DocumentEmbeddingCompleted {
	base := NewBase(DocumentEmbeddingCompletedName, tenantID, optionalActor(userID))
	return DocumentEmbeddingCompleted{Base: base, DocumentID: documentID, UserID: userID}
}

type DocumentSummaryCompleted struct {
	Base
	DocumentID uuid.UUID
	UserID     uuid.UUID
}

func NewDocumentSummaryCompleted(tenantID, documentID, userID uuid.UUID) DocumentSummaryCompleted {
	base := NewBase(DocumentSummaryCompletedName, tenantID, optionalActor(userID))
	return DocumentSummaryCompleted{Base: base, DocumentID: documentID, UserID: userID}
}

type DocumentMetadataCompleted struct {
	Base
	DocumentID uuid.UUID
	UserID     uuid.UUID
}

func NewDocumentMetadataCompleted(tenantID, documentID, userID uuid.UUID) DocumentMetadataCompleted {
	base := NewBase(DocumentMetadataCompletedName, tenantID, optionalActor(userID))
	return DocumentMetadataCompleted{Base: base, DocumentID: documentID, UserID: userID}
}

type DocumentTagsGenerated struct {
	Base
	DocumentID uuid.UUID
	UserID     uuid.UUID
}

func NewDocumentTagsGenerated(tenantID, documentID, userID uuid.UUID) DocumentTagsGenerated {
	base := NewBase(DocumentTagsGeneratedName, tenantID, optionalActor(userID))
	return DocumentTagsGenerated{Base: base, DocumentID: documentID, UserID: userID}
}

type DocumentEntitiesExtracted struct {
	Base
	DocumentID uuid.UUID
	UserID     uuid.UUID
}

func NewDocumentEntitiesExtracted(tenantID, documentID, userID uuid.UUID) DocumentEntitiesExtracted {
	base := NewBase(DocumentEntitiesExtractedName, tenantID, optionalActor(userID))
	return DocumentEntitiesExtracted{Base: base, DocumentID: documentID, UserID: userID}
}

type DocumentClassificationCompleted struct {
	Base
	DocumentID uuid.UUID
	UserID     uuid.UUID
}

func NewDocumentClassificationCompleted(tenantID, documentID, userID uuid.UUID) DocumentClassificationCompleted {
	base := NewBase(DocumentClassificationCompletedName, tenantID, optionalActor(userID))
	return DocumentClassificationCompleted{Base: base, DocumentID: documentID, UserID: userID}
}

type DocumentPipelineCompleted struct {
	Base
	DocumentID uuid.UUID
	UserID     uuid.UUID
}

func NewDocumentPipelineCompleted(tenantID, documentID, userID uuid.UUID) DocumentPipelineCompleted {
	base := NewBase(DocumentPipelineCompletedName, tenantID, optionalActor(userID))
	return DocumentPipelineCompleted{Base: base, DocumentID: documentID, UserID: userID}
}

func optionalActor(userID uuid.UUID) *uuid.UUID {
	if userID == uuid.Nil {
		return nil
	}
	id := userID
	return &id
}
