package auth

import (
	"errors"
	"strings"

	"fifriz/internal/audit"
	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var input RegisterRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	authResponse, err := h.service.Register(audit.WithRequestMeta(c.UserContext(), audit.RequestMeta{
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
	}), input)
	if err != nil {
		if errors.Is(err, ErrInvalidRegistration) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_registration", "accountType, fullName, valid email and password are required; organizationName is required for organization accounts"))
		}
		if errors.Is(err, ErrEmailAlreadyRegistered) {
			return c.Status(fiber.StatusConflict).JSON(response.Fail("email_already_registered", "email is already registered"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("register_failed", "could not register user"))
	}

	return c.Status(fiber.StatusCreated).JSON(response.OK(authResponse))
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var input LoginRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	authResponse, err := h.service.Login(audit.WithRequestMeta(c.UserContext(), audit.RequestMeta{
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
	}), input)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("invalid_credentials", "email or password is incorrect"))
		}
		if errors.Is(err, ErrNoTenantMembership) {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("tenant_membership_missing", "user does not belong to any tenant"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("login_failed", "could not login"))
	}

	return c.JSON(response.OK(authResponse))
}

func (h *Handler) Me(c *fiber.Ctx) error {
	authContext, err := RequireContext(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("unauthorized", "authentication is required"))
	}

	meResponse, err := h.service.Me(c.UserContext(), authContext)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("me_failed", "could not load current user"))
	}

	return c.JSON(response.OK(meResponse))
}

func (h *Handler) ListTenantMembers(c *fiber.Ctx) error {
	members, err := h.service.ListTenantMembers(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("tenant_members_list_failed", "could not list tenant members"))
	}
	return c.JSON(response.OK(members))
}

func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var input ForgotPasswordRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	result, err := h.service.ForgotPassword(audit.WithRequestMeta(c.UserContext(), audit.RequestMeta{
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
	}), input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("forgot_password_failed", "could not process password reset request"))
	}

	return c.JSON(response.OK(result))
}

func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	var input ResetPasswordRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	result, err := h.service.ResetPassword(audit.WithRequestMeta(c.UserContext(), audit.RequestMeta{
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
	}), input)
	if err != nil {
		if errors.Is(err, ErrInvalidResetToken) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_reset_token", "şifre sıfırlama bağlantısı geçersiz veya süresi dolmuş"))
		}
		if errors.Is(err, ErrWeakPassword) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("weak_password", "şifre en az 8 karakter olmalıdır"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("reset_password_failed", "şifre sıfırlanamadı"))
	}

	return c.JSON(response.OK(result))
}

func (h *Handler) GetProfile(c *fiber.Ctx) error {
	authContext, err := RequireContext(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("unauthorized", "authentication is required"))
	}

	profile, err := h.service.GetProfile(c.UserContext(), authContext)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("profile_load_failed", "profil bilgisi yüklenemedi"))
	}

	return c.JSON(response.OK(profile))
}

func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	authContext, err := RequireContext(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("unauthorized", "authentication is required"))
	}

	var input UpdateProfileRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	profile, err := h.service.UpdateProfile(audit.WithRequestMeta(c.UserContext(), audit.RequestMeta{
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
	}), authContext, input)
	if err != nil {
		if errors.Is(err, ErrInvalidProfile) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_profile", "ad soyad gereklidir"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("profile_update_failed", "profil güncellenemedi"))
	}

	return c.JSON(response.OK(profile))
}

func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	authContext, err := RequireContext(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("unauthorized", "authentication is required"))
	}

	var input ChangePasswordRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_request_body", "request body is invalid"))
	}

	result, err := h.service.ChangePassword(audit.WithRequestMeta(c.UserContext(), audit.RequestMeta{
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
	}), authContext, input)
	if err != nil {
		if errors.Is(err, ErrInvalidCurrentPassword) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("invalid_current_password", "mevcut şifre hatalı"))
		}
		if errors.Is(err, ErrWeakPassword) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Fail("weak_password", "yeni şifre en az 8 karakter olmalıdır"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("password_change_failed", "şifre değiştirilemedi"))
	}

	return c.JSON(response.OK(result))
}

func (h *Handler) GetProfileTenant(c *fiber.Ctx) error {
	authContext, err := RequireContext(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("unauthorized", "authentication is required"))
	}

	tenantInfo, err := h.service.GetProfileTenant(c.UserContext(), authContext)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("tenant_info_failed", "tenant bilgisi yüklenemedi"))
	}

	return c.JSON(response.OK(tenantInfo))
}

func (h *Handler) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("unauthorized", "authorization bearer token is required"))
		}

		authContext, err := h.service.ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Fail("invalid_token", "authorization token is invalid"))
		}
		authContext, err = h.service.VerifyMembership(c.UserContext(), authContext)
		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(response.Fail("tenant_access_denied", "active tenant access is denied"))
		}

		c.SetUserContext(WithContext(c.UserContext(), authContext))
		c.Locals("userId", authContext.UserID.String())
		c.Locals("tenantId", authContext.TenantID.String())
		c.Locals("role", string(authContext.Role))

		return c.Next()
	}
}
