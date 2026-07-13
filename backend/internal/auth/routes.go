package auth

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/auth")
	group.Post("/register", handler.Register)
	group.Post("/login", handler.Login)
	group.Post("/forgot-password", handler.ForgotPassword)
	group.Post("/reset-password", handler.ResetPassword)
	group.Get("/me", handler.Middleware(), handler.Me)
}

func RegisterProfileRoutes(router fiber.Router, handler *Handler) {
	group := router.Group("/profile")
	group.Get("/", handler.GetProfile)
	group.Put("/", handler.UpdateProfile)
	group.Patch("/change-password", handler.ChangePassword)
	group.Get("/tenant", handler.GetProfileTenant)
}
