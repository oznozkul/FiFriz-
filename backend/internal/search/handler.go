package search

import (
	"errors"

	"fifriz/internal/documents"
	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Global(c *fiber.Ctx) error {
	result, err := h.service.Global(c.UserContext(), c.Query("q"))
	if err != nil {
		if errors.Is(err, documents.ErrInvalidSearchQuery) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_search_query", "geçerli bir arama sorgusu girin"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("global_search_failed", "global arama tamamlanamadı"))
	}
	return c.JSON(response.OK(result))
}
