package documents

import (
	"encoding/json"
	"strings"
	"time"

	"fifriz/internal/tags"

	"github.com/google/uuid"
)

type DocumentResponse struct {
	ID               uuid.UUID      `json:"id"`
	TenantID         uuid.UUID      `json:"tenant_id"`
	FolderID         *uuid.UUID     `json:"folder_id"`
	FolderName       string         `json:"folder_name"`
	UploadedByUserID uuid.UUID      `json:"uploaded_by_user_id"`
	OriginalFileName string         `json:"original_file_name"`
	StoredFileName   string         `json:"stored_file_name"`
	FilePath         string         `json:"file_path"`
	FileURL          string         `json:"fileUrl"`
	FileExtension    string         `json:"file_extension"`
	MimeType         string         `json:"mime_type"`
	FileSize         int64          `json:"file_size"`
	Status           DocumentStatus `json:"status"`
	LifecycleStatus  string         `json:"lifecycleStatus"`
	ApprovalStatus   string         `json:"approvalStatus"`
	StorageProvider  string         `json:"storageProvider"`
	SHA256Checksum   string         `json:"sha256Checksum"`
	CompressedSize   *int64         `json:"compressedSize,omitempty"`
	ContentEncoding  string         `json:"contentEncoding"`
	StorageStatus    string         `json:"storageStatus"`
	IsLocked         bool           `json:"isLocked"`
	LockedAt         *time.Time     `json:"lockedAt,omitempty"`
	SourceType       string         `json:"sourceType"`
	SourceMetadata   map[string]any `json:"sourceMetadata"`
	HasExtractedText bool           `json:"hasExtractedText"`
	IsFavorite       bool           `json:"isFavorite"`
	FavoritedAt      *time.Time     `json:"favoritedAt"`
	AIDocumentType           *string    `json:"aiDocumentType,omitempty"`
	AIDocumentTypeConfidence *float64   `json:"aiDocumentTypeConfidence,omitempty"`
	AIClassifiedAt           *time.Time `json:"aiClassifiedAt,omitempty"`
	Tags             []tags.TagBrief `json:"tags"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type UpdateDocumentRequest struct {
	OriginalFileName string         `json:"original_file_name"`
	Status           DocumentStatus `json:"status"`
	FolderID         *string        `json:"folder_id"`
}

type DocumentTextResponse struct {
	DocumentID uuid.UUID `json:"document_id"`
	FullText   string    `json:"full_text"`
	CreatedAt  time.Time `json:"created_at"`
}

type DocumentSearchResultResponse struct {
	DocumentID       uuid.UUID `json:"documentId"`
	Title            string    `json:"title"`
	OriginalFileName string    `json:"originalFileName"`
	FileExtension    string    `json:"fileExtension"`
	MatchedSnippet   string    `json:"matchedSnippet"`
	CreatedAt        time.Time `json:"createdAt"`
}

type SemanticSearchResultResponse struct {
	DocumentID uuid.UUID `json:"documentId"`
	ChunkID    uuid.UUID `json:"chunkId"`
	ChunkIndex int       `json:"chunkIndex"`
	Content    string    `json:"content"`
	Similarity float64   `json:"similarity"`
}

type DocumentChunkResponse struct {
	ID            uuid.UUID `json:"id"`
	DocumentID    uuid.UUID `json:"documentId"`
	ChunkIndex    int       `json:"chunkIndex"`
	Content       string    `json:"content"`
	TokenEstimate int       `json:"tokenEstimate"`
	CreatedAt     time.Time `json:"createdAt"`
}

type DocumentChunkStatsResponse struct {
	DocumentID       uuid.UUID `json:"documentId"`
	TotalCharacters  int       `json:"totalCharacters"`
	TotalWords       int       `json:"totalWords"`
	ChunkCount       int       `json:"chunkCount"`
	AverageChunkSize int       `json:"averageChunkSize"`
	ChunkIndexes     []int     `json:"chunkIndexes"`
}

func toDocumentResponse(document Document) DocumentResponse {
	return DocumentResponse{
		ID:               document.ID,
		TenantID:         document.TenantID,
		FolderID:         document.FolderID,
		FolderName:       document.FolderName,
		UploadedByUserID: document.UploadedByUserID,
		OriginalFileName: document.OriginalFileName,
		StoredFileName:   document.StoredFileName,
		FilePath:         document.FilePath,
		FileURL:          "/api/v1/documents/" + document.ID.String() + "/file",
		FileExtension:    document.FileExtension,
		MimeType:         document.MimeType,
		FileSize:         document.FileSize,
		Status:           NormalizeStatus(document.Status),
		LifecycleStatus:  string(EffectiveLifecycleStatus(&document, false)),
		ApprovalStatus:   string(document.ApprovalStatus),
		StorageProvider:  document.StorageProvider,
		SHA256Checksum:   document.SHA256Checksum,
		CompressedSize:   document.CompressedSize,
		ContentEncoding:  document.ContentEncoding,
		StorageStatus:    document.StorageStatus,
		IsLocked:         document.IsLocked,
		LockedAt:         document.LockedAt,
		SourceType:       document.SourceType,
		SourceMetadata:   decodeDocumentSourceMetadata(document.SourceMetadata),
		HasExtractedText: hasExtractedText(document),
		IsFavorite:       document.IsFavorite,
		FavoritedAt:      document.FavoritedAt,
		AIDocumentType:           document.AIDocumentType,
		AIDocumentTypeConfidence: document.AIDocumentTypeConfidence,
		AIClassifiedAt:           document.AIClassifiedAt,
		Tags:             []tags.TagBrief{},
		CreatedAt:        document.CreatedAt,
		UpdatedAt:        document.UpdatedAt,
	}
}

func toDocumentTextResponse(documentText DocumentText) DocumentTextResponse {
	return DocumentTextResponse{
		DocumentID: documentText.DocumentID,
		FullText:   documentText.FullText,
		CreatedAt:  documentText.CreatedAt,
	}
}

func toDocumentChunkResponses(chunks []DocumentChunk) []DocumentChunkResponse {
	responses := make([]DocumentChunkResponse, 0, len(chunks))
	for _, chunk := range chunks {
		responses = append(responses, DocumentChunkResponse{
			ID:            chunk.ID,
			DocumentID:    chunk.DocumentID,
			ChunkIndex:    chunk.ChunkIndex,
			Content:       chunk.Content,
			TokenEstimate: chunk.TokenEstimate,
			CreatedAt:     chunk.CreatedAt,
		})
	}
	return responses
}

func hasExtractedText(document Document) bool {
	return document.Text != nil && strings.TrimSpace(document.Text.FullText) != ""
}

func decodeDocumentSourceMetadata(raw []byte) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil
	}
	return result
}
