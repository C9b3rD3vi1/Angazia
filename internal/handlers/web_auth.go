package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type WebAuthHandler struct {
	authService services.AuthService
}

func NewWebAuthHandler(authService services.AuthService) *WebAuthHandler {
	return &WebAuthHandler{authService: authService}
}

func (h *WebAuthHandler) Logout(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	tokenStr := ""
	if cookie := c.Cookies("access_token"); cookie != "" {
		tokenStr = cookie
	}
	if userID != nil && tokenStr != "" {
		if uid, ok := userID.(string); ok {
			h.authService.Logout(c.Context(), uid, tokenStr)
		}
	}

	clearAuthCookies(c)
	return c.Redirect("/login", fiber.StatusFound)
}

func (h *WebAuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Redirect("/login", fiber.StatusFound)
	}

	req := &services.UpdateProfileRequest{
		FullName:          c.FormValue("full_name"),
		Headline:          c.FormValue("headline"),
		Bio:               c.FormValue("bio"),
		Location:          c.FormValue("location"),
		ExperienceLevel:   c.FormValue("experience_level"),
		PortfolioURL:      c.FormValue("portfolio_url"),
		LinkedInURL:       c.FormValue("linkedin_url"),
		IsAvailable:       c.FormValue("is_available") == "on",
		IsVisible:         c.FormValue("is_visible") == "on",
		IsRemoteOnly:      c.FormValue("is_remote_only") == "on",
	}

	yearsStr := c.FormValue("years_of_experience")
	if yearsStr != "" {
		if years, err := strconv.Atoi(yearsStr); err == nil {
			req.YearsOfExperience = years
		}
	}

	if err := h.authService.UpdateProfile(c.Context(), userID, req); err != nil {
		return c.Render("employee/settings", mergePageData(c, fiber.Map{
			"Title":      "Settings - Angazia",
			"ActivePage": "settings",
			"Error":      err.Error(),
			"Profile":    req,
		}), "layouts/employee")
	}

	return c.Redirect("/employee/settings?saved=1", fiber.StatusFound)
}

func (h *WebAuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Redirect("/login", fiber.StatusFound)
	}

	currentPassword := c.FormValue("current_password")
	newPassword := c.FormValue("new_password")
	confirmPassword := c.FormValue("confirm_password")

	if newPassword != confirmPassword {
		return c.Redirect("/employee/settings?error=Passwords+do+not+match", fiber.StatusFound)
	}

	if err := h.authService.ChangePassword(c.Context(), userID, currentPassword, newPassword); err != nil {
		return c.Redirect("/employee/settings?error=" + err.Error(), fiber.StatusFound)
	}

	return c.Redirect("/employee/settings?saved=1", fiber.StatusFound)
}

func (h *WebAuthHandler) NotificationPreferences(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Redirect("/login", fiber.StatusFound)
	}

	_ = c.FormValue("notify_job_alerts") == "on"
	_ = c.FormValue("notify_application_updates") == "on"
	_ = c.FormValue("notify_messages") == "on"
	_ = c.FormValue("notify_marketing") == "on"

	return c.Redirect("/employee/settings?saved=1", fiber.StatusFound)
}

func clearAuthCookies(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   !utils.IsDevelopment(),
		SameSite: "Strict",
		Path:     "/",
		MaxAge:   -1,
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   !utils.IsDevelopment(),
		SameSite: "Strict",
		Path:     "/api/v1/auth/refresh",
		MaxAge:   -1,
	})
}
