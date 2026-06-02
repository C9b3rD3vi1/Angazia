package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type WebAuthHandler struct {
	authService services.AuthService
}

func NewWebAuthHandler(authService services.AuthService) *WebAuthHandler {
	return &WebAuthHandler{authService: authService}
}

func (h *WebAuthHandler) Login(c *fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	if email == "" || password == "" {
		return c.Render("auth/login", fiber.Map{
			"Title": "Login to Angazia",
			"Error": "Email and password are required",
			"Email": email,
		}, "layouts/auth")
	}

	ipAddress := c.IP()
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := c.Get("User-Agent")

	response, err := h.authService.Login(c.Context(), email, password, ipAddress, userAgent)
	if err != nil {
		return c.Render("auth/login", fiber.Map{
			"Title": "Login to Angazia",
			"Error": err.Error(),
			"Email": email,
		}, "layouts/auth")
	}

	setAuthCookies(c, response)

	redirect := c.FormValue("redirect")
	if redirect == "" {
		switch response.User.Role {
		case "employee":
			redirect = "/employee/dashboard"
		case "employer":
			redirect = "/employer/dashboard"
		case "admin":
			redirect = "/admin/dashboard"
		default:
			redirect = "/"
		}
	}
	return c.Redirect(redirect, fiber.StatusFound)
}

func (h *WebAuthHandler) Register(c *fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	fullName := c.FormValue("full_name")
	role := c.FormValue("role")

	if email == "" || password == "" {
		return c.Render("auth/register", fiber.Map{
			"Title":    "Create Account - Angazia",
			"Error":    "Email and password are required",
			"Email":    email,
			"FullName": fullName,
		}, "layouts/auth")
	}

	if role != "employee" && role != "employer" {
		role = "employee"
	}

	ipAddress := c.IP()
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}

	_, err := h.authService.Register(c.Context(), &services.RegisterRequest{
		Email:     email,
		Password:  password,
		FullName:  fullName,
		Role:      models.UserRole(role),
		IPAddress: ipAddress,
	})
	if err != nil {
		return c.Render("auth/register", fiber.Map{
			"Title":    "Create Account - Angazia",
			"Error":    err.Error(),
			"Email":    email,
			"FullName": fullName,
		}, "layouts/auth")
	}

	return c.Redirect("/login?registered=1", fiber.StatusFound)
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

func (h *WebAuthHandler) ForgotPassword(c *fiber.Ctx) error {
	email := c.FormValue("email")

	if email == "" {
		return c.Render("auth/forgot-password", fiber.Map{
			"Title": "Reset Password - Angazia",
			"Error": "Email is required",
		}, "layouts/auth")
	}

	ipAddress := c.IP()
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}

	err := h.authService.ForgotPassword(c.Context(), email, ipAddress)
	if err != nil {
		return c.Render("auth/forgot-password", fiber.Map{
			"Title": "Reset Password - Angazia",
			"Error": err.Error(),
			"Email": email,
		}, "layouts/auth")
	}

	return c.Render("auth/forgot-password", fiber.Map{
		"Title":   "Reset Password - Angazia",
		"Success": "If the email exists, a password reset link has been sent.",
		"Email":   email,
	}, "layouts/auth")
}

func (h *WebAuthHandler) ResetPassword(c *fiber.Ctx) error {
	token := c.FormValue("token")
	password := c.FormValue("password")

	if token == "" || password == "" {
		return c.Render("auth/reset-password", fiber.Map{
			"Title": "Set New Password",
			"Token": token,
			"Error": "Token and password are required",
		}, "layouts/auth")
	}

	err := h.authService.ResetPassword(c.Context(), token, password)
	if err != nil {
		return c.Render("auth/reset-password", fiber.Map{
			"Title": "Set New Password",
			"Token": token,
			"Error": err.Error(),
		}, "layouts/auth")
	}

	return c.Redirect("/login?reset=1", fiber.StatusFound)
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

	// Notification preferences stored separately; for now just acknowledge.
	return c.Redirect("/employee/settings?saved=1", fiber.StatusFound)
}

func setAuthCookies(c *fiber.Ctx, response *services.AuthResponse) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    response.AccessToken,
		HTTPOnly: true,
		Secure:   !utils.IsDevelopment(),
		SameSite: "Strict",
		Path:     "/",
		MaxAge:   3600 * 24,
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    response.RefreshToken,
		HTTPOnly: true,
		Secure:   !utils.IsDevelopment(),
		SameSite: "Strict",
		Path:     "/api/v1/auth/refresh",
		MaxAge:   7 * 24 * 3600,
	})
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
