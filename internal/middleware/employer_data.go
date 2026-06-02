package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

func EmployerPageData(authService services.AuthService) fiber.Handler {
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

		if profile.EmployerProfile != nil {
			ep := profile.EmployerProfile
			companyName = ep.CompanyName
			verified = ep.IsVerified()
			plan = ep.SubscriptionPlan
			jobsUsed = ep.TotalJobsPosted

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
			"Name":     companyName,
			"Email":    profile.User.Email,
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
			"Logo":     "",
			"Verified": verified,
		}

		subMap := fiber.Map{
			"Plan":       plan,
			"Percentage": pct,
			"JobsUsed":   jobsUsed,
			"JobsLimit":  jobsLimit,
		}

		data := fiber.Map{
			"User":           userMap,
			"Company":        companyMap,
			"Subscription":   subMap,
			"JobCount":       0,
			"CandidateCount": 0,
			"UnreadCount":    0,
		}

		c.Locals("_pageData", data)
		return c.Next()
	}
}

func planLimit(plan string) int {
	switch plan {
	case "free":
		return 3
	case "basic":
		return 10
	case "pro":
		return 50
	case "enterprise":
		return 999
	default:
		return 3
	}
}
