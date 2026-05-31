package models

import (
	"time"
)

// ApplicationTrend represents application volume over time
type ApplicationTrend struct {
	Date          string `json:"date"`
	Total         int    `json:"total"`
	Pending       int    `json:"pending"`
	Viewed        int    `json:"viewed"`
	Shortlisted   int    `json:"shortlisted"`
	Rejected      int    `json:"rejected"`
	Hired         int    `json:"hired"`
	Withdrawn     int    `json:"withdrawn"`
}

// ConversionFunnel represents the application pipeline
type ConversionFunnel struct {
	Stage        string  `json:"stage"`
	Count        int     `json:"count"`
	ConversionRate float64 `json:"conversion_rate"`
	DropOffRate    float64 `json:"drop_off_rate"`
}

// JobPerformance represents performance metrics for a job
type JobPerformance struct {
	JobID              string    `json:"job_id"`
	Title              string    `json:"title"`
	Views              int       `json:"views"`
	Applications       int       `json:"applications"`
	Shortlisted        int       `json:"shortlisted"`
	Hired              int       `json:"hired"`
	ViewToAppRate      float64   `json:"view_to_app_rate"`
	AppToShortlistRate float64   `json:"app_to_shortlist_rate"`
	ShortlistToHireRate float64  `json:"shortlist_to_hire_rate"`
	OverallSuccessRate float64   `json:"overall_success_rate"`
	DaysToHire         *int      `json:"days_to_hire,omitempty"`
	CostPerHire        *float64  `json:"cost_per_hire,omitempty"`
	PostedAt           time.Time `json:"posted_at"`
}

// TimeToHireMetric represents time-to-hire statistics
type TimeToHireMetric struct {
	AverageDays    int     `json:"average_days"`
	MedianDays     int     `json:"median_days"`
	MinDays        int     `json:"min_days"`
	MaxDays        int     `json:"max_days"`
	ByJobTitle     map[string]int `json:"by_job_title"`
	ByExperienceLevel map[string]int `json:"by_experience_level"`
}

// ApplicationQualityScore represents the quality of applications
type ApplicationQualityScore struct {
	AverageMatchScore   float64 `json:"average_match_score"`
	HighQualityCount    int     `json:"high_quality_count"`    // Score >= 80
	MediumQualityCount  int     `json:"medium_quality_count"`  // Score 50-79
	LowQualityCount     int     `json:"low_quality_count"`     // Score < 50
	AverageResponseTime float64 `json:"average_response_time"` // Hours
}

// SourceAnalytics tracks where applicants come from
type SourceAnalytics struct {
	Source      string  `json:"source"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
	ConversionRate float64 `json:"conversion_rate"`
}

// ApplicationTrendsResponse represents complete trends data
type ApplicationTrendsResponse struct {
	Daily   []ApplicationTrend `json:"daily"`
	Weekly  []ApplicationTrend `json:"weekly"`
	Monthly []ApplicationTrend `json:"monthly"`
	Summary TrendSummary        `json:"summary"`
}

// TrendSummary provides high-level metrics
type TrendSummary struct {
	TotalApplications    int     `json:"total_applications"`
	AveragePerDay        float64 `json:"average_per_day"`
	GrowthRate           float64 `json:"growth_rate"`           // Percentage change
	PeakDay              string  `json:"peak_day"`
	PeakApplications     int     `json:"peak_applications"`
	BestPerformingSource string  `json:"best_performing_source"`
}

// FunnelResponse represents the complete funnel analysis
type FunnelResponse struct {
	Stages      []ConversionFunnel `json:"stages"`
	OverallRate float64            `json:"overall_rate"`
	DropOffPoints []DropOffPoint   `json:"drop_off_points"`
}

// DropOffPoint identifies where candidates leave
type DropOffPoint struct {
	FromStage   string  `json:"from_stage"`
	ToStage     string  `json:"to_stage"`
	DropOffCount int    `json:"drop_off_count"`
	DropOffRate  float64 `json:"drop_off_rate"`
	Suggestion   string  `json:"suggestion"`
}

// ExportFormat defines export options
type ExportFormat string

const (
	FormatCSV  ExportFormat = "csv"
	FormatPDF  ExportFormat = "pdf"
	FormatJSON ExportFormat = "json"
)