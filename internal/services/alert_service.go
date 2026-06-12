package services

import (
	"context"
	"fmt"
	"time"

	
	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type AlertService interface {
	// Saved search management
	CreateSavedSearch(ctx context.Context, userID string, req *CreateSavedSearchRequest) (*models.SavedSearch, error)
	GetSavedSearch(ctx context.Context, id, userID string) (*models.SavedSearch, error)
	UpdateSavedSearch(ctx context.Context, id, userID string, req *UpdateSavedSearchRequest) (*models.SavedSearch, error)
	DeleteSavedSearch(ctx context.Context, id, userID string) error
	ListSavedSearches(ctx context.Context, userID string, page, limit int) (*SavedSearchListResponse, error)
	
	// Alert execution
	ProcessAlert(ctx context.Context, savedSearch *models.SavedSearch) (*AlertResult, error)
	ProcessAllAlerts(ctx context.Context) error
	SendTestAlert(ctx context.Context, savedSearchID, userID string) (*AlertResult, error)
	
	// Alert settings
	GetAlertSettings(ctx context.Context, userID string) (*models.AlertSettings, error)
	UpdateAlertSettings(ctx context.Context, userID string, req *UpdateAlertSettingsRequest) (*models.AlertSettings, error)
	
	// History
	GetAlertHistory(ctx context.Context, userID string, days int) ([]*models.AlertHistory, error)
}

type CreateSavedSearchRequest struct {
	Name        string                  `json:"name" validate:"required,min=3,max=255"`
	Filters     models.SearchFilters    `json:"filters" validate:"required"`
	Frequency   string                  `json:"frequency" validate:"oneof=daily weekly instant"`
	IsActive    bool                    `json:"is_active"`
}

type UpdateSavedSearchRequest struct {
	Name        *string                 `json:"name"`
	Filters     *models.SearchFilters   `json:"filters"`
	Frequency   *string                 `json:"frequency"`
	IsActive    *bool                   `json:"is_active"`
}

type UpdateAlertSettingsRequest struct {
	EmailEnabled       *bool `json:"email_enabled"`
	PushEnabled        *bool `json:"push_enabled"`
	NewJobAlerts       *bool `json:"new_job_alerts"`
	ApplicationUpdates *bool `json:"application_updates"`
	MarketingEmails    *bool `json:"marketing_emails"`
	EmailDigestHour    *int  `json:"email_digest_hour"`
	EmailDigestDay     *int  `json:"email_digest_day"`
}

type SavedSearchListResponse struct {
	Searches   []*models.SavedSearch `json:"searches"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	TotalPages int                   `json:"total_pages"`
}

type AlertResult struct {
	SavedSearchID string   `json:"saved_search_id"`
	JobsFound     int      `json:"jobs_found"`
	JobsSent      int      `json:"jobs_sent"`
	JobIDs        []string `json:"job_ids"`
	SentAt        time.Time `json:"sent_at"`
}

type AlertServiceImpl struct {
	cfg             *config.Config
	alertRepo       repository.AlertRepository
	jobRepo         repository.JobRepository
	emailService    EmailService
	notificationSvc NotificationService
}

func NewAlertService(
	cfg *config.Config,
	alertRepo repository.AlertRepository,
	jobRepo repository.JobRepository,
	emailService EmailService,
	notificationSvc NotificationService,
) AlertService {
	return &AlertServiceImpl{
		cfg:             cfg,
		alertRepo:       alertRepo,
		jobRepo:         jobRepo,
		emailService:    emailService,
		notificationSvc: notificationSvc,
	}
}

func (s *AlertServiceImpl) CreateSavedSearch(ctx context.Context, userID string, req *CreateSavedSearchRequest) (*models.SavedSearch, error) {
	// Validate frequency
	if req.Frequency == "" {
		req.Frequency = "daily"
	}
	
	// Create saved search
	savedSearch := &models.SavedSearch{
		UserID:    userID,
		Name:      req.Name,
		Filters:   req.Filters,
		Frequency: req.Frequency,
		IsActive:  req.IsActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if !req.IsActive {
		savedSearch.IsActive = true
	}
	
	if err := s.alertRepo.CreateSavedSearch(ctx, savedSearch); err != nil {
		return nil, fmt.Errorf("failed to create saved search: %w", err)
	}
	
	// If instant alerts are enabled, process immediately
	if req.Frequency == "instant" && req.IsActive {
		go s.ProcessAlert(context.Background(), savedSearch)
	}
	
	return savedSearch, nil
}

func (s *AlertServiceImpl) GetSavedSearch(ctx context.Context, id, userID string) (*models.SavedSearch, error) {
	return s.alertRepo.GetSavedSearchByUser(ctx, id, userID)
}

func (s *AlertServiceImpl) UpdateSavedSearch(ctx context.Context, id, userID string, req *UpdateSavedSearchRequest) (*models.SavedSearch, error) {
	savedSearch, err := s.alertRepo.GetSavedSearchByUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	
	if req.Name != nil {
		savedSearch.Name = *req.Name
	}
	if req.Filters != nil {
		savedSearch.Filters = *req.Filters
	}
	if req.Frequency != nil {
		savedSearch.Frequency = *req.Frequency
	}
	if req.IsActive != nil {
		savedSearch.IsActive = *req.IsActive
	}
	
	savedSearch.UpdatedAt = time.Now()
	
	if err := s.alertRepo.UpdateSavedSearch(ctx, savedSearch); err != nil {
		return nil, fmt.Errorf("failed to update saved search: %w", err)
	}
	
	return savedSearch, nil
}

func (s *AlertServiceImpl) DeleteSavedSearch(ctx context.Context, id, userID string) error {
	return s.alertRepo.DeleteSavedSearch(ctx, id, userID)
}

func (s *AlertServiceImpl) ListSavedSearches(ctx context.Context, userID string, page, limit int) (*SavedSearchListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 100 {
		limit = 100
	}
	
	searches, total, err := s.alertRepo.ListSavedSearchesByUser(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}
	
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	
	return &SavedSearchListResponse{
		Searches:   searches,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *AlertServiceImpl) ProcessAlert(ctx context.Context, savedSearch *models.SavedSearch) (*AlertResult, error) {
	// Build job filters from saved search criteria
	filters := s.buildJobFilters(savedSearch.Filters)
	
	// Find matching jobs
	jobs, total, err := s.jobRepo.List(ctx, filters, 1, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching jobs: %w", err)
	}
	
	// Filter out jobs that were already sent
	jobIDs := make([]string, 0, len(jobs))
	for _, job := range jobs {
		jobIDs = append(jobIDs, job.ID)
	}
	
	result := &AlertResult{
		SavedSearchID: savedSearch.ID,
		JobsFound:     int(total),
		JobsSent:      len(jobs),
		JobIDs:        jobIDs,
		SentAt:        time.Now(),
	}
	
	// Send email alert if user has email enabled and there are jobs
	if len(jobs) > 0 && savedSearch.User != nil && savedSearch.User.Email != "" {
		settings, _ := s.alertRepo.GetAlertSettings(ctx, savedSearch.UserID)
		if settings == nil || settings.EmailEnabled {
			s.sendAlertEmail(savedSearch, jobs, result)
		}
	}
	
	// Send in-app notification when new jobs found
	if len(jobs) > 0 && savedSearch.UserID != "" {
		s.notificationSvc.NotifyJobAlert(ctx, savedSearch.UserID, len(jobs))
	}

	// Save to history
	history := &models.AlertHistory{
		SavedSearchID: savedSearch.ID,
		JobsFound:     int(total),
		JobsSent:      len(jobs),
		JobIDs:        jobIDs,
	}
	
	if err := s.alertRepo.CreateAlertHistory(ctx, history); err != nil {
		return result, fmt.Errorf("failed to save alert history: %w", err)
	}
	
	// Update last sent timestamp
	s.alertRepo.UpdateLastSent(ctx, savedSearch.ID, time.Now())
	
	return result, nil
}

func (s *AlertServiceImpl) ProcessAllAlerts(ctx context.Context) error {
	searches, err := s.alertRepo.ListActiveSavedSearches(ctx)
	if err != nil {
		return fmt.Errorf("failed to list active searches: %w", err)
	}
	
	for _, search := range searches {
		// Check if alert should be sent based on frequency
		if !s.shouldSendAlert(search) {
			continue
		}
		
		go func(search *models.SavedSearch) {
			if _, err := s.ProcessAlert(context.Background(), search); err != nil {
				fmt.Printf("Failed to process alert for search %s: %v\n", search.ID, err)
			}
		}(search)
	}
	
	return nil
}

func (s *AlertServiceImpl) SendTestAlert(ctx context.Context, savedSearchID, userID string) (*AlertResult, error) {
	savedSearch, err := s.alertRepo.GetSavedSearchByUser(ctx, savedSearchID, userID)
	if err != nil {
		return nil, err
	}
	
	return s.ProcessAlert(ctx, savedSearch)
}

func (s *AlertServiceImpl) GetAlertSettings(ctx context.Context, userID string) (*models.AlertSettings, error) {
	settings, err := s.alertRepo.GetAlertSettings(ctx, userID)
	if err != nil {
		// Create default settings if not exists
		defaultSettings := &models.AlertSettings{
			UserID:            userID,
			EmailEnabled:      true,
			PushEnabled:       false,
			NewJobAlerts:      true,
			ApplicationUpdates: true,
			MarketingEmails:   false,
			EmailDigestHour:   8,
			EmailDigestDay:    1,
		}
		if err := s.alertRepo.CreateAlertSettings(ctx, defaultSettings); err != nil {
			return nil, err
		}
		return defaultSettings, nil
	}
	return settings, nil
}

func (s *AlertServiceImpl) UpdateAlertSettings(ctx context.Context, userID string, req *UpdateAlertSettingsRequest) (*models.AlertSettings, error) {
	updates := make(map[string]interface{})
	
	if req.EmailEnabled != nil {
		updates["email_enabled"] = *req.EmailEnabled
	}
	if req.PushEnabled != nil {
		updates["push_enabled"] = *req.PushEnabled
	}
	if req.NewJobAlerts != nil {
		updates["new_job_alerts"] = *req.NewJobAlerts
	}
	if req.ApplicationUpdates != nil {
		updates["application_updates"] = *req.ApplicationUpdates
	}
	if req.MarketingEmails != nil {
		updates["marketing_emails"] = *req.MarketingEmails
	}
	if req.EmailDigestHour != nil {
		updates["email_digest_hour"] = *req.EmailDigestHour
	}
	if req.EmailDigestDay != nil {
		updates["email_digest_day"] = *req.EmailDigestDay
	}
	
	if len(updates) > 0 {
		if err := s.alertRepo.UpdateAlertSettings(ctx, userID, updates); err != nil {
			return nil, fmt.Errorf("failed to update settings: %w", err)
		}
	}
	
	return s.GetAlertSettings(ctx, userID)
}

func (s *AlertServiceImpl) GetAlertHistory(ctx context.Context, userID string, days int) ([]*models.AlertHistory, error) {
	return s.alertRepo.GetRecentAlerts(ctx, userID, days)
}

// Private helper methods

func (s *AlertServiceImpl) buildJobFilters(filters models.SearchFilters) repository.JobFilters {
	repoFilters := repository.JobFilters{
		Title:           filters.JobTitle,
		Location:        filters.Location,
		IsRemote:        filters.IsRemote,
		EmploymentType:  filters.EmploymentType,
		ExperienceLevel: filters.ExperienceLevel,
		MinSalary:       filters.MinSalary,
		MaxSalary:       filters.MaxSalary,
		Skills:          filters.Skills,
		IsActive:        boolPtr(true),
	}
	return repoFilters
}

func (s *AlertServiceImpl) shouldSendAlert(search *models.SavedSearch) bool {
	if !search.IsActive {
		return false
	}
	
	now := time.Now()
	
	switch search.Frequency {
	case "instant":
		return true
	case "daily":
		if search.LastSentAt == nil {
			return true
		}
		// Send once per day
		return now.Sub(*search.LastSentAt) >= 24*time.Hour
	case "weekly":
		if search.LastSentAt == nil {
			return true
		}
		// Send once per week
		return now.Sub(*search.LastSentAt) >= 7*24*time.Hour
	default:
		return false
	}
}

func (s *AlertServiceImpl) sendAlertEmail(search *models.SavedSearch, jobs []*models.Job, result *AlertResult) {
	// Prepare job alerts for email
	jobAlerts := make([]JobAlert, 0, len(jobs))
	for _, job := range jobs {
		salaryRange := ""
		if job.SalaryMin > 0 && job.SalaryMax > 0 {
			salaryRange = formatMoney(job.SalaryMin) + " - " + formatMoney(job.SalaryMax) + " " + job.SalaryCurrency
		}
		
		jobAlerts = append(jobAlerts, JobAlert{
			Title:       job.Title,
			Company:     job.Employer.CompanyName,
			Location:    job.GetLocationDisplay(),
			SalaryRange: salaryRange,
			JobURL:      fmt.Sprintf("%s/jobs/%s", s.cfg.AppURL, job.ID),
		})
	}
	
	s.emailService.SendJobAlertEmail(search.User.Email, jobAlerts, search.User.Email)
}

func boolPtr(b bool) *bool {
	return &b
}

func formatMoney(amount int) string {
	if amount >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(amount)/1000000)
	}
	if amount >= 1000 {
		return fmt.Sprintf("%dK", amount/1000)
	}
	return fmt.Sprintf("%d", amount)
}