package documents

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"fifriz/internal/audit"
	"fifriz/internal/auth"
	"fifriz/internal/embedding"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrRelatedSuggestionNotFound      = errors.New("related document suggestion not found")
	ErrRelatedSuggestionInvalidStatus = errors.New("related document suggestion is not pending")
	ErrAIRelatedDocumentsDisabled     = errors.New("related documents provider is disabled")
	ErrAIRelatedDocumentsNoEmbeddings = errors.New("document has no embeddings for related document matching")
)

const (
	relatedSuggestionMinScore    = 0.50
	relatedSuggestionMaxCount    = 10
	relatedSuggestionSearchLimit = 25
)

type RelatedSuggestionGenerateRequest struct {
	TenantID         uuid.UUID
	SourceDocumentID uuid.UUID
	UserID           uuid.UUID
	AccessUserID     *uuid.UUID
}

type relatedDocumentCandidate struct {
	DocumentID uuid.UUID
	Similarity float64
	Document   *Document
}

func (s *Service) ListRelatedSuggestions(ctx context.Context, documentIDRaw string) ([]DocumentRelatedSuggestionResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	sourceDocumentID, document, err := s.loadDocumentForRelatedSuggestions(ctx, authContext, documentIDRaw)
	if err != nil {
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	rows, err := s.repository.ListSuggestedRelatedDocuments(ctx, authContext.TenantID, sourceDocumentID)
	if err != nil {
		return nil, err
	}
	return toRelatedSuggestionResponses(rows), nil
}

func (s *Service) GenerateRelatedSuggestions(ctx context.Context, documentIDRaw string) ([]DocumentRelatedSuggestionResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	sourceDocumentID, document, err := s.loadDocumentForRelatedSuggestions(ctx, authContext, documentIDRaw)
	if err != nil {
		return nil, err
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	return s.GenerateRelatedSuggestionsForRequest(ctx, RelatedSuggestionGenerateRequest{
		TenantID:         authContext.TenantID,
		SourceDocumentID: sourceDocumentID,
		UserID:           authContext.UserID,
		AccessUserID:     s.memberAccessUserID(authContext),
	})
}

func (s *Service) AcceptRelatedSuggestion(ctx context.Context, documentIDRaw, suggestionIDRaw string) (*DocumentRelationshipEdge, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	sourceDocumentID, suggestionID, _, suggestion, err := s.loadPendingRelatedSuggestion(ctx, authContext, documentIDRaw, suggestionIDRaw)
	if err != nil {
		return nil, err
	}

	relationshipType := RelationshipTypeRelated
	if suggestion.RelationshipType != nil && strings.TrimSpace(*suggestion.RelationshipType) != "" {
		candidate := strings.TrimSpace(*suggestion.RelationshipType)
		if IsAllowedRelationshipType(candidate) {
			relationshipType = candidate
		}
	}

	targetDocument, err := s.repository.GetByID(ctx, authContext.TenantID, suggestion.TargetDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, targetDocument); err != nil {
		return nil, err
	}

	exists, err := s.repository.RelationshipExists(ctx, authContext.TenantID, sourceDocumentID, suggestion.TargetDocumentID, relationshipType)
	if err != nil {
		return nil, err
	}
	if !exists {
		relationship := &DocumentRelationship{
			TenantID:         authContext.TenantID,
			SourceDocumentID: sourceDocumentID,
			TargetDocumentID: suggestion.TargetDocumentID,
			RelationshipType: relationshipType,
			CreatedByUserID:  authContext.UserID,
		}
		if err := s.repository.CreateRelationship(ctx, relationship); err != nil {
			return nil, err
		}
	}

	suggestion.Status = RelatedSuggestionStatusAccepted
	if err := s.repository.UpdateRelatedSuggestion(ctx, suggestion); err != nil {
		return nil, err
	}

	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, authContext.TenantID, &userID, audit.ActionAIRelatedDocumentAccepted, sourceDocumentID, "ai related document accepted", map[string]interface{}{
		"suggestionId":     suggestionID,
		"targetDocumentId": suggestion.TargetDocumentID,
		"relationshipType": relationshipType,
	})

	rows, err := s.repository.ListRelationshipsByDocument(ctx, authContext.TenantID, sourceDocumentID)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.TargetDocumentID == suggestion.TargetDocumentID && row.RelationshipType == relationshipType {
			response := buildRelationshipsResponse(sourceDocumentID, []RelationshipListRow{row})
			if len(response.Items) > 0 {
				item := response.Items[0]
				return &item, nil
			}
		}
	}
	return nil, ErrDocumentRelationshipNotFound
}

func (s *Service) RejectRelatedSuggestion(ctx context.Context, documentIDRaw, suggestionIDRaw string) (*DocumentRelatedSuggestionResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	sourceDocumentID, suggestionID, _, suggestion, err := s.loadPendingRelatedSuggestion(ctx, authContext, documentIDRaw, suggestionIDRaw)
	if err != nil {
		return nil, err
	}

	suggestion.Status = RelatedSuggestionStatusRejected
	if err := s.repository.UpdateRelatedSuggestion(ctx, suggestion); err != nil {
		return nil, err
	}

	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, authContext.TenantID, &userID, audit.ActionAIRelatedDocumentRejected, sourceDocumentID, "ai related document rejected", map[string]interface{}{
		"suggestionId":     suggestionID,
		"targetDocumentId": suggestion.TargetDocumentID,
	})

	targetDocument, err := s.repository.GetByID(ctx, authContext.TenantID, suggestion.TargetDocumentID)
	if err != nil {
		response := toRelatedSuggestionResponse(RelatedSuggestionListRow{DocumentRelatedSuggestion: *suggestion})
		return &response, nil
	}

	response := toRelatedSuggestionResponse(RelatedSuggestionListRow{
		DocumentRelatedSuggestion: *suggestion,
		TargetOriginalFileName:    targetDocument.OriginalFileName,
		TargetFileExtension:       targetDocument.FileExtension,
		TargetMimeType:            targetDocument.MimeType,
		TargetFolderID:            targetDocument.FolderID,
		TargetFolderName:          targetDocument.FolderName,
	})
	return &response, nil
}

func (s *Service) GenerateRelatedSuggestionsForRequest(ctx context.Context, request RelatedSuggestionGenerateRequest) ([]DocumentRelatedSuggestionResponse, error) {
	sourceDocument, err := s.repository.GetByID(ctx, request.TenantID, request.SourceDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	chunks, err := s.repository.ListChunksByDocumentID(ctx, request.SourceDocumentID)
	if err != nil {
		return nil, err
	}

	avgEmbedding, ok := averageChunkEmbeddings(chunks)
	if !ok {
		return nil, ErrAIRelatedDocumentsNoEmbeddings
	}

	scores, err := s.repository.SearchSimilarDocumentsByEmbedding(
		ctx,
		request.TenantID,
		avgEmbedding,
		request.SourceDocumentID,
		relatedSuggestionSearchLimit,
		request.AccessUserID,
	)
	if err != nil {
		return nil, err
	}

	sourceTagIDs, err := s.loadDocumentTagIDs(ctx, request.TenantID, request.SourceDocumentID)
	if err != nil {
		return nil, err
	}

	candidates := make([]relatedDocumentCandidate, 0, len(scores))
	for _, score := range scores {
		if score.Similarity < relatedSuggestionMinScore {
			continue
		}

		targetDocument, err := s.repository.GetByID(ctx, request.TenantID, score.DocumentID)
		if err != nil {
			continue
		}

		boosted := s.scoreRelatedDocumentCandidate(ctx, request, sourceDocument, targetDocument, sourceTagIDs, score.Similarity)
		if boosted < relatedSuggestionMinScore {
			continue
		}

		candidates = append(candidates, relatedDocumentCandidate{
			DocumentID: score.DocumentID,
			Similarity: boosted,
			Document:   targetDocument,
		})
	}

	candidates = sortAndLimitCandidates(candidates, relatedSuggestionMaxCount)
	if len(candidates) == 0 {
		if err := s.repository.DeleteSuggestedRelatedSuggestions(ctx, request.TenantID, request.SourceDocumentID); err != nil {
			return nil, err
		}
		return []DocumentRelatedSuggestionResponse{}, nil
	}

	if err := s.repository.DeleteSuggestedRelatedSuggestions(ctx, request.TenantID, request.SourceDocumentID); err != nil {
		return nil, err
	}

	created := make([]DocumentRelatedSuggestion, 0, len(candidates))
	for _, candidate := range candidates {
		existsSuggestion, err := s.repository.HasRelatedSuggestionPair(ctx, request.TenantID, request.SourceDocumentID, candidate.DocumentID)
		if err != nil {
			return nil, err
		}
		if existsSuggestion {
			continue
		}

		alreadyLinked, err := s.repository.HasAnyRelationshipBetweenDocuments(ctx, request.TenantID, request.SourceDocumentID, candidate.DocumentID)
		if err != nil {
			return nil, err
		}
		if alreadyLinked {
			continue
		}

		relationshipType := inferRelatedRelationshipType(sourceDocument.AIDocumentType, candidate.Document.AIDocumentType)
		targetTagIDs, _ := s.loadDocumentTagIDs(ctx, request.TenantID, candidate.DocumentID)
		reason := buildRelatedSuggestionReason(sourceDocument, candidate.Document, sourceTagIDs, targetTagIDs, candidate.Similarity)
		sharedEntities, _ := s.repository.CountSharedEntities(ctx, request.TenantID, sourceDocument.ID, candidate.DocumentID)
		if sharedEntities > 0 {
			reason = strings.TrimSuffix(reason, ".") + ", ortak varlık."
		}
		similarity := clampConfidence(candidate.Similarity)

		item := DocumentRelatedSuggestion{
			TenantID:         request.TenantID,
			SourceDocumentID: request.SourceDocumentID,
			TargetDocumentID: candidate.DocumentID,
			RelationshipType: &relationshipType,
			Reason:           &reason,
			Similarity:       &similarity,
			Status:           RelatedSuggestionStatusSuggested,
			GeneratedBy:      RelatedSuggestionGeneratedByAI,
		}
		if err := s.repository.CreateRelatedSuggestions(ctx, []DocumentRelatedSuggestion{item}); err != nil {
			return nil, err
		}
		created = append(created, item)
	}

	userID := request.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, request.TenantID, &userID, audit.ActionAIRelatedDocumentsGenerated, request.SourceDocumentID, "ai related documents generated", map[string]interface{}{
		"count": len(created),
	})

	rows, err := s.repository.ListSuggestedRelatedDocuments(ctx, request.TenantID, request.SourceDocumentID)
	if err != nil {
		return nil, err
	}
	return toRelatedSuggestionResponses(rows), nil
}

func (s *Service) loadDocumentForRelatedSuggestions(ctx context.Context, authContext auth.Context, documentIDRaw string) (uuid.UUID, *Document, error) {
	documentID, err := uuid.Parse(strings.TrimSpace(documentIDRaw))
	if err != nil {
		return uuid.Nil, nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, nil, ErrDocumentNotFound
		}
		return uuid.Nil, nil, err
	}
	return documentID, document, nil
}

func (s *Service) loadPendingRelatedSuggestion(ctx context.Context, authContext auth.Context, documentIDRaw, suggestionIDRaw string) (uuid.UUID, uuid.UUID, *Document, *DocumentRelatedSuggestion, error) {
	sourceDocumentID, sourceDocument, err := s.loadDocumentForRelatedSuggestions(ctx, authContext, documentIDRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, nil, err
	}
	if err := s.requireEditAccess(ctx, authContext, sourceDocument); err != nil {
		return uuid.Nil, uuid.Nil, nil, nil, err
	}

	suggestionID, err := uuid.Parse(strings.TrimSpace(suggestionIDRaw))
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, nil, ErrRelatedSuggestionNotFound
	}

	suggestion, err := s.repository.GetRelatedSuggestionByID(ctx, authContext.TenantID, sourceDocumentID, suggestionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, uuid.Nil, nil, nil, ErrRelatedSuggestionNotFound
		}
		return uuid.Nil, uuid.Nil, nil, nil, err
	}
	if suggestion.Status != RelatedSuggestionStatusSuggested {
		return uuid.Nil, uuid.Nil, nil, nil, ErrRelatedSuggestionInvalidStatus
	}
	return sourceDocumentID, suggestionID, sourceDocument, suggestion, nil
}

func (s *Service) loadDocumentTagIDs(ctx context.Context, tenantID, documentID uuid.UUID) ([]uuid.UUID, error) {
	tagMap, err := s.tags.MapTagsForDocuments(ctx, tenantID, []uuid.UUID{documentID})
	if err != nil {
		return nil, err
	}
	tagsForDocument := tagMap[documentID]
	ids := make([]uuid.UUID, 0, len(tagsForDocument))
	for _, tag := range tagsForDocument {
		ids = append(ids, tag.ID)
	}
	return ids, nil
}

func (s *Service) scoreRelatedDocumentCandidate(
	ctx context.Context,
	request RelatedSuggestionGenerateRequest,
	sourceDocument, targetDocument *Document,
	sourceTagIDs []uuid.UUID,
	baseSimilarity float64,
) float64 {
	score := baseSimilarity

	if sourceDocument.FolderID != nil && targetDocument.FolderID != nil && *sourceDocument.FolderID == *targetDocument.FolderID {
		score += 0.05
	}

	if sourceDocument.AIDocumentType != nil && targetDocument.AIDocumentType != nil &&
		strings.EqualFold(strings.TrimSpace(*sourceDocument.AIDocumentType), strings.TrimSpace(*targetDocument.AIDocumentType)) {
		score += 0.04
	}

	if sourceDocument.SourceType != "" && sourceDocument.SourceType == targetDocument.SourceType && sourceDocument.SourceType != "manual_upload" {
		score += 0.02
	}

	if sourceDocument.CreatedAt.Year() == targetDocument.CreatedAt.Year() {
		score += 0.02
	}

	targetTagIDs, err := s.loadDocumentTagIDs(ctx, request.TenantID, targetDocument.ID)
	if err == nil && len(sourceTagIDs) > 0 && len(targetTagIDs) > 0 {
		shared := countSharedUUIDs(sourceTagIDs, targetTagIDs)
		if shared > 0 {
			boost := math.Min(float64(shared)*0.03, 0.09)
			score += boost
		}
	}

	sharedEntities, err := s.repository.CountSharedEntities(ctx, request.TenantID, sourceDocument.ID, targetDocument.ID)
	if err == nil && sharedEntities > 0 {
		score += math.Min(float64(sharedEntities)*0.025, 0.08)
	}

	return math.Min(score, 1.0)
}

func buildRelatedSuggestionReason(
	sourceDocument, targetDocument *Document,
	sourceTagIDs, targetTagIDs []uuid.UUID,
	similarity float64,
) string {
	parts := []string{fmt.Sprintf("Semantik benzerlik %%%.0f", similarity*100)}

	if sourceDocument.FolderID != nil && targetDocument.FolderID != nil && *sourceDocument.FolderID == *targetDocument.FolderID {
		parts = append(parts, "aynı klasör")
	}

	if countSharedUUIDs(sourceTagIDs, targetTagIDs) > 0 {
		parts = append(parts, "ortak etiket")
	}

	if sourceDocument.AIDocumentType != nil && targetDocument.AIDocumentType != nil &&
		strings.EqualFold(strings.TrimSpace(*sourceDocument.AIDocumentType), strings.TrimSpace(*targetDocument.AIDocumentType)) {
		parts = append(parts, "aynı AI belge türü")
	}

	if sourceDocument.SourceType != "" && sourceDocument.SourceType == targetDocument.SourceType && sourceDocument.SourceType != "manual_upload" {
		parts = append(parts, "aynı kaynak")
	}

	if sourceDocument.CreatedAt.Year() == targetDocument.CreatedAt.Year() {
		parts = append(parts, "aynı yıl")
	}

	return strings.Join(parts, ", ") + "."
}

func averageChunkEmbeddings(chunks []DocumentChunk) ([]float32, bool) {
	if len(chunks) == 0 {
		return nil, false
	}

	var vectors [][]float32
	for _, chunk := range chunks {
		if chunk.Embedding == nil {
			continue
		}
		slice := chunk.Embedding.Slice()
		if len(slice) != embedding.DefaultEmbeddingDimensions {
			continue
		}
		vectors = append(vectors, slice)
	}
	if len(vectors) == 0 {
		return nil, false
	}

	avg := make([]float32, embedding.DefaultEmbeddingDimensions)
	for _, vector := range vectors {
		for index, value := range vector {
			avg[index] += value
		}
	}
	count := float32(len(vectors))
	for index := range avg {
		avg[index] /= count
	}
	return avg, true
}

func sortAndLimitCandidates(candidates []relatedDocumentCandidate, limit int) []relatedDocumentCandidate {
	if len(candidates) <= 1 {
		if len(candidates) > limit {
			return candidates[:limit]
		}
		return candidates
	}

	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].Similarity > candidates[i].Similarity {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}
	if len(candidates) > limit {
		return candidates[:limit]
	}
	return candidates
}

func countSharedUUIDs(left, right []uuid.UUID) int {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	lookup := make(map[uuid.UUID]struct{}, len(left))
	for _, id := range left {
		lookup[id] = struct{}{}
	}
	shared := 0
	for _, id := range right {
		if _, ok := lookup[id]; ok {
			shared++
		}
	}
	return shared
}

func inferRelatedRelationshipType(sourceType, targetType *string) string {
	source := strings.TrimSpace(stringValue(sourceType))
	target := strings.TrimSpace(stringValue(targetType))

	switch {
	case source == "contract" && (target == "identity_document" || target == "passport"):
		return RelationshipTypeIdentityOf
	case source == "invoice" && target == "bank_receipt":
		return RelationshipTypeReference
	case source == "contract" && target == "insurance_policy":
		return RelationshipTypeAttachment
	case source == "vehicle_registration" && target == "insurance_policy":
		return RelationshipTypeAttachment
	case source == "court_decision" && (target == "contract" || target == "identity_document"):
		return RelationshipTypeReference
	default:
		return RelationshipTypeRelated
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
