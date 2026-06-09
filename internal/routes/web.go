package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

func SetupWebRoutes(
	app *fiber.App,
	webHandler *handlers.WebHandler,
	authHandler *handlers.AuthHandler,
	authService services.AuthService,
	adminWebHandler *handlers.AdminWebHandler,
	notificationService services.NotificationService,
	jobService services.JobService,
) {
	// Public pages
	app.Get("/", webHandler.HomePage)
	app.Get("/jobs", webHandler.JobsPage)
	app.Get("/jobs/:id", webHandler.JobDetailPage)
	app.Get("/companies/:id", webHandler.CompanyPage)
	app.Get("/about", webHandler.AboutPage)
	app.Get("/contact", webHandler.ContactPage)
	app.Get("/pricing", webHandler.PricingPage)

	// Auth pages (GET - show form; POST - handled via JS API /api/v1/auth/*)
	app.Get("/login", middleware.WebGuestMiddleware(), webHandler.LoginPage)
	app.Get("/register", middleware.WebGuestMiddleware(), webHandler.RegisterPage)
	app.Get("/forgot-password", webHandler.ForgotPasswordPage)
	app.Get("/reset-password", webHandler.ResetPasswordPage)
	app.Get("/verify-email", webHandler.VerifyEmailPage)
	app.Post("/logout", authHandler.WebLogout)
	app.Get("/logout", authHandler.WebLogout)

	// Employee pages
	employee := app.Group("/employee",
		middleware.WebAuthMiddleware(),
		middleware.WebRequireRole("employee"),
		middleware.EmployeePageData(authService, notificationService),
	)
	employee.Get("/dashboard", webHandler.EmployeeDashboardPage)
	employee.Get("/jobs", webHandler.EmployeeJobsPage)
	employee.Get("/applications", webHandler.EmployeeApplicationsPage)
	employee.Get("/saved", webHandler.EmployeeSavedJobsPage)
	employee.Get("/alerts", webHandler.EmployeeJobAlertsPage)
	employee.Get("/skills", webHandler.EmployeeSkillsPage)
	employee.Get("/settings", webHandler.EmployeeSettingsPage)

	// Employer pages
	employer := app.Group("/employer",
		middleware.WebAuthMiddleware(),
		middleware.WebRequireRole("employer"),
		middleware.EmployerPageData(authService, notificationService, jobService),
	)
	employer.Get("/dashboard", webHandler.EmployerDashboardPage)
	employer.Get("/jobs", webHandler.EmployerJobsPage)
	employer.Get("/jobs/:id", webHandler.EmployerJobDetailPage)
	employer.Get("/candidates", webHandler.EmployerCandidatesPage)
	employer.Get("/candidates/:id", webHandler.EmployerCandidateDetailPage)
	employer.Get("/applications", webHandler.EmployerApplicationsPage)
	employer.Get("/talent-pool", webHandler.EmployerTalentPoolPage)
	employer.Get("/analytics", webHandler.EmployerAnalyticsPage)
	employer.Get("/company", webHandler.EmployerCompanyPage)
	employer.Get("/company-edit", webHandler.EmployerCompanyEditPage)
	employer.Get("/matches", webHandler.EmployerMatchesPage)
	employer.Get("/job-post", webHandler.EmployerJobPostPage)
	employer.Get("/job-edit/:id", webHandler.EmployerJobEditPage)
	employer.Get("/job-applications/:id", webHandler.EmployerJobApplicationsPage)
	employer.Get("/billing", webHandler.EmployerBillingPage)
	employer.Get("/billing/invoices", webHandler.EmployerBillingInvoicesPage)
	employer.Get("/billing/upgrade/:plan", webHandler.EmployerBillingUpgradePage)
	employer.Get("/settings", webHandler.EmployerSettingsPage)

	// Notifications page (authenticated users of any role)
	app.Get("/notifications", middleware.WebAuthMiddleware(), webHandler.NotificationsPage)

	// 2FA pages (authenticated users of any role)
	app.Get("/auth/2fa/setup", middleware.WebAuthMiddleware(), webHandler.TwoFASetupPage)
	app.Get("/auth/2fa/verify", webHandler.TwoFAVerifyPage)

	// Admin login portal (guest-only — redirects to dashboard if already authenticated)
	app.Get("/admin/login", middleware.WebGuestMiddleware(), adminWebHandler.LoginPage)
	app.Get("/admin/logout", adminWebHandler.LogoutPage)

	// Admin pages
	SetupAdminWebRoutes(app, adminWebHandler)
}

func SetupAdminWebRoutes(app *fiber.App, h *handlers.AdminWebHandler) {
	admin := app.Group("/admin", middleware.WebAuthMiddleware(), middleware.WebRequireRole("admin"))

	// Overview
	admin.Get("/dashboard", h.DashboardPage)
	admin.Get("/analytics", h.AnalyticsPage)

	// User management
	admin.Get("/users", h.UsersPage)
	admin.Get("/users/:id", h.UserDetailPage)

	// Company management
	admin.Get("/companies", h.CompaniesPage)
	admin.Get("/companies/:id", h.CompanyDetailPage)

	// Job management
	admin.Get("/jobs", h.JobsPage)
	admin.Get("/jobs/:id", h.JobDetailPage)

	// Moderation & reports
	admin.Get("/reports", h.ReportsPage)
	admin.Get("/audit", h.AuditLogsPage)

	// Subscriptions & billing
	admin.Get("/subscriptions", h.SubscriptionsPage)

	// Notifications
	admin.Get("/notifications", h.NotificationsPage)

	// Profile & settings
	admin.Get("/profile", h.ProfilePage)
	admin.Post("/profile/password", h.ProfileUpdatePassword)
	admin.Get("/settings", h.SettingsPage)
}
