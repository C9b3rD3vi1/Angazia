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
	router.Post("/auth/admin/login", authHandler.AdminLogin)
	router.Post("/auth/refresh", authHandler.RefreshToken)
	router.Post("/auth/forgot-password", authHandler.ForgotPassword)
	router.Post("/auth/reset-password", authHandler.ResetPassword)
	router.Get("/auth/verify-email/:token", authHandler.VerifyEmail)
	router.Post("/auth/resend-verification", authHandler.ResendVerificationEmail)

	// Protected routes (authentication required)
	protected := router.Group("/auth", middleware.AuthMiddleware())
	protected.Post("/logout", authHandler.Logout)
	protected.Post("/change-password", authHandler.ChangePassword)
	protected.Get("/sessions", authHandler.GetSessions)
	protected.Post("/sessions/revoke", authHandler.RevokeSession)
	protected.Post("/delete-account", authHandler.DeleteAccount)
	
	// Profile routes (authentication required)
	profile := router.Group("/profile", middleware.AuthMiddleware())
	profile.Get("/", authHandler.GetProfile)
	profile.Put("/", authHandler.UpdateProfile)

	// Employer candidate routes (authentication required)
	employerCandidate := router.Group("/employer/candidates", middleware.AuthMiddleware(), middleware.RequireRole("employer"))
	employerCandidate.Get("/:id", authHandler.GetCandidateProfile)
}