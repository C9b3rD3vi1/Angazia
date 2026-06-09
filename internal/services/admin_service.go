package services

import (
	"context"
	"encoding/json"
	"fmt"
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

type SearchServiceImpl struct {
	cfg         *config.Config
	searchRepo  repository.SearchRepository
	jobRepo     repository.JobRepository
	userRepo    repository.UserRepository
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
	for i, job := range jobs {
		results[i] = models.SearchResult{
			ID:          job.ID,
			Type:        "job",
			Title:       job.Title,
			Description: job.Description,
			Score:       1.0,
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
	
	return &models.SearchResponse{
		Results:      results,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
		Facets:       *facets,
		SearchTimeMs: searchTimeMs,
	}, nil
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
	for i, candidate := range candidates {
		results[i] = models.SearchResult{
			ID:          candidate.UserID,
			Type:        "candidate",
			Title:       candidate.FullName,
			Description: candidate.Headline,
			Score:       1.0,
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
	for i, company := range companies {
		results[i] = models.SearchResult{
			ID:          company.UserID,
			Type:        "company",
			Title:       company.CompanyName,
			Description: company.CompanyDescription,
			Score:       1.0,
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
	filtersJSON, _ := json.Marshal(filters)
	
	searchQuery := &models.SearchQuery{
		UserID:       userID,
		Query:        query,
		Filters:      filters,
		EntityType:   entityType,
		ResultsCount: resultsCount,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}
	_ = filtersJSON
	
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
	
	var suggestions []string
	
	switch entityType {
	case "job":
		var titles []string
		s.searchRepo.SearchJobs(ctx, models.SearchFilters{Keywords: prefix}, 1, limit)
		_ = titles
		// Simplified - would query actual job titles
		suggestions = []string{prefix + " Developer", prefix + " Engineer", "Senior " + prefix}
	case "skill":
		suggestions = []string{prefix, prefix + " Development", prefix + " Programming", "Advanced " + prefix}
	default:
		suggestions = []string{prefix}
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
	GetModerationQueue(ctx context.Context, entityType, status string, page, limit int) ([]*models.ModerationQueue, int64, error)
	ApproveContent(ctx context.Context, id, reviewerID string) error
	RejectContent(ctx context.Context, id, reviewerID, reason string) error
	GetSettings(ctx context.Context, category string) ([]*models.SystemSetting, error)
	UpdateSetting(ctx context.Context, key, value string) error
	ApproveCompanyVerification(ctx context.Context, adminID, companyID string) error
	RejectCompanyVerification(ctx context.Context, adminID, companyID, reason string) error
	GetPendingVerifications(ctx context.Context, page, limit int) (*PendingVerificationsResponse, error)
	GetReportReasons(ctx context.Context, entityType string) ([]*models.ReportReason, error)
	GetAuditLogs(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.AdminActionLog, int64, error)
	ReportContent(ctx context.Context, submittedBy string, req ReportContentRequest) error
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
    TotalPending   int64 `json:"total_pending"`
    TotalApproved  int64 `json:"total_approved"`
    TotalRejected  int64 `json:"total_rejected"`
    TotalSubmitted int64 `json:"total_submitted"`
    AverageWaitTimeDays float64 `json:"average_wait_time_days"`
}

type AdminServiceImpl struct {
	cfg       *config.Config
	adminRepo repository.AdminRepository
}

func NewAdminService(cfg *config.Config, adminRepo repository.AdminRepository) AdminService {
	return &AdminServiceImpl{
		cfg:       cfg,
		adminRepo: adminRepo,
	}
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
	return s.adminRepo.SuspendUser(ctx, userID, true)
}

func (s *AdminServiceImpl) ActivateUser(ctx context.Context, userID string) error {
	return s.adminRepo.SuspendUser(ctx, userID, false)
}

func (s *AdminServiceImpl) DeleteUser(ctx context.Context, userID string) error {
	return s.adminRepo.DeleteUser(ctx, userID)
}

func (s *AdminServiceImpl) VerifyUser(ctx context.Context, userID string) error {
	return s.adminRepo.VerifyUser(ctx, userID, true)
}

func (s *AdminServiceImpl) GetModerationQueue(ctx context.Context, entityType, status string, page, limit int) ([]*models.ModerationQueue, int64, error) {
	return s.adminRepo.GetModerationQueue(ctx, entityType, status, page, limit)
}

func (s *AdminServiceImpl) ApproveContent(ctx context.Context, id, reviewerID string) error {
	return s.adminRepo.ApproveModeration(ctx, id, reviewerID)
}

func (s *AdminServiceImpl) RejectContent(ctx context.Context, id, reviewerID, reason string) error {
	return s.adminRepo.RejectModeration(ctx, id, reviewerID, reason)
}

func (s *AdminServiceImpl) GetSettings(ctx context.Context, category string) ([]*models.SystemSetting, error) {
	return s.adminRepo.GetSettings(ctx, category)
}

func (s *AdminServiceImpl) UpdateSetting(ctx context.Context, key, value string) error {
	return s.adminRepo.UpdateSetting(ctx, key, value)
}

// ApproveCompanyVerification approves a company verification request
func (s *AdminServiceImpl) ApproveCompanyVerification(ctx context.Context, adminID, companyID string) error {
    // First, check if employer profile exists
    profile, err := s.adminRepo.GetUserDetails(ctx, companyID)
    if err != nil || profile == nil {
        return fmt.Errorf("company not found")
    }
    
    // Directly update employer profile verification status
    if err := s.adminRepo.UpdateEmployerVerificationStatus(ctx, companyID, "verified"); err != nil {
        return fmt.Errorf("failed to update company status: %w", err)
    }
    
    // Update or create verification record
    verification, err := s.adminRepo.GetVerificationByCompanyID(ctx, companyID)
    if err != nil {
        // Create new verification record if doesn't exist
        verification = &models.CompanyVerification{
            CompanyID:   companyID,
            Status:      "approved",
            VerifiedBy:  &adminID,
            VerifiedAt:  timePtr(time.Now()),
            SubmittedAt: time.Now(),
        }
        if err := s.adminRepo.CreateCompanyVerification(ctx, verification); err != nil {
            // Log error but don't fail since employer profile is already updated
            fmt.Printf("Warning: Failed to create verification record: %v\n", err)
        }
    } else {
        // Update existing verification
        if err := s.adminRepo.ApproveCompanyVerification(ctx, companyID, adminID); err != nil {
            // Log error but don't fail since employer profile is already updated
            fmt.Printf("Warning: Failed to update verification record: %v\n", err)
        }
    }
    
    // Log action
    log := &models.AdminActionLog{
        AdminID:    adminID,
        Action:     "approve_company",
        EntityType: "company",
        EntityID:   companyID,
        NewValue:   map[string]interface{}{"verification_status": "verified"},
    }
    s.adminRepo.LogAction(ctx, log)
    
    return nil
}

// RejectCompanyVerification rejects a company verification request
// RejectCompanyVerification rejects a company verification request
func (s *AdminServiceImpl) RejectCompanyVerification(ctx context.Context, adminID, companyID, reason string) error {
    // First, check if employer profile exists
    profile, err := s.adminRepo.GetUserDetails(ctx, companyID)
    if err != nil || profile == nil {
        return fmt.Errorf("company not found")
    }
    
    // Validate reason
    if reason == "" {
        return fmt.Errorf("rejection reason is required")
    }
    
    // Directly update employer profile verification status
    if err := s.adminRepo.UpdateEmployerVerificationStatus(ctx, companyID, "rejected"); err != nil {
        return fmt.Errorf("failed to update company status: %w", err)
    }
    
    // Update or create verification record
    verification, err := s.adminRepo.GetVerificationByCompanyID(ctx, companyID)
    if err != nil {
        // Create new verification record if doesn't exist
        verification = &models.CompanyVerification{
            CompanyID:        companyID,
            Status:           "rejected",
            RejectionReason:  reason,
            VerifiedBy:       &adminID,
            VerifiedAt:       timePtr(time.Now()),
            SubmittedAt:      time.Now(),
        }
        if err := s.adminRepo.CreateCompanyVerification(ctx, verification); err != nil {
            // Log error but don't fail since employer profile is already updated
            fmt.Printf("Warning: Failed to create verification record: %v\n", err)
        }
    } else {
        // Update existing verification
        if err := s.adminRepo.RejectCompanyVerification(ctx, companyID, reason, adminID); err != nil {
            // Log error but don't fail since employer profile is already updated
            fmt.Printf("Warning: Failed to update verification record: %v\n", err)
        }
    }
    
    // Log action
    log := &models.AdminActionLog{
        AdminID:    adminID,
        Action:     "reject_company",
        EntityType: "company",
        EntityID:   companyID,
        NewValue: map[string]interface{}{
            "verification_status": "rejected",
            "reason": reason,
        },
    }
    s.adminRepo.LogAction(ctx, log)
    
    return nil
}

func timePtr(t time.Time) *time.Time {
    return &t
}


func (s *AdminServiceImpl) GetReportReasons(ctx context.Context, entityType string) ([]*models.ReportReason, error) {
	return s.adminRepo.GetReportReasons(ctx, entityType)
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
	return s.adminRepo.AddToModerationQueue(ctx, item)
}


// GetPendingVerifications - Get all pending verifications with pagination
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

