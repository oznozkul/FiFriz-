package navigation

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

func (h *Handler) ListPins(c *fiber.Ctx) error {
	items, err := h.service.ListPins(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("pins_list_failed", "could not list pinned items"))
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) CreatePin(c *fiber.Ctx) error {
	var input CreatePinRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	item, err := h.service.CreatePin(c.UserContext(), input)
	if err != nil {
		if errors.Is(err, ErrInvalidPinData) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_pin_data", "pin data is invalid"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("pin_create_failed", "could not create pin"))
	}
	return c.Status(fiber.StatusCreated).JSON(response.OK(item))
}

func (h *Handler) DeletePin(c *fiber.Ctx) error {
	if err := h.service.DeletePin(c.UserContext(), c.Params("id")); err != nil {
		if errors.Is(err, ErrPinNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("pin_not_found", "pin not found"))
		}
		if errors.Is(err, ErrInvalidPinData) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_pin_id", "pin id is invalid"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("pin_delete_failed", "could not delete pin"))
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListRecent(c *fiber.Ctx) error {
	items, err := h.service.ListRecent(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("recent_list_failed", "could not list recent items"))
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) TrackVisit(c *fiber.Ctx) error {
	var input TrackVisitRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}
	if err := h.service.TrackVisit(c.UserContext(), input); err != nil {
		if errors.Is(err, ErrInvalidPinData) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_visit_data", "visit data is invalid"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("visit_track_failed", "could not track visit"))
	}
	return c.SendStatus(fiber.StatusNoContent)
}
