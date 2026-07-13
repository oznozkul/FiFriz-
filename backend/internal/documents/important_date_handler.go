package documents

import (
	"errors"

	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ListImportantDates(c *fiber.Ctx) error {
	items, err := h.service.ListImportantDates(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeImportantDateError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) CreateImportantDate(c *fiber.Ctx) error {
	var request CreateDocumentImportantDateRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "request body is invalid"))
	}

	result, err := h.service.CreateImportantDate(c.UserContext(), c.Params("id"), request)
	if err != nil {
		return writeImportantDateError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(response.OK(result))
}

func (h *Handler) UpdateImportantDate(c *fiber.Ctx) error {
	var request UpdateDocumentImportantDateRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "request body is invalid"))
	}

	result, err := h.service.UpdateImportantDate(c.UserContext(), c.Params("id"), c.Params("dateId"), request)
	if err != nil {
		return writeImportantDateError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) DeleteImportantDate(c *fiber.Ctx) error {
	if err := h.service.DeleteImportantDate(c.UserContext(), c.Params("id"), c.Params("dateId")); err != nil {
		return writeImportantDateError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) CompleteImportantDate(c *fiber.Ctx) error {
	result, err := h.service.CompleteImportantDate(c.UserContext(), c.Params("id"), c.Params("dateId"))
	if err != nil {
		return writeImportantDateError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) ListImportantDatesGlobal(c *fiber.Ctx) error {
	query := parseImportantDateListQuery(c)
	items, err := h.service.ListImportantDatesGlobal(c.UserContext(), query)
	if err != nil {
		return writeImportantDateError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) ListImportantDateSuggestions(c *fiber.Ctx) error {
	items, err := h.service.ListImportantDateSuggestions(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeImportantDateError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) ExtractAIImportantDates(c *fiber.Ctx) error {
	items, err := h.service.ExtractAIImportantDates(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeAIImportantDateError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) AcceptImportantDateSuggestion(c *fiber.Ctx) error {
	result, err := h.service.AcceptImportantDateSuggestion(c.UserContext(), c.Params("id"), c.Params("suggestionId"))
	if err != nil {
		return writeAIImportantDateError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) RejectImportantDateSuggestion(c *fiber.Ctx) error {
	result, err := h.service.RejectImportantDateSuggestion(c.UserContext(), c.Params("id"), c.Params("suggestionId"))
	if err != nil {
		return writeAIImportantDateError(c, err)
	}
	return c.JSON(response.OK(result))
}

func writeAIImportantDateError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrAIImportantDatesProviderDisabled):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("important_dates_provider_disabled", "important dates service is not configured"))
	case errors.Is(err, ErrSummaryProviderUnavailable):
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("important_dates_provider_unavailable", "AI önemli tarih servisi şu anda kullanılamıyor"))
	case errors.Is(err, ErrAIImportantDatesNoContent):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("important_dates_no_content", "document has no extractable content for important dates"))
	case errors.Is(err, ErrAIImportantDatesParseFailed):
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Fail("important_dates_parse_failed", "AI önemli tarih yanıtı işlenemedi"))
	case errors.Is(err, ErrImportantDateSuggestionNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("important_date_suggestion_not_found", "important date suggestion not found"))
	case errors.Is(err, ErrImportantDateSuggestionInvalid):
		return c.Status(fiber.StatusConflict).JSON(response.Fail("important_date_suggestion_invalid", "important date suggestion is not pending"))
	default:
		return writeImportantDateError(c, err)
	}
}

func writeImportantDateError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidDocumentID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
	case errors.Is(err, ErrInvalidImportantDateID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_important_date_id", "important date id must be a valid UUID"))
	case errors.Is(err, ErrEmptyImportantDateTitle):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("empty_important_date_title", "title must not be empty"))
	case errors.Is(err, ErrInvalidImportantDateValue):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_important_date_value", "date must be in YYYY-MM-DD format"))
	case errors.Is(err, ErrInvalidImportantDateType):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_important_date_type", "type must be one of expiration, renewal, payment_due, review, custom"))
	case errors.Is(err, ErrInvalidAssignedUser):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_assigned_user", "assigned user must belong to the tenant"))
	case errors.Is(err, ErrDocumentNotFound), errors.Is(err, ErrDocumentImportantDateNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("not_found", "resource not found"))
	case errors.Is(err, ErrForbidden):
		return response.WriteForbidden(c, "forbidden", "you do not have permission for this action")
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("important_date_failed", "could not process important date"))
	}
}
