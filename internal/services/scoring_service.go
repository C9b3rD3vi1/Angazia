package services

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/ai"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type ScoringService interface {
	// Core scoring
	ScoreJobCandidate(ctx context.Context, job *models.Job, employee *models.EmployeeProfile, github *models.GithubProfile) (*MatchResult, error)
	ScoreCandidatesForJob(ctx context.Context, job *models.Job, employees []*models.EmployeeProfile) ([]*MatchResult, error)
	ScoreJobsForCandidate(ctx context.Context, employee *models.EmployeeProfile, github *models.GithubProfile, jobs []*models.Job) ([]*MatchResult, error)

	// Profile scoring
	ScoreProfileComplete(profile *models.EmployeeProfile) int
	ScoreGitHubActivity(github *models.GithubProfile) int

	// Weight management
	GetDefaultWeights() *MatchWeights
	CalculateOverall(skillsScore, experienceScore, locationScore, cultureScore, salaryScore int, weights *MatchWeights) int
}

type MatchWeights struct {
	Skills     int
	Experience int
	Location   int
	Culture    int
	Salary     int
}

func DefaultMatchWeights() *MatchWeights {
	return &MatchWeights{
		Skills:     40,
		Experience: 25,
		Location:   15,
		Culture:    10,
		Salary:     10,
	}
}

type ScoringServiceImpl struct {
	cfg       *ai.Config
	matcher   *ai.Matcher
	aiFactory *ai.ProviderFactory
	aiProvider ai.AIProvider
	jobRepo   repository.JobRepository
	userRepo  repository.UserRepository
	mu        sync.RWMutex
}

func NewScoringService(
	cfg *ai.Config,
	jobRepo repository.JobRepository,
	userRepo repository.UserRepository,
) (*ScoringServiceImpl, error) {
	matcher := ai.NewMatcher()
	factory := ai.NewProviderFactory(cfg)
	provider, err := factory.GetProvider()
	if err != nil {
		provider = nil
	}

	return &ScoringServiceImpl{
		cfg:       cfg,
		matcher:   matcher,
		aiFactory: factory,
		aiProvider: provider,
		jobRepo:   jobRepo,
		userRepo:  userRepo,
	}, nil
}

func (s *ScoringServiceImpl) ScoreJobCandidate(ctx context.Context, job *models.Job, employee *models.EmployeeProfile, github *models.GithubProfile) (*MatchResult, error) {
	if job == nil || employee == nil {
		return nil, fmt.Errorf("job and employee are required")
	}

	// Algorithmic scoring first
	matchingSkills, missingSkills, skillsScore := s.matcher.CalculateSkillMatch(job.RequiredSkills, employee.Skills)
	experienceScore := s.matcher.CalculateExperienceMatch(job.MinExperience, job.MaxExperience, employee.YearsOfExperience)
	locationScore := s.matcher.CalculateLocationMatch(job.Location, job.IsRemote, employee.Location, employee.IsRemoteOnly)

	var githubActivity *ai.GithubActivity
	if github != nil {
		githubActivity = &ai.GithubActivity{
			PublicRepos:        github.PublicRepos,
			TotalCommits:       github.TotalCommits,
			Followers:          github.Followers,
			ContributionStreak: github.ContributionStreak,
			TopLanguages:       github.GetTopLanguages(5),
			AccountAgeDays:     github.AccountAgeDays,
			ActivityScore:      github.ActivityScore,
			QualityScore:       github.QualityScore,
		}
	}

	profileComplete := s.ScoreProfileComplete(employee)
	cultureScore := s.matcher.CalculateCultureMatch(githubActivity, profileComplete)
	expectedSalary := 0
	salaryScore := s.calculateSalaryMatch(job.SalaryMin, job.SalaryMax, expectedSalary)

	weights := DefaultMatchWeights()
	overallScore := s.CalculateOverall(skillsScore, experienceScore, locationScore, cultureScore, salaryScore, weights)

	recommendation := s.matcher.GetRecommendation(overallScore)
	summary := s.matcher.GenerateMatchSummary(job.Title, employee.FullName, overallScore, matchingSkills, missingSkills)

	initials := ""
	for _, part := range strings.Fields(employee.FullName) {
		if len(part) > 0 {
			initials += strings.ToUpper(part[:1])
		}
	}

	companyLogo := ""
	if job.Employer != nil {
		companyLogo = job.Employer.CompanyLogo
	}
	result := &MatchResult{
		JobID:             job.ID,
		EmployeeID:        employee.UserID,
		JobTitle:          job.Title,
		CompanyName:       job.Employer.CompanyName,
		CompanyLogo:       companyLogo,
		OverallScore:      overallScore,
		SkillsScore:       skillsScore,
		ExperienceScore:   experienceScore,
		CultureScore:      cultureScore,
		LocationScore:     locationScore,
		Summary:           summary,
		Recommendation:    recommendation,
		MatchingSkills:    matchingSkills,
		MissingSkills:     missingSkills,
		MatchID:           fmt.Sprintf("match-%s-%s", job.ID[:8], employee.UserID[:8]),
		AnalyzedAt:        time.Now(),
		CandidateName:     employee.FullName,
		CandidateHeadline: employee.Headline,
		CandidateLocation: employee.Location,
		CandidateAvatar:   avatarFromProfile(employee),
		CandidateInitials: initials,
		ExperienceYears:   employee.YearsOfExperience,
		Skills:            employee.Skills,
	}

	// Try to enhance with AI if available
	if s.aiProvider != nil {
		aiResult, err := s.enhanceWithAI(ctx, job, employee, githubActivity, result)
		if err == nil && aiResult != nil {
			result = aiResult
		}
	}

	return result, nil
}

func (s *ScoringServiceImpl) ScoreCandidatesForJob(ctx context.Context, job *models.Job, employees []*models.EmployeeProfile) ([]*MatchResult, error) {
	if job == nil {
		return nil, fmt.Errorf("job is required")
	}

	results := make([]*MatchResult, 0, len(employees))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	for _, emp := range employees {
		wg.Add(1)
		go func(employee *models.EmployeeProfile) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			github, _ := s.userRepo.GetGithubProfileByEmployeeID(ctx, employee.UserID)
			match, err := s.ScoreJobCandidate(ctx, job, employee, github)
			if err != nil {
				return
			}

			mu.Lock()
			results = append(results, match)
			mu.Unlock()
		}(emp)
	}

	wg.Wait()

	// Sort by score descending
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].OverallScore > results[i].OverallScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results, nil
}

func (s *ScoringServiceImpl) ScoreJobsForCandidate(ctx context.Context, employee *models.EmployeeProfile, github *models.GithubProfile, jobs []*models.Job) ([]*MatchResult, error) {
	if employee == nil {
		return nil, fmt.Errorf("employee is required")
	}

	results := make([]*MatchResult, 0, len(jobs))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	for _, job := range jobs {
		wg.Add(1)
		go func(j *models.Job) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			match, err := s.ScoreJobCandidate(ctx, j, employee, github)
			if err != nil {
				return
			}

			mu.Lock()
			results = append(results, match)
			mu.Unlock()
		}(job)
	}

	wg.Wait()

	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].OverallScore > results[i].OverallScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results, nil
}

func (s *ScoringServiceImpl) ScoreProfileComplete(profile *models.EmployeeProfile) int {
	if profile == nil {
		return 0
	}

	score := 0
	maxScore := 10

	if profile.FullName != "" && len(profile.FullName) > 2 {
		score++
	}
	if profile.Headline != "" {
		score++
	}
	if profile.Bio != "" && len(profile.Bio) > 50 {
		score++
	}
	if profile.Location != "" {
		score++
	}
	if profile.YearsOfExperience > 0 {
		score++
	}
	if len(profile.Skills) >= 5 {
		score++
	}
	if profile.ExperienceLevel != "" {
		score++
	}
	if profile.GithubConnected {
		score++
	}
	if avatarFromProfile(profile) != "" {
		score++
	}
	if profile.ResumeURL != "" {
		score++
	}

	return int(math.Round(float64(score) / float64(maxScore) * 100))
}

func (s *ScoringServiceImpl) ScoreGitHubActivity(github *models.GithubProfile) int {
	if github == nil {
		return 0
	}

	score := 0
	if github.TotalCommits > 500 {
		score = 100
	} else if github.TotalCommits > 200 {
		score = 80
	} else if github.TotalCommits > 50 {
		score = 60
	} else if github.TotalCommits > 10 {
		score = 40
	} else if github.TotalCommits > 0 {
		score = 20
	}

	if github.ContributionStreak > 30 {
		score = int(math.Min(float64(score+20), 100))
	} else if github.ContributionStreak > 14 {
		score = int(math.Min(float64(score+10), 100))
	}

	if github.PublicRepos > 10 {
		score = int(math.Min(float64(score+10), 100))
	}

	if github.ActivityScore > 0 {
		score = int(math.Min(float64(score+github.ActivityScore)/2, 100))
	}

	return score
}

func (s *ScoringServiceImpl) GetDefaultWeights() *MatchWeights {
	return DefaultMatchWeights()
}

func (s *ScoringServiceImpl) CalculateOverall(skillsScore, experienceScore, locationScore, cultureScore, salaryScore int, weights *MatchWeights) int {
	if weights == nil {
		weights = DefaultMatchWeights()
	}

	total := weights.Skills + weights.Experience + weights.Location + weights.Culture + weights.Salary
	if total == 0 {
		return 0
	}

	overall := (skillsScore*weights.Skills +
		experienceScore*weights.Experience +
		locationScore*weights.Location +
		cultureScore*weights.Culture +
		salaryScore*weights.Salary) / total

	if overall > 100 {
		overall = 100
	}
	if overall < 0 {
		overall = 0
	}

	return overall
}

func (s *ScoringServiceImpl) enhanceWithAI(ctx context.Context, job *models.Job, employee *models.EmployeeProfile, githubActivity *ai.GithubActivity, base *MatchResult) (*MatchResult, error) {
	if s.aiProvider == nil {
		return base, nil
	}

	jobDesc := ai.JobDescription{
		ID:               job.ID,
		Title:            job.Title,
		Description:      job.Description,
		Requirements:     job.Requirements,
		Responsibilities: job.Responsibilities,
		RequiredSkills:   job.RequiredSkills,
		NiceToHaveSkills: job.NiceToHaveSkills,
		ExperienceLevel:  job.ExperienceLevel,
		MinExperience:    job.MinExperience,
		MaxExperience:    job.MaxExperience,
		EmploymentType:   job.EmploymentType,
		Location:         job.Location,
		IsRemote:         job.IsRemote,
	}

	candidateProfile := ai.CandidateProfile{
		ID:                employee.UserID,
		FullName:          employee.FullName,
		Headline:          employee.Headline,
		Bio:               employee.Bio,
		Skills:            employee.Skills,
		ExperienceLevel:   employee.ExperienceLevel,
		YearsOfExperience: employee.YearsOfExperience,
		Location:          employee.Location,
		IsRemoteOnly:      employee.IsRemoteOnly,
		GithubUsername:    employee.GithubUsername,
		GithubActivity:    githubActivity,
	}

	analysis, err := s.aiProvider.GenerateMatchAnalysis(ctx, jobDesc, candidateProfile)
	if err != nil {
		return base, nil
	}

	base.OverallScore = blendScores(base.OverallScore, analysis.OverallScore, 60)
	base.SkillsScore = blendScores(base.SkillsScore, analysis.SkillsScore, 50)
	base.ExperienceScore = blendScores(base.ExperienceScore, analysis.ExperienceScore, 50)
	base.CultureScore = blendScores(base.CultureScore, analysis.CultureScore, 60)
	base.LocationScore = blendScores(base.LocationScore, analysis.LocationScore, 50)

	if len(analysis.MatchingSkills) > 0 {
		seen := make(map[string]bool)
		for _, s := range base.MatchingSkills {
			seen[strings.ToLower(s)] = true
		}
		for _, s := range analysis.MatchingSkills {
			if !seen[strings.ToLower(s)] {
				base.MatchingSkills = append(base.MatchingSkills, s)
			}
		}
	}

	if len(analysis.MissingSkills) > 0 {
		seen := make(map[string]bool)
		for _, s := range base.MissingSkills {
			seen[strings.ToLower(s)] = true
		}
		for _, s := range analysis.MissingSkills {
			if !seen[strings.ToLower(s)] {
				base.MissingSkills = append(base.MissingSkills, s)
			}
		}
	}

	if analysis.Summary != "" {
		base.Summary = analysis.Summary
	}
	if analysis.Recommendation != "" {
		base.Recommendation = analysis.Recommendation
	}

	base.Recommendation = s.matcher.GetRecommendation(base.OverallScore)
	return base, nil
}

func (s *ScoringServiceImpl) calculateSalaryMatch(jobMin, jobMax, expectedSalary int) int {
	if expectedSalary <= 0 || (jobMin <= 0 && jobMax <= 0) {
		return 70
	}

	if jobMin > 0 && jobMax > 0 {
		if expectedSalary >= jobMin && expectedSalary <= jobMax {
			return 100
		}
		if expectedSalary < jobMin {
			ratio := float64(expectedSalary) / float64(jobMin)
			return int(math.Min(ratio*100, 90))
		}
		if expectedSalary > jobMax {
			ratio := float64(jobMax) / float64(expectedSalary)
			return int(math.Min(ratio*100, 80))
		}
	}

	if jobMin > 0 && expectedSalary >= jobMin {
		ratio := float64(jobMin) / float64(expectedSalary)
		return int(math.Min(ratio*100, 90))
	}

	return 50
}

func blendScores(algorithmic, aiScore, algorithmWeight int) int {
	if aiScore <= 0 {
		return algorithmic
	}
	if algorithmWeight > 100 {
		algorithmWeight = 50
	}
	aiWeight := 100 - algorithmWeight
	return (algorithmic*algorithmWeight + aiScore*aiWeight) / 100
}

func avatarFromProfile(profile *models.EmployeeProfile) string {
	if profile == nil {
		return ""
	}
	if profile.User != nil && profile.User.AvatarURL != "" {
		return profile.User.AvatarURL
	}
	return ""
}
