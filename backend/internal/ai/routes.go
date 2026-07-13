package ai

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(group fiber.Router, handler *Handler) {
	group.Get("/ai/workspace-insights", handler.WorkspaceInsights)
	group.Get("/ai/action-suggestions", handler.ListActionSuggestions)
	group.Post("/ai/action-suggestions/:id/accept", handler.AcceptActionSuggestion)
	group.Post("/ai/action-suggestions/:id/reject", handler.RejectActionSuggestion)
	group.Post("/ai/chat", handler.Chat)
	group.Post("/ai/agent-chat", handler.AgentChat)
	group.Post("/ai/chat-sessions", handler.CreateSession)
	group.Get("/ai/chat-sessions", handler.ListSessions)
	group.Get("/ai/chat-sessions/:id", handler.GetSession)
	group.Delete("/ai/chat-sessions/:id", handler.DeleteSession)
	group.Post("/ai/chat-sessions/:id/messages", handler.SendSessionMessage)
}
