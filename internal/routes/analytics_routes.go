package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupAnalyticsRoutes(router fiber.Router, analyticsHandler *handlers.AnalyticsHandler) {
	// All analytics routes require authentication and employer role
	analytics := router.Group("/employer/analytics", middleware.AuthMiddleware(), middleware.RequireRole("employer"))
	
	// Application trends
	analytics.Get("/trends", analyticsHandler.GetApplicationTrends)
	
	// Conversion funnel
	analytics.Get("/funnel", analyticsHandler.GetConversionFunnel)
	
	// Job performance
	analytics.Get("/jobs", analyticsHandler.GetJobPerformance)
	analytics.Get("/jobs/:jobId", analyticsHandler.GetJobPerformanceByID)
	
	// Time to hire
	analytics.Get("/time-to-hire", analyticsHandler.GetTimeToHireMetrics)
	
	// Application quality
	analytics.Get("/quality", analyticsHandler.GetApplicationQualityScores)
	
	// Source analytics
	analytics.Get("/sources", analyticsHandler.GetSourceAnalytics)
	
	// Export
	analytics.Get("/export", analyticsHandler.ExportAnalytics)
}