package cache

import (
	"github.com/gofiber/fiber/v2"
)

// Middleware is a placeholder for future HTTP response caching.
// Read caching is implemented at the service layer via GetOrSet.
func Middleware(_ *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}
