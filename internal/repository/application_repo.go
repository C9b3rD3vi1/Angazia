package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type ApplicationRepository interface {
	// CRUD operations
	Create(ctx context.Context, application *models.Application) error
	GetByID(ctx context.Context, id string) (*models.Application, error)
	GetByIDWithDetails(ctx context.Context, id string) (*models.Application, error)
	Update(ctx context.Context, application *models.Application) error
	UpdateStatus(ctx context.Context, id string, status string, employerNotes string) error
	Delete(ctx context.Context, id string) error
	
	// Application listings
	ListByJob(ctx context.Context, jobID string, status string, page, limit int) ([]*models.Application, int64, error)
	ListByEmployee(ctx context.Context, employeeID string, page, limit int) ([]*models.Application, int64, error)
	ListByEmployer(ctx context.Context, employerID string, status string, page, limit int) ([]*models.Application, int64, error)
	ListPending(ctx context.Context, employerID string, page, limit int) ([]*models.Application, int64, error)
	ListShortlisted(ctx context.Context, employerID string, page, limit int) ([]*models.Application, int64, error)
	
	// Status updates
	Shortlist(ctx context.Context, id string, employerNotes string) error
	Reject(ctx context.Context, id string, employerNotes string) error
	MarkAsViewed(ctx context.Context, id string) error
	ScheduleInterview(ctx context.Context, id string, interviewDate time.Time, interviewType string) error
	
	// Checks
	HasApplied(ctx context.Context, jobID, employeeID string) (bool, error)
	GetApplicationCount(ctx context.Context, jobID string) (int64, error)
	GetEmployeeApplicationStats(ctx context.Context, employeeID string) (*ApplicationStats, error)
	GetEmployerApplicationStats(ctx context.Context, employerID string) (*ApplicationStats, error)
	
	// Bulk operations
	BulkUpdateStatus(ctx context.Context, ids []string, status string) error
	DeleteByJob(ctx context.Context, jobID string) error

	// Raw database access
	DB() *gorm.DB
}

type ApplicationStats struct {
	TotalApplications   int64            `json:"total_applications"`
	PendingCount        int64            `json:"pending_count"`
	ViewedCount         int64            `json:"viewed_count"`
	ShortlistedCount    int64            `json:"shortlisted_count"`
	RejectedCount       int64            `json:"rejected_count"`
	HiredCount          int64            `json:"hired_count"`
	WithdrawnCount      int64            `json:"withdrawn_count"`
	AverageMatchScore   float64          `json:"average_match_score"`
	ApplicationsByDate  map[string]int64 `json:"applications_by_date"`
}

type ApplicationRepositoryImpl struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &ApplicationRepositoryImpl{db: db}
}

func (r *ApplicationRepositoryImpl) DB() *gorm.DB {
	return r.db
}

func (r *ApplicationRepositoryImpl) Create(ctx context.Context, application *models.Application) error {
	application.ID = uuid.New().String()
	application.AppliedAt = time.Now()
	application.Status = "pending"
	
	return r.db.WithContext(ctx).Create(application).Error
}

func (r *ApplicationRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Application, error) {
	var application models.Application
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&application).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &application, nil
}

func (r *ApplicationRepositoryImpl) GetByIDWithDetails(ctx context.Context, id string) (*models.Application, error) {
	var application models.Application
	err := r.db.WithContext(ctx).
		Preload("Job").
		Preload("Job.Employer").
		Preload("Employee").
		Preload("Employee.User").
		Where("id = ?", id).
		First(&application).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &application, nil
}

func (r *ApplicationRepositoryImpl) Update(ctx context.Context, application *models.Application) error {
	return r.db.WithContext(ctx).Save(application).Error
}

func (r *ApplicationRepositoryImpl) UpdateStatus(ctx context.Context, id string, status string, employerNotes string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":        status,
		"employer_notes": employerNotes,
		"reviewed_at":   &now,
		"updated_at":    now,
	}
	
	// Add responded_at if final status
	if status == "shortlisted" || status == "rejected" || status == "hired" {
		updates["responded_at"] = &now
	}
	
	return r.db.WithContext(ctx).
		Model(&models.Application{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *ApplicationRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.Application{}).Error
}

func (r *ApplicationRepositoryImpl) ListByJob(ctx context.Context, jobID string, status string, page, limit int) ([]*models.Application, int64, error) {
	var applications []*models.Application
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Application{}).
		Where("job_id = ?", jobID)
	
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("Employee").
		Preload("Employee.User").
		Order("match_score DESC, applied_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&applications).Error
	
	return applications, total, err
}

func (r *ApplicationRepositoryImpl) ListByEmployee(ctx context.Context, employeeID string, page, limit int) ([]*models.Application, int64, error) {
	var applications []*models.Application
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Application{}).
		Where("employee_id = ?", employeeID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("Job").
		Preload("Job.Employer").
		Order("applied_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&applications).Error
	
	return applications, total, err
}

func (r *ApplicationRepositoryImpl) ListByEmployer(ctx context.Context, employerID string, status string, page, limit int) ([]*models.Application, int64, error) {
	var applications []*models.Application
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Application{}).
		Joins("JOIN jobs ON applications.job_id = jobs.id").
		Where("jobs.employer_id = ?", employerID)
	
	if status != "" && status != "all" {
		query = query.Where("applications.status = ?", status)
	}
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("Job").
		Preload("Employee").
		Preload("Employee.User").
		Order("applications.match_score DESC, applications.applied_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&applications).Error
	
	return applications, total, err
}

func (r *ApplicationRepositoryImpl) ListPending(ctx context.Context, employerID string, page, limit int) ([]*models.Application, int64, error) {
	return r.ListByEmployer(ctx, employerID, "pending", page, limit)
}

func (r *ApplicationRepositoryImpl) ListShortlisted(ctx context.Context, employerID string, page, limit int) ([]*models.Application, int64, error) {
	return r.ListByEmployer(ctx, employerID, "shortlisted", page, limit)
}

func (r *ApplicationRepositoryImpl) Shortlist(ctx context.Context, id string, employerNotes string) error {
	return r.UpdateStatus(ctx, id, "shortlisted", employerNotes)
}

func (r *ApplicationRepositoryImpl) Reject(ctx context.Context, id string, employerNotes string) error {
	return r.UpdateStatus(ctx, id, "rejected", employerNotes)
}

func (r *ApplicationRepositoryImpl) MarkAsViewed(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Application{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"viewed_at": &now,
			"updated_at": now,
		}).Error
}

func (r *ApplicationRepositoryImpl) ScheduleInterview(ctx context.Context, id string, interviewDate time.Time, interviewType string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"interview_date": &interviewDate,
		"interview_type": interviewType,
		"status":         "interview",
		"reviewed_at":    &now,
		"responded_at":   &now,
		"updated_at":     now,
	}
	
	return r.db.WithContext(ctx).
		Model(&models.Application{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *ApplicationRepositoryImpl) HasApplied(ctx context.Context, jobID, employeeID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Application{}).
		Where("job_id = ? AND employee_id = ?", jobID, employeeID).
		Count(&count).Error
	return count > 0, err
}

func (r *ApplicationRepositoryImpl) GetApplicationCount(ctx context.Context, jobID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Application{}).
		Where("job_id = ?", jobID).
		Count(&count).Error
	return count, err
}

func (r *ApplicationRepositoryImpl) GetEmployeeApplicationStats(ctx context.Context, employeeID string) (*ApplicationStats, error) {
	stats := &ApplicationStats{
		ApplicationsByDate: make(map[string]int64),
	}
	
	// Get counts by status
	var results []struct {
		Status string
		Count  int64
	}
	
	err := r.db.WithContext(ctx).
		Model(&models.Application{}).
		Select("status, COUNT(*) as count").
		Where("employee_id = ?", employeeID).
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
		case "viewed":
			stats.ViewedCount = r.Count
		case "shortlisted":
			stats.ShortlistedCount = r.Count
		case "rejected":
			stats.RejectedCount = r.Count
		case "hired":
			stats.HiredCount = r.Count
		case "withdrawn":
			stats.WithdrawnCount = r.Count
		}
	}
	
	// Get average match score
	var avgScore float64
	r.db.WithContext(ctx).
		Model(&models.Application{}).
		Where("employee_id = ? AND match_score > 0", employeeID).
		Select("AVG(match_score)").
		Scan(&avgScore)
	stats.AverageMatchScore = avgScore
	
	// Get applications by date (last 30 days)
	var dateResults []struct {
		Date  time.Time
		Count int64
	}
	
	r.db.WithContext(ctx).
		Model(&models.Application{}).
		Select("DATE(applied_at) as date, COUNT(*) as count").
		Where("employee_id = ? AND applied_at >= ?", employeeID, time.Now().AddDate(0, 0, -30)).
		Group("DATE(applied_at)").
		Scan(&dateResults)
	
	for _, dr := range dateResults {
		stats.ApplicationsByDate[dr.Date.Format("2006-01-02")] = dr.Count
	}
	
	return stats, nil
}

func (r *ApplicationRepositoryImpl) GetEmployerApplicationStats(ctx context.Context, employerID string) (*ApplicationStats, error) {
	stats := &ApplicationStats{
		ApplicationsByDate: make(map[string]int64),
	}
	
	// Get counts by status
	var results []struct {
		Status string
		Count  int64
	}
	
	err := r.db.WithContext(ctx).
		Model(&models.Application{}).
		Select("applications.status, COUNT(*) as count").
		Joins("JOIN jobs ON applications.job_id = jobs.id").
		Where("jobs.employer_id = ?", employerID).
		Group("applications.status").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	
	for _, r := range results {
		stats.TotalApplications += r.Count
		switch r.Status {
		case "pending":
			stats.PendingCount = r.Count
		case "viewed":
			stats.ViewedCount = r.Count
		case "shortlisted":
			stats.ShortlistedCount = r.Count
		case "rejected":
			stats.RejectedCount = r.Count
		case "hired":
			stats.HiredCount = r.Count
		case "withdrawn":
			stats.WithdrawnCount = r.Count
		}
	}
	
	// Get average match score
	var avgScore float64
	r.db.WithContext(ctx).
		Model(&models.Application{}).
		Select("AVG(applications.match_score)").
		Joins("JOIN jobs ON applications.job_id = jobs.id").
		Where("jobs.employer_id = ? AND applications.match_score > 0", employerID).
		Scan(&avgScore)
	stats.AverageMatchScore = avgScore
	
	// Get applications by date (last 30 days)
	var dateResults []struct {
		Date  time.Time
		Count int64
	}
	
	r.db.WithContext(ctx).
		Model(&models.Application{}).
		Select("DATE(applications.applied_at) as date, COUNT(*) as count").
		Joins("JOIN jobs ON applications.job_id = jobs.id").
		Where("jobs.employer_id = ? AND applications.applied_at >= ?", employerID, time.Now().AddDate(0, 0, -30)).
		Group("DATE(applications.applied_at)").
		Scan(&dateResults)
	
	for _, dr := range dateResults {
		stats.ApplicationsByDate[dr.Date.Format("2006-01-02")] = dr.Count
	}
	
	return stats, nil
}

func (r *ApplicationRepositoryImpl) BulkUpdateStatus(ctx context.Context, ids []string, status string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Application{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":      status,
			"reviewed_at": &now,
			"updated_at":  now,
		}).Error
}

func (r *ApplicationRepositoryImpl) DeleteByJob(ctx context.Context, jobID string) error {
	return r.db.WithContext(ctx).
		Where("job_id = ?", jobID).
		Delete(&models.Application{}).Error
}