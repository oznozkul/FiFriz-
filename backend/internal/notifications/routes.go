package notifications

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/notifications")
	group.Get("", handler.List)
	group.Get("/unread-count", handler.UnreadCount)
	group.Patch("/read-all", handler.MarkAllRead)
	group.Patch("/:id/read", handler.MarkRead)
}
