package routes

import (
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupAdminRoutes(router fiber.Router, adminHandler *handlers.AdminHandler) {
	// All admin routes require authentication and admin role
	admin := router.Group("/admin", middleware.AuthMiddleware(), middleware.RequireRole("admin"))

	// Dashboard & Statistics
	admin.Get("/stats/platform", adminHandler.GetPlatformStats)
	admin.Get("/stats/users", adminHandler.GetUserStats)
	admin.Get("/stats/jobs", adminHandler.GetJobStats)
	admin.Get("/stats/engagement", adminHandler.GetEngagementStats)
	admin.Get("/stats/chart-data", adminHandler.GetChartData)

	// Jobs
	admin.Get("/jobs", adminHandler.GetAdminJobs)
	admin.Get("/jobs/:id/applications", adminHandler.GetJobApplications)

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
	admin.Post("/settings", adminHandler.CreateSetting)
	admin.Put("/settings/:key", adminHandler.UpdateSetting)

	// Company Verification
	admin.Get("/companies", adminHandler.GetCompanies)
	admin.Get("/companies/pending", adminHandler.GetPendingVerifications)
	admin.Post("/companies/:id/verify", adminHandler.ApproveCompanyVerification)
	admin.Post("/companies/:id/reject", adminHandler.RejectCompanyVerification)

	// Report Reasons
	admin.Get("/report-reasons", adminHandler.GetReportReasons)
	admin.Post("/report-reasons", adminHandler.CreateReportReason)
	admin.Put("/report-reasons/:id", adminHandler.UpdateReportReason)
	admin.Delete("/report-reasons/:id", adminHandler.DeleteReportReason)

	// Audit Logs
	admin.Get("/audit-logs", adminHandler.GetAuditLogs)

	// Public report endpoint (authenticated users can report content)
	protected := router.Group("/report", middleware.AuthMiddleware())
	protected.Post("/", adminHandler.ReportContent)
}

func SetupAdminSubscriptionRoutes(router fiber.Router, subHandler *handlers.AdminSubscriptionHandler) {
	admin := router.Group("/admin/subscriptions", middleware.AuthMiddleware(), middleware.RequireRole("admin"))

	admin.Get("/", subHandler.ListSubscriptions)
	admin.Get("/:id", subHandler.GetSubscription)
	admin.Post("/:id/cancel", subHandler.CancelSubscription)
	admin.Post("/:id/reactivate", subHandler.ReactivateSubscription)
	admin.Post("/:id/change-plan", subHandler.ChangePlan)
	admin.Post("/", subHandler.AssignSubscription)
}

func SetupAdminContactRoutes(router fiber.Router, contactHandler *handlers.ContactHandler) {
	admin := router.Group("/admin/contacts", middleware.AuthMiddleware(), middleware.RequireRole("admin"))

	admin.Get("/", contactHandler.AdminList)
	admin.Get("/:id", contactHandler.AdminGet)
	admin.Post("/:id/read", contactHandler.AdminMarkRead)
	admin.Delete("/:id", contactHandler.AdminDelete)
}
