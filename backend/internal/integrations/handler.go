package integrations

import (
	"errors"
	"strings"

	"fifriz/internal/applog"
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
	items, err := h.service.List(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("integrations_list_failed", "could not list integrations"))
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var request CreateIntegrationRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "request body is invalid"))
	}
	result, err := h.service.Create(c.UserContext(), request)
	if err != nil {
		return writeIntegrationError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(response.OK(result))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	result, err := h.service.GetByID(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeIntegrationError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var request UpdateIntegrationRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request", "request body is invalid"))
	}
	result, err := h.service.Update(c.UserContext(), c.Params("id"), request)
	if err != nil {
		return writeIntegrationError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.service.Delete(c.UserContext(), c.Params("id")); err != nil {
		return writeIntegrationError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) QueueSync(c *fiber.Ctx) error {
	result, err := h.service.QueueSync(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeIntegrationError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) ListSyncLogs(c *fiber.Ctx) error {
	items, err := h.service.ListSyncLogs(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeIntegrationError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) ListProviders(c *fiber.Ctx) error {
	return c.JSON(response.OK(h.service.ListProviders()))
}

func (h *Handler) ProviderConnect(c *fiber.Ctx) error {
	authURL, err := h.service.ProviderConnectURL(c.UserContext(), c.Params("provider"))
	if err != nil {
		return writeIntegrationError(c, err)
	}
	return c.JSON(response.OK(GmailConnectResponse{AuthURL: authURL}))
}

func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	result, err := h.service.CheckHealth(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeIntegrationError(c, err)
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) GmailConnect(c *fiber.Ctx) error {
	authURL, err := h.service.GmailConnectURL(c.UserContext())
	if err != nil {
		return writeIntegrationError(c, err)
	}
	return c.JSON(response.OK(GmailConnectResponse{AuthURL: authURL}))
}

func (h *Handler) GmailCallback(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	if code == "" || state == "" {
		return c.Redirect(h.service.FrontendIntegrationsURL() + "?gmail=error")
	}
	if _, err := h.service.CompleteGmailOAuth(c.UserContext(), code, state); err != nil {
		applog.Warn("gmail oauth callback failed: error=%v", err)
		return c.Redirect(h.service.FrontendIntegrationsURL() + "?gmail=error")
	}
	return c.Redirect(h.service.FrontendIntegrationsURL() + "?gmail=connected")
}

func (h *Handler) OutlookConnect(c *fiber.Ctx) error {
	authURL, err := h.service.OutlookConnectURL(c.UserContext())
	if err != nil {
		return writeIntegrationError(c, err)
	}
	return c.JSON(response.OK(OutlookConnectResponse{AuthURL: authURL}))
}

func (h *Handler) OutlookCallback(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	if code == "" || state == "" {
		return c.Redirect(h.service.FrontendIntegrationsURL() + "?outlook=error")
	}
	if _, err := h.service.CompleteOutlookOAuth(c.UserContext(), code, state); err != nil {
		applog.Warn("outlook oauth callback failed: error=%v", err)
		return c.Redirect(h.service.FrontendIntegrationsURL() + "?outlook=error")
	}
	return c.Redirect(h.service.FrontendIntegrationsURL() + "?outlook=connected")
}

func (h *Handler) GoogleDriveConnect(c *fiber.Ctx) error {
	authURL, err := h.service.GoogleDriveConnectURL(c.UserContext())
	if err != nil {
		return writeIntegrationError(c, err)
	}
	return c.JSON(response.OK(GoogleDriveConnectResponse{AuthURL: authURL}))
}

func (h *Handler) GoogleDriveCallback(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	if code == "" || state == "" {
		return c.Redirect(h.service.FrontendIntegrationsURL() + "?google_drive=error")
	}
	if _, err := h.service.CompleteGoogleDriveOAuth(c.UserContext(), code, state); err != nil {
		applog.Warn("google drive oauth callback failed: error=%v", err)
		return c.Redirect(h.service.FrontendIntegrationsURL() + "?google_drive=error")
	}
	return c.Redirect(h.service.FrontendIntegrationsURL() + "?google_drive=connected")
}

func (h *Handler) OneDriveConnect(c *fiber.Ctx) error {
	authURL, err := h.service.OneDriveConnectURL(c.UserContext())
	if err != nil {
		return writeIntegrationError(c, err)
	}
	return c.JSON(response.OK(OutlookConnectResponse{AuthURL: authURL}))
}

func (h *Handler) OneDriveCallback(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	if code == "" || state == "" {
		return c.Redirect(h.service.FrontendIntegrationsURL() + "?onedrive=error")
	}
	if _, err := h.service.CompleteOneDriveOAuth(c.UserContext(), code, state); err != nil {
		applog.Warn("onedrive oauth callback failed: error=%v", err)
		return c.Redirect(h.service.FrontendIntegrationsURL() + "?onedrive=error")
	}
	return c.Redirect(h.service.FrontendIntegrationsURL() + "?onedrive=connected")
}

func writeIntegrationError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrIntegrationNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("integration_not_found", "integration not found"))
	case errors.Is(err, ErrIMAPPasswordRequired):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("imap_password_required", "IMAP password is required"))
	case errors.Is(err, ErrIMAPEmailRequired):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("imap_email_required", "IMAP email address is required"))
	case errors.Is(err, ErrIntegrationConnection):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("integration_connection_failed", err.Error()))
	case errors.Is(err, ErrInvalidIntegrationID), errors.Is(err, ErrInvalidIntegrationInput), errors.Is(err, ErrInvalidProvider):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_integration_request", err.Error()))
	default:
		message := strings.TrimSpace(err.Error())
		if message == "" {
			message = "integration request failed"
		}
		lower := strings.ToLower(message)
		if strings.Contains(lower, "imap") || strings.Contains(lower, "connection") || strings.Contains(lower, "login") {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("integration_connection_failed", message))
		}
		if strings.Contains(lower, "relation") && strings.Contains(lower, "integration") {
			return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("integration_database_error", "Integration tables are missing. Restart the backend API to run migrations."))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("integration_request_failed", message))
	}
}
