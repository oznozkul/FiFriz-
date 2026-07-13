package documents

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DuplicateDocumentBrief struct {
	ID               uuid.UUID `json:"id"`
	OriginalFileName string    `json:"originalFileName"`
}

type DocumentHealthResponse struct {
	PreviewGenerated   bool   `json:"previewGenerated"`
	OCRGenerated       bool   `json:"ocrGenerated"`
	EmbeddingGenerated bool   `json:"embeddingGenerated"`
	SummaryGenerated   bool   `json:"summaryGenerated"`
	MetadataGenerated  bool   `json:"metadataGenerated"`
	AICompleted        bool   `json:"aiCompleted"`
	VersionCount       int    `json:"versionCount"`
	StorageStatus      string `json:"storageStatus"`
	ProcessingStatus   string `json:"processingStatus"`
}

type DocumentUploadResult struct {
	Document     DocumentResponse        `json:"document"`
	Warnings     []string              `json:"warnings,omitempty"`
	DuplicateOf  *DuplicateDocumentBrief `json:"duplicateOf,omitempty"`
}

func (r *Repository) FindActiveByChecksum(ctx context.Context, tenantID uuid.UUID, checksum string) (*Document, error) {
	checksum = strings.TrimSpace(checksum)
	if checksum == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var document Document
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND sha256_checksum = ? AND deleted_at IS NULL", tenantID, checksum).
		Order("created_at DESC").
		First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *Repository) CountVersions(ctx context.Context, tenantID, documentID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&DocumentVersion{}).
		Where("tenant_id = ? AND document_id = ?", tenantID, documentID).
		Count(&count).Error
	return int(count), err
}

func (r *Repository) HasEmbeddings(ctx context.Context, documentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&DocumentChunk{}).
		Where("document_id = ? AND embedding IS NOT NULL", documentID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) HasSummary(ctx context.Context, tenantID, documentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&DocumentSummary{}).
		Where("tenant_id = ? AND document_id = ?", tenantID, documentID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) HasAIMetadata(ctx context.Context, tenantID, documentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&DocumentAIMetadata{}).
		Where("tenant_id = ? AND document_id = ?", tenantID, documentID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) ListTenantDocumentPolicies(ctx context.Context) ([]TenantDocumentPolicy, error) {
	var items []TenantDocumentPolicy
	err := r.db.WithContext(ctx).Find(&items).Error
	return items, err
}

func ParseTenantPolicyMIMEList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err == nil {
		return values
	}
	parts := strings.Split(raw, ",")
	for index, part := range parts {
		parts[index] = strings.TrimSpace(part)
	}
	return parts
}
