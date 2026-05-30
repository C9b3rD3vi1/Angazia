package routes

import (
	"github.com/gofiber/fiber/v2"
	
	"github.com/Angazia/internal/config"
	"github.com/Angazia/internal/handlers"
)

func SetupRoutes(
	app *fiber.App,
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	githubHandler *handlers.GitHubHandler,
	jobHandler *handlers.JobHandler,
	applicationHandler *handlers.ApplicationHandler,
	employeeHandler *handlers.EmployeeHandler,
	employerHandler *handlers.EmployerHandler,
	webHandler *handlers.WebHandler,
) {
	// API v1 group
	api := app.Group("/api/v1")
	
	// ========== PUBLIC ROUTES (No Auth) ==========
	// Health
	app.Get("/health", healthCheck)
	
	// Web routes (HTML pages)
	setupWebRoutes(app, webHandler)
	
	// Auth routes
	setupAuthRoutes(api, authHandler)
	
	// GitHub OAuth routes
	setupGitHubRoutes(api, githubHandler)
	
	// Public job routes (read-only)
	api.Get("/jobs", jobHandler.ListJobs)
	api.Get("/jobs/:id", jobHandler.GetJobDetails)
	api.Get("/companies/:name", jobHandler.GetCompanyJobs)
	
	// ========== PROTECTED ROUTES (Auth Required) ==========
	protected := api.Group("/", authMiddleware)
	
	// User profile routes
	protected.Get("/profile", authHandler.GetProfile)
	protected.Put("/profile", authHandler.UpdateProfile)
	
	// Job routes for authenticated users
	protected.Post("/jobs/:id/save", jobHandler.SaveJob)
	protected.Delete("/jobs/:id/save", jobHandler.UnsaveJob)
	protected.Get("/saved-jobs", jobHandler.GetSavedJobs)
	
	// Application routes
	setupApplicationRoutes(protected, applicationHandler)
	
	// Match routes
	protected.Get("/matches", matchingHandler.GetMatches)
	protected.Get("/matches/:id", matchingHandler.GetMatchDetails)
	protected.Post("/matches/:id/feedback", matchingHandler.SubmitFeedback)
	
	// ========== EMPLOYEE ONLY ROUTES ==========
	employeeRoutes := protected.Group("/employee", employeeOnly)
	setupEmployeeRoutes(employeeRoutes, employeeHandler, jobHandler, applicationHandler)
	
	// ========== EMPLOYER ONLY ROUTES ==========
	employerRoutes := protected.Group("/employer", employerOnly)
	setupEmployerRoutes(employerRoutes, employerHandler, jobHandler)
}

func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"time":   fiber.Now(),
	})
}