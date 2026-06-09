package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupApplicationRoutes(
	router fiber.Router,
	applicationHandler *handlers.ApplicationHandler,
) {
	// Protected routes (authentication required)
	protected := router.Group("/", middleware.AuthMiddleware())
	
	// Candidate application routes
	candidate := protected.Group("/employee", middleware.RequireRole("employee"))
	candidate.Post("/applications", applicationHandler.Apply)
	candidate.Get("/applications", applicationHandler.ListMyApplications)
	candidate.Get("/applications/stats", applicationHandler.GetApplicationStats)
	candidate.Post("/applications/:id/withdraw", applicationHandler.WithdrawApplication)
	
	// Employer application routes
	employer := protected.Group("/employer", middleware.RequireRole("employer"))
	employer.Get("/applications", applicationHandler.ListCompanyApplications)
	employer.Get("/jobs/:jobId/applications", applicationHandler.ListJobApplications)
	employer.Post("/applications/:id/shortlist", applicationHandler.ShortlistApplication)
	employer.Post("/applications/:id/reject", applicationHandler.RejectApplication)
	employer.Post("/applications/:id/interview", applicationHandler.ScheduleInterview)
	employer.Post("/applications/bulk-shortlist", applicationHandler.BulkShortlist)
	employer.Post("/applications/bulk-reject", applicationHandler.BulkReject)
	employer.Post("/applications/:id/hire", applicationHandler.MarkAsHired)
	
	// Shared application view (both roles can view individual applications)
	protected.Get("/applications/:id", applicationHandler.GetApplication)
}