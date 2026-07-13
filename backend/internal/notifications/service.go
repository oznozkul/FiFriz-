package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"fifriz/internal/auth"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrInvalidNotificationID = errors.New("invalid notification id")
)

type Creator interface {
	CreateNotification(ctx context.Context, input CreateInput) error
}

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateNotification(ctx context.Context, input CreateInput) error {
	if s == nil || s.repository == nil {
		return nil
	}
	if input.UserID == uuid.Nil || input.TenantID == uuid.Nil {
		return nil
	}
	if input.ActorUserID != uuid.Nil && input.UserID == input.ActorUserID {
		return nil
	}

	input.Type = strings.TrimSpace(input.Type)
	input.Title = strings.TrimSpace(input.Title)
	input.Message = strings.TrimSpace(input.Message)
	if input.Type == "" || input.Title == "" || input.Message == "" {
		return errors.New("notification type, title and message are required")
	}

	var metadata []byte
	if len(input.Metadata) > 0 {
		encoded, err := json.Marshal(input.Metadata)
		if err != nil {
			return err
		}
		metadata = encoded
	}

	notification := &Notification{
		TenantID:   input.TenantID,
		UserID:     input.UserID,
		Type:       input.Type,
		Title:      input.Title,
		Message:    input.Message,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Metadata:   metadata,
	}

	return s.repository.Create(ctx, notification)
}

func Notify(creator Creator, ctx context.Context, input CreateInput) {
	if creator == nil {
		return
	}
	if err := creator.CreateNotification(ctx, input); err != nil {
		log.Printf("notification failed: type=%s user_id=%s error=%v", input.Type, input.UserID, err)
	}
}

func (s *Service) List(ctx context.Context, filter ListFilter) (*NotificationListResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	filter.TenantID = authContext.TenantID
	filter.UserID = authContext.UserID

	items, totalCount, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]NotificationResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toNotificationResponse(item))
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	return &NotificationListResponse{
		Items:      responses,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s *Service) UnreadCount(ctx context.Context) (*UnreadCountResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := s.repository.CountUnread(ctx, authContext.TenantID, authContext.UserID)
	if err != nil {
		return nil, err
	}

	return &UnreadCountResponse{Count: count}, nil
}

func (s *Service) MarkRead(ctx context.Context, idRaw string) (*NotificationResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idRaw)
	if err != nil {
		return nil, ErrInvalidNotificationID
	}

	readAt := time.Now().UTC()
	if err := s.repository.MarkRead(ctx, authContext.TenantID, authContext.UserID, id, readAt); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotificationNotFound
		}
		return nil, err
	}

	notification, err := s.repository.GetByID(ctx, authContext.TenantID, authContext.UserID, id)
	if err != nil {
		return nil, err
	}

	response := toNotificationResponse(*notification)
	return &response, nil
}

func (s *Service) MarkAllRead(ctx context.Context) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	return s.repository.MarkAllRead(ctx, authContext.TenantID, authContext.UserID, time.Now().UTC())
}
