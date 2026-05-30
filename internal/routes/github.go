package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/Angazia/internal/handlers"
)

func setupGitHubRoutes(router fiber.Router, githubHandler *handlers.GitHubHandler) {
	// OAuth flow (public)
	router.Get("/github/auth", githubHandler.Login)
	router.Get("/github/callback", githubHandler.Callback)
	
	// Webhook (public)
	router.Post("/github/webhook", githubHandler.Webhook)
}