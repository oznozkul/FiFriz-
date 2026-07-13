package documents

import (
	"errors"

	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ListComments(c *fiber.Ctx) error {
	items, err := h.service.ListComments(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeCommentError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) CreateComment(c *fiber.Ctx) error {
	var request CreateDocumentCommentRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "request body is invalid"))
	}

	result, err := h.service.CreateComment(c.UserContext(), c.Params("id"), request)
	if err != nil {
		return writeCommentError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(response.OK(result))
}

func (h *Handler) UpdateComment(c *fiber.Ctx) error {
	var request UpdateDocumentCommentRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "request body is invalid"))
	}

	result, err := h.service.UpdateComment(c.UserContext(), c.Params("id"), c.Params("commentId"), request)
	if err != nil {
		return writeCommentError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) DeleteComment(c *fiber.Ctx) error {
	if err := h.service.DeleteComment(c.UserContext(), c.Params("id"), c.Params("commentId")); err != nil {
		return writeCommentError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func writeCommentError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidDocumentID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
	case errors.Is(err, ErrInvalidCommentID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_comment_id", "comment id must be a valid UUID"))
	case errors.Is(err, ErrInvalidParentComment):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_parent_comment", "parent comment must belong to the same document"))
	case errors.Is(err, ErrEmptyCommentContent):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("empty_comment_content", "comment content must not be empty"))
	case errors.Is(err, ErrDocumentNotFound), errors.Is(err, ErrDocumentCommentNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("not_found", "resource not found"))
	case errors.Is(err, ErrForbidden):
		return response.WriteForbidden(c, "forbidden", "you do not have permission for this action")
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_comment_failed", "could not process document comment"))
	}
}
