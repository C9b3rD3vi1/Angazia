package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupSearchRoutes(router fiber.Router, searchHandler *handlers.SearchHandler) {
	// Public search endpoints
	router.Get("/search/jobs", searchHandler.SearchJobs)
	router.Get("/search/jobs/facets", searchHandler.GetJobFacets)
	router.Get("/search/companies", searchHandler.SearchCompanies)
	router.Get("/search/auto-complete", searchHandler.AutoComplete)
	router.Get("/search/popular", searchHandler.GetPopularSearches)
	
	// Protected search endpoints (require authentication)
	protected := router.Group("/search", middleware.AuthMiddleware())
	
	// Candidate search (employers only)
	protected.Get("/candidates", searchHandler.SearchCandidates)
	
	// Search history
	protected.Get("/history", searchHandler.GetSearchHistory)
	
	// Saved searches
	protected.Get("/saved", searchHandler.GetSavedSearches)
	protected.Post("/saved", searchHandler.SaveSearch)
	protected.Delete("/saved/:id", searchHandler.DeleteSavedSearch)
	protected.Get("/saved/:id/run", searchHandler.RunSavedSearch)
}