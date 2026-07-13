package documents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"fifriz/internal/audit"
	"fifriz/internal/auth"
	domain "fifriz/internal/events/events"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrClassificationProviderDisabled = errors.New("classification provider is disabled")
	ErrClassificationNoContent        = errors.New("document has no extractable content for classification")
	ErrClassificationParseFailed      = errors.New("classification response parse failed")
	ErrDocumentNotClassified          = errors.New("document is not classified")
)

var validAIDocumentTypes = map[string]struct{}{
	"invoice":              {},
	"contract":             {},
	"court_decision":       {},
	"vehicle_registration": {},
	"identity_document":    {},
	"passport":             {},
	"insurance_policy":     {},
	"warranty_document":    {},
	"bank_receipt":         {},
	"academic_article":     {},
	"hr_document":          {},
	"medical_document":     {},
	"tax_document":         {},
	"certificate":          {},
	"other":                {},
}

const aiClassificationContextMaxRunes = 12000

var aiClassificationJSONFencePattern = regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*\\})\\s*```")

func (s *Service) GetDocumentClassification(ctx context.Context, documentID string) (*DocumentClassificationResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, document, err := s.loadDocumentForClassification(ctx, authContext, documentID)
	if err != nil {
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	s.logClassificationViewed(ctx, authContext, parsedDocumentID)
	return toDocumentClassificationResponse(document), nil
}

func (s *Service) GenerateDocumentClassification(ctx context.Context, documentID string) (*DocumentClassificationResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, document, err := s.loadDocumentForClassification(ctx, authContext, documentID)
	if err != nil {
		return nil, err
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	return s.GenerateClassificationForRequest(ctx, ClassificationGenerateRequest{
		TenantID:   authContext.TenantID,
		DocumentID: parsedDocumentID,
		UserID:     authContext.UserID,
	})
}

func (s *Service) GenerateClassificationForRequest(ctx context.Context, request ClassificationGenerateRequest) (*DocumentClassificationResponse, error) {
	if s.summaryChatProvider == nil {
		return nil, ErrClassificationProviderDisabled
	}

	document, err := s.repository.GetByID(ctx, request.TenantID, request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	contextText, err := s.buildClassificationContext(ctx, request.DocumentID)
	if err != nil {
		return nil, err
	}

	parsed, rawJSON, err := s.generateClassificationFromContext(ctx, document.OriginalFileName, document.FileExtension, contextText)
	if err != nil {
		return nil, err
	}

	documentType := normalizeAIDocumentType(parsed.DocumentType)
	confidence := parsed.Confidence
	if confidence != nil {
		clamped := clampConfidence(*confidence)
		confidence = &clamped
	}

	reason := strings.TrimSpace(parsed.Reason)
	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}

	now := time.Now().UTC()
	document.AIDocumentType = &documentType
	document.AIDocumentTypeConfidence = confidence
	document.AIClassifiedAt = &now
	document.AIClassificationReason = reasonPtr

	if err := s.repository.Update(ctx, document); err != nil {
		return nil, err
	}

	userID := request.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, request.TenantID, &userID, audit.ActionAIDocumentClassified, request.DocumentID, "ai document classified", map[string]interface{}{
		"model":        s.summaryChatModel,
		"documentType": documentType,
		"confidence":   confidence,
		"rawSize":      len(rawJSON),
	})

	s.publisher.Publish(ctx, domain.NewDocumentClassificationCompleted(request.TenantID, request.DocumentID, userID))

	return toDocumentClassificationResponse(document), nil
}

func (s *Service) loadDocumentForClassification(ctx context.Context, authContext auth.Context, documentID string) (uuid.UUID, *Document, error) {
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

func (s *Service) buildClassificationContext(ctx context.Context, documentID uuid.UUID) (string, error) {
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
			if usedRunes+chunkRunes > aiClassificationContextMaxRunes {
				remaining := aiClassificationContextMaxRunes - usedRunes
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
			if usedRunes >= aiClassificationContextMaxRunes {
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
			return "", ErrClassificationNoContent
		}
		return "", err
	}

	fullText := strings.TrimSpace(documentText.FullText)
	if fullText == "" {
		return "", ErrClassificationNoContent
	}

	if utf8.RuneCountInString(fullText) > aiClassificationContextMaxRunes {
		fullText = string([]rune(fullText)[:aiClassificationContextMaxRunes])
	}

	return fullText, nil
}

func (s *Service) generateClassificationFromContext(ctx context.Context, fileName, fileExtension, contextText string) (*extractedDocumentClassification, []byte, error) {
	completion, err := s.summaryChatProvider.Complete(ctx, aiClassificationSystemPrompt(), buildClassificationUserPrompt(fileName, fileExtension, contextText))
	if err != nil {
		log.Printf("document classification failed: %v", err)
		if errors.Is(err, ErrSummaryProviderUnavailable) {
			return nil, nil, ErrSummaryProviderUnavailable
		}
		return nil, nil, err
	}

	parsed, rawJSON, err := parseClassificationLLMResponse(completion)
	if err != nil {
		log.Printf("document classification parse failed: %v", err)
		return nil, rawJSON, fmt.Errorf("%w: %v", ErrClassificationParseFailed, err)
	}

	return parsed, rawJSON, nil
}

func aiClassificationSystemPrompt() string {
	return strings.TrimSpace(`
Sen FiFriz belge sınıflandırma asistanısın.
Yalnızca verilen belge içeriğine dayanarak belge türünü tahmin et.
Bağlamda olmayan bilgileri uydurma.
Belge türünü aşağıdaki listeden birini seçerek ver:
invoice, contract, court_decision, vehicle_registration, identity_document, passport, insurance_policy, warranty_document, bank_receipt, academic_article, hr_document, medical_document, tax_document, certificate, other
Yanıtını yalnızca geçerli JSON olarak ver:
{
  "documentType": "contract",
  "confidence": 0.91,
  "reason": "Belgede taraflar, sözleşme başlangıç/bitiş tarihleri ve kira bedeli geçiyor."
}
`)
}

func buildClassificationUserPrompt(fileName, fileExtension, contextText string) string {
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

func parseClassificationLLMResponse(completion string) (*extractedDocumentClassification, []byte, error) {
	raw := strings.TrimSpace(completion)
	if raw == "" {
		return nil, nil, fmt.Errorf("empty completion")
	}

	jsonPayload := raw
	if match := aiClassificationJSONFencePattern.FindStringSubmatch(raw); len(match) > 1 {
		jsonPayload = strings.TrimSpace(match[1])
	} else {
		start := strings.Index(raw, "{")
		end := strings.LastIndex(raw, "}")
		if start >= 0 && end > start {
			jsonPayload = raw[start : end+1]
		}
	}

	var parsed extractedDocumentClassification
	if err := json.Unmarshal([]byte(jsonPayload), &parsed); err != nil {
		return nil, []byte(jsonPayload), err
	}

	if strings.TrimSpace(parsed.DocumentType) == "" {
		return nil, []byte(jsonPayload), fmt.Errorf("documentType is required")
	}

	return &parsed, []byte(jsonPayload), nil
}

func normalizeAIDocumentType(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")

	if _, ok := validAIDocumentTypes[normalized]; ok {
		return normalized
	}

	aliases := map[string]string{
		"invoice_document": "invoice",
		"fatura":           "invoice",
		"sozlesme":         "contract",
		"sözleşme":         "contract",
		"lease":            "contract",
		"rental_agreement": "contract",
		"identity":         "identity_document",
		"id_document":      "identity_document",
		"insurance":        "insurance_policy",
		"policy":           "insurance_policy",
		"warranty":         "warranty_document",
		"garanti":          "warranty_document",
		"receipt":          "bank_receipt",
		"makbuz":           "bank_receipt",
		"tax":              "tax_document",
		"vergi":            "tax_document",
		"medical":          "medical_document",
		"hr":               "hr_document",
		"human_resources":  "hr_document",
		"academic":         "academic_article",
		"article":          "academic_article",
		"court":            "court_decision",
		"judgment":         "court_decision",
		"vehicle":          "vehicle_registration",
		"registration":     "vehicle_registration",
	}

	if mapped, ok := aliases[normalized]; ok {
		return mapped
	}

	return "other"
}

func (s *Service) logClassificationViewed(ctx context.Context, authContext auth.Context, documentID uuid.UUID) {
	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, authContext.TenantID, &userID, audit.ActionAIDocumentClassificationViewed, documentID, "ai document classification viewed", nil)
}
