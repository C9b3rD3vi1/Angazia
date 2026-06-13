package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
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
	injectFlash(c, data)
	return data
}

func injectFlash(c *fiber.Ctx, data fiber.Map) {
	if flash := c.Locals("_flash"); flash != nil {
		if _, exists := data["Flash"]; !exists {
			data["Flash"] = flash
		}
	}
}

type WebHandler struct {
	companyService      services.CompanyService
	notificationService services.NotificationService
}

func NewWebHandler(companyService services.CompanyService) *WebHandler {
	return &WebHandler{
		companyService: companyService,
	}
}

func NewWebHandlerWithNotifications(
	companyService services.CompanyService,
	notificationService services.NotificationService,
) *WebHandler {
	return &WebHandler{
		companyService:      companyService,
		notificationService: notificationService,
	}
}

// HomePage renders the landing page
func (h *WebHandler) HomePage(c *fiber.Ctx) error {
	return c.Render("public/index", fiber.Map{
		"Title":       "Angazia - Find Your Dream Tech Job in Kenya",
		"Description": "Connect with top tech employers in Kenya. AI-powered job matching for developers, engineers, and tech professionals.",
		"ActivePage":  "home",
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
	data := fiber.Map{
		"Title": "Login to Angazia",
	}
	if flash := c.Query("flash"); flash != "" {
		data["Flash"] = utils.FlashMessage{
			Type:    c.Query("type", "info"),
			Message: flash,
		}
	}
	return c.Render("auth/login", data, "layouts/auth")
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

// TwoFASetupPage renders the 2FA setup page (for all authenticated roles)
func (h *WebHandler) TwoFASetupPage(c *fiber.Ctx) error {
	return c.Render("auth/2fa-setup", fiber.Map{
		"Title": "Two-Factor Authentication Setup - Angazia",
	}, "layouts/auth")
}

// TwoFAVerifyPage renders the 2FA verification page
func (h *WebHandler) TwoFAVerifyPage(c *fiber.Ctx) error {
	return c.Render("auth/2fa-verify", fiber.Map{
		"Title": "Verify Your Identity - Angazia",
	}, "layouts/auth")
}

// ── Employee pages (non-job) ──

// EmployeeApplicationsPage renders the candidate's applications
func (h *WebHandler) EmployeeApplicationsPage(c *fiber.Ctx) error {
	return c.Render("employee/applications", mergePageData(c, fiber.Map{
		"Title":      "My Applications - Angazia",
		"ActivePage": "applications",
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

// ── Employer pages (non-job) ──

// EmployerDashboardPage renders the employer dashboard
func (h *WebHandler) EmployerDashboardPage(c *fiber.Ctx) error {
	return c.Render("employer/dashboard", mergePageData(c, fiber.Map{
		"Title":      "Employer Dashboard - Angazia",
		"ActivePage": "dashboard",
	}), "layouts/employer")
}

// EmployerApplicationsPage renders the employer's application list
func (h *WebHandler) EmployerApplicationsPage(c *fiber.Ctx) error {
	return c.Render("employer/applications", mergePageData(c, fiber.Map{
		"Title":      "Applications - Angazia",
		"ActivePage": "applications",
	}), "layouts/employer")
}

// EmployerApplicationDetailPage renders the employer's application detail page
func (h *WebHandler) EmployerApplicationDetailPage(c *fiber.Ctx) error {
	return c.Render("employer/application-detail", mergePageData(c, fiber.Map{
		"Title":      "Application Detail - Angazia",
		"ActivePage": "applications",
	}), "layouts/employer")
}

// EmployerCandidatesPage renders the employer's candidate search
func (h *WebHandler) EmployerCandidatesPage(c *fiber.Ctx) error {
	return c.Render("employer/candidates", mergePageData(c, fiber.Map{
		"Title":      "Find Candidates - Angazia",
		"ActivePage": "candidates",
	}), "layouts/employer")
}

// EmployerTalentPoolPage renders the employer's talent pool
func (h *WebHandler) EmployerTalentPoolPage(c *fiber.Ctx) error {
	return c.Render("employer/talent-pool", mergePageData(c, fiber.Map{
		"Title":      "Talent Pool - Angazia",
		"ActivePage": "talent-pool",
	}), "layouts/employer")
}

// EmployerAnalyticsPage renders the employer analytics page
func (h *WebHandler) EmployerAnalyticsPage(c *fiber.Ctx) error {
	return c.Render("employer/analytics", mergePageData(c, fiber.Map{
		"Title":      "Analytics - Angazia",
		"ActivePage": "analytics",
	}), "layouts/employer")
}

// EmployerCompanyPage renders the employer company profile page
func (h *WebHandler) EmployerCompanyPage(c *fiber.Ctx) error {
	return c.Render("employer/company", mergePageData(c, fiber.Map{
		"Title":      "Company Profile - Angazia",
		"ActivePage": "company",
	}), "layouts/employer")
}

// EmployerMatchesPage renders the employer matches page
func (h *WebHandler) EmployerMatchesPage(c *fiber.Ctx) error {
	return c.Render("employer/matches", mergePageData(c, fiber.Map{
		"Title":      "AI Job Matches - Angazia",
		"ActivePage": "matches",
	}), "layouts/employer")
}

// EmployerCandidateDetailPage renders the employer's candidate detail page
func (h *WebHandler) EmployerCandidateDetailPage(c *fiber.Ctx) error {
	return c.Render("employer/candidate-detail", mergePageData(c, fiber.Map{
		"Title":      "Candidate Profile - Angazia",
		"ActivePage": "candidates",
	}), "layouts/employer")
}

// EmployerBillingPage renders the employer billing page
func (h *WebHandler) EmployerBillingPage(c *fiber.Ctx) error {
	return c.Render("employer/billing", mergePageData(c, fiber.Map{
		"Title":      "Billing - Angazia",
		"ActivePage": "billing",
	}), "layouts/employer")
}

// EmployerBillingInvoicesPage renders the employer invoices page
func (h *WebHandler) EmployerBillingInvoicesPage(c *fiber.Ctx) error {
	return c.Render("employer/billing-invoices", mergePageData(c, fiber.Map{
		"Title":      "Invoices - Angazia",
		"ActivePage": "billing",
	}), "layouts/employer")
}

// EmployerBillingUpgradePage renders the employer billing upgrade page
func (h *WebHandler) EmployerBillingUpgradePage(c *fiber.Ctx) error {
	return c.Render("employer/billing-upgrade", mergePageData(c, fiber.Map{
		"Title":      "Upgrade Plan - Angazia",
		"ActivePage": "billing",
	}), "layouts/employer")
}

// EmployerSettingsPage renders the employer settings page
func (h *WebHandler) EmployerSettingsPage(c *fiber.Ctx) error {
	return c.Render("employer/settings", mergePageData(c, fiber.Map{
		"Title":      "Settings - Angazia",
		"ActivePage": "settings",
	}), "layouts/employer")
}

// NotificationsPage renders the notifications page
func (h *WebHandler) NotificationsPage(c *fiber.Ctx) error {
	role, _ := c.Locals("user_role").(string)
	layout := "layouts/employee"
	if role == "employer" {
		layout = "layouts/employer"
	}
	return c.Render("employee/notifications", mergePageData(c, fiber.Map{
		"Title":      "Notifications - Angazia",
		"ActivePage": "notifications",
	}), layout)
}
