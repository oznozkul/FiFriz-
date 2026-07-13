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
	ErrDocumentEntityNotFound      = errors.New("document entity not found")
	ErrAIEntitiesProviderDisabled  = errors.New("entities provider is disabled")
	ErrAIEntitiesNoContent         = errors.New("document has no extractable content for entities")
	ErrAIEntitiesParseFailed       = errors.New("entities response parse failed")
)

const aiEntitiesContextMaxRunes = 12000

var aiEntitiesJSONFencePattern = regexp.MustCompile("(?s)```(?:json)?\\s*(\\[.*\\])\\s*```")

var validEntityTypes = map[string]struct{}{
	"person":            {},
	"company":           {},
	"organization":      {},
	"email":             {},
	"phone":             {},
	"address":           {},
	"tax_number":        {},
	"identity_number":   {},
	"passport_number":   {},
	"license_plate":     {},
	"vehicle_vin":       {},
	"invoice_number":    {},
	"contract_number":   {},
	"court_case_number": {},
	"date":              {},
	"amount":            {},
	"currency":          {},
	"product":           {},
	"location":          {},
	"custom":            {},
}

type EntityGenerateRequest struct {
	TenantID   uuid.UUID
	DocumentID uuid.UUID
	UserID     uuid.UUID
}

type aiEntityLLMItem struct {
	EntityType string   `json:"entityType"`
	Value      string   `json:"value"`
	Label      string   `json:"label"`
	Confidence *float64 `json:"confidence"`
	SourceText string   `json:"sourceText"`
	PageNumber *int     `json:"pageNumber"`
}

func (s *Service) ListDocumentEntities(ctx context.Context, documentIDRaw string) ([]DocumentEntityResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, document, err := s.loadDocumentForEntities(ctx, authContext, documentIDRaw)
	if err != nil {
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	rows, err := s.repository.ListEntitiesByDocument(ctx, authContext.TenantID, documentID)
	if err != nil {
		return nil, err
	}

	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, authContext.TenantID, &userID, audit.ActionAIEntitiesViewed, documentID, "ai entities viewed", map[string]interface{}{
		"count": len(rows),
	})

	return toDocumentEntityResponses(rows), nil
}

func (s *Service) ExtractAIDocumentEntities(ctx context.Context, documentIDRaw string) ([]DocumentEntityResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, document, err := s.loadDocumentForEntities(ctx, authContext, documentIDRaw)
	if err != nil {
		return nil, err
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	return s.GenerateEntitiesForRequest(ctx, EntityGenerateRequest{
		TenantID:   authContext.TenantID,
		DocumentID: documentID,
		UserID:     authContext.UserID,
	})
}

func (s *Service) DeleteDocumentEntity(ctx context.Context, documentIDRaw, entityIDRaw string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	documentID, entityID, document, item, err := s.loadDocumentEntity(ctx, authContext, documentIDRaw, entityIDRaw)
	if err != nil {
		return err
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return err
	}

	if err := s.repository.SoftDeleteEntity(ctx, authContext.TenantID, documentID, entityID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDocumentEntityNotFound
		}
		return err
	}

	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, authContext.TenantID, &userID, audit.ActionAIEntityDeleted, documentID, "ai entity deleted", map[string]interface{}{
		"entityId":   entityID,
		"entityType": item.EntityType,
		"value":      item.Value,
	})
	return nil
}

func (s *Service) GenerateEntitiesForRequest(ctx context.Context, request EntityGenerateRequest) ([]DocumentEntityResponse, error) {
	if s.summaryChatProvider == nil {
		return nil, ErrAIEntitiesProviderDisabled
	}

	document, err := s.repository.GetByID(ctx, request.TenantID, request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	contextText, err := s.buildEntitiesContext(ctx, request.DocumentID)
	if err != nil {
		return nil, err
	}

	parsedItems, err := s.generateEntitiesFromContext(ctx, document.OriginalFileName, document.FileExtension, contextText)
	if err != nil {
		return nil, err
	}

	if err := s.repository.DeleteAIEntitiesByDocument(ctx, request.TenantID, request.DocumentID); err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	created := make([]DocumentEntity, 0, len(parsedItems))
	for _, candidate := range parsedItems {
		entityType := normalizeEntityType(candidate.EntityType)
		value := strings.TrimSpace(candidate.Value)
		if entityType == "" || value == "" {
			continue
		}

		dedupKey := entityType + "\x00" + normalizeEntityValue(value)
		if _, exists := seen[dedupKey]; exists {
			continue
		}

		exists, err := s.repository.HasEntity(ctx, request.TenantID, request.DocumentID, entityType, value)
		if err != nil {
			return nil, err
		}
		if exists {
			continue
		}

		confidence := candidate.Confidence
		if confidence != nil {
			clamped := clampConfidence(*confidence)
			confidence = &clamped
		}

		item := DocumentEntity{
			TenantID:    request.TenantID,
			DocumentID:  request.DocumentID,
			EntityType:  entityType,
			Value:       value,
			Label:       normalizeOptionalText(stringPtrOrNil(candidate.Label)),
			Confidence:  confidence,
			SourceText:  normalizeOptionalText(stringPtrOrNil(candidate.SourceText)),
			PageNumber:  candidate.PageNumber,
			ExtractedBy: EntityExtractedByAI,
		}
		if err := s.repository.CreateEntities(ctx, []DocumentEntity{item}); err != nil {
			return nil, err
		}
		seen[dedupKey] = struct{}{}
		created = append(created, item)
	}

	userID := request.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, request.TenantID, &userID, audit.ActionAIEntitiesExtracted, request.DocumentID, "ai entities extracted", map[string]interface{}{
		"count": len(created),
		"model": s.summaryChatModel,
	})

	rows, err := s.repository.ListEntitiesByDocument(ctx, request.TenantID, request.DocumentID)
	if err != nil {
		return nil, err
	}
	return toDocumentEntityResponses(rows), nil
}

func (s *Service) loadDocumentForEntities(ctx context.Context, authContext auth.Context, documentIDRaw string) (uuid.UUID, *Document, error) {
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

func (s *Service) loadDocumentEntity(ctx context.Context, authContext auth.Context, documentIDRaw, entityIDRaw string) (uuid.UUID, uuid.UUID, *Document, *DocumentEntity, error) {
	documentID, document, err := s.loadDocumentForEntities(ctx, authContext, documentIDRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, nil, err
	}

	entityID, err := uuid.Parse(strings.TrimSpace(entityIDRaw))
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, nil, ErrDocumentEntityNotFound
	}

	item, err := s.repository.GetEntityByID(ctx, authContext.TenantID, documentID, entityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, uuid.Nil, nil, nil, ErrDocumentEntityNotFound
		}
		return uuid.Nil, uuid.Nil, nil, nil, err
	}
	return documentID, entityID, document, item, nil
}

func (s *Service) buildEntitiesContext(ctx context.Context, documentID uuid.UUID) (string, error) {
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
			if usedRunes+chunkRunes > aiEntitiesContextMaxRunes {
				remaining := aiEntitiesContextMaxRunes - usedRunes
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
			if usedRunes >= aiEntitiesContextMaxRunes {
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
			return "", ErrAIEntitiesNoContent
		}
		return "", err
	}

	fullText := strings.TrimSpace(documentText.FullText)
	if fullText == "" {
		return "", ErrAIEntitiesNoContent
	}

	if utf8.RuneCountInString(fullText) > aiEntitiesContextMaxRunes {
		fullText = string([]rune(fullText)[:aiEntitiesContextMaxRunes])
	}
	return fullText, nil
}

func (s *Service) generateEntitiesFromContext(ctx context.Context, fileName, fileExtension, contextText string) ([]aiEntityLLMItem, error) {
	completion, err := s.summaryChatProvider.Complete(ctx, aiEntitiesSystemPrompt(), buildEntitiesUserPrompt(fileName, fileExtension, contextText))
	if err != nil {
		log.Printf("document entity extraction failed: %v", err)
		if errors.Is(err, ErrSummaryProviderUnavailable) {
			return nil, ErrSummaryProviderUnavailable
		}
		return nil, err
	}

	parsed, err := parseEntitiesLLMResponse(completion)
	if err != nil {
		log.Printf("document entity parse failed: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrAIEntitiesParseFailed, err)
	}
	return parsed, nil
}

func aiEntitiesSystemPrompt() string {
	return strings.TrimSpace(`
Sen FiFriz belge varlık çıkarma asistanısın.
Yalnızca verilen belge içeriğine dayanarak önemli varlıkları çıkar.
Bağlamda olmayan bilgileri uydurma.
entityType değerini şu listeden seç:
person, company, organization, email, phone, address, tax_number, identity_number, passport_number, license_plate, vehicle_vin, invoice_number, contract_number, court_case_number, date, amount, currency, product, location, custom
Yanıtını yalnızca geçerli JSON array olarak ver:
[
  {
    "entityType": "person",
    "value": "Ali Ozan Özkul",
    "label": "Kiracı",
    "confidence": 0.92,
    "sourceText": "Kiracı Ali Ozan Özkul..."
  },
  {
    "entityType": "amount",
    "value": "25000",
    "label": "Kira Bedeli",
    "confidence": 0.88,
    "sourceText": "Aylık kira bedeli 25.000 TL..."
  }
]
`)
}

func buildEntitiesUserPrompt(fileName, fileExtension, contextText string) string {
	title := strings.TrimSpace(fileName)
	if title == "" {
		title = "adsız belge"
	}
	extension := strings.TrimSpace(fileExtension)
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	return fmt.Sprintf("Belge adı: %s%s\n\nBelge içeriği:\n%s", title, extension, contextText)
}

func parseEntitiesLLMResponse(completion string) ([]aiEntityLLMItem, error) {
	raw := strings.TrimSpace(completion)
	if raw == "" {
		return nil, fmt.Errorf("empty completion")
	}

	jsonPayload := raw
	if match := aiEntitiesJSONFencePattern.FindStringSubmatch(raw); len(match) > 1 {
		jsonPayload = strings.TrimSpace(match[1])
	} else {
		start := strings.Index(raw, "[")
		end := strings.LastIndex(raw, "]")
		if start >= 0 && end > start {
			jsonPayload = raw[start : end+1]
		}
	}

	var parsed []aiEntityLLMItem
	if err := json.Unmarshal([]byte(jsonPayload), &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

func normalizeEntityType(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")

	if _, ok := validEntityTypes[normalized]; ok {
		return normalized
	}

	aliases := map[string]string{
		"person_name":      "person",
		"individual":       "person",
		"firma":            "company",
		"business":         "company",
		"org":              "organization",
		"telephone":        "phone",
		"mobile":           "phone",
		"tc_no":            "identity_number",
		"tc_kimlik":        "identity_number",
		"vergi_no":         "tax_number",
		"plate":            "license_plate",
		"vin":              "vehicle_vin",
		"invoice_no":       "invoice_number",
		"contract_no":      "contract_number",
		"case_number":      "court_case_number",
		"money":            "amount",
		"price":            "amount",
		"place":            "location",
	}

	if mapped, ok := aliases[normalized]; ok {
		return mapped
	}
	return "custom"
}

func normalizeEntityValue(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}
