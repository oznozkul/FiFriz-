package search

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/search")
	group.Get("/global", handler.Global)
}
