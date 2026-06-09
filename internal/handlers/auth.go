package handlers

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	// "github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
)

type AuthHandler struct {
	authService services.AuthService
	validator   *validator.Validate
	cfg         *config.Config
	twoFAService services.TwoFAService
}

func (h *AuthHandler) SetTwoFAService(s services.TwoFAService) {
	h.twoFAService = s
}

type RegisterRequest struct {
	Email     string          `json:"email" validate:"required,email"`
	Password  string          `json:"password" validate:"required,min=8,max=72"`
	FullName  string          `json:"full_name" validate:"omitempty,max=100"`
	Role      models.UserRole `json:"role" validate:"required,oneof=employee employer"`
}

type LoginRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

type UpdateProfileRequest struct {
	FullName          string   `json:"full_name"`
	Headline          string   `json:"headline"`
	Bio               string   `json:"bio"`
	Location          string   `json:"location"`
	IsRemoteOnly      bool     `json:"is_remote_only"`
	ExperienceLevel   string   `json:"experience_level"`
	YearsOfExperience int      `json:"years_of_experience"`
	Skills            []string `json:"skills"`
	ResumeURL         string   `json:"resume_url"`
	PortfolioURL      string   `json:"portfolio_url"`
	LinkedInURL       string   `json:"linkedin_url"`
	IsVisible         bool     `json:"is_visible"`
	IsAvailable       bool     `json:"is_available"`
	
	// Employer specific
	CompanyName        string `json:"company_name"`
	CompanyWebsite     string `json:"company_website"`
	CompanyLinkedIn    string `json:"company_linkedin"`
	CompanyDescription string `json:"company_description"`
	Industry           string `json:"industry"`
	CompanySize        string `json:"company_size"`
}

func NewAuthHandler(authService services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
		cfg:         cfg,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 409 {object} utils.APIResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	ipAddress := c.IP()
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = strings.Split(forwarded, ",")[0]
	}
	
	authReq := &services.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FullName:  req.FullName,
		Role:      req.Role,
		IPAddress: ipAddress,
	}
	
	response, err := h.authService.Register(c.Context(), authReq)
	if err != nil {
		message := err.Error()
		
		if strings.Contains(err.Error(), "already exists") {
			return utils.Conflict(c, message)
		} else if strings.Contains(err.Error(), "invalid") {
			return utils.BadRequest(c, message)
		}
		
		return utils.InternalServerError(c, message)
	}
	
	return utils.SuccessCreated(c, "User registered successfully. Please check your email for verification.", response)
}

// Login handles user authentication
// @Summary Login user
// @Description Authenticate user and return tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	ipAddress := c.IP()
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = strings.Split(forwarded, ",")[0]
	}
	userAgent := c.Get("User-Agent")
	
	response, err := h.authService.Login(c.Context(), req.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		message := err.Error()
		
		if strings.Contains(err.Error(), "verification") {
			return utils.Forbidden(c, message)
		}
		
		return utils.Unauthorized(c, message)
	}

	if h.twoFAService != nil {
		userID := response.User.ID
		if enabled, _ := h.twoFAService.IsEnabled(c.Context(), userID); enabled {
			response.RequiresTwoFA = true
		}
	}
	
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    response.AccessToken,
		HTTPOnly: true,
		Secure:   !utils.IsDevelopment(),
		SameSite: "Strict",
		Path:     "/",
		MaxAge:   24 * 3600,
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
	
	return utils.SuccessWithMessage(c, "Login successful", response)
}

type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *AuthHandler) AdminLogin(c *fiber.Ctx) error {
	var req AdminLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if req.Email != h.cfg.AdminEmail || req.Password != h.cfg.AdminPassword {
		return utils.Unauthorized(c, "Invalid admin credentials")
	}

	ctx := c.Context()
	user, err := h.authService.GetOrCreateAdmin(ctx, req.Email)
	if err != nil {
		return utils.InternalServerError(c, "Failed to authenticate admin")
	}

	ipAddress := c.IP()
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = strings.Split(forwarded, ",")[0]
	}
	userAgent := c.Get("User-Agent")

	response, err := h.authService.GenerateTokens(ctx, user.ID, user.Role, user.Email, ipAddress, userAgent)
	if err != nil {
		return utils.InternalServerError(c, "Failed to generate tokens")
	}

	if h.twoFAService != nil {
		if enabled, _ := h.twoFAService.IsEnabled(ctx, user.ID); enabled {
			response.RequiresTwoFA = true
		}
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    response.AccessToken,
		HTTPOnly: true,
		Secure:   !utils.IsDevelopment(),
		SameSite: "Strict",
		Path:     "/",
		MaxAge:   24 * 3600,
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

	return utils.SuccessWithMessage(c, "Admin login successful", response)
}

// Logout handles user logout
// @Summary Logout user
// @Description Invalidate user tokens
// @Tags Auth
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	authHeader := c.Get("Authorization")
	token := ""
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}
	
	if err := h.authService.Logout(c.Context(), userID.(string), token); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	c.Cookie(&fiber.Cookie{Name: "access_token", Value: "", Path: "/", MaxAge: -1})
	c.Cookie(&fiber.Cookie{Name: "refresh_token", Value: "", Path: "/api/v1/auth/refresh", MaxAge: -1})
	
	return utils.SuccessWithMessage(c, "Logout successful", nil)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	
	refreshToken := c.Cookies("refresh_token")
	
	if refreshToken == "" {
		if err := c.BodyParser(&req); err == nil && req.RefreshToken != "" {
			refreshToken = req.RefreshToken
		}
	}
	
	if refreshToken == "" {
		return utils.BadRequest(c, "Refresh token is required")
	}
	
	ipAddress := c.IP()
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = strings.Split(forwarded, ",")[0]
	}
	
	response, err := h.authService.RefreshToken(c.Context(), refreshToken, ipAddress)
	if err != nil {
		return utils.Unauthorized(c, err.Error())
	}
	
	if c.Get("X-Client-Type") == "web" {
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
	
	return utils.SuccessWithMessage(c, "Token refreshed successfully", response)
}

// ForgotPassword handles password reset request
// @Summary Request password reset
// @Description Send password reset email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Email address"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	ipAddress := c.IP()
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = strings.Split(forwarded, ",")[0]
	}
	
	message := "If an account exists with this email, you will receive a password reset link"
	
	if err := h.authService.ForgotPassword(c.Context(), req.Email, ipAddress); err != nil {
		return utils.SuccessWithMessage(c, message, nil)
	}
	
	return utils.SuccessWithMessage(c, message, nil)
}

// ResetPassword handles password reset
// @Summary Reset password
// @Description Reset password using token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset password details"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.authService.ResetPassword(c.Context(), req.Token, req.NewPassword); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Password reset successfully. You can now login with your new password.", nil)
}

// ChangePassword handles password change for authenticated users
// @Summary Change password
// @Description Change user password
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Password change details"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.authService.ChangePassword(c.Context(), userID.(string), req.OldPassword, req.NewPassword); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Password changed successfully", nil)
}

// VerifyEmail handles email verification
// @Summary Verify email address
// @Description Verify user email using token
// @Tags Auth
// @Param token path string true "Verification token"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Router /auth/verify-email/{token} [get]
func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return utils.BadRequest(c, "Verification token is required")
	}
	
	if err := h.authService.VerifyEmail(c.Context(), token); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Email verified successfully. You can now login to your account.", nil)
}

// ResendVerificationEmail resends verification email
// @Summary Resend verification email
// @Description Resend email verification link
// @Tags Auth
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Router /auth/resend-verification [post]
func (h *AuthHandler) ResendVerificationEmail(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"user_id"`
	}
	c.BodyParser(&req)

	userID := c.Locals("user_id")
	if userID == nil {
		if req.UserID == "" {
			return utils.Unauthorized(c, "User not authenticated")
		}
		userID = req.UserID
	}

	ipAddress := c.IP()
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = strings.Split(forwarded, ",")[0]
	}

	if err := h.authService.ResendVerificationEmail(c.Context(), userID.(string), ipAddress); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Verification email sent successfully. Please check your inbox.", nil)
}

// GetProfile retrieves user profile
// @Summary Get user profile
// @Description Get authenticated user's profile
// @Tags Auth
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Router /profile [get]
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	profile, err := h.authService.GetProfile(c.Context(), userID.(string))
	if err != nil {
		return utils.NotFound(c, "Profile")
	}
	
	return utils.Success(c, profile)
}

// GetCandidateProfile returns a candidate's profile for employer viewing
// @Summary Get candidate profile
// @Description Get a candidate's full profile (employer access)
// @Tags Candidates
// @Security BearerAuth
// @Param id path string true "Candidate ID"
// @Success 200 {object} APIResponse
// @Router /employer/candidates/{id} [get]
func (h *AuthHandler) GetCandidateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	candidateID := c.Params("id")
	if candidateID == "" {
		return utils.BadRequest(c, "Candidate ID is required")
	}

	profile, err := h.authService.GetCandidateProfile(c.Context(), candidateID)
	if err != nil {
		return utils.NotFound(c, "Candidate")
	}

	return utils.Success(c, profile)
}

// UpdateProfile updates user profile
// @Summary Update user profile
// @Description Update authenticated user's profile
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile update details"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Router /profile [put]
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	serviceReq := &services.UpdateProfileRequest{
		FullName:          req.FullName,
		Headline:          req.Headline,
		Bio:               req.Bio,
		Location:          req.Location,
		IsRemoteOnly:      req.IsRemoteOnly,
		ExperienceLevel:   req.ExperienceLevel,
		YearsOfExperience: req.YearsOfExperience,
		Skills:            req.Skills,
		ResumeURL:         req.ResumeURL,
		PortfolioURL:      req.PortfolioURL,
		LinkedInURL:       req.LinkedInURL,
		IsVisible:         req.IsVisible,
		IsAvailable:       req.IsAvailable,
		CompanyName:       req.CompanyName,
		CompanyWebsite:    req.CompanyWebsite,
		CompanyLinkedIn:   req.CompanyLinkedIn,
		CompanyDescription: req.CompanyDescription,
		Industry:          req.Industry,
		CompanySize:       req.CompanySize,
	}
	
	if err := h.authService.UpdateProfile(c.Context(), userID.(string), serviceReq); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Profile updated successfully", nil)
}

// WebLogout handles GET/POST /logout — clears cookies and redirects
func (h *AuthHandler) WebLogout(c *fiber.Ctx) error {
	if userID := c.Locals("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			if tokenStr := c.Cookies("access_token"); tokenStr != "" {
				h.authService.Logout(c.Context(), uid, tokenStr)
			}
		}
	}
	c.Cookie(&fiber.Cookie{Name: "access_token", Value: "", Path: "/", MaxAge: -1})
	c.Cookie(&fiber.Cookie{Name: "refresh_token", Value: "", Path: "/api/v1/auth/refresh", MaxAge: -1})
	return c.Redirect("/login", fiber.StatusFound)
}

// Helper function to get user ID from context
func (h *AuthHandler) getUserID(c *fiber.Ctx) string {
	if userID := c.Locals("user_id"); userID != nil {
		return userID.(string)
	}
	return ""
}
