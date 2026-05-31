package handlers

import (
	"github.com/gofiber/fiber/v2"
)

type WebHandler struct{}

func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

// HomePage renders the landing page
func (h *WebHandler) HomePage(c *fiber.Ctx) error {
	return c.Render("public/index", fiber.Map{
		"Title":       "Angazia - Find Your Dream Tech Job in Kenya",
		"Description": "Connect with top tech employers in Kenya. AI-powered job matching for developers, engineers, and tech professionals.",
		"ActivePage":  "home",
	})
}

// JobsPage renders the job listings page
func (h *WebHandler) JobsPage(c *fiber.Ctx) error {
	return c.Render("public/jobs", fiber.Map{
		"Title":      "Tech Jobs in Kenya - Find Your Next Role",
		"ActivePage": "jobs",
	})
}

// JobDetailPage renders the job details page
func (h *WebHandler) JobDetailPage(c *fiber.Ctx) error {
	jobID := c.Params("id")
	return c.Render("public/job-detail", fiber.Map{
		"Title":      "Job Details",
		"JobID":      jobID,
		"ActivePage": "jobs",
	})
}

// CompanyPage renders the company profile page
func (h *WebHandler) CompanyPage(c *fiber.Ctx) error {
	companyID := c.Params("id")
	return c.Render("public/company", fiber.Map{
		"Title":      "Company Profile",
		"CompanyID":  companyID,
		"ActivePage": "companies",
	})
}

// AboutPage renders the about page
func (h *WebHandler) AboutPage(c *fiber.Ctx) error {
	return c.Render("public/about", fiber.Map{
		"Title":      "About Angazia - Connecting Kenyan Tech Talent",
		"ActivePage": "about",
	})
}

// ContactPage renders the contact page
func (h *WebHandler) ContactPage(c *fiber.Ctx) error {
	return c.Render("public/contact", fiber.Map{
		"Title":      "Contact Us",
		"ActivePage": "contact",
	})
}

// PricingPage renders the pricing page
func (h *WebHandler) PricingPage(c *fiber.Ctx) error {
	return c.Render("public/pricing", fiber.Map{
		"Title":      "Pricing Plans - Simple, Transparent Pricing",
		"ActivePage": "pricing",
	})
}

// LoginPage renders the login page
func (h *WebHandler) LoginPage(c *fiber.Ctx) error {
	return c.Render("auth/login", fiber.Map{
		"Title": "Login to Angazia",
	})
}

// RegisterPage renders the registration page
func (h *WebHandler) RegisterPage(c *fiber.Ctx) error {
	return c.Render("auth/register", fiber.Map{
		"Title": "Create Account - Angazia",
	})
}

// ForgotPasswordPage renders the forgot password page
func (h *WebHandler) ForgotPasswordPage(c *fiber.Ctx) error {
	return c.Render("auth/forgot-password", fiber.Map{
		"Title": "Reset Password - Angazia",
	})
}

// ResetPasswordPage renders the reset password page
func (h *WebHandler) ResetPasswordPage(c *fiber.Ctx) error {
	token := c.Query("token")
	return c.Render("auth/reset-password", fiber.Map{
		"Title": "Set New Password",
		"Token": token,
	})
}

// VerifyEmailPage renders the email verification page
func (h *WebHandler) VerifyEmailPage(c *fiber.Ctx) error {
	token := c.Query("token")
	userID := c.Query("user_id")
	return c.Render("auth/verify-email", fiber.Map{
		"Title":   "Verify Email",
		"Token":   token,
		"UserID":  userID,
	})
}

// EmployeeDashboardPage renders the candidate dashboard
func (h *WebHandler) EmployeeDashboardPage(c *fiber.Ctx) error {
	return c.Render("employee/dashboard", fiber.Map{
		"Title":      "Dashboard - Angazia",
		"Layout":     "employee",
		"ActivePage": "dashboard",
	})
}

// EmployerDashboardPage renders the employer dashboard
func (h *WebHandler) EmployerDashboardPage(c *fiber.Ctx) error {
	return c.Render("employer/dashboard", fiber.Map{
		"Title":      "Employer Dashboard - Angazia",
		"Layout":     "employer",
		"ActivePage": "dashboard",
	})
}

// AdminDashboardPage renders the admin dashboard
func (h *WebHandler) AdminDashboardPage(c *fiber.Ctx) error {
	return c.Render("admin/dashboard", fiber.Map{
		"Title":      "Admin Dashboard - Angazia",
		"Layout":     "admin",
		"ActivePage": "dashboard",
	})
}