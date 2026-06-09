package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type JobRepository interface {
	// Job CRUD operations
	Create(ctx context.Context, job *models.Job) error
	GetByID(ctx context.Context, id string) (*models.Job, error)
	GetByIDWithEmployer(ctx context.Context, id string) (*models.Job, error)
	Update(ctx context.Context, job *models.Job) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, isActive bool) error
	IncrementViews(ctx context.Context, id string) error
	IncrementApplications(ctx context.Context, id string) error
	
	// Listing with filters
	List(ctx context.Context, filters JobFilters, page, limit int) ([]*models.Job, int64, error)
	ListByEmployer(ctx context.Context, employerID string, page, limit int) ([]*models.Job, int64, error)
	ListActive(ctx context.Context, page, limit int) ([]*models.Job, int64, error)
	ListFeatured(ctx context.Context, limit int) ([]*models.Job, error)
	ListRecent(ctx context.Context, limit int) ([]*models.Job, error)
	
	// Search
	Search(ctx context.Context, query string, filters JobFilters, page, limit int) ([]*models.Job, int64, error)
	
	// Saved jobs
	SaveJob(ctx context.Context, employeeID, jobID string) error
	UnsaveJob(ctx context.Context, employeeID, jobID string) error
	GetSavedJobs(ctx context.Context, employeeID string, page, limit int) ([]*models.Job, int64, error)
	IsJobSaved(ctx context.Context, employeeID, jobID string) (bool, error)
	
	// Statistics
	GetStatsByEmployer(ctx context.Context, employerID string) (*JobStats, error)
	GetTotalJobCount(ctx context.Context) (int64, error)
	GetActiveJobCount(ctx context.Context) (int64, error)
	
	// Cleanup
	DeleteExpiredJobs(ctx context.Context) error
}

type JobFilters struct {
	EmployerID      string
	Title           string
	Location        string
	IsRemote        *bool
	IsHybrid        *bool
	EmploymentType  string
	ExperienceLevel string
	MinSalary       int
	MaxSalary       int
	Skills          []string
	Industry        string
	PostedAfter     time.Time
	PostedBefore    time.Time
	IsActive        *bool
	IsFeatured      *bool
	IsUrgent        *bool
}

type JobStats struct {
	TotalJobs        int64 `json:"total_jobs"`
	ActiveJobs       int64 `json:"active_jobs"`
	TotalApplications int64 `json:"total_applications"`
	TotalViews       int64 `json:"total_views"`
	JobsByStatus     map[string]int64 `json:"jobs_by_status"`
}

type JobRepositoryImpl struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) JobRepository {
	return &JobRepositoryImpl{db: db}
}

func (r *JobRepositoryImpl) Create(ctx context.Context, job *models.Job) error {
    job.ID = uuid.New().String()
    job.PostedAt = time.Now()
    job.UpdatedAt = time.Now()
    job.ViewsCount = 0
    job.ApplicationsCount = 0
    job.ShortlistedCount = 0
    job.HiredCount = 0
    
    // Use raw SQL for array types to ensure proper formatting
    return r.db.WithContext(ctx).Create(job).Error
}

func (r *JobRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Job, error) {
	var job models.Job
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&job).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (r *JobRepositoryImpl) GetByIDWithEmployer(ctx context.Context, id string) (*models.Job, error) {
	var job models.Job
	err := r.db.WithContext(ctx).
		Preload("Employer").
		Preload("Employer.User").
		Where("id = ?", id).
		First(&job).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (r *JobRepositoryImpl) Update(ctx context.Context, job *models.Job) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *JobRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.Job{}).Error
}

func (r *JobRepositoryImpl) UpdateStatus(ctx context.Context, id string, isActive bool) error {
	return r.db.WithContext(ctx).
		Model(&models.Job{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":  isActive,
			"updated_at": time.Now(),
		}).Error
}

func (r *JobRepositoryImpl) IncrementViews(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.Job{}).
		Where("id = ?", id).
		UpdateColumn("views_count", gorm.Expr("views_count + 1")).Error
}

func (r *JobRepositoryImpl) IncrementApplications(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.Job{}).
		Where("id = ?", id).
		UpdateColumn("applications_count", gorm.Expr("applications_count + 1")).Error
}

func (r *JobRepositoryImpl) List(ctx context.Context, filters JobFilters, page, limit int) ([]*models.Job, int64, error) {
	var jobs []*models.Job
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Job{}).Preload("Employer")
	
	// Apply filters
	if filters.EmployerID != "" {
		query = query.Where("employer_id = ?", filters.EmployerID)
	}
	
	if filters.Title != "" {
		query = query.Where("title ILIKE ?", "%"+filters.Title+"%")
	}
	
	if filters.Location != "" {
		query = query.Where("location ILIKE ?", "%"+filters.Location+"%")
	}
	
	if filters.IsRemote != nil {
		query = query.Where("is_remote = ?", *filters.IsRemote)
	}
	
	if filters.IsHybrid != nil {
		query = query.Where("is_hybrid = ?", *filters.IsHybrid)
	}
	
	if filters.EmploymentType != "" {
		query = query.Where("employment_type = ?", filters.EmploymentType)
	}
	
	if filters.ExperienceLevel != "" {
		query = query.Where("experience_level = ?", filters.ExperienceLevel)
	}
	
	if filters.MinSalary > 0 {
		query = query.Where("salary_max >= ?", filters.MinSalary)
	}
	
	if filters.MaxSalary > 0 {
		query = query.Where("salary_min <= ?", filters.MaxSalary)
	}
	
	if len(filters.Skills) > 0 {
		for _, skill := range filters.Skills {
			query = query.Where("? = ANY(required_skills)", skill)
		}
	}
	
	if filters.Industry != "" {
		query = query.Joins("JOIN employer_profiles ON jobs.employer_id = employer_profiles.user_id").
			Where("employer_profiles.industry = ?", filters.Industry)
	}
	
	if !filters.PostedAfter.IsZero() {
		query = query.Where("posted_at >= ?", filters.PostedAfter)
	}
	
	if !filters.PostedBefore.IsZero() {
		query = query.Where("posted_at <= ?", filters.PostedBefore)
	}
	
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	} else {
		query = query.Where("is_active = ?", true)
	}
	
	if filters.IsFeatured != nil {
		query = query.Where("is_featured = ?", *filters.IsFeatured)
	}
	
	if filters.IsUrgent != nil {
		query = query.Where("is_urgent = ?", *filters.IsUrgent)
	}
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination and ordering
	offset := (page - 1) * limit
	err := query.Order("posted_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&jobs).Error
	
	return jobs, total, err
}

func (r *JobRepositoryImpl) ListByEmployer(ctx context.Context, employerID string, page, limit int) ([]*models.Job, int64, error) {
	var jobs []*models.Job
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Job{}).Where("employer_id = ?", employerID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.Order("posted_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&jobs).Error
	
	return jobs, total, err
}

func (r *JobRepositoryImpl) ListActive(ctx context.Context, page, limit int) ([]*models.Job, int64, error) {
	filters := JobFilters{
		IsActive: boolPtr(true),
	}
	return r.List(ctx, filters, page, limit)
}

func (r *JobRepositoryImpl) ListFeatured(ctx context.Context, limit int) ([]*models.Job, error) {
	var jobs []*models.Job
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND is_featured = ?", true, true).
		Order("posted_at DESC").
		Limit(limit).
		Find(&jobs).Error
	return jobs, err
}

func (r *JobRepositoryImpl) ListRecent(ctx context.Context, limit int) ([]*models.Job, error) {
	var jobs []*models.Job
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("posted_at DESC").
		Limit(limit).
		Find(&jobs).Error
	return jobs, err
}

func (r *JobRepositoryImpl) Search(ctx context.Context, query string, filters JobFilters, page, limit int) ([]*models.Job, int64, error) {
	var jobs []*models.Job
	var total int64
	
	dbQuery := r.db.WithContext(ctx).Model(&models.Job{}).Preload("Employer")
	
	// Full-text search on title and description
	if query != "" {
		dbQuery = dbQuery.Where(
			"to_tsvector('english', title || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', ?)",
			query,
		)
	}
	
	// Apply other filters
	if filters.Location != "" {
		dbQuery = dbQuery.Where("location ILIKE ?", "%"+filters.Location+"%")
	}
	
	if filters.IsRemote != nil {
		dbQuery = dbQuery.Where("is_remote = ?", *filters.IsRemote)
	}
	
	if filters.EmploymentType != "" {
		dbQuery = dbQuery.Where("employment_type = ?", filters.EmploymentType)
	}
	
	if filters.ExperienceLevel != "" {
		dbQuery = dbQuery.Where("experience_level = ?", filters.ExperienceLevel)
	}
	
	if len(filters.Skills) > 0 {
		for _, skill := range filters.Skills {
			dbQuery = dbQuery.Where("? = ANY(required_skills)", skill)
		}
	}
	
	dbQuery = dbQuery.Where("is_active = ?", true)
	
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := dbQuery.Order("posted_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&jobs).Error
	
	return jobs, total, err
}

func (r *JobRepositoryImpl) SaveJob(ctx context.Context, employeeID, jobID string) error {
	savedJob := &models.SavedJob{
		ID:         uuid.New().String(),
		EmployeeID: employeeID,
		JobID:      jobID,
		SavedAt:    time.Now(),
	}
	return r.db.WithContext(ctx).Create(savedJob).Error
}

func (r *JobRepositoryImpl) UnsaveJob(ctx context.Context, employeeID, jobID string) error {
	return r.db.WithContext(ctx).
		Where("employee_id = ? AND job_id = ?", employeeID, jobID).
		Delete(&models.SavedJob{}).Error
}

func (r *JobRepositoryImpl) GetSavedJobs(ctx context.Context, employeeID string, page, limit int) ([]*models.Job, int64, error) {
	var jobs []*models.Job
	var total int64
	
	query := r.db.WithContext(ctx).
		Model(&models.Job{}).
		Joins("INNER JOIN saved_jobs ON saved_jobs.job_id = jobs.id").
		Where("saved_jobs.employee_id = ?", employeeID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.Order("saved_jobs.saved_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&jobs).Error
	
	return jobs, total, err
}

func (r *JobRepositoryImpl) IsJobSaved(ctx context.Context, employeeID, jobID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.SavedJob{}).
		Where("employee_id = ? AND job_id = ?", employeeID, jobID).
		Count(&count).Error
	return count > 0, err
}

func (r *JobRepositoryImpl) GetStatsByEmployer(ctx context.Context, employerID string) (*JobStats, error) {
	stats := &JobStats{
		JobsByStatus: make(map[string]int64),
	}
	
	// Total jobs
	if err := r.db.WithContext(ctx).
		Model(&models.Job{}).
		Where("employer_id = ?", employerID).
		Count(&stats.TotalJobs).Error; err != nil {
		return nil, err
	}
	
	// Active jobs
	if err := r.db.WithContext(ctx).
		Model(&models.Job{}).
		Where("employer_id = ? AND is_active = ?", employerID, true).
		Count(&stats.ActiveJobs).Error; err != nil {
		return nil, err
	}
	
	// Total applications
	if err := r.db.WithContext(ctx).
		Model(&models.Job{}).
		Where("employer_id = ?", employerID).
		Select("COALESCE(SUM(applications_count), 0)").
		Scan(&stats.TotalApplications).Error; err != nil {
		return nil, err
	}
	
	// Total views
	if err := r.db.WithContext(ctx).
		Model(&models.Job{}).
		Where("employer_id = ?", employerID).
		Select("COALESCE(SUM(views_count), 0)").
		Scan(&stats.TotalViews).Error; err != nil {
		return nil, err
	}
	
	return stats, nil
}

func (r *JobRepositoryImpl) GetTotalJobCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Job{}).
		Count(&count).Error
	return count, err
}

func (r *JobRepositoryImpl) GetActiveJobCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Job{}).
		Where("is_active = ?", true).
		Count(&count).Error
	return count, err
}

func (r *JobRepositoryImpl) DeleteExpiredJobs(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&models.Job{}).Error
}

func boolPtr(b bool) *bool {
	return &b
}