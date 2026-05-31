package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupMatchingRoutes(router fiber.Router, matchingHandler *handlers.MatchingHandler) {
	// Protected routes (authentication required)
	protected := router.Group("/", middleware.AuthMiddleware())
	
	// Candidate matching endpoints
	employee := protected.Group("/employee", middleware.RequireRole("employee"))
	employee.Get("/matches/jobs", matchingHandler.GetJobMatches)
	employee.Post("/matches/cover-letter", matchingHandler.GenerateCoverLetter)
	employee.Get("/matches/skills-gap/:jobId", matchingHandler.AnalyzeSkillsGap)
	employee.Get("/matches/analysis/:jobId/:employeeId", matchingHandler.GetDetailedMatchAnalysis)
	
	// Employer matching endpoints
	employer := protected.Group("/employer", middleware.RequireRole("employer"))
	employer.Get("/matches/candidates/:jobId", matchingHandler.GetCandidateMatches)
	employer.Get("/matches/interview-questions/:jobId", matchingHandler.GenerateInterviewQuestions)
}