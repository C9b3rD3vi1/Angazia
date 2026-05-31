package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupJobRoutes(
	router fiber.Router,
	jobHandler *handlers.JobHandler,
) {
	// Public job routes (no authentication required)
	router.Get("/jobs", jobHandler.ListJobs)
	router.Get("/jobs/featured", jobHandler.GetFeaturedJobs)
	router.Get("/jobs/recent", jobHandler.GetRecentJobs)
	router.Get("/jobs/search", jobHandler.SearchJobs)
	router.Get("/jobs/:id", jobHandler.GetJob)
	router.Get("/jobs/:id/similar", jobHandler.GetSimilarJobs)
	
	// Protected job routes (authentication required)
	protected := router.Group("/", middleware.AuthMiddleware())
	
	// Candidate job actions
	protected.Post("/jobs/:id/save", jobHandler.SaveJob)
	protected.Delete("/jobs/:id/save", jobHandler.UnsaveJob)
	protected.Get("/employee/saved-jobs", jobHandler.GetSavedJobs)
	
	// Employer job management
	employer := protected.Group("/employer", middleware.RequireRole("employer"))
	employer.Post("/jobs", jobHandler.CreateJob)
	employer.Get("/jobs", jobHandler.ListMyJobs)
	employer.Put("/jobs/:id", jobHandler.UpdateJob)
	employer.Delete("/jobs/:id", jobHandler.DeleteJob)
	employer.Post("/jobs/:id/close", jobHandler.CloseJob)
}