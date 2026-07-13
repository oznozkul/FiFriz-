package audit

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrForbidden        = errors.New("forbidden")
	ErrAuditLogNotFound = errors.New("audit log not found")
)

type Recorder interface {
	CreateLog(ctx context.Context, input CreateLogInput) error
}

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateLog(ctx context.Context, input CreateLogInput) error {
	if s == nil || s.repository == nil {
		return nil
	}

	input.Action = strings.TrimSpace(input.Action)
	input.EntityType = strings.TrimSpace(input.EntityType)
	if input.Action == "" || input.EntityType == "" {
		return errors.New("audit action and entity type are required")
	}

	if input.TenantID == uuid.Nil {
		return errors.New("audit tenant id is required")
	}

	requestMeta := RequestMetaFromContext(ctx)
	if input.IPAddress == "" {
		input.IPAddress = requestMeta.IPAddress
	}
	if input.UserAgent == "" {
		input.UserAgent = requestMeta.UserAgent
	}

	var metadata []byte
	if len(input.Metadata) > 0 {
		encoded, err := json.Marshal(input.Metadata)
		if err != nil {
			return err
		}
		metadata = encoded
	}

	var description *string
	if strings.TrimSpace(input.Description) != "" {
		value := strings.TrimSpace(input.Description)
		description = &value
	}

	var ipAddress *string
	if strings.TrimSpace(input.IPAddress) != "" {
		value := strings.TrimSpace(input.IPAddress)
		ipAddress = &value
	}

	var userAgent *string
	if strings.TrimSpace(input.UserAgent) != "" {
		value := strings.TrimSpace(input.UserAgent)
		userAgent = &value
	}

	logEntry := &AuditLog{
		TenantID:    input.TenantID,
		UserID:      input.UserID,
		Action:      input.Action,
		EntityType:  input.EntityType,
		EntityID:    input.EntityID,
		Description: description,
		Metadata:    metadata,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now().UTC(),
	}

	return s.repository.Create(ctx, logEntry)
}

func Log(ctx context.Context, recorder Recorder, input CreateLogInput) {
	if recorder == nil {
		return
	}
	if err := recorder.CreateLog(ctx, input); err != nil {
		log.Printf("audit log failed: action=%s entity_type=%s error=%v", input.Action, input.EntityType, err)
	}
}

func (s *Service) List(ctx context.Context, tenantID uuid.UUID, filter ListFilter) (*AuditLogListResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant id is required")
	}

	filter.TenantID = tenantID
	rows, totalCount, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]AuditLogResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAuditLogResponse(row))
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	return &AuditLogListResponse{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, tenantID uuid.UUID, id string) (*AuditLogResponse, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant id is required")
	}

	logID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid audit log id")
	}

	row, err := s.repository.GetByID(ctx, tenantID, logID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditLogNotFound
		}
		return nil, err
	}

	response := toAuditLogResponse(*row)
	return &response, nil
}
