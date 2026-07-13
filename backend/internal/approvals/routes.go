package approvals

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	router.Get("/workflows/pending-for-me", handler.ListPendingForMe)
	router.Get("/workflows/:id", handler.GetByID)
	router.Post("/workflows/:id/approve", handler.Approve)
	router.Post("/workflows/:id/reject", handler.Reject)
	router.Post("/workflows/:id/cancel", handler.Cancel)

	router.Get("/documents/:id/workflows", handler.ListByDocument)
	router.Post("/documents/:id/workflows", handler.CreateForDocument)
}
