package integrations

import "github.com/gofiber/fiber/v2"

func RegisterPublicRoutes(router fiber.Router, handler *Handler) {
	router.Get("/integrations/gmail/callback", handler.GmailCallback)
	router.Get("/integrations/outlook/callback", handler.OutlookCallback)
	router.Get("/integrations/google-drive/callback", handler.GoogleDriveCallback)
	router.Get("/integrations/onedrive/callback", handler.OneDriveCallback)
}

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/integrations")
	group.Get("/providers", handler.ListProviders)
	group.Get("/providers/:provider/connect", handler.ProviderConnect)
	group.Get("/gmail/connect", handler.GmailConnect)
	group.Get("/outlook/connect", handler.OutlookConnect)
	group.Get("/google-drive/connect", handler.GoogleDriveConnect)
	group.Get("/onedrive/connect", handler.OneDriveConnect)
	group.Get("/", handler.List)
	group.Post("/", handler.Create)
	group.Get("/:id/sync-logs", handler.ListSyncLogs)
	group.Post("/:id/sync", handler.QueueSync)
	group.Get("/:id/health", handler.HealthCheck)
	group.Get("/:id", handler.GetByID)
	group.Put("/:id", handler.Update)
	group.Delete("/:id", handler.Delete)
}
