package models

import (
	"time"
)

// CandidateDashboard represents the complete candidate dashboard data
type CandidateDashboard struct {
	ProfileStrength     *ProfileStrength      `json:"profile_strength"`
	ApplicationStats    *ApplicationStats     `json:"application_stats"`
	SuccessRate         *SuccessRate          `json:"success_rate"`
	SkillGapAnalysis    *SkillGapAnalysisData `json:"skill_gap_analysis"`
	MarketPositioning   *MarketPositioning    `json:"market_positioning"`
	Recommendations     []Recommendation      `json:"recommendations"`
	RecentActivity      []RecentActivity      `json:"recent_activity"`
}

// ProfileStrength represents profile completeness score
type ProfileStrength struct {
	OverallScore    int                       `json:"overall_score"` // 0-100
	CategoryScores  map[string]CategoryScore  `json:"category_scores"`
	ImprovementTips []ImprovementTip          `json:"improvement_tips"`
	NextSteps       []string                  `json:"next_steps"`
	CompletedSteps  []string                  `json:"completed_steps"`
	TotalSteps      int                       `json:"total_steps"`
}

// CategoryScore represents score for a profile category
type CategoryScore struct {
	Name        string `json:"name"`
	Score       int    `json:"score"`
	MaxScore    int    `json:"max_score"`
	Completed   bool   `json:"completed"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// ImprovementTip provides actionable advice
type ImprovementTip struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"` // high, medium, low
	ActionURL   string `json:"action_url"`
}

// ApplicationStats represents application statistics
type ApplicationStats struct {
	TotalApplications   int            `json:"total_applications"`
	ActiveApplications  int            `json:"active_applications"`
	ShortlistedCount    int            `json:"shortlisted_count"`
	RejectedCount       int            `json:"rejected_count"`
	HiredCount          int            `json:"hired_count"`
	WithdrawnCount      int            `json:"withdrawn_count"`
	PendingCount        int            `json:"pending_count"`
	InterviewCount      int            `json:"interview_count"`
	ResponseRate        float64        `json:"response_rate"`
	AverageResponseTime float64        `json:"average_response_time"` // days
	ByMonth             []MonthlyStats `json:"by_month"`
	ByStatus            map[string]int `json:"by_status"`
}

// MonthlyStats represents application statistics by month
type MonthlyStats struct {
	Month         string `json:"month"`
	Applications  int    `json:"applications"`
	Responses     int    `json:"responses"`
	ResponseRate  float64 `json:"response_rate"`
}

// SuccessRate represents application success metrics
type SuccessRate struct {
	OverallRate        float64                `json:"overall_rate"`
	ApplicationToShortlistRate float64        `json:"application_to_shortlist_rate"`
	ShortlistToInterviewRate float64          `json:"shortlist_to_interview_rate"`
	InterviewToHireRate float64               `json:"interview_to_hire_rate"`
	ByExperienceLevel  map[string]float64     `json:"by_experience_level"`
	ByJobType          map[string]float64     `json:"by_job_type"`
	BySkills           map[string]float64     `json:"by_skills"`
	IndustryAverages   map[string]float64     `json:"industry_averages"`
	PercentileRank     int                    `json:"percentile_rank"` // 0-100
}

// SkillGapAnalysisData represents skill gap analysis for candidate
type SkillGapAnalysisData struct {
	MatchingSkills     []SkillMatch           `json:"matching_skills"`
	MissingSkills      []SkillGapDetail       `json:"missing_skills"`
	TransferableSkills []string               `json:"transferable_skills"`
	RecommendedSkills  []RecommendedSkill     `json:"recommended_skills"`
	SkillDemandTrends  map[string]int         `json:"skill_demand_trends"`
	TopJobMatches      []JobMatchSuggestion   `json:"top_job_matches"`
}

// SkillMatch represents a skill the candidate has
type SkillMatch struct {
	Name          string  `json:"name"`
	DemandScore   int     `json:"demand_score"`   // 0-100 how in-demand this skill is
	MatchScore    int     `json:"match_score"`    // 0-100 how well they match
	JobCount      int     `json:"job_count"`      // number of jobs requesting this skill
}

// SkillGapDetail represents a missing skill with learning resources
type SkillGapDetail struct {
	Name           string   `json:"name"`
	Importance     string   `json:"importance"` // critical, important, nice-to-have
	JobCount       int      `json:"job_count"`  // number of jobs requiring this
	LearningResources []LearningResource `json:"learning_resources"`
	EstimatedTime  string   `json:"estimated_time"`
}

// LearningResource represents a course or tutorial
type LearningResource struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
	URL      string `json:"url"`
	Duration string `json:"duration"`
	Free     bool   `json:"free"`
}

// RecommendedSkill represents a skill to learn
type RecommendedSkill struct {
	Name           string   `json:"name"`
	Reason         string   `json:"reason"`
	Priority       string   `json:"priority"` // high, medium, low
	TimeToLearn    string   `json:"time_to_learn"`
	JobOpportunities int    `json:"job_opportunities"`
}

// JobMatchSuggestion represents a job recommendation
type JobMatchSuggestion struct {
	JobID       string  `json:"job_id"`
	Title       string  `json:"title"`
	Company     string  `json:"company"`
	MatchScore  int     `json:"match_score"`
	SkillsMatch []string `json:"skills_match"`
	MissingSkills []string `json:"missing_skills"`
}

// MarketPositioning represents how candidate compares to others
type MarketPositioning struct {
	TotalCandidates     int            `json:"total_candidates"`
	YourRank            int            `json:"your_rank"`
	Percentile          int            `json:"percentile"`
	SkillPercentile     map[string]int `json:"skill_percentile"`
	ExperiencePercentile int           `json:"experience_percentile"`
	SalaryPercentile    int            `json:"salary_percentile"`
	TopSkills           []string       `json:"top_skills"`
	UniqueStrengths     []string       `json:"unique_strengths"`
	AreasForImprovement []string       `json:"areas_for_improvement"`
}

// Recommendation represents personalized advice
type Recommendation struct {
	Type        string `json:"type"` // skill, job, profile, networking
	Title       string `json:"title"`
	Description string `json:"description"`
	ActionURL   string `json:"action_url"`
	Priority    string `json:"priority"`
}

// RecentActivity represents user's recent actions
type RecentActivity struct {
	Type      string    `json:"type"` // application, profile_update, job_view, skill_added
	Title     string    `json:"title"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}