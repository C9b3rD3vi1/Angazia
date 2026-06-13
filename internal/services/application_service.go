package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type ApplicationService interface {
	Apply(ctx context.Context, employeeID string, req *ApplyRequest) (*models.Application, error)
	WithdrawApplication(ctx context.Context, applicationID string, employeeID string) error
	GetApplication(ctx context.Context, applicationID string, userID string, role string) (*models.Application, error)
	ListJobApplications(ctx context.Context, jobID string, employerID string, status string, page, limit int) (*ApplicationListResponse, error)
	ListMyApplications(ctx context.Context, employeeID string, page, limit int) (*ApplicationListResponse, error)
	ListCompanyApplications(ctx context.Context, employerID string, status string, page, limit int) (*ApplicationListResponse, error)
	ShortlistApplication(ctx context.Context, applicationID string, employerID string, notes string) error
	RejectApplication(ctx context.Context, applicationID string, employerID string, notes string) error
	SaveNotes(ctx context.Context, applicationID string, employerID string, notes string) error
	ScheduleInterview(ctx context.Context, applicationID string, employerID string, interviewDate time.Time, interviewType string) error
	MarkAsHired(ctx context.Context, applicationID string, employerID string) error
	GetApplicationStats(ctx context.Context, userID string, role string) (*repository.ApplicationStats, error)
	GetJobApplicationStats(ctx context.Context, jobID string, employerID string) (*JobApplicationStats, error)
	BulkShortlist(ctx context.Context, applicationIDs []string, employerID string) error
	BulkReject(ctx context.Context, applicationIDs []string, employerID string) error
	SetNotificationService(ns NotificationService)
}

type ApplyRequest struct {
	JobID       string `json:"job_id" validate:"required"`
	CoverLetter string `json:"cover_letter" validate:"required,min=50"`
	ResumeURL   string `json:"resume_url"`
	PortfolioURL string `json:"portfolio_url"`
}

type ApplicationListResponse struct {
	Applications []*models.Application `json:"applications"`
	Total        int64                  `json:"total"`
	Page         int                    `json:"page"`
	Limit        int                    `json:"limit"`
	TotalPages   int                    `json:"total_pages"`
}

type JobApplicationStats struct {
	TotalApplications int64   `json:"total_applications"`
	PendingCount      int64   `json:"pending_count"`
	ShortlistedCount  int64   `json:"shortlisted_count"`
	RejectedCount     int64   `json:"rejected_count"`
	HiredCount        int64   `json:"hired_count"`
	AverageMatchScore float64 `json:"average_match_score"`
	ViewCount         int64   `json:"view_count"`
}

type ApplicationServiceImpl struct {
	cfg                *config.Config
	applicationRepo    repository.ApplicationRepository
	jobRepo            repository.JobRepository
	userRepo           repository.UserRepository
	emailService       EmailService
	notificationService NotificationService
}

func NewApplicationService(
	cfg *config.Config,
	applicationRepo repository.ApplicationRepository,
	jobRepo repository.JobRepository,
	userRepo repository.UserRepository,
	emailService EmailService,
) ApplicationService {
	return &ApplicationServiceImpl{
		cfg:             cfg,
		applicationRepo: applicationRepo,
		jobRepo:         jobRepo,
		userRepo:        userRepo,
		emailService:    emailService,
	}
}

func (s *ApplicationServiceImpl) SetNotificationService(ns NotificationService) {
	s.notificationService = ns
}

func (s *ApplicationServiceImpl) Apply(ctx context.Context, employeeID string, req *ApplyRequest) (*models.Application, error) {
	job, err := s.jobRepo.GetByID(ctx, req.JobID)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, errors.New("job not found")
	}

	if !job.IsActive {
		return nil, errors.New("this job is no longer accepting applications")
	}

	hasApplied, err := s.applicationRepo.HasApplied(ctx, req.JobID, employeeID)
	if err != nil {
		return nil, err
	}
	if hasApplied {
		return nil, errors.New("you have already applied for this job")
	}

	employeeProfile, err := s.userRepo.GetEmployeeProfile(ctx, employeeID)
	if err != nil {
		return nil, errors.New("employee profile not found")
	}

	matchScore := s.calculateMatchScore(employeeProfile, job)

	application := &models.Application{
		ID:           uuid.New().String(),
		JobID:        req.JobID,
		EmployeeID:   employeeID,
		CoverLetter:  req.CoverLetter,
		ResumeURL:    req.ResumeURL,
		PortfolioURL: req.PortfolioURL,
		MatchScore:   matchScore,
		Status:       "pending",
		AppliedAt:    time.Now(),
	}

	if err := s.applicationRepo.Create(ctx, application); err != nil {
		return nil, fmt.Errorf("failed to submit application: %w", err)
	}

	s.jobRepo.IncrementApplications(ctx, req.JobID)

	employerProfile, err := s.userRepo.GetEmployerProfile(ctx, job.EmployerID)
	if err == nil && employerProfile != nil && employerProfile.User != nil {
		go s.emailService.SendNewApplicationNotification(
			employerProfile.User.Email,
			job.Title,
			employeeProfile.FullName,
			application.ID,
			employerProfile.User.Email,
		)
	}

	if employeeProfile.User != nil {
		go s.emailService.SendApplicationConfirmation(
			employeeProfile.User.Email,
			job.Title,
			job.Employer.CompanyName,
			application.ID,
			employeeProfile.User.Email,
		)
	}

	if s.notificationService != nil {
		go s.notificationService.NotifyNewApplication(context.Background(), req.JobID, job.EmployerID, employeeID)
	}

	return application, nil
}

func (s *ApplicationServiceImpl) WithdrawApplication(ctx context.Context, applicationID string, employeeID string) error {
	application, err := s.applicationRepo.GetByID(ctx, applicationID)
	if err != nil {
		return err
	}
	if application == nil {
		return errors.New("application not found")
	}

	if application.EmployeeID != employeeID {
		return errors.New("unauthorized: you don't own this application")
	}

	if application.Status != "pending" && application.Status != "viewed" {
		return errors.New("application cannot be withdrawn at this stage")
	}

	return s.applicationRepo.UpdateStatus(ctx, applicationID, "withdrawn", "Withdrawn by candidate")
}

func (s *ApplicationServiceImpl) GetApplication(ctx context.Context, applicationID string, userID string, role string) (*models.Application, error) {
	application, err := s.applicationRepo.GetByIDWithDetails(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, errors.New("application not found")
	}

	if role == "employee" && application.EmployeeID != userID {
		return nil, errors.New("unauthorized: you don't own this application")
	}

	if role == "employer" && application.Job.EmployerID != userID {
		return nil, errors.New("unauthorized: you don't own this job")
	}

	if role == "employer" && application.ViewedAt == nil {
		go s.applicationRepo.MarkAsViewed(ctx, applicationID)
	}

	return application, nil
}

func (s *ApplicationServiceImpl) ListJobApplications(ctx context.Context, jobID string, employerID string, status string, page, limit int) (*ApplicationListResponse, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, errors.New("job not found")
	}

	if job.EmployerID != employerID {
		return nil, errors.New("unauthorized: you don't own this job")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 100 {
		limit = 100
	}

	applications, total, err := s.applicationRepo.ListByJob(ctx, jobID, status, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &ApplicationListResponse{
		Applications: applications,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
	}, nil
}

func (s *ApplicationServiceImpl) ListMyApplications(ctx context.Context, employeeID string, page, limit int) (*ApplicationListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 100 {
		limit = 100
	}

	applications, total, err := s.applicationRepo.ListByEmployee(ctx, employeeID, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &ApplicationListResponse{
		Applications: applications,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
	}, nil
}

func (s *ApplicationServiceImpl) ListCompanyApplications(ctx context.Context, employerID string, status string, page, limit int) (*ApplicationListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 100 {
		limit = 100
	}

	applications, total, err := s.applicationRepo.ListByEmployer(ctx, employerID, status, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &ApplicationListResponse{
		Applications: applications,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
	}, nil
}

func (s *ApplicationServiceImpl) ShortlistApplication(ctx context.Context, applicationID string, employerID string, notes string) error {
	application, err := s.applicationRepo.GetByIDWithDetails(ctx, applicationID)
	if err != nil {
		return err
	}
	if application == nil {
		return errors.New("application not found")
	}

	if application.Job.EmployerID != employerID {
		return errors.New("unauthorized: you don't own this job")
	}

	if err := s.applicationRepo.Shortlist(ctx, applicationID, notes); err != nil {
		return err
	}

	if application.Employee != nil && application.Employee.User != nil {
		go s.emailService.SendApplicationStatusUpdate(
			application.Employee.User.Email,
			application.Job.Title,
			application.Job.Employer.CompanyName,
			"shortlisted",
			application.Employee.User.Email,
		)
	}

	if s.notificationService != nil {
		go s.notificationService.NotifyApplicationStatusChange(context.Background(), applicationID, application.EmployeeID, employerID, "shortlisted")
	}

	return nil
}

func (s *ApplicationServiceImpl) RejectApplication(ctx context.Context, applicationID string, employerID string, notes string) error {
	application, err := s.applicationRepo.GetByIDWithDetails(ctx, applicationID)
	if err != nil {
		return err
	}
	if application == nil {
		return errors.New("application not found")
	}

	if application.Job.EmployerID != employerID {
		return errors.New("unauthorized: you don't own this job")
	}

	if err := s.applicationRepo.Reject(ctx, applicationID, notes); err != nil {
		return err
	}

	if application.Employee != nil && application.Employee.User != nil {
		go s.emailService.SendApplicationStatusUpdate(
			application.Employee.User.Email,
			application.Job.Title,
			application.Job.Employer.CompanyName,
			"rejected",
			application.Employee.User.Email,
		)
	}

	if s.notificationService != nil {
		go s.notificationService.NotifyApplicationStatusChange(context.Background(), applicationID, application.EmployeeID, employerID, "rejected")
	}

	return nil
}

func (s *ApplicationServiceImpl) SaveNotes(ctx context.Context, applicationID string, employerID string, notes string) error {
	application, err := s.applicationRepo.GetByIDWithDetails(ctx, applicationID)
	if err != nil {
		return err
	}
	if application == nil {
		return errors.New("application not found")
	}

	if application.Job.EmployerID != employerID {
		return errors.New("unauthorized: you don't own this job")
	}

	return s.applicationRepo.UpdateNotes(ctx, applicationID, notes)
}

func (s *ApplicationServiceImpl) ScheduleInterview(ctx context.Context, applicationID string, employerID string, interviewDate time.Time, interviewType string) error {
	application, err := s.applicationRepo.GetByIDWithDetails(ctx, applicationID)
	if err != nil {
		return err
	}
	if application == nil {
		return errors.New("application not found")
	}

	if application.Job.EmployerID != employerID {
		return errors.New("unauthorized: you don't own this job")
	}

	if err := s.applicationRepo.ScheduleInterview(ctx, applicationID, interviewDate, interviewType); err != nil {
		return err
	}

	if application.Employee != nil && application.Employee.User != nil {
		go s.emailService.SendInterviewInvitation(
			application.Employee.User.Email,
			application.Job.Title,
			application.Job.Employer.CompanyName,
			interviewDate,
			interviewType,
			applicationID,
			application.Employee.User.Email,
		)
	}

	if s.notificationService != nil {
		go s.notificationService.NotifyInterviewScheduled(context.Background(), applicationID, application.EmployeeID, employerID, interviewDate)
	}

	return nil
}

func (s *ApplicationServiceImpl) MarkAsHired(ctx context.Context, applicationID string, employerID string) error {
	application, err := s.applicationRepo.GetByIDWithDetails(ctx, applicationID)
	if err != nil {
		return err
	}
	if application == nil {
		return errors.New("application not found")
	}

	if application.Job.EmployerID != employerID {
		return errors.New("unauthorized: you don't own this job")
	}

	if err := s.applicationRepo.UpdateStatus(ctx, applicationID, "hired", "Hired"); err != nil {
		return err
	}

	s.userRepo.IncrementHiresCount(ctx, employerID)

	if application.Employee != nil && application.Employee.User != nil {
		go s.emailService.SendHiredNotification(
			application.Employee.User.Email,
			application.Job.Title,
			application.Job.Employer.CompanyName,
			application.Employee.User.Email,
		)
	}

	if s.notificationService != nil {
		go s.notificationService.NotifyApplicationStatusChange(context.Background(), applicationID, application.EmployeeID, employerID, "hired")
	}

	return nil
}

func (s *ApplicationServiceImpl) GetApplicationStats(ctx context.Context, userID string, role string) (*repository.ApplicationStats, error) {
	if role == "employee" {
		return s.applicationRepo.GetEmployeeApplicationStats(ctx, userID)
	}
	return s.applicationRepo.GetEmployerApplicationStats(ctx, userID)
}

func (s *ApplicationServiceImpl) GetJobApplicationStats(ctx context.Context, jobID string, employerID string) (*JobApplicationStats, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, errors.New("job not found")
	}

	if job.EmployerID != employerID {
		return nil, errors.New("unauthorized: you don't own this job")
	}

	stats := &JobApplicationStats{
		ViewCount: int64(job.ViewsCount),
	}

	var results []struct {
		Status string
		Count  int64
	}

	err = s.applicationRepo.DB().WithContext(ctx).
		Model(&models.Application{}).
		Select("status, COUNT(*) as count").
		Where("job_id = ?", jobID).
		Group("status").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	for _, r := range results {
		stats.TotalApplications += r.Count
		switch r.Status {
		case "pending":
			stats.PendingCount = r.Count
		case "shortlisted":
			stats.ShortlistedCount = r.Count
		case "rejected":
			stats.RejectedCount = r.Count
		case "hired":
			stats.HiredCount = r.Count
		}
	}

	var avgScore float64
	s.applicationRepo.DB().WithContext(ctx).
		Model(&models.Application{}).
		Where("job_id = ? AND match_score > 0", jobID).
		Select("AVG(match_score)").
		Scan(&avgScore)
	stats.AverageMatchScore = avgScore

	return stats, nil
}

func (s *ApplicationServiceImpl) BulkShortlist(ctx context.Context, applicationIDs []string, employerID string) error {
	for _, id := range applicationIDs {
		app, err := s.applicationRepo.GetByIDWithDetails(ctx, id)
		if err != nil {
			return err
		}
		if app == nil {
			return fmt.Errorf("application %s not found", id)
		}
		if app.Job.EmployerID != employerID {
			return fmt.Errorf("unauthorized: application %s belongs to another employer", id)
		}
	}
	return s.applicationRepo.BulkUpdateStatus(ctx, applicationIDs, "shortlisted")
}

func (s *ApplicationServiceImpl) BulkReject(ctx context.Context, applicationIDs []string, employerID string) error {
	for _, id := range applicationIDs {
		app, err := s.applicationRepo.GetByIDWithDetails(ctx, id)
		if err != nil {
			return err
		}
		if app == nil {
			return fmt.Errorf("application %s not found", id)
		}
		if app.Job.EmployerID != employerID {
			return fmt.Errorf("unauthorized: application %s belongs to another employer", id)
		}
	}
	return s.applicationRepo.BulkUpdateStatus(ctx, applicationIDs, "rejected")
}

func (s *ApplicationServiceImpl) calculateMatchScore(employee *models.EmployeeProfile, job *models.Job) int {
	score := 0

	if len(job.RequiredSkills) > 0 {
		matchingSkills := 0.0
		employeeSkills := make(map[string]bool)
		for _, skill := range employee.Skills {
			employeeSkills[normalizeString(skill)] = true
		}

		for _, reqSkill := range job.RequiredSkills {
			normalizedReq := normalizeString(reqSkill)
			if employeeSkills[normalizedReq] {
				matchingSkills++
			} else {
				for empSkill := range employeeSkills {
					if strings.Contains(empSkill, normalizedReq) || strings.Contains(normalizedReq, empSkill) {
						matchingSkills += 0.5
						break
					}
				}
			}
		}

		skillsScore := int((matchingSkills / float64(len(job.RequiredSkills))) * 60)
		score += skillsScore
	}

	if job.MinExperience > 0 {
		if employee.YearsOfExperience >= job.MinExperience {
			expScore := 20
			if job.MaxExperience > 0 && employee.YearsOfExperience <= job.MaxExperience {
				expScore = 30
			} else if employee.YearsOfExperience > job.MinExperience {
				expScore = 25
			}
			score += expScore
		} else if employee.YearsOfExperience > 0 {
			experienceRatio := float64(employee.YearsOfExperience) / float64(job.MinExperience)
			score += int(experienceRatio * 20)
		}
	} else if employee.YearsOfExperience > 0 {
		score += 10
	}

	if job.IsRemote {
		if employee.IsRemoteOnly {
			score += 10
		} else {
			score += 7
		}
	} else if job.IsHybrid && employee.IsRemoteOnly {
		score += 5
	} else if !job.IsRemote && job.Location != "" && employee.Location != "" {
		if strings.Contains(strings.ToLower(job.Location), strings.ToLower(employee.Location)) ||
			strings.Contains(strings.ToLower(employee.Location), strings.ToLower(job.Location)) {
			score += 8
		} else if strings.ToLower(job.Location) == "nairobi" {
			score += 5
		}
	}

	if len(employee.Skills) >= 5 && employee.YearsOfExperience > 0 && employee.Bio != "" {
		score += 5
	}

	if employee.GithubConnected {
		score += 5
	}

	if score > 100 {
		score = 100
	}

	return score
}

func normalizeString(s string) string {
	if s == "" {
		return s
	}

	normalized := strings.ToLower(s)
	normalized = strings.TrimSpace(normalized)

	punctuations := []string{".", ",", "!", "?", ";", ":", "'", "\"", "(", ")", "[", "]", "{", "}"}
	for _, p := range punctuations {
		normalized = strings.ReplaceAll(normalized, p, "")
	}

	replacements := map[string]string{
		"javascript": "js",
		"typescript": "ts",
		"reactjs":    "react",
		"react.js":   "react",
		"nodejs":     "node",
		"node.js":    "node",
		"postgresql": "postgres",
		"mongodb":    "mongo",
		"aws":        "amazon web services",
		"gcp":        "google cloud",
		"azure":      "microsoft azure",
	}

	if val, ok := replacements[normalized]; ok {
		normalized = val
	}

	return normalized
}

func (s *ApplicationServiceImpl) DB() interface{} {
	return s.applicationRepo.DB()
}