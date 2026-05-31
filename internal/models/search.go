package models

import (
	"time"
)

// SearchQuery represents a user's search query with filters
type SearchQuery struct {
	ID          string                 `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string                 `json:"user_id" gorm:"type:uuid;not null;index"`
	Query       string                 `json:"query" gorm:"type:text"`
	Filters     SearchFilters          `json:"filters" gorm:"type:jsonb"`
	EntityType  string                 `json:"entity_type" gorm:"size:50;default:'job'"` // job, candidate, company
	ResultsCount int                   `json:"results_count" gorm:"default:0"`
	IPAddress   string                 `json:"ip_address" gorm:"size:45"`
	UserAgent   string                 `json:"user_agent" gorm:"type:text"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime;index"`
}

// SearchFilters represents advanced search criteria
type SearchFilters struct {
	// Common filters
	Keywords        string   `json:"keywords,omitempty"`
	Location        string   `json:"location,omitempty"`
	Radius          int      `json:"radius,omitempty"` // km
	IsRemote        *bool    `json:"is_remote,omitempty"`
	IsHybrid        *bool    `json:"is_hybrid,omitempty"`
	
	// Job-specific filters
	JobTitle        string   `json:"job_title,omitempty"`
	CompanyName     string   `json:"company_name,omitempty"`
	Industry        string   `json:"industry,omitempty"`
	EmploymentType  string   `json:"employment_type,omitempty"`
	ExperienceLevel string   `json:"experience_level,omitempty"`
	MinSalary       int      `json:"min_salary,omitempty"`
	MaxSalary       int      `json:"max_salary,omitempty"`
	Skills          []string `json:"skills,omitempty"`
	PostedWithin    string   `json:"posted_within,omitempty"` // 24h, 7d, 30d, 90d
	
	// Candidate-specific filters
	MinExperience   int      `json:"min_experience,omitempty"`
	MaxExperience   int      `json:"max_experience,omitempty"`
	EducationLevel  string   `json:"education_level,omitempty"`
	GitHubConnected *bool    `json:"github_connected,omitempty"`
	MinMatchScore   int      `json:"min_match_score,omitempty"`
	
	// Company-specific filters
	CompanySize     string   `json:"company_size,omitempty"`
	IsVerified      *bool    `json:"is_verified,omitempty"`
	
	// Sorting
	SortBy          string   `json:"sort_by,omitempty"`  // relevance, date, salary, match_score
	SortOrder       string   `json:"sort_order,omitempty"` // asc, desc
}

// SavedSearch represents a user's saved search
type SavedSearch struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string         `json:"user_id" gorm:"type:uuid;not null;index"`
	Name        string         `json:"name" gorm:"size:255;not null"`
	Filters     SearchFilters  `json:"filters" gorm:"type:jsonb;not null"`
	EntityType  string         `json:"entity_type" gorm:"size:50;default:'job'"`
	Frequency   string         `json:"frequency" gorm:"size:50;default:'daily'"` // daily, weekly, instant
	IsActive    bool           `json:"is_active" gorm:"default:true;index"`
	LastRunAt   *time.Time     `json:"last_run_at,omitempty"`
	LastSentAt  *time.Time     `json:"last_sent_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User        *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	History     []AlertHistory `json:"history,omitempty" gorm:"foreignKey:SavedSearchID"`
}

// SearchResult represents a search result item
type SearchResult struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"` // job, candidate, company
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Score       float64     `json:"score"`
	Highlights  []Highlight `json:"highlights,omitempty"`
	Data        interface{} `json:"data"`
}

// Highlight represents a text highlight in search results
type Highlight struct {
	Field       string   `json:"field"`
	Text        string   `json:"text"`
	MatchedTerms []string `json:"matched_terms"`
}

// SearchResponse represents paginated search results
type SearchResponse struct {
	Results     []SearchResult `json:"results"`
	Total       int64          `json:"total"`
	Page        int            `json:"page"`
	Limit       int            `json:"limit"`
	TotalPages  int            `json:"total_pages"`
	Facets      FacetResult    `json:"facets,omitempty"`
	SearchTimeMs int64         `json:"search_time_ms"`
}

// FacetResult represents aggregation results for filtering
type FacetResult struct {
	Locations      []FacetCount `json:"locations,omitempty"`
	Skills         []FacetCount `json:"skills,omitempty"`
	ExperienceLevels []FacetCount `json:"experience_levels,omitempty"`
	EmploymentTypes []FacetCount `json:"employment_types,omitempty"`
	Industries     []FacetCount `json:"industries,omitempty"`
	SalaryRanges   []FacetCount `json:"salary_ranges,omitempty"`
}

// FacetCount represents a facet with count
type FacetCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

func (SearchQuery) TableName() string {
	return "search_queries"
}

func (SavedSearch) TableName() string {
	return "saved_searches"
}