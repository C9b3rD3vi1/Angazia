package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupGitHubRoutes(router fiber.Router, githubHandler *handlers.GitHubHandler) {
	// OAuth flow (public)
	router.Get("/github/auth", githubHandler.Login)
	router.Get("/github/callback", githubHandler.Callback)
	
	// Webhook (public)
	router.Post("/github/webhook", githubHandler.Webhook)

	// Protected routes (authentication required)
	protected := router.Group("/github", middleware.AuthMiddleware())
	protected.Post("/connect", githubHandler.Connect)
	protected.Post("/disconnect", githubHandler.Disconnect)
	protected.Post("/sync", githubHandler.Sync)
	protected.Get("/profile", githubHandler.GetProfile)
	protected.Get("/repos", githubHandler.GetRepos)
	protected.Get("/contributions", githubHandler.GetContributions)
}