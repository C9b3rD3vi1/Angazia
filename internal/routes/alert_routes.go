package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupAlertRoutes(router fiber.Router, alertHandler *handlers.AlertHandler) {
	// Protected routes (authentication required)
	protected := router.Group("/", middleware.AuthMiddleware())
	
	// Saved search endpoints
	protected.Post("/alerts/search", alertHandler.CreateSavedSearch)
	protected.Get("/alerts", alertHandler.ListSavedSearches)
	protected.Get("/alerts/:id", alertHandler.GetSavedSearch)
	protected.Put("/alerts/:id", alertHandler.UpdateSavedSearch)
	protected.Delete("/alerts/:id", alertHandler.DeleteSavedSearch)
	protected.Post("/alerts/:id/test", alertHandler.TestAlert)
	
	// Alert settings endpoints
	protected.Get("/alerts/settings", alertHandler.GetAlertSettings)
	protected.Put("/alerts/settings", alertHandler.UpdateAlertSettings)
	
	// Alert history endpoint
	protected.Get("/alerts/history", alertHandler.GetAlertHistory)
}