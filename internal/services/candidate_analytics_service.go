package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type CandidateAnalyticsService interface {
	// Dashboard
	GetDashboard(ctx context.Context, employeeID string) (*models.CandidateDashboard, error)

	// Profile strength
	GetProfileStrength(ctx context.Context, employeeID string) (*models.ProfileStrength, error)

	// Application analytics
	GetApplicationStats(ctx context.Context, employeeID string) (*models.ApplicationStats, error)
	GetMonthlyStats(ctx context.Context, employeeID string, months int) ([]models.MonthlyStats, error)

	// Success metrics
	GetSuccessRates(ctx context.Context, employeeID string) (*models.SuccessRate, error)

	// Skill analysis
	GetSkillGapAnalysis(ctx context.Context, employeeID string) (*models.SkillGapAnalysisData, error)

	// Market positioning
	GetMarketPositioning(ctx context.Context, employeeID string) (*models.MarketPositioning, error)

	// Recommendations
	GetRecommendations(ctx context.Context, employeeID string) ([]models.Recommendation, error)

	// Activity feed
	GetRecentActivity(ctx context.Context, employeeID string, limit int) ([]models.RecentActivity, error)
}

type CandidateAnalyticsServiceImpl struct {
	cfg             *config.Config
	candidateRepo   repository.CandidateAnalyticsRepository
	jobRepo         repository.JobRepository
	applicationRepo repository.ApplicationRepository
	matchingSvc     MatchingService
}

func NewCandidateAnalyticsService(
	cfg *config.Config,
	candidateRepo repository.CandidateAnalyticsRepository,
	jobRepo repository.JobRepository,
	applicationRepo repository.ApplicationRepository,
	matchingSvc MatchingService,
) CandidateAnalyticsService {
	return &CandidateAnalyticsServiceImpl{
		cfg:             cfg,
		candidateRepo:   candidateRepo,
		jobRepo:         jobRepo,
		applicationRepo: applicationRepo,
		matchingSvc:     matchingSvc,
	}
}

func (s *CandidateAnalyticsServiceImpl) GetDashboard(ctx context.Context, employeeID string) (*models.CandidateDashboard, error) {
	dashboard := &models.CandidateDashboard{}

	var err error

	// Get profile strength
	dashboard.ProfileStrength, err = s.GetProfileStrength(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile strength: %w", err)
	}

	// Get application stats
	dashboard.ApplicationStats, err = s.GetApplicationStats(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get application stats: %w", err)
	}

	// Get success rates
	dashboard.SuccessRate, err = s.GetSuccessRates(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get success rates: %w", err)
	}

	// Get skill gap analysis
	dashboard.SkillGapAnalysis, err = s.GetSkillGapAnalysis(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill gap analysis: %w", err)
	}

	// Get market positioning
	dashboard.MarketPositioning, err = s.GetMarketPositioning(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get market positioning: %w", err)
	}

	// Get recommendations
	dashboard.Recommendations, err = s.GetRecommendations(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}

	// Get recent activity
	dashboard.RecentActivity, err = s.GetRecentActivity(ctx, employeeID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	return dashboard, nil
}

func (s *CandidateAnalyticsServiceImpl) GetProfileStrength(ctx context.Context, employeeID string) (*models.ProfileStrength, error) {
	stats := &models.ProfileStrength{
		CategoryScores:  make(map[string]models.CategoryScore),
		ImprovementTips: []models.ImprovementTip{},
		NextSteps:       []string{},
		CompletedSteps:  []string{},
	}

	totalScore := 0
	maxScore := 0

	// Get employee profile
	profile, err := s.candidateRepo.GetEmployeeProfile(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	// Category 1: Basic Information (25 points)
	basicScore := 0
	basicMax := 25

	if profile.FullName != "" && len(profile.FullName) > 2 {
		basicScore += 10
		stats.CompletedSteps = append(stats.CompletedSteps, "basic_info")
	} else {
		stats.ImprovementTips = append(stats.ImprovementTips, models.ImprovementTip{
			Title:       "Complete Your Profile",
			Description: "Add your full name to help employers recognize you",
			Priority:    "high",
			ActionURL:   "/employee/profile/edit",
		})
		stats.NextSteps = append(stats.NextSteps, "Add your full name")
	}

	if profile.Headline != "" {
		basicScore += 15
	} else {
		stats.ImprovementTips = append(stats.ImprovementTips, models.ImprovementTip{
			Title:       "Add a Professional Headline",
			Description: "Example: 'Senior Full Stack Developer with 5+ years experience'",
			Priority:    "high",
			ActionURL:   "/employee/profile/edit",
		})
		stats.NextSteps = append(stats.NextSteps, "Create a professional headline")
	}

	stats.CategoryScores["basic_info"] = models.CategoryScore{
		Name:        "Basic Information",
		Score:       basicScore,
		MaxScore:    basicMax,
		Completed:   basicScore >= 20,
		Required:    true,
		Description: "Your name, headline, and basic details",
	}
	totalScore += basicScore
	maxScore += basicMax

	// Category 2: Skills (30 points)
	skillsScore := 0
	skillsMax := 30
	skillCount := len(profile.Skills)

	if skillCount >= 8 {
		skillsScore = 30
		stats.CompletedSteps = append(stats.CompletedSteps, "skills")
	} else if skillCount >= 5 {
		skillsScore = 25
	} else if skillCount >= 3 {
		skillsScore = 15
	} else if skillCount >= 1 {
		skillsScore = 10
	}

	if skillCount < 5 {
		stats.ImprovementTips = append(stats.ImprovementTips, models.ImprovementTip{
			Title:       "Add More Skills",
			Description: "Add at least 5 skills to get matched with relevant jobs",
			Priority:    "high",
			ActionURL:   "/employee/profile/skills",
		})
		stats.NextSteps = append(stats.NextSteps, "Add more skills to your profile")
	}

	stats.CategoryScores["skills"] = models.CategoryScore{
		Name:        "Skills",
		Score:       skillsScore,
		MaxScore:    skillsMax,
		Completed:   skillCount >= 5,
		Required:    true,
		Description: "Technical skills and competencies",
	}
	totalScore += skillsScore
	maxScore += skillsMax

	// Category 3: Experience (25 points)
	expScore := 0
	expMax := 25

	if profile.YearsOfExperience > 0 {
		if profile.YearsOfExperience >= 7 {
			expScore = 25
		} else if profile.YearsOfExperience >= 5 {
			expScore = 20
		} else if profile.YearsOfExperience >= 3 {
			expScore = 15
		} else {
			expScore = 10
		}
		stats.CompletedSteps = append(stats.CompletedSteps, "experience")
	} else {
		stats.ImprovementTips = append(stats.ImprovementTips, models.ImprovementTip{
			Title:       "Add Your Experience",
			Description: "List your work experience to showcase your expertise",
			Priority:    "medium",
			ActionURL:   "/employee/profile/experience",
		})
		stats.NextSteps = append(stats.NextSteps, "Add your work experience")
	}

	if profile.ExperienceLevel == "" {
		expScore -= 5
		stats.ImprovementTips = append(stats.ImprovementTips, models.ImprovementTip{
			Title:       "Set Your Experience Level",
			Description: "Choose your seniority level (Entry, Junior, Mid, Senior, Lead)",
			Priority:    "medium",
			ActionURL:   "/employee/profile/edit",
		})
	}

	if expScore < 0 {
		expScore = 0
	}

	stats.CategoryScores["experience"] = models.CategoryScore{
		Name:        "Experience",
		Score:       expScore,
		MaxScore:    expMax,
		Completed:   expScore >= 15,
		Required:    true,
		Description: "Work experience and seniority level",
	}
	totalScore += expScore
	maxScore += expMax

	// Category 4: GitHub Connection (20 points)
	githubScore := 0
	githubMax := 20

	if profile.GithubConnected && profile.GetGithubUsername() != "" {
		githubScore = 20
		stats.CompletedSteps = append(stats.CompletedSteps, "github")
	} else {
		stats.ImprovementTips = append(stats.ImprovementTips, models.ImprovementTip{
			Title:       "Connect GitHub",
			Description: "Showcase your code and contributions to stand out",
			Priority:    "medium",
			ActionURL:   "/employee/github/connect",
		})
		stats.NextSteps = append(stats.NextSteps, "Connect your GitHub account")
	}

	stats.CategoryScores["github"] = models.CategoryScore{
		Name:        "GitHub Integration",
		Score:       githubScore,
		MaxScore:    githubMax,
		Completed:   githubScore == githubMax,
		Required:    false,
		Description: "Connect GitHub to showcase your code",
	}
	totalScore += githubScore
	maxScore += githubMax

	stats.OverallScore = (totalScore * 100) / maxScore
	stats.TotalSteps = len(stats.CategoryScores)

	return stats, nil
}

func (s *CandidateAnalyticsServiceImpl) GetApplicationStats(ctx context.Context, employeeID string) (*models.ApplicationStats, error) {
	stats, err := s.candidateRepo.GetApplicationStats(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	// Get monthly stats
	monthlyStats, err := s.candidateRepo.GetMonthlyApplicationStats(ctx, employeeID, 6)
	if err == nil {
		stats.ByMonth = monthlyStats
	}

	return stats, nil
}

func (s *CandidateAnalyticsServiceImpl) GetMonthlyStats(ctx context.Context, employeeID string, months int) ([]models.MonthlyStats, error) {
	return s.candidateRepo.GetMonthlyApplicationStats(ctx, employeeID, months)
}

func (s *CandidateAnalyticsServiceImpl) GetSuccessRates(ctx context.Context, employeeID string) (*models.SuccessRate, error) {
	return s.candidateRepo.GetSuccessRates(ctx, employeeID)
}

func (s *CandidateAnalyticsServiceImpl) GetSkillGapAnalysis(ctx context.Context, employeeID string) (*models.SkillGapAnalysisData, error) {
	analysis := &models.SkillGapAnalysisData{
		MatchingSkills:     []models.SkillMatch{},
		MissingSkills:      []models.SkillGapDetail{},
		TransferableSkills: []string{},
		RecommendedSkills:  []models.RecommendedSkill{},
		SkillDemandTrends:  make(map[string]int),
		TopJobMatches:      []models.JobMatchSuggestion{},
	}

	// Get employee skills
	profile, err := s.candidateRepo.GetEmployeeProfile(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	// Get top in-demand skills from job postings
	var topSkills []struct {
		Skill    string
		JobCount int
	}

	s.candidateRepo.GetTopInDemandSkills(ctx, 20, &topSkills)

	employeeSkillsMap := make(map[string]bool)
	for _, s := range profile.Skills {
		employeeSkillsMap[s] = true
	}

	// Categorize skills
	for _, ts := range topSkills {
		if employeeSkillsMap[ts.Skill] {
			analysis.MatchingSkills = append(analysis.MatchingSkills, models.SkillMatch{
				Name:        ts.Skill,
				DemandScore: ts.JobCount,
				MatchScore:  85,
				JobCount:    ts.JobCount,
			})
		} else {
			analysis.MissingSkills = append(analysis.MissingSkills, models.SkillGapDetail{
				Name:       ts.Skill,
				Importance: "important",
				JobCount:   ts.JobCount,
				LearningResources: []models.LearningResource{
					{
						Name:     "Complete " + ts.Skill + " Course",
						Platform: "Coursera",
						URL:      "https://www.coursera.org/search?query=" + ts.Skill,
						Duration: "4-6 weeks",
						Free:     false,
					},
					{
						Name:     ts.Skill + " Tutorial",
						Platform: "YouTube",
						URL:      "https://www.youtube.com/results?search_query=" + ts.Skill + "+tutorial",
						Duration: "5-10 hours",
						Free:     true,
					},
				},
				EstimatedTime: "2-4 weeks",
			})
		}
	}

	// Get recommended skills (top 5 missing skills)
	for i, ms := range analysis.MissingSkills {
		if i >= 5 {
			break
		}
		analysis.RecommendedSkills = append(analysis.RecommendedSkills, models.RecommendedSkill{
			Name:             ms.Name,
			Reason:           "In high demand with " + strconv.Itoa(ms.JobCount) + " job opportunities",
			Priority:         "high",
			TimeToLearn:      "2-4 weeks",
			JobOpportunities: ms.JobCount,
		})
	}

	// Get skill demand trends
	analysis.SkillDemandTrends, _ = s.candidateRepo.GetSkillDemandTrends(ctx, profile.Skills)

	// Get top job matches from matching service
	matches, err := s.matchingSvc.GetJobMatches(ctx, employeeID, 5)
	if err == nil {
		for _, match := range matches {
			analysis.TopJobMatches = append(analysis.TopJobMatches, models.JobMatchSuggestion{
				JobID:         match.JobID,
				Title:         match.JobTitle,
				Company:       match.CompanyName,
				MatchScore:    match.OverallScore,
				SkillsMatch:   match.MatchingSkills,
				MissingSkills: match.MissingSkills,
			})
		}
	}

	return analysis, nil
}

func (s *CandidateAnalyticsServiceImpl) GetMarketPositioning(ctx context.Context, employeeID string) (*models.MarketPositioning, error) {
	positioning, err := s.candidateRepo.GetMarketPositioning(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	// Get employee profile for skills
	profile, err := s.candidateRepo.GetEmployeeProfile(ctx, employeeID)
	if err != nil {
		return positioning, nil
	}

	// Get skill percentiles
	if len(profile.Skills) > 0 {
		percentiles, err := s.candidateRepo.GetSkillPercentiles(ctx, profile.Skills)
		if err == nil {
			positioning.SkillPercentile = percentiles
		}
	}

	// Calculate areas for improvement
	positioning.AreasForImprovement = []string{}

	if len(profile.Skills) < 5 {
		positioning.AreasForImprovement = append(positioning.AreasForImprovement, "Expand your skill set")
	}

	if profile.YearsOfExperience < 2 {
		positioning.AreasForImprovement = append(positioning.AreasForImprovement, "Gain more experience")
	}

	if !profile.GithubConnected {
		positioning.AreasForImprovement = append(positioning.AreasForImprovement, "Connect GitHub to showcase your work")
	}

	return positioning, nil
}

func (s *CandidateAnalyticsServiceImpl) GetRecommendations(ctx context.Context, employeeID string) ([]models.Recommendation, error) {
	recommendations := []models.Recommendation{}

	// Get profile strength to identify gaps
	strength, err := s.GetProfileStrength(ctx, employeeID)
	if err == nil {
		for _, tip := range strength.ImprovementTips {
			recType := "profile"
			if tip.Priority == "high" {
				recType = "urgent"
			}
			recommendations = append(recommendations, models.Recommendation{
				Type:        recType,
				Title:       tip.Title,
				Description: tip.Description,
				ActionURL:   tip.ActionURL,
				Priority:    tip.Priority,
			})
		}
	}

	// Get skill recommendations
	skillGap, err := s.GetSkillGapAnalysis(ctx, employeeID)
	if err == nil && len(skillGap.RecommendedSkills) > 0 {
		for i, rs := range skillGap.RecommendedSkills {
			if i >= 3 {
				break
			}
			recommendations = append(recommendations, models.Recommendation{
				Type:        "skill",
				Title:       "Learn " + rs.Name,
				Description: rs.Reason,
				ActionURL:   "/employee/learning/resources",
				Priority:    rs.Priority,
			})
		}
	}

	// Get job recommendations
	matches, err := s.matchingSvc.GetJobMatches(ctx, employeeID, 3)
	if err == nil && len(matches) > 0 {
		for _, match := range matches {
			if match.OverallScore >= 80 {
				recommendations = append(recommendations, models.Recommendation{
					Type:        "job",
					Title:       "Apply to " + match.JobTitle,
					Description: "This job matches your profile with " + strconv.Itoa(match.OverallScore) + "% match score",
					ActionURL:   "/jobs/" + match.JobID,
					Priority:    "high",
				})
				break
			}
		}
	}

	// Limit to top 5 recommendations
	if len(recommendations) > 5 {
		recommendations = recommendations[:5]
	}

	return recommendations, nil
}

func (s *CandidateAnalyticsServiceImpl) GetRecentActivity(ctx context.Context, employeeID string, limit int) ([]models.RecentActivity, error) {
	return s.candidateRepo.GetRecentActivity(ctx, employeeID, limit)
}
