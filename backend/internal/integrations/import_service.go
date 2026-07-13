package integrations

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"fifriz/internal/applog"
	"fifriz/internal/documents"
	"fifriz/internal/jobs"
	"fifriz/internal/storage"

	"github.com/google/uuid"
)

const SourceTypeIMAP = "imap"
const SourceTypeGmail = "gmail"
const SourceTypeOutlook = "outlook"
const SourceTypeGoogleDrive = "google_drive"
const SourceTypeOneDrive = "onedrive"

func sourceTypeForProvider(provider string) string {
	switch provider {
	case ProviderIMAP:
		return SourceTypeIMAP
	case ProviderGmail:
		return SourceTypeGmail
	case ProviderOutlook:
		return SourceTypeOutlook
	case ProviderGoogleDrive:
		return SourceTypeGoogleDrive
	case ProviderOneDrive:
		return SourceTypeOneDrive
	default:
		return provider
	}
}

type ImportService struct {
	integrationRepo   *Repository
	documentRepo      *documents.Repository
	storage           *storage.Manager
	jobQueue          jobs.Enqueuer
	documentScheduler interface {
		ScheduleProcessing(ctx context.Context, tenantID, documentID, userID uuid.UUID, trigger string)
	}
}

func NewImportService(
	integrationRepo *Repository,
	documentRepo *documents.Repository,
	storageManager *storage.Manager,
	jobQueue jobs.Enqueuer,
) *ImportService {
	return &ImportService{
		integrationRepo: integrationRepo,
		documentRepo:    documentRepo,
		storage:         storageManager,
		jobQueue:        jobQueue,
	}
}

func (s *ImportService) SetDocumentScheduler(scheduler interface {
	ScheduleProcessing(ctx context.Context, tenantID, documentID, userID uuid.UUID, trigger string)
}) {
	s.documentScheduler = scheduler
}

type ImportFileInput struct {
	Account            *IntegrationAccount
	ExternalID         string
	ExternalType       string
	OriginalFileName   string
	FileExtension      string
	MimeType           string
	Content            []byte
	SourceMetadata     map[string]any
	Checksum           string
	AuditAction        string
	TriggeredByUserID  uuid.UUID
}

type ImportOutcome struct {
	Imported bool
	Skipped  bool
	Failed   bool
	DocumentID *uuid.UUID
}

func (s *ImportService) ImportFile(ctx context.Context, input ImportFileInput) ImportOutcome {
	if s == nil || input.Account == nil {
		return ImportOutcome{Failed: true}
	}

	exists, err := s.integrationRepo.ExistsImportedExternal(ctx, input.Account.TenantID, input.Account.ID, input.ExternalID, input.ExternalType)
	if err != nil {
		applog.Warn("imap import duplicate check failed: integration_id=%s external_id=%s error=%v", input.Account.ID, input.ExternalID, err)
		return ImportOutcome{Failed: true}
	}
	if exists {
		return ImportOutcome{Skipped: true}
	}

	if len(input.Content) == 0 {
		return ImportOutcome{Failed: true}
	}

	documentID := uuid.New()
	storedFileName := documentID.String() + normalizeStoredExtension(input.FileExtension)
	location := s.storage.NewUploadLocation(
		input.Account.TenantID,
		documentID,
		storedFileName,
		&input.Account.ID,
		0,
	)
	if err := s.storage.Upload(ctx, location, bytes.NewReader(input.Content), int64(len(input.Content)), input.MimeType); err != nil {
		applog.Warn("integration import storage upload failed: integration_id=%s error=%v", input.Account.ID, err)
		return ImportOutcome{Failed: true}
	}

	sourceMetadata, err := encodeSourceMetadata(input.SourceMetadata)
	if err != nil {
		_ = s.storage.Delete(ctx, location)
		return ImportOutcome{Failed: true}
	}

	integrationAccountID := input.Account.ID
	sourceExternalID := input.ExternalID
	document := &documents.Document{
		ID:                         documentID,
		TenantID:                   input.Account.TenantID,
		UploadedByUserID:           input.Account.UserID,
		OriginalFileName:           input.OriginalFileName,
		StoredFileName:             storedFileName,
		StorageProvider:            location.Provider,
		StorageKey:                 location.Key,
		BucketName:                 location.BucketName,
		FilePath:                   location.FilePath,
		FileExtension:              strings.TrimPrefix(strings.ToLower(input.FileExtension), "."),
		MimeType:                   input.MimeType,
		FileSize:                   int64(len(input.Content)),
		Status:                     documents.DocumentStatusProcessing,
		SourceType:                 sourceTypeForProvider(input.Account.Provider),
		SourceIntegrationAccountID: &integrationAccountID,
		SourceExternalID:           &sourceExternalID,
		SourceMetadata:             sourceMetadata,
	}

	if err := s.documentRepo.CreateWithInitialVersion(ctx, document); err != nil {
		_ = s.storage.Delete(ctx, location)
		applog.Warn("imap import document create failed: integration_id=%s error=%v", input.Account.ID, err)
		return ImportOutcome{Failed: true}
	}

	checksum := input.Checksum
	if checksum == "" {
		checksum = checksumBytes(input.Content)
	}

	importedItem := &ImportedExternalItem{
		TenantID:             input.Account.TenantID,
		IntegrationAccountID: input.Account.ID,
		ExternalID:           input.ExternalID,
		ExternalType:         input.ExternalType,
		DocumentID:           &documentID,
		Checksum:             &checksum,
		ImportedAt:           time.Now().UTC(),
	}
	if err := s.integrationRepo.CreateImportedExternal(ctx, importedItem); err != nil {
		if errorsIsDuplicate(err) {
			_ = s.storage.Delete(ctx, location)
			return ImportOutcome{Skipped: true}
		}
		applog.Warn("imap import external item failed: integration_id=%s error=%v", input.Account.ID, err)
		return ImportOutcome{Failed: true}
	}

	s.enqueueProcessing(ctx, input.Account.TenantID, documentID, input.TriggeredByUserID)

	return ImportOutcome{Imported: true, DocumentID: &documentID}
}

func (s *ImportService) enqueueProcessing(ctx context.Context, tenantID, documentID, userID uuid.UUID) {
	if s.documentScheduler != nil {
		s.documentScheduler.ScheduleProcessing(ctx, tenantID, documentID, userID, jobs.TriggerIntegrationImport)
		return
	}
	if err := jobs.EnqueueProcessDocument(ctx, s.jobQueue, tenantID, documentID, userID, jobs.TriggerIntegrationImport); err != nil {
		applog.Warn("failed to enqueue integration import processing: document_id=%s error=%v", documentID, err)
	}
}

func checksumBytes(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func normalizeStoredExtension(extension string) string {
	extension = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(extension)), ".")
	if extension == "" {
		return ""
	}
	return "." + extension
}

func errorsIsDuplicate(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate") ||
		strings.Contains(strings.ToLower(err.Error()), "unique")
}

func encodeSourceMetadata(values map[string]any) ([]byte, error) {
	if len(values) == 0 {
		return nil, nil
	}
	return json.Marshal(values)
}
