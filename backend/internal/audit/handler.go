package audit

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	tenantID, _, err := auditAccessContext(c)
	if err != nil {
		return err
	}

	filter, err := parseListFilter(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_audit_log_filter", err.Error()))
	}

	result, err := h.service.List(c.UserContext(), tenantID, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("audit_logs_list_failed", "could not list audit logs"))
	}

	return c.JSON(response.OK(result))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	tenantID, _, err := auditAccessContext(c)
	if err != nil {
		return err
	}

	item, err := h.service.GetByID(c.UserContext(), tenantID, c.Params("id"))
	if err != nil {
		if errors.Is(err, ErrAuditLogNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Fail("audit_log_not_found", "audit log not found"))
		}
		if err.Error() == "invalid audit log id" {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_audit_log_id", "audit log id must be a valid UUID"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("audit_log_get_failed", "could not get audit log"))
	}

	return c.JSON(response.OK(item))
}

func auditAccessContext(c *fiber.Ctx) (uuid.UUID, string, error) {
	role, _ := c.Locals("role").(string)
	if role != "owner" && role != "admin" {
		return uuid.Nil, "", c.Status(fiber.StatusForbidden).JSON(response.Fail("forbidden", "you are not allowed to view audit logs"))
	}

	tenantIDRaw, _ := c.Locals("tenantId").(string)
	tenantID, err := uuid.Parse(tenantIDRaw)
	if err != nil {
		return uuid.Nil, "", c.Status(fiber.StatusUnauthorized).JSON(response.Fail("unauthorized", "authentication is required"))
	}

	return tenantID, role, nil
}

func parseListFilter(c *fiber.Ctx) (ListFilter, error) {
	filter := ListFilter{
		Action:     strings.TrimSpace(c.Query("action")),
		EntityType: strings.TrimSpace(c.Query("entityType")),
		Page:       parsePositiveInt(c.Query("page"), 1),
		PageSize:   parsePositiveInt(c.Query("pageSize"), 20),
	}

	userIDRaw := strings.TrimSpace(c.Query("userId"))
	if userIDRaw != "" {
		userID, err := uuid.Parse(userIDRaw)
		if err != nil {
			return ListFilter{}, errors.New("userId must be a valid UUID")
		}
		filter.UserID = &userID
	}

	if startDateRaw := strings.TrimSpace(c.Query("startDate")); startDateRaw != "" {
		startDate, err := time.Parse(time.RFC3339, startDateRaw)
		if err != nil {
			return ListFilter{}, errors.New("startDate must be RFC3339")
		}
		filter.StartDate = &startDate
	}

	if endDateRaw := strings.TrimSpace(c.Query("endDate")); endDateRaw != "" {
		endDate, err := time.Parse(time.RFC3339, endDateRaw)
		if err != nil {
			return ListFilter{}, errors.New("endDate must be RFC3339")
		}
		filter.EndDate = &endDate
	}

	return filter, nil
}

func parsePositiveInt(value string, fallback int) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}
