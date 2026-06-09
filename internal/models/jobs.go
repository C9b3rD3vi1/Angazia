package models

import (
	"time"
	"github.com/lib/pq"
)

// Job represents a job posting
type Job struct {
	ID               string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployerID       string     `json:"employer_id" gorm:"type:uuid;not null;index"`
	Title            string     `json:"title" gorm:"size:255;not null;index"`
	Description      string     `json:"description" gorm:"type:text;not null"`
	Requirements     string     `json:"requirements" gorm:"type:text"`
	Responsibilities string     `json:"responsibilities" gorm:"type:text"`
	
	// Skills
	RequiredSkills    pq.StringArray `json:"required_skills" gorm:"type:text[]"`
    NiceToHaveSkills  pq.StringArray `json:"nice_to_have_skills" gorm:"type:text[]"`
	
	// Experience & Education
	ExperienceLevel  string     `json:"experience_level" gorm:"size:50"` // entry, junior, mid, senior, lead
	MinExperience    int        `json:"min_experience"` // years
	MaxExperience    int        `json:"max_experience"` // years
	EducationLevel   string     `json:"education_level" gorm:"size:100"` // high school, diploma, bachelor, master, phd
	
	// Compensation
	SalaryMin        int        `json:"salary_min"`
	SalaryMax        int        `json:"salary_max"`
	SalaryCurrency   string     `json:"salary_currency" gorm:"default:'KES';size:3"`
	IsSalaryVisible  bool       `json:"is_salary_visible" gorm:"default:true"`
	Benefits         []string   `json:"benefits_list" gorm:"type:text[]"` // health, lunch, transport, etc.
	
	// Location
	Location         string     `json:"location" gorm:"size:255;index"`
	IsRemote         bool       `json:"is_remote" gorm:"default:false"`
	IsHybrid         bool       `json:"is_hybrid" gorm:"default:false"`
	RemotePolicy     string     `json:"remote_policy" gorm:"size:255"` // "fully remote", "remote within Kenya", etc.
	
	// Employment Details
	EmploymentType   string     `json:"employment_type" gorm:"default:'full-time';size:50"` // full-time, part-time, contract, internship, freelance
	WorkHours        string     `json:"work_hours" gorm:"size:100"` // "9-5", "flexible", "shift"
	
	// Status & Dates
	IsActive         bool       `json:"is_active" gorm:"default:true;index"`
	IsFeatured       bool       `json:"is_featured" gorm:"default:false"`
	IsUrgent         bool       `json:"is_urgent" gorm:"default:false"`
	PostedAt         time.Time  `json:"posted_at" gorm:"autoCreateTime;index"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime;index"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty" gorm:"index"`
	ClosedAt         *time.Time `json:"closed_at,omitempty"`
	
	// Statistics
	ViewsCount       int        `json:"views_count" gorm:"default:0"`
	ApplicationsCount int       `json:"applications_count" gorm:"default:0"`
	ShortlistedCount int        `json:"shortlisted_count" gorm:"default:0"`
	HiredCount       int        `json:"hired_count" gorm:"default:0"`
	
	// Relationships
	Employer         *EmployerProfile `json:"employer,omitempty" gorm:"foreignKey:EmployerID"`
	Applications     []*Application   `json:"applications,omitempty" gorm:"foreignKey:JobID"`
}

// Application represents a job application from an employee
type Application struct {
	ID            string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	JobID         string     `json:"job_id" gorm:"type:uuid;not null;index"`
	EmployeeID    string     `json:"employee_id" gorm:"type:uuid;not null;index"`
	
	// Application details
	CoverLetter   string     `json:"cover_letter" gorm:"type:text"`
	ResumeURL     string     `json:"resume_url,omitempty" gorm:"size:512"`
	PortfolioURL  string     `json:"portfolio_url,omitempty" gorm:"size:512"`
	
	// Matching & Scoring
	MatchScore    int        `json:"match_score" gorm:"default:0"` // 1-100
	SkillMatch    int        `json:"skill_match" gorm:"default:0"` // Skills match percentage
	ExperienceMatch int      `json:"experience_match" gorm:"default:0"` // Experience match percentage
	AiInsights    string     `json:"ai_insights,omitempty" gorm:"type:text"` // AI-generated feedback
	
	// Status tracking
	Status        string     `json:"status" gorm:"default:'pending';size:50;index"` // pending, viewed, shortlisted, rejected, hired, withdrawn
	StatusHistory JSONArray   `json:"status_history" gorm:"type:jsonb"` // Array of status changes
	EmployerNotes string     `json:"employer_notes,omitempty" gorm:"type:text"`
	EmployerRating int       `json:"employer_rating"` // 1-5 stars
	
	// Interview tracking
	InterviewDate *time.Time `json:"interview_date,omitempty"`
	InterviewType string     `json:"interview_type,omitempty" gorm:"size:50"` // phone, technical, onsite, final
	InterviewNotes string    `json:"interview_notes,omitempty" gorm:"type:text"`
	
	// Dates
	AppliedAt     time.Time  `json:"applied_at" gorm:"autoCreateTime;index"`
	ViewedAt      *time.Time `json:"viewed_at,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	RespondedAt   *time.Time `json:"responded_at,omitempty"`
	
	// Relationships
	Job           *Job           `json:"job,omitempty" gorm:"foreignKey:JobID"`
	Employee      *EmployeeProfile `json:"employee,omitempty" gorm:"foreignKey:EmployeeID"`
}

// SavedJob allows employees to save jobs for later
type SavedJob struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID string    `json:"employee_id" gorm:"type:uuid;not null;index"`
	JobID      string    `json:"job_id" gorm:"type:uuid;not null;index"`
	Notes      string    `json:"notes,omitempty" gorm:"type:text"`
	SavedAt    time.Time `json:"saved_at" gorm:"autoCreateTime"`
	
	// Relationships
	Employee   *EmployeeProfile `json:"employee,omitempty" gorm:"foreignKey:EmployeeID"`
	Job        *Job             `json:"job,omitempty" gorm:"foreignKey:JobID"`
}

// JobView tracks who viewed which jobs
type JobView struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	JobID      string    `json:"job_id" gorm:"type:uuid;not null;index"`
	EmployeeID *string   `json:"employee_id,omitempty" gorm:"type:uuid;index"` // null if anonymous
	IPAddress  string    `json:"ip_address" gorm:"size:45"`
	UserAgent  string    `json:"user_agent" gorm:"type:text"`
	ViewedAt   time.Time `json:"viewed_at" gorm:"autoCreateTime;index"`
}

// TableName specifies table names
func (Job) TableName() string {
	return "jobs"
}

func (Application) TableName() string {
	return "applications"
}

func (SavedJob) TableName() string {
	return "saved_jobs"
}

func (JobView) TableName() string {
	return "job_views"
}

// Helper methods for Job
func (j *Job) IsExpired() bool {
	if j.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*j.ExpiresAt)
}

func (j *Job) GetSalaryRange() string {
	if j.SalaryMin == 0 && j.SalaryMax == 0 {
		return "Not specified"
	}
	if j.SalaryMin == j.SalaryMax {
		return formatMoney(j.SalaryMin) + " " + j.SalaryCurrency
	}
	return formatMoney(j.SalaryMin) + " - " + formatMoney(j.SalaryMax) + " " + j.SalaryCurrency
}

func (j *Job) GetLocationDisplay() string {
	if j.IsRemote {
		if j.IsHybrid {
			return "Hybrid / " + j.Location
		}
		return "Remote"
	}
	return j.Location
}

func (j *Job) GetRequiredSkillsCount() int {
	return len(j.RequiredSkills)
}

// Helper methods for Application
func (a *Application) IsPending() bool {
	return a.Status == "pending"
}

func (a *Application) IsShortlisted() bool {
	return a.Status == "shortlisted"
}

func (a *Application) IsRejected() bool {
	return a.Status == "rejected"
}

func (a *Application) IsHired() bool {
	return a.Status == "hired"
}

func (a *Application) GetMatchScorePercent() int {
	return a.MatchScore
}

func (a *Application) GetMatchGrade() string {
	if a.MatchScore >= 90 {
		return "Excellent"
	} else if a.MatchScore >= 70 {
		return "Good"
	} else if a.MatchScore >= 50 {
		return "Fair"
	} else {
		return "Poor"
	}
}

// Helper functions
func formatMoney(amount int) string {
	if amount >= 1000000 {
		return formatNumber(amount/1000000) + "M"
	} else if amount >= 1000 {
		return formatNumber(amount/1000) + "K"
	}
	return formatNumber(amount)
}

func formatNumber(n int) string {
	if n < 1000 {
		return string(rune(n))
	}
	// Simple formatting - can be improved
	return string(rune(n))
}

// Validation methods
func (j *Job) Validate() map[string]string {
	errors := make(map[string]string)
	
	if j.Title == "" {
		errors["title"] = "Job title is required"
	}
	
	if j.Description == "" {
		errors["description"] = "Job description is required"
	}
	
	if len(j.RequiredSkills) == 0 {
		errors["required_skills"] = "At least one required skill is needed"
	}
	
	if j.SalaryMin > j.SalaryMax && j.SalaryMax != 0 {
		errors["salary"] = "Minimum salary cannot be greater than maximum salary"
	}
	
	if j.ExpiresAt != nil && j.ExpiresAt.Before(time.Now()) {
		errors["expires_at"] = "Expiration date cannot be in the past"
	}
	
	return errors
}
