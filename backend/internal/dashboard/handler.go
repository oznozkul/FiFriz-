package dashboard

import (
	"fifriz/internal/auth"
	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Summary(c *fiber.Ctx) error {
	if _, err := auth.RequireContext(c.UserContext()); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("unauthorized", "authentication is required"))
	}

	summary, err := h.service.GetSummary(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("dashboard_summary_failed", "could not load dashboard summary"))
	}

	return c.JSON(response.OK(summary))
}
