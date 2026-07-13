package ai

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrSessionNotFound = errors.New("chat session not found")

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateSession(ctx context.Context, session *ChatSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *SessionRepository) UpdateSession(ctx context.Context, session *ChatSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *SessionRepository) GetSession(ctx context.Context, tenantID, userID, sessionID uuid.UUID) (*ChatSession, error) {
	var session ChatSession
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND user_id = ?", sessionID, tenantID, userID).
		First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) ListSessions(ctx context.Context, tenantID, userID uuid.UUID, scopeType *string, scopeID *uuid.UUID) ([]ChatSession, error) {
	query := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Order("updated_at DESC")

	if scopeType != nil && *scopeType != "" {
		query = query.Where("scope_type = ?", *scopeType)
	}
	if scopeID != nil {
		query = query.Where("scope_id = ?", *scopeID)
	}

	var sessions []ChatSession
	err := query.Find(&sessions).Error
	return sessions, err
}

func (r *SessionRepository) DeleteSession(ctx context.Context, tenantID, userID, sessionID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND user_id = ?", sessionID, tenantID, userID).
		Delete(&ChatSession{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (r *SessionRepository) CreateMessage(ctx context.Context, message *ChatMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *SessionRepository) ListMessages(ctx context.Context, tenantID, userID, sessionID uuid.UUID) ([]ChatMessage, error) {
	if _, err := r.GetSession(ctx, tenantID, userID, sessionID); err != nil {
		return nil, err
	}

	var messages []ChatMessage
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND session_id = ?", tenantID, sessionID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func (r *SessionRepository) CountUserMessages(ctx context.Context, tenantID, sessionID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ChatMessage{}).
		Where("tenant_id = ? AND session_id = ? AND role = ?", tenantID, sessionID, MessageRoleUser).
		Count(&count).Error
	return count, err
}
