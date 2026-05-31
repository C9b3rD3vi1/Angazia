package ai

import (
	"context"
	"time"
)

// AIProvider defines the interface for all AI providers
type AIProvider interface {
	// GenerateMatchAnalysis analyzes job and candidate fit
	GenerateMatchAnalysis(ctx context.Context, job JobDescription, candidate CandidateProfile) (*MatchAnalysis, error)
	
	// GenerateCoverLetter creates a personalized cover letter
	GenerateCoverLetter(ctx context.Context, job JobDescription, candidate CandidateProfile) (string, error)
	
	// AnalyzeSkillsGap identifies missing skills and provides recommendations
	AnalyzeSkillsGap(ctx context.Context, job JobDescription, candidate CandidateProfile) (*SkillsGapAnalysis, error)
	
	// GenerateInterviewQuestions creates role-specific interview questions
	GenerateInterviewQuestions(ctx context.Context, job JobDescription) ([]string, error)
	
	// GetProviderName returns the provider name
	GetProviderName() string
	
	// HealthCheck checks if the provider is available
	HealthCheck(ctx context.Context) error
}

// JobDescription represents a job posting for AI analysis
type JobDescription struct {
	ID               string
	Title            string
	Description      string
	Requirements     string
	Responsibilities string
	RequiredSkills   []string
	NiceToHaveSkills []string
	ExperienceLevel  string
	MinExperience    int
	MaxExperience    int
	EducationLevel   string
	Industry         string
	EmploymentType   string
	Location         string
	IsRemote         bool
}

// CandidateProfile represents a candidate for AI analysis
type CandidateProfile struct {
	ID                string
	FullName          string
	Headline          string
	Bio               string
	Skills            []string
	ExperienceLevel   string
	YearsOfExperience int
	Education         string
	Location          string
	IsRemoteOnly      bool
	GithubUsername    string
	GithubActivity    *GithubActivity
	ResumeText        string
}

// GithubActivity represents GitHub metrics
type GithubActivity struct {
	PublicRepos        int
	TotalCommits       int
	Followers          int
	ContributionStreak int
	TopLanguages       []string
	AccountAgeDays     int
	ActivityScore      int
	QualityScore       int
}

// MatchAnalysis contains AI-generated match insights
type MatchAnalysis struct {
	OverallScore      int                  `json:"overall_score"`
	SkillsScore       int                  `json:"skills_score"`
	ExperienceScore   int                  `json:"experience_score"`
	CultureScore      int                  `json:"culture_score"`
	LocationScore     int                  `json:"location_score"`
	
	MatchingSkills    []string             `json:"matching_skills"`
	MissingSkills     []string             `json:"missing_skills"`
	StrongPoints      []string             `json:"strong_points"`
	WeakPoints        []string             `json:"weak_points"`
	
	Summary           string               `json:"summary"`
	Recommendation    string               `json:"recommendation"`
	InterviewTips     []string             `json:"interview_tips"`
	
	AnalysisMetadata  AnalysisMetadata     `json:"analysis_metadata"`
}

// SkillsGapAnalysis contains detailed gap analysis
type SkillsGapAnalysis struct {
	MissingSkills        []SkillGap         `json:"missing_skills"`
	RecommendedCourses   []CourseRecommendation `json:"recommended_courses"`
	ImprovementPlan      string             `json:"improvement_plan"`
	EstimatedTimeToFill  string             `json:"estimated_time_to_fill"`
	TransferableSkills   []string           `json:"transferable_skills"`
	PriorityLevel        string             `json:"priority_level"` // high, medium, low
}

// SkillGap represents a missing skill with details
type SkillGap struct {
	SkillName        string   `json:"skill_name"`
	Importance       string   `json:"importance"` // critical, important, nice-to-have
	Description      string   `json:"description"`
	LearningResources []string `json:"learning_resources"`
}

// CourseRecommendation represents a recommended course
type CourseRecommendation struct {
	Name        string `json:"name"`
	Platform    string `json:"platform"` // Coursera, Udemy, Pluralsight, etc.
	URL         string `json:"url"`
	Duration    string `json:"duration"`
	Difficulty  string `json:"difficulty"` // beginner, intermediate, advanced
}

// AnalysisMetadata contains metadata about the analysis
type AnalysisMetadata struct {
	Provider        string    `json:"provider"`
	Model           string    `json:"model"`
	ProcessingTimeMs int64    `json:"processing_time_ms"`
	AnalyzedAt      time.Time `json:"analyzed_at"`
	TokensUsed      int       `json:"tokens_used,omitempty"`
}

// Config holds AI provider configuration
type Config struct {
	Provider      string        // openai, anthropic, gemini, local
	APIKey        string
	Model         string
	BaseURL       string        // for local LLMs
	Timeout       time.Duration
	MaxTokens     int
	Temperature   float64
	RetryAttempts int
	RetryDelay    time.Duration
}

// DefaultConfig returns default AI configuration
func DefaultConfig() *Config {
	return &Config{
		Provider:      "openai",
		Model:         "gpt-4-turbo-preview",
		Timeout:       30 * time.Second,
		MaxTokens:     2000,
		Temperature:   0.7,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
	}
}