package documents

import (
	"context"
	"errors"
	"strings"

	"fifriz/internal/ai/pipeline"
	"fifriz/internal/auth"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *Service) GetDocumentHealth(ctx context.Context, id string) (*DocumentHealthResponse, error) {
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

	hasText := hasExtractedText(*document)
	extension := strings.ToLower(strings.TrimSpace(document.FileExtension))
	isImage := extension == "png" || extension == "jpg" || extension == "jpeg"

	hasEmbeddings, _ := s.repository.HasEmbeddings(ctx, documentID)
	hasSummary, _ := s.repository.HasSummary(ctx, authContext.TenantID, documentID)
	hasMetadata, _ := s.repository.HasAIMetadata(ctx, authContext.TenantID, documentID)
	versionCount, _ := s.repository.CountVersions(ctx, authContext.TenantID, documentID)

	storageStatus := strings.TrimSpace(document.StorageStatus)
	if storageStatus == "" {
		storageStatus = "available"
	}

	aiCompleted := IsReadyStatus(document.Status)
	if progress, progressErr := s.GetAIProgress(ctx, id); progressErr == nil && len(progress) > 0 {
		aiCompleted = pipelineStepsCompleted(progress)
	}

	return &DocumentHealthResponse{
		PreviewGenerated:   isPreviewSupportedExtension(extension),
		OCRGenerated:       isImage && hasText,
		EmbeddingGenerated: hasEmbeddings,
		SummaryGenerated:   hasSummary,
		MetadataGenerated:  hasMetadata,
		AICompleted:        aiCompleted,
		VersionCount:       versionCount,
		StorageStatus:      storageStatus,
		ProcessingStatus:   string(EffectiveLifecycleStatus(document, false)),
	}, nil
}

func isPreviewSupportedExtension(extension string) bool {
	switch extension {
	case "pdf", "png", "jpg", "jpeg", "docx", "xlsx", "txt":
		return true
	default:
		return false
	}
}

func pipelineStepsCompleted(progress []pipeline.StepProgress) bool {
	for _, step := range progress {
		switch step.Status {
		case "completed", "skipped":
			continue
		default:
			return false
		}
	}
	return true
}
