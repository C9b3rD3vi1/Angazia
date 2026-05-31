package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupAuthRoutes(router fiber.Router, authHandler *handlers.AuthHandler) {
	// Public routes (no authentication required)
	router.Post("/auth/register", authHandler.Register)
	router.Post("/auth/login", authHandler.Login)
	router.Post("/auth/refresh", authHandler.RefreshToken)
	router.Post("/auth/forgot-password", authHandler.ForgotPassword)
	router.Post("/auth/reset-password", authHandler.ResetPassword)
	router.Get("/auth/verify-email/:token", authHandler.VerifyEmail)
	
	// Protected routes (authentication required)
	protected := router.Group("/auth", middleware.AuthMiddleware())
	protected.Post("/logout", authHandler.Logout)
	protected.Post("/change-password", authHandler.ChangePassword)
	protected.Post("/resend-verification", authHandler.ResendVerificationEmail)
	
	// Profile routes (authentication required)
	profile := router.Group("/profile", middleware.AuthMiddleware())
	profile.Get("/", authHandler.GetProfile)
	profile.Put("/", authHandler.UpdateProfile)
}