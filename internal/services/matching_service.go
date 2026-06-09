package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/ai"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type MatchingService interface {
	// Job-Candidate Matching
	GetJobMatches(ctx context.Context, employeeID string, limit int) ([]*MatchResult, error)
	GetCandidateMatches(ctx context.Context, jobID string, employerID string, limit int) ([]*MatchResult, error)
	GetDetailedMatchAnalysis(ctx context.Context, jobID, employeeID string) (*ai.MatchAnalysis, error)
	
	// Skills Analysis
	AnalyzeSkillsGap(ctx context.Context, jobID, employeeID string) (*ai.SkillsGapAnalysis, error)
	
	// Cover Letter Generation
	GenerateCoverLetter(ctx context.Context, jobID, employeeID string) (string, error)
	
	// Interview Preparation
	GenerateInterviewQuestions(ctx context.Context, jobID string) ([]string, error)
	
	// Match Feedback
	SubmitMatchFeedback(ctx context.Context, matchID string, userID string, feedback string, rating int) error
	
	// Batch Processing
	BatchMatchJobs(ctx context.Context, jobID string) error
	BatchMatchCandidates(ctx context.Context, employeeID string) error
}

type MatchResult struct {
	JobID             string    `json:"job_id"`
	EmployeeID        string    `json:"employee_id"`
	JobTitle          string    `json:"job_title"`
	CompanyName       string    `json:"company_name"`
	OverallScore      int       `json:"overall_score"`
	SkillsScore       int       `json:"skills_score"`
	ExperienceScore   int       `json:"experience_score"`
	CultureScore      int       `json:"culture_score"`
	LocationScore     int       `json:"location_score"`
	Summary           string    `json:"summary"`
	Recommendation    string    `json:"recommendation"`
	MatchingSkills    []string  `json:"matching_skills"`
	MissingSkills     []string  `json:"missing_skills"`
	MatchID           string    `json:"match_id"`
	AnalyzedAt        time.Time `json:"analyzed_at"`
	CandidateName     string    `json:"candidate_name"`
	CandidateHeadline string    `json:"candidate_headline"`
	CandidateLocation string    `json:"candidate_location"`
	CandidateAvatar   string    `json:"candidate_avatar"`
	CandidateInitials string    `json:"candidate_initials"`
	ExperienceYears   int       `json:"experience_years"`
	Skills            []string  `json:"skills"`
}

type MatchingServiceImpl struct {
	cfg            *config.Config
	aiProvider     ai.AIProvider
	jobRepo        repository.JobRepository
	userRepo       repository.UserRepository
	githubRepo     repository.GitHubRepository
	matchRepo      repository.MatchRepository
	mu             sync.RWMutex
}

func NewMatchingService(
	cfg *config.Config,
	aiProvider ai.AIProvider,
	jobRepo repository.JobRepository,
	userRepo repository.UserRepository,
	githubRepo repository.GitHubRepository,
	matchRepo repository.MatchRepository,
) MatchingService {
	return &MatchingServiceImpl{
		cfg:        cfg,
		aiProvider: aiProvider,
		jobRepo:    jobRepo,
		userRepo:   userRepo,
		githubRepo: githubRepo,
		matchRepo:  matchRepo,
	}
}

func (s *MatchingServiceImpl) GetJobMatches(ctx context.Context, employeeID string, limit int) ([]*MatchResult, error) {
	employee, githubProfile, err := s.userRepo.GetEmployeeWithGitHub(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee profile: %w", err)
	}
	
	candidateProfile := s.buildCandidateProfile(employee, githubProfile)
	
	jobs, _, err := s.jobRepo.ListActive(ctx, 1, limit*2)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}
	
	var results []*MatchResult
	var wg sync.WaitGroup
	var mu sync.Mutex
	semaphore := make(chan struct{}, 5)
	
	for _, job := range jobs {
		wg.Add(1)
		go func(j *models.Job) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			jobDesc := s.buildJobDescription(j)
			
			analysis, err := s.aiProvider.GenerateMatchAnalysis(ctx, jobDesc, candidateProfile)
			if err != nil {
				return
			}
			
			match := &MatchResult{
				JobID:           j.ID,
				EmployeeID:      employeeID,
				JobTitle:        j.Title,
				CompanyName:     j.Employer.CompanyName,
				OverallScore:    analysis.OverallScore,
				SkillsScore:     analysis.SkillsScore,
				ExperienceScore: analysis.ExperienceScore,
				CultureScore:    analysis.CultureScore,
				LocationScore:   analysis.LocationScore,
				Summary:         analysis.Summary,
				Recommendation:  analysis.Recommendation,
				MatchingSkills:  analysis.MatchingSkills,
				MissingSkills:   analysis.MissingSkills,
				MatchID:         uuid.New().String(),
				AnalyzedAt:      time.Now(),
			}
			
			mu.Lock()
			results = append(results, match)
			mu.Unlock()
			
			s.saveMatch(ctx, j.ID, employeeID, analysis)
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
	
	if len(results) > limit {
		results = results[:limit]
	}
	
	return results, nil
}

func (s *MatchingServiceImpl) GetCandidateMatches(ctx context.Context, jobID string, employerID string, limit int) ([]*MatchResult, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	
	if job.EmployerID != employerID {
		return nil, fmt.Errorf("unauthorized: you don't own this job")
	}
	
	jobDesc := s.buildJobDescription(job)
	
	employees, _, err := s.userRepo.ListActiveEmployees(ctx, 1, limit*2)
	if err != nil {
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}
	
	var results []*MatchResult
	var wg sync.WaitGroup
	var mu sync.Mutex
	semaphore := make(chan struct{}, 5)
	
		for _, employee := range employees {
			wg.Add(1)
			go func(emp *models.EmployeeProfile) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				githubProfile, _ := s.githubRepo.GetProfileByEmployeeID(ctx, emp.UserID)
				candidateProfile := s.buildCandidateProfile(emp, githubProfile)

				analysis, err := s.aiProvider.GenerateMatchAnalysis(ctx, jobDesc, candidateProfile)
				if err != nil {
					return
				}

				initials := ""
				parts := splitName(emp.FullName)
				for _, p := range parts {
					if len(p) > 0 {
						initials += string(p[0])
					}
				}

				match := &MatchResult{
					JobID:             jobID,
					EmployeeID:        emp.UserID,
					OverallScore:      analysis.OverallScore,
					SkillsScore:       analysis.SkillsScore,
					ExperienceScore:   analysis.ExperienceScore,
					CultureScore:      analysis.CultureScore,
					LocationScore:     analysis.LocationScore,
					Summary:           analysis.Summary,
					Recommendation:    analysis.Recommendation,
					MatchingSkills:    analysis.MatchingSkills,
					MissingSkills:     analysis.MissingSkills,
					MatchID:           uuid.New().String(),
					AnalyzedAt:        time.Now(),
					CandidateName:     emp.FullName,
					CandidateHeadline: emp.Headline,
					CandidateLocation: emp.Location,
					CandidateInitials: initials,
					ExperienceYears:   emp.YearsOfExperience,
					Skills:            emp.Skills,
				}
			
			mu.Lock()
			results = append(results, match)
			mu.Unlock()
			
			s.saveMatch(ctx, jobID, emp.UserID, analysis)
		}(employee)
	}
	
	wg.Wait()
	
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].OverallScore > results[i].OverallScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	
	if len(results) > limit {
		results = results[:limit]
	}
	
	return results, nil
}

func (s *MatchingServiceImpl) GetDetailedMatchAnalysis(ctx context.Context, jobID, employeeID string) (*ai.MatchAnalysis, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	
	employee, githubProfile, err := s.userRepo.GetEmployeeWithGitHub(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	
	jobDesc := s.buildJobDescription(job)
	candidateProfile := s.buildCandidateProfile(employee, githubProfile)
	
	return s.aiProvider.GenerateMatchAnalysis(ctx, jobDesc, candidateProfile)
}

func (s *MatchingServiceImpl) AnalyzeSkillsGap(ctx context.Context, jobID, employeeID string) (*ai.SkillsGapAnalysis, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	
	employee, githubProfile, err := s.userRepo.GetEmployeeWithGitHub(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	
	jobDesc := s.buildJobDescription(job)
	candidateProfile := s.buildCandidateProfile(employee, githubProfile)
	
	return s.aiProvider.AnalyzeSkillsGap(ctx, jobDesc, candidateProfile)
}

func (s *MatchingServiceImpl) GenerateCoverLetter(ctx context.Context, jobID, employeeID string) (string, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return "", err
	}
	
	employee, githubProfile, err := s.userRepo.GetEmployeeWithGitHub(ctx, employeeID)
	if err != nil {
		return "", err
	}
	
	jobDesc := s.buildJobDescription(job)
	candidateProfile := s.buildCandidateProfile(employee, githubProfile)
	
	return s.aiProvider.GenerateCoverLetter(ctx, jobDesc, candidateProfile)
}

func (s *MatchingServiceImpl) GenerateInterviewQuestions(ctx context.Context, jobID string) ([]string, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	
	jobDesc := s.buildJobDescription(job)
	
	return s.aiProvider.GenerateInterviewQuestions(ctx, jobDesc)
}

func (s *MatchingServiceImpl) SubmitMatchFeedback(ctx context.Context, matchID string, userID string, feedback string, rating int) error {
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("match not found: %w", err)
	}
	
	matchFeedback := &models.MatchFeedback{
		ID:          uuid.New().String(),
		MatchID:     matchID,
		UserType:    s.getUserTypeFromID(ctx, userID, match),
		Rating:      rating,
		WasAccurate: rating >= 3,
		Feedback:    feedback,
		CreatedAt:   time.Now(),
	}
	
	if err := s.matchRepo.CreateFeedback(ctx, matchFeedback); err != nil {
		return fmt.Errorf("failed to save feedback: %w", err)
	}
	
	if rating < 3 {
		s.matchRepo.UpdateMatchScore(ctx, matchID, match.OverallScore-10)
	} else if rating > 4 {
		s.matchRepo.UpdateMatchScore(ctx, matchID, match.OverallScore+5)
	}
	
	return nil
}

func (s *MatchingServiceImpl) BatchMatchJobs(ctx context.Context, jobID string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	
	employees, total, err := s.userRepo.ListActiveEmployees(ctx, 1, 100)
	if err != nil {
		return err
	}
	
	jobDesc := s.buildJobDescription(job)
	
	for _, employee := range employees {
		githubProfile, _ := s.githubRepo.GetProfileByEmployeeID(ctx, employee.UserID)
		candidateProfile := s.buildCandidateProfile(employee, githubProfile)
		
		analysis, err := s.aiProvider.GenerateMatchAnalysis(ctx, jobDesc, candidateProfile)
		if err != nil {
			continue
		}
		
		s.saveMatch(ctx, jobID, employee.UserID, analysis)
	}
	
	fmt.Printf("Batch matching completed for job %s: processed %d candidates\n", jobID, total)
	return nil
}

func (s *MatchingServiceImpl) BatchMatchCandidates(ctx context.Context, employeeID string) error {
	employee, githubProfile, err := s.userRepo.GetEmployeeWithGitHub(ctx, employeeID)
	if err != nil {
		return err
	}
	
	jobs, total, err := s.jobRepo.ListActive(ctx, 1, 100)
	if err != nil {
		return err
	}
	
	candidateProfile := s.buildCandidateProfile(employee, githubProfile)
	
	for _, job := range jobs {
		jobDesc := s.buildJobDescription(job)
		
		analysis, err := s.aiProvider.GenerateMatchAnalysis(ctx, jobDesc, candidateProfile)
		if err != nil {
			continue
		}
		
		s.saveMatch(ctx, job.ID, employeeID, analysis)
	}
	
	fmt.Printf("Batch matching completed for employee %s: processed %d jobs\n", employeeID, total)
	return nil
}

// Helper methods

func (s *MatchingServiceImpl) buildJobDescription(job *models.Job) ai.JobDescription {
	return ai.JobDescription{
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
		EducationLevel:   job.EducationLevel,
		EmploymentType:   job.EmploymentType,
		Location:         job.Location,
		IsRemote:         job.IsRemote,
	}
}

func (s *MatchingServiceImpl) buildCandidateProfile(employee *models.EmployeeProfile, githubProfile *models.GithubProfile) ai.CandidateProfile {
	profile := ai.CandidateProfile{
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
	}
	
	if githubProfile != nil {
		profile.GithubActivity = &ai.GithubActivity{
			PublicRepos:        githubProfile.PublicRepos,
			TotalCommits:       githubProfile.TotalCommits,
			Followers:          githubProfile.Followers,
			ContributionStreak: githubProfile.ContributionStreak,
			TopLanguages:       githubProfile.GetTopLanguages(5),
			AccountAgeDays:     githubProfile.AccountAgeDays,
			ActivityScore:      githubProfile.ActivityScore,
			QualityScore:       githubProfile.QualityScore,
		}
	}
	
	return profile
}

func (s *MatchingServiceImpl) saveMatch(ctx context.Context, jobID, employeeID string, analysis *ai.MatchAnalysis) error {
	existingMatch, _ := s.matchRepo.GetByJobAndEmployee(ctx, jobID, employeeID)
	if existingMatch != nil {
		existingMatch.OverallScore = analysis.OverallScore
		existingMatch.SkillsScore = analysis.SkillsScore
		existingMatch.ExperienceScore = analysis.ExperienceScore
		existingMatch.LocationScore = analysis.LocationScore
		existingMatch.CultureScore = analysis.CultureScore
		existingMatch.MatchingSkills = analysis.MatchingSkills
		existingMatch.MissingSkills = analysis.MissingSkills
		existingMatch.MatchReason = analysis.Summary
		return s.matchRepo.Update(ctx, existingMatch)
	}
	
	match := &models.Match{
		ID:              uuid.New().String(),
		JobID:           jobID,
		EmployeeID:      employeeID,
		OverallScore:    analysis.OverallScore,
		SkillsScore:     analysis.SkillsScore,
		ExperienceScore: analysis.ExperienceScore,
		LocationScore:   analysis.LocationScore,
		CultureScore:    analysis.CultureScore,
		MatchingSkills:  analysis.MatchingSkills,
		MissingSkills:   analysis.MissingSkills,
		MatchReason:     analysis.Summary,
		CreatedAt:       time.Now(),
		ExpiresAt:       time.Now().AddDate(0, 1, 0),
	}
	
	return s.matchRepo.Create(ctx, match)
}

func (s *MatchingServiceImpl) getUserTypeFromID(ctx context.Context, userID string, match *models.Match) string {
	if match.EmployeeID == userID {
		return "employee"
	}
	return "employer"
}

func splitName(name string) []string {
	var result []string
	current := ""
	for _, r := range name {
		if r == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}