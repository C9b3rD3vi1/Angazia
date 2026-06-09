package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type TalentPoolRepository interface {
	// Talent pool operations
	CreatePool(ctx context.Context, pool *models.TalentPool) error
	GetPool(ctx context.Context, id string) (*models.TalentPool, error)
	GetPoolByEmployer(ctx context.Context, id, employerID string) (*models.TalentPool, error)
	UpdatePool(ctx context.Context, pool *models.TalentPool) error
	DeletePool(ctx context.Context, id, employerID string) error
	ListPoolsByEmployer(ctx context.Context, employerID string, page, limit int) ([]*models.TalentPool, int64, error)
	
	// Candidate operations
	AddCandidate(ctx context.Context, candidate *models.TalentPoolCandidate) error
	GetCandidate(ctx context.Context, id string) (*models.TalentPoolCandidate, error)
	GetCandidateByPoolAndEmployee(ctx context.Context, poolID, employeeID string) (*models.TalentPoolCandidate, error)
	UpdateCandidate(ctx context.Context, candidate *models.TalentPoolCandidate) error
	RemoveCandidate(ctx context.Context, id, poolID string) error
	ListCandidatesByPool(ctx context.Context, poolID string, filters CandidateFilters, page, limit int) ([]*models.TalentPoolCandidate, int64, error)
	ListCandidatesByEmployer(ctx context.Context, employerID string, page, limit int) ([]*models.TalentPoolCandidate, int64, error)
	
	// Candidate pool lookup
	GetCandidatePools(ctx context.Context, employerID, employeeID string) ([]*models.TalentPool, error)

	// Bulk operations
	BulkAddCandidates(ctx context.Context, candidates []*models.TalentPoolCandidate) error
	BulkUpdateStatus(ctx context.Context, candidateIDs []string, status string) error
	BulkRemoveCandidates(ctx context.Context, candidateIDs []string, poolID string) error
	
	// Statistics
	GetPoolStats(ctx context.Context, poolID string) (*models.TalentPoolStats, error)
	GetEmployerTalentStats(ctx context.Context, employerID string) (map[string]int, error)
	
	// Search
	SearchCandidatesInPool(ctx context.Context, poolID, query string, page, limit int) ([]*models.TalentPoolCandidate, int64, error)
	
	// Cleanup
	ArchiveOldCandidates(ctx context.Context, days int) error
}

type CandidateFilters struct {
	Status       string   `json:"status"`
	Tags         []string `json:"tags"`
	MinMatchScore int     `json:"min_match_score"`
	MaxMatchScore int     `json:"max_match_score"`
	AddedAfter   time.Time `json:"added_after"`
	AddedBefore  time.Time `json:"added_before"`
	Contacted    *bool    `json:"contacted"`
	Hired        *bool    `json:"hired"`
}

type TalentPoolRepositoryImpl struct {
	db *gorm.DB
}

func NewTalentPoolRepository(db *gorm.DB) TalentPoolRepository {
	return &TalentPoolRepositoryImpl{db: db}
}

func (r *TalentPoolRepositoryImpl) CreatePool(ctx context.Context, pool *models.TalentPool) error {
	pool.ID = uuid.New().String()
	pool.CreatedAt = time.Now()
	pool.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(pool).Error
}

func (r *TalentPoolRepositoryImpl) GetPool(ctx context.Context, id string) (*models.TalentPool, error) {
	var pool models.TalentPool
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&pool).Error
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

func (r *TalentPoolRepositoryImpl) GetPoolByEmployer(ctx context.Context, id, employerID string) (*models.TalentPool, error) {
	var pool models.TalentPool
	err := r.db.WithContext(ctx).
		Where("id = ? AND employer_id = ?", id, employerID).
		First(&pool).Error
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

func (r *TalentPoolRepositoryImpl) UpdatePool(ctx context.Context, pool *models.TalentPool) error {
	pool.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(pool).Error
}

func (r *TalentPoolRepositoryImpl) DeletePool(ctx context.Context, id, employerID string) error {
	// Delete all candidates in the pool first
	if err := r.db.WithContext(ctx).
		Where("talent_pool_id = ?", id).
		Delete(&models.TalentPoolCandidate{}).Error; err != nil {
		return err
	}
	
	return r.db.WithContext(ctx).
		Where("id = ? AND employer_id = ?", id, employerID).
		Delete(&models.TalentPool{}).Error
}

func (r *TalentPoolRepositoryImpl) ListPoolsByEmployer(ctx context.Context, employerID string, page, limit int) ([]*models.TalentPool, int64, error) {
	var pools []*models.TalentPool
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.TalentPool{}).
		Where("employer_id = ?", employerID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&pools).Error
	
	// Get candidate count for each pool
	for _, pool := range pools {
		var count int64
		r.db.WithContext(ctx).Model(&models.TalentPoolCandidate{}).
			Where("talent_pool_id = ?", pool.ID).
			Count(&count)
		pool.CandidateCount = int(count)
	}
	
	return pools, total, err
}

func (r *TalentPoolRepositoryImpl) AddCandidate(ctx context.Context, candidate *models.TalentPoolCandidate) error {
	candidate.ID = uuid.New().String()
	candidate.AddedAt = time.Now()
	candidate.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(candidate).Error
}

func (r *TalentPoolRepositoryImpl) GetCandidate(ctx context.Context, id string) (*models.TalentPoolCandidate, error) {
	var candidate models.TalentPoolCandidate
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Preload("Employee").
		Preload("Employee.User").
		First(&candidate).Error
	if err != nil {
		return nil, err
	}
	return &candidate, nil
}

func (r *TalentPoolRepositoryImpl) GetCandidateByPoolAndEmployee(ctx context.Context, poolID, employeeID string) (*models.TalentPoolCandidate, error) {
	var candidate models.TalentPoolCandidate
	err := r.db.WithContext(ctx).
		Where("talent_pool_id = ? AND employee_id = ?", poolID, employeeID).
		First(&candidate).Error
	if err != nil {
		return nil, err
	}
	return &candidate, nil
}

func (r *TalentPoolRepositoryImpl) UpdateCandidate(ctx context.Context, candidate *models.TalentPoolCandidate) error {
	candidate.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(candidate).Error
}

func (r *TalentPoolRepositoryImpl) RemoveCandidate(ctx context.Context, id, poolID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND talent_pool_id = ?", id, poolID).
		Delete(&models.TalentPoolCandidate{}).Error
}

func (r *TalentPoolRepositoryImpl) ListCandidatesByPool(ctx context.Context, poolID string, filters CandidateFilters, page, limit int) ([]*models.TalentPoolCandidate, int64, error) {
	var candidates []*models.TalentPoolCandidate
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.TalentPoolCandidate{}).
		Where("talent_pool_id = ?", poolID)
	
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	
	if len(filters.Tags) > 0 {
		for _, tag := range filters.Tags {
			query = query.Where("? = ANY(tags)", tag)
		}
	}
	
	if filters.MinMatchScore > 0 {
		query = query.Where("match_score >= ?", filters.MinMatchScore)
	}
	
	if filters.MaxMatchScore > 0 {
		query = query.Where("match_score <= ?", filters.MaxMatchScore)
	}
	
	if !filters.AddedAfter.IsZero() {
		query = query.Where("added_at >= ?", filters.AddedAfter)
	}
	
	if !filters.AddedBefore.IsZero() {
		query = query.Where("added_at <= ?", filters.AddedBefore)
	}
	
	if filters.Contacted != nil {
		if *filters.Contacted {
			query = query.Where("contacted_at IS NOT NULL")
		} else {
			query = query.Where("contacted_at IS NULL")
		}
	}
	
	if filters.Hired != nil {
		if *filters.Hired {
			query = query.Where("hired_at IS NOT NULL")
		} else {
			query = query.Where("hired_at IS NULL")
		}
	}
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("Employee").
		Preload("Employee.User").
		Order("match_score DESC, added_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&candidates).Error
	
	return candidates, total, err
}

func (r *TalentPoolRepositoryImpl) ListCandidatesByEmployer(ctx context.Context, employerID string, page, limit int) ([]*models.TalentPoolCandidate, int64, error) {
	var candidates []*models.TalentPoolCandidate
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.TalentPoolCandidate{}).
		Joins("JOIN talent_pools ON talent_pools.id = talent_pool_candidates.talent_pool_id").
		Where("talent_pools.employer_id = ?", employerID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("Employee").
		Preload("Employee.User").
		Preload("TalentPool").
		Order("talent_pool_candidates.added_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&candidates).Error
	
	return candidates, total, err
}

func (r *TalentPoolRepositoryImpl) GetCandidatePools(ctx context.Context, employerID, employeeID string) ([]*models.TalentPool, error) {
	var pools []*models.TalentPool
	err := r.db.WithContext(ctx).
		Model(&models.TalentPool{}).
		Joins("JOIN talent_pool_candidates ON talent_pool_candidates.talent_pool_id = talent_pools.id").
		Where("talent_pools.employer_id = ? AND talent_pool_candidates.employee_id = ?", employerID, employeeID).
		Find(&pools).Error
	if err != nil {
		return nil, err
	}
	return pools, nil
}

func (r *TalentPoolRepositoryImpl) BulkAddCandidates(ctx context.Context, candidates []*models.TalentPoolCandidate) error {
	for _, candidate := range candidates {
		candidate.ID = uuid.New().String()
		candidate.AddedAt = time.Now()
		candidate.UpdatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(&candidates).Error
}

func (r *TalentPoolRepositoryImpl) BulkUpdateStatus(ctx context.Context, candidateIDs []string, status string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": now,
	}
	
	if status == "contacted" {
		updates["contacted_at"] = now
	} else if status == "hired" {
		updates["hired_at"] = now
	}
	
	return r.db.WithContext(ctx).
		Model(&models.TalentPoolCandidate{}).
		Where("id IN ?", candidateIDs).
		Updates(updates).Error
}

func (r *TalentPoolRepositoryImpl) BulkRemoveCandidates(ctx context.Context, candidateIDs []string, poolID string) error {
	return r.db.WithContext(ctx).
		Where("id IN ? AND talent_pool_id = ?", candidateIDs, poolID).
		Delete(&models.TalentPoolCandidate{}).Error
}

func (r *TalentPoolRepositoryImpl) GetPoolStats(ctx context.Context, poolID string) (*models.TalentPoolStats, error) {
	stats := &models.TalentPoolStats{
		TagsDistribution: make(map[string]int),
	}
	
	var candidates []struct {
		Status     string
		MatchScore int
		Tags       []string
	}
	
	err := r.db.WithContext(ctx).
		Model(&models.TalentPoolCandidate{}).
		Where("talent_pool_id = ?", poolID).
		Select("status, match_score, tags").
		Scan(&candidates).Error
	if err != nil {
		return nil, err
	}
	
	var totalMatchScore int
	for _, c := range candidates {
		stats.TotalCandidates++
		
		switch c.Status {
		case "active":
			stats.ActiveCandidates++
		case "contacted":
			stats.ContactedCount++
		case "hired":
			stats.HiredCount++
		case "archived":
			stats.ArchivedCount++
		}
		
		totalMatchScore += c.MatchScore
		
		for _, tag := range c.Tags {
			stats.TagsDistribution[tag]++
		}
	}
	
	if stats.TotalCandidates > 0 {
		stats.AverageMatchScore = float64(totalMatchScore) / float64(stats.TotalCandidates)
	}
	
	return stats, nil
}

func (r *TalentPoolRepositoryImpl) GetEmployerTalentStats(ctx context.Context, employerID string) (map[string]int, error) {
	stats := make(map[string]int)
	
	// Get total pools count
	var poolCount int64
	r.db.WithContext(ctx).Model(&models.TalentPool{}).
		Where("employer_id = ?", employerID).
		Count(&poolCount)
	stats["total_pools"] = int(poolCount)
	
	// Get total candidates count
	var candidateCount int64
	r.db.WithContext(ctx).Model(&models.TalentPoolCandidate{}).
		Joins("JOIN talent_pools ON talent_pools.id = talent_pool_candidates.talent_pool_id").
		Where("talent_pools.employer_id = ?", employerID).
		Count(&candidateCount)
	stats["total_candidates"] = int(candidateCount)
	
	// Get candidates by status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	
	r.db.WithContext(ctx).Model(&models.TalentPoolCandidate{}).
		Select("status, COUNT(*) as count").
		Joins("JOIN talent_pools ON talent_pools.id = talent_pool_candidates.talent_pool_id").
		Where("talent_pools.employer_id = ?", employerID).
		Group("status").
		Scan(&statusCounts)
	
	for _, sc := range statusCounts {
		stats["status_"+sc.Status] = int(sc.Count)
	}
	
	return stats, nil
}

func (r *TalentPoolRepositoryImpl) SearchCandidatesInPool(ctx context.Context, poolID, query string, page, limit int) ([]*models.TalentPoolCandidate, int64, error) {
	var candidates []*models.TalentPoolCandidate
	var total int64
	
	dbQuery := r.db.WithContext(ctx).Model(&models.TalentPoolCandidate{}).
		Where("talent_pool_id = ?", poolID)
	
	if query != "" {
		dbQuery = dbQuery.Where(
			"notes ILIKE ? OR EXISTS (SELECT 1 FROM employee_profiles WHERE employee_profiles.user_id = talent_pool_candidates.employee_id AND (employee_profiles.full_name ILIKE ? OR employee_profiles.headline ILIKE ?))",
			"%"+query+"%", "%"+query+"%", "%"+query+"%",
		)
	}
	
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := dbQuery.
		Preload("Employee").
		Preload("Employee.User").
		Order("match_score DESC").
		Offset(offset).
		Limit(limit).
		Find(&candidates).Error
	
	return candidates, total, err
}

func (r *TalentPoolRepositoryImpl) ArchiveOldCandidates(ctx context.Context, days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return r.db.WithContext(ctx).
		Model(&models.TalentPoolCandidate{}).
		Where("status = ? AND updated_at < ?", "active", cutoffDate).
		Update("status", "archived").Error
}