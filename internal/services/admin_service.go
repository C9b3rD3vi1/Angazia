package services

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type SearchService interface {
	// Search operations
	SearchJobs(ctx context.Context, filters models.SearchFilters, page, limit int) (*models.SearchResponse, error)
	SearchCandidates(ctx context.Context, filters models.SearchFilters, page, limit int) (*models.SearchResponse, error)
	SearchCompanies(ctx context.Context, filters models.SearchFilters, page, limit int) (*models.SearchResponse, error)

	// Facets
	GetJobFacets(ctx context.Context, filters models.SearchFilters) (*models.FacetResult, error)

	// Search history
	SaveSearchHistory(ctx context.Context, userID string, query string, filters models.SearchFilters, entityType string, resultsCount int, ipAddress, userAgent string) error
	GetSearchHistory(ctx context.Context, userID string, limit int) ([]*models.SearchQuery, error)
	GetPopularSearches(ctx context.Context, days, limit int) ([]PopularSearch, error)

	// Saved searches
	SaveSearch(ctx context.Context, userID string, name string, filters models.SearchFilters, entityType string, frequency string) (*models.SavedSearch, error)
	GetSavedSearches(ctx context.Context, userID string, entityType string) ([]*models.SavedSearch, error)
	DeleteSavedSearch(ctx context.Context, id, userID string) error
	RunSavedSearch(ctx context.Context, id, userID string) (*models.SearchResponse, error)

	// Auto-complete
	AutoComplete(ctx context.Context, prefix string, entityType string, limit int) ([]string, error)
}

type PopularSearch struct {
	Query string `json:"query"`
	Count int    `json:"count"`
}

type SystemHealth struct {
	API             bool  `json:"api"`
	APILatency      int64 `json:"api_latency"`
	Database        bool  `json:"database"`
	DatabaseLatency int64 `json:"database_latency"`
	Redis           bool  `json:"redis"`
	Elasticsearch   bool  `json:"elasticsearch"`
}

type SearchServiceImpl struct {
	cfg        *config.Config
	searchRepo repository.SearchRepository
	jobRepo    repository.JobRepository
	userRepo   repository.UserRepository
}

func NewSearchService(
	cfg *config.Config,
	searchRepo repository.SearchRepository,
	jobRepo repository.JobRepository,
	userRepo repository.UserRepository,
) SearchService {
	return &SearchServiceImpl{
		cfg:        cfg,
		searchRepo: searchRepo,
		jobRepo:    jobRepo,
		userRepo:   userRepo,
	}
}

func (s *SearchServiceImpl) SearchJobs(ctx context.Context, filters models.SearchFilters, page, limit int) (*models.SearchResponse, error) {
	startTime := time.Now()

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 100 {
		limit = 100
	}

	jobs, total, err := s.searchRepo.SearchJobs(ctx, filters, page, limit)
	if err != nil {
		return nil, err
	}

	results := make([]models.SearchResult, len(jobs))
	keywords := strings.Fields(filters.Keywords)
	for i, job := range jobs {
		score := 1.0
		if len(keywords) > 0 {
			matchCount := 0
			checks := 0
			for _, kw := range keywords {
				kwLower := strings.ToLower(kw)
				if strings.Contains(strings.ToLower(job.Title), kwLower) {
					matchCount++
				}
				checks++
				if strings.Contains(strings.ToLower(job.Description), kwLower) {
					matchCount++
				}
				checks++
				for _, skill := range job.RequiredSkills {
					if strings.Contains(strings.ToLower(skill), kwLower) {
						matchCount++
						break
					}
				}
				checks++
				if strings.Contains(strings.ToLower(job.Location), kwLower) {
					matchCount++
				}
				checks++
			}
			if checks > 0 {
				score = float64(matchCount) / float64(checks)
			}
		}
		results[i] = models.SearchResult{
			ID:          job.ID,
			Type:        "job",
			Title:       job.Title,
			Description: job.Description,
			Score:       score,
			Data:        job,
		}
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	// Get facets
	facets, _ := s.searchRepo.GetJobFacets(ctx, filters)

	searchTimeMs := time.Since(startTime).Milliseconds()

	resp := &models.SearchResponse{
		Results:      results,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
		SearchTimeMs: searchTimeMs,
	}
	if facets != nil {
		resp.Facets = *facets
	}
	return resp, nil
}

func (s *SearchServiceImpl) SearchCandidates(ctx context.Context, filters models.SearchFilters, page, limit int) (*models.SearchResponse, error) {
	startTime := time.Now()

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 100 {
		limit = 100
	}

	candidates, total, err := s.searchRepo.SearchCandidates(ctx, filters, page, limit)
	if err != nil {
		return nil, err
	}

	results := make([]models.SearchResult, len(candidates))
	keywords := strings.Fields(filters.Keywords)
	for i, candidate := range candidates {
		score := 1.0
		if len(keywords) > 0 {
			matchCount := 0
			checks := 0
			for _, kw := range keywords {
				kwLower := strings.ToLower(kw)
				if strings.Contains(strings.ToLower(candidate.FullName), kwLower) {
					matchCount++
				}
				checks++
				if strings.Contains(strings.ToLower(candidate.Headline), kwLower) {
					matchCount++
				}
				checks++
				for _, skill := range candidate.Skills {
					if strings.Contains(strings.ToLower(skill), kwLower) {
						matchCount++
						break
					}
				}
				checks++
			}
			if checks > 0 {
				score = float64(matchCount) / float64(checks)
			}
		}
		results[i] = models.SearchResult{
			ID:          candidate.UserID,
			Type:        "candidate",
			Title:       candidate.FullName,
			Description: candidate.Headline,
			Score:       score,
			Data:        candidate,
		}
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	searchTimeMs := time.Since(startTime).Milliseconds()

	return &models.SearchResponse{
		Results:      results,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
		SearchTimeMs: searchTimeMs,
	}, nil
}

func (s *SearchServiceImpl) SearchCompanies(ctx context.Context, filters models.SearchFilters, page, limit int) (*models.SearchResponse, error) {
	startTime := time.Now()

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = s.cfg.PageSize
	}
	if limit > 100 {
		limit = 100
	}

	companies, total, err := s.searchRepo.SearchCompanies(ctx, filters, page, limit)
	if err != nil {
		return nil, err
	}

	results := make([]models.SearchResult, len(companies))
	keywords := strings.Fields(filters.Keywords)
	for i, company := range companies {
		score := 1.0
		if len(keywords) > 0 {
			matchCount := 0
			checks := 0
			for _, kw := range keywords {
				kwLower := strings.ToLower(kw)
				if strings.Contains(strings.ToLower(company.CompanyName), kwLower) {
					matchCount++
				}
				checks++
				if strings.Contains(strings.ToLower(company.CompanyDescription), kwLower) {
					matchCount++
				}
				checks++
			}
			if checks > 0 {
				score = float64(matchCount) / float64(checks)
			}
		}
		results[i] = models.SearchResult{
			ID:          company.UserID,
			Type:        "company",
			Title:       company.CompanyName,
			Description: company.CompanyDescription,
			Score:       score,
			Data:        company,
		}
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	searchTimeMs := time.Since(startTime).Milliseconds()

	return &models.SearchResponse{
		Results:      results,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
		SearchTimeMs: searchTimeMs,
	}, nil
}

func (s *SearchServiceImpl) GetJobFacets(ctx context.Context, filters models.SearchFilters) (*models.FacetResult, error) {
	return s.searchRepo.GetJobFacets(ctx, filters)
}

func (s *SearchServiceImpl) SaveSearchHistory(ctx context.Context, userID string, query string, filters models.SearchFilters, entityType string, resultsCount int, ipAddress, userAgent string) error {
	searchQuery := &models.SearchQuery{
		UserID:       userID,
		Query:        query,
		Filters:      filters,
		EntityType:   entityType,
		ResultsCount: resultsCount,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	return s.searchRepo.LogSearch(ctx, searchQuery)
}

func (s *SearchServiceImpl) GetSearchHistory(ctx context.Context, userID string, limit int) ([]*models.SearchQuery, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	return s.searchRepo.GetSearchHistory(ctx, userID, limit)
}

func (s *SearchServiceImpl) GetPopularSearches(ctx context.Context, days, limit int) ([]PopularSearch, error) {
	if days < 1 {
		days = 7
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	results, err := s.searchRepo.GetPopularSearches(ctx, days, limit)
	if err != nil {
		return nil, err
	}

	popular := make([]PopularSearch, len(results))
	for i, r := range results {
		popular[i] = PopularSearch{
			Query: r.Query,
			Count: r.Count,
		}
	}

	return popular, nil
}

func (s *SearchServiceImpl) SaveSearch(ctx context.Context, userID string, name string, filters models.SearchFilters, entityType string, frequency string) (*models.SavedSearch, error) {
	if frequency == "" {
		frequency = "daily"
	}

	savedSearch := &models.SavedSearch{
		UserID:     userID,
		Name:       name,
		Filters:    filters,
		EntityType: entityType,
		Frequency:  frequency,
		IsActive:   true,
	}

	if err := s.searchRepo.SaveSearch(ctx, savedSearch); err != nil {
		return nil, err
	}

	return savedSearch, nil
}

func (s *SearchServiceImpl) GetSavedSearches(ctx context.Context, userID string, entityType string) ([]*models.SavedSearch, error) {
	return s.searchRepo.GetSavedSearches(ctx, userID, entityType)
}

func (s *SearchServiceImpl) DeleteSavedSearch(ctx context.Context, id, userID string) error {
	return s.searchRepo.DeleteSavedSearch(ctx, id, userID)
}

func (s *SearchServiceImpl) RunSavedSearch(ctx context.Context, id, userID string) (*models.SearchResponse, error) {
	searches, err := s.searchRepo.GetSavedSearches(ctx, userID, "")
	if err != nil {
		return nil, err
	}

	var target *models.SavedSearch
	for _, search := range searches {
		if search.ID == id {
			target = search
			break
		}
	}

	if target == nil {
		return nil, fmt.Errorf("saved search not found")
	}

	// Update last run timestamp
	s.searchRepo.UpdateSavedSearchLastRun(ctx, id)

	switch target.EntityType {
	case "job":
		return s.SearchJobs(ctx, target.Filters, 1, 20)
	case "candidate":
		return s.SearchCandidates(ctx, target.Filters, 1, 20)
	case "company":
		return s.SearchCompanies(ctx, target.Filters, 1, 20)
	default:
		return nil, fmt.Errorf("invalid entity type")
	}
}

func (s *SearchServiceImpl) AutoComplete(ctx context.Context, prefix string, entityType string, limit int) ([]string, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	seen := make(map[string]struct{})
	var suggestions []string

	switch entityType {
	case "job":
		jobs, _, err := s.searchRepo.SearchJobs(ctx, models.SearchFilters{Keywords: prefix}, 1, limit)
		if err == nil {
			for _, j := range jobs {
				if j != nil && j.Title != "" {
					if _, ok := seen[j.Title]; !ok {
						seen[j.Title] = struct{}{}
						suggestions = append(suggestions, j.Title)
					}
				}
			}
		}
		if len(suggestions) == 0 {
			suggestions = append(suggestions, prefix)
		}
	case "skill":
		jobs, _, err := s.searchRepo.SearchJobs(ctx, models.SearchFilters{Keywords: prefix}, 1, limit)
		if err == nil {
			for _, j := range jobs {
				if j != nil {
					for _, skill := range j.RequiredSkills {
						if _, ok := seen[skill]; !ok {
							seen[skill] = struct{}{}
							suggestions = append(suggestions, skill)
						}
					}
				}
			}
		}
		if len(suggestions) == 0 {
			suggestions = append(suggestions, prefix)
		}
	default:
		suggestions = append(suggestions, prefix)
	}

	return suggestions, nil
}

// ============================================================
// Admin Service
// ============================================================

type AdminService interface {
	GetPlatformStats(ctx context.Context) (*models.PlatformStats, error)
	GetUserStats(ctx context.Context) (map[string]int, error)
	GetJobStats(ctx context.Context) (map[string]int, error)
	GetEngagementStats(ctx context.Context) (map[string]int, error)
	GetAllUsers(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.UserReport, int64, error)
	GetUserDetails(ctx context.Context, userID string) (*models.UserReport, error)
	SuspendUser(ctx context.Context, userID string) error
	ActivateUser(ctx context.Context, userID string) error
	DeleteUser(ctx context.Context, userID string) error
	VerifyUser(ctx context.Context, userID string) error
	GetModerationQueue(ctx context.Context, entityType, status string, page, limit int, dateFrom, dateTo *time.Time) ([]*models.ModerationQueue, int64, error)
	GetModerationItem(ctx context.Context, id string) (*models.ModerationQueue, error)
	ApproveContent(ctx context.Context, id, reviewerID string) error
	RejectContent(ctx context.Context, id, reviewerID, reason string) error
	GetSettings(ctx context.Context, category string) ([]*models.SystemSetting, error)
	UpdateSetting(ctx context.Context, key, value string) error
	CreateSetting(ctx context.Context, key, value, settingType, category, description string, isPublic bool) error
	ApproveCompanyVerification(ctx context.Context, adminID, companyID string) error
	RejectCompanyVerification(ctx context.Context, adminID, companyID, reason string) error
	GetPendingVerifications(ctx context.Context, page, limit int) (*PendingVerificationsResponse, error)
	GetReportReasons(ctx context.Context, entityType string) ([]*models.ReportReason, error)
	CreateReportReason(ctx context.Context, reason *models.ReportReason) error
	UpdateReportReason(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteReportReason(ctx context.Context, id string) error
	LogAdminAction(ctx context.Context, adminID, action, entityType, entityID, ipAddress, userAgent string, oldVal, newVal interface{}) error
	GetAuditLogs(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.AdminActionLog, int64, error)
	ReportContent(ctx context.Context, submittedBy string, req ReportContentRequest) error
	CheckHealth(ctx context.Context) (*SystemHealth, error)
	GetPendingReportsCount(ctx context.Context) (int64, error)
	GetJobEngagementMetrics(ctx context.Context, jobID string) (uniqueViews int, savesCount int, err error)
	GetAdminJobs(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.AdminJobReport, int64, error)
	GetJobApplications(ctx context.Context, jobID, status string, page, limit int) ([]*models.Application, int64, error)
	GetChartData(ctx context.Context, period int) (*models.ChartData, error)
	GetUsersByIDs(ctx context.Context, ids []string) (map[string]*models.User, error)
	GetAdminUserIDs(ctx context.Context) ([]string, error)
	GetEntityNames(ctx context.Context, entityType string, ids []string) (map[string]string, error)
}

type ReportContentRequest struct {
	EntityType  string `json:"entity_type" validate:"required"`
	EntityID    string `json:"entity_id" validate:"required"`
	ReasonID    string `json:"reason_id" validate:"required"`
	Description string `json:"description"`
}

type PendingVerificationsResponse struct {
	Verifications []*models.CompanyVerification `json:"verifications"`
	Total         int64                         `json:"total"`
	Page          int                           `json:"page"`
	Limit         int                           `json:"limit"`
	TotalPages    int                           `json:"total_pages"`
}

type VerificationStats struct {
	TotalPending        int64   `json:"total_pending"`
	TotalApproved       int64   `json:"total_approved"`
	TotalRejected       int64   `json:"total_rejected"`
	TotalSubmitted      int64   `json:"total_submitted"`
	AverageWaitTimeDays float64 `json:"average_wait_time_days"`
}

type AdminServiceImpl struct {
	cfg                 *config.Config
	adminRepo           repository.AdminRepository
	notificationService NotificationService
	emailService        EmailService
}

func NewAdminService(cfg *config.Config, adminRepo repository.AdminRepository) AdminService {
	return &AdminServiceImpl{
		cfg:       cfg,
		adminRepo: adminRepo,
	}
}

func toJSONMap(v interface{}) map[string]interface{} {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	if m, ok := v.(map[string]string); ok {
		jm := make(map[string]interface{}, len(m))
		for k, val := range m {
			jm[k] = val
		}
		return jm
	}
	return nil
}

func (s *AdminServiceImpl) SetNotificationService(ns NotificationService) {
	s.notificationService = ns
}

func (s *AdminServiceImpl) SetEmailService(es EmailService) {
	s.emailService = es
}

func (s *AdminServiceImpl) GetPlatformStats(ctx context.Context) (*models.PlatformStats, error) {
	return s.adminRepo.GetPlatformStats(ctx)
}

func (s *AdminServiceImpl) GetUserStats(ctx context.Context) (map[string]int, error) {
	return s.adminRepo.GetUserStats(ctx)
}

func (s *AdminServiceImpl) GetJobStats(ctx context.Context) (map[string]int, error) {
	return s.adminRepo.GetJobStats(ctx)
}

func (s *AdminServiceImpl) GetEngagementStats(ctx context.Context) (map[string]int, error) {
	return s.adminRepo.GetEngagementStats(ctx)
}

func (s *AdminServiceImpl) GetAllUsers(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.UserReport, int64, error) {
	return s.adminRepo.GetAllUsers(ctx, filters, page, limit)
}

func (s *AdminServiceImpl) GetUserDetails(ctx context.Context, userID string) (*models.UserReport, error) {
	return s.adminRepo.GetUserDetails(ctx, userID)
}

func (s *AdminServiceImpl) SuspendUser(ctx context.Context, userID string) error {
	if err := s.adminRepo.SuspendUser(ctx, userID, true); err != nil {
		return err
	}

	// Send notification
	if s.notificationService != nil {
		s.notificationService.NotifyUserSuspended(ctx, userID)
	}

	return nil
}

func (s *AdminServiceImpl) ActivateUser(ctx context.Context, userID string) error {
	if err := s.adminRepo.SuspendUser(ctx, userID, false); err != nil {
		return err
	}

	// Send notification
	if s.notificationService != nil {
		s.notificationService.NotifyUserActivated(ctx, userID)
	}

	return nil
}

func (s *AdminServiceImpl) DeleteUser(ctx context.Context, userID string) error {
	return s.adminRepo.DeleteUser(ctx, userID)
}

func (s *AdminServiceImpl) VerifyUser(ctx context.Context, userID string) error {
	return s.adminRepo.VerifyUser(ctx, userID, true)
}

func (s *AdminServiceImpl) GetModerationQueue(ctx context.Context, entityType, status string, page, limit int, dateFrom, dateTo *time.Time) ([]*models.ModerationQueue, int64, error) {
	return s.adminRepo.GetModerationQueue(ctx, entityType, status, page, limit, dateFrom, dateTo)
}

func (s *AdminServiceImpl) GetModerationItem(ctx context.Context, id string) (*models.ModerationQueue, error) {
	return s.adminRepo.GetModerationItem(ctx, id)
}

func (s *AdminServiceImpl) ApproveContent(ctx context.Context, id, reviewerID string) error {
	item, err := s.adminRepo.GetModerationItem(ctx, id)
	if err != nil {
		return fmt.Errorf("moderation item not found: %w", err)
	}

	if err := s.adminRepo.ApproveModeration(ctx, id, reviewerID); err != nil {
		return err
	}

	// Re-activate the entity if it was disabled
	if item.EntityType == "job" || item.EntityType == "review" {
		_ = s.adminRepo.SetEntityActive(ctx, item.EntityType, item.EntityID, true)
	}

	// Send notification to submitter
	if item.SubmittedBy != "" && s.notificationService != nil {
		s.notificationService.NotifyContentApproved(ctx, item.SubmittedBy, item.EntityType, item.EntityID)
	}

	return nil
}

func (s *AdminServiceImpl) RejectContent(ctx context.Context, id, reviewerID, reason string) error {
	item, err := s.adminRepo.GetModerationItem(ctx, id)
	if err != nil {
		return fmt.Errorf("moderation item not found: %w", err)
	}

	if err := s.adminRepo.RejectModeration(ctx, id, reviewerID, reason); err != nil {
		return err
	}

	// Enforce moderation: disable/hide the content
	_ = s.adminRepo.SetEntityActive(ctx, item.EntityType, item.EntityID, false)

	// Send notification to submitter
	if item.SubmittedBy != "" && s.notificationService != nil {
		s.notificationService.NotifyContentRejected(ctx, item.SubmittedBy, item.EntityType, item.EntityID, reason)
	}

	return nil
}

func (s *AdminServiceImpl) GetSettings(ctx context.Context, category string) ([]*models.SystemSetting, error) {
	return s.adminRepo.GetSettings(ctx, category)
}

func (s *AdminServiceImpl) UpdateSetting(ctx context.Context, key, value string) error {
	return s.adminRepo.UpdateSetting(ctx, key, value)
}

func (s *AdminServiceImpl) CreateSetting(ctx context.Context, key, value, settingType, category, description string, isPublic bool) error {
	return s.adminRepo.SetSetting(ctx, key, value, settingType, category, description, isPublic)
}

// ApproveCompanyVerification approves a company verification request
func (s *AdminServiceImpl) ApproveCompanyVerification(ctx context.Context, adminID, companyID string) error {
	profile, err := s.adminRepo.GetUserDetails(ctx, companyID)
	if err != nil || profile == nil {
		return fmt.Errorf("company not found")
	}

	if err := s.adminRepo.UpdateEmployerVerificationStatus(ctx, companyID, "verified"); err != nil {
		return fmt.Errorf("failed to update company status: %w", err)
	}

	verification, err := s.adminRepo.GetVerificationByCompanyID(ctx, companyID)
	if err != nil {
		verification = &models.CompanyVerification{
			CompanyID:   companyID,
			Status:      "approved",
			VerifiedBy:  &adminID,
			VerifiedAt:  timePtr(time.Now()),
			SubmittedAt: time.Now(),
		}
		if err := s.adminRepo.CreateCompanyVerification(ctx, verification); err != nil {
			fmt.Printf("Warning: Failed to create verification record: %v\n", err)
		}
	} else {
		if err := s.adminRepo.ApproveCompanyVerification(ctx, companyID, adminID); err != nil {
			fmt.Printf("Warning: Failed to update verification record: %v\n", err)
		}
	}

	log := &models.AdminActionLog{
		AdminID:    adminID,
		Action:     "approve_company",
		EntityType: "company",
		EntityID:   companyID,
		NewValue:   map[string]interface{}{"verification_status": "verified"},
	}
	s.adminRepo.LogAction(ctx, log)

	// Send notification and email to company owner
	companyName := profile.CompanyName
	if companyName == "" {
		companyName = profile.FullName
	}
	if s.notificationService != nil {
		s.notificationService.NotifyCompanyVerified(ctx, companyID, companyName)
	}
	if s.emailService != nil {
		s.emailService.SendVerificationApprovedEmail(profile.Email, companyName)
	}

	return nil
}

func (s *AdminServiceImpl) GetJobApplications(ctx context.Context, jobID, status string, page, limit int) ([]*models.Application, int64, error) {
	return s.adminRepo.GetJobApplications(ctx, jobID, status, page, limit)
}

func (s *AdminServiceImpl) RejectCompanyVerification(ctx context.Context, adminID, companyID, reason string) error {
	profile, err := s.adminRepo.GetUserDetails(ctx, companyID)
	if err != nil || profile == nil {
		return fmt.Errorf("company not found")
	}

	if reason == "" {
		return fmt.Errorf("rejection reason is required")
	}

	if err := s.adminRepo.UpdateEmployerVerificationStatus(ctx, companyID, "rejected"); err != nil {
		return fmt.Errorf("failed to update company status: %w", err)
	}

	verification, err := s.adminRepo.GetVerificationByCompanyID(ctx, companyID)
	if err != nil {
		verification = &models.CompanyVerification{
			CompanyID:       companyID,
			Status:          "rejected",
			RejectionReason: reason,
			VerifiedBy:      &adminID,
			VerifiedAt:      timePtr(time.Now()),
			SubmittedAt:     time.Now(),
		}
		if err := s.adminRepo.CreateCompanyVerification(ctx, verification); err != nil {
			fmt.Printf("Warning: Failed to create verification record: %v\n", err)
		}
	} else {
		if err := s.adminRepo.RejectCompanyVerification(ctx, companyID, reason, adminID); err != nil {
			fmt.Printf("Warning: Failed to update verification record: %v\n", err)
		}
	}

	log := &models.AdminActionLog{
		AdminID:    adminID,
		Action:     "reject_company",
		EntityType: "company",
		EntityID:   companyID,
		NewValue: map[string]interface{}{
			"verification_status": "rejected",
			"reason":              reason,
		},
	}
	s.adminRepo.LogAction(ctx, log)

	// Send notification and email to company owner
	companyName := profile.CompanyName
	if companyName == "" {
		companyName = profile.FullName
	}
	if s.notificationService != nil {
		s.notificationService.NotifyCompanyRejected(ctx, companyID, companyName, reason)
	}
	if s.emailService != nil {
		s.emailService.SendVerificationRejectedEmail(profile.Email, companyName, reason)
	}

	return nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func (s *AdminServiceImpl) CheckHealth(ctx context.Context) (*SystemHealth, error) {
	health := &SystemHealth{API: true}

	dbLatency, dbErr := s.adminRepo.PingDatabase(ctx)
	health.Database = dbErr == nil
	health.DatabaseLatency = dbLatency

	if s.cfg.RedisHost != "" {
		addr := net.JoinHostPort(s.cfg.RedisHost, s.cfg.RedisPort)
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			health.Redis = true
			conn.Close()
		}
	}

	if s.cfg.ElasticsearchURL != "" {
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(s.cfg.ElasticsearchURL + "/_cluster/health")
		if err == nil {
			health.Elasticsearch = true
			resp.Body.Close()
		}
	}

	return health, nil
}

func (s *AdminServiceImpl) GetPendingReportsCount(ctx context.Context) (int64, error) {
	return s.adminRepo.GetPendingReportsCount(ctx)
}

func (s *AdminServiceImpl) GetJobEngagementMetrics(ctx context.Context, jobID string) (int, int, error) {
	return s.adminRepo.GetJobEngagementMetrics(ctx, jobID)
}

func (s *AdminServiceImpl) GetReportReasons(ctx context.Context, entityType string) ([]*models.ReportReason, error) {
	return s.adminRepo.GetReportReasons(ctx, entityType)
}

func (s *AdminServiceImpl) CreateReportReason(ctx context.Context, reason *models.ReportReason) error {
	return s.adminRepo.CreateReportReason(ctx, reason)
}

func (s *AdminServiceImpl) UpdateReportReason(ctx context.Context, id string, updates map[string]interface{}) error {
	return s.adminRepo.UpdateReportReason(ctx, id, updates)
}

func (s *AdminServiceImpl) DeleteReportReason(ctx context.Context, id string) error {
	return s.adminRepo.DeleteReportReason(ctx, id)
}

func (s *AdminServiceImpl) GetUsersByIDs(ctx context.Context, ids []string) (map[string]*models.User, error) {
	return s.adminRepo.GetUsersByIDs(ctx, ids)
}

func (s *AdminServiceImpl) GetEntityNames(ctx context.Context, entityType string, ids []string) (map[string]string, error) {
	return s.adminRepo.GetEntityNames(ctx, entityType, ids)
}

func (s *AdminServiceImpl) GetAdminUserIDs(ctx context.Context) ([]string, error) {
	return s.adminRepo.GetAdminUserIDs(ctx)
}

func (s *AdminServiceImpl) LogAdminAction(ctx context.Context, adminID, action, entityType, entityID, ipAddress, userAgent string, oldVal, newVal interface{}) error {
	log := &models.AdminActionLog{
		AdminID:    adminID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValue:   toJSONMap(oldVal),
		NewValue:   toJSONMap(newVal),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
	if err := s.adminRepo.LogAction(ctx, log); err != nil {
		return err
	}
	return nil
}

func (s *AdminServiceImpl) GetAuditLogs(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.AdminActionLog, int64, error) {
	return s.adminRepo.GetActionLogs(ctx, filters, page, limit)
}

func (s *AdminServiceImpl) ReportContent(ctx context.Context, submittedBy string, req ReportContentRequest) error {
	item := &models.ModerationQueue{
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		Status:      "pending",
		Reason:      req.Description,
		SubmittedBy: submittedBy,
	}
	if err := s.adminRepo.AddToModerationQueue(ctx, item); err != nil {
		return err
	}

	// Notify all admin users about the new report
	if s.notificationService != nil {
		adminIDs, err := s.adminRepo.GetAdminUserIDs(ctx)
		if err == nil && len(adminIDs) > 0 {
			for _, adminID := range adminIDs {
				s.notificationService.NotifyAdminNewReport(ctx, adminID, req.EntityType, req.EntityID)
			}
		}
	}

	return nil
}

func (s *AdminServiceImpl) GetChartData(ctx context.Context, period int) (*models.ChartData, error) {
	if period < 1 {
		period = 30
	}
	if period > 365 {
		period = 365
	}
	return s.adminRepo.GetChartData(ctx, period)
}

// GetPendingVerifications - Get all pending verifications with pagination
func (s *AdminServiceImpl) GetAdminJobs(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.AdminJobReport, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.adminRepo.GetAdminJobs(ctx, filters, page, limit)
}

func (s *AdminServiceImpl) GetPendingVerifications(ctx context.Context, page, limit int) (*PendingVerificationsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	verifications, total, err := s.adminRepo.GetPendingVerifications(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &PendingVerificationsResponse{
		Verifications: verifications,
		Total:         total,
		Page:          page,
		Limit:         limit,
		TotalPages:    totalPages,
	}, nil
}
