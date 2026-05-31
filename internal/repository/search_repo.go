package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type SearchRepository interface {
	// Search operations
	SearchJobs(ctx context.Context, filters models.SearchFilters, page, limit int) ([]*models.Job, int64, error)
	SearchCandidates(ctx context.Context, filters models.SearchFilters, page, limit int) ([]*models.EmployeeProfile, int64, error)
	SearchCompanies(ctx context.Context, filters models.SearchFilters, page, limit int) ([]*models.EmployerProfile, int64, error)
	
	// Facets
	GetJobFacets(ctx context.Context, filters models.SearchFilters) (*models.FacetResult, error)
	
	// Search history
	LogSearch(ctx context.Context, query *models.SearchQuery) error
	GetSearchHistory(ctx context.Context, userID string, limit int) ([]*models.SearchQuery, error)
	GetPopularSearches(ctx context.Context, days int, limit int) ([]struct{ Query string; Count int }, error)
	
	// Saved searches
	SaveSearch(ctx context.Context, saved *models.SavedSearch) error
	GetSavedSearches(ctx context.Context, userID string, entityType string) ([]*models.SavedSearch, error)
	DeleteSavedSearch(ctx context.Context, id, userID string) error
	UpdateSavedSearchLastRun(ctx context.Context, id string) error
}

type SearchRepositoryImpl struct {
	db *gorm.DB
}

func NewSearchRepository(db *gorm.DB) SearchRepository {
	return &SearchRepositoryImpl{db: db}
}

func (r *SearchRepositoryImpl) SearchJobs(ctx context.Context, filters models.SearchFilters, page, limit int) ([]*models.Job, int64, error) {
	var jobs []*models.Job
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Job{}).Preload("Employer")
	
	// Full-text search on title and description
	if filters.Keywords != "" {
		query = query.Where(
			"to_tsvector('english', title || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', ?)",
			filters.Keywords,
		)
	}
	
	// Location filter
	if filters.Location != "" {
		if filters.Radius > 0 {
			// Simplified location search - in production, use PostGIS
			query = query.Where("location ILIKE ?", "%"+filters.Location+"%")
		} else {
			query = query.Where("location ILIKE ?", "%"+filters.Location+"%")
		}
	}
	
	// Remote/Hybrid filters
	if filters.IsRemote != nil {
		query = query.Where("is_remote = ?", *filters.IsRemote)
	}
	if filters.IsHybrid != nil {
		query = query.Where("is_hybrid = ?", *filters.IsHybrid)
	}
	
	// Job-specific filters
	if filters.JobTitle != "" {
		query = query.Where("title ILIKE ?", "%"+filters.JobTitle+"%")
	}
	if filters.CompanyName != "" {
		query = query.Joins("JOIN employer_profiles ON jobs.employer_id = employer_profiles.user_id").
			Where("employer_profiles.company_name ILIKE ?", "%"+filters.CompanyName+"%")
	}
	if filters.Industry != "" {
		query = query.Joins("JOIN employer_profiles ON jobs.employer_id = employer_profiles.user_id").
			Where("employer_profiles.industry = ?", filters.Industry)
	}
	if filters.EmploymentType != "" {
		query = query.Where("employment_type = ?", filters.EmploymentType)
	}
	if filters.ExperienceLevel != "" {
		query = query.Where("experience_level = ?", filters.ExperienceLevel)
	}
	
	// Salary range
	if filters.MinSalary > 0 {
		query = query.Where("salary_max >= ?", filters.MinSalary)
	}
	if filters.MaxSalary > 0 {
		query = query.Where("salary_min <= ?", filters.MaxSalary)
	}
	
	// Skills filter
	if len(filters.Skills) > 0 {
		for _, skill := range filters.Skills {
			query = query.Where("? = ANY(required_skills)", skill)
		}
	}
	
	// Posted within
	if filters.PostedWithin != "" {
		var days int
		switch filters.PostedWithin {
		case "24h":
			days = 1
		case "7d":
			days = 7
		case "30d":
			days = 30
		case "90d":
			days = 90
		default:
			days = 30
		}
		query = query.Where("posted_at >= NOW() - INTERVAL '? days'", days)
	}
	
	// Only active jobs
	query = query.Where("is_active = ?", true)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Sorting
	switch filters.SortBy {
	case "date":
		order := "posted_at DESC"
		if filters.SortOrder == "asc" {
			order = "posted_at ASC"
		}
		query = query.Order(order)
	case "salary":
		order := "salary_max DESC"
		if filters.SortOrder == "asc" {
			order = "salary_max ASC"
		}
		query = query.Order(order)
	case "relevance":
		fallthrough
	default:
		query = query.Order("posted_at DESC")
	}
	
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&jobs).Error
	
	return jobs, total, err
}

func (r *SearchRepositoryImpl) SearchCandidates(ctx context.Context, filters models.SearchFilters, page, limit int) ([]*models.EmployeeProfile, int64, error) {
	var candidates []*models.EmployeeProfile
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.EmployeeProfile{}).Preload("User")
	
	// Full-text search
	if filters.Keywords != "" {
		query = query.Where(
			"to_tsvector('english', COALESCE(full_name, '') || ' ' || COALESCE(headline, '') || ' ' || COALESCE(bio, '')) @@ plainto_tsquery('english', ?)",
			filters.Keywords,
		)
	}
	
	// Location filter
	if filters.Location != "" {
		query = query.Where("location ILIKE ?", "%"+filters.Location+"%")
	}
	
	// Experience range
	if filters.MinExperience > 0 {
		query = query.Where("years_of_experience >= ?", filters.MinExperience)
	}
	if filters.MaxExperience > 0 {
		query = query.Where("years_of_experience <= ?", filters.MaxExperience)
	}
	
	// Skills filter
	if len(filters.Skills) > 0 {
		for _, skill := range filters.Skills {
			query = query.Where("? = ANY(skills)", skill)
		}
	}
	
	// GitHub connection
	if filters.GitHubConnected != nil {
		query = query.Where("github_connected = ?", *filters.GitHubConnected)
	}
	
	// Match score (requires joining with matches table)
	if filters.MinMatchScore > 0 {
		// This would require a specific job ID to calculate match score
		// For now, just filter by profile completeness
	}
	
	// Only visible candidates
	query = query.Where("is_visible = ?", true)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Sorting
	switch filters.SortBy {
	case "experience":
		order := "years_of_experience DESC"
		if filters.SortOrder == "asc" {
			order = "years_of_experience ASC"
		}
		query = query.Order(order)
	default:
		query = query.Order("created_at DESC")
	}
	
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&candidates).Error
	
	return candidates, total, err
}

func (r *SearchRepositoryImpl) SearchCompanies(ctx context.Context, filters models.SearchFilters, page, limit int) ([]*models.EmployerProfile, int64, error) {
	var companies []*models.EmployerProfile
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.EmployerProfile{}).Preload("User")
	
	if filters.Keywords != "" {
		query = query.Where(
			"company_name ILIKE ? OR company_description ILIKE ?",
			"%"+filters.Keywords+"%", "%"+filters.Keywords+"%",
		)
	}
	
	if filters.Industry != "" {
		query = query.Where("industry = ?", filters.Industry)
	}
	
	if filters.CompanySize != "" {
		query = query.Where("company_size = ?", filters.CompanySize)
	}
	
	if filters.IsVerified != nil {
		if *filters.IsVerified {
			query = query.Where("verification_status = ?", "verified")
		}
	}
	
	query = query.Where("verification_status IN (?)", []string{"verified", "pending"})
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&companies).Error
	
	return companies, total, err
}

func (r *SearchRepositoryImpl) GetJobFacets(ctx context.Context, filters models.SearchFilters) (*models.FacetResult, error) {
	facets := &models.FacetResult{
		Locations:        []models.FacetCount{},
		Skills:           []models.FacetCount{},
		ExperienceLevels: []models.FacetCount{},
		EmploymentTypes:  []models.FacetCount{},
		Industries:       []models.FacetCount{},
		SalaryRanges:     []models.FacetCount{},
	}
	
	// Build base query
	baseQuery := r.db.WithContext(ctx).Model(&models.Job{}).Where("is_active = ?", true)
	
	if filters.Keywords != "" {
		baseQuery = baseQuery.Where(
			"to_tsvector('english', title || ' ' || COALESCE(description, '')) @@ plainto_tsquery('english', ?)",
			filters.Keywords,
		)
	}
	
	// Location facets
	var locations []struct {
		Location string
		Count    int
	}
	baseQuery.Select("location, COUNT(*) as count").
		Where("location != ''").
		Group("location").
		Order("count DESC").
		Limit(10).
		Scan(&locations)
	for _, l := range locations {
		facets.Locations = append(facets.Locations, models.FacetCount{Value: l.Location, Count: l.Count})
	}
	
	// Skills facets
	var skills []struct {
		Skill string
		Count int
	}
	baseQuery.Raw(`
		SELECT skill, COUNT(*) as count
		FROM (
			SELECT UNNEST(required_skills) as skill
			FROM jobs
			WHERE is_active = true
		) skills
		GROUP BY skill
		ORDER BY count DESC
		LIMIT 10
	`).Scan(&skills)
	for _, s := range skills {
		facets.Skills = append(facets.Skills, models.FacetCount{Value: s.Skill, Count: s.Count})
	}
	
	// Experience levels facets
	var expLevels []struct {
		Level string
		Count int
	}
	baseQuery.Select("experience_level, COUNT(*) as count").
		Where("experience_level != ''").
		Group("experience_level").
		Scan(&expLevels)
	for _, el := range expLevels {
		facets.ExperienceLevels = append(facets.ExperienceLevels, models.FacetCount{Value: el.Level, Count: el.Count})
	}
	
	// Employment types facets
	var empTypes []struct {
		Type  string
		Count int
	}
	baseQuery.Select("employment_type, COUNT(*) as count").
		Where("employment_type != ''").
		Group("employment_type").
		Scan(&empTypes)
	for _, et := range empTypes {
		facets.EmploymentTypes = append(facets.EmploymentTypes, models.FacetCount{Value: et.Type, Count: et.Count})
	}
	
	// Industries facets
	var industries []struct {
		Industry string
		Count    int
	}
	baseQuery.Joins("JOIN employer_profiles ON jobs.employer_id = employer_profiles.user_id").
		Select("employer_profiles.industry, COUNT(*) as count").
		Where("employer_profiles.industry != ''").
		Group("employer_profiles.industry").
		Scan(&industries)
	for _, i := range industries {
		facets.Industries = append(facets.Industries, models.FacetCount{Value: i.Industry, Count: i.Count})
	}
	
	// Salary ranges
	salaryRanges := []struct {
		Min    int
		Max    int
		Label  string
		Count  int
	}{
		{0, 50000, "0 - 50K", 0},
		{50001, 100000, "50K - 100K", 0},
		{100001, 150000, "100K - 150K", 0},
		{150001, 200000, "150K - 200K", 0},
		{200001, 300000, "200K - 300K", 0},
		{300001, 500000, "300K - 500K", 0},
		{500001, 1000000, "500K - 1M", 0},
		{1000001, 9999999, "1M+", 0},
	}
	
	for i := range salaryRanges {
		var count int64
		baseQuery.Where("salary_min >= ? AND salary_max <= ?", salaryRanges[i].Min, salaryRanges[i].Max).Count(&count)
		salaryRanges[i].Count = int(count)
		facets.SalaryRanges = append(facets.SalaryRanges, models.FacetCount{Value: salaryRanges[i].Label, Count: salaryRanges[i].Count})
	}
	
	return facets, nil
}

func (r *SearchRepositoryImpl) LogSearch(ctx context.Context, query *models.SearchQuery) error {
	query.ID = uuid.New().String()
	query.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(query).Error
}

func (r *SearchRepositoryImpl) GetSearchHistory(ctx context.Context, userID string, limit int) ([]*models.SearchQuery, error) {
	var queries []*models.SearchQuery
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&queries).Error
	return queries, err
}

func (r *SearchRepositoryImpl) GetPopularSearches(ctx context.Context, days int, limit int) ([]struct{ Query string; Count int }, error) {
	var results []struct{ Query string; Count int }
	err := r.db.WithContext(ctx).Raw(`
		SELECT query, COUNT(*) as count
		FROM search_queries
		WHERE created_at >= NOW() - INTERVAL '? days'
		GROUP BY query
		ORDER BY count DESC
		LIMIT ?
	`, days, limit).Scan(&results).Error
	return results, err
}

func (r *SearchRepositoryImpl) SaveSearch(ctx context.Context, saved *models.SavedSearch) error {
	saved.ID = uuid.New().String()
	saved.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(saved).Error
}

func (r *SearchRepositoryImpl) GetSavedSearches(ctx context.Context, userID string, entityType string) ([]*models.SavedSearch, error) {
	var searches []*models.SavedSearch
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	err := query.Order("created_at DESC").Find(&searches).Error
	return searches, err
}

func (r *SearchRepositoryImpl) DeleteSavedSearch(ctx context.Context, id, userID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.SavedSearch{}).Error
}

func (r *SearchRepositoryImpl) UpdateSavedSearchLastRun(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.SavedSearch{}).
		Where("id = ?", id).
		Update("last_run_at", now).Error
}