package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupCandidateAnalyticsRoutes(router fiber.Router, handler *handlers.CandidateAnalyticsHandler) {
	// All routes require authentication and employee role
	analytics := router.Group("/employee/analytics", middleware.AuthMiddleware(), middleware.RequireRole("employee"))
	
	// Dashboard
	analytics.Get("/dashboard", handler.GetDashboard)
	
	// Profile strength
	analytics.Get("/profile-strength", handler.GetProfileStrength)
	
	// Application statistics
	analytics.Get("/applications/stats", handler.GetApplicationStats)
	analytics.Get("/applications/monthly", handler.GetMonthlyStats)
	
	// Success metrics
	analytics.Get("/success-rates", handler.GetSuccessRates)
	
	// Skill analysis
	analytics.Get("/skill-gap", handler.GetSkillGapAnalysis)
	
	// Market positioning
	analytics.Get("/market-positioning", handler.GetMarketPositioning)
	
	// Recommendations
	analytics.Get("/recommendations", handler.GetRecommendations)
	
	// Activity feed
	analytics.Get("/recent-activity", handler.GetRecentActivity)
}