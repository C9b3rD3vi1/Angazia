package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type AdminWebHandler struct {
	adminService        services.AdminService
	jobService          services.JobService
	subscriptionService services.SubscriptionService
	companyService      services.CompanyService
	authService         services.AuthService
}

func NewAdminWebHandler(adminService services.AdminService, jobService services.JobService, subscriptionService services.SubscriptionService, companyService services.CompanyService, authService services.AuthService) *AdminWebHandler {
	return &AdminWebHandler{
		adminService:        adminService,
		jobService:          jobService,
		subscriptionService: subscriptionService,
		companyService:      companyService,
		authService:         authService,
	}
}

func (h *AdminWebHandler) LoginPage(c *fiber.Ctx) error {
	return c.Render("admin/login", fiber.Map{
		"Title": "Admin Login - Angazia",
	}, "layouts/auth")
}

func (h *AdminWebHandler) LogoutPage(c *fiber.Ctx) error {
	c.ClearCookie("access_token", "refresh_token")
	return c.Redirect("/admin/login", fiber.StatusFound)
}

func (h *AdminWebHandler) sidebarData(ctx *fiber.Ctx) fiber.Map {
	stats, _ := h.adminService.GetPlatformStats(ctx.Context())
	userStats, _ := h.adminService.GetUserStats(ctx.Context())
	jobStats, _ := h.adminService.GetJobStats(ctx.Context())

	pendingUsers := 0
	if userStats != nil {
		if v, ok := userStats["pending"]; ok {
			pendingUsers = v
		}
	}

	pendingJobs := 0
	if jobStats != nil {
		if v, ok := jobStats["pending"]; ok {
			pendingJobs = v
		}
	}

	pendingVerifications := 0
	pendingReports := 0
	if stats != nil {
		pendingVerifications = int(stats.TotalEmployers - stats.VerifiedEmployers)
	}

	adminAvatar := ""
	adminInitials := "A"
	if uid, ok := ctx.Locals("user_id").(string); ok && uid != "" {
		if p, err := h.authService.GetProfile(ctx.Context(), uid); err == nil && p != nil && p.User != nil {
			adminAvatar = p.User.AvatarURL
			if name := p.User.FullName; name != "" {
				parts := strings.Fields(name)
				initials := ""
				for _, part := range parts {
					if len(part) > 0 {
						initials += strings.ToUpper(part[:1])
					}
				}
				if initials != "" {
					adminInitials = initials
				}
			}
		}
	}

	return fiber.Map{
		"PendingUsers":         pendingUsers,
		"PendingVerifications": pendingVerifications,
		"PendingJobs":          pendingJobs,
		"PendingReports":       pendingReports,
		"Env":                  "development",
		"Admin": fiber.Map{
			"Name":     ctx.Locals("user_email"),
			"Email":    ctx.Locals("user_email"),
			"Avatar":   adminAvatar,
			"Initials": adminInitials,
		},
	}
}

func (h *AdminWebHandler) DashboardPage(c *fiber.Ctx) error {
	stats, _ := h.adminService.GetPlatformStats(c.Context())
	userStats, _ := h.adminService.GetUserStats(c.Context())
	jobStats, _ := h.adminService.GetJobStats(c.Context())
	engagementStats, _ := h.adminService.GetEngagementStats(c.Context())

	recentUsers, _, _ := h.adminService.GetAllUsers(c.Context(), map[string]interface{}{}, 1, 10)

	var recentUserItems []fiber.Map
	for _, u := range recentUsers {
		initials := ""
		if len(u.FullName) > 0 {
			initials = string(u.FullName[0])
		}
		recentUserItems = append(recentUserItems, fiber.Map{
			"ID":             u.ID,
			"Name":           u.FullName,
			"Email":          u.Email,
			"Role":           u.Role,
			"Active":         u.IsActive,
			"Avatar":         "",
			"Initials":       initials,
			"RegisteredDate": u.CreatedAt.Format("2006-01-02"),
		})
	}

	var (
		totalUsers            int64
		totalCompanies        int64
		totalJobs             int64
		activeJobs            int64
		pendingVerifications  int64
		verifiedEmployers     int64
		activeUsers30Days     int
		newUsers7Days         int64
		newUsers30Days        int64
		totalApplications     int64
		totalRevenue          float64
		mrr                   float64
		userGrowthRate        float64
		jobGrowthRate         float64
		totalProfileViews     int
		totalJobViews         int
		avgMatchScore         float64
		avgResponseDays       int
		conversionRate        int
		verifiedCount         int
		pendingCount          int
		activeJobsCount       int
		inactiveJobsCount     int
	)

	if stats != nil {
		totalUsers = stats.TotalUsers
		totalCompanies = stats.TotalEmployers
		totalJobs = stats.TotalJobs
		activeJobs = stats.ActiveJobs
		verifiedEmployers = stats.VerifiedEmployers
		activeUsers30Days = stats.ActiveUsers30Days
		newUsers7Days = stats.NewUsers7Days
		newUsers30Days = stats.NewUsers30Days
		totalApplications = stats.TotalApplications
		totalRevenue = stats.TotalRevenue
		mrr = stats.MRR
		userGrowthRate = stats.UserGrowthRate
		jobGrowthRate = stats.JobGrowthRate
		totalProfileViews = stats.TotalProfileViews
		totalJobViews = stats.TotalJobViews
		avgMatchScore = stats.AverageMatchScore
	}
	pendingVerifications = totalCompanies - verifiedEmployers

	if userStats != nil {
		verifiedCount = userStats["verification_verified"]
		pendingCount = userStats["verification_pending"]
	}

	if jobStats != nil {
		activeJobsCount = jobStats["status_active"]
		inactiveJobsCount = jobStats["status_inactive"]
	}

	if engagementStats != nil {
		avgResponseDays = engagementStats["avg_response_days"]
		conversionRate = engagementStats["conversion_rate"]
	}

	data := fiber.Map{
		"Title":      "Admin Dashboard - Angazia",
		"ActivePage": "dashboard",
		"CurrentDate": time.Now().Format("Monday, January 2, 2006"),
		"Stats": fiber.Map{
			"TotalUsers":           totalUsers,
			"TotalCompanies":       totalCompanies,
			"TotalJobs":            totalJobs,
			"ActiveJobs":           activeJobs,
			"PendingVerifications": pendingVerifications,
			"PendingReports":       0,
			"TotalApplications":    totalApplications,
			"TotalRevenue":         totalRevenue,
			"MRR":                  mrr,
			"ActiveUsers30Days":    activeUsers30Days,
			"NewUsers7Days":        newUsers7Days,
			"NewUsers30Days":       newUsers30Days,
			"UserGrowthRate":       userGrowthRate,
			"JobGrowthRate":        jobGrowthRate,
			"TotalProfileViews":    totalProfileViews,
			"TotalJobViews":        totalJobViews,
			"AvgMatchScore":        avgMatchScore,
			"AvgResponseDays":      avgResponseDays,
			"ConversionRate":       conversionRate,
			"VerifiedEmployers":    verifiedCount,
			"PendingEmployers":     pendingCount,
			"ActiveJobsCount":      activeJobsCount,
			"InactiveJobsCount":    inactiveJobsCount,
		},
		"RecentUsers": recentUserItems,
		"SystemHealth": fiber.Map{
			"API":             true,
			"APILatency":      12,
			"Database":        true,
			"DatabaseLatency": 3,
			"Redis":           false,
			"Elasticsearch":   false,
		},
	}

	for k, v := range h.sidebarData(c) {
		data[k] = v
	}

	return c.Render("admin/dashboard", data, "layouts/admin")
}

func (h *AdminWebHandler) UsersPage(c *fiber.Ctx) error {
	data := fiber.Map{
		"Title":      "User Management - Angazia",
		"ActivePage": "users",
	}
	for k, v := range h.sidebarData(c) {
		data[k] = v
	}
	return c.Render("admin/users", data, "layouts/admin")
}

func (h *AdminWebHandler) UserDetailPage(c *fiber.Ctx) error {
	userID := c.Params("id")
	user, _ := h.adminService.GetUserDetails(c.Context(), userID)

	data := fiber.Map{
		"Title":      "User Detail - Angazia",
		"ActivePage": "users",
		"UserDetail": user,
	}
	for k, v := range h.sidebarData(c) {
		data[k] = v
	}
	return c.Render("admin/user-detail", data, "layouts/admin")
}

func (h *AdminWebHandler) CompaniesPage(c *fiber.Ctx) error {
	pendingItems, _, _ := h.adminService.GetModerationQueue(c.Context(), "company", "pending", 1, 100)
	allItems, totalComps, _ := h.adminService.GetModerationQueue(c.Context(), "company", "", 1, 100)

	type companyItem struct {
		ID                string
		Name              string
		Email             string
		Logo              string
		Initials          string
		VerificationStatus string
		JobsCount         int
		JoinedDate        string
		SubmittedDate     string
		DocumentCount     int
	}

	var pendingComps []companyItem
	for _, item := range pendingItems {
		initials := "C"
		if len(item.EntityID) > 0 {
			initials = string(item.EntityID[0])
		}
		pendingComps = append(pendingComps, companyItem{
			ID:                item.EntityID,
			Name:              item.EntityType,
			VerificationStatus: "pending",
			SubmittedDate:     item.CreatedAt.Format("2006-01-02"),
			Initials:          initials,
		})
	}

	var allComps []companyItem
	for _, item := range allItems {
		initials := "C"
		if len(item.EntityID) > 0 {
			initials = string(item.EntityID[0])
		}
		status := item.Status
		if status == "" {
			status = "unverified"
		}
		allComps = append(allComps, companyItem{
			ID:                item.EntityID,
			Name:              item.EntityType,
			VerificationStatus: status,
			JoinedDate:        item.CreatedAt.Format("2006-01-02"),
			Initials:          initials,
		})
	}

	data := fiber.Map{
		"Title":             "Company Management - Angazia",
		"ActivePage":        "companies",
		"PendingCompanies":  pendingComps,
		"Companies":         allComps,
		"TotalCompanies":    int(totalComps),
	}
	for k, v := range h.sidebarData(c) {
		data[k] = v
	}
	return c.Render("admin/company", data, "layouts/admin")
}

func (h *AdminWebHandler) CompanyDetailPage(c *fiber.Ctx) error {
	companyID := c.Params("id")

	detail := fiber.Map{
		"id":                   companyID,
		"name":                 "",
		"email":                "",
		"verification_status":  "unverified",
		"created_at":           "",
		"initials":             "?",
		"description":          "",
		"website":              "",
		"location":             "",
		"industry":             "",
		"size":                 "",
		"active_jobs":          0,
		"logo":                 nil,
		"verification_history": []fiber.Map{},
		"badges":               []fiber.Map{},
		"team_members":         []fiber.Map{},
		"review_stats": fiber.Map{
			"total":          0,
			"average_rating": 0.0,
			"five_star":      0,
		},
		"recent_reviews": []fiber.Map{},
	}

	// Fetch basic user info
	user, _ := h.adminService.GetUserDetails(c.Context(), companyID)
	if user != nil {
		detail["email"] = user.Email
		if !user.CreatedAt.IsZero() {
			detail["created_at"] = user.CreatedAt.Format("Jan 02, 2006")
		}
	}

	// Fetch rich company profile
	profileResp, err := h.companyService.GetCompanyProfile(c.Context(), companyID)
	if err == nil && profileResp != nil {
		p := profileResp.Profile
		if p != nil {
			name := p.CompanyName
			if name == "" && user != nil {
				name = user.FullName
			}
			detail["name"] = name
			detail["initials"] = computeInitials(name)
			detail["verification_status"] = p.VerificationStatus
			detail["description"] = p.CompanyDescription
			detail["website"] = p.CompanyWebsite
			detail["location"] = p.Location
			detail["industry"] = p.Industry
			detail["size"] = p.CompanySize
			detail["logo"] = p.CompanyLogo
		}

		if profileResp.Stats != nil {
			detail["active_jobs"] = profileResp.Stats.ActiveJobs
		}

		// Verification history
		if profileResp.Verification != nil {
			v := profileResp.Verification
			historyEntry := fiber.Map{
				"type":   v.Status,
				"action": "Verification " + v.Status,
				"date":   v.UpdatedAt.Format("Jan 02, 2006"),
				"note":   v.RejectionReason,
				"by":     "",
			}
			if v.VerifiedBy != nil {
				historyEntry["by"] = *v.VerifiedBy
			}
			detail["verification_history"] = []fiber.Map{historyEntry}
		}

		// Badges
		if len(profileResp.Badges) > 0 {
			var badges []fiber.Map
			for _, b := range profileResp.Badges {
				badges = append(badges, fiber.Map{
					"icon":        b.IconURL,
					"name":        b.BadgeName,
					"description": b.Description,
					"awarded":     b.IsActive,
				})
			}
			detail["badges"] = badges
		}
	}

	// Fetch team members
	teamMembers, err := h.companyService.GetTeamMembers(c.Context(), companyID)
	if err == nil && len(teamMembers) > 0 {
		var members []fiber.Map
		for _, m := range teamMembers {
			members = append(members, fiber.Map{
				"avatar":   "",
				"initials": computeInitials(m.FullName),
				"user_id":  m.ID,
				"name":     m.FullName,
				"role":     m.Role,
				"active":   true,
			})
		}
		detail["team_members"] = members
	}

	// Fetch review stats
	reviewStats, err := h.companyService.GetReviewStats(c.Context(), companyID)
	if err == nil && reviewStats != nil {
		fiveStar := 0
		if reviewStats.RatingDistribution != nil {
			fiveStar = reviewStats.RatingDistribution[5]
		}
		detail["review_stats"] = fiber.Map{
			"total":          reviewStats.TotalReviews,
			"average_rating": reviewStats.AverageRating,
			"five_star":      fiveStar,
		}
	}

	// Fetch recent reviews
	reviewsResp, err := h.companyService.GetCompanyReviews(c.Context(), companyID, 1, 5)
	if err == nil && reviewsResp != nil && len(reviewsResp.Reviews) > 0 {
		var reviews []fiber.Map
		for _, r := range reviewsResp.Reviews {
			author := ""
			if r.Reviewer != nil {
				author = r.Reviewer.Email
			}
			reviews = append(reviews, fiber.Map{
				"author":  author,
				"rating":  r.Rating,
				"date":    r.CreatedAt.Format("Jan 02, 2006"),
				"comment": r.Content,
			})
		}
		detail["recent_reviews"] = reviews
	}

	data := fiber.Map{
		"Title":         "Company Detail - Angazia",
		"ActivePage":    "companies",
		"CompanyDetail": detail,
	}
	for k, v := range h.sidebarData(c) {
		data[k] = v
	}
	return c.Render("admin/company-detail", data, "layouts/admin")
}

func computeInitials(name string) string {
	if name == "" {
		return "?"
	}
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "?"
	}
	var initials strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			initials.WriteByte(p[0])
		}
	}
	result := strings.ToUpper(initials.String())
	if len(result) > 2 {
		result = result[:2]
	}
	return result
}

func (h *AdminWebHandler) JobsPage(c *fiber.Ctx) error {
	jobStats, _ := h.adminService.GetJobStats(c.Context())
	jobList, _ := h.jobService.ListJobs(c.Context(), &services.JobFilters{}, 1, 100)

	totalJobs := 0
	activeJobs := 0
	pendingJobs := 0
	closedJobs := 0
	if jobStats != nil {
		totalJobs = jobStats["total"]
		activeJobs = jobStats["active"]
		pendingJobs = jobStats["pending"]
		closedJobs = jobStats["closed"]
	}

	var jobs []fiber.Map
	if jobList != nil {
		for _, j := range jobList.Jobs {
			companyName := ""
			companyInitials := ""
			if j.Employer != nil {
				companyName = j.Employer.CompanyName
				if len(companyName) > 0 {
					companyInitials = string(companyName[0])
				}
			}
			status := "active"
			if !j.IsActive {
				status = "closed"
			}
			companyLogo := ""
			if j.Employer != nil {
				companyLogo = j.Employer.CompanyLogo
			}
			jobs = append(jobs, fiber.Map{
				"ID":                j.ID,
				"Title":             j.Title,
				"CompanyID":         j.EmployerID,
				"CompanyName":       companyName,
				"CompanyInitials":   companyInitials,
				"CompanyLogo":       companyLogo,
				"Status":            status,
				"ApplicationsCount": j.ApplicationsCount,
				"ViewsCount":        j.ViewsCount,
				"PostedDate":        j.PostedAt.Format("2006-01-02"),
			})
		}
	}

	data := fiber.Map{
		"Title":      "Job Management - Angazia",
		"ActivePage": "jobs",
		"Stats": fiber.Map{
			"TotalJobs":  totalJobs,
			"ActiveJobs": activeJobs,
			"PendingJobs": pendingJobs,
			"ClosedJobs": closedJobs,
		},
		"TotalJobs": totalJobs,
		"Jobs":      jobs,
	}
	for k, v := range h.sidebarData(c) {
		data[k] = v
	}
	return c.Render("admin/jobs", data, "layouts/admin")
}

func (h *AdminWebHandler) JobDetailPage(c *fiber.Ctx) error {
	jobID := c.Params("id")
	job, _ := h.jobService.GetJob(c.Context(), jobID)

	var jobDetail fiber.Map
	if job != nil {
		companyName := ""
		companyInitials := ""
		if job.Employer != nil {
			companyName = job.Employer.CompanyName
			if len(companyName) > 0 {
				companyInitials = string(companyName[0])
			}
		}
		status := "active"
		if !job.IsActive {
			status = "closed"
		}
		companyLogo := ""
		if job.Employer != nil {
			companyLogo = job.Employer.CompanyLogo
		}
		jobDetail = fiber.Map{
			"ID":                job.ID,
			"Title":             job.Title,
			"Slug":              job.ID,
			"CompanyID":         job.EmployerID,
			"CompanyName":       companyName,
			"CompanyInitials":   companyInitials,
			"CompanyLogo":       companyLogo,
			"Status":            status,
			"Description":       job.Description,
			"Requirements":      job.Requirements,
			"Responsibilities":  job.Responsibilities,
			"Skills":            job.RequiredSkills,
			"Location":          job.Location,
			"Type":              job.EmploymentType,
			"Category":          job.ExperienceLevel,
			"SalaryRange":       "",
			"ExperienceLevel":   job.ExperienceLevel,
			"PostedDate":        job.PostedAt.Format("2006-01-02"),
			"ApplicationsCount": job.ApplicationsCount,
			"ViewsCount":        job.ViewsCount,
			"UniqueViews":       0,
			"SavesCount":        0,
		}
		if job.SalaryMin > 0 || job.SalaryMax > 0 {
			currency := job.SalaryCurrency
			if currency == "" {
				currency = "KES"
			}
			jobDetail["SalaryRange"] = fmt.Sprintf("%s %d - %d", currency, job.SalaryMin, job.SalaryMax)
		}
		if job.ExpiresAt != nil {
			jobDetail["ExpiresAt"] = job.ExpiresAt.Format("2006-01-02")
		}
	}

	data := fiber.Map{
		"Title":      "Job Detail - Angazia",
		"ActivePage": "jobs",
		"JobDetail":  jobDetail,
	}
	for k, v := range h.sidebarData(c) {
		data[k] = v
	}
	return c.Render("admin/job-detail", data, "layouts/admin")
}

func (h *AdminWebHandler) ReportsPage(c *fiber.Ctx) error {
	data := fiber.Map{
		"Title":            "Reports & Analytics - Angazia",
		"ActivePage":       "reports",
		"CurrentTab":       c.Query("tab", "reported"),
		"EntityTypeFilter": c.Query("entity_type", ""),
	}
	for k, v := range h.sidebarData(c) {
		data[k] = v
	}
	return c.Render("admin/reports", data, "layouts/admin")
}

func (h *AdminWebHandler) AuditLogsPage(c *fiber.Ctx) error {
	data := fiber.Map{
		"Title":            "Audit Logs - Angazia",
		"ActivePage":       "audit-logs",
		"ActionFilter":     c.Query("action", ""),
		"EntityTypeFilter": c.Query("entity_type", ""),
		"DateFrom":         c.Query("date_from", ""),
		"DateTo":           c.Query("date_to", ""),
		"Page":             1,
		"TotalPages":       0,
		"Limit":            20,
	}
	for k, v := range h.sidebarData(c) {
		data[k] = v
	}
	return c.Render("admin/audit-logs", data, "layouts/admin")
}

func (h *AdminWebHandler) AnalyticsPage(c *fiber.Ctx) error {
	stats, _ := h.adminService.GetPlatformStats(c.Context())
	userStats, _ := h.adminService.GetUserStats(c.Context())
	jobStats, _ := h.adminService.GetJobStats(c.Context())

	data := fiber.Map{
		"Title":      "System Analytics - Angazia",
		"ActivePage": "analytics",
		"Stats":      stats,
		"UserStats":  userStats,
		"JobStats":   jobStats,
	}
	for k, v := range h.sidebarData(c) {
		data[k] = v
	}
	return c.Render("admin/analytics", data, "layouts/admin")
}

func (h *AdminWebHandler) SubscriptionsPage(c *fiber.Ctx) error {
	plans, _ := h.subscriptionService.GetAllPlans(c.Context(), true)
	subs, totalSubs, _ := h.subscriptionService.GetAllSubscriptions(c.Context(), map[string]interface{}{}, 1, 50)

	data := fiber.Map{
		"Title":              "Subscription Plans - Angazia",
		"ActivePage":         "subscriptions",
		"Plans":              plans,
		"Subscriptions":      subs,
		"TotalSubscriptions": totalSubs,
	}
	for k, v := range h.sidebarData(c) {
		data[k] = v
	}
	return c.Render("admin/subscriptions", data, "layouts/admin")
}

func (h *AdminWebHandler) NotificationsPage(c *fiber.Ctx) error {
	data := fiber.Map{
		"Title":      "Admin Notifications - Angazia",
		"ActivePage": "notifications",
	}
	for k, v := range h.sidebarData(c) {
		data[k] = v
	}
	return c.Render("admin/notifications", data, "layouts/admin")
}

func (h *AdminWebHandler) ProfilePage(c *fiber.Ctx) error {
	email := c.Locals("user_email").(string)

	data := fiber.Map{
		"Title":      "Admin Profile - Angazia",
		"ActivePage": "profile",
		"ProfileUser": fiber.Map{
			"Email": email,
			"Role":  "admin",
			"Name":  email,
		},
	}
	for k, v := range h.sidebarData(c) {
		data[k] = v
	}
	return c.Render("admin/profile", data, "layouts/admin")
}

func (h *AdminWebHandler) ProfileUpdatePassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	oldPwd := c.FormValue("current_password")
	newPwd := c.FormValue("new_password")
	confirm := c.FormValue("confirm_password")

	email := c.Locals("user_email").(string)
	if newPwd != confirm {
		data := fiber.Map{"Title": "Admin Profile - Angazia", "ActivePage": "profile", "Error": "New passwords do not match", "ProfileUser": fiber.Map{"Email": email, "Role": "admin", "Name": email}}
		for k, v := range h.sidebarData(c) { data[k] = v }
		return c.Render("admin/profile", data, "layouts/admin")
	}
	if len(newPwd) < 8 {
		data := fiber.Map{"Title": "Admin Profile - Angazia", "ActivePage": "profile", "Error": "Password must be at least 8 characters", "ProfileUser": fiber.Map{"Email": email, "Role": "admin", "Name": email}}
		for k, v := range h.sidebarData(c) { data[k] = v }
		return c.Render("admin/profile", data, "layouts/admin")
	}

	if err := h.authService.ChangePassword(c.Context(), userID, oldPwd, newPwd); err != nil {
		data := fiber.Map{"Title": "Admin Profile - Angazia", "ActivePage": "profile", "Error": err.Error(), "ProfileUser": fiber.Map{"Email": email, "Role": "admin", "Name": email}}
		for k, v := range h.sidebarData(c) { data[k] = v }
		return c.Render("admin/profile", data, "layouts/admin")
	}

	data := fiber.Map{"Title": "Admin Profile - Angazia", "ActivePage": "profile", "Success": "Password changed successfully", "ProfileUser": fiber.Map{"Email": email, "Role": "admin", "Name": email}}
	for k, v := range h.sidebarData(c) { data[k] = v }
	return c.Render("admin/profile", data, "layouts/admin")
}

func (h *AdminWebHandler) SettingsPage(c *fiber.Ctx) error {
	settings, _ := h.adminService.GetSettings(c.Context(), "")

	grouped := make(map[string][]*models.SystemSetting)
	for _, s := range settings {
		grouped[s.Category] = append(grouped[s.Category], s)
	}

	data := fiber.Map{
		"Title":          "System Settings - Angazia",
		"ActivePage":     "settings",
		"Settings":       settings,
		"GroupedSettings": grouped,
	}
	for k, v := range h.sidebarData(c) {
		data[k] = v
	}
	return c.Render("admin/settings", data, "layouts/admin")
}
