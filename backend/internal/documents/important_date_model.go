package documents

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ImportantDateTypeExpiration = "expiration"
	ImportantDateTypeRenewal    = "renewal"
	ImportantDateTypePaymentDue = "payment_due"
	ImportantDateTypeReview     = "review"
	ImportantDateTypeCustom     = "custom"
)

const (
	ImportantDateSourceManual = "manual"
	ImportantDateSourceAI     = "ai"
)

const (
	ImportantDateStatusSuggested = "suggested"
	ImportantDateStatusActive    = "active"
	ImportantDateStatusRejected  = "rejected"
)

type DocumentImportantDate struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID           uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	DocumentID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"document_id"`
	CreatedByUserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"created_by_user_id"`
	AssignedToUserID   *uuid.UUID     `gorm:"type:uuid;index" json:"assigned_to_user_id"`
	Title              string         `gorm:"type:text;not null" json:"title"`
	Date               time.Time      `gorm:"type:date;not null;index" json:"date"`
	ReminderDaysBefore *int           `json:"reminder_days_before"`
	Note               *string        `gorm:"type:text" json:"note"`
	Type               *string        `gorm:"type:text" json:"type"`
	IsCompleted        bool           `gorm:"not null;default:false;index" json:"is_completed"`
	CompletedAt        *time.Time     `json:"completed_at"`
	Source             string         `gorm:"type:text;not null;default:manual;index" json:"source"`
	Status             string         `gorm:"type:text;not null;default:active;index" json:"status"`
	Confidence         *float64       `json:"confidence"`
	SourceText         *string        `gorm:"type:text" json:"source_text"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (d *DocumentImportantDate) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if strings.TrimSpace(d.Source) == "" {
		d.Source = ImportantDateSourceManual
	}
	if strings.TrimSpace(d.Status) == "" {
		d.Status = ImportantDateStatusActive
	}
	return nil
}

func (d *DocumentImportantDate) TableName() string {
	return "document_important_dates"
}
