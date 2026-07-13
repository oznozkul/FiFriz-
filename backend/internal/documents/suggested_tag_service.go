package documents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"unicode/utf8"

	"fifriz/internal/audit"
	"fifriz/internal/auth"
	domain "fifriz/internal/events/events"
	"fifriz/internal/tags"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type suggestedTagLLMItem struct {
	Name       string   `json:"name"`
	Confidence *float64 `json:"confidence"`
}

type suggestedTagLLMResponse struct {
	Tags []suggestedTagLLMItem `json:"tags"`
}

var suggestedTagJSONFencePattern = regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*\\})\\s*```")

const suggestedTagContextMaxRunes = 10000
const suggestedTagMinCount = 5
const suggestedTagMaxCount = 10

func (s *Service) GetSuggestedTags(ctx context.Context, documentID string) ([]DocumentSuggestedTagResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, document, err := s.loadDocumentForSuggestedTags(ctx, authContext, documentID)
	if err != nil {
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	items, err := s.repository.ListSuggestedTagsByDocument(ctx, authContext.TenantID, parsedDocumentID, SuggestedTagStatusPending)
	if err != nil {
		return nil, err
	}

	responses := make([]DocumentSuggestedTagResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toSuggestedTagResponse(item))
	}
	return responses, nil
}

func (s *Service) AcceptSuggestedTag(ctx context.Context, documentID, suggestedTagID string) (*AcceptSuggestedTagsResult, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, parsedSuggestedTagID, document, suggestedTag, err := s.loadPendingSuggestedTag(ctx, authContext, documentID, suggestedTagID)
	if err != nil {
		return nil, err
	}

	accepted, documentTags, err := s.acceptSuggestedTagRecord(ctx, authContext, document, suggestedTag)
	if err != nil {
		return nil, err
	}

	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, authContext.TenantID, &userID, audit.ActionAITagAccepted, parsedDocumentID, "ai suggested tag accepted", map[string]interface{}{
		"suggestedTagId": parsedSuggestedTagID,
		"tagName":        suggestedTag.TagName,
	})

	return &AcceptSuggestedTagsResult{
		Accepted: []DocumentSuggestedTagResponse{accepted},
		Tags:     documentTags,
	}, nil
}

func (s *Service) RejectSuggestedTag(ctx context.Context, documentID, suggestedTagID string) (*DocumentSuggestedTagResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, parsedSuggestedTagID, _, suggestedTag, err := s.loadPendingSuggestedTag(ctx, authContext, documentID, suggestedTagID)
	if err != nil {
		return nil, err
	}

	if err := s.repository.UpdateSuggestedTagStatus(ctx, authContext.TenantID, parsedDocumentID, parsedSuggestedTagID, SuggestedTagStatusRejected); err != nil {
		return nil, err
	}
	suggestedTag.Status = SuggestedTagStatusRejected

	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, authContext.TenantID, &userID, audit.ActionAITagRejected, parsedDocumentID, "ai suggested tag rejected", map[string]interface{}{
		"suggestedTagId": parsedSuggestedTagID,
		"tagName":        suggestedTag.TagName,
	})

	return toSuggestedTagResponsePtr(*suggestedTag), nil
}

func (s *Service) AcceptAllSuggestedTags(ctx context.Context, documentID string) (*AcceptSuggestedTagsResult, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, document, err := s.loadDocumentForSuggestedTags(ctx, authContext, documentID)
	if err != nil {
		return nil, err
	}
	if !canAssignDocumentTags(authContext, document) {
		return nil, ErrForbidden
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	pending, err := s.repository.ListSuggestedTagsByDocument(ctx, authContext.TenantID, parsedDocumentID, SuggestedTagStatusPending)
	if err != nil {
		return nil, err
	}

	accepted := make([]DocumentSuggestedTagResponse, 0, len(pending))
	for _, item := range pending {
		record, _, acceptErr := s.acceptSuggestedTagRecord(ctx, authContext, document, &item)
		if acceptErr != nil {
			return nil, acceptErr
		}
		accepted = append(accepted, record)
	}

	tagMap, err := s.tags.MapTagsForDocuments(ctx, authContext.TenantID, []uuid.UUID{parsedDocumentID})
	if err != nil {
		return nil, err
	}

	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, authContext.TenantID, &userID, audit.ActionAITagAccepted, parsedDocumentID, "all ai suggested tags accepted", map[string]interface{}{
		"count": len(accepted),
	})

	return &AcceptSuggestedTagsResult{
		Accepted: accepted,
		Tags:     tagMap[parsedDocumentID],
	}, nil
}

func (s *Service) GenerateSuggestedTagsForRequest(ctx context.Context, request SuggestedTagGenerateRequest) ([]DocumentSuggestedTagResponse, error) {
	if s.summaryChatProvider == nil {
		return nil, ErrSuggestedTagsProviderDisabled
	}
	if s.tags == nil {
		return nil, ErrSuggestedTagsProviderDisabled
	}

	document, err := s.repository.GetByID(ctx, request.TenantID, request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	contextText, err := s.buildSuggestedTagContext(ctx, request.DocumentID)
	if err != nil {
		return nil, err
	}

	existingTagNames, err := s.tags.ListKnownTagTerms(ctx, request.TenantID)
	if err != nil {
		return nil, err
	}

	parsedSuggestions, err := s.generateSuggestedTagsFromContext(ctx, document.OriginalFileName, document.FileExtension, contextText, existingTagNames)
	if err != nil {
		return nil, err
	}
	if len(parsedSuggestions) == 0 {
		return []DocumentSuggestedTagResponse{}, nil
	}

	documentTagNames, err := s.listDocumentTagNames(ctx, request.TenantID, request.DocumentID)
	if err != nil {
		return nil, err
	}

	if err := s.repository.DeletePendingSuggestedTags(ctx, request.TenantID, request.DocumentID); err != nil {
		return nil, err
	}

	records := make([]DocumentSuggestedTag, 0, len(parsedSuggestions))
	seenNames := make(map[string]struct{}, len(parsedSuggestions))
	for _, suggestion := range parsedSuggestions {
		normalizedName := normalizeSuggestedTagName(suggestion.Name, existingTagNames)
		if normalizedName == "" {
			continue
		}
		lookupKey := strings.ToLower(normalizedName)
		if _, exists := seenNames[lookupKey]; exists {
			continue
		}
		if containsTagName(documentTagNames, normalizedName) {
			continue
		}
		seenNames[lookupKey] = struct{}{}

		confidence := suggestion.Confidence
		if confidence != nil {
			clamped := clampConfidence(*confidence)
			confidence = &clamped
		}

		records = append(records, DocumentSuggestedTag{
			TenantID:    request.TenantID,
			DocumentID:  request.DocumentID,
			TagName:     normalizedName,
			Confidence:  confidence,
			Status:      SuggestedTagStatusPending,
			GeneratedBy: "ai",
		})
	}

	if len(records) == 0 {
		return []DocumentSuggestedTagResponse{}, nil
	}

	if err := s.repository.CreateSuggestedTags(ctx, records); err != nil {
		return nil, err
	}

	userID := request.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, request.TenantID, &userID, audit.ActionAITagsGenerated, request.DocumentID, "ai tag suggestions generated", map[string]interface{}{
		"count": len(records),
		"model": s.summaryChatModel,
	})

	responses := make([]DocumentSuggestedTagResponse, 0, len(records))
	for _, record := range records {
		responses = append(responses, toSuggestedTagResponse(record))
	}
	return responses, nil
}

func (s *Service) loadDocumentForSuggestedTags(ctx context.Context, authContext auth.Context, documentID string) (uuid.UUID, *Document, error) {
	parsedDocumentID, err := uuid.Parse(strings.TrimSpace(documentID))
	if err != nil {
		return uuid.Nil, nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, parsedDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, nil, ErrDocumentNotFound
		}
		return uuid.Nil, nil, err
	}

	return parsedDocumentID, document, nil
}

func (s *Service) loadPendingSuggestedTag(ctx context.Context, authContext auth.Context, documentID, suggestedTagID string) (uuid.UUID, uuid.UUID, *Document, *DocumentSuggestedTag, error) {
	parsedDocumentID, document, err := s.loadDocumentForSuggestedTags(ctx, authContext, documentID)
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, nil, err
	}
	if !canAssignDocumentTags(authContext, document) {
		return uuid.Nil, uuid.Nil, nil, nil, ErrForbidden
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return uuid.Nil, uuid.Nil, nil, nil, err
	}

	parsedSuggestedTagID, err := uuid.Parse(strings.TrimSpace(suggestedTagID))
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, nil, ErrSuggestedTagNotFound
	}

	suggestedTag, err := s.repository.GetSuggestedTagByID(ctx, authContext.TenantID, parsedDocumentID, parsedSuggestedTagID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, uuid.Nil, nil, nil, ErrSuggestedTagNotFound
		}
		return uuid.Nil, uuid.Nil, nil, nil, err
	}
	if suggestedTag.Status != SuggestedTagStatusPending {
		return uuid.Nil, uuid.Nil, nil, nil, ErrSuggestedTagInvalidStatus
	}

	return parsedDocumentID, parsedSuggestedTagID, document, suggestedTag, nil
}

func (s *Service) acceptSuggestedTagRecord(ctx context.Context, authContext auth.Context, document *Document, suggestedTag *DocumentSuggestedTag) (DocumentSuggestedTagResponse, []tags.TagBrief, error) {
	if s.tags == nil {
		return DocumentSuggestedTagResponse{}, nil, fmt.Errorf("tag service is not configured")
	}

	tag, err := s.tags.ResolveTagByNameOrAlias(ctx, authContext.TenantID, authContext.UserID, suggestedTag.TagName)
	if err != nil {
		return DocumentSuggestedTagResponse{}, nil, err
	}

	if err := s.tags.AppendDocumentTag(ctx, authContext.TenantID, document.ID, tag.ID); err != nil {
		return DocumentSuggestedTagResponse{}, nil, err
	}

	if err := s.repository.UpdateSuggestedTagStatus(ctx, authContext.TenantID, document.ID, suggestedTag.ID, SuggestedTagStatusAccepted); err != nil {
		return DocumentSuggestedTagResponse{}, nil, err
	}
	suggestedTag.Status = SuggestedTagStatusAccepted

	tagMap, err := s.tags.MapTagsForDocuments(ctx, authContext.TenantID, []uuid.UUID{document.ID})
	if err != nil {
		return DocumentSuggestedTagResponse{}, nil, err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewTagAssigned(authContext.TenantID, authContext.UserID, document.ID, []uuid.UUID{tag.ID}))

	return toSuggestedTagResponse(*suggestedTag), tagMap[document.ID], nil
}

func (s *Service) buildSuggestedTagContext(ctx context.Context, documentID uuid.UUID) (string, error) {
	chunks, err := s.repository.ListChunksByDocumentID(ctx, documentID)
	if err != nil {
		return "", err
	}

	if len(chunks) > 0 {
		var builder strings.Builder
		usedRunes := 0
		for _, chunk := range chunks {
			content := strings.TrimSpace(chunk.Content)
			if content == "" {
				continue
			}
			chunkRunes := utf8.RuneCountInString(content)
			if usedRunes+chunkRunes > suggestedTagContextMaxRunes {
				remaining := suggestedTagContextMaxRunes - usedRunes
				if remaining <= 0 {
					break
				}
				runes := []rune(content)
				if len(runes) > remaining {
					content = string(runes[:remaining])
				}
			}

			if builder.Len() > 0 {
				builder.WriteString("\n\n")
			}
			builder.WriteString(content)
			usedRunes += utf8.RuneCountInString(content)
			if usedRunes >= suggestedTagContextMaxRunes {
				break
			}
		}

		if builder.Len() > 0 {
			return builder.String(), nil
		}
	}

	documentText, err := s.repository.GetTextByDocumentID(ctx, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrSuggestedTagsNoContent
		}
		return "", err
	}

	fullText := strings.TrimSpace(documentText.FullText)
	if fullText == "" {
		return "", ErrSuggestedTagsNoContent
	}

	if utf8.RuneCountInString(fullText) > suggestedTagContextMaxRunes {
		fullText = string([]rune(fullText)[:suggestedTagContextMaxRunes])
	}

	return fullText, nil
}

func (s *Service) generateSuggestedTagsFromContext(ctx context.Context, fileName, fileExtension, contextText string, existingTagNames []string) ([]suggestedTagLLMItem, error) {
	completion, err := s.summaryChatProvider.Complete(ctx, suggestedTagSystemPrompt(), buildSuggestedTagUserPrompt(fileName, fileExtension, contextText, existingTagNames))
	if err != nil {
		log.Printf("document tag suggestion generation failed: %v", err)
		return nil, err
	}

	parsed, err := parseSuggestedTagLLMResponse(completion)
	if err != nil {
		return nil, fmt.Errorf("tag suggestion response parse failed: %w", err)
	}
	if len(parsed.Tags) == 0 {
		return nil, fmt.Errorf("tag suggestion response was empty")
	}

	return parsed.Tags, nil
}

func suggestedTagSystemPrompt() string {
	return strings.TrimSpace(`
Sen FiFriz belge etiketleme asistanısın.
Yalnızca verilen belge içeriğine dayanarak Türkçe, kısa ve anlamlı etiket öner.
Etiketler 1-3 kelime olmalı, küçük harf tercih et (özel isimler hariç).
Bağlamda olmayan etiketleri uydurma.
5 ile 10 arasında etiket öner.
Yanıtını yalnızca geçerli JSON olarak ver:
{"tags":[{"name":"fatura","confidence":0.92},{"name":"kira sözleşmesi","confidence":0.85}]}
confidence değeri 0 ile 1 arasında olmalı.
`)
}

func buildSuggestedTagUserPrompt(fileName, fileExtension, contextText string, existingTagNames []string) string {
	title := strings.TrimSpace(fileName)
	if title == "" {
		title = "Belge"
	}
	extension := strings.TrimSpace(fileExtension)
	if extension != "" {
		title = fmt.Sprintf("%s (%s)", title, strings.ToUpper(extension))
	}

	existingSection := "Mevcut tenant etiketleri: yok"
	if len(existingTagNames) > 0 {
		existingSection = "Mevcut tenant etiketleri: " + strings.Join(existingTagNames, ", ")
	}

	return fmt.Sprintf(`Aşağıdaki belge için etiket öner.

BELGE ADI: %s
%s

BELGE İÇERİĞİ:
%s

Tenant içinde zaten var olan etiketlerle aynı anlama geliyorsa aynı ismi kullan.
JSON formatında tags dizisi döndür.`, title, existingSection, contextText)
}

func parseSuggestedTagLLMResponse(content string) (*suggestedTagLLMResponse, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("empty llm response")
	}

	candidates := []string{content}
	if match := suggestedTagJSONFencePattern.FindStringSubmatch(content); len(match) == 2 {
		candidates = append([]string{match[1]}, candidates...)
	}
	if start := strings.Index(content, "{"); start >= 0 {
		if end := strings.LastIndex(content, "}"); end > start {
			candidates = append([]string{content[start : end+1]}, candidates...)
		}
	}

	var lastErr error
	for _, candidate := range candidates {
		var parsed suggestedTagLLMResponse
		if err := json.Unmarshal([]byte(candidate), &parsed); err != nil {
			lastErr = err
			continue
		}
		parsed.Tags = normalizeSuggestedTagItems(parsed.Tags)
		return &parsed, nil
	}

	return nil, lastErr
}

func normalizeSuggestedTagItems(items []suggestedTagLLMItem) []suggestedTagLLMItem {
	if len(items) == 0 {
		return []suggestedTagLLMItem{}
	}

	normalized := make([]suggestedTagLLMItem, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, suggestedTagLLMItem{
			Name:       name,
			Confidence: item.Confidence,
		})
		if len(normalized) >= suggestedTagMaxCount {
			break
		}
	}
	return normalized
}

func normalizeSuggestedTagName(candidate string, existingTagNames []string) string {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return ""
	}
	for _, existing := range existingTagNames {
		if strings.EqualFold(strings.TrimSpace(existing), candidate) {
			return strings.TrimSpace(existing)
		}
	}
	return candidate
}

func containsTagName(names []string, candidate string) bool {
	for _, name := range names {
		if strings.EqualFold(strings.TrimSpace(name), strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}

func clampConfidence(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func (s *Service) listDocumentTagNames(ctx context.Context, tenantID, documentID uuid.UUID) ([]string, error) {
	tagMap, err := s.tags.MapTagsForDocuments(ctx, tenantID, []uuid.UUID{documentID})
	if err != nil {
		return nil, err
	}
	briefs := tagMap[documentID]
	names := make([]string, 0, len(briefs))
	for _, brief := range briefs {
		names = append(names, brief.Name)
	}
	return names, nil
}

func toSuggestedTagResponse(item DocumentSuggestedTag) DocumentSuggestedTagResponse {
	return DocumentSuggestedTagResponse{
		ID:          item.ID,
		DocumentID:  item.DocumentID,
		TagName:     item.TagName,
		Confidence:  item.Confidence,
		Status:      item.Status,
		GeneratedBy: item.GeneratedBy,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func toSuggestedTagResponsePtr(item DocumentSuggestedTag) *DocumentSuggestedTagResponse {
	response := toSuggestedTagResponse(item)
	return &response
}
