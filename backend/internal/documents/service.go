package documents

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"

	"fifriz/internal/ai/pipeline"
	"fifriz/internal/applog"
	"fifriz/internal/audit"
	"fifriz/internal/auth"
	"fifriz/internal/cache"
	"fifriz/internal/documents/validation"
	"fifriz/internal/documents/virusscan"
	"fifriz/internal/embedding"
	"fifriz/internal/events"
	domain "fifriz/internal/events/events"
	"fifriz/internal/jobs"
	"fifriz/internal/permissions"
	"fifriz/internal/storage"
	"fifriz/internal/storage/compression"
	"fifriz/internal/tags"
)

var (
	ErrDocumentNotFound          = errors.New("document not found")
	ErrInvalidDocumentID         = errors.New("invalid document id")
	ErrMissingUploadFile         = errors.New("upload file is required")
	ErrEmptyUploadFile           = errors.New("upload file is empty")
	ErrUnsupportedUploadFile     = errors.New("unsupported upload file type")
	ErrUploadMetadataSaveFail    = errors.New("could not save upload metadata")
	ErrInvalidUpdateData         = errors.New("invalid update document data")
	ErrDocumentTextNotFound      = errors.New("document text not found")
	ErrInvalidSearchQuery        = errors.New("invalid search query")
	ErrInvalidFolderID           = errors.New("invalid folder id")
	ErrFolderNotFound            = errors.New("folder not found")
	ErrForbidden                 = errors.New("forbidden")
	ErrEmbeddingProviderDisabled = errors.New("embedding provider is disabled")
	ErrTextExtractionUnsupported = errors.New("text extraction unsupported for document")
	ErrTextExtractionFailed      = errors.New("text extraction failed")
	ErrInvalidTagID              = errors.New("invalid tag id")
	ErrTagNotFound               = errors.New("tag not found")
	ErrDocumentLocked            = errors.New("document is locked")
	ErrFileTooLarge              = errors.New("file exceeds maximum allowed size")
	ErrMIMETypeNotAllowed        = errors.New("mime type is not allowed")
	ErrInvalidMIMEContent        = errors.New("file content does not match extension")
)

type Service struct {
	repository          *Repository
	storage             *storage.Manager
	embeddingProvider   embedding.Provider
	publisher           events.Publisher
	tags                *tags.Service
	shareAccess         *ShareAccessService
	jobQueue            jobs.Enqueuer
	cache               *cache.Service
	summaryChatProvider SummaryChatCompleter
	summaryChatModel    string
	auditRecorder       audit.Recorder
	pipelineRunner      *pipeline.Runner
	filePolicy          *validation.PolicyResolver
	virusScanner        virusscan.Scanner
	downloadMaxRetries  int
	downloadTimeout     time.Duration
}

type DocumentFile struct {
	Reader           io.ReadCloser
	OriginalFileName string
	MimeType         string
	Size             int64
}

func NewService(repository *Repository, storageManager *storage.Manager, embeddingProvider embedding.Provider, publisher events.Publisher, tagService *tags.Service, shareAccess *ShareAccessService, jobQueue jobs.Enqueuer, cacheService *cache.Service) *Service {
	return &Service{
		repository:        repository,
		storage:           storageManager,
		embeddingProvider: embeddingProvider,
		publisher:         publisher,
		tags:              tagService,
		shareAccess:       shareAccess,
		jobQueue:          jobQueue,
		cache:             cacheService,
		virusScanner:      virusscan.NoOpScanner{},
		downloadMaxRetries: 3,
		downloadTimeout:   30 * time.Second,
	}
}

func (s *Service) List(ctx context.Context, folderIDFilter string, tagIDFilter string) ([]DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	accessUserID := s.memberAccessUserID(authContext)
	var items []Document
	tagIDFilter = strings.TrimSpace(tagIDFilter)
	if tagIDFilter != "" {
		tagID, err := uuid.Parse(tagIDFilter)
		if err != nil {
			return nil, ErrInvalidTagID
		}
		if s.tags != nil {
			if _, err := s.tags.GetByID(ctx, tagIDFilter); err != nil {
				if errors.Is(err, tags.ErrTagNotFound) {
					return nil, ErrTagNotFound
				}
				if errors.Is(err, tags.ErrInvalidTagID) {
					return nil, ErrInvalidTagID
				}
				return nil, err
			}
		}
		items, err = s.repository.ListByTag(ctx, authContext.TenantID, tagID, nil, accessUserID)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(folderIDFilter) != "" {
			items, err = s.filterDocumentsByFolder(ctx, authContext.TenantID, items, folderIDFilter)
			if err != nil {
				return nil, err
			}
		}
	} else if strings.TrimSpace(folderIDFilter) == "" {
		items, err = s.repository.List(ctx, authContext.TenantID, accessUserID)
	} else {
		items, err = s.listByFolderID(ctx, authContext.TenantID, folderIDFilter, accessUserID)
	}
	if err != nil {
		return nil, err
	}

	return s.mapDocumentsToResponses(ctx, authContext.TenantID, items)
}

func (s *Service) ListByTag(ctx context.Context, tagID string) ([]DocumentResponse, error) {
	return s.List(ctx, "", tagID)
}

func (s *Service) ListByFolder(ctx context.Context, folderID string) ([]DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	items, err := s.listByFolderID(ctx, authContext.TenantID, folderID, s.memberAccessUserID(authContext))
	if err != nil {
		return nil, err
	}

	return s.mapDocumentsToResponses(ctx, authContext.TenantID, items)
}

func (s *Service) filterDocumentsByFolder(ctx context.Context, tenantID uuid.UUID, items []Document, folderIDFilter string) ([]Document, error) {
	folderIDFilter = strings.TrimSpace(folderIDFilter)
	if folderIDFilter == "" {
		return items, nil
	}

	var targetFolderID *uuid.UUID
	if folderIDFilter == "root" {
		targetFolderID = nil
	} else {
		folderID, err := uuid.Parse(folderIDFilter)
		if err != nil {
			return nil, ErrInvalidFolderID
		}
		exists, err := s.repository.FolderExists(ctx, tenantID, folderID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrFolderNotFound
		}
		targetFolderID = &folderID
	}

	filtered := make([]Document, 0, len(items))
	for _, item := range items {
		if targetFolderID == nil {
			if item.FolderID == nil {
				filtered = append(filtered, item)
			}
			continue
		}
		if item.FolderID != nil && *item.FolderID == *targetFolderID {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

func (s *Service) listByFolderID(ctx context.Context, tenantID uuid.UUID, folderIDFilter string, accessUserID *uuid.UUID) ([]Document, error) {
	folderIDFilter = strings.TrimSpace(folderIDFilter)
	if folderIDFilter == "" || folderIDFilter == "root" {
		return s.repository.ListByFolder(ctx, tenantID, nil, accessUserID)
	}

	folderID, err := uuid.Parse(folderIDFilter)
	if err != nil {
		return nil, ErrInvalidFolderID
	}

	exists, err := s.repository.FolderExists(ctx, tenantID, folderID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrFolderNotFound
	}

	return s.repository.ListByFolder(ctx, tenantID, &folderID, accessUserID)
}

func (s *Service) ListDeleted(ctx context.Context) ([]DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	items, err := s.repository.ListDeleted(ctx, authContext.TenantID, s.memberAccessUserID(authContext))
	if err != nil {
		return nil, err
	}

	return s.mapDocumentsToResponses(ctx, authContext.TenantID, items)
}

func (s *Service) ListFavorites(ctx context.Context) ([]DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	items, err := s.repository.ListFavorites(ctx, authContext.TenantID, s.memberAccessUserID(authContext))
	if err != nil {
		return nil, err
	}

	return s.mapDocumentsToResponses(ctx, authContext.TenantID, items)
}

func (s *Service) Favorite(ctx context.Context, id string) (*DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	if err := s.requireViewAccessForFavorite(ctx, authContext, document); err != nil {
		return nil, err
	}

	if !document.IsFavorite {
		now := time.Now().UTC()
		document.IsFavorite = true
		document.FavoritedAt = &now
		if err := s.repository.Update(ctx, document); err != nil {
			return nil, err
		}

		s.publisher.PublishAfterCommit(ctx, domain.NewFavoriteAdded(authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName))
	}

	return s.toResponsePtr(ctx, authContext.TenantID, *document)
}

func (s *Service) Unfavorite(ctx context.Context, id string) (*DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	if err := s.requireViewAccessForFavorite(ctx, authContext, document); err != nil {
		return nil, err
	}

	if document.IsFavorite {
		document.IsFavorite = false
		document.FavoritedAt = nil
		if err := s.repository.Update(ctx, document); err != nil {
			return nil, err
		}

		s.publisher.PublishAfterCommit(ctx, domain.NewFavoriteRemoved(authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName))
	}

	return s.toResponsePtr(ctx, authContext.TenantID, *document)
}

func (s *Service) GetByID(ctx context.Context, id string) (*DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	s.publisher.Publish(ctx, domain.NewDocumentOpened(
		authContext.TenantID,
		authContext.UserID,
		document.ID,
		document.OriginalFileName,
		"/documents/"+document.ID.String(),
	))

	return s.toResponsePtr(ctx, authContext.TenantID, *document)
}

func (s *Service) GetWorkspace(ctx context.Context, id string) (*DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	if s.cache == nil {
		return s.getWorkspaceUncached(ctx, authContext, documentID)
	}

	result, err := cache.GetOrSetJSON(ctx, s.cache, cache.DocumentWorkspaceKey(authContext.TenantID, documentID), cache.TTLDocumentWorkspace, func() (*DocumentResponse, error) {
		return s.loadWorkspaceDocument(ctx, authContext, documentID)
	})
	if err != nil {
		return nil, err
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	s.publisher.Publish(ctx, domain.NewDocumentOpened(
		authContext.TenantID,
		authContext.UserID,
		documentID,
		result.OriginalFileName,
		"/documents/"+documentID.String(),
	))

	return result, nil
}

func (s *Service) getWorkspaceUncached(ctx context.Context, authContext auth.Context, documentID uuid.UUID) (*DocumentResponse, error) {
	result, err := s.loadWorkspaceDocument(ctx, authContext, documentID)
	if err != nil {
		return nil, err
	}

	s.publisher.Publish(ctx, domain.NewDocumentOpened(
		authContext.TenantID,
		authContext.UserID,
		documentID,
		result.OriginalFileName,
		"/documents/"+documentID.String(),
	))

	return result, nil
}

func (s *Service) loadWorkspaceDocument(ctx context.Context, authContext auth.Context, documentID uuid.UUID) (*DocumentResponse, error) {
	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	return s.toResponsePtr(ctx, authContext.TenantID, *document)
}

func (s *Service) PrepareUpload(fileHeader *multipart.FileHeader) (*Document, error) {
	if fileHeader == nil {
		return nil, ErrMissingUploadFile
	}

	if fileHeader.Size <= 0 {
		return nil, ErrEmptyUploadFile
	}

	extension := normalizeExtension(fileHeader.Filename)
	if !isAllowedExtension(extension) {
		return nil, ErrUnsupportedUploadFile
	}

	documentID := uuid.New()
	storedFileName := documentID.String() + extension

	document := &Document{
		ID:               documentID,
		OriginalFileName: filepath.Base(fileHeader.Filename),
		StoredFileName:   storedFileName,
		FileExtension:    strings.TrimPrefix(extension, "."),
		MimeType:         detectMimeType(fileHeader),
		FileSize:         fileHeader.Size,
		Status:           DocumentStatusUploading,
	}

	return document, nil
}

func (s *Service) SaveUploaded(ctx context.Context, document *Document, reader io.Reader, size int64, folderIDRaw string, tagIDStrings []string) (*DocumentUploadResult, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	extension := normalizeExtension(document.OriginalFileName)
	validated, err := s.validateUploadContent(ctx, authContext.TenantID, &multipart.FileHeader{
		Filename: document.OriginalFileName,
		Size:     size,
	}, reader, extension)
	if err != nil {
		return nil, err
	}

	document.TenantID = authContext.TenantID
	document.UploadedByUserID = authContext.UserID
	document.MimeType = validated.mimeType
	document.SHA256Checksum = validated.checksum
	document.StorageStatus = "available"
	document.Status = DocumentStatusUploading
	document.FileSize = int64(len(validated.data))
	folderID, err := s.parseFolderID(ctx, authContext.TenantID, folderIDRaw)
	if err != nil {
		return nil, err
	}
	document.FolderID = folderID

	if err := s.storeDocumentFile(ctx, authContext.TenantID, document.ID, document, bytes.NewReader(validated.data), int64(len(validated.data)), 0, nil); err != nil {
		return nil, ErrUploadMetadataSaveFail
	}

	applog.Info("upload started: document_id=%s tenant_id=%s user_id=%s file=%s", document.ID, authContext.TenantID, authContext.UserID, document.OriginalFileName)

	if err := s.repository.CreateWithInitialVersion(ctx, document); err != nil {
		return nil, ErrUploadMetadataSaveFail
	}

	document.Status = DocumentStatusProcessing
	if err := s.repository.Update(ctx, document); err != nil {
		return nil, ErrUploadMetadataSaveFail
	}

	s.scheduleDocumentProcessing(ctx, authContext.TenantID, document.ID, authContext.UserID, jobs.TriggerUpload)

	applog.Info("upload queued: document_id=%s tenant_id=%s", document.ID, authContext.TenantID)

	s.publisher.PublishAfterCommit(ctx, domain.NewDocumentCreated(authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName, document.FileExtension, document.FolderID))

	if len(tagIDStrings) > 0 && s.tags != nil {
		tagIDs, err := s.tags.ParseTagIDs(tagIDStrings)
		if err != nil {
			return nil, ErrInvalidTagID
		}
		if err := s.tags.ReplaceDocumentTags(ctx, authContext.TenantID, document.ID, tagIDs); err != nil {
			if errors.Is(err, tags.ErrTagNotFound) {
				return nil, ErrTagNotFound
			}
			return nil, err
		}
	}

	if reloadedDocument, err := s.repository.GetByID(ctx, authContext.TenantID, document.ID); err == nil {
		document = reloadedDocument
	}

	response, err := s.toResponsePtr(ctx, authContext.TenantID, *document)
	if err != nil {
		return nil, err
	}

	return &DocumentUploadResult{
		Document:    *response,
		Warnings:    validated.warnings,
		DuplicateOf: validated.duplicate,
	}, nil
}

func (s *Service) parseFolderID(ctx context.Context, tenantID uuid.UUID, raw string) (*uuid.UUID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	folderID, err := uuid.Parse(raw)
	if err != nil {
		return nil, ErrInvalidFolderID
	}

	exists, err := s.repository.FolderExists(ctx, tenantID, folderID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrFolderNotFound
	}

	return &folderID, nil
}

func (s *Service) GetFile(ctx context.Context, id string, accessMode string) (*DocumentFile, error) {
	if accessMode == "" {
		accessMode = "preview"
	}
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	if err := s.requireDownloadAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	location, err := documentStorageLocation(s.storage, document)
	if err != nil {
		return nil, err
	}

	reader, err := s.storage.DownloadWithRetry(ctx, location, s.downloadMaxRetries, s.downloadTimeout)
	if err != nil {
		document.StorageStatus = "unavailable"
		_ = s.repository.Update(ctx, document)
		return nil, err
	}

	s.logDocumentFileAccess(ctx, authContext, document, accessMode)

	return &DocumentFile{
		Reader:           reader,
		OriginalFileName: document.OriginalFileName,
		MimeType:         document.MimeType,
		Size:             document.FileSize,
	}, nil
}

func (s *Service) GetText(ctx context.Context, id string) (*DocumentTextResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	documentText, err := s.repository.GetTextByDocumentID(ctx, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentTextNotFound
		}
		return nil, err
	}

	response := toDocumentTextResponse(*documentText)
	return &response, nil
}

func (s *Service) GetChunks(ctx context.Context, id string) ([]DocumentChunkResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	chunks, err := s.repository.ListChunksByDocumentID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	return toDocumentChunkResponses(chunks), nil
}

func (s *Service) GetChunkStats(ctx context.Context, id string) (*DocumentChunkStatsResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	chunks, err := s.repository.ListChunksByDocumentID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	fullText := ""
	if document.Text != nil {
		fullText = document.Text.FullText
	}

	totalWords := countWords(fullText)
	chunkIndexes := make([]int, 0, len(chunks))
	totalChunkWords := 0

	for _, chunk := range chunks {
		chunkIndexes = append(chunkIndexes, chunk.ChunkIndex)
		totalChunkWords += countWords(chunk.Content)
	}

	averageChunkSize := 0
	if len(chunks) > 0 {
		averageChunkSize = totalChunkWords / len(chunks)
	}

	return &DocumentChunkStatsResponse{
		DocumentID:       documentID,
		TotalCharacters:  utf8.RuneCountInString(fullText),
		TotalWords:       totalWords,
		ChunkCount:       len(chunks),
		AverageChunkSize: averageChunkSize,
		ChunkIndexes:     chunkIndexes,
	}, nil
}

func (s *Service) ReprocessText(ctx context.Context, id string) (*DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	if !isTextExtractionSupported(document.FileExtension) {
		return nil, ErrTextExtractionUnsupported
	}

	document.Status = DocumentStatusProcessing
	if err := s.repository.Update(ctx, document); err != nil {
		return nil, err
	}

	s.scheduleDocumentProcessing(ctx, authContext.TenantID, document.ID, authContext.UserID, jobs.TriggerReprocess)

	s.publisher.PublishAfterCommit(ctx, domain.NewDocumentUpdated(authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName))

	return s.toResponsePtr(ctx, authContext.TenantID, *document)
}

func (s *Service) Search(ctx context.Context, query string) ([]DocumentSearchResultResponse, error) {
	return s.searchDocuments(ctx, query, true)
}

func (s *Service) SearchForGlobal(ctx context.Context, query string) ([]DocumentSearchResultResponse, error) {
	return s.searchDocuments(ctx, query, false)
}

func (s *Service) searchDocuments(ctx context.Context, query string, publishEvent bool) ([]DocumentSearchResultResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	normalizedQuery := buildPrefixSearchQuery(query)
	if normalizedQuery == "" {
		return nil, ErrInvalidSearchQuery
	}

	results, err := s.repository.SearchText(ctx, authContext.TenantID, normalizedQuery, s.memberAccessUserID(authContext))
	if err != nil {
		return nil, err
	}

	responses := make([]DocumentSearchResultResponse, 0, len(results))
	for _, result := range results {
		responses = append(responses, DocumentSearchResultResponse{
			DocumentID:       result.DocumentID,
			Title:            result.Title,
			OriginalFileName: result.OriginalFileName,
			FileExtension:    result.FileExtension,
			MatchedSnippet:   formatSearchSnippet(result.FullText, query, result.MatchedSnippet),
			CreatedAt:        result.CreatedAt,
		})
	}

	if publishEvent {
		s.publisher.PublishAfterCommit(ctx, domain.NewDocumentSearched(authContext.TenantID, authContext.UserID, strings.TrimSpace(query), len(responses)))
	}

	return responses, nil
}

func (s *Service) SemanticSearch(ctx context.Context, query string) ([]SemanticSearchResultResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return nil, ErrInvalidSearchQuery
	}

	if s.embeddingProvider == nil {
		return nil, ErrEmbeddingProviderDisabled
	}

	embeddingValues, err := s.embeddingProvider.CreateEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(embeddingValues) != embedding.DefaultEmbeddingDimensions {
		return nil, fmt.Errorf("query embedding dimension mismatch: got %d want %d", len(embeddingValues), embedding.DefaultEmbeddingDimensions)
	}

	results, err := s.repository.SearchChunksByEmbedding(ctx, authContext.TenantID, embeddingValues, 10, s.memberAccessUserID(authContext))
	if err != nil {
		return nil, err
	}

	responses := make([]SemanticSearchResultResponse, 0, len(results))
	for _, result := range results {
		responses = append(responses, SemanticSearchResultResponse{
			DocumentID: result.DocumentID,
			ChunkID:    result.ChunkID,
			ChunkIndex: result.ChunkIndex,
			Content:    result.Content,
			Similarity: result.Similarity,
		})
	}

	return responses, nil
}

func formatSearchSnippet(fullText, query, fallback string) string {
	if snippet := buildOriginalTextSnippet(fullText, query); snippet != "" {
		return formatSnippet(snippet)
	}
	return formatSnippet(fallback)
}

func formatSnippet(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if value == "" {
		return ""
	}
	return "... " + value + " ..."
}

func buildOriginalTextSnippet(fullText, query string) string {
	terms := extractNormalizedSearchTerms(query)
	if fullText == "" || len(terms) == 0 {
		return ""
	}

	normalizedText, normalizedRuneToOriginalRune := normalizeTextForSearch(fullText)
	if normalizedText == "" {
		return ""
	}

	bestOriginalRune := -1
	for _, term := range terms {
		matchByteIndex := strings.Index(normalizedText, term)
		if matchByteIndex < 0 {
			continue
		}

		normalizedRuneIndex := utf8.RuneCountInString(normalizedText[:matchByteIndex])
		if normalizedRuneIndex >= 0 && normalizedRuneIndex < len(normalizedRuneToOriginalRune) {
			originalRuneIndex := normalizedRuneToOriginalRune[normalizedRuneIndex]
			if bestOriginalRune == -1 || originalRuneIndex < bestOriginalRune {
				bestOriginalRune = originalRuneIndex
			}
		}
	}

	if bestOriginalRune == -1 {
		return ""
	}

	originalRunes := []rune(fullText)
	start := bestOriginalRune - 18
	if start < 0 {
		start = 0
	}
	end := bestOriginalRune + 80
	if end > len(originalRunes) {
		end = len(originalRunes)
	}

	return string(originalRunes[start:end])
}

func normalizeTextForSearch(value string) (string, []int) {
	var builder strings.Builder
	runeMap := make([]int, 0, len(value))

	for originalRuneIndex, currentRune := range []rune(value) {
		normalizedRune := normalizeSearchRune(currentRune)
		builder.WriteRune(normalizedRune)
		runeMap = append(runeMap, originalRuneIndex)
	}

	return builder.String(), runeMap
}

func buildPrefixSearchQuery(query string) string {
	terms := extractNormalizedSearchTerms(query)
	for index, term := range terms {
		terms[index] = term + ":*"
	}

	return strings.Join(terms, " & ")
}

func extractNormalizedSearchTerms(query string) []string {
	terms := make([]string, 0)
	var builder strings.Builder

	flush := func() {
		if builder.Len() == 0 {
			return
		}
		term := normalizeSearchTerm(builder.String())
		terms = append(terms, term)
		builder.Reset()
	}

	for _, value := range query {
		if unicode.IsLetter(value) || unicode.IsNumber(value) {
			builder.WriteRune(value)
			continue
		}
		flush()
	}
	flush()

	return terms
}

func normalizeSearchTerm(value string) string {
	var builder strings.Builder
	for _, currentRune := range value {
		builder.WriteRune(normalizeSearchRune(currentRune))
	}
	return builder.String()
}

func normalizeSearchRune(value rune) rune {
	switch value {
	case 'ç', 'Ç':
		return 'c'
	case 'ğ', 'Ğ':
		return 'g'
	case 'ı', 'I', 'İ', 'i':
		return 'i'
	case 'ö', 'Ö':
		return 'o'
	case 'ş', 'Ş':
		return 's'
	case 'ü', 'Ü':
		return 'u'
	case 'â', 'Â':
		return 'a'
	case 'î', 'Î':
		return 'i'
	case 'û', 'Û':
		return 'u'
	default:
		return unicode.ToLower(value)
	}
}

func (s *Service) Update(ctx context.Context, id string, input UpdateDocumentRequest) (*DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return nil, err
	}
	if err := s.requireDocumentUnlocked(document); err != nil {
		return nil, err
	}

	input.OriginalFileName = strings.TrimSpace(input.OriginalFileName)
	if input.OriginalFileName == "" {
		return nil, ErrInvalidUpdateData
	}

	if input.Status != "" && !isValidStatus(input.Status) {
		return nil, ErrInvalidUpdateData
	}
	if input.Status != "" {
		document.Status = NormalizeStatus(input.Status)
	}
	document.OriginalFileName = input.OriginalFileName

	if input.FolderID != nil {
		folderRaw := strings.TrimSpace(*input.FolderID)
		if folderRaw == "" || folderRaw == "root" {
			document.FolderID = nil
		} else {
			folderID, err := s.parseFolderID(ctx, authContext.TenantID, folderRaw)
			if err != nil {
				if errors.Is(err, ErrInvalidFolderID) {
					return nil, ErrInvalidFolderID
				}
				if errors.Is(err, ErrFolderNotFound) {
					return nil, ErrFolderNotFound
				}
				return nil, err
			}
			document.FolderID = folderID
		}
	}

	if err := s.repository.Update(ctx, document); err != nil {
		return nil, err
	}

	return s.toResponsePtr(ctx, authContext.TenantID, *document)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDocumentNotFound
		}
		return err
	}
	if !permissions.CanDeleteDocument(permissions.SubjectFromAuth(authContext), permissions.DocumentFromFields(document.TenantID, document.UploadedByUserID, "")) {
		return ErrForbidden
	}
	if err := s.requireDocumentUnlocked(document); err != nil {
		return err
	}

	if err := s.repository.Delete(ctx, document); err != nil {
		return err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewDocumentDeleted(authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName, false))

	return nil
}

func (s *Service) Restore(ctx context.Context, id string) (*DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetDeletedByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if !permissions.CanDeleteDocument(permissions.SubjectFromAuth(authContext), permissions.DocumentFromFields(document.TenantID, document.UploadedByUserID, "")) {
		return nil, ErrForbidden
	}

	if err := s.repository.Restore(ctx, document); err != nil {
		return nil, err
	}

	document.DeletedAt = gorm.DeletedAt{}
	s.publisher.PublishAfterCommit(ctx, domain.NewDocumentRestored(authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName))

	return s.toResponsePtr(ctx, authContext.TenantID, *document)
}

func (s *Service) PermanentlyDelete(ctx context.Context, id string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	documentID, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidDocumentID
	}

	document, err := s.repository.GetDeletedByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDocumentNotFound
		}
		return err
	}
	if !permissions.CanDeleteDocument(permissions.SubjectFromAuth(authContext), permissions.DocumentFromFields(document.TenantID, document.UploadedByUserID, "")) {
		return ErrForbidden
	}

	_ = s.repository.DeleteTextByDocumentID(ctx, document.ID)
	_ = s.repository.DeleteChunksByDocumentID(ctx, document.ID)

	if err := s.repository.PermanentlyDelete(ctx, document); err != nil {
		return err
	}

	_ = s.deleteDocumentStorage(ctx, document)

	s.publisher.PublishAfterCommit(ctx, domain.NewDocumentDeleted(authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName, true))

	return nil
}

func (s *Service) SetDocumentTags(ctx context.Context, documentIDRaw string, tagIDStrings []string) (*DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	if !canAssignDocumentTags(authContext, document) {
		return nil, ErrForbidden
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	if s.tags == nil {
		return s.GetByID(ctx, documentIDRaw)
	}

	tagIDs, err := s.tags.ParseTagIDs(tagIDStrings)
	if err != nil {
		return nil, ErrInvalidTagID
	}
	if err := s.tags.ReplaceDocumentTags(ctx, authContext.TenantID, document.ID, tagIDs); err != nil {
		if errors.Is(err, tags.ErrTagNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewTagAssigned(authContext.TenantID, authContext.UserID, document.ID, tagIDs))

	return s.GetByID(ctx, documentIDRaw)
}

func (s *Service) ToResponses(ctx context.Context, tenantID uuid.UUID, items []Document) ([]DocumentResponse, error) {
	return s.mapDocumentsToResponses(ctx, tenantID, items)
}

func (s *Service) mapDocumentsToResponses(ctx context.Context, tenantID uuid.UUID, items []Document) ([]DocumentResponse, error) {
	responses := make([]DocumentResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toDocumentResponse(item))
	}
	return s.attachTags(ctx, tenantID, responses)
}

func (s *Service) attachTags(ctx context.Context, tenantID uuid.UUID, responses []DocumentResponse) ([]DocumentResponse, error) {
	if s.tags == nil || len(responses) == 0 {
		return responses, nil
	}

	documentIDs := make([]uuid.UUID, 0, len(responses))
	for _, response := range responses {
		documentIDs = append(documentIDs, response.ID)
	}

	tagMap, err := s.tags.MapTagsForDocuments(ctx, tenantID, documentIDs)
	if err != nil {
		return nil, err
	}

	for index := range responses {
		tagsForDocument := tagMap[responses[index].ID]
		if tagsForDocument == nil {
			tagsForDocument = []tags.TagBrief{}
		}
		responses[index].Tags = tagsForDocument
	}

	return responses, nil
}

func (s *Service) toResponsePtr(ctx context.Context, tenantID uuid.UUID, document Document) (*DocumentResponse, error) {
	responses, err := s.attachTags(ctx, tenantID, []DocumentResponse{toDocumentResponse(document)})
	if err != nil {
		return nil, err
	}
	if len(responses) == 0 {
		response := toDocumentResponse(document)
		return &response, nil
	}
	return &responses[0], nil
}

func (s *Service) memberAccessUserID(authContext auth.Context) *uuid.UUID {
	if HasTenantWideDocumentAccess(authContext.Role) {
		return nil
	}
	userID := authContext.UserID
	return &userID
}

func (s *Service) requireViewAccess(ctx context.Context, authContext auth.Context, document *Document) error {
	if s.shareAccess == nil {
		return nil
	}
	allowed, err := s.shareAccess.CanView(ctx, authContext, document)
	if err != nil {
		return err
	}
	if !allowed {
		applog.PermissionDenied("view_document", document.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
		return ErrForbidden
	}
	return nil
}

func (s *Service) requireDownloadAccess(ctx context.Context, authContext auth.Context, document *Document) error {
	if s.shareAccess == nil {
		return nil
	}
	allowed, err := s.shareAccess.CanDownload(ctx, authContext, document)
	if err != nil {
		return err
	}
	if !allowed {
		applog.PermissionDenied("download_document", document.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
		return ErrForbidden
	}
	return nil
}

func (s *Service) requireEditAccess(ctx context.Context, authContext auth.Context, document *Document) error {
	if s.shareAccess == nil {
		return nil
	}
	allowed, err := s.shareAccess.CanEdit(ctx, authContext, document)
	if err != nil {
		return err
	}
	if !allowed {
		applog.PermissionDenied("edit_document", document.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
		return ErrForbidden
	}
	return nil
}

func (s *Service) requireViewAccessForFavorite(ctx context.Context, authContext auth.Context, document *Document) error {
	if s.shareAccess != nil {
		allowed, err := s.shareAccess.CanView(ctx, authContext, document)
		if err != nil {
			return err
		}
		if !allowed {
			return ErrForbidden
		}
		return nil
	}
	if !legacyCanFavoriteDocument(authContext, document) {
		return ErrForbidden
	}
	return nil
}

func legacyCanFavoriteDocument(authContext auth.Context, document *Document) bool {
	switch authContext.Role {
	case auth.TenantRoleOwner, auth.TenantRoleAdmin:
		return true
	case auth.TenantRoleMember:
		return document.TenantID == authContext.TenantID
	default:
		return false
	}
}

func canAssignDocumentTags(authContext auth.Context, document *Document) bool {
	return document.TenantID == authContext.TenantID
}

func (s *Service) ProcessCompressDocumentJob(ctx context.Context, payload jobs.CompressDocumentPayload) error {
	if s.repository == nil {
		return fmt.Errorf("document repository is not configured")
	}

	document, err := s.repository.GetByID(ctx, payload.TenantID, payload.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("document not found: %s", payload.DocumentID)
		}
		return err
	}

	compressor := compression.NewGzipCompressor()
	if !compressor.ShouldCompress(document.FileExtension) {
		return nil
	}

	location, err := documentStorageLocation(s.storage, document)
	if err != nil {
		applog.Warn("document compression location failed: document_id=%s error=%v", document.ID, err)
		return nil
	}
	if err := s.storage.CompressLocalFile(ctx, location, compressor); err != nil {
		applog.Warn("document compression failed: document_id=%s error=%v", document.ID, err)
		return nil
	}

	applog.Info("document compressed: document_id=%s key=%s", document.ID, storage.CompressedKey(location.Key))
	return nil
}

func normalizeExtension(filename string) string {
	extension := strings.TrimSpace(strings.ToLower(filename))
	extension = filepath.Ext(extension)
	if extension == "" {
		extension = strings.TrimSpace(strings.ToLower(filename))
	}
	if extension == "" {
		return ""
	}
	return "." + strings.TrimPrefix(extension, ".")
}

func isAllowedExtension(extension string) bool {
	switch extension {
	case ".pdf", ".png", ".jpg", ".jpeg", ".docx", ".xlsx":
		return true
	default:
		return false
	}
}

func isValidStatus(status DocumentStatus) bool {
	return isValidLifecycleStatus(status) || status == DocumentStatusUploaded || status == DocumentStatusProcessed
}

func detectMimeType(fileHeader *multipart.FileHeader) string {
	if value := fileHeader.Header.Get("Content-Type"); value != "" {
		return value
	}
	return "application/octet-stream"
}

func (s *Service) StorageManager() *storage.Manager {
	return s.storage
}

func (s *Service) extractAndSaveDocumentText(ctx context.Context, document *Document) error {
	location, err := documentStorageLocation(s.storage, document)
	if err != nil {
		return err
	}

	extension := normalizeExtension("." + document.FileExtension)
	var fullText string
	err = s.storage.WithLocalFile(ctx, location, extension, func(filePath string) error {
		extracted, extractErr := extractDocumentText(document.FileExtension, filePath)
		if extractErr != nil {
			return extractErr
		}
		fullText = extracted
		return nil
	})
	if err != nil {
		log.Printf("document text extraction failed: document_id=%s extension=%s error=%v", document.ID, document.FileExtension, err)
		return err
	}

	if strings.TrimSpace(fullText) == "" {
		log.Printf("document text extraction returned empty text: document_id=%s extension=%s key=%s", document.ID, document.FileExtension, location.Key)
		_ = s.repository.DeleteChunksByDocumentID(ctx, document.ID)
		return nil
	}

	if err := s.saveDocumentTextAndChunks(ctx, document, fullText); err != nil {
		return err
	}

	document.Text = &DocumentText{
		DocumentID: document.ID,
		FullText:   fullText,
	}
	return nil
}

func (s *Service) saveDocumentTextAndChunks(ctx context.Context, document *Document, fullText string) error {
	fullText = strings.TrimSpace(fullText)
	if fullText == "" {
		_ = s.repository.DeleteChunksByDocumentID(ctx, document.ID)
		return nil
	}

	documentText := &DocumentText{
		DocumentID: document.ID,
		FullText:   fullText,
	}
	chunks := buildDocumentChunks(document.ID, fullText)
	chunks = s.embedChunks(ctx, chunks)

	embeddedCount := countEmbeddedChunks(chunks)
	if err := s.repository.SaveTextChunksAndStatus(ctx, document.TenantID, document.ID, DocumentStatusReady, documentText, chunks); err != nil {
		log.Printf(
			"embedding save fail: document_id=%s chunk_count=%d embedded_count=%d error=%v",
			document.ID,
			len(chunks),
			embeddedCount,
			err,
		)
		return err
	}

	if len(chunks) > 0 {
		log.Printf(
			"embedding save success: document_id=%s chunk_count=%d embedded_count=%d",
			document.ID,
			len(chunks),
			embeddedCount,
		)
	}

	document.Status = DocumentStatusReady
	return nil
}

func buildDocumentChunks(documentID uuid.UUID, fullText string) []DocumentChunk {
	chunkContents := chunkText(fullText)
	chunks := make([]DocumentChunk, 0, len(chunkContents))
	for _, content := range chunkContents {
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}
		chunks = append(chunks, DocumentChunk{
			ID:            uuid.New(),
			DocumentID:    documentID,
			ChunkIndex:    len(chunks),
			Content:       content,
			TokenEstimate: estimateTokens(content),
		})
	}
	return chunks
}

func (s *Service) embedChunks(ctx context.Context, chunks []DocumentChunk) []DocumentChunk {
	if s.embeddingProvider == nil {
		log.Printf("embedding skipped: provider=disabled chunk_count=%d", len(chunks))
		return chunks
	}

	providerName := s.embeddingProvider.ProviderName()
	modelName := s.embeddingProvider.ModelName()

	for index := range chunks {
		chunk := &chunks[index]
		embeddingValues, err := s.embeddingProvider.CreateEmbedding(ctx, chunk.Content)
		if err != nil {
			log.Printf(
				"embedding fail: provider=%s model=%s document_id=%s chunk_id=%s chunk_index=%d error=%v",
				providerName,
				modelName,
				chunk.DocumentID,
				chunk.ID,
				chunk.ChunkIndex,
				err,
			)
			continue
		}

		dimension := len(embeddingValues)
		if dimension != embedding.DefaultEmbeddingDimensions {
			log.Printf(
				"embedding fail: provider=%s model=%s document_id=%s chunk_id=%s chunk_index=%d dimension=%d expected=%d error=vector dimension mismatch",
				providerName,
				modelName,
				chunk.DocumentID,
				chunk.ID,
				chunk.ChunkIndex,
				dimension,
				embedding.DefaultEmbeddingDimensions,
			)
			continue
		}

		vector := pgvector.NewVector(embeddingValues)
		chunk.Embedding = &vector
		log.Printf(
			"embedding success: provider=%s model=%s document_id=%s chunk_id=%s chunk_index=%d dimension=%d",
			providerName,
			modelName,
			chunk.DocumentID,
			chunk.ID,
			chunk.ChunkIndex,
			dimension,
		)
	}

	return chunks
}

func countEmbeddedChunks(chunks []DocumentChunk) int {
	count := 0
	for _, chunk := range chunks {
		if chunk.Embedding != nil {
			count++
		}
	}
	return count
}

func isTextExtractionSupported(extension string) bool {
	extension = normalizeStoredExtension(extension)
	return extension == "pdf" || extension == "docx" || extension == "xlsx" || extension == "txt" || isImageExtension(extension)
}

func extractDocumentText(extension, filePath string) (string, error) {
	switch normalizeStoredExtension(extension) {
	case "pdf":
		return extractPDFText(filePath)
	case "docx":
		return extractDOCXText(filePath)
	case "xlsx":
		return extractXLSXText(filePath)
	case "txt":
		return extractTXTText(filePath)
	case "png", "jpg", "jpeg":
		return extractImageText(filePath)
	default:
		return "", ErrTextExtractionUnsupported
	}
}

func normalizeStoredExtension(extension string) string {
	return strings.TrimPrefix(normalizeExtension(extension), ".")
}
