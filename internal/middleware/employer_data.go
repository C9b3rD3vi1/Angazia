package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

func EmployerPageData(authService services.AuthService, notificationService services.NotificationService, jobService services.JobService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, _ := c.Locals("user_id").(string)
		if userID == "" {
			return c.Next()
		}

		profile, err := authService.GetProfile(c.Context(), userID)
		if err != nil || profile == nil {
			return c.Next()
		}

		companyName := ""
		initials := ""
		verified := false
		plan := "free"
		jobsUsed := 0
		logo := ""

		if profile.EmployerProfile != nil {
			ep := profile.EmployerProfile
			companyName = ep.CompanyName
			verified = ep.IsVerified()
			plan = ep.SubscriptionPlan
			jobsUsed = ep.TotalJobsPosted
			logo = ep.CompanyLogo

			parts := strings.Fields(ep.CompanyName)
			for _, p := range parts {
				if len(p) > 0 {
					initials += strings.ToUpper(p[:1])
				}
			}
		}

		if companyName == "" {
			companyName = profile.User.CompanyName
		}
		if companyName == "" {
			companyName = profile.User.Email
		}
		if initials == "" && len(companyName) > 0 {
			initials = strings.ToUpper(companyName[:1])
		}

		userMap := fiber.Map{
			"ID":       profile.User.ID,
			"Name":     companyName,
			"Email":    profile.User.Email,
			"Role":     profile.User.Role,
			"Avatar":   profile.User.AvatarURL,
			"Initials": initials,
		}

		jobsLimit := planLimit(plan)
		pct := 0
		if jobsLimit > 0 {
			pct = (jobsUsed * 100) / jobsLimit
		}

		companyMap := fiber.Map{
			"Name":     companyName,
			"Initials": initials,
			"Logo":     logo,
			"Verified": verified,
		}

		subMap := fiber.Map{
			"Plan":       plan,
			"Percentage": pct,
			"JobsUsed":   jobsUsed,
			"JobsLimit":  jobsLimit,
		}

		unreadCount := 0
		if counts, err := notificationService.GetUnreadCount(c.Context(), userID); err == nil {
			unreadCount = counts.TotalUnread
		}

		jobCount := 0
		candidateCount := 0
		if stats, err := jobService.GetJobStats(c.Context(), userID); err == nil {
			jobCount = int(stats.ActiveJobs)
			candidateCount = int(stats.TotalApplications)
		}

		data := fiber.Map{
			"User":           userMap,
			"Company":        companyMap,
			"Subscription":   subMap,
			"JobCount":       jobCount,
			"CandidateCount": candidateCount,
			"UnreadCount":    unreadCount,
		}

		c.Locals("_pageData", data)
		return c.Next()
	}
}

func planLimit(plan string) int {
	switch plan {
	case "free":
		return 3
	case "pro_monthly", "pro_yearly":
		return 20
	case "business_monthly", "business_yearly":
		return 100
	case "enterprise":
		return 1000
	default:
		return 3
	}
}
