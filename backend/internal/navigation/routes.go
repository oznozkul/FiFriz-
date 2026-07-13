package navigation

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	router.Get("/pins", handler.ListPins)
	router.Post("/pins", handler.CreatePin)
	router.Delete("/pins/:id", handler.DeletePin)
	router.Get("/recent", handler.ListRecent)
	router.Post("/navigation/visited", handler.TrackVisit)
}
