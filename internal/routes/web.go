package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/Angazia/internal/handlers"
)

func setupWebRoutes(app *fiber.App, webHandler *handlers.WebHandler) {
	// Landing page
	app.Get("/", webHandler.HomePage)
	
	// Authentication pages
	app.Get("/login", webHandler.LoginPage)
	app.Get("/register", webHandler.RegisterPage)
	app.Get("/forgot-password", webHandler.ForgotPasswordPage)
	app.Get("/reset-password", webHandler.ResetPasswordPage)
	app.Get("/verify-email", webHandler.VerifyEmailPage)
	
	// GitHub setup page
	app.Get("/github-setup", webHandler.GitHubSetupPage)
	
	// Employee pages
	app.Get("/employee/dashboard", webHandler.EmployeeDashboardPage)
	app.Get("/employee/profile", webHandler.EmployeeProfilePage)
	app.Get("/employee/profile/edit", webHandler.EmployeeProfileEditPage)
	app.Get("/employee/matches", webHandler.MatchesPage)
	app.Get("/employee/applications", webHandler.ApplicationsPage)
	app.Get("/employee/saved-jobs", webHandler.SavedJobsPage)
	app.Get("/employee/settings", webHandler.EmployeeSettingsPage)
	
	// Employer pages
	app.Get("/employer/dashboard", webHandler.EmployerDashboardPage)
	app.Get("/employer/company", webHandler.CompanyProfilePage)
	app.Get("/employer/jobs", webHandler.EmployerJobsPage)
	app.Get("/employer/jobs/post", webHandler.PostJobPage)
	app.Get("/employer/jobs/:id/edit", webHandler.EditJobPage)
	app.Get("/employer/jobs/:id/applicants", webHandler.ApplicantsPage)
	app.Get("/employer/talent-pool", webHandler.TalentPoolPage)
	app.Get("/employer/billing", webHandler.BillingPage)
	
	// Shared pages
	app.Get("/jobs", webHandler.JobsPage)
	app.Get("/jobs/:id", webHandler.JobDetailPage)
	app.Get("/companies/:name", webHandler.CompanyPage)
	app.Get("/search", webHandler.SearchPage)
	
	// Error pages
	app.Get("/404", webHandler.NotFoundPage)
	app.Get("/500", webHandler.ErrorPage)
}