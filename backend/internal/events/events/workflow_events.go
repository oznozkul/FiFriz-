package domain

import "github.com/google/uuid"

const (
	WorkflowCreatedName  = "workflow.created"
	WorkflowApprovedName = "workflow.approved"
	WorkflowRejectedName = "workflow.rejected"
)

type WorkflowCreated struct {
	Base
	WorkflowID       uuid.UUID
	DocumentID       uuid.UUID
	OriginalFileName string
	WorkflowTitle    string
	OwnerUserID      uuid.UUID
}

func NewWorkflowCreated(tenantID, actorID, workflowID, documentID, ownerUserID uuid.UUID, fileName, title string) WorkflowCreated {
	base := NewBase(WorkflowCreatedName, tenantID, &actorID)
	return WorkflowCreated{Base: base, WorkflowID: workflowID, DocumentID: documentID, OriginalFileName: fileName, WorkflowTitle: title, OwnerUserID: ownerUserID}
}

type WorkflowApproved struct {
	Base
	WorkflowID       uuid.UUID
	DocumentID       uuid.UUID
	OriginalFileName string
	WorkflowTitle    string
	RecipientUserID  uuid.UUID
}

func NewWorkflowApproved(tenantID, actorID, workflowID, documentID, recipientUserID uuid.UUID, fileName, title string) WorkflowApproved {
	base := NewBase(WorkflowApprovedName, tenantID, &actorID)
	return WorkflowApproved{Base: base, WorkflowID: workflowID, DocumentID: documentID, OriginalFileName: fileName, WorkflowTitle: title, RecipientUserID: recipientUserID}
}

type WorkflowRejected struct {
	Base
	WorkflowID       uuid.UUID
	DocumentID       uuid.UUID
	OriginalFileName string
	WorkflowTitle    string
	RecipientUserID  uuid.UUID
}

func NewWorkflowRejected(tenantID, actorID, workflowID, documentID, recipientUserID uuid.UUID, fileName, title string) WorkflowRejected {
	base := NewBase(WorkflowRejectedName, tenantID, &actorID)
	return WorkflowRejected{Base: base, WorkflowID: workflowID, DocumentID: documentID, OriginalFileName: fileName, WorkflowTitle: title, RecipientUserID: recipientUserID}
}
