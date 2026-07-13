package health

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	handler := NewHandler(db)
	app.Get("/health", handler.Check)
}
