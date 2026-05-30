package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/Angazia/internal/handlers"
)

func setupAuthRoutes(router fiber.Router, authHandler *handlers.AuthHandler) {
	router.Post("/auth/register", authHandler.Register)
	router.Post("/auth/login", authHandler.Login)
	router.Post("/auth/logout", authHandler.Logout)
	router.Post("/auth/refresh", authHandler.RefreshToken)
	router.Post("/auth/forgot-password", authHandler.ForgotPassword)
	router.Post("/auth/reset-password", authHandler.ResetPassword)
	router.Get("/auth/verify-email/:token", authHandler.VerifyEmail)
}