package documents

import (
	"context"
	"fmt"
	"io"

	"fifriz/internal/storage"

	"github.com/google/uuid"
)

func documentStorageLocation(manager *storage.Manager, document *Document) (storage.DocumentLocation, error) {
	if manager == nil {
		return storage.DocumentLocation{}, fmt.Errorf("storage manager is not configured")
	}
	return manager.ResolveLocation(
		document.TenantID,
		document.StorageProvider,
		document.StorageKey,
		document.FilePath,
		document.BucketName,
		document.StoredFileName,
	)
}

func applyStorageLocation(document *Document, location storage.DocumentLocation) {
	document.StorageProvider = location.Provider
	document.StorageKey = location.Key
	document.BucketName = location.BucketName
	document.FilePath = location.FilePath
}

func versionStorageLocation(manager *storage.Manager, version *DocumentVersion) (storage.DocumentLocation, error) {
	if manager == nil {
		return storage.DocumentLocation{}, fmt.Errorf("storage manager is not configured")
	}
	return manager.ResolveLocation(
		version.TenantID,
		version.StorageProvider,
		version.StorageKey,
		version.FilePath,
		version.BucketName,
		version.StoredFileName,
	)
}

func applyStorageLocationToVersion(version *DocumentVersion, location storage.DocumentLocation) {
	version.StorageProvider = location.Provider
	version.StorageKey = location.Key
	version.BucketName = location.BucketName
	version.FilePath = location.FilePath
}

func (s *Service) storeDocumentFile(
	ctx context.Context,
	tenantID, documentID uuid.UUID,
	document *Document,
	reader io.Reader,
	size int64,
	versionNumber int,
	integrationAccountID *uuid.UUID,
) error {
	location := s.storage.NewUploadLocation(tenantID, documentID, document.StoredFileName, integrationAccountID, versionNumber)
	if err := s.storage.Upload(ctx, location, reader, size, document.MimeType); err != nil {
		return err
	}
	applyStorageLocation(document, location)
	return nil
}

func (s *Service) deleteDocumentStorage(ctx context.Context, document *Document) error {
	location, err := documentStorageLocation(s.storage, document)
	if err != nil {
		return err
	}
	return s.storage.Delete(ctx, location)
}
