package handlers

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type DashboardHandler struct {
	analyticsService         services.AnalyticsService
	subscriptionService      services.SubscriptionService
	candidateAnalyticsService services.CandidateAnalyticsService
	matchingService          services.MatchingService
	jobService               services.JobService
	applicationService       services.ApplicationService
}

func NewDashboardHandler(
	analyticsService services.AnalyticsService,
	subscriptionService services.SubscriptionService,
	candidateAnalyticsService services.CandidateAnalyticsService,
	matchingService services.MatchingService,
	jobService services.JobService,
	applicationService services.ApplicationService,
) *DashboardHandler {
	return &DashboardHandler{
		analyticsService:         analyticsService,
		subscriptionService:      subscriptionService,
		candidateAnalyticsService: candidateAnalyticsService,
		matchingService:          matchingService,
		jobService:               jobService,
		applicationService:       applicationService,
	}
}

func (h *DashboardHandler) GetEmployerDashboard(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	days, _ := strconv.Atoi(c.Query("days", "30"))
	if days < 1 {
		days = 30
	}
	if days > 365 {
		days = 365
	}

	dashboard, err := h.analyticsService.GetDashboard(c.Context(), userID.(string), days)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	if sub, subErr := h.subscriptionService.GetCurrentSubscription(c.Context(), userID.(string)); subErr == nil && sub != nil {
		jobsUsed := 0
		if dashboard.Stats != nil {
			jobsUsed = dashboard.Stats.ActiveJobs
		}
		dashboard.Subscription = &models.SubscriptionInfo{
			PlanName:  sub.PlanName,
			Amount:    sub.Amount,
			Currency:  sub.Currency,
			Interval:  sub.Interval,
			JobsUsed:  jobsUsed,
			JobsLimit: sub.JobPostLimit,
			Status:    sub.Status,
			Features:  sub.Features,
		}
	}

	return utils.Success(c, dashboard)
}

// EmployeeDashboardPage renders the candidate dashboard with real data
func (h *DashboardHandler) EmployeeDashboardPage(c *fiber.Ctx) error {
	ctx := c.Context()
	employeeID, _ := c.Locals("user_id").(string)

	data := fiber.Map{
		"Title":      "Dashboard - Angazia",
		"ActivePage": "dashboard",
	}

	stats := fiber.Map{
		"ProfileViews":         0,
		"ApplicationsSent":     0,
		"SavedJobs":            0,
		"InterviewInvites":     0,
		"ProfileViewsTrend":    0,
		"ProfileViewsPositive": true,
		"ApplicationsTrend":    0,
		"ApplicationsPositive": true,
		"InterviewRate":        0,
	}
	data["Stats"] = stats

	data["ProfileCompletion"] = 0

	appStatus := fiber.Map{
		"Pending":          0,
		"Viewed":           0,
		"Shortlisted":      0,
		"Interview":        0,
		"Rejected":         0,
		"PendingPercent":   0,
		"ViewedPercent":    0,
		"ShortlistedPercent": 0,
		"InterviewPercent": 0,
		"RejectedPercent":  0,
	}
	data["AppStatus"] = appStatus

	// ── 1. Fetch Candidate Dashboard ──
	dashboard, dashErr := h.candidateAnalyticsService.GetDashboard(ctx, employeeID)
	if dashErr == nil && dashboard != nil {
		if as := dashboard.ApplicationStats; as != nil {
			total := as.TotalApplications
			if total == 0 {
				total = as.PendingCount + as.InterviewCount + as.ShortlistedCount + as.RejectedCount + as.HiredCount + as.WithdrawnCount
			}

			stats["ApplicationsSent"] = total
			stats["InterviewInvites"] = as.InterviewCount

			viewedCount := 0
			if as.ByStatus != nil {
				viewedCount = as.ByStatus["viewed"]
			}
			if viewedCount == 0 && total > 0 {
				viewedCount = total - as.PendingCount - as.ShortlistedCount - as.InterviewCount - as.RejectedCount - as.HiredCount - as.WithdrawnCount
			}

			appStatus["Pending"] = as.PendingCount
			appStatus["Viewed"] = viewedCount
			appStatus["Shortlisted"] = as.ShortlistedCount
			appStatus["Interview"] = as.InterviewCount
			appStatus["Rejected"] = as.RejectedCount

			if total > 0 {
				appStatus["PendingPercent"] = math.Round(float64(as.PendingCount) / float64(total) * 100)
				appStatus["ViewedPercent"] = math.Round(float64(viewedCount) / float64(total) * 100)
				appStatus["ShortlistedPercent"] = math.Round(float64(as.ShortlistedCount) / float64(total) * 100)
				appStatus["InterviewPercent"] = math.Round(float64(as.InterviewCount) / float64(total) * 100)
				appStatus["RejectedPercent"] = math.Round(float64(as.RejectedCount) / float64(total) * 100)
			}

			if len(as.ByMonth) >= 2 {
				last := as.ByMonth[len(as.ByMonth)-1]
				prev := as.ByMonth[len(as.ByMonth)-2]
				appsTrend := 0
				if prev.Applications > 0 {
					appsTrend = int(math.Round(float64(last.Applications-prev.Applications) / float64(prev.Applications) * 100))
				}
				stats["ApplicationsTrend"] = appsTrend
				if appsTrend < 0 {
					stats["ApplicationsPositive"] = false
				}
			}

			reviewed := as.ShortlistedCount + as.InterviewCount + as.RejectedCount
			if reviewed > 0 {
				stats["InterviewRate"] = int(math.Round(float64(as.InterviewCount) / float64(reviewed) * 100))
			}
		}

		if len(dashboard.RecentActivity) > 0 {
			activityList := make([]fiber.Map, 0, len(dashboard.RecentActivity))
			for _, ra := range dashboard.RecentActivity {
				text := ra.Title
				if ra.Details != "" {
					text = ra.Details
				}
				ts := friendlyTime(ra.CreatedAt)
				activityList = append(activityList, fiber.Map{
					"Type":      ra.Type,
					"Text":      text,
					"Timestamp": ts,
				})
			}
			data["RecentActivity"] = activityList
		}

		if sga := dashboard.SkillGapAnalysis; sga != nil {
			if len(sga.MatchingSkills) > 0 {
				topSkills := make([]string, 0, len(sga.MatchingSkills))
				for _, ms := range sga.MatchingSkills {
					topSkills = append(topSkills, ms.Name)
				}
				data["TopSkills"] = topSkills
			}

			if len(sga.MissingSkills) > 0 {
				missingSkills := make([]string, 0, len(sga.MissingSkills))
				for _, mg := range sga.MissingSkills {
					missingSkills = append(missingSkills, mg.Name)
				}
				data["MissingSkills"] = missingSkills
			}
		}

		if ps := dashboard.ProfileStrength; ps != nil {
			data["ProfileCompletion"] = ps.OverallScore

			if len(ps.ImprovementTips) > 0 {
				tips := make([]fiber.Map, 0, len(ps.ImprovementTips))
				for _, tip := range ps.ImprovementTips {
					icon := "💡"
					switch tip.Priority {
					case "high":
						icon = "⚠️"
					case "medium":
						icon = "📌"
					}
					tips = append(tips, fiber.Map{
						"Icon": icon,
						"Text": tip.Description,
						"Link": tip.ActionURL,
					})
				}
				data["ProfileTips"] = tips
			}
		}
	}

	// ── 2. Fallback data from middleware page data ──
	if pd := c.Locals("_pageData"); pd != nil {
		if m, ok := pd.(fiber.Map); ok {
			if pv, ok := m["ProfileViews"]; ok {
				switch v := pv.(type) {
				case int:
					stats["ProfileViews"] = v
				case float64:
					stats["ProfileViews"] = int(v)
				}
			}
			// Fallback ProfileCompletion from middleware's ProfileStrength
			if pc, ok := data["ProfileCompletion"].(int); !ok || pc == 0 {
				if ps, ok := m["ProfileStrength"]; ok {
					switch v := ps.(type) {
					case int:
						data["ProfileCompletion"] = v
					case float64:
						data["ProfileCompletion"] = int(v)
					}
				}
			}
		}
	}

	// ── 3. Saved Jobs Count ──
	savedList, savedErr := h.jobService.GetSavedJobs(ctx, employeeID, 1, 1)
	if savedErr == nil && savedList != nil {
		stats["SavedJobs"] = savedList.Total
	}

	// ── 4. Job Recommendations ──
	matches, matchErr := h.matchingService.GetJobMatches(ctx, employeeID, 6)
	if matchErr == nil && len(matches) > 0 {
		recommendedJobs := make([]fiber.Map, 0, len(matches))
		for _, m := range matches {
			ci := ""
			for _, w := range strings.Fields(m.CompanyName) {
				if len(w) > 0 {
					ci += strings.ToUpper(w[:1])
				}
			}

			postedDate := ""
			if !m.AnalyzedAt.IsZero() {
				postedDate = friendlyDate(m.AnalyzedAt)
			}

			recommendedJobs = append(recommendedJobs, fiber.Map{
				"ID":              m.JobID,
				"Title":           m.JobTitle,
				"Company":         m.CompanyName,
				"CompanyLogo":     "",
				"CompanyInitials": ci,
				"MatchScore":      m.OverallScore,
				"MatchingSkills":  m.MatchingSkills,
				"MissingSkills":   m.MissingSkills,
				"Location":        "",
				"EmploymentType":  "",
				"Salary":          "",
				"PostedDate":      postedDate,
			})
		}
		data["RecommendedJobs"] = recommendedJobs
	}

	// ── 5. Upcoming Interviews ──
	appList, appErr := h.applicationService.ListMyApplications(ctx, employeeID, 1, 50)
	if appErr == nil && appList != nil && len(appList.Applications) > 0 {
		now := time.Now()
		type upcomingItem struct {
			date time.Time
			m    fiber.Map
		}
		items := make([]upcomingItem, 0)
		for _, a := range appList.Applications {
			if a.InterviewDate != nil && a.InterviewDate.After(now) {
				companyName := ""
				role := ""
				if a.Job != nil {
					role = a.Job.Title
					if a.Job.Employer != nil {
						companyName = a.Job.Employer.CompanyName
					}
				}
				items = append(items, upcomingItem{
					date: *a.InterviewDate,
					m: fiber.Map{
						"ID":      a.ID,
						"Day":     a.InterviewDate.Format("02"),
						"Month":   a.InterviewDate.Format("Jan"),
						"Role":    role,
						"Company": companyName,
						"Time":    a.InterviewDate.Format("15:04"),
						"Type":    a.InterviewType,
					},
				})
			}
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i].date.Before(items[j].date)
		})
		if len(items) > 0 {
			upcoming := make([]fiber.Map, 0, len(items))
			for _, item := range items {
				upcoming = append(upcoming, item.m)
			}
			data["UpcomingInterviews"] = upcoming
		}
	}

	// ── 6. Greeting, Date, Time ──
	data["Greeting"] = greeting()
	data["CurrentDate"] = time.Now().Format("Monday, January 2, 2006")
	data["CurrentTime"] = time.Now().Format("3:04 PM")

	return c.Render("employee/dashboard", mergePageData(c, data), "layouts/employee")
}

func greeting() string {
	h := time.Now().Hour()
	switch {
	case h < 12:
		return "Good morning"
	case h < 17:
		return "Good afternoon"
	default:
		return "Good evening"
	}
}

func friendlyTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "Just now"
	case d < 2*time.Minute:
		return "1 minute ago"
	case d < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	case d < 2*time.Hour:
		return "1 hour ago"
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	case d < 48*time.Hour:
		return "Yesterday"
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 2")
	}
}

func friendlyDate(t time.Time) string {
	return t.Format("2 Jan")
}
