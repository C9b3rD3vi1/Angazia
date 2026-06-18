package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type AdminRepository interface {
	// Action logging
	LogAction(ctx context.Context, log *models.AdminActionLog) error
	GetActionLogs(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.AdminActionLog, int64, error)

	// Platform statistics
	GetPlatformStats(ctx context.Context) (*models.PlatformStats, error)
	GetUserStats(ctx context.Context) (map[string]int, error)
	GetJobStats(ctx context.Context) (map[string]int, error)
	GetEngagementStats(ctx context.Context) (map[string]int, error)
	GetJobEngagementMetrics(ctx context.Context, jobID string) (uniqueViews int, savesCount int, err error)
	GetPendingReportsCount(ctx context.Context) (int64, error)
	PingDatabase(ctx context.Context) (latencyMs int64, err error)

	// User management
	GetAllUsers(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.UserReport, int64, error)
	SuspendUser(ctx context.Context, userID string, suspended bool) error
	VerifyUser(ctx context.Context, userID string, verified bool) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserDetails(ctx context.Context, userID string) (*models.UserReport, error)

	// Moderation
	AddToModerationQueue(ctx context.Context, item *models.ModerationQueue) error
	GetModerationQueue(ctx context.Context, entityType, status string, page, limit int, dateFrom, dateTo *time.Time) ([]*models.ModerationQueue, int64, error)
	GetModerationItem(ctx context.Context, id string) (*models.ModerationQueue, error)
	ApproveModeration(ctx context.Context, id string, reviewerID string) error
	RejectModeration(ctx context.Context, id string, reviewerID, reason string) error

	// System settings
	GetSetting(ctx context.Context, key string) (*models.SystemSetting, error)
	GetSettings(ctx context.Context, category string) ([]*models.SystemSetting, error)
	SetSetting(ctx context.Context, key, value, settingType, category, description string, isPublic bool) error
	UpdateSetting(ctx context.Context, key, value string) error

	// Company verification
	ApproveCompanyVerification(ctx context.Context, companyID, adminID string) error
	RejectCompanyVerification(ctx context.Context, companyID, reason, adminID string) error
	UpdateVerification(ctx context.Context, companyID string, updates map[string]interface{}) error
	UpdateEmployerVerificationStatus(ctx context.Context, companyID, status string) error
	GetOrCreateVerification(ctx context.Context, companyID string) (*models.CompanyVerification, error)
	CreateCompanyVerification(ctx context.Context, verification *models.CompanyVerification) error
	GetPendingVerifications(ctx context.Context, page, limit int) ([]*models.CompanyVerification, int64, error)
	GetVerificationByCompanyID(ctx context.Context, companyID string) (*models.CompanyVerification, error)

	// Report reasons
	GetReportReasons(ctx context.Context, entityType string) ([]*models.ReportReason, error)
	CreateReportReason(ctx context.Context, reason *models.ReportReason) error
	UpdateReportReason(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteReportReason(ctx context.Context, id string) error

	// Dashboard chart data
	GetChartData(ctx context.Context, period int) (*models.ChartData, error)

	// Admin job management
	GetAdminJobs(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.AdminJobReport, int64, error)
	GetJobApplications(ctx context.Context, jobID, status string, page, limit int) ([]*models.Application, int64, error)

	// Batch user fetch
	GetUsersByIDs(ctx context.Context, ids []string) (map[string]*models.User, error)
	GetAdminUserIDs(ctx context.Context) ([]string, error)
}

type AdminRepositoryImpl struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &AdminRepositoryImpl{db: db}
}

func (r *AdminRepositoryImpl) LogAction(ctx context.Context, log *models.AdminActionLog) error {
	log.ID = uuid.New().String()
	log.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *AdminRepositoryImpl) GetActionLogs(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.AdminActionLog, int64, error) {
	var logs []*models.AdminActionLog
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AdminActionLog{}).
		Preload("Admin")

	if adminID, ok := filters["admin_id"].(string); ok && adminID != "" {
		query = query.Where("admin_id = ?", adminID)
	}
	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if entityType, ok := filters["entity_type"].(string); ok && entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if entityID, ok := filters["entity_id"].(string); ok && entityID != "" {
		query = query.Where("entity_id = ?", entityID)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error

	return logs, total, err
}

func (r *AdminRepositoryImpl) GetPlatformStats(ctx context.Context) (*models.PlatformStats, error) {
	stats := &models.PlatformStats{}

	// User statistics
	r.db.WithContext(ctx).Model(&models.User{}).Count(&stats.TotalUsers)
	r.db.WithContext(ctx).Model(&models.EmployeeProfile{}).Count(&stats.TotalCandidates)
	r.db.WithContext(ctx).Model(&models.EmployerProfile{}).Count(&stats.TotalEmployers)
	r.db.WithContext(ctx).Model(&models.EmployerProfile{}).Where("verification_status = ?", "verified").Count(&stats.VerifiedEmployers)

	// Active users in last 30 days
	var activeUsers int64
	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT id) FROM users 
		WHERE last_login_at >= NOW() - INTERVAL '30 days'
	`).Scan(&activeUsers)
	stats.ActiveUsers30Days = int(activeUsers)

	// New users
	r.db.WithContext(ctx).Model(&models.User{}).Where("created_at >= NOW() - INTERVAL '7 days'").Count(&stats.NewUsers7Days)
	r.db.WithContext(ctx).Model(&models.User{}).Where("created_at >= NOW() - INTERVAL '30 days'").Count(&stats.NewUsers30Days)

	r.db.WithContext(ctx).Model(&models.User{}).Where("is_active = ?", false).Count(&stats.SuspendedUsers)
	r.db.WithContext(ctx).Model(&models.User{}).Where("is_verified = ?", false).Count(&stats.UnverifiedUsers)

	// Job statistics
	r.db.WithContext(ctx).Model(&models.Job{}).Count(&stats.TotalJobs)
	r.db.WithContext(ctx).Model(&models.Job{}).Where("is_active = ?", true).Count(&stats.ActiveJobs)
	r.db.WithContext(ctx).Model(&models.Application{}).Count(&stats.TotalApplications)
	r.db.WithContext(ctx).Model(&models.Job{}).Where("posted_at >= NOW() - INTERVAL '7 days'").Count(&stats.JobsPosted7Days)
	r.db.WithContext(ctx).Model(&models.Job{}).Where("posted_at >= NOW() - INTERVAL '30 days'").Count(&stats.JobsPosted30Days)

	// Engagement metrics
	var totalProfileViews, totalJobViews int64
	r.db.WithContext(ctx).Model(&models.CompanyAnalytics{}).Select("COALESCE(SUM(profile_views), 0)").Scan(&totalProfileViews)
	stats.TotalProfileViews = int(totalProfileViews)

	r.db.WithContext(ctx).Model(&models.Job{}).Select("COALESCE(SUM(views_count), 0)").Scan(&totalJobViews)
	stats.TotalJobViews = int(totalJobViews)

	var avgMatchScore float64
	r.db.WithContext(ctx).Model(&models.Application{}).Select("COALESCE(AVG(match_score), 0)").Scan(&avgMatchScore)
	stats.AverageMatchScore = avgMatchScore

	// Calculate growth rates
	var prevUsers, prevJobs int64
	r.db.WithContext(ctx).Model(&models.User{}).Where("created_at < NOW() - INTERVAL '30 days' AND created_at >= NOW() - INTERVAL '60 days'").Count(&prevUsers)
	if prevUsers > 0 {
		stats.UserGrowthRate = float64(stats.NewUsers30Days-prevUsers) / float64(prevUsers) * 100
	}

	r.db.WithContext(ctx).Model(&models.Job{}).Where("posted_at < NOW() - INTERVAL '30 days' AND posted_at >= NOW() - INTERVAL '60 days'").Count(&prevJobs)
	if prevJobs > 0 {
		stats.JobGrowthRate = float64(stats.JobsPosted30Days-prevJobs) / float64(prevJobs) * 100
	}

	// Revenue statistics
	r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("status = ?", "completed").
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalRevenue)

	r.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("status = ?", "active").
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.MRR)

	// Application growth rate
	var currentApps, prevApps int64
	r.db.WithContext(ctx).Model(&models.Application{}).Where("applied_at >= NOW() - INTERVAL '30 days'").Count(&currentApps)
	r.db.WithContext(ctx).Model(&models.Application{}).Where("applied_at >= NOW() - INTERVAL '60 days' AND applied_at < NOW() - INTERVAL '30 days'").Count(&prevApps)
	if prevApps > 0 {
		stats.ApplicationGrowthRate = float64(currentApps-prevApps) / float64(prevApps) * 100
	}

	stats.UpdatedAt = time.Now()

	return stats, nil
}

func (r *AdminRepositoryImpl) GetUserStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	var byRole []struct {
		Role  string
		Count int
	}
	r.db.WithContext(ctx).Model(&models.User{}).Select("role, COUNT(*) as count").Group("role").Scan(&byRole)
	for _, br := range byRole {
		stats["role_"+br.Role] = br.Count
	}

	var byVerification []struct {
		Status string
		Count  int
	}
	r.db.WithContext(ctx).Model(&models.EmployerProfile{}).Select("verification_status, COUNT(*) as count").Group("verification_status").Scan(&byVerification)
	for _, bv := range byVerification {
		stats["verification_"+bv.Status] = bv.Count
	}

	return stats, nil
}

func (r *AdminRepositoryImpl) GetJobStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	var byStatus []struct {
		IsActive bool
		Count    int
	}
	r.db.WithContext(ctx).Model(&models.Job{}).Select("is_active, COUNT(*) as count").Group("is_active").Scan(&byStatus)
	for _, bs := range byStatus {
		status := "inactive"
		if bs.IsActive {
			status = "active"
		}
		stats["status_"+status] = bs.Count
	}

	var byType []struct {
		EmploymentType string
		Count          int
	}
	r.db.WithContext(ctx).Model(&models.Job{}).Select("employment_type, COUNT(*) as count").Group("employment_type").Scan(&byType)
	for _, bt := range byType {
		stats["type_"+bt.EmploymentType] = bt.Count
	}

	return stats, nil
}

func (r *AdminRepositoryImpl) GetEngagementStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	var avgResponseTime float64
	r.db.WithContext(ctx).Model(&models.Application{}).Select("COALESCE(AVG(EXTRACT(EPOCH FROM (reviewed_at - applied_at))/86400), 0)").Scan(&avgResponseTime)
	stats["avg_response_days"] = int(avgResponseTime)

	var conversionRate float64
	r.db.WithContext(ctx).Raw(`
		SELECT 
			COUNT(CASE WHEN status = 'hired' THEN 1 END)::float / COUNT(*) * 100
		FROM applications
	`).Scan(&conversionRate)
	stats["conversion_rate"] = int(conversionRate)

	return stats, nil
}

func (r *AdminRepositoryImpl) GetJobEngagementMetrics(ctx context.Context, jobID string) (int, int, error) {
	var uniqueViews int64
	r.db.WithContext(ctx).Model(&models.JobView{}).
		Where("job_id = ?", jobID).
		Select("COALESCE(COUNT(DISTINCT employee_id), 0)").Scan(&uniqueViews)

	var savesCount int64
	r.db.WithContext(ctx).Model(&models.SavedJob{}).
		Where("job_id = ?", jobID).
		Count(&savesCount)

	return int(uniqueViews), int(savesCount), nil
}

func (r *AdminRepositoryImpl) GetPendingReportsCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.ModerationQueue{}).
		Where("status = ?", "pending").
		Count(&count).Error
	return count, err
}

func (r *AdminRepositoryImpl) PingDatabase(ctx context.Context) (int64, error) {
	start := time.Now()
	sqlDB, err := r.db.DB()
	if err != nil {
		return 0, err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return 0, err
	}
	return time.Since(start).Milliseconds(), nil
}

func (r *AdminRepositoryImpl) GetAllUsers(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.UserReport, int64, error) {
	var users []*models.UserReport
	var total int64

	query := r.db.WithContext(ctx).Table("users").
		Select(`
			users.id,
			users.email,
			users.role,
			users.is_verified,
			users.is_active,
			users.created_at,
			users.last_login_at,
			users.avatar_url,
			ep.full_name,
			emp.company_name,
			emp.company_logo,
			emp.verification_status,
			(SELECT COUNT(*) FROM jobs WHERE employer_id = users.id) as job_count,
			(SELECT COUNT(*) FROM applications WHERE employee_id = users.id) as application_count,
			(SELECT COUNT(*) FROM moderation_queue WHERE entity_id = users.id) as reports_count,
			(SELECT COALESCE(jsonb_array_length(cv.documents), 0) FROM company_verifications cv WHERE cv.company_id = users.id) as document_count
		`).
		Joins("LEFT JOIN employee_profiles ep ON ep.user_id = users.id").
		Joins("LEFT JOIN employer_profiles emp ON emp.user_id = users.id")

	if role, ok := filters["role"].(string); ok && role != "" {
		query = query.Where("users.role = ?", role)
	}
	if isActive, ok := filters["is_active"].(bool); ok {
		query = query.Where("users.is_active = ?", isActive)
	}
	if isVerified, ok := filters["is_verified"].(bool); ok {
		query = query.Where("users.is_verified = ?", isVerified)
	}
	if verificationStatus, ok := filters["verification_status"].(string); ok && verificationStatus != "" {
		if verificationStatus == "unverified" {
			query = query.Where("(emp.verification_status IS NULL OR emp.verification_status = '')")
		} else {
			query = query.Where("emp.verification_status = ?", verificationStatus)
		}
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("users.email ILIKE ? OR ep.full_name ILIKE ? OR emp.company_name ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if hasDocs, ok := filters["has_documents"].(string); ok && hasDocs != "" {
		if hasDocs == "yes" {
			query = query.Where("EXISTS (SELECT 1 FROM company_verifications cv WHERE cv.company_id = users.id AND jsonb_array_length(cv.documents) > 0)")
		} else if hasDocs == "no" {
			query = query.Where("NOT EXISTS (SELECT 1 FROM company_verifications cv WHERE cv.company_id = users.id AND jsonb_array_length(cv.documents) > 0)")
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Order("users.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&users).Error

	return users, total, err
}

func (r *AdminRepositoryImpl) SuspendUser(ctx context.Context, userID string, suspended bool) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_active", !suspended).Error
}

func (r *AdminRepositoryImpl) VerifyUser(ctx context.Context, userID string, verified bool) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_verified", verified).Error
}

func (r *AdminRepositoryImpl) DeleteUser(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", userID).Error
}

func (r *AdminRepositoryImpl) GetUserDetails(ctx context.Context, userID string) (*models.UserReport, error) {
	var user models.UserReport
	err := r.db.WithContext(ctx).Table("users").
		Select(`
			users.id,
			users.email,
			users.role,
			users.is_verified,
			users.is_active,
			users.created_at,
			users.last_login_at,
			users.avatar_url,
			ep.full_name,
			emp.company_name,
			emp.company_logo,
			emp.verification_status
		`).
		Joins("LEFT JOIN employee_profiles ep ON ep.user_id = users.id").
		Joins("LEFT JOIN employer_profiles emp ON emp.user_id = users.id").
		Where("users.id = ?", userID).
		Scan(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AdminRepositoryImpl) AddToModerationQueue(ctx context.Context, item *models.ModerationQueue) error {
	item.ID = uuid.New().String()
	item.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *AdminRepositoryImpl) GetModerationQueue(ctx context.Context, entityType, status string, page, limit int, dateFrom, dateTo *time.Time) ([]*models.ModerationQueue, int64, error) {
	var items []*models.ModerationQueue
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ModerationQueue{})

	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if dateFrom != nil {
		query = query.Where("created_at >= ?", dateFrom)
	}
	if dateTo != nil {
		query = query.Where("created_at <= ?", dateTo)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error

	return items, total, err
}

func (r *AdminRepositoryImpl) GetModerationItem(ctx context.Context, id string) (*models.ModerationQueue, error) {
	var item models.ModerationQueue
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *AdminRepositoryImpl) ApproveModeration(ctx context.Context, id string, reviewerID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.ModerationQueue{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      "approved",
			"reviewed_by": reviewerID,
			"reviewed_at": &now,
		}).Error
}

func (r *AdminRepositoryImpl) RejectModeration(ctx context.Context, id string, reviewerID string, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.ModerationQueue{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      "rejected",
			"reason":      reason,
			"reviewed_by": reviewerID,
			"reviewed_at": &now,
		}).Error
}

func (r *AdminRepositoryImpl) GetSetting(ctx context.Context, key string) (*models.SystemSetting, error) {
	var setting models.SystemSetting
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *AdminRepositoryImpl) GetSettings(ctx context.Context, category string) ([]*models.SystemSetting, error) {
	var settings []*models.SystemSetting
	query := r.db.WithContext(ctx)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	err := query.Order("key ASC").Find(&settings).Error
	return settings, err
}

func (r *AdminRepositoryImpl) SetSetting(ctx context.Context, key, value, settingType, category, description string, isPublic bool) error {
	var existing models.SystemSetting
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&existing).Error

	if err == nil {
		return r.UpdateSetting(ctx, key, value)
	}

	setting := &models.SystemSetting{
		Key:         key,
		Value:       value,
		Type:        settingType,
		Category:    category,
		Description: description,
		IsPublic:    isPublic,
	}
	return r.db.WithContext(ctx).Create(setting).Error
}

func (r *AdminRepositoryImpl) UpdateSetting(ctx context.Context, key, value string) error {
	return r.db.WithContext(ctx).
		Model(&models.SystemSetting{}).
		Where("key = ?", key).
		Update("value", value).Error
}

// RejectCompanyVerification rejects a company verification request (updates regardless of current status)
func (r *AdminRepositoryImpl) RejectCompanyVerification(ctx context.Context, companyID, reason, adminID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.CompanyVerification{}).
		Where("company_id = ?", companyID).
		Updates(map[string]interface{}{
			"status":           "rejected",
			"rejection_reason": reason,
			"verified_by":      adminID,
			"verified_at":      now,
			"updated_at":       now,
		}).Error
}

// ApproveCompanyVerification approves a company verification request (updates regardless of current status)
func (r *AdminRepositoryImpl) ApproveCompanyVerification(ctx context.Context, companyID, adminID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.CompanyVerification{}).
		Where("company_id = ?", companyID).
		Updates(map[string]interface{}{
			"status":      "approved",
			"verified_by": adminID,
			"verified_at": now,
			"updated_at":  now,
		}).Error
}

// UpdateEmployerVerificationStatus updates the verification status on employer profile
// Uses upsert to handle cases where the employer_profiles record may not exist yet
func (r *AdminRepositoryImpl) UpdateEmployerVerificationStatus(ctx context.Context, companyID, status string) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO employer_profiles (user_id, company_name, verification_status, created_at, updated_at)
		VALUES (?, '', ?, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET verification_status = ?, updated_at = NOW()
	`, companyID, status, status).Error
}

// GetOrCreateVerification gets existing verification or creates a new one
func (r *AdminRepositoryImpl) GetOrCreateVerification(ctx context.Context, companyID string) (*models.CompanyVerification, error) {
	var verification models.CompanyVerification
	err := r.db.WithContext(ctx).Where("company_id = ?", companyID).First(&verification).Error

	if err == nil {
		return &verification, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new verification record
		verification = models.CompanyVerification{
			CompanyID:   companyID,
			Status:      "pending",
			SubmittedAt: time.Now(),
		}
		if err := r.db.WithContext(ctx).Create(&verification).Error; err != nil {
			return nil, err
		}
		return &verification, nil
	}

	return nil, err
}

// CreateCompanyVerification creates a new company verification record
func (r *AdminRepositoryImpl) CreateCompanyVerification(ctx context.Context, verification *models.CompanyVerification) error {
	verification.ID = uuid.New().String()
	verification.SubmittedAt = time.Now()
	return r.db.WithContext(ctx).Create(verification).Error
}

func (r *AdminRepositoryImpl) GetPendingVerifications(ctx context.Context, page, limit int) ([]*models.CompanyVerification, int64, error) {
	var verifications []*models.CompanyVerification
	var total int64
	query := r.db.WithContext(ctx).Model(&models.CompanyVerification{}).
		Where("status = ?", "pending").
		Preload("Company").
		Preload("Company.User")
	query.Count(&total)
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("submitted_at DESC").Find(&verifications).Error
	return verifications, total, err
}

// UpdateVerification updates a company verification record
func (r *AdminRepositoryImpl) UpdateVerification(ctx context.Context, companyID string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&models.CompanyVerification{}).
		Where("company_id = ?", companyID).
		Updates(updates).Error
}

func (r *AdminRepositoryImpl) GetVerificationByCompanyID(ctx context.Context, companyID string) (*models.CompanyVerification, error) {
	var verification models.CompanyVerification
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

func (r *AdminRepositoryImpl) GetReportReasons(ctx context.Context, entityType string) ([]*models.ReportReason, error) {
	var reasons []*models.ReportReason
	query := r.db.WithContext(ctx).Where("is_active = ?", true)
	if entityType != "" {
		query = query.Where("entity_type = ? OR entity_type IS NULL", entityType)
	}
	err := query.Order("sort_order ASC").Find(&reasons).Error
	return reasons, err
}

func (r *AdminRepositoryImpl) CreateReportReason(ctx context.Context, reason *models.ReportReason) error {
	reason.ID = uuid.New().String()
	return r.db.WithContext(ctx).Create(reason).Error
}

func (r *AdminRepositoryImpl) UpdateReportReason(ctx context.Context, id string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&models.ReportReason{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *AdminRepositoryImpl) DeleteReportReason(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.ReportReason{}, "id = ?", id).Error
}

func (r *AdminRepositoryImpl) GetChartData(ctx context.Context, period int) (*models.ChartData, error) {
	chartData := &models.ChartData{}

	var userPoints []models.ChartDataPoint
	r.db.WithContext(ctx).Raw(`
		SELECT to_char(d.date, 'YYYY-MM-DD') as date, COALESCE(COUNT(u.id), 0) as count
		FROM generate_series(
			CURRENT_DATE - ($1 || ' days')::INTERVAL,
			CURRENT_DATE,
			'1 day'::INTERVAL
		) d(date)
		LEFT JOIN users u ON DATE(u.created_at) = d.date
		GROUP BY d.date
		ORDER BY d.date
	`, fmt.Sprintf("%d", period)).Scan(&userPoints)
	chartData.UserGrowth = userPoints

	var jobPoints []models.ChartDataPoint
	r.db.WithContext(ctx).Raw(`
		SELECT to_char(d.date, 'YYYY-MM-DD') as date, COALESCE(COUNT(j.id), 0) as count
		FROM generate_series(
			CURRENT_DATE - ($1 || ' days')::INTERVAL,
			CURRENT_DATE,
			'1 day'::INTERVAL
		) d(date)
		LEFT JOIN jobs j ON DATE(j.posted_at) = d.date
		GROUP BY d.date
		ORDER BY d.date
	`, fmt.Sprintf("%d", period)).Scan(&jobPoints)
	chartData.JobPostings = jobPoints

	var appPoints []models.ChartDataPoint
	r.db.WithContext(ctx).Raw(`
		SELECT to_char(d.date, 'YYYY-MM-DD') as date, COALESCE(COUNT(a.id), 0) as count
		FROM generate_series(
			CURRENT_DATE - ($1 || ' days')::INTERVAL,
			CURRENT_DATE,
			'1 day'::INTERVAL
		) d(date)
		LEFT JOIN applications a ON DATE(a.applied_at) = d.date
		GROUP BY d.date
		ORDER BY d.date
	`, fmt.Sprintf("%d", period)).Scan(&appPoints)
	chartData.Applications = appPoints

	return chartData, nil
}

func (r *AdminRepositoryImpl) GetUsersByIDs(ctx context.Context, ids []string) (map[string]*models.User, error) {
	var users []*models.User
	err := r.db.WithContext(ctx).
		Preload("EmployeeProfile").
		Preload("EmployerProfile").
		Where("id IN ?", ids).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]*models.User, len(users))
	for _, u := range users {
		result[u.ID] = u
	}
	return result, nil
}

func (r *AdminRepositoryImpl) GetAdminUserIDs(ctx context.Context) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("role = ?", "admin").
		Pluck("id", &ids).Error
	return ids, err
}

func (r *AdminRepositoryImpl) GetJobApplications(ctx context.Context, jobID, status string, page, limit int) ([]*models.Application, int64, error) {
	var apps []*models.Application
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Application{}).
		Preload("Employee").
		Where("job_id = ?", jobID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Order("applied_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&apps).Error

	return apps, total, err
}

func (r *AdminRepositoryImpl) GetAdminJobs(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.AdminJobReport, int64, error) {
	var jobs []*models.AdminJobReport
	var total int64

	query := r.db.WithContext(ctx).Table("jobs").
		Select(`
			jobs.id,
			jobs.title,
			jobs.employer_id,
			COALESCE(emp.company_name, '') as company_name,
			COALESCE(emp.company_logo, '') as company_logo,
			COALESCE(emp.user_id::text, '') as company_id,
			COALESCE(u.email, '') as email,
			CASE
				WHEN EXISTS (SELECT 1 FROM moderation_queue mq WHERE mq.entity_id = jobs.id AND mq.entity_type = 'job' AND mq.status = 'pending') THEN 'pending'
				WHEN jobs.is_active THEN 'active'
				ELSE 'closed'
			END as status,
			COALESCE(jobs.location, '') as location,
			COALESCE(jobs.employment_type, '') as employment_type,
			COALESCE(jobs.experience_level, '') as experience_level,
			jobs.salary_min,
			jobs.salary_max,
			COALESCE(jobs.salary_currency, 'KES') as salary_currency,
			jobs.is_active,
			jobs.applications_count,
			jobs.views_count,
			to_char(jobs.posted_at, 'YYYY-MM-DD') as posted_at,
			to_char(jobs.posted_at, 'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"') as created_at,
			COALESCE(to_char(jobs.expires_at, 'YYYY-MM-DD'), '') as expires_at
		`).
		Joins("LEFT JOIN employer_profiles emp ON emp.user_id = jobs.employer_id").
		Joins("LEFT JOIN users u ON u.id = jobs.employer_id")

	if status, ok := filters["status"].(string); ok && status != "" {
		switch status {
		case "active":
			query = query.Where("jobs.is_active = ?", true).
				Where("NOT EXISTS (SELECT 1 FROM moderation_queue mq WHERE mq.entity_id = jobs.id AND mq.entity_type = 'job' AND mq.status = 'pending')")
		case "closed":
			query = query.Where("jobs.is_active = ?", false)
		case "pending":
			query = query.Where("EXISTS (SELECT 1 FROM moderation_queue mq WHERE mq.entity_id = jobs.id AND mq.entity_type = 'job' AND mq.status = 'pending')")
		}
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("jobs.title ILIKE ? OR emp.company_name ILIKE ? OR u.email ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if employmentType, ok := filters["employment_type"].(string); ok && employmentType != "" {
		query = query.Where("jobs.employment_type = ?", employmentType)
	}
	if experienceLevel, ok := filters["experience_level"].(string); ok && experienceLevel != "" {
		query = query.Where("jobs.experience_level = ?", experienceLevel)
	}
	if location, ok := filters["location"].(string); ok && location != "" {
		query = query.Where("jobs.location ILIKE ?", "%"+location+"%")
	}
	if employerID, ok := filters["employer_id"].(string); ok && employerID != "" {
		query = query.Where("jobs.employer_id = ?", employerID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Order("jobs.posted_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&jobs).Error

	return jobs, total, err
}
