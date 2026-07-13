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
	ErrImportantDateSuggestionNotFound  = errors.New("important date suggestion not found")
	ErrImportantDateSuggestionInvalid   = errors.New("important date suggestion is not pending")
	ErrAIImportantDatesProviderDisabled = errors.New("important dates provider is disabled")
	ErrAIImportantDatesNoContent        = errors.New("document has no extractable content for important dates")
	ErrAIImportantDatesParseFailed      = errors.New("important dates response parse failed")
)

const aiImportantDatesContextMaxRunes = 12000

var aiImportantDatesJSONFencePattern = regexp.MustCompile("(?s)```(?:json)?\\s*(\\[.*\\])\\s*```")

type AIImportantDateGenerateRequest struct {
	TenantID   uuid.UUID
	DocumentID uuid.UUID
	UserID     uuid.UUID
}

type aiImportantDateLLMItem struct {
	Title              string   `json:"title"`
	Date               string   `json:"date"`
	Type               string   `json:"type"`
	ReminderDaysBefore *int     `json:"reminderDaysBefore"`
	SourceText         string   `json:"sourceText"`
	Confidence         *float64 `json:"confidence"`
}

func (s *Service) ListImportantDateSuggestions(ctx context.Context, documentIDRaw string) ([]DocumentImportantDateResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, document, err := s.loadDocumentForAIImportantDates(ctx, authContext, documentIDRaw)
	if err != nil {
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	rows, err := s.repository.ListSuggestedImportantDatesByDocument(ctx, authContext.TenantID, documentID)
	if err != nil {
		return nil, err
	}
	return toImportantDateResponses(rows), nil
}

func (s *Service) ExtractAIImportantDates(ctx context.Context, documentIDRaw string) ([]DocumentImportantDateResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, document, err := s.loadDocumentForAIImportantDates(ctx, authContext, documentIDRaw)
	if err != nil {
		return nil, err
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	return s.GenerateAIImportantDatesForRequest(ctx, AIImportantDateGenerateRequest{
		TenantID:   authContext.TenantID,
		DocumentID: documentID,
		UserID:     authContext.UserID,
	})
}

func (s *Service) AcceptImportantDateSuggestion(ctx context.Context, documentIDRaw, suggestionIDRaw string) (*DocumentImportantDateResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, suggestionID, document, item, err := s.loadSuggestedImportantDate(ctx, authContext, documentIDRaw, suggestionIDRaw)
	if err != nil {
		return nil, err
	}

	item.Status = ImportantDateStatusActive
	if err := s.repository.UpdateImportantDate(ctx, item); err != nil {
		return nil, err
	}

	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, authContext.TenantID, &userID, audit.ActionAIImportantDateAccepted, documentID, "ai important date accepted", map[string]interface{}{
		"suggestionId": suggestionID,
		"title":        item.Title,
	})

	row, err := s.repository.GetImportantDateWithDetails(ctx, authContext.TenantID, documentID, item.ID)
	if err != nil {
		response := toDocumentImportantDateResponse(ImportantDateListRow{DocumentImportantDate: *item, DocumentName: document.OriginalFileName})
		return &response, nil
	}
	response := toDocumentImportantDateResponse(*row)
	return &response, nil
}

func (s *Service) RejectImportantDateSuggestion(ctx context.Context, documentIDRaw, suggestionIDRaw string) (*DocumentImportantDateResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, suggestionID, document, item, err := s.loadSuggestedImportantDate(ctx, authContext, documentIDRaw, suggestionIDRaw)
	if err != nil {
		return nil, err
	}

	item.Status = ImportantDateStatusRejected
	if err := s.repository.UpdateImportantDate(ctx, item); err != nil {
		return nil, err
	}

	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, authContext.TenantID, &userID, audit.ActionAIImportantDateRejected, documentID, "ai important date rejected", map[string]interface{}{
		"suggestionId": suggestionID,
		"title":        item.Title,
	})

	row, err := s.repository.GetImportantDateWithDetails(ctx, authContext.TenantID, documentID, item.ID)
	if err != nil {
		response := toDocumentImportantDateResponse(ImportantDateListRow{DocumentImportantDate: *item, DocumentName: document.OriginalFileName})
		return &response, nil
	}
	response := toDocumentImportantDateResponse(*row)
	return &response, nil
}

func (s *Service) GenerateAIImportantDatesForRequest(ctx context.Context, request AIImportantDateGenerateRequest) ([]DocumentImportantDateResponse, error) {
	if s.summaryChatProvider == nil {
		return nil, ErrAIImportantDatesProviderDisabled
	}

	document, err := s.repository.GetByID(ctx, request.TenantID, request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	contextText, metadataHint, err := s.buildAIImportantDatesContext(ctx, request.TenantID, request.DocumentID)
	if err != nil {
		return nil, err
	}

	parsedItems, err := s.generateAIImportantDatesFromContext(ctx, document.OriginalFileName, document.FileExtension, contextText, metadataHint)
	if err != nil {
		return nil, err
	}
	if len(parsedItems) == 0 {
		return []DocumentImportantDateResponse{}, nil
	}

	if err := s.repository.DeleteSuggestedImportantDatesByDocument(ctx, request.TenantID, request.DocumentID); err != nil {
		return nil, err
	}

	createdByUserID := request.UserID
	if createdByUserID == uuid.Nil {
		createdByUserID = document.UploadedByUserID
	}

	created := make([]DocumentImportantDate, 0, len(parsedItems))
	for _, candidate := range parsedItems {
		parsedDate, err := parseImportantDate(candidate.Date)
		if err != nil {
			continue
		}
		title := strings.TrimSpace(candidate.Title)
		if title == "" {
			continue
		}

		exists, err := s.repository.HasSimilarImportantDate(ctx, request.TenantID, request.DocumentID, title, parsedDate)
		if err != nil {
			return nil, err
		}
		if exists {
			continue
		}

		dateType := normalizeImportantDateType(stringPtrOrNil(candidate.Type))
		if dateType != nil {
			if err := validateImportantDateType(dateType); err != nil {
				dateType = stringPtr(ImportantDateTypeCustom)
			}
		}

		confidence := candidate.Confidence
		if confidence != nil {
			clamped := clampConfidence(*confidence)
			confidence = &clamped
		}

		item := DocumentImportantDate{
			TenantID:           request.TenantID,
			DocumentID:         request.DocumentID,
			CreatedByUserID:    createdByUserID,
			Title:              title,
			Date:               parsedDate,
			ReminderDaysBefore: candidate.ReminderDaysBefore,
			Type:               dateType,
			Source:             ImportantDateSourceAI,
			Status:             ImportantDateStatusSuggested,
			Confidence:         confidence,
			SourceText:         normalizeOptionalText(stringPtrOrNil(candidate.SourceText)),
		}
		if err := s.repository.CreateImportantDate(ctx, &item); err != nil {
			return nil, err
		}
		created = append(created, item)
	}

	userID := request.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, request.TenantID, &userID, audit.ActionAIImportantDatesExtracted, request.DocumentID, "ai important dates extracted", map[string]interface{}{
		"count": len(created),
		"model": s.summaryChatModel,
	})

	rows, err := s.repository.ListSuggestedImportantDatesByDocument(ctx, request.TenantID, request.DocumentID)
	if err != nil {
		return nil, err
	}
	return toImportantDateResponses(rows), nil
}

func (s *Service) loadDocumentForAIImportantDates(ctx context.Context, authContext auth.Context, documentIDRaw string) (uuid.UUID, *Document, error) {
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

func (s *Service) loadSuggestedImportantDate(ctx context.Context, authContext auth.Context, documentIDRaw, suggestionIDRaw string) (uuid.UUID, uuid.UUID, *Document, *DocumentImportantDate, error) {
	documentID, document, err := s.loadDocumentForAIImportantDates(ctx, authContext, documentIDRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, nil, err
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return uuid.Nil, uuid.Nil, nil, nil, err
	}

	suggestionID, err := uuid.Parse(strings.TrimSpace(suggestionIDRaw))
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, nil, ErrImportantDateSuggestionNotFound
	}

	item, err := s.repository.GetImportantDateByID(ctx, authContext.TenantID, documentID, suggestionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, uuid.Nil, nil, nil, ErrImportantDateSuggestionNotFound
		}
		return uuid.Nil, uuid.Nil, nil, nil, err
	}
	if item.Status != ImportantDateStatusSuggested {
		return uuid.Nil, uuid.Nil, nil, nil, ErrImportantDateSuggestionInvalid
	}
	return documentID, suggestionID, document, item, nil
}

func (s *Service) buildAIImportantDatesContext(ctx context.Context, tenantID, documentID uuid.UUID) (string, string, error) {
	chunks, err := s.repository.ListChunksByDocumentID(ctx, documentID)
	if err != nil {
		return "", "", err
	}

	var builder strings.Builder
	usedRunes := 0
	if len(chunks) > 0 {
		for _, chunk := range chunks {
			content := strings.TrimSpace(chunk.Content)
			if content == "" {
				continue
			}
			chunkRunes := utf8.RuneCountInString(content)
			if usedRunes+chunkRunes > aiImportantDatesContextMaxRunes {
				remaining := aiImportantDatesContextMaxRunes - usedRunes
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
			if usedRunes >= aiImportantDatesContextMaxRunes {
				break
			}
		}
	}

	if builder.Len() == 0 {
		documentText, err := s.repository.GetTextByDocumentID(ctx, documentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", "", ErrAIImportantDatesNoContent
			}
			return "", "", err
		}
		fullText := strings.TrimSpace(documentText.FullText)
		if fullText == "" {
			return "", "", ErrAIImportantDatesNoContent
		}
		if utf8.RuneCountInString(fullText) > aiImportantDatesContextMaxRunes {
			fullText = string([]rune(fullText)[:aiImportantDatesContextMaxRunes])
		}
		builder.WriteString(fullText)
	}

	metadataHint := ""
	if record, err := s.repository.GetAIMetadataByDocumentID(ctx, tenantID, documentID); err == nil && len(record.Metadata) > 0 {
		metadataHint = string(record.Metadata)
	}

	return builder.String(), metadataHint, nil
}

func (s *Service) generateAIImportantDatesFromContext(ctx context.Context, fileName, fileExtension, contextText, metadataHint string) ([]aiImportantDateLLMItem, error) {
	completion, err := s.summaryChatProvider.Complete(ctx, aiImportantDatesSystemPrompt(), buildAIImportantDatesUserPrompt(fileName, fileExtension, contextText, metadataHint))
	if err != nil {
		log.Printf("document important dates generation failed: %v", err)
		if errors.Is(err, ErrSummaryProviderUnavailable) {
			return nil, ErrAIImportantDatesProviderDisabled
		}
		return nil, err
	}

	parsed, err := parseAIImportantDatesLLMResponse(completion)
	if err != nil {
		log.Printf("document important dates parse failed: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrAIImportantDatesParseFailed, err)
	}
	return parsed, nil
}

func aiImportantDatesSystemPrompt() string {
	return strings.TrimSpace(`
Sen FiFriz belge önemli tarih çıkarma asistanısın.
Yalnızca verilen belge içeriğine dayanarak önemli tarihleri bul.
Örnekler: sözleşme bitiş tarihi, fatura son ödeme tarihi, garanti bitiş tarihi, sigorta yenileme tarihi, pasaport geçerlilik tarihi.
Bağlamda olmayan tarihleri uydurma.
Tarihleri ISO 8601 (YYYY-MM-DD) formatında ver.
type değeri şunlardan biri olmalı: expiration, renewal, payment_due, review, custom.
Yanıtını yalnızca geçerli JSON dizisi olarak ver:
[
  {
    "title": "Sözleşme Bitiş Tarihi",
    "date": "2025-12-31",
    "type": "expiration",
    "reminderDaysBefore": 30,
    "sourceText": "...",
    "confidence": 0.85
  }
]
`)
}

func buildAIImportantDatesUserPrompt(fileName, fileExtension, contextText, metadataHint string) string {
	title := strings.TrimSpace(fileName)
	if title == "" {
		title = "Belge"
	}
	extension := strings.TrimSpace(fileExtension)
	if extension != "" {
		title = fmt.Sprintf("%s (%s)", title, strings.ToUpper(extension))
	}

	metadataSection := ""
	if strings.TrimSpace(metadataHint) != "" {
		metadataSection = "\n\nMEVCUT AI METADATA (yardımcı bağlam):\n" + metadataHint
	}

	return fmt.Sprintf(`Aşağıdaki belgeden önemli tarihleri çıkar.

BELGE ADI: %s

BELGE İÇERİĞİ:
%s%s

Yalnızca JSON dizisi döndür.`, title, contextText, metadataSection)
}

func parseAIImportantDatesLLMResponse(content string) ([]aiImportantDateLLMItem, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("empty llm response")
	}

	candidates := []string{content}
	if match := aiImportantDatesJSONFencePattern.FindStringSubmatch(content); len(match) == 2 {
		candidates = append([]string{match[1]}, candidates...)
	}
	if start := strings.Index(content, "["); start >= 0 {
		if end := strings.LastIndex(content, "]"); end > start {
			candidates = append([]string{content[start : end+1]}, candidates...)
		}
	}

	var lastErr error
	for _, candidate := range candidates {
		var parsed []aiImportantDateLLMItem
		if err := json.Unmarshal([]byte(candidate), &parsed); err != nil {
			lastErr = err
			continue
		}
		return normalizeAIImportantDateItems(parsed), nil
	}
	return nil, lastErr
}

func normalizeAIImportantDateItems(items []aiImportantDateLLMItem) []aiImportantDateLLMItem {
	if len(items) == 0 {
		return []aiImportantDateLLMItem{}
	}

	normalized := make([]aiImportantDateLLMItem, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		date := strings.TrimSpace(item.Date)
		if title == "" || date == "" {
			continue
		}
		key := strings.ToLower(title) + "|" + date
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, aiImportantDateLLMItem{
			Title:              title,
			Date:               date,
			Type:               strings.TrimSpace(item.Type),
			ReminderDaysBefore: item.ReminderDaysBefore,
			SourceText:         strings.TrimSpace(item.SourceText),
			Confidence:         item.Confidence,
		})
	}
	return normalized
}

func stringPtrOrNil(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func stringPtr(value string) *string {
	return &value
}
