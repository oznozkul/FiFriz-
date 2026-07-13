package domain

import (
	"time"

	"github.com/google/uuid"
)

type Base struct {
	Name     string
	Tenant   uuid.UUID
	Actor    *uuid.UUID
	Occurred time.Time
	Metadata map[string]interface{}
}

func NewBase(name string, tenantID uuid.UUID, actorID *uuid.UUID) Base {
	return Base{
		Name:     name,
		Tenant:   tenantID,
		Actor:    actorID,
		Occurred: time.Now().UTC(),
		Metadata: map[string]interface{}{},
	}
}

func (b Base) EventName() string       { return b.Name }
func (b Base) TenantID() uuid.UUID     { return b.Tenant }
func (b Base) ActorUserID() *uuid.UUID { return b.Actor }
func (b Base) OccurredAt() time.Time   { return b.Occurred }
