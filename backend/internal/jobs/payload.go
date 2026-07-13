package jobs

import "github.com/google/uuid"

type Job struct {
	Type                 string    `json:"type"`
	TenantID             uuid.UUID `json:"tenantId"`
	DocumentID           uuid.UUID `json:"documentId,omitempty"`
	TriggeredByUserID    uuid.UUID `json:"triggeredByUserId,omitempty"`
	Trigger              string    `json:"trigger,omitempty"`
	PipelineStepName     string    `json:"pipelineStepName,omitempty"`
	IntegrationAccountID uuid.UUID `json:"integrationAccountId,omitempty"`
	SyncType             string    `json:"syncType,omitempty"`
}

type ProcessDocumentPayload struct {
	TenantID          uuid.UUID
	DocumentID        uuid.UUID
	TriggeredByUserID uuid.UUID
	Trigger           string
}

func (j Job) ProcessDocumentPayload() ProcessDocumentPayload {
	return ProcessDocumentPayload{
		TenantID:          j.TenantID,
		DocumentID:        j.DocumentID,
		TriggeredByUserID: j.TriggeredByUserID,
		Trigger:           j.Trigger,
	}
}

func NewProcessDocumentJob(tenantID, documentID, triggeredByUserID uuid.UUID, trigger string) Job {
	return Job{
		Type:              TypeProcessDocument,
		TenantID:          tenantID,
		DocumentID:        documentID,
		TriggeredByUserID: triggeredByUserID,
		Trigger:           trigger,
	}
}

type SyncIntegrationPayload struct {
	TenantID              uuid.UUID
	IntegrationAccountID uuid.UUID
	TriggeredByUserID   uuid.UUID
	SyncType            string
}

func (j Job) SyncIntegrationPayload() SyncIntegrationPayload {
	return SyncIntegrationPayload{
		TenantID:              j.TenantID,
		IntegrationAccountID:  j.IntegrationAccountID,
		TriggeredByUserID:     j.TriggeredByUserID,
		SyncType:              j.SyncType,
	}
}

func NewSyncIntegrationJob(tenantID, integrationAccountID, userID uuid.UUID, syncType string) Job {
	return Job{
		Type:                 TypeSyncIntegration,
		TenantID:             tenantID,
		IntegrationAccountID: integrationAccountID,
		TriggeredByUserID:    userID,
		SyncType:             syncType,
	}
}

func NewCompressDocumentJob(tenantID, documentID uuid.UUID) Job {
	return Job{
		Type:       TypeCompressDocument,
		TenantID:   tenantID,
		DocumentID: documentID,
	}
}

func (j Job) CompressDocumentPayload() CompressDocumentPayload {
	return CompressDocumentPayload{
		TenantID:   j.TenantID,
		DocumentID: j.DocumentID,
	}
}

type CompressDocumentPayload struct {
	TenantID   uuid.UUID
	DocumentID uuid.UUID
}
