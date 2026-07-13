package smartcollections

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/smart-collections")
	group.Get("/", handler.List)
	group.Get("/:key/documents", handler.ListDocuments)
}
