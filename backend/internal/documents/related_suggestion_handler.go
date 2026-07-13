package documents

import (
	"errors"

	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ListRelatedSuggestions(c *fiber.Ctx) error {
	items, err := h.service.ListRelatedSuggestions(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeRelatedSuggestionError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) GenerateRelatedSuggestions(c *fiber.Ctx) error {
	items, err := h.service.GenerateRelatedSuggestions(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeAIRelatedSuggestionError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) AcceptRelatedSuggestion(c *fiber.Ctx) error {
	result, err := h.service.AcceptRelatedSuggestion(c.UserContext(), c.Params("id"), c.Params("suggestionId"))
	if err != nil {
		return writeRelatedSuggestionError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) RejectRelatedSuggestion(c *fiber.Ctx) error {
	result, err := h.service.RejectRelatedSuggestion(c.UserContext(), c.Params("id"), c.Params("suggestionId"))
	if err != nil {
		return writeRelatedSuggestionError(c, err)
	}
	return c.JSON(response.OK(result))
}

func writeAIRelatedSuggestionError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrAIRelatedDocumentsNoEmbeddings):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("related_documents_no_embeddings", "document has no embeddings for related document matching"))
	default:
		return writeRelatedSuggestionError(c, err)
	}
}

func writeRelatedSuggestionError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidDocumentID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
	case errors.Is(err, ErrDocumentNotFound), errors.Is(err, ErrRelatedSuggestionNotFound), errors.Is(err, ErrDocumentRelationshipNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("not_found", "resource not found"))
	case errors.Is(err, ErrRelatedSuggestionInvalidStatus):
		return c.Status(fiber.StatusConflict).JSON(response.Fail("related_suggestion_invalid", "related document suggestion is not pending"))
	case errors.Is(err, ErrForbidden):
		return response.WriteForbidden(c, "forbidden", "you do not have permission for this action")
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("related_suggestion_failed", "could not process related document suggestion"))
	}
}
