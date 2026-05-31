package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupUnsubscribeRoutes(router fiber.Router, unsubscribeHandler *handlers.UnsubscribeHandler) {
	// Public unsubscribe endpoint
	router.Get("/unsubscribe", unsubscribeHandler.Unsubscribe)
	
	// Protected preferences endpoints
	protected := router.Group("/preferences", middleware.AuthMiddleware())
	protected.Get("/", unsubscribeHandler.GetPreferences)
	protected.Put("/", unsubscribeHandler.UpdatePreferences)
}