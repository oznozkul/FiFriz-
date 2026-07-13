package documents

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentVersion struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID         uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	DocumentID       uuid.UUID `gorm:"type:uuid;not null;index" json:"document_id"`
	VersionNumber    int       `gorm:"not null;index" json:"version_number"`
	OriginalFileName string    `gorm:"type:text;not null" json:"original_file_name"`
	StoredFileName   string    `gorm:"type:text;not null" json:"stored_file_name"`
	StorageProvider  string    `gorm:"type:text;not null;default:local" json:"storage_provider"`
	StorageKey       string    `gorm:"type:text" json:"storage_key"`
	BucketName       string    `gorm:"type:text" json:"bucket_name"`
	FilePath         string    `gorm:"type:text" json:"file_path"`
	FileExtension    string    `gorm:"type:text;not null" json:"file_extension"`
	MimeType         string    `gorm:"type:text;not null" json:"mime_type"`
	FileSize         int64     `gorm:"not null" json:"file_size"`
	UploadedByUserID uuid.UUID `gorm:"type:uuid;not null;index" json:"uploaded_by_user_id"`
	ChangeNote       *string   `gorm:"type:text" json:"change_note"`
	IsCurrent        bool      `gorm:"not null;default:false;index" json:"is_current"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (v *DocumentVersion) BeforeCreate(_ *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

func (v *DocumentVersion) TableName() string {
	return "document_versions"
}

func versionFromDocument(document *Document, versionNumber int, isCurrent bool, changeNote *string, uploadedBy uuid.UUID) *DocumentVersion {
	return &DocumentVersion{
		TenantID:         document.TenantID,
		DocumentID:       document.ID,
		VersionNumber:    versionNumber,
		OriginalFileName: document.OriginalFileName,
		StoredFileName:   document.StoredFileName,
		StorageProvider:  document.StorageProvider,
		StorageKey:       document.StorageKey,
		BucketName:       document.BucketName,
		FilePath:         document.FilePath,
		FileExtension:    document.FileExtension,
		MimeType:         document.MimeType,
		FileSize:         document.FileSize,
		UploadedByUserID: uploadedBy,
		ChangeNote:       changeNote,
		IsCurrent:        isCurrent,
	}
}

func applyVersionToDocument(document *Document, version *DocumentVersion) {
	document.OriginalFileName = version.OriginalFileName
	document.StoredFileName = version.StoredFileName
	document.StorageProvider = version.StorageProvider
	document.StorageKey = version.StorageKey
	document.BucketName = version.BucketName
	document.FilePath = version.FilePath
	document.FileExtension = version.FileExtension
	document.MimeType = version.MimeType
	document.FileSize = version.FileSize
}
