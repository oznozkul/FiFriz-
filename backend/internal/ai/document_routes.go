package ai

import "github.com/gofiber/fiber/v2"

func RegisterDocumentRoutes(group fiber.Router, handler *Handler) {
	documents := group.Group("/documents")
	documents.Get("/:id/ai/action-suggestions", handler.ListDocumentActionSuggestions)
	documents.Post("/:id/ai/action-suggestions/generate", handler.GenerateDocumentActionSuggestions)
}
