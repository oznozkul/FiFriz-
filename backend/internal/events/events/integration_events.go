package domain

import "github.com/google/uuid"

const IntegrationSyncCompletedName = "integration.sync.completed"
const IntegrationSyncFailedName = "integration.sync.failed"

type IntegrationSyncCompleted struct {
	Base
	IntegrationAccountID uuid.UUID
	Provider             string
	ImportedCount        int
	TargetUserID         uuid.UUID
	DisplayName          string
}

func NewIntegrationSyncCompleted(tenantID, actorID, integrationID, targetUserID uuid.UUID, provider, displayName string, imported int) IntegrationSyncCompleted {
	base := NewBase(IntegrationSyncCompletedName, tenantID, &actorID)
	return IntegrationSyncCompleted{
		Base:                 base,
		IntegrationAccountID: integrationID,
		Provider:             provider,
		ImportedCount:        imported,
		TargetUserID:         targetUserID,
		DisplayName:          displayName,
	}
}

type IntegrationSyncFailed struct {
	Base
	IntegrationAccountID uuid.UUID
	Provider             string
	TargetUserID         uuid.UUID
	DisplayName          string
	Message              string
}

func NewIntegrationSyncFailed(tenantID, actorID, integrationID, targetUserID uuid.UUID, provider, displayName, message string) IntegrationSyncFailed {
	base := NewBase(IntegrationSyncFailedName, tenantID, &actorID)
	return IntegrationSyncFailed{
		Base:                 base,
		IntegrationAccountID: integrationID,
		Provider:             provider,
		TargetUserID:         targetUserID,
		DisplayName:          displayName,
		Message:              message,
	}
}
