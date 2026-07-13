package audit

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID      *uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Action      string     `gorm:"type:text;not null;index" json:"action"`
	EntityType  string     `gorm:"type:text;not null;index" json:"entity_type"`
	EntityID    *uuid.UUID `gorm:"type:uuid;index" json:"entity_id"`
	Description *string    `gorm:"type:text" json:"description"`
	Metadata    []byte     `gorm:"type:jsonb" json:"metadata"`
	IPAddress   *string    `gorm:"type:text" json:"ip_address"`
	UserAgent   *string    `gorm:"type:text" json:"user_agent"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;not null;index" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

func (a *AuditLog) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
