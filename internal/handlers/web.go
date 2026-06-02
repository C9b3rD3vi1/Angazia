package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func mergePageData(c *fiber.Ctx, data fiber.Map) fiber.Map {
	if pd := c.Locals("_pageData"); pd != nil {
		if m, ok := pd.(fiber.Map); ok {
			for k, v := range m {
				if _, exists := data[k]; !exists {
					data[k] = v
				}
			}
		}
	}
	return data
}

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
	}, "layouts/base")
}

// JobsPage renders the job listings page
func (h *WebHandler) JobsPage(c *fiber.Ctx) error {
	return c.Render("public/jobs", fiber.Map{
		"Title":      "Tech Jobs in Kenya - Find Your Next Role",
		"ActivePage": "jobs",
	}, "layouts/base")
}

// JobDetailPage renders the job details page
func (h *WebHandler) JobDetailPage(c *fiber.Ctx) error {
	jobID := c.Params("id")
	return c.Render("public/job-detail", fiber.Map{
		"Title":      "Job Details",
		"JobID":      jobID,
		"ActivePage": "jobs",
	}, "layouts/base")
}

// CompanyPage renders the company profile page
func (h *WebHandler) CompanyPage(c *fiber.Ctx) error {
	companyID := c.Params("id")
	return c.Render("public/company", fiber.Map{
		"Title":      "Company Profile",
		"CompanyID":  companyID,
		"ActivePage": "companies",
	}, "layouts/base")
}

// AboutPage renders the about page
func (h *WebHandler) AboutPage(c *fiber.Ctx) error {
	return c.Render("public/about", fiber.Map{
		"Title":      "About Angazia - Connecting Kenyan Tech Talent",
		"ActivePage": "about",
	}, "layouts/base")
}

// ContactPage renders the contact page
func (h *WebHandler) ContactPage(c *fiber.Ctx) error {
	return c.Render("public/contact", fiber.Map{
		"Title":      "Contact Us",
		"ActivePage": "contact",
	}, "layouts/base")
}

// PricingPage renders the pricing page
func (h *WebHandler) PricingPage(c *fiber.Ctx) error {
	return c.Render("public/pricing", fiber.Map{
		"Title":      "Pricing Plans - Simple, Transparent Pricing",
		"ActivePage": "pricing",
	}, "layouts/base")
}

// LoginPage renders the login page
func (h *WebHandler) LoginPage(c *fiber.Ctx) error {
	return c.Render("auth/login", fiber.Map{
		"Title": "Login to Angazia",
	}, "layouts/auth")
}

// RegisterPage renders the registration page
func (h *WebHandler) RegisterPage(c *fiber.Ctx) error {
	return c.Render("auth/register", fiber.Map{
		"Title": "Create Account - Angazia",
	}, "layouts/auth")
}

// ForgotPasswordPage renders the forgot password page
func (h *WebHandler) ForgotPasswordPage(c *fiber.Ctx) error {
	return c.Render("auth/forgot-password", fiber.Map{
		"Title": "Reset Password - Angazia",
	}, "layouts/auth")
}

// ResetPasswordPage renders the reset password page
func (h *WebHandler) ResetPasswordPage(c *fiber.Ctx) error {
	token := c.Query("token")
	return c.Render("auth/reset-password", fiber.Map{
		"Title": "Set New Password",
		"Token": token,
	}, "layouts/auth")
}

// VerifyEmailPage renders the email verification page
func (h *WebHandler) VerifyEmailPage(c *fiber.Ctx) error {
	token := c.Query("token")
	userID := c.Query("user_id")
	return c.Render("auth/verify-email", fiber.Map{
		"Title":  "Verify Email",
		"Token":  token,
		"UserID": userID,
	}, "layouts/auth")
}

// EmployeeDashboardPage renders the candidate dashboard
func (h *WebHandler) EmployeeDashboardPage(c *fiber.Ctx) error {
	return c.Render("employee/dashboard", mergePageData(c, fiber.Map{
		"Title":      "Dashboard - Angazia",
		"ActivePage": "dashboard",
	}), "layouts/employee")
}

// EmployeeJobsPage renders the job listings with AI matching
func (h *WebHandler) EmployeeJobsPage(c *fiber.Ctx) error {
	return c.Render("employee/jobs", mergePageData(c, fiber.Map{
		"Title":      "Find Jobs - Angazia",
		"ActivePage": "jobs",
	}), "layouts/employee")
}

// EmployeeApplicationsPage renders the candidate's applications
func (h *WebHandler) EmployeeApplicationsPage(c *fiber.Ctx) error {
	return c.Render("employee/applications", mergePageData(c, fiber.Map{
		"Title":      "My Applications - Angazia",
		"ActivePage": "applications",
	}), "layouts/employee")
}

// EmployeeSavedJobsPage renders the candidate's saved jobs
func (h *WebHandler) EmployeeSavedJobsPage(c *fiber.Ctx) error {
	return c.Render("employee/saved", mergePageData(c, fiber.Map{
		"Title":      "Saved Jobs - Angazia",
		"ActivePage": "saved",
	}), "layouts/employee")
}

// EmployeeJobAlertsPage renders the candidate's job alerts
func (h *WebHandler) EmployeeJobAlertsPage(c *fiber.Ctx) error {
	return c.Render("employee/alerts", mergePageData(c, fiber.Map{
		"Title":      "Job Alerts - Angazia",
		"ActivePage": "alerts",
	}), "layouts/employee")
}

// EmployeeSkillsPage renders the candidate's skills profile
func (h *WebHandler) EmployeeSkillsPage(c *fiber.Ctx) error {
	return c.Render("employee/skills", mergePageData(c, fiber.Map{
		"Title":      "Skills Profile - Angazia",
		"ActivePage": "skills",
	}), "layouts/employee")
}

// EmployeeSettingsPage renders the candidate's settings
func (h *WebHandler) EmployeeSettingsPage(c *fiber.Ctx) error {
	return c.Render("employee/settings", mergePageData(c, fiber.Map{
		"Title":      "Settings - Angazia",
		"ActivePage": "settings",
	}), "layouts/employee")
}

// EmployerDashboardPage renders the employer dashboard
func (h *WebHandler) EmployerDashboardPage(c *fiber.Ctx) error {
	return c.Render("employer/dashboard", mergePageData(c, fiber.Map{
		"Title":      "Employer Dashboard - Angazia",
		"ActivePage": "dashboard",
	}), "layouts/employer")
}

// EmployerJobsPage renders the employer's job listings
func (h *WebHandler) EmployerJobsPage(c *fiber.Ctx) error {
	return c.Render("employer/jobs", mergePageData(c, fiber.Map{
		"Title":      "My Jobs - Angazia",
		"ActivePage": "jobs",
	}), "layouts/employer")
}

// EmployerCandidatesPage renders the candidate search
func (h *WebHandler) EmployerCandidatesPage(c *fiber.Ctx) error {
	return c.Render("employer/candidates", mergePageData(c, fiber.Map{
		"Title":      "Candidates - Angazia",
		"ActivePage": "candidates",
	}), "layouts/employer")
}

// EmployerTalentPoolPage renders the employer's talent pool
func (h *WebHandler) EmployerTalentPoolPage(c *fiber.Ctx) error {
	return c.Render("employer/talent-pool", mergePageData(c, fiber.Map{
		"Title":      "Talent Pool - Angazia",
		"ActivePage": "talent",
	}), "layouts/employer")
}

// EmployerAnalyticsPage renders the employer analytics
func (h *WebHandler) EmployerAnalyticsPage(c *fiber.Ctx) error {
	return c.Render("employer/analytics", mergePageData(c, fiber.Map{
		"Title":      "Analytics - Angazia",
		"ActivePage": "analytics",
	}), "layouts/employer")
}

// EmployerCompanyPage renders the employer's company profile
func (h *WebHandler) EmployerCompanyPage(c *fiber.Ctx) error {
	return c.Render("employer/company", mergePageData(c, fiber.Map{
		"Title":      "Company Profile - Angazia",
		"ActivePage": "company",
	}), "layouts/employer")
}

// EmployerJobPostPage renders the job posting form
func (h *WebHandler) EmployerJobPostPage(c *fiber.Ctx) error {
	return c.Render("employer/job-post", mergePageData(c, fiber.Map{
		"Title":      "Post a Job - Angazia",
		"ActivePage": "jobs",
	}), "layouts/employer")
}

// EmployerBillingPage renders the billing/subscription page
func (h *WebHandler) EmployerBillingPage(c *fiber.Ctx) error {
	return c.Render("employer/billing", mergePageData(c, fiber.Map{
		"Title":      "Billing - Angazia",
		"ActivePage": "billing",
	}), "layouts/employer")
}

// EmployerMatchesPage renders the AI candidate matches page
func (h *WebHandler) EmployerMatchesPage(c *fiber.Ctx) error {
	return c.Render("employer/matches", mergePageData(c, fiber.Map{
		"Title":      "AI Matches - Angazia",
		"ActivePage": "matches",
	}), "layouts/employer")
}

// EmployerCompanyEditPage renders the company profile edit form
func (h *WebHandler) EmployerCompanyEditPage(c *fiber.Ctx) error {
	return c.Render("employer/company-edit", mergePageData(c, fiber.Map{
		"Title":      "Edit Company Profile - Angazia",
		"ActivePage": "company",
	}), "layouts/employer")
}

// AdminDashboardPage renders the admin dashboard
func (h *WebHandler) AdminDashboardPage(c *fiber.Ctx) error {
	return c.Render("admin/dashboard", fiber.Map{
		"Title":      "Admin Dashboard - Angazia",
		"ActivePage": "dashboard",
	}, "layouts/admin")
}
