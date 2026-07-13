package ai

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ScopeTypeAll        = "all"
	ScopeTypeDocument   = "document"
	ScopeTypeFolder     = "folder"
	ScopeTypeTag        = "tag"
	ScopeTypeCollection = "collection"
	ScopeTypeWorkspace  = "workspace"
	ScopeTypeAgent      = "agent"

	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"
)

type ChatSession struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index:idx_ai_chat_sessions_tenant_user,priority:1" json:"tenantId"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_ai_chat_sessions_tenant_user,priority:2" json:"userId"`
	Title     string         `gorm:"type:text;not null" json:"title"`
	ScopeType *string        `gorm:"type:text;index" json:"scopeType"`
	ScopeID   *uuid.UUID     `gorm:"type:uuid;index" json:"scopeId"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (s *ChatSession) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (ChatSession) TableName() string {
	return "ai_chat_sessions"
}

type ChatMessage struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenantId"`
	SessionID uuid.UUID  `gorm:"type:uuid;not null;index" json:"sessionId"`
	UserID    *uuid.UUID `gorm:"type:uuid;index" json:"userId"`
	Role      string     `gorm:"type:text;not null;index" json:"role"`
	Content   string     `gorm:"type:text;not null" json:"content"`
	Sources   []byte     `gorm:"type:jsonb" json:"sources"`
	CreatedAt time.Time  `json:"createdAt"`
}

func (m *ChatMessage) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (ChatMessage) TableName() string {
	return "ai_chat_messages"
}
