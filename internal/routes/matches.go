package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/Angazia/internal/handlers"
)

// SetupMatchRoutes configures AI matching endpoints
func SetupMatchRoutes(router fiber.Router) {
	// For employees
	router.Get("/jobs", handlers.GetRecommendedJobs)
	router.Get("/jobs/:id/analysis", handlers.GetJobFitAnalysis)
	router.Post("/analyze-skills", handlers.AnalyzeSkillsGap)
	
	// For employers
	router.Get("/candidates", handlers.GetRecommendedCandidates)
	router.Get("/candidates/:id/analysis", handlers.GetCandidateFitAnalysis)
	router.Post("/jobs/:id/match-candidates", handlers.MatchCandidatesForJob)
	
	// Match feedback and improvement
	router.Post("/feedback", handlers.SubmitMatchFeedback)
	router.Get("/insights", handlers.GetMatchInsights)
}