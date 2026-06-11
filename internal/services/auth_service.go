package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	
	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

// Update the AuthService interface to include all methods
type AuthService interface {
	// Registration & Login
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, email, password, ipAddress, userAgent string) (*AuthResponse, error)
	Logout(ctx context.Context, userID, token string) error
	
	// Token management
	RefreshToken(ctx context.Context, refreshToken string, ipAddress string) (*AuthResponse, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
	
	// Password management
	ForgotPassword(ctx context.Context, email string, ipAddress string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	
	// Email verification
	VerifyEmail(ctx context.Context, token string) error
	ResendVerificationEmail(ctx context.Context, userID string, ipAddress string) error
	
	// Profile management
	GetProfile(ctx context.Context, userID string) (*ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) error
	GetCandidateProfile(ctx context.Context, candidateID string) (*ProfileResponse, error)

	// Session management
	GetSessions(ctx context.Context, userID string) ([]*models.UserSession, error)
	RevokeSession(ctx context.Context, userID, sessionID string) error

	// Account management
	DeleteAccount(ctx context.Context, userID, password string) error

	// Admin
	GetOrCreateAdmin(ctx context.Context, email string) (*models.User, error)
	GenerateTokens(ctx context.Context, userID string, role models.UserRole, email, ipAddress, userAgent string) (*AuthResponse, error)
}



type RegisterRequest struct {
	Email     string          `json:"email" validate:"required,email"`
	Password  string          `json:"password" validate:"required,min=8"`
	FullName  string          `json:"full_name"`
	Role      models.UserRole `json:"role" validate:"oneof=employee employer"`
	IPAddress string          `json:"-"`
}

type AuthResponse struct {
	User          *UserResponse `json:"user"`
	AccessToken   string        `json:"access_token"`
	RefreshToken  string        `json:"refresh_token"`
	ExpiresAt     int64         `json:"expires_at"`
	RequiresTwoFA bool          `json:"requires_2fa,omitempty"`
}

type UserResponse struct {
	ID          string          `json:"id"`
	Email       string          `json:"email"`
	Role        models.UserRole `json:"role"`
	IsVerified  bool            `json:"is_verified"`
	FullName    string          `json:"full_name,omitempty"`
	CompanyName string          `json:"company_name,omitempty"`
	AvatarURL   string          `json:"avatar_url,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type TokenClaims struct {
	UserID string          `json:"user_id"`
	Email  string          `json:"email"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type ProfileResponse struct {
	User            *UserResponse               `json:"user"`
	EmployeeProfile *models.EmployeeProfile     `json:"employee_profile,omitempty"`
	EmployerProfile *models.EmployerProfile     `json:"employer_profile,omitempty"`
}

type UpdateProfileRequest struct {
	FullName          string                      `json:"full_name"`
	Headline          string                      `json:"headline"`
	Bio               string                      `json:"bio"`
	Location          string                      `json:"location"`
	IsRemoteOnly      bool                        `json:"is_remote_only"`
	ExperienceLevel   string                      `json:"experience_level"`
	YearsOfExperience int                         `json:"years_of_experience"`
	Skills            []string                    `json:"skills"`
	Experience        []models.WorkExperienceItem `json:"experience"`
	Certifications    []models.CertificationItem  `json:"certifications"`
	ResumeURL         string                      `json:"resume_url"`
	PortfolioURL      string                      `json:"portfolio_url"`
	LinkedInURL       string                      `json:"linkedin_url"`
	IsVisible         bool                        `json:"is_visible"`
	IsAvailable       bool                        `json:"is_available"`
	
	// Employer specific
	CompanyName        string `json:"company_name"`
	CompanyWebsite     string `json:"company_website"`
	CompanyLinkedIn    string `json:"company_linkedin"`
	CompanyDescription string `json:"company_description"`
	Industry           string `json:"industry"`
	CompanySize        string `json:"company_size"`
}

// StoredToken represents a temporary token in database
type StoredToken struct {
	ID        string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string     `gorm:"type:uuid;not null;index"`
	Token     string     `gorm:"uniqueIndex;not null;size:255"`
	Type      string     `gorm:"size:50;not null;index"`
	ExpiresAt time.Time  `gorm:"not null;index"`
	UsedAt    *time.Time `gorm:"index"`
	IPAddress string     `gorm:"size:45"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
}

func (StoredToken) TableName() string {
	return "auth_tokens"
}

// BlacklistedToken for JWT blacklist
type BlacklistedToken struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Token     string    `gorm:"uniqueIndex;not null;size:512"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (BlacklistedToken) TableName() string {
	return "blacklisted_tokens"
}

type AuthServiceImpl struct {
	cfg         *config.Config
	userRepo    repository.UserRepository
	db          *gorm.DB
	emailSvc    EmailService
	redisClient *redis.Client
}

func NewAuthService(cfg *config.Config, userRepo repository.UserRepository, db *gorm.DB, emailSvc EmailService, redisClient *redis.Client) AuthService {
	// Auto-migrate token tables
	db.AutoMigrate(&StoredToken{}, &BlacklistedToken{})
	
	return &AuthServiceImpl{
		cfg:         cfg,
		userRepo:    userRepo,
		db:          db,
		emailSvc:    emailSvc,
		redisClient: redisClient,
	}
}

func (s *AuthServiceImpl) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// Validate input
	if err := s.validateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := s.validatePassword(req.Password); err != nil {
		return nil, err
	}
	
	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, errors.New("user already exists with this email")
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
		IsVerified:   false,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	if err := tx.WithContext(ctx).Create(user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	// Create role-specific profile
	if req.Role == models.RoleEmployee {
		displayName := req.FullName
		if displayName == "" {
			displayName = req.Email
		}
		employeeProfile := &models.EmployeeProfile{
			UserID:      user.ID,
			FullName:    displayName,
			Headline:    "Software Developer",
			IsVisible:   true,
			IsAvailable: true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := tx.WithContext(ctx).Create(employeeProfile).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create employee profile: %w", err)
		}
	} else if req.Role == models.RoleEmployer {
		companyName := req.FullName
		if companyName == "" {
			companyName = req.Email
		}
		employerProfile := &models.EmployerProfile{
			UserID:             user.ID,
			CompanyName:        companyName,
			VerificationStatus: "pending",
			SubscriptionPlan:   "free",
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}
		if err := tx.WithContext(ctx).Create(employerProfile).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create employer profile: %w", err)
		}

		// Auto-create a trial subscription
		var plan models.SubscriptionPlan
		if err := tx.WithContext(ctx).Where("plan_id = ? AND is_active = ?", "pro_monthly", true).First(&plan).Error; err != nil {
			if err := tx.WithContext(ctx).Where("plan_id = ?", "free").First(&plan).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to find subscription plan: %w", err)
			}
		}
		now := time.Now()
		trialEnd := now.AddDate(0, 0, plan.TrialDays)
		sub := &models.Subscription{
			ID:                uuid.New().String(),
			UserID:            user.ID,
			PlanID:            plan.PlanID,
			PlanName:          plan.Name,
			Amount:            plan.Price,
			Currency:          plan.Currency,
			Interval:          plan.Interval,
			Status:            "trialing",
			StartDate:         now,
			EndDate:           trialEnd,
			AutoRenew:         false,
			JobPostLimit:      plan.JobPostLimit,
			Features:          plan.Features,
			FeatureFlags:      plan.FeatureFlags,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:  trialEnd,
			TrialEndsAt:       &trialEnd,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := tx.WithContext(ctx).Create(sub).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create trial subscription: %w", err)
		}
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	// Generate email verification token
	verificationToken, err := s.createToken(user.ID, "email_verification", 24*time.Hour, req.IPAddress)
	if err != nil {
		// Log error but don't fail registration
		fmt.Printf("Failed to create verification token: %v\n", err)
	} else {
		// Send verification email (async)
		go s.emailSvc.SendVerificationEmail(user.Email, verificationToken, user.ID, user.Email)
	}
	
	// Generate auth response
	return s.generateAuthResponse(ctx, user)
}

func (s *AuthServiceImpl) Login(ctx context.Context, email, password, ipAddress, userAgent string) (*AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("invalid email or password")
	}
	
	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated. Please contact support")
	}
	
	// Check if email is verified
	if s.cfg.RequireEmailVerification && !user.IsVerified {
		return nil, errors.New("email not verified. Please check your inbox for verification link")
	}
	
	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// Log failed login attempt
		s.logFailedLoginAttempt(ctx, user.ID, ipAddress, userAgent)
		return nil, errors.New("invalid email or password")
	}
	
	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		fmt.Printf("Failed to update last login: %v\n", err)
	}
	
	// Log successful login
	s.logSuccessfulLogin(ctx, user.ID, ipAddress, userAgent)
	
	// Generate auth response
	resp, err := s.generateAuthResponse(ctx, user)
	if err == nil {
		s.createSession(ctx, user.ID, resp.AccessToken, userAgent, ipAddress)
	}
	return resp, err
}

func (s *AuthServiceImpl) Logout(ctx context.Context, userID, token string) error {
	// Parse token to get expiration
	claims, err := s.ValidateToken(token)
	if err != nil {
		// If token is already invalid, still consider logout successful
		return nil
	}
	
	// Add token to blacklist
	blacklistedToken := &BlacklistedToken{
		ID:        uuid.New().String(),
		Token:     token,
		UserID:    userID,
		ExpiresAt: claims.ExpiresAt.Time,
		CreatedAt: time.Now(),
	}
	
	if err := s.db.WithContext(ctx).Create(blacklistedToken).Error; err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	
	// Also store in Redis for faster lookups
	if s.redisClient != nil {
		ttl := time.Until(claims.ExpiresAt.Time)
		if ttl > 0 {
			s.redisClient.Set(ctx, "blacklist:"+token, userID, ttl)
		}
	}
	
	return nil
}

func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken string, ipAddress string) (*AuthResponse, error) {
	// Check if token is blacklisted
	var blacklisted BlacklistedToken
	if err := s.db.WithContext(ctx).Where("token = ?", refreshToken).First(&blacklisted).Error; err == nil {
		return nil, errors.New("token has been revoked")
	}
	
	// Validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}
	
	// Get user
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}
	
	// Blacklist the old refresh token
	go func() {
		blacklistedToken := &BlacklistedToken{
			ID:        uuid.New().String(),
			Token:     refreshToken,
			UserID:    user.ID,
			ExpiresAt: claims.ExpiresAt.Time,
			CreatedAt: time.Now(),
		}
		s.db.Create(blacklistedToken)
	}()
	
	// Generate new tokens
	return s.generateAuthResponse(ctx, user)
}

func (s *AuthServiceImpl) ValidateToken(tokenString string) (*TokenClaims, error) {
	// Check if token is blacklisted in Redis
	if s.redisClient != nil {
		exists, err := s.redisClient.Exists(context.Background(), "blacklist:"+tokenString).Result()
		if err == nil && exists == 1 {
			return nil, errors.New("token has been revoked")
		}
	}
	
	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, errors.New("invalid token")
}

func (s *AuthServiceImpl) ForgotPassword(ctx context.Context, email, ipAddress string) error {
	// Get user (don't reveal if doesn't exist for security)
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		// Return success even if user doesn't exist to prevent email enumeration
		return nil
	}
	
	// Check for existing recent tokens (prevent spam)
	var existingToken StoredToken
	err = s.db.WithContext(ctx).
		Where("user_id = ? AND type = ? AND used_at IS NULL AND created_at > ?", user.ID, "password_reset", time.Now().Add(-5*time.Minute)).
		First(&existingToken).Error
	
	if err == nil {
		return errors.New("password reset email already sent. Please wait 5 minutes")
	}
	
	// Generate password reset token (expires in 1 hour)
	resetToken, err := s.createToken(user.ID, "password_reset", 1*time.Hour, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}
	
	// Send reset email (async)
	go s.emailSvc.SendPasswordResetEmail(user.Email, resetToken, user.ID, user.Email)
	
	return nil
}

func (s *AuthServiceImpl) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}
	
	// Get token from database
	var storedToken StoredToken
	err := s.db.WithContext(ctx).
		Where("token = ? AND type = ? AND expires_at > ? AND used_at IS NULL", token, "password_reset", time.Now()).
		First(&storedToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid or expired reset token")
		}
		return fmt.Errorf("failed to validate token: %w", err)
	}
	
	// Mark token as used
	now := time.Now()
	storedToken.UsedAt = &now
	if err := s.db.WithContext(ctx).Save(&storedToken).Error; err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}
	
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Update user password
	if err := s.userRepo.UpdatePassword(ctx, storedToken.UserID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	
	// Send confirmation email
	user, _ := s.userRepo.GetByID(ctx, storedToken.UserID)
	if user != nil {
		go s.emailSvc.SendPasswordChangedEmail(user.Email, user.Email)
	}
	
	return nil
}

func (s *AuthServiceImpl) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}
	
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}
	
	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid current password")
	}
	
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	
	// Blacklist all active tokens for this user
	go s.blacklistAllUserTokens(ctx, userID)
	
	// Send confirmation email
	go s.emailSvc.SendPasswordChangedEmail(user.Email, user.Email)
	
	return nil
}

func (s *AuthServiceImpl) VerifyEmail(ctx context.Context, token string) error {
	// Get token from database
	var storedToken StoredToken
	err := s.db.WithContext(ctx).
		Where("token = ? AND type = ? AND expires_at > ? AND used_at IS NULL", token, "email_verification", time.Now()).
		First(&storedToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid or expired verification token")
		}
		return fmt.Errorf("failed to validate token: %w", err)
	}
	
	// Mark token as used
	now := time.Now()
	storedToken.UsedAt = &now
	if err := s.db.WithContext(ctx).Save(&storedToken).Error; err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}
	
	// Verify user email
	if err := s.userRepo.VerifyEmail(ctx, storedToken.UserID); err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}
	
	// Send welcome email
	user, _ := s.userRepo.GetByID(ctx, storedToken.UserID)
	if user != nil {
		// Get profile for name
		var fullName string
		if user.Role == models.RoleEmployee {
			profile, _ := s.userRepo.GetEmployeeProfile(ctx, user.ID)
			if profile != nil {
				fullName = profile.FullName
			}
		}
		go s.emailSvc.SendWelcomeEmail(user.Email, fullName, user.Email)
	}
	
	return nil
}

func (s *AuthServiceImpl) ResendVerificationEmail(ctx context.Context, userID, ipAddress string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}
	
	if user.IsVerified {
		return errors.New("email already verified")
	}
	
	// Check for recent token (prevent spam)
	var existingToken StoredToken
	err = s.db.WithContext(ctx).
		Where("user_id = ? AND type = ? AND used_at IS NULL AND created_at > ?", userID, "email_verification", time.Now().Add(-5*time.Minute)).
		First(&existingToken).Error
	
	if err == nil {
		return errors.New("verification email already sent recently. Please wait 5 minutes")
	}
	
	// Generate new verification token
	verificationToken, err := s.createToken(user.ID, "email_verification", 24*time.Hour, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}
	
	// Send verification email (async)
	go s.emailSvc.SendVerificationEmail(user.Email, verificationToken, user.ID, user.Email)
	
	return nil
}

func (s *AuthServiceImpl) GetProfile(ctx context.Context, userID string) (*ProfileResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	
	response := &ProfileResponse{
		User: &UserResponse{
			ID:         user.ID,
			Email:      user.Email,
			Role:       user.Role,
			IsVerified: user.IsVerified,
			AvatarURL:  user.AvatarURL,
			CreatedAt:  user.CreatedAt,
		},
	}
	
	// Get role-specific profile
	if user.Role == models.RoleEmployee {
		profile, err := s.userRepo.GetEmployeeProfile(ctx, userID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to get employee profile: %w", err)
		}
		if profile != nil {
			computeProfileStrength(profile)
			response.EmployeeProfile = profile
			response.User.FullName = profile.FullName
		}
	} else if user.Role == models.RoleEmployer {
		profile, err := s.userRepo.GetEmployerProfile(ctx, userID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to get employer profile: %w", err)
		}
		if profile != nil {
			response.EmployerProfile = profile
			response.User.CompanyName = profile.CompanyName
		}
	}
	
	return response, nil
}

func computeProfileStrength(p *models.EmployeeProfile) {
	score := 0

	// Full name (10)
	if p.FullName != "" && len(p.FullName) > 2 {
		score += 10
	}

	// Headline (10)
	if p.Headline != "" {
		score += 10
	}

	// Skills (max 20)
	if len(p.Skills) >= 3 {
		score += 15
		if len(p.Skills) >= 5 {
			score += 5
		}
	}

	// Experience (15)
	if p.YearsOfExperience > 0 || len(p.Experience) > 0 {
		score += 15
	}

	// Location (10)
	if p.Location != "" {
		score += 10
	}

	// GitHub (15)
	if p.GithubConnected {
		score += 15
	}

	// Portfolio or LinkedIn (10)
	if p.PortfolioURL != "" || p.LinkedInURL != "" {
		score += 10
	}

	// Resume (10)
	if p.ResumeURL != "" {
		score += 10
	}

	p.ProfileStrength = score

	// Individual match percentages (simplified — based on collected data)
	if len(p.Skills) > 0 {
		p.SkillsMatchPercent = len(p.Skills) * 20
		if p.SkillsMatchPercent > 100 {
			p.SkillsMatchPercent = 100
		}
	}
	if p.YearsOfExperience > 0 || len(p.Experience) > 0 {
		p.ExperienceMatchPercent = 80
		if p.YearsOfExperience >= 3 {
			p.ExperienceMatchPercent = 100
		}
	}
	if p.Location != "" {
		p.LocationMatchPercent = 100
	}
}

func (s *AuthServiceImpl) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}
	
	// Update role-specific profile
	if user.Role == models.RoleEmployee {
		profile := models.EmployeeProfile{
			FullName:          req.FullName,
			Headline:          req.Headline,
			Bio:               req.Bio,
			Location:          req.Location,
			IsRemoteOnly:      req.IsRemoteOnly,
			ExperienceLevel:   req.ExperienceLevel,
			YearsOfExperience: req.YearsOfExperience,
			Skills:            req.Skills,
			Experience:        models.WorkExperiences(req.Experience),
			Certifications:    models.Certifications(req.Certifications),
			ResumeURL:         req.ResumeURL,
			PortfolioURL:      req.PortfolioURL,
			LinkedInURL:       req.LinkedInURL,
			IsVisible:         req.IsVisible,
			IsAvailable:       req.IsAvailable,
		}
		return s.db.WithContext(ctx).
			Model(&models.EmployeeProfile{}).
			Where("user_id = ?", userID).
			Updates(&profile).Error
	} else if user.Role == models.RoleEmployer {
		updates := map[string]interface{}{
			"company_name":        req.CompanyName,
			"company_website":     req.CompanyWebsite,
			"company_linkedin":    req.CompanyLinkedIn,
			"company_description": req.CompanyDescription,
			"industry":            req.Industry,
			"company_size":        req.CompanySize,
			"updated_at":          time.Now(),
		}
		return s.userRepo.UpdateEmployerProfile(ctx, userID, updates)
	}
	
	return errors.New("invalid user role")
}

func (s *AuthServiceImpl) GetOrCreateAdmin(ctx context.Context, email string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(s.cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash admin password: %w", err)
	}

	adminUser := &models.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         models.RoleAdmin,
		IsVerified:   true,
		IsActive:     true,
	}
	if err := s.userRepo.Create(ctx, adminUser); err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}
	return adminUser, nil
}

func (s *AuthServiceImpl) GetCandidateProfile(ctx context.Context, candidateID string) (*ProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, candidateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("candidate not found")
	}

	response := &ProfileResponse{
		User: &UserResponse{
			ID:         user.ID,
			Email:      user.Email,
			Role:       user.Role,
			IsVerified: user.IsVerified,
			CreatedAt:  user.CreatedAt,
		},
	}

	profile, githubProfile, err := s.userRepo.GetEmployeeWithGitHub(ctx, candidateID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to get candidate profile: %w", err)
	}
	if profile != nil {
		response.EmployeeProfile = profile
		response.User.FullName = profile.FullName
	}

	if githubProfile != nil {
		if response.EmployeeProfile != nil {
			response.EmployeeProfile.GithubProfile = githubProfile
		}
	}

	return response, nil
}

func (s *AuthServiceImpl) GenerateTokens(ctx context.Context, userID string, role models.UserRole, email, ipAddress, userAgent string) (*AuthResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	s.logSuccessfulLogin(ctx, user.ID, ipAddress, userAgent)

	return s.generateAuthResponse(ctx, user)
}

// Private helper methods

func (s *AuthServiceImpl) generateAuthResponse(ctx context.Context, user *models.User) (*AuthResponse, error) {
	// Get profile for additional info
	var fullName, companyName string
	
	if user.Role == models.RoleEmployee {
		profile, _ := s.userRepo.GetEmployeeProfile(ctx, user.ID)
		if profile != nil {
			fullName = profile.FullName
		}
	} else if user.Role == models.RoleEmployer {
		profile, _ := s.userRepo.GetEmployerProfile(ctx, user.ID)
		if profile != nil {
			companyName = profile.CompanyName
		}
	}
	
	// Generate access token (short-lived)
	accessExpiration := time.Duration(s.cfg.JWTExpiryHours) * time.Hour
	accessToken, err := s.generateToken(user, accessExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	
	// Generate refresh token (long-lived, 7 days)
	refreshExpiration := 7 * 24 * time.Hour
	refreshToken, err := s.generateToken(user, refreshExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	
	return &AuthResponse{
		User: &UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			Role:        user.Role,
			IsVerified:  user.IsVerified,
			FullName:    fullName,
			CompanyName: companyName,
			CreatedAt:   user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(accessExpiration).Unix(),
	}, nil
}

// ── Session management ──

func (s *AuthServiceImpl) createSession(ctx context.Context, userID, token, userAgent, ipAddress string) {
	device, browser, os := parseUserAgent(userAgent)
	session := &models.UserSession{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		Device:    device,
		Browser:   browser,
		OS:        os,
		IPAddress: ipAddress,
		LastActive: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	s.db.WithContext(ctx).Create(session)
}

func (s *AuthServiceImpl) GetSessions(ctx context.Context, userID string) ([]*models.UserSession, error) {
	var sessions []*models.UserSession
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *AuthServiceImpl) RevokeSession(ctx context.Context, userID, sessionID string) error {
	result := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Delete(&models.UserSession{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("session not found")
	}
	return nil
}

// ── Account management ──

func (s *AuthServiceImpl) DeleteAccount(ctx context.Context, userID, password string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return errors.New("invalid password")
	}

	// Delete all user data in a transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete sessions
	tx.Where("user_id = ?", userID).Delete(&models.UserSession{})
	// Delete notifications
	tx.Where("user_id = ?", userID).Delete(&models.Notification{})
	// Delete notification preferences
	tx.Where("user_id = ?", userID).Delete(&models.NotificationPreferences{})
	// Delete employee profile
	tx.Where("user_id = ?", userID).Delete(&models.EmployeeProfile{})
	// Delete employer profile
	tx.Where("user_id = ?", userID).Delete(&models.EmployerProfile{})
	// Delete auth tokens
	tx.Where("user_id = ?", userID).Delete(&StoredToken{})
	tx.Where("user_id = ?", userID).Delete(&BlacklistedToken{})
	// Delete the user
	tx.Where("id = ?", userID).Delete(&models.User{})

	return tx.Commit().Error
}

func (s *AuthServiceImpl) generateToken(user *models.User, expiration time.Duration) (string, error) {
	claims := &TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.cfg.AppName,
			Subject:   user.ID,
			ID:        uuid.New().String(),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthServiceImpl) createToken(userID, tokenType string, expiration time.Duration, ipAddress string) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)
	
	// Store in database
	storedToken := &StoredToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		Type:      tokenType,
		ExpiresAt: time.Now().Add(expiration),
		IPAddress: ipAddress,
		CreatedAt: time.Now(),
	}
	
	if err := s.db.Create(storedToken).Error; err != nil {
		return "", err
	}
	
	return token, nil
}

func (s *AuthServiceImpl) blacklistAllUserTokens(ctx context.Context, userID string) {
	// In production, you would blacklist all active tokens for this user
	// This requires storing token IDs in Redis or a separate table
}

func (s *AuthServiceImpl) logFailedLoginAttempt(ctx context.Context, userID, ipAddress, userAgent string) {
	// In production, store failed login attempts for security monitoring
	// Could use Redis to track and implement rate limiting
}

func (s *AuthServiceImpl) logSuccessfulLogin(ctx context.Context, userID, ipAddress, userAgent string) {
	// In production, log successful logins for audit trail
	// Store in database or send to analytics service
}

func (s *AuthServiceImpl) validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	// Add additional email validation if needed
	return nil
}

func (s *AuthServiceImpl) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > 72 {
		return errors.New("password must be less than 72 characters")
	}
	return nil
}

func parseUserAgent(userAgent string) (device, browser, os string) {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "iphone"):
		device = "iPhone"
	case strings.Contains(ua, "ipad"):
		device = "iPad"
	case strings.Contains(ua, "android"):
		device = "Android"
	case strings.Contains(ua, "linux"):
		device = "Linux Desktop"
	case strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os"):
		device = "Mac"
	case strings.Contains(ua, "windows"):
		device = "Windows"
	default:
		device = "Unknown"
	}

	switch {
	case strings.Contains(ua, "edg/"):
		browser = "Edge"
	case strings.Contains(ua, "chrome/") && !strings.Contains(ua, "edg/"):
		browser = "Chrome"
	case strings.Contains(ua, "safari/") && !strings.Contains(ua, "chrome/"):
		browser = "Safari"
	case strings.Contains(ua, "firefox/"):
		browser = "Firefox"
	case strings.Contains(ua, "opera") || strings.Contains(ua, "opr/"):
		browser = "Opera"
	default:
		browser = "Unknown"
	}

	switch {
	case strings.Contains(ua, "windows nt"):
		os = "Windows"
	case strings.Contains(ua, "mac os x") || strings.Contains(ua, "macintosh"):
		os = "macOS"
	case strings.Contains(ua, "linux") && !strings.Contains(ua, "android"):
		os = "Linux"
	case strings.Contains(ua, "android"):
		os = "Android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		os = "iOS"
	default:
		os = "Unknown"
	}
	return
}