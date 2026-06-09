package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/parser"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type ProfileService interface {
	ParseAndUpdateProfile(ctx context.Context, userID string, file multipart.File, filename string) (*ProfileCompletionResponse, error)
	GetProfileCompletion(ctx context.Context, userID string) (*ProfileCompletion, error)
	GetSuggestedSkills(ctx context.Context, userID string) ([]string, error)
	GetProfileWizard(ctx context.Context, userID string) (*ProfileWizard, error)
	CompleteWizardStep(ctx context.Context, userID string, step int, data map[string]interface{}) error
	UploadAvatar(ctx context.Context, userID string, file multipart.File, filename string) (string, error)
}

type ProfileCompletion struct {
	Percentage      int      `json:"percentage"`
	CompletedSteps  []string `json:"completed_steps"`
	RemainingSteps  []string `json:"remaining_steps"`
	MissingFields   []string `json:"missing_fields"`
	Score           int      `json:"score"` // 0-100
}

type ProfileCompletionResponse struct {
	Success     bool               `json:"success"`
	Message     string             `json:"message"`
	Profile     *models.EmployeeProfile `json:"profile"`
	Completion  *ProfileCompletion `json:"completion"`
	NewSkills   []string           `json:"new_skills"`
}

type ProfileWizard struct {
	CurrentStep   int                    `json:"current_step"`
	TotalSteps    int                    `json:"total_steps"`
	Steps         []WizardStep           `json:"steps"`
	Profile       *models.EmployeeProfile `json:"profile"`
	Completion    *ProfileCompletion     `json:"completion"`
}

type WizardStep struct {
	Number      int         `json:"number"`
	Name        string      `json:"name"`
	Completed   bool        `json:"completed"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Data        interface{} `json:"data,omitempty"`
}

type ProfileServiceImpl struct {
	cfg         *config.Config
	userRepo    repository.UserRepository
	githubRepo  repository.GitHubRepository
	pdfParser   *parser.PDFParser
}

func NewProfileService(
	cfg *config.Config,
	userRepo repository.UserRepository,
	githubRepo repository.GitHubRepository,
) ProfileService {
	return &ProfileServiceImpl{
		cfg:        cfg,
		userRepo:   userRepo,
		githubRepo: githubRepo,
		pdfParser:  parser.NewPDFParser(),
	}
}

func (s *ProfileServiceImpl) ParseAndUpdateProfile(ctx context.Context, userID string, file multipart.File, filename string) (*ProfileCompletionResponse, error) {
	// Parse resume
	parsed, err := s.pdfParser.Parse(ctx, file, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resume: %w", err)
	}
	
	// Get existing profile
	profile, err := s.userRepo.GetEmployeeProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	
	newSkills := []string{}
	updates := make(map[string]interface{})
	
	// Update name if empty
	if profile.FullName == "" && parsed.FullName != "" {
		updates["full_name"] = parsed.FullName
	}
	
	// Update location if empty
	if profile.Location == "" && parsed.Location != "" {
		updates["location"] = parsed.Location
	}
	
	// Update bio if empty
	if profile.Bio == "" && parsed.Summary != "" {
		updates["bio"] = parsed.Summary
	}
	
	// Merge skills
	existingSkills := make(map[string]bool)
	for _, skill := range profile.Skills {
		existingSkills[strings.ToLower(skill)] = true
	}
	
	for _, skill := range parsed.Skills {
		if !existingSkills[strings.ToLower(skill)] {
			newSkills = append(newSkills, skill)
			existingSkills[strings.ToLower(skill)] = true
		}
	}
	
	if len(newSkills) > 0 {
		allSkills := append(profile.Skills, newSkills...)
		updates["skills"] = allSkills
	}
	
	// Update LinkedIn URL if empty
	if profile.LinkedInURL == "" && parsed.LinkedInURL != "" {
		updates["linkedin_url"] = parsed.LinkedInURL
	}
	
	// Update portfolio URL if empty
	if profile.PortfolioURL == "" && parsed.PortfolioURL != "" {
		updates["portfolio_url"] = parsed.PortfolioURL
	}
	
	// Update experience level based on total years
	if profile.ExperienceLevel == "" && parsed.TotalExperience > 0 {
		experienceLevel := "entry"
		if parsed.TotalExperience >= 7 {
			experienceLevel = "senior"
		} else if parsed.TotalExperience >= 3 {
			experienceLevel = "mid"
		} else if parsed.TotalExperience >= 1 {
			experienceLevel = "junior"
		}
		updates["experience_level"] = experienceLevel
		
		if profile.YearsOfExperience == 0 {
			updates["years_of_experience"] = parsed.TotalExperience
		}
	}
	
	// Apply updates
	if len(updates) > 0 {
		if err := s.userRepo.UpdateEmployeeProfile(ctx, userID, updates); err != nil {
			return nil, fmt.Errorf("failed to update profile: %w", err)
		}
	}
	
	// Get updated profile
	updatedProfile, _ := s.userRepo.GetEmployeeProfile(ctx, userID)
	
	// Calculate completion
	completion := s.calculateCompletion(updatedProfile)
	
	return &ProfileCompletionResponse{
		Success:    true,
		Message:    fmt.Sprintf("Resume parsed successfully! Added %d new skills.", len(newSkills)),
		Profile:    updatedProfile,
		Completion: completion,
		NewSkills:  newSkills,
	}, nil
}

func (s *ProfileServiceImpl) GetProfileCompletion(ctx context.Context, userID string) (*ProfileCompletion, error) {
	profile, err := s.userRepo.GetEmployeeProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	return s.calculateCompletion(profile), nil
}

func (s *ProfileServiceImpl) calculateCompletion(profile *models.EmployeeProfile) *ProfileCompletion {
	completion := &ProfileCompletion{
		CompletedSteps: []string{},
		RemainingSteps: []string{},
		MissingFields:  []string{},
	}
	
	score := 0
	totalFields := 0
	
	// Check full name
	totalFields++
	if profile.FullName != "" && len(profile.FullName) > 2 {
		score += 10
		completion.CompletedSteps = append(completion.CompletedSteps, "basic_info")
	} else {
		completion.RemainingSteps = append(completion.RemainingSteps, "basic_info")
		completion.MissingFields = append(completion.MissingFields, "full_name")
	}
	
	// Check bio/headline
	totalFields++
	if profile.Headline != "" {
		score += 10
	} else {
		completion.MissingFields = append(completion.MissingFields, "headline")
	}
	
	// Check skills (at least 3)
	totalFields++
	if len(profile.Skills) >= 3 {
		score += 15
		if len(profile.Skills) >= 5 {
			score += 5
		}
	} else {
		completion.MissingFields = append(completion.MissingFields, "skills")
	}
	
	// Check experience
	totalFields++
	if profile.YearsOfExperience > 0 {
		score += 15
	} else {
		completion.MissingFields = append(completion.MissingFields, "experience")
	}
	
	// Check location
	totalFields++
	if profile.Location != "" {
		score += 10
	} else {
		completion.MissingFields = append(completion.MissingFields, "location")
	}
	
	// Check GitHub connection
	totalFields++
	if profile.GithubConnected {
		score += 15
		completion.CompletedSteps = append(completion.CompletedSteps, "github")
	} else {
		completion.RemainingSteps = append(completion.RemainingSteps, "github")
		completion.MissingFields = append(completion.MissingFields, "github_connection")
	}
	
	// Check portfolio/LinkedIn
	totalFields++
	if profile.PortfolioURL != "" || profile.LinkedInURL != "" {
		score += 10
	} else {
		completion.MissingFields = append(completion.MissingFields, "portfolio")
	}
	
	// Check resume
	totalFields++
	if profile.ResumeURL != "" {
		score += 10
	} else {
		completion.MissingFields = append(completion.MissingFields, "resume")
	}
	
	completion.Percentage = (score * 100) / 100
	completion.Score = score
	
	return completion
}

func (s *ProfileServiceImpl) GetSuggestedSkills(ctx context.Context, userID string) ([]string, error) {
	profile, err := s.userRepo.GetEmployeeProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Get GitHub languages
	githubProfile, _ := s.githubRepo.GetProfileByEmployeeID(ctx, userID)
	
	suggestedSkills := make(map[string]bool)
	
	// Suggest based on current skills
	skillCategories := map[string][]string{
		"go":      {"microservices", "grpc", "docker", "kubernetes"},
		"python":  {"django", "flask", "fastapi", "pandas", "data analysis"},
		"react":   {"next.js", "typescript", "redux", "tailwind"},
		"javascript": {"typescript", "node.js", "react", "vue"},
		"aws":     {"terraform", "kubernetes", "serverless", "cloudformation"},
	}
	
	for _, skill := range profile.Skills {
		skillLower := strings.ToLower(skill)
		if suggestions, ok := skillCategories[skillLower]; ok {
			for _, sugg := range suggestions {
				suggestedSkills[sugg] = true
			}
		}
	}
	
	// Suggest based on GitHub activity
	if githubProfile != nil && len(githubProfile.GetTopLanguages(3)) > 0 {
		topLang := githubProfile.GetTopLanguages(1)
		if len(topLang) > 0 {
			langMap := map[string][]string{
				"go":      {"gin", "fiber", "echo", "gorm"},
				"python":  {"django", "flask", "fastapi"},
				"javascript": {"react", "vue", "node.js"},
				"typescript": {"react", "next.js", "nestjs"},
				"java":    {"spring boot", "hibernate"},
				"php":     {"laravel", "symfony"},
			}
			
			if suggestions, ok := langMap[strings.ToLower(topLang[0])]; ok {
				for _, sugg := range suggestions {
					suggestedSkills[sugg] = true
				}
			}
		}
	}
	
	// Remove skills the user already has
	for _, skill := range profile.Skills {
		delete(suggestedSkills, strings.ToLower(skill))
	}
	
	result := make([]string, 0, len(suggestedSkills))
	for skill := range suggestedSkills {
		result = append(result, skill)
	}
	
	if len(result) > 10 {
		result = result[:10]
	}
	
	return result, nil
}

func (s *ProfileServiceImpl) GetProfileWizard(ctx context.Context, userID string) (*ProfileWizard, error) {
	profile, err := s.userRepo.GetEmployeeProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	completion := s.calculateCompletion(profile)
	
	steps := []WizardStep{
		{
			Number:      1,
			Name:        "Basic Information",
			Completed:   profile.FullName != "" && profile.Headline != "",
			Required:    true,
			Description: "Tell us about yourself",
		},
		{
			Number:      2,
			Name:        "Skills & Experience",
			Completed:   len(profile.Skills) >= 3 && profile.YearsOfExperience > 0,
			Required:    true,
			Description: "Add your technical skills and work history",
		},
		{
			Number:      3,
			Name:        "Location & Availability",
			Completed:   profile.Location != "" && (profile.IsRemoteOnly || !profile.IsRemoteOnly),
			Required:    true,
			Description: "Where do you want to work?",
		},
		{
			Number:      4,
			Name:        "GitHub Integration",
			Completed:   profile.GithubConnected,
			Required:    false,
			Description: "Connect GitHub to showcase your work",
		},
		{
			Number:      5,
			Name:        "Portfolio & Links",
			Completed:   profile.PortfolioURL != "" || profile.LinkedInURL != "",
			Required:    false,
			Description: "Share your work and professional profiles",
		},
		{
			Number:      6,
			Name:        "Resume Upload",
			Completed:   profile.ResumeURL != "",
			Required:    true,
			Description: "Upload your CV for employers",
		},
	}
	
	// Determine current step
	currentStep := 1
	for i, step := range steps {
		if !step.Completed {
			currentStep = step.Number
			break
		}
		if i == len(steps)-1 {
			currentStep = len(steps)
		}
	}
	
	return &ProfileWizard{
		CurrentStep: currentStep,
		TotalSteps:  len(steps),
		Steps:       steps,
		Profile:     profile,
		Completion:  completion,
	}, nil
}

func (s *ProfileServiceImpl) CompleteWizardStep(ctx context.Context, userID string, step int, data map[string]interface{}) error {
	return nil
}

func (s *ProfileServiceImpl) UploadAvatar(ctx context.Context, userID string, file multipart.File, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowedExts[ext] {
		return "", fmt.Errorf("invalid file type. Allowed: JPG, PNG, WEBP")
	}

	newFilename := fmt.Sprintf("avatar_%s_%d%s", userID, time.Now().Unix(), ext)
	uploadDir := filepath.Join(s.cfg.UploadDir, "avatars")
	filePath := filepath.Join(uploadDir, newFilename)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	avatarURL := fmt.Sprintf("%s/uploads/avatars/%s", s.cfg.AppURL, newFilename)

	if err := s.userRepo.UpdateAvatar(ctx, userID, avatarURL); err != nil {
		return "", fmt.Errorf("failed to update avatar: %w", err)
	}

	return avatarURL, nil
}