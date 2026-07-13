package documents

import (
	"errors"

	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ListRelationships(c *fiber.Ctx) error {
	result, err := h.service.ListRelationships(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeRelationshipError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) CreateRelationship(c *fiber.Ctx) error {
	var request CreateDocumentRelationshipRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "request body is invalid"))
	}

	result, err := h.service.CreateRelationship(c.UserContext(), c.Params("id"), request)
	if err != nil {
		return writeRelationshipError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(response.OK(result))
}

func (h *Handler) DeleteRelationship(c *fiber.Ctx) error {
	if err := h.service.DeleteRelationship(c.UserContext(), c.Params("id")); err != nil {
		return writeRelationshipError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func writeRelationshipError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidDocumentID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
	case errors.Is(err, ErrInvalidRelationshipID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_relationship_id", "relationship id must be a valid UUID"))
	case errors.Is(err, ErrInvalidRelationshipType):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_relationship_type", "relationship type is not supported"))
	case errors.Is(err, ErrSelfDocumentRelationship):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("self_document_relationship", "document cannot be related to itself"))
	case errors.Is(err, ErrDuplicateDocumentRelationship):
		return c.Status(fiber.StatusConflict).JSON(response.Fail("duplicate_document_relationship", "this relationship already exists"))
	case errors.Is(err, ErrDocumentNotFound), errors.Is(err, ErrDocumentRelationshipNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("not_found", "resource not found"))
	case errors.Is(err, ErrForbidden):
		return response.WriteForbidden(c, "forbidden", "you do not have permission for this action")
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("document_relationship_failed", "could not process document relationship"))
	}
}
