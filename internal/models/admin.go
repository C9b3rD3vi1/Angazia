package models

import (
	"time"
)

// AdminActionLog represents admin actions for audit trail
type AdminActionLog struct {
	ID          string                 `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AdminID     string                 `json:"admin_id" gorm:"type:uuid;not null;index"`
	Action      string                 `json:"action" gorm:"size:100;not null"` // create, update, delete, suspend, verify
	EntityType  string                 `json:"entity_type" gorm:"size:50;not null"` // user, job, company, review
	EntityID    string                 `json:"entity_id" gorm:"type:uuid;index"`
	OldValue    JSONMap                `json:"old_value" gorm:"type:jsonb"`
	NewValue    JSONMap                `json:"new_value" gorm:"type:jsonb"`
	IPAddress   string                 `json:"ip_address" gorm:"size:45"`
	UserAgent   string                 `json:"user_agent" gorm:"type:text"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime;index"`
	
	// Relationships
	Admin       *User                  `json:"admin,omitempty" gorm:"foreignKey:AdminID"`
}

// PlatformStats represents overall platform statistics
type PlatformStats struct {
	// User statistics
	TotalUsers          int64          `json:"total_users"`
	TotalCandidates     int64          `json:"total_candidates"`
	TotalEmployers      int64          `json:"total_employers"`
	VerifiedEmployers   int64          `json:"verified_employers"`
	ActiveUsers30Days   int            `json:"active_users_30_days"`
	NewUsers7Days       int64          `json:"new_users_7_days"`
	NewUsers30Days      int64          `json:"new_users_30_days"`
	
	// Job statistics
	TotalJobs           int64          `json:"total_jobs"`
	ActiveJobs          int64          `json:"active_jobs"`
	TotalApplications   int64          `json:"total_applications"`
	JobsPosted7Days     int64          `json:"jobs_posted_7_days"`
	JobsPosted30Days    int64          `json:"jobs_posted_30_days"`
	
	// Engagement metrics
	TotalProfileViews   int            `json:"total_profile_views"`
	TotalJobViews       int            `json:"total_job_views"`
	AverageMatchScore   float64        `json:"average_match_score"`
	
	// Growth metrics
	UserGrowthRate      float64        `json:"user_growth_rate"`
	JobGrowthRate       float64        `json:"job_growth_rate"`
	ApplicationGrowthRate float64      `json:"application_growth_rate"`
	
	// Revenue (future)
	TotalRevenue        float64        `json:"total_revenue"`
	MRR                 float64        `json:"mrr"`
	
	UpdatedAt           time.Time      `json:"updated_at"`
}

// UserReport represents user data for admin reporting
type UserReport struct {
	ID              string    `json:"id"`
	Email           string    `json:"email"`
	Role            string    `json:"role"`
	FullName        string    `json:"full_name"`
	CompanyName     string    `json:"company_name"`
	IsVerified      bool      `json:"is_verified"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	JobCount        int       `json:"job_count,omitempty"`
	ApplicationCount int      `json:"application_count,omitempty"`
	ReportsCount    int       `json:"reports_count"`
}

// ModerationQueue represents items pending moderation
type ModerationQueue struct {
	ID          string                 `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EntityType  string                 `json:"entity_type" gorm:"size:50;not null"` // job, review, company
	EntityID    string                 `json:"entity_id" gorm:"type:uuid;not null;index"`
	Status      string                 `json:"status" gorm:"size:50;default:'pending'"` // pending, approved, rejected
	Reason      string                 `json:"reason" gorm:"type:text"`
	SubmittedBy string                 `json:"submitted_by" gorm:"type:uuid"`
	ReviewedBy  *string                `json:"reviewed_by,omitempty" gorm:"type:uuid"`
	ReviewedAt  *time.Time             `json:"reviewed_at,omitempty"`
	Metadata    JSONMap                `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// SystemSetting represents platform configuration
type SystemSetting struct {
	ID          string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Key         string    `json:"key" gorm:"size:255;uniqueIndex;not null"`
	Value       string    `json:"value" gorm:"type:text"`
	Type        string    `json:"type" gorm:"size:50"` // string, int, bool, json
	Description string    `json:"description" gorm:"type:text"`
	Category    string    `json:"category" gorm:"size:100;index"`
	IsPublic    bool      `json:"is_public" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ReportReason represents reasons for reporting content
type ReportReason struct {
	ID          string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"type:text"`
	EntityType  string    `json:"entity_type" gorm:"size:50"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (AdminActionLog) TableName() string {
	return "admin_action_logs"
}

func (ModerationQueue) TableName() string {
	return "moderation_queue"
}

func (SystemSetting) TableName() string {
	return "system_settings"
}

func (ReportReason) TableName() string {
	return "report_reasons"
}