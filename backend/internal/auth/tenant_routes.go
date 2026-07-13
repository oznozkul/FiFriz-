package auth

import "github.com/gofiber/fiber/v2"

func RegisterTenantRoutes(router fiber.Router, handler *Handler) {
	router.Get("/tenant/members", handler.ListTenantMembers)
}
