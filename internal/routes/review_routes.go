package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupReviewRoutes(router fiber.Router, reviewHandler *handlers.ReviewHandler) {
	// Public review routes (read-only, no authentication)
	router.Get("/companies/:id/reviews", reviewHandler.GetCompanyReviews)
	router.Get("/companies/:id/reviews/stats", reviewHandler.GetReviewStats)

	// Protected routes (authentication required)
	protected := router.Group("/", middleware.AuthMiddleware())

	// Review submission and interaction
	protected.Post("/companies/:id/reviews", reviewHandler.SubmitReview)
	protected.Post("/reviews/:id/helpful", reviewHandler.MarkReviewHelpful)

	// User's own reviews
	protected.Get("/user/reviews", reviewHandler.GetMyReviews)

	// Admin only routes
	admin := protected.Group("/admin", middleware.RequireRole("admin"))
	admin.Delete("/reviews/:id", reviewHandler.DeleteReview)
}