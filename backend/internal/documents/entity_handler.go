package documents

import (
	"errors"

	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ListDocumentEntities(c *fiber.Ctx) error {
	items, err := h.service.ListDocumentEntities(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeEntityError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) ExtractAIDocumentEntities(c *fiber.Ctx) error {
	items, err := h.service.ExtractAIDocumentEntities(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeAIEntityError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) DeleteDocumentEntity(c *fiber.Ctx) error {
	if err := h.service.DeleteDocumentEntity(c.UserContext(), c.Params("id"), c.Params("entityId")); err != nil {
		return writeEntityError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func writeAIEntityError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrAIEntitiesProviderDisabled):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("entities_provider_disabled", "entities service is not configured"))
	case errors.Is(err, ErrSummaryProviderUnavailable):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("entities_provider_unavailable", "AI varlık çıkarma servisi şu anda kullanılamıyor"))
	case errors.Is(err, ErrAIEntitiesNoContent):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("entities_no_content", "document has no extractable content for entities"))
	case errors.Is(err, ErrAIEntitiesParseFailed):
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Fail("entities_parse_failed", "AI varlık yanıtı işlenemedi"))
	default:
		return writeEntityError(c, err)
	}
}

func writeEntityError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidDocumentID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
	case errors.Is(err, ErrDocumentNotFound), errors.Is(err, ErrDocumentEntityNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("not_found", "resource not found"))
	case errors.Is(err, ErrForbidden):
		return response.WriteForbidden(c, "forbidden", "you do not have permission for this action")
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("entity_failed", "could not process document entity"))
	}
}
