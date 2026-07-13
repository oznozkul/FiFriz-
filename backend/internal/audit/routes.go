package audit

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/audit-logs")
	group.Get("", handler.List)
	group.Get("/", handler.List)
	group.Get("/:id", handler.GetByID)
}
