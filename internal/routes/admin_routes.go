package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupAdminRoutes(router fiber.Router, adminHandler *handlers.AdminHandler) {
	// All admin routes require authentication and admin role
	admin := router.Group("/admin", middleware.AuthMiddleware(), middleware.RequireRole("admin"))
	
	// Dashboard & Statistics
	admin.Get("/stats/platform", adminHandler.GetPlatformStats)
	admin.Get("/stats/users", adminHandler.GetUserStats)
	admin.Get("/stats/jobs", adminHandler.GetJobStats)
	admin.Get("/stats/engagement", adminHandler.GetEngagementStats)
	
	// User Management
	admin.Get("/users", adminHandler.GetAllUsers)
	admin.Get("/users/:id", adminHandler.GetUserDetails)
	admin.Post("/users/:id/suspend", adminHandler.SuspendUser)
	admin.Post("/users/:id/activate", adminHandler.ActivateUser)
	admin.Delete("/users/:id", adminHandler.DeleteUser)
	admin.Post("/users/:id/verify", adminHandler.VerifyUser)
	
	// Moderation Queue
	admin.Get("/moderation", adminHandler.GetModerationQueue)
	admin.Post("/moderation/:id/approve", adminHandler.ApproveContent)
	admin.Post("/moderation/:id/reject", adminHandler.RejectContent)
	
	// System Settings
	admin.Get("/settings", adminHandler.GetSettings)
	admin.Put("/settings/:key", adminHandler.UpdateSetting)
	
	// Company Verification
	admin.Post("/companies/:id/verify", adminHandler.ApproveCompanyVerification)
	admin.Post("/companies/:id/reject", adminHandler.RejectCompanyVerification)

	// Report Reasons
	admin.Get("/report-reasons", adminHandler.GetReportReasons)
	
	// Audit Logs
	admin.Get("/audit-logs", adminHandler.GetAuditLogs)
	
	// Public report endpoint (authenticated users can report content)
	protected := router.Group("/report", middleware.AuthMiddleware())
	protected.Post("/", adminHandler.ReportContent)
}