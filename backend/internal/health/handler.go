package health

import (
	"fifriz/internal/response"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Check(c *fiber.Ctx) error {
	sqlDB, err := h.db.DB()
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("database_unavailable", "database connection is unavailable"))
	}

	if err := sqlDB.PingContext(c.UserContext()); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("database_unavailable", "database ping failed"))
	}

	return c.JSON(response.OK(fiber.Map{
		"status": "ok",
	}))
}
