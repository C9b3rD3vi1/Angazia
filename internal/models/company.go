package models

import (
	"time"
)

// CompanyVerification represents verification documents and status
type CompanyVerification struct {
	ID                         string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID                  string     `json:"company_id" gorm:"type:uuid;not null;uniqueIndex"`
	BusinessRegistrationNumber string     `json:"business_registration_number" gorm:"size:100"`
	TaxID                      string     `json:"tax_id" gorm:"size:100"`
	Documents                  JSONArray  `json:"documents" gorm:"type:jsonb"`
	Status                     string     `json:"status" gorm:"size:50;default:'pending';index"` // pending, approved, rejected
	RejectionReason            string     `json:"rejection_reason,omitempty" gorm:"type:text"`
	VerifiedBy                 *string    `json:"verified_by,omitempty" gorm:"type:uuid"`
	VerifiedAt                 *time.Time `json:"verified_at,omitempty"`
	SubmittedAt                time.Time  `json:"submitted_at" gorm:"autoCreateTime"`
	UpdatedAt                  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	Company                    *EmployerProfile `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	Verifier                   *User            `json:"verifier,omitempty" gorm:"foreignKey:VerifiedBy"`
}

// TrustBadge represents badges awarded to companies
type TrustBadge struct {
	ID          string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID   string     `json:"company_id" gorm:"type:uuid;not null;index"`
	BadgeType   string     `json:"badge_type" gorm:"size:50;not null"` // verified, top_employer, fast_responder, safe_place
	BadgeName   string     `json:"badge_name" gorm:"size:100"`
	Description string     `json:"description" gorm:"type:text"`
	IconURL     string     `json:"icon_url" gorm:"size:512"`
	AwardedAt   time.Time  `json:"awarded_at" gorm:"autoCreateTime"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	
	// Relationships
	Company     *EmployerProfile `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
}

// CompanyReview represents a review of a company by a candidate
type CompanyReview struct {
	ID               string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID        string     `json:"company_id" gorm:"type:uuid;not null;index"`
	ReviewerID       string     `json:"reviewer_id" gorm:"type:uuid;not null;index"`
	Rating           int        `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Title            string     `json:"title" gorm:"size:255"`
	Content          string     `json:"content" gorm:"type:text;not null"`
	Pros             string     `json:"pros" gorm:"type:text"`
	Cons             string     `json:"cons" gorm:"type:text"`
	WouldRecommend   bool       `json:"would_recommend" gorm:"default:false"`
	EmploymentStatus string     `json:"employment_status" gorm:"size:50"` // former, current, interviewed
	IsVerified       bool       `json:"is_verified" gorm:"default:false"`
	HelpfulCount     int        `json:"helpful_count" gorm:"default:0"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	Company          *EmployerProfile `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	Reviewer         *User            `json:"reviewer,omitempty" gorm:"foreignKey:ReviewerID"`
}

// TeamInvitation represents an invitation to join a company team
type TeamInvitation struct {
	ID          string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID   string     `json:"company_id" gorm:"type:uuid;not null;index"`
	Email       string     `json:"email" gorm:"size:255;not null;index"`
	Role        string     `json:"role" gorm:"size:50;not null"` // admin, recruiter, viewer
	Token       string     `json:"token" gorm:"uniqueIndex;not null;size:255"`
	Status      string     `json:"status" gorm:"size:50;default:'pending'"` // pending, accepted, declined, expired
	InvitedBy   string     `json:"invited_by" gorm:"type:uuid;not null"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	ExpiresAt   time.Time  `json:"expires_at" gorm:"not null"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	
	// Relationships
	Company     *EmployerProfile `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	Inviter     *User            `json:"inviter,omitempty" gorm:"foreignKey:InvitedBy"`
}

// CompanyAnalytics tracks daily company statistics
type CompanyAnalytics struct {
	ID                 string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID          string    `json:"company_id" gorm:"type:uuid;not null;index"`
	Date               time.Time `json:"date" gorm:"type:date;not null;index"`
	ProfileViews       int       `json:"profile_views" gorm:"default:0"`
	JobViews           int       `json:"job_views" gorm:"default:0"`
	ApplicationsReceived int     `json:"applications_received" gorm:"default:0"`
	SearchesFound      int       `json:"searches_found" gorm:"default:0"`
	ProfileCompletionRate int    `json:"profile_completion_rate" gorm:"default:0"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	
	// Relationships
	Company            *EmployerProfile `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
}

// ReviewStats represents aggregated review statistics
type ReviewStats struct {
	AverageRating     float64   `json:"average_rating"`
	TotalReviews      int       `json:"total_reviews"`
	RatingDistribution map[int]int `json:"rating_distribution"`
	RecommendationRate float64   `json:"recommendation_rate"`
	WouldRecommend    int       `json:"would_recommend"`
	WouldNotRecommend int       `json:"would_not_recommend"`
}

func (CompanyVerification) TableName() string {
	return "company_verifications"
}

func (TrustBadge) TableName() string {
	return "trust_badges"
}

func (CompanyReview) TableName() string {
	return "company_reviews"
}

func (TeamInvitation) TableName() string {
	return "team_invitations"
}

func (CompanyAnalytics) TableName() string {
	return "company_analytics"
}