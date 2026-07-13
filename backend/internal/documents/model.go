package documents

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type DocumentStatus string

type DocumentApprovalStatus string

const (
	DocumentApprovalStatusNone     DocumentApprovalStatus = "none"
	DocumentApprovalStatusPending  DocumentApprovalStatus = "pending"
	DocumentApprovalStatusApproved DocumentApprovalStatus = "approved"
	DocumentApprovalStatusRejected DocumentApprovalStatus = "rejected"
)

type Document struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	FolderID         *uuid.UUID     `gorm:"type:uuid;index" json:"folder_id"`
	UploadedByUserID uuid.UUID      `gorm:"type:uuid;not null;index" json:"uploaded_by_user_id"`
	OriginalFileName string         `gorm:"size:255" json:"original_file_name"`
	StoredFileName   string         `gorm:"size:255;uniqueIndex" json:"stored_file_name"`
	StorageProvider  string         `gorm:"type:text;not null;default:local" json:"storage_provider"`
	StorageKey       string         `gorm:"type:text" json:"storage_key"`
	BucketName       string         `gorm:"type:text" json:"bucket_name"`
	FilePath         string         `gorm:"size:500" json:"file_path"`
	FileExtension    string         `gorm:"size:20" json:"file_extension"`
	MimeType         string         `gorm:"size:150" json:"mime_type"`
	FileSize         int64          `json:"file_size"`
	CompressedSize   *int64         `json:"compressed_size"`
	ContentEncoding  string         `gorm:"size:32" json:"content_encoding"`
	SHA256Checksum   string         `gorm:"size:64;index" json:"sha256_checksum"`
	StorageStatus    string         `gorm:"size:32;not null;default:available" json:"storage_status"`
	Status           DocumentStatus         `gorm:"size:50;not null;default:uploading" json:"status"`
	ApprovalStatus   DocumentApprovalStatus `gorm:"type:text;not null;default:none" json:"approval_status"`
	IsLocked         bool           `gorm:"not null;default:false;index" json:"is_locked"`
	LockedAt         *time.Time     `gorm:"type:timestamptz" json:"locked_at"`
	LockedByUserID   *uuid.UUID     `gorm:"type:uuid" json:"locked_by_user_id"`
	ArchivedAt       *time.Time     `gorm:"type:timestamptz" json:"archived_at"`
	ExpiresAt        *time.Time     `gorm:"type:timestamptz" json:"expires_at"`
	RetentionLockedAt *time.Time    `gorm:"type:timestamptz" json:"retention_locked_at"`
	SourceType                    string     `gorm:"type:text;not null;default:manual_upload" json:"source_type"`
	SourceIntegrationAccountID    *uuid.UUID `gorm:"type:uuid;index" json:"source_integration_account_id"`
	SourceExternalID              *string    `gorm:"type:text" json:"source_external_id"`
	SourceMetadata                []byte     `gorm:"type:jsonb" json:"source_metadata"`
	IsFavorite       bool                   `gorm:"not null;default:false;index" json:"is_favorite"`
	FavoritedAt      *time.Time     `gorm:"type:timestamptz" json:"favorited_at"`
	AIDocumentType            *string    `gorm:"type:text" json:"ai_document_type"`
	AIDocumentTypeConfidence  *float64   `json:"ai_document_type_confidence"`
	AIClassifiedAt            *time.Time `gorm:"type:timestamptz" json:"ai_classified_at"`
	AIClassificationReason   *string    `gorm:"type:text" json:"ai_classification_reason"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	Text             *DocumentText  `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	FolderName       string         `gorm:"->" json:"-"`
}

func (d *Document) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}

	if d.Status == "" {
		d.Status = DocumentStatusUploading
	}
	if d.ApprovalStatus == "" {
		d.ApprovalStatus = DocumentApprovalStatusNone
	}
	if d.SourceType == "" {
		d.SourceType = "manual_upload"
	}

	return nil
}

type DocumentText struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DocumentID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"document_id"`
	FullText   string    `gorm:"type:text;not null" json:"full_text"`
	CreatedAt  time.Time `json:"created_at"`
}

func (d *DocumentText) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}

	return nil
}

type DocumentChunk struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	DocumentID    uuid.UUID        `gorm:"type:uuid;not null;index;uniqueIndex:idx_document_chunks_document_chunk" json:"document_id"`
	ChunkIndex    int              `gorm:"not null;uniqueIndex:idx_document_chunks_document_chunk" json:"chunk_index"`
	Content       string           `gorm:"type:text;not null" json:"content"`
	TokenEstimate int              `gorm:"not null" json:"token_estimate"`
	Embedding     *pgvector.Vector `gorm:"type:vector(1024)" json:"-"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

func (d *DocumentChunk) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}

	return nil
}
