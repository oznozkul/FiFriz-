package audit

import (
	"context"

	"github.com/google/uuid"
)

func LogWithActor(ctx context.Context, recorder Recorder, tenantID uuid.UUID, userID *uuid.UUID, input CreateLogInput) {
	if input.TenantID == uuid.Nil {
		input.TenantID = tenantID
	}
	if input.UserID == nil && userID != nil {
		input.UserID = userID
	}
	Log(ctx, recorder, input)
}

func LogDocumentAction(ctx context.Context, recorder Recorder, tenantID uuid.UUID, userID *uuid.UUID, action string, documentID uuid.UUID, description string, metadata map[string]interface{}) {
	entityID := documentID
	LogWithActor(ctx, recorder, tenantID, userID, CreateLogInput{
		TenantID:    tenantID,
		Action:      action,
		EntityType:  EntityTypeDocument,
		EntityID:    &entityID,
		Description: description,
		Metadata:    metadata,
	})
}
