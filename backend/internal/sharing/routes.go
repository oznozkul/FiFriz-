package sharing

import "github.com/gofiber/fiber/v2"

func RegisterProtectedRoutes(router fiber.Router, handler *Handler) {
	router.Get("/documents/shared-with-me", handler.ListSharedWithMe)
	router.Post("/documents/bulk-share-links", handler.BulkCreateShareLinks)
	router.Post("/documents/:id/share-links", handler.CreateShareLink)
	router.Get("/documents/:id/share-links", handler.ListShareLinks)
	router.Delete("/documents/:id/share-links/:shareLinkId", handler.DeleteShareLink)
	router.Post("/documents/:id/share-links/:shareLinkId/revoke", handler.RevokeShareLink)
	router.Get("/documents/share-links/:shareLinkId/analytics", handler.GetShareLinkAnalytics)
	router.Post("/documents/share-links/:shareLinkId/channel-log", handler.LogShareChannel)
	router.Get("/documents/:id/internal-shares", handler.ListInternalShares)
	router.Post("/documents/:id/internal-shares", handler.CreateInternalShare)
	router.Put("/documents/:id/internal-shares/:shareId", handler.UpdateInternalShare)
	router.Delete("/documents/:id/internal-shares/:shareId", handler.DeleteInternalShare)
}

func RegisterPublicRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/public/share")
	group.Get("/:token", handler.GetPublicShare)
	group.Post("/:token/verify-password", handler.VerifyPassword)
	group.Get("/:token/download", handler.DownloadPublicShare)
}
