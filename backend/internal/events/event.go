package events

import (
	"time"

	"github.com/google/uuid"
)

type Event interface {
	EventName() string
	TenantID() uuid.UUID
	ActorUserID() *uuid.UUID
	OccurredAt() time.Time
}
