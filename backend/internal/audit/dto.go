package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateLogInput struct {
	TenantID    uuid.UUID
	UserID      *uuid.UUID
	Action      string
	EntityType  string
	EntityID    *uuid.UUID
	Description string
	Metadata    map[string]interface{}
	IPAddress   string
	UserAgent   string
}

type AuditLogResponse struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenantId"`
	UserID       *uuid.UUID             `json:"userId"`
	UserEmail    string                 `json:"userEmail"`
	UserFullName string                 `json:"userFullName"`
	Action       string                 `json:"action"`
	EntityType   string                 `json:"entityType"`
	EntityID     *uuid.UUID             `json:"entityId"`
	Description  string                 `json:"description"`
	Metadata     map[string]interface{} `json:"metadata"`
	IPAddress    string                 `json:"ipAddress"`
	UserAgent    string                 `json:"userAgent"`
	CreatedAt    time.Time              `json:"createdAt"`
}

type AuditLogListResponse struct {
	Items      []AuditLogResponse `json:"items"`
	TotalCount int64              `json:"totalCount"`
	Page       int                `json:"page"`
	PageSize   int                `json:"pageSize"`
}

type ListFilter struct {
	TenantID   uuid.UUID
	Action     string
	EntityType string
	UserID     *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	PageSize   int
}

type auditLogRow struct {
	ID           uuid.UUID  `gorm:"column:id"`
	TenantID     uuid.UUID  `gorm:"column:tenant_id"`
	UserID       *uuid.UUID `gorm:"column:user_id"`
	UserEmail    string     `gorm:"column:user_email"`
	UserFullName string     `gorm:"column:user_full_name"`
	Action       string     `gorm:"column:action"`
	EntityType   string     `gorm:"column:entity_type"`
	EntityID     *uuid.UUID `gorm:"column:entity_id"`
	Description  *string    `gorm:"column:description"`
	Metadata     []byte     `gorm:"column:metadata"`
	IPAddress    *string    `gorm:"column:ip_address"`
	UserAgent    *string    `gorm:"column:user_agent"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
}

func toAuditLogResponse(row auditLogRow) AuditLogResponse {
	description := ""
	if row.Description != nil {
		description = *row.Description
	}

	ipAddress := ""
	if row.IPAddress != nil {
		ipAddress = *row.IPAddress
	}

	userAgent := ""
	if row.UserAgent != nil {
		userAgent = *row.UserAgent
	}

	metadata := map[string]interface{}{}
	if len(row.Metadata) > 0 {
		_ = json.Unmarshal(row.Metadata, &metadata)
	}

	return AuditLogResponse{
		ID:           row.ID,
		TenantID:     row.TenantID,
		UserID:       row.UserID,
		UserEmail:    row.UserEmail,
		UserFullName: row.UserFullName,
		Action:       row.Action,
		EntityType:   row.EntityType,
		EntityID:     row.EntityID,
		Description:  description,
		Metadata:     metadata,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    row.CreatedAt,
	}
}
