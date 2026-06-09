package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
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
	return data
}

type WebHandler struct {
	jobService          services.JobService
	companyService      services.CompanyService
	notificationService services.NotificationService
}

func NewWebHandler(jobService services.JobService, companyService services.CompanyService) *WebHandler {
	return &WebHandler{
		jobService:     jobService,
		companyService: companyService,
	}
}

func NewWebHandlerWithNotifications(
	jobService services.JobService,
	companyService services.CompanyService,
	notificationService services.NotificationService,
) *WebHandler {
	return &WebHandler{
		jobService:          jobService,
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

	job, err := h.jobService.GetJob(c.Context(), jobID)
	if err != nil {
		return c.Render("public/job-detail", fiber.Map{
			"Title":      "Job Not Found",
			"JobID":      jobID,
			"ActivePage": "jobs",
		}, "layouts/base")
	}

	var requirements []string
	if job.Requirements != "" {
		requirements = strings.Split(job.Requirements, "\n")
		for i := range requirements {
			requirements[i] = strings.TrimSpace(requirements[i])
		}
	}

	var responsibilities []string
	if job.Responsibilities != "" {
		responsibilities = strings.Split(job.Responsibilities, "\n")
		for i := range responsibilities {
			responsibilities[i] = strings.TrimSpace(responsibilities[i])
		}
	}

	company, _ := h.companyService.GetCompanyProfile(c.Context(), job.EmployerID)

	companyName := ""
	if job.Employer != nil {
		companyName = job.Employer.CompanyName
	}

	jobMap := fiber.Map{
		"title":           job.Title,
		"company":         companyName,
		"companyID":       job.EmployerID,
		"location":        job.Location,
		"salary":          formatJobSalary(job),
		"type":            job.EmploymentType,
		"description":     job.Description,
		"requirements":    requirements,
		"responsibilities": responsibilities,
		"postedDate":      job.PostedAt.Format("Jan 2, 2006"),
		"expiresDate":     "",
	}

	if job.ExpiresAt != nil {
		jobMap["expiresDate"] = job.ExpiresAt.Format("Jan 2, 2006")
	}

	companyMap := fiber.Map{}
	if company != nil && company.Profile != nil {
		companyMap = fiber.Map{
			"name":     company.Profile.CompanyName,
			"logo":     company.Profile.CompanyLogo,
			"industry": company.Profile.Industry,
			"size":     company.Profile.CompanySize,
		}
	}

	similarJobs, _ := h.jobService.GetSimilarJobs(c.Context(), jobID, 5)
	var similarJobsMap []fiber.Map
	for _, sj := range similarJobs {
		companyName := ""
		if sj.Employer != nil {
			companyName = sj.Employer.CompanyName
		}
		similarJobsMap = append(similarJobsMap, fiber.Map{
			"id":      sj.ID,
			"title":   sj.Title,
			"company": companyName,
			"salary":  formatJobSalary(sj),
		})
	}

	return c.Render("public/job-detail", fiber.Map{
		"Title":       job.Title + " - Angazia",
		"JobID":       jobID,
		"ActivePage":  "jobs",
		"job":         jobMap,
		"company":     companyMap,
		"similarJobs": similarJobsMap,
	}, "layouts/base")
}

func formatJobSalary(job *models.Job) string {
	if job.SalaryMin == 0 && job.SalaryMax == 0 {
		return ""
	}
	currency := job.SalaryCurrency
	if currency == "" {
		currency = "KES"
	}
	if job.SalaryMin > 0 && job.SalaryMax > 0 {
		return fmt.Sprintf("%s %s - %s", currency, formatNumber(job.SalaryMin), formatNumber(job.SalaryMax))
	}
	if job.SalaryMin > 0 {
		return fmt.Sprintf("%s %s+", currency, formatNumber(job.SalaryMin))
	}
	if job.SalaryMax > 0 {
		return fmt.Sprintf("Up to %s %s", currency, formatNumber(job.SalaryMax))
	}
	return ""
}

func formatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
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

// EmployerApplicationsPage renders the employer's applications
func (h *WebHandler) EmployerApplicationsPage(c *fiber.Ctx) error {
	return c.Render("employer/applications", mergePageData(c, fiber.Map{
		"Title":      "Applications - Angazia",
		"ActivePage": "applications",
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

// EmployerJobDetailPage renders the employer's job detail page
func (h *WebHandler) EmployerJobDetailPage(c *fiber.Ctx) error {
	jobID := c.Params("id")
	return c.Render("employer/job-details", mergePageData(c, fiber.Map{
		"Title":      "Job Details - Angazia",
		"ActivePage": "jobs",
		"JobID":      jobID,
	}), "layouts/employer")
}

// EmployerJobEditPage renders the job edit page
func (h *WebHandler) EmployerJobEditPage(c *fiber.Ctx) error {
	jobID := c.Params("id")
	return c.Render("employer/job-edit", mergePageData(c, fiber.Map{
		"Title":      "Edit Job - Angazia",
		"ActivePage": "jobs",
		"JobID":      jobID,
	}), "layouts/employer")
}

// EmployerJobApplicationsPage renders the job-specific applications page
func (h *WebHandler) EmployerJobApplicationsPage(c *fiber.Ctx) error {
	jobID := c.Params("id")
	return c.Render("employer/job-applications", mergePageData(c, fiber.Map{
		"Title":      "Job Applications - Angazia",
		"ActivePage": "applications",
		"JobID":      jobID,
	}), "layouts/employer")
}

// EmployerBillingPage renders the billing/subscription page
func (h *WebHandler) EmployerBillingPage(c *fiber.Ctx) error {
	return c.Render("employer/billing", mergePageData(c, fiber.Map{
		"Title":      "Billing - Angazia",
		"ActivePage": "billing",
	}), "layouts/employer")
}

// EmployerBillingInvoicesPage renders the invoices history page
func (h *WebHandler) EmployerBillingInvoicesPage(c *fiber.Ctx) error {
	return c.Render("employer/billing-invoices", mergePageData(c, fiber.Map{
		"Title":      "Invoices - Angazia",
		"ActivePage": "billing",
	}), "layouts/employer")
}

// EmployerBillingUpgradePage renders the plan upgrade page
func (h *WebHandler) EmployerBillingUpgradePage(c *fiber.Ctx) error {
	plan := c.Params("plan")
	return c.Render("employer/billing-upgrade", mergePageData(c, fiber.Map{
		"Title":      "Upgrade Plan - Angazia",
		"ActivePage": "billing",
		"Plan":       plan,
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

// EmployerCandidateDetailPage renders a candidate's profile for employer view
func (h *WebHandler) EmployerCandidateDetailPage(c *fiber.Ctx) error {
	candidateID := c.Params("id")
	return c.Render("employer/candidate-detail", mergePageData(c, fiber.Map{
		"Title":       "Candidate Profile - Angazia",
		"ActivePage":  "candidates",
		"CandidateID": candidateID,
	}), "layouts/employer")
}

// EmployerSettingsPage renders the employer's settings
func (h *WebHandler) EmployerSettingsPage(c *fiber.Ctx) error {
	return c.Render("employer/settings", mergePageData(c, fiber.Map{
		"Title":      "Settings - Angazia",
		"ActivePage": "settings",
	}), "layouts/employer")
}

// NotificationsPage renders the notification center (for both employee & employer)
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
