package tags

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/tags")
	group.Get("", handler.List)
	group.Get("/", handler.List)
	group.Get("/search", handler.Search)
	group.Post("", handler.Create)
	group.Post("/", handler.Create)
	group.Get("/:id/stats", handler.GetStats)
	group.Get("/:id/aliases", handler.ListAliases)
	group.Post("/:id/aliases", handler.CreateAlias)
	group.Delete("/:id/aliases/:aliasId", handler.DeleteAlias)
	group.Post("/:id/merge", handler.Merge)
	group.Get("/:id", handler.GetByID)
	group.Put("/:id", handler.Update)
	group.Delete("/:id", handler.Delete)
}
