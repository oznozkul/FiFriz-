package folders

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/folders")
	group.Get("", handler.List)
	group.Get("/", handler.List)
	group.Get("/search", handler.Search)
	group.Post("", handler.Create)
	group.Get("/:id/documents", handler.ListDocuments)
	group.Get("/:id/metadata", handler.GetMetadata)
	group.Get("/:id/breadcrumb", handler.GetBreadcrumb)
	group.Get("/:id/permissions", handler.GetPermissions)
	group.Put("/:id/permissions", handler.UpdatePermissions)
	group.Post("/:id/move", handler.Move)
	group.Post("/:id/lock", handler.Lock)
	group.Post("/:id/unlock", handler.Unlock)
	group.Post("/:id/restore", handler.Restore)
	group.Get("/:id", handler.GetByID)
	group.Put("/:id", handler.Update)
	group.Delete("/:id", handler.Delete)
}
