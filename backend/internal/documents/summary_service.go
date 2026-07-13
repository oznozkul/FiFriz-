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
	ErrSummaryProviderDisabled    = errors.New("summary chat provider is disabled")
	ErrSummaryNoContent           = errors.New("document has no extractable content for summary")
	ErrSummaryProviderUnavailable = errors.New("summary provider unavailable")
)

const summaryContextMaxRunes = 12000

type summaryLLMResponse struct {
	Summary   string   `json:"summary"`
	KeyPoints []string `json:"keyPoints"`
}

var summaryJSONFencePattern = regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*\\})\\s*```")

func (s *Service) SetSummaryChatProvider(provider SummaryChatCompleter, modelName string) {
	s.summaryChatProvider = provider
	s.summaryChatModel = strings.TrimSpace(modelName)
}

func (s *Service) SetAuditRecorder(recorder audit.Recorder) {
	s.auditRecorder = recorder
}

func (s *Service) GetSummary(ctx context.Context, documentID string) (*DocumentSummaryResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, document, err := s.loadDocumentForSummary(ctx, authContext, documentID)
	if err != nil {
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	summary, err := s.repository.GetSummaryByDocumentID(ctx, authContext.TenantID, parsedDocumentID)
	if err != nil {
		return nil, err
	}

	s.logSummaryViewed(ctx, authContext, parsedDocumentID)
	return toDocumentSummaryResponse(summary), nil
}

func (s *Service) GenerateSummary(ctx context.Context, documentID string) (*DocumentSummaryResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, document, err := s.loadDocumentForSummary(ctx, authContext, documentID)
	if err != nil {
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	return s.GenerateSummaryForRequest(ctx, SummaryGenerateRequest{
		TenantID:   authContext.TenantID,
		DocumentID: parsedDocumentID,
		UserID:     authContext.UserID,
	})
}

func (s *Service) GenerateSummaryForRequest(ctx context.Context, request SummaryGenerateRequest) (*DocumentSummaryResponse, error) {
	if s.summaryChatProvider == nil {
		return nil, ErrSummaryProviderDisabled
	}

	document, err := s.repository.GetByID(ctx, request.TenantID, request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	contextText, sourceChunkIDs, err := s.buildSummaryContext(ctx, request.DocumentID)
	if err != nil {
		return nil, err
	}

	parsedSummary, err := s.generateSummaryFromContext(ctx, document.OriginalFileName, document.FileExtension, contextText)
	if err != nil {
		return nil, err
	}

	keyPointsJSON, err := json.Marshal(parsedSummary.KeyPoints)
	if err != nil {
		return nil, err
	}
	sourceChunkIDsJSON, err := json.Marshal(sourceChunkIDs)
	if err != nil {
		return nil, err
	}

	var modelName *string
	if s.summaryChatModel != "" {
		modelName = &s.summaryChatModel
	}

	record := &DocumentSummary{
		TenantID:          request.TenantID,
		DocumentID:        request.DocumentID,
		GeneratedByUserID: request.UserID,
		Summary:           parsedSummary.Summary,
		KeyPoints:         keyPointsJSON,
		SourceChunkIDs:    sourceChunkIDsJSON,
		Model:             modelName,
	}
	if err := s.repository.SaveDocumentSummary(ctx, record); err != nil {
		return nil, err
	}

	saved, err := s.repository.GetSummaryByDocumentID(ctx, request.TenantID, request.DocumentID)
	if err != nil {
		return nil, err
	}

	userID := request.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, request.TenantID, &userID, audit.ActionDocumentSummaryGenerated, request.DocumentID, "document summary generated", map[string]interface{}{
		"model": s.summaryChatModel,
	})

	return toDocumentSummaryResponse(saved), nil
}

func (s *Service) loadDocumentForSummary(ctx context.Context, authContext auth.Context, documentID string) (uuid.UUID, *Document, error) {
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

func (s *Service) buildSummaryContext(ctx context.Context, documentID uuid.UUID) (string, []uuid.UUID, error) {
	chunks, err := s.repository.ListChunksByDocumentID(ctx, documentID)
	if err != nil {
		return "", nil, err
	}

	sourceChunkIDs := make([]uuid.UUID, 0, len(chunks))
	if len(chunks) > 0 {
		var builder strings.Builder
		usedRunes := 0
		for _, chunk := range chunks {
			content := strings.TrimSpace(chunk.Content)
			if content == "" {
				continue
			}
			chunkRunes := utf8.RuneCountInString(content)
			if usedRunes+chunkRunes > summaryContextMaxRunes {
				remaining := summaryContextMaxRunes - usedRunes
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
			sourceChunkIDs = append(sourceChunkIDs, chunk.ID)
			usedRunes += utf8.RuneCountInString(content)
			if usedRunes >= summaryContextMaxRunes {
				break
			}
		}

		if builder.Len() > 0 {
			return builder.String(), sourceChunkIDs, nil
		}
	}

	documentText, err := s.repository.GetTextByDocumentID(ctx, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrSummaryNoContent
		}
		return "", nil, err
	}

	fullText := strings.TrimSpace(documentText.FullText)
	if fullText == "" {
		return "", nil, ErrSummaryNoContent
	}

	if utf8.RuneCountInString(fullText) > summaryContextMaxRunes {
		fullText = string([]rune(fullText)[:summaryContextMaxRunes])
	}

	return fullText, nil, nil
}

func (s *Service) generateSummaryFromContext(ctx context.Context, fileName, fileExtension, contextText string) (*summaryLLMResponse, error) {
	completion, err := s.summaryChatProvider.Complete(ctx, summarySystemPrompt(), buildSummaryUserPrompt(fileName, fileExtension, contextText))
	if err != nil {
		log.Printf("document summary generation failed: %v", err)
		return nil, err
	}

	parsed, err := parseSummaryLLMResponse(completion)
	if err != nil {
		return nil, fmt.Errorf("summary response parse failed: %w", err)
	}
	if strings.TrimSpace(parsed.Summary) == "" {
		return nil, fmt.Errorf("summary response was empty")
	}

	return parsed, nil
}

func summarySystemPrompt() string {
	return strings.TrimSpace(`
Sen FiFriz belge özet asistanısın.
Yalnızca verilen belge içeriğine dayanarak Türkçe, sade ve profesyonel bir özet üret.
Bağlamda olmayan bilgileri uydurma.
Belirsiz veya belgede geçmeyen bilgiler için "belgede açıkça belirtilmemiş" ifadesini kullan.
Önemli tarihleri varsa belirt.
Belge türü hakkında kısa bir tahmin ekle (summary metninde).
Yanıtını yalnızca geçerli JSON olarak ver:
{"summary":"...","keyPoints":["...","..."]}
`)
}

func buildSummaryUserPrompt(fileName, fileExtension, contextText string) string {
	title := strings.TrimSpace(fileName)
	if title == "" {
		title = "Belge"
	}
	extension := strings.TrimSpace(fileExtension)
	if extension != "" {
		title = fmt.Sprintf("%s (%s)", title, strings.ToUpper(extension))
	}

	return fmt.Sprintf(`Aşağıdaki belge içeriğini özetle.

BELGE ADI: %s

BELGE İÇERİĞİ:
%s

JSON formatında summary ve keyPoints döndür.`, title, contextText)
}

func parseSummaryLLMResponse(content string) (*summaryLLMResponse, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("empty llm response")
	}

	candidates := []string{content}
	if match := summaryJSONFencePattern.FindStringSubmatch(content); len(match) == 2 {
		candidates = append([]string{match[1]}, candidates...)
	}
	if start := strings.Index(content, "{"); start >= 0 {
		if end := strings.LastIndex(content, "}"); end > start {
			candidates = append([]string{content[start : end+1]}, candidates...)
		}
	}

	var lastErr error
	for _, candidate := range candidates {
		var parsed summaryLLMResponse
		if err := json.Unmarshal([]byte(candidate), &parsed); err != nil {
			lastErr = err
			continue
		}
		parsed.Summary = strings.TrimSpace(parsed.Summary)
		parsed.KeyPoints = normalizeKeyPoints(parsed.KeyPoints)
		return &parsed, nil
	}

	return nil, lastErr
}

func normalizeKeyPoints(points []string) []string {
	if len(points) == 0 {
		return []string{}
	}
	normalized := make([]string, 0, len(points))
	for _, point := range points {
		point = strings.TrimSpace(point)
		if point == "" {
			continue
		}
		normalized = append(normalized, point)
	}
	return normalized
}

func toDocumentSummaryResponse(summary *DocumentSummary) *DocumentSummaryResponse {
	response := &DocumentSummaryResponse{
		DocumentID: summary.DocumentID,
		Summary:    summary.Summary,
		KeyPoints:  []string{},
		CreatedAt:  summary.CreatedAt,
		UpdatedAt:  summary.UpdatedAt,
		Model:      summary.Model,
	}

	if len(summary.KeyPoints) > 0 {
		var keyPoints []string
		if err := json.Unmarshal(summary.KeyPoints, &keyPoints); err == nil {
			response.KeyPoints = normalizeKeyPoints(keyPoints)
		}
	}

	return response
}

func (s *Service) logSummaryViewed(ctx context.Context, authContext auth.Context, documentID uuid.UUID) {
	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.auditRecorder, authContext.TenantID, &userID, audit.ActionDocumentSummaryViewed, documentID, "document summary viewed", nil)
}
