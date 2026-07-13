package documents

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"strings"

	"fifriz/internal/applog"
	"fifriz/internal/auth"
	domain "fifriz/internal/events/events"
	"fifriz/internal/jobs"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *Service) UploadVersion(ctx context.Context, documentIDRaw string, prepared *Document, changeNoteRaw string, reader io.Reader, size int64) (*DocumentVersionResponse, error) {
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
	if err := s.requireUploadVersionAccess(ctx, authContext, document); err != nil {
		return nil, err
	}
	if err := s.requireDocumentUnlocked(document); err != nil {
		return nil, err
	}

	maxVersion, err := s.repository.GetMaxVersionNumber(ctx, authContext.TenantID, documentID)
	if err != nil {
		return nil, err
	}

	var changeNote *string
	if trimmed := strings.TrimSpace(changeNoteRaw); trimmed != "" {
		changeNote = &trimmed
	}

	extension := normalizeExtension(prepared.OriginalFileName)
	validated, err := s.validateUploadContent(ctx, authContext.TenantID, &multipart.FileHeader{
		Filename: prepared.OriginalFileName,
		Size:     size,
	}, reader, extension)
	if err != nil {
		return nil, err
	}
	prepared.MimeType = validated.mimeType
	prepared.SHA256Checksum = validated.checksum
	prepared.FileSize = int64(len(validated.data))

	if err := s.storeDocumentFile(ctx, authContext.TenantID, document.ID, prepared, bytes.NewReader(validated.data), int64(len(validated.data)), maxVersion+1, nil); err != nil {
		return nil, err
	}

	version := versionFromDocument(prepared, maxVersion+1, true, changeNote, authContext.UserID)
	version.TenantID = authContext.TenantID
	version.DocumentID = document.ID

	applyVersionToDocument(document, version)
	document.Status = DocumentStatusUploading

	if err := s.repository.CreateVersionAndUpdateDocument(ctx, document, version); err != nil {
		return nil, err
	}

	document.Status = DocumentStatusProcessing
	if err := s.repository.Update(ctx, document); err != nil {
		return nil, err
	}
	s.scheduleDocumentProcessing(ctx, authContext.TenantID, document.ID, authContext.UserID, jobs.TriggerVersion)

	s.publisher.PublishAfterCommit(ctx, domain.NewDocumentVersionCreated(
		authContext.TenantID,
		authContext.UserID,
		document.ID,
		version.ID,
		document.UploadedByUserID,
		version.VersionNumber,
		document.OriginalFileName,
	))

	if reloaded, err := s.repository.ListVersionsByDocument(ctx, authContext.TenantID, documentID); err == nil {
		for _, item := range reloaded {
			if item.ID == version.ID {
				response := toDocumentVersionResponse(item)
				return &response, nil
			}
		}
	}

	response := toDocumentVersionResponse(VersionListRow{DocumentVersion: *version})
	return &response, nil
}

var (
	ErrDocumentVersionNotFound = errors.New("document version not found")
	ErrInvalidVersionID        = errors.New("invalid version id")
)

func (s *Service) ListVersions(ctx context.Context, documentIDRaw string) ([]DocumentVersionResponse, error) {
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
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	rows, err := s.repository.ListVersionsByDocument(ctx, authContext.TenantID, documentID)
	if err != nil {
		return nil, err
	}

	responses := make([]DocumentVersionResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, toDocumentVersionResponse(row))
	}

	return responses, nil
}

func (s *Service) RestoreVersion(ctx context.Context, documentIDRaw, versionIDRaw string) (*DocumentResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	versionID, err := uuid.Parse(versionIDRaw)
	if err != nil {
		return nil, ErrInvalidVersionID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireRestoreVersionAccess(ctx, authContext, document); err != nil {
		return nil, err
	}
	if err := s.requireDocumentUnlocked(document); err != nil {
		return nil, err
	}

	version, err := s.repository.GetVersionByID(ctx, authContext.TenantID, documentID, versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentVersionNotFound
		}
		return nil, err
	}
	if version.IsCurrent {
		return s.toResponsePtr(ctx, authContext.TenantID, *document)
	}

	applyVersionToDocument(document, version)
	document.Status = DocumentStatusUploading

	if err := s.repository.RestoreVersionAndUpdateDocument(ctx, document, version); err != nil {
		return nil, err
	}

	document.Status = DocumentStatusProcessing
	if err := s.repository.Update(ctx, document); err != nil {
		return nil, err
	}
	s.scheduleDocumentProcessing(ctx, authContext.TenantID, document.ID, authContext.UserID, jobs.TriggerRestore)

	s.publisher.PublishAfterCommit(ctx, domain.NewDocumentVersionRestored(
		authContext.TenantID,
		authContext.UserID,
		document.ID,
		version.ID,
		document.UploadedByUserID,
		version.VersionNumber,
		document.OriginalFileName,
	))

	return s.toResponsePtr(ctx, authContext.TenantID, *document)
}

func (s *Service) requireUploadVersionAccess(ctx context.Context, authContext auth.Context, document *Document) error {
	if s.shareAccess == nil {
		return nil
	}
	allowed, err := s.shareAccess.CanUploadVersion(ctx, authContext, document)
	if err != nil {
		return err
	}
	if !allowed {
		applog.PermissionDenied("upload_document_version", document.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
		return ErrForbidden
	}
	return nil
}

func (s *Service) requireRestoreVersionAccess(ctx context.Context, authContext auth.Context, document *Document) error {
	if s.shareAccess == nil {
		return nil
	}
	allowed, err := s.shareAccess.CanRestoreVersion(ctx, authContext, document)
	if err != nil {
		return err
	}
	if !allowed {
		applog.PermissionDenied("restore_document_version", document.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
		return ErrForbidden
	}
	return nil
}
