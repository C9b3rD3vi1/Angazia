package models

import (
	"time"
)

// TalentPool represents a collection of candidates saved by an employer
type TalentPool struct {
	ID          string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployerID  string    `json:"employer_id" gorm:"type:uuid;not null;index"`
	Name        string    `json:"name" gorm:"size:255;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Color       string    `json:"color" gorm:"size:50;default:'#667eea'"`
	Icon        string    `json:"icon" gorm:"size:50;default:'users'"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CandidateCount int     `json:"candidate_count" gorm:"-"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	Employer    *EmployerProfile      `json:"employer,omitempty" gorm:"foreignKey:EmployerID"`
	Candidates  []*TalentPoolCandidate `json:"candidates,omitempty" gorm:"foreignKey:TalentPoolID"`
}

// TalentPoolCandidate represents a candidate saved to a talent pool
type TalentPoolCandidate struct {
	ID            string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TalentPoolID  string     `json:"talent_pool_id" gorm:"type:uuid;not null;index"`
	EmployeeID    string     `json:"employee_id" gorm:"type:uuid;not null;index"`
	MatchScore    int        `json:"match_score" gorm:"default:0"`
	Notes         string     `json:"notes" gorm:"type:text"`
	Tags          []string   `json:"tags" gorm:"type:text[]"`
	Status        string     `json:"status" gorm:"size:50;default:'active'"` // active, contacted, hired, archived
	ContactedAt   *time.Time `json:"contacted_at,omitempty"`
	HiredAt       *time.Time `json:"hired_at,omitempty"`
	AddedAt       time.Time  `json:"added_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	TalentPool    *TalentPool        `json:"talent_pool,omitempty" gorm:"foreignKey:TalentPoolID"`
	Employee      *EmployeeProfile   `json:"employee,omitempty" gorm:"foreignKey:EmployeeID"`
}

// TalentPoolCandidateWithDetails extends TalentPoolCandidate with employee details
type TalentPoolCandidateWithDetails struct {
	TalentPoolCandidate
	EmployeeName     string   `json:"employee_name"`
	EmployeeEmail    string   `json:"employee_email"`
	EmployeeHeadline string   `json:"employee_headline"`
	EmployeeSkills   []string `json:"employee_skills"`
	EmployeeLocation string   `json:"employee_location"`
	YearsExperience  int      `json:"years_experience"`
	GitHubUsername   string   `json:"github_username"`
	ProfileScore     int      `json:"profile_score"`
}

// CreateTalentPoolRequest represents request to create a talent pool
type CreateTalentPoolRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=255"`
	Description string `json:"description"`
	Color       string `json:"color" validate:"omitempty,hexcolor"`
	Icon        string `json:"icon"`
}

// UpdateTalentPoolRequest represents request to update a talent pool
type UpdateTalentPoolRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
	Icon        *string `json:"icon"`
	IsActive    *bool   `json:"is_active"`
}

// AddCandidateRequest represents request to add a candidate to talent pool
type AddCandidateRequest struct {
	EmployeeID string   `json:"employee_id" validate:"required"`
	MatchScore int      `json:"match_score"`
	Notes      string   `json:"notes"`
	Tags       []string `json:"tags"`
}

// UpdateCandidateRequest represents request to update a candidate in talent pool
type UpdateCandidateRequest struct {
	Notes *string   `json:"notes"`
	Tags  []string  `json:"tags"`
	Status *string  `json:"status" validate:"omitempty,oneof=active contacted hired archived"`
}

// TalentPoolListResponse represents paginated list of talent pools
type TalentPoolListResponse struct {
	Pools      []*TalentPool `json:"pools"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

// CandidateListResponse represents paginated list of candidates in a pool
type CandidateListResponse struct {
	Candidates []*TalentPoolCandidateWithDetails `json:"candidates"`
	Total      int64                             `json:"total"`
	Page       int                               `json:"page"`
	Limit      int                               `json:"limit"`
	TotalPages int                               `json:"total_pages"`
}

// TalentPoolStats represents statistics for a talent pool
type TalentPoolStats struct {
	TotalCandidates   int            `json:"total_candidates"`
	ActiveCandidates  int            `json:"active_candidates"`
	ContactedCount    int            `json:"contacted_count"`
	HiredCount        int            `json:"hired_count"`
	ArchivedCount     int            `json:"archived_count"`
	TagsDistribution  map[string]int `json:"tags_distribution"`
	AverageMatchScore float64        `json:"average_match_score"`
}

func (TalentPool) TableName() string {
	return "talent_pools"
}

func (TalentPoolCandidate) TableName() string {
	return "talent_pool_candidates"
}