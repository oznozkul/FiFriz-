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

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrMetadataProviderDisabled = errors.New("metadata provider is disabled")
	ErrMetadataNoContent        = errors.New("document has no extractable content for metadata")
	ErrMetadataParseFailed      = errors.New("metadata response parse failed")
)

const aiMetadataContextMaxRunes = 12000

var aiMetadataJSONFencePattern = regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*\\})\\s*```")

func (s *Service) GetAIMetadata(ctx context.Context, documentID string) (*DocumentAIMetadataResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, document, err := s.loadDocumentForAIMetadata(ctx, authContext, documentID)
	if err != nil {
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	record, err := s.repository.GetAIMetadataByDocumentID(ctx, authContext.TenantID, parsedDocumentID)
	if err != nil {
		return nil, err
	}

	s.logAIMetadataViewed(ctx, authContext, parsedDocumentID)
	return toDocumentAIMetadataResponse(record)
}

func (s *Service) GenerateAIMetadata(ctx context.Context, documentID string) (*DocumentAIMetadataResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, document, err := s.loadDocumentForAIMetadata(ctx, authContext, documentID)
	if err != nil {
		return nil, err
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	return s.GenerateAIMetadataForRequest(ctx, AIMetadataGenerateRequest{
		TenantID:   authContext.TenantID,
		DocumentID: parsedDocumentID,
		UserID:     authContext.UserID,
	})
}

func (s *Service) GenerateAIMetadataForRequest(ctx context.Context, request AIMetadataGenerateRequest) (*DocumentAIMetadataResponse, error) {
	if s.summaryChatProvider == nil {
		return nil, ErrMetadataProviderDisabled
	}

	document, err := s.repository.GetByID(ctx, request.TenantID, request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	contextText, err := s.buildAIMetadataContext(ctx, request.DocumentID)
	if err != nil {
		return nil, err
	}

	parsed, rawJSON, err := s.generateAIMetadataFromContext(ctx, document.OriginalFileName, document.FileExtension, contextText)
	if err != nil {
		return nil, err
	}

	metadataJSON, err := json.Marshal(parsed)
	if err != nil {
		return nil, err
	}

	var documentType *string
	if parsed.DocumentType != "" {
		value := strings.TrimSpace(parsed.DocumentType)
		documentType = &value
	}

	confidence := parsed.Confidence
	if confidence != nil {
		clamped := clampConfidence(*confidence)
		confidence = &clamped
	}

	var modelName *string
	if s.summaryChatModel != "" {
		modelName = &s.summaryChatModel
	}

	record := &DocumentAIMetadata{
		TenantID:     request.TenantID,
		DocumentID:   request.DocumentID,
		ExtractedBy:  "ai",
		DocumentType: documentType,
		Metadata:     metadataJSON,
		Confidence:   confidence,
		Model:        modelName,
	}
	if err := s.repository.SaveAIMetadata(ctx, record); err != nil {
		return nil, err
	}

	saved, err := s.repository.GetAIMetadataByDocumentID(ctx, request.TenantID, request.DocumentID)
	if err != nil {
		return nil, err
	}

	userID := request.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, request.TenantID, &userID, audit.ActionAIMetadataGenerated, request.DocumentID, "ai metadata generated", map[string]interface{}{
		"model":        s.summaryChatModel,
		"documentType": documentType,
		"rawSize":      len(rawJSON),
	})

	return toDocumentAIMetadataResponse(saved)
}

func (s *Service) loadDocumentForAIMetadata(ctx context.Context, authContext auth.Context, documentID string) (uuid.UUID, *Document, error) {
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

func (s *Service) buildAIMetadataContext(ctx context.Context, documentID uuid.UUID) (string, error) {
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
			if usedRunes+chunkRunes > aiMetadataContextMaxRunes {
				remaining := aiMetadataContextMaxRunes - usedRunes
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
			if usedRunes >= aiMetadataContextMaxRunes {
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
			return "", ErrMetadataNoContent
		}
		return "", err
	}

	fullText := strings.TrimSpace(documentText.FullText)
	if fullText == "" {
		return "", ErrMetadataNoContent
	}

	if utf8.RuneCountInString(fullText) > aiMetadataContextMaxRunes {
		fullText = string([]rune(fullText)[:aiMetadataContextMaxRunes])
	}

	return fullText, nil
}

func (s *Service) generateAIMetadataFromContext(ctx context.Context, fileName, fileExtension, contextText string) (*ExtractedDocumentMetadata, []byte, error) {
	completion, err := s.summaryChatProvider.Complete(ctx, aiMetadataSystemPrompt(), buildAIMetadataUserPrompt(fileName, fileExtension, contextText))
	if err != nil {
		log.Printf("document metadata generation failed: %v", err)
		if errors.Is(err, ErrSummaryProviderUnavailable) {
			return nil, nil, ErrSummaryProviderUnavailable
		}
		return nil, nil, err
	}

	parsed, rawJSON, err := parseAIMetadataLLMResponse(completion)
	if err != nil {
		log.Printf("document metadata parse failed: %v", err)
		return nil, rawJSON, fmt.Errorf("%w: %v", ErrMetadataParseFailed, err)
	}

	return parsed, rawJSON, nil
}

func aiMetadataSystemPrompt() string {
	return strings.TrimSpace(`
Sen FiFriz belge metadata çıkarma asistanısın.
Yalnızca verilen belge içeriğine dayanarak yapılandırılmış metadata üret.
Bağlamda olmayan bilgileri uydurma; emin olmadığın alanları boş bırak.
Tarihleri ISO 8601 (YYYY-MM-DD) formatında ver.
Tutarları sayı olarak ver; para birimini currency alanında belirt (ör. TRY, USD, EUR).
Belge türünü İngilizce snake_case veya kısa kod olarak ver (ör. contract, invoice, identity, receipt, policy).
Yanıtını yalnızca geçerli JSON olarak ver:
{
  "documentType": "contract",
  "title": "Kira Sözleşmesi",
  "parties": [{"name":"...","role":"..."}],
  "dates": [{"label":"Sözleşme Başlangıç Tarihi","value":"2025-01-01"}],
  "amounts": [{"label":"Kira Bedeli","value":25000,"currency":"TRY"}],
  "identifiers": [{"label":"Vergi No","value":"..."}],
  "addresses": [],
  "keywords": [],
  "summary": "",
  "confidence": 0.85
}
`)
}

func buildAIMetadataUserPrompt(fileName, fileExtension, contextText string) string {
	title := strings.TrimSpace(fileName)
	if title == "" {
		title = "Belge"
	}
	extension := strings.TrimSpace(fileExtension)
	if extension != "" {
		title = fmt.Sprintf("%s (%s)", title, strings.ToUpper(extension))
	}

	return fmt.Sprintf(`Aşağıdaki belgeden metadata çıkar.

BELGE ADI: %s

BELGE İÇERİĞİ:
%s

Yalnızca JSON döndür.`, title, contextText)
}

func parseAIMetadataLLMResponse(content string) (*ExtractedDocumentMetadata, []byte, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, nil, errors.New("empty llm response")
	}

	candidates := []string{content}
	if match := aiMetadataJSONFencePattern.FindStringSubmatch(content); len(match) == 2 {
		candidates = append([]string{match[1]}, candidates...)
	}
	if start := strings.Index(content, "{"); start >= 0 {
		if end := strings.LastIndex(content, "}"); end > start {
			candidates = append([]string{content[start : end+1]}, candidates...)
		}
	}

	var lastErr error
	for _, candidate := range candidates {
		rawJSON := []byte(candidate)
		var parsed ExtractedDocumentMetadata
		if err := json.Unmarshal(rawJSON, &parsed); err != nil {
			lastErr = err
			continue
		}
		normalizeExtractedMetadata(&parsed)
		return &parsed, rawJSON, nil
	}

	return nil, []byte(content), lastErr
}

func normalizeExtractedMetadata(metadata *ExtractedDocumentMetadata) {
	if metadata == nil {
		return
	}
	metadata.DocumentType = strings.TrimSpace(metadata.DocumentType)
	metadata.Title = strings.TrimSpace(metadata.Title)
	metadata.Summary = strings.TrimSpace(metadata.Summary)

	metadata.Parties = normalizeMetadataParties(metadata.Parties)
	metadata.Dates = normalizeMetadataDates(metadata.Dates)
	metadata.Amounts = normalizeMetadataAmounts(metadata.Amounts)
	metadata.Identifiers = normalizeMetadataIdentifiers(metadata.Identifiers)
	metadata.Addresses = normalizeStringList(metadata.Addresses)
	metadata.Keywords = normalizeStringList(metadata.Keywords)
}

func normalizeMetadataParties(items []MetadataParty) []MetadataParty {
	if len(items) == 0 {
		return []MetadataParty{}
	}
	normalized := make([]MetadataParty, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		normalized = append(normalized, MetadataParty{
			Name: name,
			Role: strings.TrimSpace(item.Role),
		})
	}
	return normalized
}

func normalizeMetadataDates(items []MetadataDate) []MetadataDate {
	if len(items) == 0 {
		return []MetadataDate{}
	}
	normalized := make([]MetadataDate, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		value := strings.TrimSpace(item.Value)
		if label == "" || value == "" {
			continue
		}
		normalized = append(normalized, MetadataDate{Label: label, Value: value})
	}
	return normalized
}

func normalizeMetadataAmounts(items []MetadataAmount) []MetadataAmount {
	if len(items) == 0 {
		return []MetadataAmount{}
	}
	normalized := make([]MetadataAmount, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		if label == "" {
			continue
		}
		normalized = append(normalized, MetadataAmount{
			Label:    label,
			Value:    item.Value,
			Currency: strings.TrimSpace(item.Currency),
		})
	}
	return normalized
}

func normalizeMetadataIdentifiers(items []MetadataIdentifier) []MetadataIdentifier {
	if len(items) == 0 {
		return []MetadataIdentifier{}
	}
	normalized := make([]MetadataIdentifier, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		value := strings.TrimSpace(item.Value)
		if label == "" || value == "" {
			continue
		}
		normalized = append(normalized, MetadataIdentifier{Label: label, Value: value})
	}
	return normalized
}

func normalizeStringList(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	normalized := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, item)
	}
	return normalized
}

func toDocumentAIMetadataResponse(record *DocumentAIMetadata) (*DocumentAIMetadataResponse, error) {
	var metadata ExtractedDocumentMetadata
	if len(record.Metadata) > 0 {
		if err := json.Unmarshal(record.Metadata, &metadata); err != nil {
			return nil, err
		}
	}

	return &DocumentAIMetadataResponse{
		DocumentID:   record.DocumentID,
		DocumentType: record.DocumentType,
		Metadata:     metadata,
		Confidence:   record.Confidence,
		Model:        record.Model,
		ExtractedBy:  record.ExtractedBy,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
		RawJSON:      json.RawMessage(record.Metadata),
	}, nil
}

func (s *Service) logAIMetadataViewed(ctx context.Context, authContext auth.Context, documentID uuid.UUID) {
	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, authContext.TenantID, &userID, audit.ActionAIMetadataViewed, documentID, "ai metadata viewed", nil)
}
