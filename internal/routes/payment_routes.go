package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupSubscriptionRoutes(router fiber.Router, subscriptionHandler *handlers.SubscriptionHandler) {
	// Public plan listing (no authentication)
	router.Get("/subscriptions/plans", subscriptionHandler.GetPlans)
	
	// Webhook endpoint (no auth — IntaSend calls this)
	router.Post("/payments/webhook", subscriptionHandler.Webhook)
	
	// Payment verification (no auth — uses transaction_id/reference query params)
	router.Get("/payments/verify", subscriptionHandler.VerifyPayment)
	
	// Protected subscription routes (authentication required)
	protected := router.Group("/subscriptions", middleware.AuthMiddleware())
	
	protected.Get("/current", subscriptionHandler.GetCurrentSubscription)
	protected.Post("/subscribe", subscriptionHandler.Subscribe)
	protected.Post("/cancel", subscriptionHandler.CancelSubscription)
	protected.Post("/reactivate", subscriptionHandler.ReactivateSubscription)
	protected.Post("/upgrade", subscriptionHandler.UpgradeSubscription)
	protected.Post("/downgrade", subscriptionHandler.DowngradeSubscription)
	protected.Post("/proration", subscriptionHandler.GetProration)
	protected.Post("/retry-payment", subscriptionHandler.RetryPayment)
	protected.Get("/invoices", subscriptionHandler.GetInvoices)
	
	// Payment methods
	methods := protected.Group("/payment-methods")
	methods.Get("/", subscriptionHandler.GetPaymentMethods)
	methods.Post("/", subscriptionHandler.AddPaymentMethod)
	methods.Delete("/:id", subscriptionHandler.RemovePaymentMethod)
	methods.Put("/:id/default", subscriptionHandler.SetDefaultPaymentMethod)
}