package documents

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"time"

	docvalidation "fifriz/internal/documents/validation"
	"fifriz/internal/documents/virusscan"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EnterpriseConfig struct {
	PolicyResolver      *docvalidation.PolicyResolver
	VirusScanner        virusscan.Scanner
	DownloadMaxRetries  int
	DownloadTimeout     time.Duration
}

func (s *Service) SetEnterpriseConfig(cfg EnterpriseConfig) {
	s.filePolicy = cfg.PolicyResolver
	if cfg.VirusScanner != nil {
		s.virusScanner = cfg.VirusScanner
	}
	if cfg.DownloadMaxRetries > 0 {
		s.downloadMaxRetries = cfg.DownloadMaxRetries
	}
	if cfg.DownloadTimeout > 0 {
		s.downloadTimeout = cfg.DownloadTimeout
	}
}

type validatedUpload struct {
	data     []byte
	mimeType string
	checksum string
	warnings []string
	duplicate *DuplicateDocumentBrief
}

func (s *Service) validateUploadContent(ctx context.Context, tenantID uuid.UUID, fileHeader *multipart.FileHeader, reader io.Reader, extension string) (*validatedUpload, error) {
	if s.filePolicy == nil {
		return s.validateUploadContentLegacy(fileHeader, reader, extension)
	}

	policy := s.filePolicy.Resolve(tenantID)
	maxSize := policy.MaxFileSizeBytes
	if maxSize <= 0 {
		maxSize = 100 << 20
	}

	data, mimeType, err := docvalidation.ValidateReader(reader, fileHeader.Size, extension, maxSize)
	if err != nil {
		if strings.Contains(err.Error(), "maximum allowed size") {
			return nil, ErrFileTooLarge
		}
		if errors.Is(err, docvalidation.ErrMIMETypeMismatch) {
			return nil, ErrInvalidMIMEContent
		}
		return nil, err
	}
	if !policy.AllowsMIME(mimeType) {
		return nil, ErrMIMETypeNotAllowed
	}

	if s.virusScanner != nil {
		scanResult, scanErr := s.virusScanner.Scan(ctx, bytes.NewReader(data), fileHeader.Filename)
		if scanErr != nil {
			return nil, scanErr
		}
		if !scanResult.Clean {
			return nil, errors.New("upload rejected by virus scanner")
		}
	}

	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])

	result := &validatedUpload{
		data:     data,
		mimeType: mimeType,
		checksum: checksum,
	}

	if duplicate, dupErr := s.repository.FindActiveByChecksum(ctx, tenantID, checksum); dupErr == nil && duplicate != nil {
		result.warnings = append(result.warnings, "Aynı checksum ile başka bir belge zaten mevcut.")
		result.duplicate = &DuplicateDocumentBrief{
			ID:               duplicate.ID,
			OriginalFileName: duplicate.OriginalFileName,
		}
	} else if dupErr != nil && !errors.Is(dupErr, gorm.ErrRecordNotFound) {
		return nil, dupErr
	}

	return result, nil
}

func (s *Service) validateUploadContentLegacy(fileHeader *multipart.FileHeader, reader io.Reader, extension string) (*validatedUpload, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	mimeType, err := docvalidation.ValidateContent(data, extension)
	if err != nil {
		if errors.Is(err, docvalidation.ErrMIMETypeMismatch) {
			return nil, ErrInvalidMIMEContent
		}
		return nil, err
	}
	sum := sha256.Sum256(data)
	return &validatedUpload{
		data:     data,
		mimeType: mimeType,
		checksum: hex.EncodeToString(sum[:]),
	}, nil
}
