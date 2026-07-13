package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	Provider      string
	LocalBasePath string
	MinIO         MinIOConfig
	ProcessingDir string
}

type DocumentLocation struct {
	Provider   string
	Key        string
	BucketName string
	FilePath   string
}

type Manager struct {
	defaultProvider string
	local           *LocalStorageProvider
	minio           *MinIOStorageProvider
	processingDir   string
}

func NewManager(cfg Config) (*Manager, error) {
	local := NewLocalStorageProvider(cfg.LocalBasePath)
	processingDir := cfg.ProcessingDir
	if processingDir == "" {
		processingDir = filepath.Join("tmp", "processing")
	}
	if err := os.MkdirAll(processingDir, 0o755); err != nil {
		return nil, fmt.Errorf("create processing dir: %w", err)
	}

	manager := &Manager{
		defaultProvider: normalizeProvider(cfg.Provider),
		local:           local,
		processingDir:   processingDir,
	}

	if manager.defaultProvider == ProviderMinIO {
		minioProvider, err := NewMinIOStorageProvider(cfg.MinIO)
		if err != nil {
			return nil, err
		}
		manager.minio = minioProvider
	}

	return manager, nil
}

func normalizeProvider(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ProviderMinIO:
		return ProviderMinIO
	default:
		return ProviderLocal
	}
}

func (m *Manager) DefaultProviderName() string {
	return m.defaultProvider
}

func (m *Manager) BucketName() string {
	if m.minio != nil {
		return m.minio.Bucket()
	}
	return ""
}

func (m *Manager) LocalProvider() *LocalStorageProvider {
	return m.local
}

func (m *Manager) MinIOProvider() *MinIOStorageProvider {
	return m.minio
}

func (m *Manager) providerByName(name string) (StorageProvider, error) {
	switch normalizeProvider(name) {
	case ProviderMinIO:
		if m.minio == nil {
			return nil, errors.New("minio storage provider is not configured")
		}
		return m.minio, nil
	default:
		return m.local, nil
	}
}

func (m *Manager) NewUploadLocation(
	tenantID, documentID uuid.UUID,
	storedFileName string,
	integrationAccountID *uuid.UUID,
	versionNumber int,
) DocumentLocation {
	if m.defaultProvider == ProviderMinIO {
		var key string
		switch {
		case integrationAccountID != nil:
			key = IntegrationDocumentKey(tenantID, *integrationAccountID, documentID, storedFileName)
		case versionNumber > 0:
			key = VersionDocumentKey(tenantID, documentID, versionNumber, storedFileName)
		default:
			key = OriginalDocumentKey(tenantID, documentID, storedFileName)
		}
		return DocumentLocation{
			Provider:   ProviderMinIO,
			Key:        key,
			BucketName: m.BucketName(),
		}
	}

	return DocumentLocation{
		Provider: ProviderLocal,
		Key:      storedFileName,
		FilePath: filepath.ToSlash(filepath.Join(m.local.BasePath(), storedFileName)),
	}
}

func (m *Manager) ResolveLocation(
	tenantID uuid.UUID,
	providerName, storageKey, filePath, bucketName, storedFileName string,
) (DocumentLocation, error) {
	providerName = normalizeProvider(providerName)
	if providerName == "" || providerName == ProviderLocal {
		if storageKey != "" {
			return DocumentLocation{
				Provider: ProviderLocal,
				Key:      storageKey,
				FilePath: filePath,
			}, nil
		}
		if filePath != "" {
			return DocumentLocation{
				Provider: ProviderLocal,
				Key:      storedFileName,
				FilePath: filePath,
			}, nil
		}
		if storedFileName != "" {
			return DocumentLocation{
				Provider: ProviderLocal,
				Key:      storedFileName,
				FilePath: filepath.ToSlash(filepath.Join(m.local.BasePath(), storedFileName)),
			}, nil
		}
		return DocumentLocation{}, errors.New("document storage location is missing")
	}

	if storageKey == "" {
		return DocumentLocation{}, errors.New("minio storage key is missing")
	}
	if tenantID != uuid.Nil && !ValidateTenantKey(tenantID, storageKey) {
		return DocumentLocation{}, errors.New("storage key does not belong to tenant")
	}
	if bucketName == "" {
		bucketName = m.BucketName()
	}
	return DocumentLocation{
		Provider:   ProviderMinIO,
		Key:        storageKey,
		BucketName: bucketName,
	}, nil
}

func (m *Manager) Upload(ctx context.Context, location DocumentLocation, reader io.Reader, size int64, contentType string) error {
	provider, err := m.providerForLocation(location)
	if err != nil {
		return err
	}
	_, err = provider.Upload(ctx, UploadInput{
		Key:         location.Key,
		Reader:      reader,
		Size:        size,
		ContentType: contentType,
	})
	return err
}

func (m *Manager) providerForLocation(location DocumentLocation) (StorageProvider, error) {
	return m.providerByName(location.Provider)
}

func (m *Manager) Download(ctx context.Context, location DocumentLocation) (io.ReadCloser, error) {
	if location.Provider == ProviderLocal && location.FilePath != "" {
		if _, err := os.Stat(location.FilePath); err == nil {
			return os.Open(location.FilePath)
		}
	}

	provider, err := m.providerForLocation(location)
	if err != nil {
		return nil, err
	}
	return provider.Download(ctx, location.Key)
}

func (m *Manager) Delete(ctx context.Context, location DocumentLocation) error {
	provider, err := m.providerForLocation(location)
	if err != nil {
		return err
	}
	return provider.Delete(ctx, location.Key)
}

func (m *Manager) Exists(ctx context.Context, location DocumentLocation) (bool, error) {
	provider, err := m.providerForLocation(location)
	if err != nil {
		return false, err
	}
	return provider.Exists(ctx, location.Key)
}

func (m *Manager) WithLocalFile(ctx context.Context, location DocumentLocation, extension string, fn func(path string) error) error {
	if location.Provider == ProviderLocal {
		path := location.FilePath
		if path == "" {
			path = m.local.LocalPath(location.Key)
		}
		if _, err := os.Stat(path); err != nil {
			return err
		}
		return fn(path)
	}

	reader, err := m.Download(ctx, location)
	if err != nil {
		return err
	}
	defer reader.Close()

	tempFile, err := os.CreateTemp(m.processingDir, "doc-*"+extension)
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	if _, err := io.Copy(tempFile, reader); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	return fn(tempPath)
}

func (m *Manager) UploadCompressed(ctx context.Context, location DocumentLocation, sourcePath string) error {
	compressedKey := CompressedKey(location.Key)
	provider, err := m.providerForLocation(location)
	if err != nil {
		return err
	}

	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	_, err = provider.Upload(ctx, UploadInput{
		Key:    compressedKey,
		Reader: file,
		Size:   info.Size(),
	})
	return err
}

func (m *Manager) CompressLocalFile(ctx context.Context, location DocumentLocation, compressor interface {
	Compress(sourcePath, destinationPath string) error
}) error {
	return m.WithLocalFile(ctx, location, "", func(sourcePath string) error {
		destPath := sourcePath + ".gz"
		if err := compressor.Compress(sourcePath, destPath); err != nil {
			return err
		}
		defer os.Remove(destPath)

		if location.Provider == ProviderMinIO {
			return m.UploadCompressed(ctx, location, destPath)
		}
		return nil
	})
}

func (m *Manager) GetPresignedURL(ctx context.Context, location DocumentLocation, expiry time.Duration) (string, error) {
	if location.Provider != ProviderMinIO {
		return "", errors.New("presigned URLs are only available for minio storage")
	}
	provider, err := m.providerForLocation(location)
	if err != nil {
		return "", err
	}
	return provider.GetPresignedURL(ctx, location.Key, expiry)
}

type MigrateDocumentInput struct {
	TenantID             uuid.UUID
	DocumentID           uuid.UUID
	StoredFileName       string
	CurrentLocation      DocumentLocation
	IntegrationAccountID *uuid.UUID
	VersionNumber        int
}

func (m *Manager) MigrateDocumentToMinIO(ctx context.Context, input MigrateDocumentInput) (DocumentLocation, error) {
	if m.minio == nil {
		return DocumentLocation{}, errors.New("minio storage provider is not configured")
	}
	if input.CurrentLocation.Provider != ProviderLocal {
		return DocumentLocation{}, errors.New("document is not stored locally")
	}

	target := m.NewUploadLocation(
		input.TenantID,
		input.DocumentID,
		input.StoredFileName,
		input.IntegrationAccountID,
		input.VersionNumber,
	)
	target.Provider = ProviderMinIO
	if input.VersionNumber > 0 {
		target.Key = VersionDocumentKey(input.TenantID, input.DocumentID, input.VersionNumber, input.StoredFileName)
	} else if input.IntegrationAccountID != nil {
		target.Key = IntegrationDocumentKey(input.TenantID, *input.IntegrationAccountID, input.DocumentID, input.StoredFileName)
	} else {
		target.Key = OriginalDocumentKey(input.TenantID, input.DocumentID, input.StoredFileName)
	}
	target.BucketName = m.BucketName()

	reader, err := m.local.Download(ctx, input.CurrentLocation.Key)
	if err != nil {
		if input.CurrentLocation.FilePath != "" {
			file, openErr := os.Open(input.CurrentLocation.FilePath)
			if openErr != nil {
				return DocumentLocation{}, err
			}
			reader = file
		} else {
			return DocumentLocation{}, err
		}
	}
	defer reader.Close()

	tempFile, err := os.CreateTemp(m.processingDir, "migrate-*")
	if err != nil {
		return DocumentLocation{}, err
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	size, err := io.Copy(tempFile, reader)
	if err != nil {
		return DocumentLocation{}, err
	}
	if err := tempFile.Close(); err != nil {
		return DocumentLocation{}, err
	}

	file, err := os.Open(tempPath)
	if err != nil {
		return DocumentLocation{}, err
	}
	defer file.Close()

	if _, err := m.minio.Upload(ctx, UploadInput{
		Key:    target.Key,
		Reader: file,
		Size:   size,
	}); err != nil {
		return DocumentLocation{}, err
	}

	return target, nil
}
