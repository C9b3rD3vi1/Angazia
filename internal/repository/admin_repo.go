package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	
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
	
	// User management
	GetAllUsers(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.UserReport, int64, error)
	SuspendUser(ctx context.Context, userID string, suspended bool) error
	VerifyUser(ctx context.Context, userID string, verified bool) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserDetails(ctx context.Context, userID string) (*models.UserReport, error)
	
	// Moderation
	AddToModerationQueue(ctx context.Context, item *models.ModerationQueue) error
	GetModerationQueue(ctx context.Context, entityType string, status string, page, limit int) ([]*models.ModerationQueue, int64, error)
	ApproveModeration(ctx context.Context, id string, reviewerID string) error
	RejectModeration(ctx context.Context, id string, reviewerID string, reason string) error
	
	// System settings
	GetSetting(ctx context.Context, key string) (*models.SystemSetting, error)
	GetSettings(ctx context.Context, category string) ([]*models.SystemSetting, error)
	SetSetting(ctx context.Context, key, value, settingType, category, description string, isPublic bool) error
	UpdateSetting(ctx context.Context, key, value string) error
	
	// Report reasons
	GetReportReasons(ctx context.Context, entityType string) ([]*models.ReportReason, error)
	CreateReportReason(ctx context.Context, reason *models.ReportReason) error
	UpdateReportReason(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteReportReason(ctx context.Context, id string) error
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
		SELECT COUNT(DISTINCT user_id) FROM users 
		WHERE last_login_at >= NOW() - INTERVAL '30 days'
	`).Scan(&activeUsers)
	stats.ActiveUsers30Days = int(activeUsers)
	
	// New users
	r.db.WithContext(ctx).Model(&models.User{}).Where("created_at >= NOW() - INTERVAL '7 days'").Count(&stats.NewUsers7Days)
	r.db.WithContext(ctx).Model(&models.User{}).Where("created_at >= NOW() - INTERVAL '30 days'").Count(&stats.NewUsers30Days)
	
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
			ep.full_name,
			emp.company_name,
			(SELECT COUNT(*) FROM jobs WHERE employer_id = users.id) as job_count,
			(SELECT COUNT(*) FROM applications WHERE employee_id = users.id) as application_count,
			(SELECT COUNT(*) FROM moderation_queue WHERE entity_id = users.id) as reports_count
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
	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("users.email ILIKE ? OR ep.full_name ILIKE ? OR emp.company_name ILIKE ?", 
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
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
			ep.full_name,
			emp.company_name
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

func (r *AdminRepositoryImpl) GetModerationQueue(ctx context.Context, entityType string, status string, page, limit int) ([]*models.ModerationQueue, int64, error) {
	var items []*models.ModerationQueue
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.ModerationQueue{})
	
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error
	
	return items, total, err
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