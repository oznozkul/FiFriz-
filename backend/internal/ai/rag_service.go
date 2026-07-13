package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"fifriz/internal/auth"
	"fifriz/internal/documents"
	"fifriz/internal/embedding"
	"fifriz/internal/smartcollections"

	"github.com/google/uuid"
)

type CollectionLister interface {
	ListDocuments(ctx context.Context, tenantID, userID uuid.UUID, accessUserID *uuid.UUID, key string, now time.Time) ([]documents.Document, error)
}

type RAGService struct {
	documents   *documents.Repository
	collections CollectionLister
	embeddings  embedding.Provider
	chat        ChatProvider
	shareAccess *documents.ShareAccessService
}

type RAGServiceConfig struct {
	Documents   *documents.Repository
	Collections CollectionLister
	Embeddings  embedding.Provider
	Chat        ChatProvider
	ShareAccess *documents.ShareAccessService
}

func NewRAGService(cfg RAGServiceConfig) *RAGService {
	return &RAGService{
		documents:   cfg.Documents,
		collections: cfg.Collections,
		embeddings:  cfg.Embeddings,
		chat:        cfg.Chat,
		shareAccess: cfg.ShareAccess,
	}
}

func (s *RAGService) Chat(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	return s.generateRAGAnswer(ctx, request)
}

func (s *RAGService) generateRAGAnswer(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	question := strings.TrimSpace(request.Question)
	if question == "" {
		return nil, ErrInvalidQuestion
	}
	if s.embeddings == nil {
		return nil, ErrEmbeddingDisabled
	}
	if s.chat == nil {
		return nil, ErrChatProviderDisabled
	}

	filter, err := s.buildChunkFilter(ctx, authContext, request)
	if err != nil {
		return nil, err
	}
	if err := s.validateScopedDocumentAccess(ctx, authContext, filter); err != nil {
		return nil, err
	}

	embeddingValues, err := s.embeddings.CreateEmbedding(ctx, question)
	if err != nil {
		return nil, err
	}
	if len(embeddingValues) != embedding.DefaultEmbeddingDimensions {
		return nil, fmt.Errorf("query embedding dimension mismatch: got %d want %d", len(embeddingValues), embedding.DefaultEmbeddingDimensions)
	}

	rawResults, err := s.documents.SearchChunksByEmbeddingFiltered(
		ctx,
		authContext.TenantID,
		embeddingValues,
		ragContextChunkLimit,
		memberAccessUserID(authContext),
		filter,
	)
	if err != nil {
		return nil, err
	}

	allowedResults, err := s.filterByViewPermission(ctx, authContext, rawResults)
	if err != nil {
		return nil, err
	}
	if len(allowedResults) == 0 {
		return &ChatResponse{
			Answer:  noContextAnswer,
			Sources: []ChatSource{},
		}, nil
	}

	sources := toChatSources(allowedResults)
	completion, err := s.chat.Complete(ctx, ChatCompletionRequest{
		SystemPrompt: ragSystemPrompt(),
		UserPrompt:   buildRAGUserPrompt(question, allowedResults),
	})
	if err != nil {
		return nil, err
	}

	return &ChatResponse{
		Answer:  strings.TrimSpace(completion.Content),
		Sources: sources,
	}, nil
}

func (s *RAGService) buildChunkFilter(ctx context.Context, authContext auth.Context, request ChatRequest) (documents.ChunkSearchFilter, error) {
	filter := documents.ChunkSearchFilter{}

	if request.DocumentID != nil && strings.TrimSpace(*request.DocumentID) != "" {
		documentID, err := uuid.Parse(strings.TrimSpace(*request.DocumentID))
		if err != nil {
			return filter, ErrInvalidDocumentFilter
		}
		filter.DocumentID = &documentID
	}

	if request.FolderID != nil && strings.TrimSpace(*request.FolderID) != "" {
		folderID, err := uuid.Parse(strings.TrimSpace(*request.FolderID))
		if err != nil {
			return filter, ErrInvalidFolderFilter
		}
		filter.FolderID = &folderID
	}

	if len(request.TagIDs) > 0 {
		tagIDs := make([]uuid.UUID, 0, len(request.TagIDs))
		for _, rawTagID := range request.TagIDs {
			rawTagID = strings.TrimSpace(rawTagID)
			if rawTagID == "" {
				continue
			}
			tagID, err := uuid.Parse(rawTagID)
			if err != nil {
				return filter, ErrInvalidTagFilter
			}
			tagIDs = append(tagIDs, tagID)
		}
		if len(tagIDs) == 0 {
			return filter, ErrInvalidTagFilter
		}
		filter.TagIDs = tagIDs
	}

	if request.CollectionKey != nil && strings.TrimSpace(*request.CollectionKey) != "" {
		key := strings.TrimSpace(*request.CollectionKey)
		if !smartcollections.IsValidKey(key) {
			return filter, ErrInvalidCollectionKey
		}
		if s.collections == nil {
			return filter, ErrInvalidCollectionKey
		}

		accessUserID := memberAccessUserID(authContext)
		items, err := s.collections.ListDocuments(ctx, authContext.TenantID, authContext.UserID, accessUserID, key, time.Now().UTC())
		if err != nil {
			return filter, err
		}

		documentIDs := make([]uuid.UUID, 0, len(items))
		for _, item := range items {
			documentIDs = append(documentIDs, item.ID)
		}
		filter.DocumentIDs = documentIDs
		filter.RestrictToDocuments = true
	}

	return filter, nil
}

func (s *RAGService) validateScopedDocumentAccess(ctx context.Context, authContext auth.Context, filter documents.ChunkSearchFilter) error {
	if filter.DocumentID == nil || s.shareAccess == nil {
		return nil
	}

	document, err := s.documents.GetByID(ctx, authContext.TenantID, *filter.DocumentID)
	if err != nil {
		return err
	}
	return requireViewAccess(ctx, s.shareAccess, authContext, document)
}

func (s *RAGService) filterByViewPermission(ctx context.Context, authContext auth.Context, results []documents.SemanticSearchResult) ([]documents.SemanticSearchResult, error) {
	if s.shareAccess == nil {
		return results, nil
	}

	allowed := make([]documents.SemanticSearchResult, 0, len(results))
	documentCache := make(map[uuid.UUID]*documents.Document)

	for _, result := range results {
		document, ok := documentCache[result.DocumentID]
		if !ok {
			item, err := s.documents.GetByID(ctx, authContext.TenantID, result.DocumentID)
			if err != nil {
				return nil, err
			}
			document = item
			documentCache[result.DocumentID] = document
		}

		if err := requireViewAccess(ctx, s.shareAccess, authContext, document); err != nil {
			if errors.Is(err, documents.ErrForbidden) {
				continue
			}
			return nil, err
		}
		allowed = append(allowed, result)
	}

	return allowed, nil
}

func requireViewAccess(ctx context.Context, shareAccess *documents.ShareAccessService, authContext auth.Context, document *documents.Document) error {
	allowed, err := shareAccess.CanView(ctx, authContext, document)
	if err != nil {
		return err
	}
	if !allowed {
		return documents.ErrForbidden
	}
	return nil
}

func memberAccessUserID(authContext auth.Context) *uuid.UUID {
	if documents.HasTenantWideDocumentAccess(authContext.Role) {
		return nil
	}
	userID := authContext.UserID
	return &userID
}

func ragSystemPrompt() string {
	return strings.TrimSpace(`
Sen FiFriz belge asistanısın.
Yalnızca sana verilen belge parçalarındaki bilgilere dayanarak cevap ver.
Bağlamda olmayan bilgileri uydurma veya tahmin etme.
Yeterli bilgi yoksa bunu açıkça belirt.
Cevaplarını Türkçe ve anlaşılır şekilde yaz.
`)
}

func buildRAGUserPrompt(question string, chunks []documents.SemanticSearchResult) string {
	var builder strings.Builder
	builder.WriteString("Aşağıdaki belge parçalarını kullanarak soruyu yanıtla.\n\n")
	builder.WriteString("BELGE PARÇALARI:\n")

	for index, chunk := range chunks {
		title := strings.TrimSpace(chunk.DocumentTitle)
		if title == "" {
			title = "Belge"
		}
		builder.WriteString(fmt.Sprintf("\n[%d] Belge: %s | Parça: %d\n", index+1, title, chunk.ChunkIndex))
		builder.WriteString(chunk.Content)
		builder.WriteString("\n")
	}

	builder.WriteString("\nSORU:\n")
	builder.WriteString(question)
	return builder.String()
}

func toChatSources(chunks []documents.SemanticSearchResult) []ChatSource {
	sources := make([]ChatSource, 0, len(chunks))
	for _, chunk := range chunks {
		title := strings.TrimSpace(chunk.DocumentTitle)
		if title == "" {
			title = "Belge"
		}
		sources = append(sources, ChatSource{
			DocumentID:     chunk.DocumentID,
			DocumentTitle:  title,
			ChunkID:        chunk.ChunkID,
			ChunkIndex:     chunk.ChunkIndex,
			Similarity:     chunk.Similarity,
			ContentPreview: previewContent(chunk.Content),
		})
	}
	return sources
}

func previewContent(content string) string {
	content = strings.Join(strings.Fields(content), " ")
	if content == "" {
		return ""
	}
	if utf8.RuneCountInString(content) <= contentPreviewMaxRunes {
		return content
	}

	runes := []rune(content)
	if len(runes) <= contentPreviewMaxRunes {
		return content
	}
	return string(runes[:contentPreviewMaxRunes]) + "..."
}
