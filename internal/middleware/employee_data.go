package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

func EmployeePageData(authService services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, _ := c.Locals("user_id").(string)
		if userID == "" {
			return c.Next()
		}

		profile, err := authService.GetProfile(c.Context(), userID)
		if err != nil || profile == nil {
			return c.Next()
		}

		userName := profile.User.FullName
		if userName == "" {
			userName = profile.User.Email
		}

		initials := ""
		parts := strings.Fields(userName)
		for _, p := range parts {
			if len(p) > 0 {
				initials += strings.ToUpper(p[:1])
			}
		}

		userMap := fiber.Map{
			"Name":     userName,
			"Email":    profile.User.Email,
			"Initials": initials,
			"Headline": "",
		}

		data := fiber.Map{
			"User":              userMap,
			"ProfileStrength":   0,
			"SkillsScore":       0,
			"SkillsOffset":      264,
			"SkillsLabel":       "No skills yet",
			"MatchCount":        0,
			"ApplicationCount":  0,
			"AlertCount":        0,
			"UnreadCount":       0,
		}

		if profile.EmployeeProfile != nil {
			ep := profile.EmployeeProfile
			userMap["Headline"] = ep.Headline

			data["ProfileStrength"] = calcProfileStrength(ep)
			skillsScore := min(len(ep.Skills)*10, 100)
			data["SkillsScore"] = skillsScore
			data["SkillsOffset"] = 264 - (264 * skillsScore / 100)
			if len(ep.Skills) == 0 {
				data["SkillsLabel"] = "No skills listed"
			} else if len(ep.Skills) == 1 {
				data["SkillsLabel"] = "1 skill listed"
			} else {
				data["SkillsLabel"] = fmt.Sprintf("%d skills listed", len(ep.Skills))
			}
			data["MatchCount"] = len(ep.Skills)
			data["ApplicationCount"] = ep.ApplicationCount
		}

		c.Locals("_pageData", data)
		return c.Next()
	}
}

func calcProfileStrength(ep *models.EmployeeProfile) int {
	s := 0
	if ep.FullName != "" {
		s += 10
	}
	if ep.Headline != "" {
		s += 15
	}
	if ep.Bio != "" {
		s += 15
	}
	if ep.Location != "" {
		s += 10
	}
	if len(ep.Skills) > 0 {
		s += 20
	}
	if ep.ResumeURL != "" {
		s += 15
	}
	if ep.PortfolioURL != "" || ep.LinkedInURL != "" {
		s += 10
	}
	if ep.YearsOfExperience > 0 {
		s += 5
	}
	return min(s, 100)
}
