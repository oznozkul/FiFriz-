package approvals

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

func (h *Handler) CreateForDocument(c *fiber.Ctx) error {
	var request CreateWorkflowRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "request body is invalid"))
	}
	result, err := h.service.CreateForDocument(c.UserContext(), c.Params("id"), request)
	if err != nil {
		return writeWorkflowError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(response.OK(result))
}

func (h *Handler) ListByDocument(c *fiber.Ctx) error {
	items, err := h.service.ListByDocument(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeWorkflowError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	result, err := h.service.GetByID(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeWorkflowError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) ListPendingForMe(c *fiber.Ctx) error {
	items, err := h.service.ListPendingForMe(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("pending_approvals_failed", "could not list pending approvals"))
	}
	if items == nil {
		items = []PendingApprovalResponse{}
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) Approve(c *fiber.Ctx) error {
	var request DecideWorkflowRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "request body is invalid"))
	}
	result, err := h.service.Approve(c.UserContext(), c.Params("id"), request)
	if err != nil {
		return writeWorkflowError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) Reject(c *fiber.Ctx) error {
	var request DecideWorkflowRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "request body is invalid"))
	}
	result, err := h.service.Reject(c.UserContext(), c.Params("id"), request)
	if err != nil {
		return writeWorkflowError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) Cancel(c *fiber.Ctx) error {
	result, err := h.service.Cancel(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeWorkflowError(c, err)
	}
	return c.JSON(response.OK(result))
}

func writeWorkflowError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidDocumentID), errors.Is(err, ErrInvalidWorkflowID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_id", "id must be a valid UUID"))
	case errors.Is(err, ErrInvalidWorkflowInput), errors.Is(err, ErrInvalidStepOrder):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_workflow_input", "workflow input is invalid"))
	case errors.Is(err, ErrApproverNotInTenant):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("approver_not_in_tenant", "approver must be a tenant member"))
	case errors.Is(err, ErrActiveWorkflowExists):
		return c.Status(fiber.StatusConflict).JSON(response.Fail("active_workflow_exists", "document already has a pending workflow"))
	case errors.Is(err, ErrWorkflowNotPending):
		return c.Status(fiber.StatusConflict).JSON(response.Fail("workflow_not_pending", "workflow is not pending"))
	case errors.Is(err, ErrNotCurrentApprover):
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("not_current_approver", "only the current step approver can perform this action"))
	case errors.Is(err, ErrDocumentNotFound), errors.Is(err, ErrWorkflowNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("not_found", "resource not found"))
	case errors.Is(err, ErrForbidden):
		return response.WriteForbidden(c, "forbidden", "you do not have permission for this action")
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("workflow_failed", "could not process workflow"))
	}
}
