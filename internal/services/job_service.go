package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type JobService interface {
	// Job CRUD
	CreateJob(ctx context.Context, employerID string, req *CreateJobRequest) (*models.Job, error)
	GetJob(ctx context.Context, jobID string) (*models.Job, error)
	UpdateJob(ctx context.Context, jobID string, employerID string, req *UpdateJobRequest) (*models.Job, error)
	DeleteJob(ctx context.Context, jobID string, employerID string) error
	CloseJob(ctx context.Context, jobID string, employerID string) error
	
	// Job listing
	ListJobs(ctx context.Context, filters *JobFilters, page, limit int) (*JobListResponse, error)
	ListMyJobs(ctx context.Context, employerID string, page, limit int) (*JobListResponse, error)
	GetFeaturedJobs(ctx context.Context, limit int) ([]*models.Job, error)
	GetRecentJobs(ctx context.Context, limit int) ([]*models.Job, error)
	GetSimilarJobs(ctx context.Context, jobID string, limit int) ([]*models.Job, error)
	
	// Search
	SearchJobs(ctx context.Context, query string, filters *JobFilters, page, limit int) (*JobListResponse, error)
	
	// Saved jobs
	SaveJob(ctx context.Context, employeeID, jobID string) error
	UnsaveJob(ctx context.Context, employeeID, jobID string) error
	GetSavedJobs(ctx context.Context, employeeID string, page, limit int) (*JobListResponse, error)
	IsJobSaved(ctx context.Context, employeeID, jobID string) (bool, error)
	
	// Statistics
	GetJobStats(ctx context.Context, employerID string) (*repository.JobStats, error)
	IncrementJobViews(ctx context.Context, jobID string) error
	
	// Validation
	ValidateJobPosting(ctx context.Context, employerID string) error
}

type CreateJobRequest struct {
	Title             string   `json:"title" validate:"required,min=5,max=255"`
	Description       string   `json:"description" validate:"required,min=20"`
	Requirements      string   `json:"requirements" validate:"required,min=20"`
	Responsibilities  string   `json:"responsibilities"`
	Benefits          []string `json:"benefits"`
	
	RequiredSkills    []string `json:"required_skills" validate:"required,min=1"`
	NiceToHaveSkills  []string `json:"nice_to_have_skills"`
	
	ExperienceLevel   string   `json:"experience_level"`
	MinExperience     int      `json:"min_experience"`
	MaxExperience     int      `json:"max_experience"`
	EducationLevel    string   `json:"education_level"`
	
	SalaryMin         int      `json:"salary_min"`
	SalaryMax         int      `json:"salary_max"`
	SalaryCurrency    string   `json:"salary_currency"`
	IsSalaryVisible   bool     `json:"is_salary_visible"`
	
	Location          string   `json:"location" validate:"required"`
	IsRemote          bool     `json:"is_remote"`
	IsHybrid          bool     `json:"is_hybrid"`
	RemotePolicy      string   `json:"remote_policy"`
	
	EmploymentType    string   `json:"employment_type" validate:"oneof=full-time part-time contract internship freelance"`
	WorkHours         string   `json:"work_hours"`
	
	IsFeatured        bool     `json:"is_featured"`
	IsUrgent          bool     `json:"is_urgent"`
	ExpiresAt         string   `json:"expires_at"`
}

type UpdateJobRequest struct {
	Title             *string   `json:"title"`
	Description       *string   `json:"description"`
	Requirements      *string   `json:"requirements"`
	Responsibilities  *string   `json:"responsibilities"`
	Benefits          *[]string `json:"benefits"`
	
	RequiredSkills    []string `json:"required_skills"`
	NiceToHaveSkills  []string `json:"nice_to_have_skills"`
	
	ExperienceLevel   *string   `json:"experience_level"`
	MinExperience     *int      `json:"min_experience"`
	MaxExperience     *int      `json:"max_experience"`
	EducationLevel    *string   `json:"education_level"`
	
	SalaryMin         *int      `json:"salary_min"`
	SalaryMax         *int      `json:"salary_max"`
	SalaryCurrency    *string   `json:"salary_currency"`
	IsSalaryVisible   *bool     `json:"is_salary_visible"`
	
	Location          *string   `json:"location"`
	IsRemote          *bool     `json:"is_remote"`
	IsHybrid          *bool     `json:"is_hybrid"`
	RemotePolicy      *string   `json:"remote_policy"`
	
	EmploymentType    *string   `json:"employment_type"`
	WorkHours         *string   `json:"work_hours"`
	
	IsFeatured        *bool   `json:"is_featured"`
	IsUrgent          *bool   `json:"is_urgent"`
	ExpiresAt         string  `json:"expires_at"`
}

type JobFilters struct {
	Title           string   `json:"title"`
	Location        string   `json:"location"`
	IsRemote        *bool    `json:"is_remote"`
	IsHybrid        *bool    `json:"is_hybrid"`
	EmploymentType  string   `json:"employment_type"`
	ExperienceLevel string   `json:"experience_level"`
	MinSalary       int      `json:"min_salary"`
	MaxSalary       int      `json:"max_salary"`
	Skills          []string `json:"skills"`
	Industry        string   `json:"industry"`
	CompanyName     string   `json:"company_name"`
	PostedAfter     time.Time `json:"posted_after"`
	PostedBefore    time.Time `json:"posted_before"`
}

type JobListResponse struct {
	Jobs       []*models.Job `json:"jobs"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

type JobServiceImpl struct {
	cfg      *config.Config
	jobRepo  repository.JobRepository
	userRepo repository.UserRepository
}

func NewJobService(cfg *config.Config, jobRepo repository.JobRepository, userRepo repository.UserRepository) JobService {
	return &JobServiceImpl{
		cfg:      cfg,
		jobRepo:  jobRepo,
		userRepo: userRepo,
	}
}

func (s *JobServiceImpl) CreateJob(ctx context.Context, employerID string, req *CreateJobRequest) (*models.Job, error) {
	// Validate employer can post job
	if err := s.ValidateJobPosting(ctx, employerID); err != nil {
		return nil, err
	}
	
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}
	
	// Set default values
	if req.SalaryCurrency == "" {
		req.SalaryCurrency = "KES"
	}
	if req.EmploymentType == "" {
		req.EmploymentType = "full-time"
	}
	if req.ExperienceLevel == "" || req.ExperienceLevel == "any" {
		req.ExperienceLevel = "mid"
	}
	
	// Set expiration date (default 30 days)
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		parsed, err := time.Parse("2006-01-02", req.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at format, expected YYYY-MM-DD")
		}
		expiresAt = &parsed
	}
	if expiresAt == nil {
		defaultExpiry := time.Now().AddDate(0, 1, 0)
		expiresAt = &defaultExpiry
	}
	
	job := &models.Job{
		ID:                uuid.New().String(),
		EmployerID:        employerID,
		Title:             req.Title,
		Description:       req.Description,
		Requirements:      req.Requirements,
		Responsibilities:  req.Responsibilities,
		Benefits:          req.Benefits,
		RequiredSkills:    req.RequiredSkills,
		NiceToHaveSkills:  req.NiceToHaveSkills,
		ExperienceLevel:   req.ExperienceLevel,
		MinExperience:     req.MinExperience,
		MaxExperience:     req.MaxExperience,
		EducationLevel:    req.EducationLevel,
		SalaryMin:         req.SalaryMin,
		SalaryMax:         req.SalaryMax,
		SalaryCurrency:    req.SalaryCurrency,
		IsSalaryVisible:   req.IsSalaryVisible,
		Location:          req.Location,
		IsRemote:          req.IsRemote,
		IsHybrid:          req.IsHybrid,
		RemotePolicy:      req.RemotePolicy,
		EmploymentType:    req.EmploymentType,
		WorkHours:         req.WorkHours,
		IsActive:          true,
		IsFeatured:        req.IsFeatured,
		IsUrgent:          req.IsUrgent,
		ExpiresAt:         expiresAt,
		PostedAt:          time.Now(),
	}
	
	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}
	
	// Update employer's job count
	s.userRepo.IncrementJobPostedCount(ctx, employerID)
	
	return job, nil
}

func (s *JobServiceImpl) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	job, err := s.jobRepo.GetByIDWithEmployer(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, errors.New("job not found")
	}
	
	// Check if job is expired
	if job.ExpiresAt != nil && job.ExpiresAt.Before(time.Now()) && job.IsActive {
		// Auto-close expired jobs
		s.jobRepo.UpdateStatus(ctx, jobID, false)
		job.IsActive = false
	}
	
	return job, nil
}

func (s *JobServiceImpl) UpdateJob(ctx context.Context, jobID string, employerID string, req *UpdateJobRequest) (*models.Job, error) {
	// Get existing job
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, errors.New("job not found")
	}
	
	// Verify ownership
	if job.EmployerID != employerID {
		return nil, errors.New("unauthorized: you don't own this job")
	}
	
	// Update fields
	if req.Title != nil {
		job.Title = *req.Title
	}
	if req.Description != nil {
		job.Description = *req.Description
	}
	if req.Requirements != nil {
		job.Requirements = *req.Requirements
	}
	if req.Responsibilities != nil {
		job.Responsibilities = *req.Responsibilities
	}
	if req.Benefits != nil {
		job.Benefits = *req.Benefits
	}
	if len(req.RequiredSkills) > 0 {
		job.RequiredSkills = req.RequiredSkills
	}
	if len(req.NiceToHaveSkills) > 0 {
		job.NiceToHaveSkills = req.NiceToHaveSkills
	}
	if req.ExperienceLevel != nil {
		job.ExperienceLevel = *req.ExperienceLevel
	}
	if req.MinExperience != nil {
		job.MinExperience = *req.MinExperience
	}
	if req.MaxExperience != nil {
		job.MaxExperience = *req.MaxExperience
	}
	if req.EducationLevel != nil {
		job.EducationLevel = *req.EducationLevel
	}
	if req.SalaryMin != nil {
		job.SalaryMin = *req.SalaryMin
	}
	if req.SalaryMax != nil {
		job.SalaryMax = *req.SalaryMax
	}
	if req.SalaryCurrency != nil {
		job.SalaryCurrency = *req.SalaryCurrency
	}
	if req.IsSalaryVisible != nil {
		job.IsSalaryVisible = *req.IsSalaryVisible
	}
	if req.Location != nil {
		job.Location = *req.Location
	}
	if req.IsRemote != nil {
		job.IsRemote = *req.IsRemote
	}
	if req.IsHybrid != nil {
		job.IsHybrid = *req.IsHybrid
	}
	if req.RemotePolicy != nil {
		job.RemotePolicy = *req.RemotePolicy
	}
	if req.EmploymentType != nil {
		job.EmploymentType = *req.EmploymentType
	}
	if req.WorkHours != nil {
		job.WorkHours = *req.WorkHours
	}
	if req.IsFeatured != nil {
		job.IsFeatured = *req.IsFeatured
	}
	if req.IsUrgent != nil {
		job.IsUrgent = *req.IsUrgent
	}
	if req.ExpiresAt != "" {
		parsed, err := time.Parse("2006-01-02", req.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at format, expected YYYY-MM-DD")
		}
		job.ExpiresAt = &parsed
	}
	
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to update job: %w", err)
	}
	
	return job, nil
}

func (s *JobServiceImpl) DeleteJob(ctx context.Context, jobID string, employerID string) error {
	// Get existing job
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("job not found")
	}
	
	// Verify ownership
	if job.EmployerID != employerID {
		return errors.New("unauthorized: you don't own this job")
	}
	
	return s.jobRepo.Delete(ctx, jobID)
}

func (s *JobServiceImpl) CloseJob(ctx context.Context, jobID string, employerID string) error {
	// Get existing job
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("job not found")
	}
	
	// Verify ownership
	if job.EmployerID != employerID {
		return errors.New("unauthorized: you don't own this job")
	}
	
	return s.jobRepo.UpdateStatus(ctx, jobID, false)
}

func (s *JobServiceImpl) ListJobs(ctx context.Context, filters *JobFilters, page, limit int) (*JobListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 100 {
		limit = 100
	}
	
	repoFilters := s.convertToRepoFilters(filters)
	jobs, total, err := s.jobRepo.List(ctx, repoFilters, page, limit)
	if err != nil {
		return nil, err
	}
	
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	
	return &JobListResponse{
		Jobs:       jobs,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *JobServiceImpl) ListMyJobs(ctx context.Context, employerID string, page, limit int) (*JobListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 100 {
		limit = 100
	}
	
	jobs, total, err := s.jobRepo.ListByEmployer(ctx, employerID, page, limit)
	if err != nil {
		return nil, err
	}
	
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	
	return &JobListResponse{
		Jobs:       jobs,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *JobServiceImpl) GetFeaturedJobs(ctx context.Context, limit int) ([]*models.Job, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	return s.jobRepo.ListFeatured(ctx, limit)
}

func (s *JobServiceImpl) GetRecentJobs(ctx context.Context, limit int) ([]*models.Job, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	return s.jobRepo.ListRecent(ctx, limit)
}

func (s *JobServiceImpl) GetSimilarJobs(ctx context.Context, jobID string, limit int) ([]*models.Job, error) {
	// Get the job to extract skills
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil || job == nil {
		return []*models.Job{}, nil
	}
	
	// Find jobs with similar skills
	filters := &JobFilters{
		Skills: job.RequiredSkills,
	}
	
	repoFilters := s.convertToRepoFilters(filters)
	jobs, _, err := s.jobRepo.List(ctx, repoFilters, 1, limit)
	if err != nil {
		return []*models.Job{}, nil
	}
	
	// Filter out the original job
	var similarJobs []*models.Job
	for _, j := range jobs {
		if j.ID != jobID {
			similarJobs = append(similarJobs, j)
		}
	}
	
	if len(similarJobs) > limit {
		similarJobs = similarJobs[:limit]
	}
	
	return similarJobs, nil
}

func (s *JobServiceImpl) SearchJobs(ctx context.Context, query string, filters *JobFilters, page, limit int) (*JobListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 100 {
		limit = 100
	}
	
	repoFilters := s.convertToRepoFilters(filters)
	jobs, total, err := s.jobRepo.Search(ctx, query, repoFilters, page, limit)
	if err != nil {
		return nil, err
	}
	
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	
	return &JobListResponse{
		Jobs:       jobs,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *JobServiceImpl) SaveJob(ctx context.Context, employeeID, jobID string) error {
	// Check if job exists and is active
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("job not found")
	}
	if !job.IsActive {
		return errors.New("cannot save inactive job")
	}
	
	// Check if already saved
	isSaved, err := s.jobRepo.IsJobSaved(ctx, employeeID, jobID)
	if err != nil {
		return err
	}
	if isSaved {
		return errors.New("job already saved")
	}
	
	return s.jobRepo.SaveJob(ctx, employeeID, jobID)
}

func (s *JobServiceImpl) UnsaveJob(ctx context.Context, employeeID, jobID string) error {
	return s.jobRepo.UnsaveJob(ctx, employeeID, jobID)
}

func (s *JobServiceImpl) GetSavedJobs(ctx context.Context, employeeID string, page, limit int) (*JobListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 100 {
		limit = 100
	}
	
	jobs, total, err := s.jobRepo.GetSavedJobs(ctx, employeeID, page, limit)
	if err != nil {
		return nil, err
	}
	
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	
	return &JobListResponse{
		Jobs:       jobs,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *JobServiceImpl) IsJobSaved(ctx context.Context, employeeID, jobID string) (bool, error) {
	return s.jobRepo.IsJobSaved(ctx, employeeID, jobID)
}

func (s *JobServiceImpl) GetJobStats(ctx context.Context, employerID string) (*repository.JobStats, error) {
	return s.jobRepo.GetStatsByEmployer(ctx, employerID)
}

func (s *JobServiceImpl) IncrementJobViews(ctx context.Context, jobID string) error {
	return s.jobRepo.IncrementViews(ctx, jobID)
}

func (s *JobServiceImpl) ValidateJobPosting(ctx context.Context, employerID string) error {
	// Get employer profile
	profile, err := s.userRepo.GetEmployerProfile(ctx, employerID)
	if err != nil {
		return errors.New("employer profile not found")
	}
	
	// Check if employer can post job
	if !profile.CanPostJob() {
		if profile.SubscriptionPlan == "free" && profile.TotalJobsPosted >= s.cfg.MaxJobPosts {
			return errors.New("free plan limit reached. Upgrade to post more jobs")
		}
		return errors.New("account not verified or inactive")
	}
	
	return nil
}

// Private helper methods

func (s *JobServiceImpl) validateCreateRequest(req *CreateJobRequest) error {
	if req.Title == "" {
		return errors.New("job title is required")
	}
	if len(req.Title) < 5 {
		return errors.New("job title must be at least 5 characters")
	}
	
	if req.Description == "" {
		return errors.New("job description is required")
	}
	if len(req.Description) < 20 {
		return errors.New("job description must be at least 50 characters")
	}
	
	if req.Requirements == "" {
		return errors.New("job requirements are required")
	}
	if len(req.Requirements) < 20 {
		return errors.New("job requirements must be at least 50 characters")
	}
	
	if len(req.RequiredSkills) == 0 {
		return errors.New("at least one required skill is needed")
	}
	
	if req.Location == "" {
		return errors.New("job location is required")
	}
	
	if req.SalaryMin > 0 && req.SalaryMax > 0 && req.SalaryMin > req.SalaryMax {
		return errors.New("minimum salary cannot be greater than maximum salary")
	}
	
	if req.ExpiresAt != "" {
		parsed, err := time.Parse("2006-01-02", req.ExpiresAt)
		if err != nil {
			return errors.New("invalid expires_at format, expected YYYY-MM-DD")
		}
		if parsed.Before(time.Now().Truncate(24 * time.Hour)) {
			return errors.New("expiration date cannot be in the past")
		}
	}
	
	return nil
}

func (s *JobServiceImpl) convertToRepoFilters(filters *JobFilters) repository.JobFilters {
	repoFilters := repository.JobFilters{}
	
	if filters == nil {
		return repoFilters
	}
	
	repoFilters.Title = filters.Title
	repoFilters.Location = filters.Location
	repoFilters.IsRemote = filters.IsRemote
	repoFilters.IsHybrid = filters.IsHybrid
	repoFilters.EmploymentType = filters.EmploymentType
	repoFilters.ExperienceLevel = filters.ExperienceLevel
	repoFilters.MinSalary = filters.MinSalary
	repoFilters.MaxSalary = filters.MaxSalary
	repoFilters.Skills = filters.Skills
	repoFilters.PostedAfter = filters.PostedAfter
	repoFilters.PostedBefore = filters.PostedBefore
	
	// Map industry to employer industry
	if filters.Industry != "" {
		repoFilters.Industry = filters.Industry
	}
	
	return repoFilters
}