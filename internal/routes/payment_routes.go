package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupSubscriptionRoutes(router fiber.Router, subscriptionHandler *handlers.SubscriptionHandler) {
	// Public plan listing (no authentication)
	router.Get("/subscriptions/plans", subscriptionHandler.GetPlans)
	
	// Protected subscription routes (authentication required)
	protected := router.Group("/subscriptions", middleware.AuthMiddleware())
	
	protected.Get("/current", subscriptionHandler.GetCurrentSubscription)
	protected.Post("/cancel", subscriptionHandler.CancelSubscription)
	protected.Get("/invoices", subscriptionHandler.GetInvoices)
}