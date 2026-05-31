package models

import (
	"time"
)

// UserRole defines the role types
type UserRole string

const (
	RoleEmployee UserRole = "employee"
	RoleEmployer UserRole = "employer"
	RoleAdmin    UserRole = "admin"
)

// User represents the base user account
type User struct {
	ID           string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null;size:255"`
	PasswordHash string    `json:"-" gorm:"not null;size:255"` // "-" hides from JSON
	Role         UserRole  `json:"role" gorm:"type:varchar(50);not null"`
	IsVerified   bool      `json:"is_verified" gorm:"default:false"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships (not stored in DB, just for joins)
	EmployeeProfile *EmployeeProfile `json:"employee_profile,omitempty" gorm:"foreignKey:UserID"`
	EmployerProfile *EmployerProfile `json:"employer_profile,omitempty" gorm:"foreignKey:UserID"`
}

// EmployeeProfile extends User for job seekers
type EmployeeProfile struct {
	UserID            string     `json:"user_id" gorm:"type:uuid;primaryKey"`
	FullName          string     `json:"full_name" gorm:"size:255;not null"`
	Headline          string     `json:"headline" gorm:"size:255"` // e.g., "Senior React Developer"
	Bio               string     `json:"bio" gorm:"type:text"`
	Location          string     `json:"location" gorm:"size:255"`
	IsRemoteOnly      bool       `json:"is_remote_only" gorm:"default:false"`
	ExperienceLevel   string     `json:"experience_level" gorm:"size:50"` // entry, junior, mid, senior, lead
	YearsOfExperience int        `json:"years_of_experience"`
	Skills            []string   `json:"skills" gorm:"type:text[]"` // PostgreSQL array
	ResumeURL         string     `json:"resume_url,omitempty" gorm:"size:512"`
	PortfolioURL      string     `json:"portfolio_url,omitempty" gorm:"size:512"`
	LinkedInURL       string     `json:"linkedin_url,omitempty" gorm:"size:512"`
	IsVisible         bool       `json:"is_visible" gorm:"default:true"`
	IsAvailable       bool       `json:"is_available" gorm:"default:true"`
	
	// GitHub integration
	GithubConnected   bool       `json:"github_connected" gorm:"default:false"`
	GithubUsername    string     `json:"github_username,omitempty" gorm:"size:255;uniqueIndex"`
	LastGithubSync    *time.Time `json:"last_github_sync,omitempty"`
	
	// Statistics
	ProfileViews      int        `json:"profile_views" gorm:"default:0"`
	ApplicationCount  int        `json:"application_count" gorm:"default:0"`
	ResponseRate      float64    `json:"response_rate" gorm:"default:0"`
	
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	User              *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	GithubProfile     *GithubProfile `json:"github_profile,omitempty" gorm:"foreignKey:EmployeeID"`
}

// EmployerProfile extends User for employers
type EmployerProfile struct {
	UserID                string     `json:"user_id" gorm:"type:uuid;primaryKey"`
	CompanyName           string     `json:"company_name" gorm:"size:255;not null;index"`
	CompanyWebsite        string     `json:"company_website,omitempty" gorm:"size:512"`
	CompanyLinkedIn       string     `json:"company_linkedin,omitempty" gorm:"size:512"`
	CompanyLogo           string     `json:"company_logo,omitempty" gorm:"size:512"`
	CompanyDescription    string     `json:"company_description" gorm:"type:text"`
	Industry              string     `json:"industry" gorm:"size:100;index"`
	CompanySize           string     `json:"company_size" gorm:"size:50"` // 1-10, 11-50, 51-200, 201-500, 500+
	Location              string     `json:"location" gorm:"size:255"`
	VerificationStatus    string     `json:"verification_status" gorm:"default:'pending';size:50"` // pending, verified, rejected
	VerifiedAt            *time.Time `json:"verified_at,omitempty"`
	VerifiedBy            string     `json:"verified_by,omitempty" gorm:"size:255"`
	TotalJobsPosted       int        `json:"total_jobs_posted" gorm:"default:0"`
	TotalHires            int        `json:"total_hires" gorm:"default:0"`
	SubscriptionPlan      string     `json:"subscription_plan" gorm:"default:'free';size:50"` // free, basic, pro, enterprise
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	User                  *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Jobs                  []*Job     `json:"jobs,omitempty" gorm:"foreignKey:EmployerID"`
}

// TableName specifies the table names for GORM
func (User) TableName() string {
	return "users"
}

func (EmployeeProfile) TableName() string {
	return "employee_profiles"
}

func (EmployerProfile) TableName() string {
	return "employer_profiles"
}

// Helper methods for EmployeeProfile
func (e *EmployeeProfile) GetFullName() string {
	if e.FullName != "" {
		return e.FullName
	}
	if e.User != nil {
		// Fallback to email username if full name not set
		return e.User.Email
	}
	return "Anonymous"
}

func (e *EmployeeProfile) IsGithubSynced() bool {
	return e.GithubConnected && e.GithubUsername != "" && e.LastGithubSync != nil
}

// Helper methods for EmployerProfile
func (e *EmployerProfile) IsVerified() bool {
	return e.VerificationStatus == "verified"
}

func (e *EmployerProfile) CanPostJob() bool {
	// Check subscription limits here
	if e.SubscriptionPlan == "free" && e.TotalJobsPosted >= 3 {
		return false
	}
	return e.IsVerified() && e.User != nil && e.User.IsActive
}
