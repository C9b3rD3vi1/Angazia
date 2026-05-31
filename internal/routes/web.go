package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupWebRoutes(app *fiber.App, webHandler *handlers.WebHandler) {
	// Public pages
	app.Get("/", webHandler.HomePage)
	app.Get("/jobs", webHandler.JobsPage)
	app.Get("/jobs/:id", webHandler.JobDetailPage)
	app.Get("/companies/:id", webHandler.CompanyPage)
	app.Get("/about", webHandler.AboutPage)
	app.Get("/contact", webHandler.ContactPage)
	app.Get("/pricing", webHandler.PricingPage)
	
	// Auth pages
	app.Get("/login", webHandler.LoginPage)
	app.Get("/register", webHandler.RegisterPage)
	app.Get("/forgot-password", webHandler.ForgotPasswordPage)
	app.Get("/reset-password", webHandler.ResetPasswordPage)
	app.Get("/verify-email", webHandler.VerifyEmailPage)
	
	// Protected dashboard pages
	employee := app.Group("/employee", middleware.AuthMiddleware(), middleware.RequireRole("employee"))
	employee.Get("/dashboard", webHandler.EmployeeDashboardPage)
	// Add other employee routes...
	
	employer := app.Group("/employer", middleware.AuthMiddleware(), middleware.RequireRole("employer"))
	employer.Get("/dashboard", webHandler.EmployerDashboardPage)
	// Add other employer routes...
	
	admin := app.Group("/admin", middleware.AuthMiddleware(), middleware.RequireRole("admin"))
	admin.Get("/dashboard", webHandler.AdminDashboardPage)
	// Add other admin routes...
}