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
	webAuthHandler *handlers.WebAuthHandler,
	authService services.AuthService,
) {
	// Public pages
	app.Get("/", webHandler.HomePage)
	app.Get("/jobs", webHandler.JobsPage)
	app.Get("/jobs/:id", webHandler.JobDetailPage)
	app.Get("/companies/:id", webHandler.CompanyPage)
	app.Get("/about", webHandler.AboutPage)
	app.Get("/contact", webHandler.ContactPage)
	app.Get("/pricing", webHandler.PricingPage)

	// Auth pages (GET - show form, POST - handle submission)
	app.Get("/login", middleware.WebGuestMiddleware(), webHandler.LoginPage)
	app.Post("/login", webAuthHandler.Login)
	app.Get("/register", middleware.WebGuestMiddleware(), webHandler.RegisterPage)
	app.Post("/register", webAuthHandler.Register)
	app.Get("/forgot-password", webHandler.ForgotPasswordPage)
	app.Post("/forgot-password", webAuthHandler.ForgotPassword)
	app.Get("/reset-password", webHandler.ResetPasswordPage)
	app.Post("/reset-password", webAuthHandler.ResetPassword)
	app.Get("/verify-email", webHandler.VerifyEmailPage)
	app.Post("/logout", webAuthHandler.Logout)
	app.Get("/logout", webAuthHandler.Logout)

	// Employee pages
	employee := app.Group("/employee",
		middleware.WebAuthMiddleware(),
		middleware.WebRequireRole("employee"),
		middleware.EmployeePageData(authService),
	)
	employee.Get("/dashboard", webHandler.EmployeeDashboardPage)
	employee.Get("/jobs", webHandler.EmployeeJobsPage)
	employee.Get("/applications", webHandler.EmployeeApplicationsPage)
	employee.Get("/saved", webHandler.EmployeeSavedJobsPage)
	employee.Get("/alerts", webHandler.EmployeeJobAlertsPage)
	employee.Get("/skills", webHandler.EmployeeSkillsPage)
	employee.Get("/settings", webHandler.EmployeeSettingsPage)
	employee.Post("/settings", webAuthHandler.UpdateProfile)
	employee.Post("/settings/password", webAuthHandler.ChangePassword)
	employee.Post("/settings/notifications", webAuthHandler.NotificationPreferences)

	// Employer pages
	employer := app.Group("/employer",
		middleware.WebAuthMiddleware(),
		middleware.WebRequireRole("employer"),
		middleware.EmployerPageData(authService),
	)
	employer.Get("/dashboard", webHandler.EmployerDashboardPage)
	employer.Get("/jobs", webHandler.EmployerJobsPage)
	employer.Get("/candidates", webHandler.EmployerCandidatesPage)
	employer.Get("/talent-pool", webHandler.EmployerTalentPoolPage)
	employer.Get("/analytics", webHandler.EmployerAnalyticsPage)
	employer.Get("/company", webHandler.EmployerCompanyPage)
	employer.Get("/company-edit", webHandler.EmployerCompanyEditPage)
	employer.Get("/matches", webHandler.EmployerMatchesPage)
	employer.Get("/job-post", webHandler.EmployerJobPostPage)
	employer.Get("/billing", webHandler.EmployerBillingPage)

	// Admin pages
	admin := app.Group("/admin", middleware.WebAuthMiddleware(), middleware.WebRequireRole("admin"))
	admin.Get("/dashboard", webHandler.AdminDashboardPage)
}
