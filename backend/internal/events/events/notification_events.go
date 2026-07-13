package domain

import "github.com/google/uuid"

const NotificationCreatedName = "notification.created"

type NotificationCreated struct {
	Base
	NotificationID uuid.UUID
	TargetUserID   uuid.UUID
	Type           string
	Title          string
}

func NewNotificationCreated(tenantID, actorID, notificationID, targetUserID uuid.UUID, notificationType, title string) NotificationCreated {
	base := NewBase(NotificationCreatedName, tenantID, &actorID)
	return NotificationCreated{
		Base:           base,
		NotificationID: notificationID,
		TargetUserID:   targetUserID,
		Type:           notificationType,
		Title:          title,
	}
}
