package sharing

import (
	"errors"
	"strings"

	"fifriz/internal/response"
	"fifriz/internal/stream"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) BulkCreateShareLinks(c *fiber.Ctx) error {
	var input BulkCreateShareLinksRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	result, err := h.service.BulkCreateShareLinks(c.UserContext(), input)
	if err != nil {
		return writeShareError(c, err, fiber.StatusInternalServerError, "bulk_share_link_create_failed", "could not create share links")
	}

	return c.Status(fiber.StatusCreated).JSON(response.OK(result))
}

func (h *Handler) GetShareLinkAnalytics(c *fiber.Ctx) error {
	result, err := h.service.GetShareLinkAnalytics(c.UserContext(), c.Params("shareLinkId"))
	if err != nil {
		return writeShareError(c, err, fiber.StatusInternalServerError, "share_link_analytics_failed", "could not load share analytics")
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) RevokeShareLink(c *fiber.Ctx) error {
	if err := h.service.RevokeShareLink(c.UserContext(), c.Params("id"), c.Params("shareLinkId")); err != nil {
		return writeShareError(c, err, fiber.StatusInternalServerError, "share_link_revoke_failed", "could not revoke share link")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) CreateShareLink(c *fiber.Ctx) error {
	var input CreateShareLinkRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	result, err := h.service.CreateShareLink(c.UserContext(), c.Params("id"), input)
	if err != nil {
		return writeShareError(c, err, fiber.StatusInternalServerError, "share_link_create_failed", "could not create share link")
	}

	return c.Status(fiber.StatusCreated).JSON(response.OK(result))
}

func (h *Handler) ListShareLinks(c *fiber.Ctx) error {
	items, err := h.service.ListShareLinks(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeShareError(c, err, fiber.StatusInternalServerError, "share_links_list_failed", "could not list share links")
	}

	return c.JSON(response.OK(items))
}

func (h *Handler) DeleteShareLink(c *fiber.Ctx) error {
	if err := h.service.DeleteShareLink(c.UserContext(), c.Params("id"), c.Params("shareLinkId")); err != nil {
		return writeShareError(c, err, fiber.StatusInternalServerError, "share_link_delete_failed", "could not delete share link")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) LogShareChannel(c *fiber.Ctx) error {
	var input ChannelLogRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	if err := h.service.LogShareChannel(c.UserContext(), c.Params("shareLinkId"), input.Channel); err != nil {
		return writeShareError(c, err, fiber.StatusInternalServerError, "share_channel_log_failed", "could not log share channel")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListInternalShares(c *fiber.Ctx) error {
	items, err := h.service.ListInternalShares(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeShareError(c, err, fiber.StatusInternalServerError, "internal_shares_list_failed", "could not list internal shares")
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) CreateInternalShare(c *fiber.Ctx) error {
	var input CreateInternalShareRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	result, err := h.service.CreateInternalShare(c.UserContext(), c.Params("id"), input)
	if err != nil {
		return writeShareError(c, err, fiber.StatusInternalServerError, "internal_share_create_failed", "could not create internal share")
	}
	return c.Status(fiber.StatusCreated).JSON(response.OK(result))
}

func (h *Handler) UpdateInternalShare(c *fiber.Ctx) error {
	var input UpdateInternalShareRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	result, err := h.service.UpdateInternalShare(c.UserContext(), c.Params("id"), c.Params("shareId"), input)
	if err != nil {
		return writeShareError(c, err, fiber.StatusInternalServerError, "internal_share_update_failed", "could not update internal share")
	}
	return c.JSON(response.OK(result))
}

func (h *Handler) DeleteInternalShare(c *fiber.Ctx) error {
	if err := h.service.DeleteInternalShare(c.UserContext(), c.Params("id"), c.Params("shareId")); err != nil {
		return writeShareError(c, err, fiber.StatusInternalServerError, "internal_share_delete_failed", "could not delete internal share")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListSharedWithMe(c *fiber.Ctx) error {
	items, err := h.service.ListSharedWithMe(c.UserContext())
	if err != nil {
		return writeShareError(c, err, fiber.StatusInternalServerError, "shared_with_me_list_failed", "could not list shared documents")
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) GetPublicShare(c *fiber.Ctx) error {
	accessToken := shareAccessTokenFromRequest(c)
	result, err := h.service.GetPublicShare(c.UserContext(), c.Params("token"), accessToken, publicAccessMetaFromRequest(c))
	if err != nil {
		return writePublicShareError(c, err)
	}

	return c.JSON(response.OK(result))
}

func (h *Handler) VerifyPassword(c *fiber.Ctx) error {
	var input VerifyPasswordRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	result, err := h.service.VerifyPassword(c.UserContext(), c.Params("token"), input.Password, publicAccessMetaFromRequest(c))
	if err != nil {
		return writePublicShareError(c, err)
	}

	return c.JSON(response.OK(result))
}

func (h *Handler) DownloadPublicShare(c *fiber.Ctx) error {
	accessToken := shareAccessTokenFromRequest(c)
	accessMode := "preview"
	if c.Query("download") == "1" {
		accessMode = "download"
	}
	file, err := h.service.GetPublicDownload(c.UserContext(), c.Params("token"), accessToken, accessMode, publicAccessMetaFromRequest(c))
	if err != nil {
		return writePublicShareError(c, err)
	}

	if c.Query("download") == "1" && !file.AllowDownload {
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("download_not_allowed", "download is not allowed for this share link"))
	}

	disposition := "inline"
	if c.Query("download") == "1" {
		disposition = "attachment"
	}

	if err := stream.SendFile(c, file.Reader, file.Size, file.MimeType, file.OriginalFileName, disposition); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("share_download_failed", "could not download shared document"))
	}
	return nil
}

func shareAccessTokenFromRequest(c *fiber.Ctx) string {
	if value := strings.TrimSpace(c.Get("X-Share-Access-Token")); value != "" {
		return value
	}
	return strings.TrimSpace(c.Query("accessToken"))
}

func publicAccessMetaFromRequest(c *fiber.Ctx) PublicAccessMeta {
	visitorEmail := strings.TrimSpace(c.Get("X-Share-Visitor-Email"))
	if visitorEmail == "" {
		visitorEmail = strings.TrimSpace(c.Query("visitorEmail"))
	}
	return PublicAccessMeta{
		IPAddress:    c.IP(),
		UserAgent:    c.Get("User-Agent"),
		VisitorEmail: visitorEmail,
	}
}

func writeShareError(c *fiber.Ctx, err error, fallbackStatus int, fallbackCode, fallbackMessage string) error {
	switch {
	case errors.Is(err, ErrInvalidDocumentID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_document_id", "document id must be a valid UUID"))
	case errors.Is(err, ErrInvalidShareLinkID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_share_link_id", "share link id must be a valid UUID"))
	case errors.Is(err, ErrDocumentNotFound), errors.Is(err, ErrShareLinkNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("not_found", "resource not found"))
	case errors.Is(err, ErrForbidden):
		return response.WriteForbidden(c, "forbidden", "you do not have permission for this action")
	case errors.Is(err, ErrInvalidSharePermission):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_share_permission", "permission must be view or download"))
	case errors.Is(err, ErrInvalidShareType):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_share_type", "share type must be public_link or email_invite"))
	case errors.Is(err, ErrInvalidShareChannel):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_share_channel", "channel must be whatsapp, telegram, email, copy, or native"))
	case errors.Is(err, ErrInvalidInternalShareID):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_internal_share_id", "internal share id must be a valid UUID"))
	case errors.Is(err, ErrInternalShareNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("internal_share_not_found", "internal share not found"))
	case errors.Is(err, ErrDuplicateInternalShare):
		return c.Status(fiber.StatusConflict).JSON(response.Fail("duplicate_internal_share", "this document is already shared with the user"))
	case errors.Is(err, ErrCannotShareWithSelf):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("cannot_share_with_self", "you cannot share a document with yourself"))
	case errors.Is(err, ErrShareUserNotInTenant):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("share_user_not_in_tenant", "the selected user is not a member of this tenant"))
	case errors.Is(err, ErrInvalidInternalPermission):
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_internal_permission", "permission must be view, download, edit, or manage"))
	default:
		return c.Status(fallbackStatus).JSON(response.Fail(fallbackCode, fallbackMessage))
	}
}

func writePublicShareError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrShareLinkNotFound), errors.Is(err, ErrInvalidShareToken):
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("share_link_not_found", "share link not found"))
	case errors.Is(err, ErrShareLinkExpired):
		return c.Status(fiber.StatusGone).JSON(response.Fail("share_link_expired", "this share link has expired"))
	case errors.Is(err, ErrShareLinkInactive):
		return c.Status(fiber.StatusGone).JSON(response.Fail("share_link_inactive", "this share link is no longer active"))
	case errors.Is(err, ErrMaxViewsReached):
		return c.Status(fiber.StatusGone).JSON(response.Fail("share_link_max_views", "this share link has reached its view limit"))
	case errors.Is(err, ErrPasswordRequired):
		return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("password_required", "password is required to access this share link"))
	case errors.Is(err, ErrInvalidPassword):
		return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("invalid_password", "password is incorrect"))
	case errors.Is(err, ErrDownloadNotAllowed):
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("download_not_allowed", "download is not allowed for this share link"))
	case errors.Is(err, ErrShareAccessDenied):
		return c.Status(fiber.StatusForbidden).JSON(response.Fail("share_access_denied", "access to this share link is denied"))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("share_access_failed", "could not access shared document"))
	}
}
