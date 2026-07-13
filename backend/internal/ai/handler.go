package ai

import (
	"errors"
	"strings"

	"fifriz/internal/ai/tools"
	"fifriz/internal/documents"
	"fifriz/internal/folders"
	"fifriz/internal/response"
	"fifriz/internal/tags"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service            *Service
	workspaceInsights  *WorkspaceInsightsService
	actionSuggestions  *ActionSuggestionService
	agentChat          *AgentChatService
}

func NewHandler(service *Service, workspaceInsights *WorkspaceInsightsService, actionSuggestions *ActionSuggestionService, agentChat *AgentChatService) *Handler {
	return &Handler{
		service:           service,
		workspaceInsights: workspaceInsights,
		actionSuggestions: actionSuggestions,
		agentChat:         agentChat,
	}
}

func (h *Handler) WorkspaceInsights(c *fiber.Ctx) error {
	items, err := h.workspaceInsights.List(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("workspace_insights_failed", "çalışma alanı içgörüleri yüklenemedi"))
	}
	if items == nil {
		items = []WorkspaceInsight{}
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) ListActionSuggestions(c *fiber.Ctx) error {
	items, err := h.actionSuggestions.ListForUser(c.UserContext())
	if err != nil {
		return h.handleActionSuggestionError(c, err)
	}
	if items == nil {
		items = []ActionSuggestionResponse{}
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) ListDocumentActionSuggestions(c *fiber.Ctx) error {
	items, err := h.actionSuggestions.ListForDocument(c.UserContext(), c.Params("id"))
	if err != nil {
		return h.handleActionSuggestionError(c, err)
	}
	if items == nil {
		items = []ActionSuggestionResponse{}
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) GenerateDocumentActionSuggestions(c *fiber.Ctx) error {
	items, err := h.actionSuggestions.GenerateForDocument(c.UserContext(), c.Params("id"))
	if err != nil {
		return h.handleActionSuggestionError(c, err)
	}
	if items == nil {
		items = []ActionSuggestionResponse{}
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) AcceptActionSuggestion(c *fiber.Ctx) error {
	result, err := h.actionSuggestions.Accept(c.UserContext(), c.Params("id"))
	if err != nil {
		return h.handleActionSuggestionError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) RejectActionSuggestion(c *fiber.Ctx) error {
	result, err := h.actionSuggestions.Reject(c.UserContext(), c.Params("id"))
	if err != nil {
		return h.handleActionSuggestionError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) handleActionSuggestionError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidDocumentID), errors.Is(err, ErrInvalidSuggestionID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_id", "geçersiz kimlik"))
	case errors.Is(err, ErrDocumentNotFound), errors.Is(err, ErrActionSuggestionNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("not_found", "kayıt bulunamadı"))
	case errors.Is(err, ErrActionSuggestionDenied), errors.Is(err, documents.ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "bu kaynağa erişim yetkiniz yok"))
	case errors.Is(err, ErrSuggestionNotPending):
		return c.Status(fiber.StatusConflict).JSON(response.Fail("suggestion_not_pending", "öneri artık beklemede değil"))
	case errors.Is(err, tools.ErrToolNotFound):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("tool_not_found", "araç bulunamadı"))
	case errors.Is(err, tools.ErrToolNotAllowed), errors.Is(err, tools.ErrInvalidToolInput):
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("tool_not_allowed", "bu işlem için yetkiniz yok"))
	case errors.Is(err, tools.ErrToolExecutionFailed), errors.Is(err, ErrToolExecutionFailed):
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Fail("tool_execution_failed", "aksiyon çalıştırılamadı"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("action_suggestion_failed", "aksiyon önerisi işlenemedi"))
	}
}

func (h *Handler) Chat(c *fiber.Ctx) error {
	var request ChatRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "geçersiz istek gövdesi"))
	}

	result, err := h.service.Chat(c.UserContext(), request)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(response.OK(result))
}

func (h *Handler) AgentChat(c *fiber.Ctx) error {
	var request AgentChatRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "geçersiz istek gövdesi"))
	}

	result, err := h.agentChat.Chat(c.UserContext(), request)
	if err != nil {
		return h.handleAgentChatError(c, err)
	}

	return c.JSON(response.OK(result))
}

func (h *Handler) CreateSession(c *fiber.Ctx) error {
	var request CreateSessionRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "geçersiz istek gövdesi"))
	}

	result, err := h.service.CreateSession(c.UserContext(), request)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(response.OK(result))
}

func (h *Handler) ListSessions(c *fiber.Ctx) error {
	scopeType := optionalQueryString(c.Query("scopeType"))
	scopeID := optionalQueryString(c.Query("scopeId"))

	result, err := h.service.ListSessions(c.UserContext(), scopeType, scopeID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(response.OK(result))
}

func (h *Handler) GetSession(c *fiber.Ctx) error {
	result, err := h.service.GetSession(c.UserContext(), c.Params("id"))
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(response.OK(result))
}

func (h *Handler) DeleteSession(c *fiber.Ctx) error {
	if err := h.service.DeleteSession(c.UserContext(), c.Params("id")); err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(response.OK(map[string]string{"status": "deleted"}))
}

func (h *Handler) SendSessionMessage(c *fiber.Ctx) error {
	var request ChatRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "geçersiz istek gövdesi"))
	}

	result, err := h.service.SendSessionMessage(c.UserContext(), c.Params("id"), request)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(response.OK(result))
}

func optionalQueryString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidQuestion):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_question", "soru boş olamaz"))
	case errors.Is(err, ErrInvalidDocumentFilter):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "geçersiz belge filtresi"))
	case errors.Is(err, ErrInvalidFolderFilter):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_folder_id", "geçersiz klasör filtresi"))
	case errors.Is(err, ErrInvalidTagFilter):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_tag_ids", "geçersiz etiket filtresi"))
	case errors.Is(err, ErrInvalidCollectionKey):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_collection_key", "geçersiz koleksiyon anahtarı"))
	case errors.Is(err, ErrInvalidSessionID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_session_id", "geçersiz sohbet kimliği"))
	case errors.Is(err, ErrInvalidScopeType):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_scope", "geçersiz sohbet kapsamı"))
	case errors.Is(err, ErrSessionNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("session_not_found", "sohbet bulunamadı"))
	case errors.Is(err, ErrEmbeddingDisabled):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("embedding_disabled", "embedding servisi kullanılamıyor"))
	case errors.Is(err, ErrChatProviderDisabled):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("chat_disabled", "sohbet servisi kullanılamıyor"))
	case errors.Is(err, documents.ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "bu kaynağa erişim yetkiniz yok"))
	case errors.Is(err, folders.ErrFolderNotFound), errors.Is(err, tags.ErrTagNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("scope_not_found", "sohbet kapsamı bulunamadı"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("chat_failed", "sohbet tamamlanamadı"))
	}
}

func (h *Handler) handleAgentChatError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidAgentMessage):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_message", "mesaj boş olamaz"))
	case errors.Is(err, ErrInvalidDocumentFilter):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "geçersiz belge filtresi"))
	case errors.Is(err, ErrInvalidFolderFilter):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_folder_id", "geçersiz klasör filtresi"))
	case errors.Is(err, ErrInvalidSessionID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_session_id", "geçersiz sohbet kimliği"))
	case errors.Is(err, ErrSessionNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("session_not_found", "sohbet bulunamadı"))
	case errors.Is(err, ErrChatProviderDisabled):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("chat_disabled", "sohbet servisi kullanılamıyor"))
	case errors.Is(err, ErrEmbeddingDisabled):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("embedding_disabled", "embedding servisi kullanılamıyor"))
	case errors.Is(err, documents.ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "bu kaynağa erişim yetkiniz yok"))
	case errors.Is(err, folders.ErrFolderNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("scope_not_found", "sohbet kapsamı bulunamadı"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("agent_chat_failed", "agent sohbeti tamamlanamadı"))
	}
}
