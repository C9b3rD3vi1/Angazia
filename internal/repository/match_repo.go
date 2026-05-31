package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type MatchRepository interface {
	Create(ctx context.Context, match *models.Match) error
	GetByID(ctx context.Context, id string) (*models.Match, error)
	GetByJobAndEmployee(ctx context.Context, jobID, employeeID string) (*models.Match, error)
	Update(ctx context.Context, match *models.Match) error
	Delete(ctx context.Context, id string) error
	
	ListByEmployee(ctx context.Context, employeeID string, page, limit int) ([]*models.Match, int64, error)
	ListByJob(ctx context.Context, jobID string, page, limit int) ([]*models.Match, int64, error)
	ListByScore(ctx context.Context, minScore int, page, limit int) ([]*models.Match, int64, error)
	
	GetTopMatchesForEmployee(ctx context.Context, employeeID string, limit int) ([]*models.Match, error)
	GetTopMatchesForJob(ctx context.Context, jobID string, limit int) ([]*models.Match, error)
	
	CreateFeedback(ctx context.Context, feedback *models.MatchFeedback) error
	GetFeedbackByMatch(ctx context.Context, matchID string) ([]*models.MatchFeedback, error)
	
	DeleteExpiredMatches(ctx context.Context) error
	UpdateMatchScore(ctx context.Context, matchID string, score int) error
}

type MatchRepositoryImpl struct {
	db *gorm.DB
}

func NewMatchRepository(db *gorm.DB) MatchRepository {
	return &MatchRepositoryImpl{db: db}
}

func (r *MatchRepositoryImpl) Create(ctx context.Context, match *models.Match) error {
	match.ID = uuid.New().String()
	match.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(match).Error
}

func (r *MatchRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Match, error) {
	var match models.Match
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&match).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *MatchRepositoryImpl) GetByJobAndEmployee(ctx context.Context, jobID, employeeID string) (*models.Match, error) {
	var match models.Match
	err := r.db.WithContext(ctx).
		Where("job_id = ? AND employee_id = ?", jobID, employeeID).
		First(&match).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *MatchRepositoryImpl) Update(ctx context.Context, match *models.Match) error {
	return r.db.WithContext(ctx).Save(match).Error
}

func (r *MatchRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Match{}, "id = ?", id).Error
}

func (r *MatchRepositoryImpl) ListByEmployee(ctx context.Context, employeeID string, page, limit int) ([]*models.Match, int64, error) {
	var matches []*models.Match
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Match{}).Where("employee_id = ?", employeeID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("Job").
		Order("overall_score DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&matches).Error
	
	return matches, total, err
}

func (r *MatchRepositoryImpl) ListByJob(ctx context.Context, jobID string, page, limit int) ([]*models.Match, int64, error) {
	var matches []*models.Match
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Match{}).Where("job_id = ?", jobID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("Employee").
		Preload("Employee.User").
		Order("overall_score DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&matches).Error
	
	return matches, total, err
}

func (r *MatchRepositoryImpl) ListByScore(ctx context.Context, minScore int, page, limit int) ([]*models.Match, int64, error) {
	var matches []*models.Match
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Match{}).Where("overall_score >= ?", minScore)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("Job").
		Preload("Employee").
		Order("overall_score DESC").
		Offset(offset).
		Limit(limit).
		Find(&matches).Error
	
	return matches, total, err
}

func (r *MatchRepositoryImpl) GetTopMatchesForEmployee(ctx context.Context, employeeID string, limit int) ([]*models.Match, error) {
	var matches []*models.Match
	err := r.db.WithContext(ctx).
		Where("employee_id = ? AND overall_score >= 70", employeeID).
		Preload("Job").
		Preload("Job.Employer").
		Order("overall_score DESC").
		Limit(limit).
		Find(&matches).Error
	return matches, err
}

func (r *MatchRepositoryImpl) GetTopMatchesForJob(ctx context.Context, jobID string, limit int) ([]*models.Match, error) {
	var matches []*models.Match
	err := r.db.WithContext(ctx).
		Where("job_id = ? AND overall_score >= 70", jobID).
		Preload("Employee").
		Preload("Employee.User").
		Order("overall_score DESC").
		Limit(limit).
		Find(&matches).Error
	return matches, err
}

func (r *MatchRepositoryImpl) CreateFeedback(ctx context.Context, feedback *models.MatchFeedback) error {
	feedback.ID = uuid.New().String()
	feedback.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(feedback).Error
}

func (r *MatchRepositoryImpl) GetFeedbackByMatch(ctx context.Context, matchID string) ([]*models.MatchFeedback, error) {
	var feedbacks []*models.MatchFeedback
	err := r.db.WithContext(ctx).
		Where("match_id = ?", matchID).
		Order("created_at DESC").
		Find(&feedbacks).Error
	return feedbacks, err
}

func (r *MatchRepositoryImpl) DeleteExpiredMatches(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.Match{}).Error
}

func (r *MatchRepositoryImpl) UpdateMatchScore(ctx context.Context, matchID string, score int) error {
	return r.db.WithContext(ctx).
		Model(&models.Match{}).
		Where("id = ?", matchID).
		Updates(map[string]interface{}{
			"overall_score": score,
			"updated_at":    time.Now(),
		}).Error
}