package notifications

import (
	"errors"
	"strconv"
	"strings"

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
	filter, err := parseListFilter(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_notification_filter", err.Error()))
	}

	result, err := h.service.List(c.UserContext(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("notifications_list_failed", "could not list notifications"))
	}

	return c.JSON(response.OK(result))
}

func (h *Handler) UnreadCount(c *fiber.Ctx) error {
	result, err := h.service.UnreadCount(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("notifications_unread_count_failed", "could not get unread count"))
	}

	return c.JSON(response.OK(result))
}

func (h *Handler) MarkRead(c *fiber.Ctx) error {
	result, err := h.service.MarkRead(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeNotificationError(c, err)
	}

	return c.JSON(response.OK(result))
}

func (h *Handler) MarkAllRead(c *fiber.Ctx) error {
	if err := h.service.MarkAllRead(c.UserContext()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("notifications_mark_all_read_failed", "could not mark notifications as read"))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func parseListFilter(c *fiber.Ctx) (ListFilter, error) {
	filter := ListFilter{
		Page:     1,
		PageSize: 20,
	}

	if unreadOnlyRaw := strings.TrimSpace(c.Query("unreadOnly")); unreadOnlyRaw != "" {
		value, err := strconv.ParseBool(unreadOnlyRaw)
		if err != nil {
			return filter, errors.New("unreadOnly must be a boolean")
		}
		filter.UnreadOnly = value
	}

	if pageRaw := strings.TrimSpace(c.Query("page")); pageRaw != "" {
		page, err := strconv.Atoi(pageRaw)
		if err != nil || page < 1 {
			return filter, errors.New("page must be a positive integer")
		}
		filter.Page = page
	}

	if pageSizeRaw := strings.TrimSpace(c.Query("pageSize")); pageSizeRaw != "" {
		pageSize, err := strconv.Atoi(pageSizeRaw)
		if err != nil || pageSize < 1 {
			return filter, errors.New("pageSize must be a positive integer")
		}
		filter.PageSize = pageSize
	}

	return filter, nil
}

func writeNotificationError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidNotificationID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_notification_id", "notification id must be a valid UUID"))
	case errors.Is(err, ErrNotificationNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("notification_not_found", "notification not found"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("notification_failed", "could not process notification"))
	}
}
