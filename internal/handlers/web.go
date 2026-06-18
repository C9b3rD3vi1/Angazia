package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
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
	subscriptionService services.SubscriptionService
	jobService          services.JobService
	testimonialService  services.TestimonialService
	contactService      services.ContactService
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

func NewWebHandlerWithAll(
	companyService services.CompanyService,
	notificationService services.NotificationService,
	subscriptionService services.SubscriptionService,
	jobService services.JobService,
	testimonialService services.TestimonialService,
	contactService services.ContactService,
) *WebHandler {
	return &WebHandler{
		companyService:      companyService,
		notificationService: notificationService,
		subscriptionService: subscriptionService,
		jobService:          jobService,
		testimonialService:  testimonialService,
		contactService:      contactService,
	}
}

// HomePage renders the landing page
func (h *WebHandler) HomePage(c *fiber.Ctx) error {
	return c.Render("public/landing", fiber.Map{
		"Title":       "Angazia - Find Your Place in Kenya's Tech Ecosystem",
		"Description": "Angazia connects developers, engineers, and tech professionals with top Kenyan employers through AI-powered matching.",
		"ActivePage":  "home",
	}, "layouts/base")
}

// CompanyPage renders the company profile page
func (h *WebHandler) CompanyPage(c *fiber.Ctx) error {
	companyID := c.Params("id")

	profile, err := h.companyService.GetPublicCompanyProfile(c.Context(), companyID)
	if err != nil {
		return c.Render("public/company", fiber.Map{
			"Title":      "Company Not Found",
			"ActivePage": "companies",
		}, "layouts/base")
	}

	companyMap := fiber.Map{
		"name":        profile.CompanyName,
		"logo":        profile.CompanyLogo,
		"description": profile.Description,
		"industry":    profile.Industry,
		"size":        profile.CompanySize,
		"location":    profile.Location,
		"website":     profile.Website,
	}
	if profile.Rating > 0 {
		companyMap["rating"] = profile.Rating
	}
	if profile.ReviewCount > 0 {
		companyMap["review_count"] = profile.ReviewCount
	}
	if profile.Stats != nil {
		companyMap["employees"] = profile.Stats.TotalHires
	}

	var jobs []fiber.Map
	if h.jobService != nil {
		jobFilters := &services.JobFilters{CompanyName: profile.CompanyName}
		jobResult, jobErr := h.jobService.ListJobs(c.Context(), jobFilters, 1, 50)
		if jobErr == nil && jobResult != nil {
			for _, j := range jobResult.Jobs {
				if !j.IsActive {
					continue
				}
				jobs = append(jobs, fiber.Map{
					"id":         j.ID,
					"title":      j.Title,
					"location":   j.Location,
					"salary":     formatJobSalary(j),
					"type":       j.EmploymentType,
					"postedDate": j.PostedAt.Format("Jan 2, 2006"),
				})
			}
		}
	}

	return c.Render("public/company", fiber.Map{
		"Title":      profile.CompanyName + " - Angazia",
		"ActivePage": "companies",
		"company":    companyMap,
		"jobs":       jobs,
	}, "layouts/base")
}

// AboutPage renders the about page
func (h *WebHandler) AboutPage(c *fiber.Ctx) error {
	team := []fiber.Map{
		{"name": "James Mwangi", "role": "CEO & Co-Founder", "bio": "Former CTO at Safaricom, passionate about connecting Kenyan tech talent with opportunities."},
		{"name": "Grace Akinyi", "role": "CTO & Co-Founder", "bio": "Software engineer with 12+ years experience building scalable platforms across Africa."},
		{"name": "David Ochieng", "role": "Head of Product", "bio": "Product leader focused on creating delightful user experiences for job seekers and employers."},
		{"name": "Sarah Wanjiku", "role": "Head of Operations", "bio": "Operations expert ensuring Angazia runs smoothly and delivers value to every user."},
	}
	values := []fiber.Map{
		{"icon": "🤝", "title": "Community First", "description": "We believe in the power of community and building connections that matter."},
		{"icon": "🔬", "title": "Innovation Driven", "description": "Leveraging AI and data to create smarter matching between talent and opportunity."},
		{"icon": "🎯", "title": "Impact Focused", "description": "Measured by the careers launched and businesses transformed through our platform."},
		{"icon": "🌍", "title": "Pan-African Vision", "description": "Starting in Kenya, building for the continent, competing on the global stage."},
	}
	milestones := []fiber.Map{
		{"year": "2023", "title": "Platform Launch", "description": "Angazia launched with AI-powered matching for Kenyan tech professionals."},
		{"year": "2024", "title": "10,000 Users", "description": "Reached 10,000 active users and 300+ employer partners across Kenya."},
		{"year": "2025", "title": "Regional Expansion", "description": "Expanded operations to Uganda, Tanzania, and Rwanda."},
	}

	return c.Render("public/about", fiber.Map{
		"Title":      "About Angazia - Connecting Kenyan Tech Talent",
		"ActivePage": "about",
		"team":       team,
		"values":     values,
		"milestones": milestones,
	}, "layouts/base")
}

// ContactPage renders the contact page
func (h *WebHandler) ContactPage(c *fiber.Ctx) error {
	flash := ""
	if c.Query("success") == "1" {
		flash = "Thank you for your message! Our team will get back to you within 24 hours."
	}
	return c.Render("public/contact", fiber.Map{
		"Title":      "Contact Us",
		"ActivePage": "contact",
		"Flash":      flash,
	}, "layouts/base")
}

// ContactSubmit handles contact form submission
func (h *WebHandler) ContactSubmit(c *fiber.Ctx) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	subject := c.FormValue("subject")
	message := c.FormValue("message")

	if name == "" || email == "" || message == "" {
		return c.Redirect("/contact?success=0")
	}

	if err := h.contactService.Submit(c.Context(), name, email, subject, message); err != nil {
		log.Printf("Contact submission error: %v", err)
		return c.Redirect("/contact?success=0")
	}

	return c.Redirect("/contact?success=1")
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
	planID := c.Params("plan")
	if planID == "" {
		planID = "free"
	}
	return c.Render("employer/billing-upgrade", mergePageData(c, fiber.Map{
		"Title":      "Upgrade Plan - Angazia",
		"ActivePage": "billing",
		"Plan":       planID,
	}), "layouts/employer")
}

// InvoiceViewPage renders a printable invoice view
func (h *WebHandler) InvoiceViewPage(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(404).SendString("Invoice not found")
	}
	if h.subscriptionService == nil {
		return c.Status(500).SendString("Billing service not available")
	}
	invoice, err := h.subscriptionService.GetInvoice(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Invoice not found")
	}
	items, err := h.subscriptionService.GetInvoiceItems(c.Context(), id)
	if err != nil {
		items = []*models.InvoiceItem{}
	}
	userName := "N/A"
	userEmail := "N/A"
	if invoice.User != nil {
		userEmail = invoice.User.Email
	}
	return c.Render("public/invoice", fiber.Map{
		"Invoice": invoice,
		"Items":   items,
		"User":    fiber.Map{"Name": userName, "Email": userEmail},
		"Title":   "Invoice - " + invoice.InvoiceNumber,
	})
}

// EmployerTeamPage renders the employer team management page
func (h *WebHandler) EmployerTeamPage(c *fiber.Ctx) error {
	return c.Render("employer/team", mergePageData(c, fiber.Map{
		"Title":      "Team Management - Angazia",
		"ActivePage": "team",
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

// EmployeeTestimonialsPage renders the employee testimonials page
func (h *WebHandler) EmployeeTestimonialsPage(c *fiber.Ctx) error {
	return c.Render("employee/testimonials", mergePageData(c, fiber.Map{
		"Title":      "My Testimonials - Angazia",
		"ActivePage": "testimonials",
	}), "layouts/employee")
}

// EmployerTestimonialsPage renders the employer testimonials page
func (h *WebHandler) EmployerTestimonialsPage(c *fiber.Ctx) error {
	return c.Render("employer/testimonials", mergePageData(c, fiber.Map{
		"Title":      "Company Testimonials - Angazia",
		"ActivePage": "testimonials",
	}), "layouts/employer")
}
