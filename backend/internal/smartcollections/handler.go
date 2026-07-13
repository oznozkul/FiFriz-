package smartcollections

import (
	"errors"

	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	items, err := h.service.ListCollections(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("smart_collections_list_failed", "could not list smart collections"))
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) ListDocuments(c *fiber.Ctx) error {
	items, err := h.service.ListDocuments(c.UserContext(), c.Params("key"))
	if err != nil {
		if errors.Is(err, ErrInvalidCollectionKey) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("smart_collection_not_found", "smart collection not found"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("smart_collection_documents_failed", "could not list smart collection documents"))
	}
	return c.JSON(response.OK(items))
}
