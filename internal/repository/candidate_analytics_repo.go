package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type CandidateAnalyticsRepository interface {
	// Application statistics
	GetApplicationStats(ctx context.Context, employeeID string) (*models.ApplicationStats, error)
	GetMonthlyApplicationStats(ctx context.Context, employeeID string, months int) ([]models.MonthlyStats, error)
	GetSuccessRates(ctx context.Context, employeeID string) (*models.SuccessRate, error)
	
	// Market positioning
	GetMarketPositioning(ctx context.Context, employeeID string) (*models.MarketPositioning, error)
	GetSkillPercentiles(ctx context.Context, skills []string) (map[string]int, error)
	
	// Skill demand
	GetSkillDemandTrends(ctx context.Context, skills []string) (map[string]int, error)
	
	// Recent activity
	GetRecentActivity(ctx context.Context, employeeID string, limit int) ([]models.RecentActivity, error)
	
	// Comparison data
	GetIndustryAverages(ctx context.Context, industry string) (map[string]float64, error)
	GetTotalCandidatesCount(ctx context.Context) (int64, error)

	// Get employee profile
	GetEmployeeProfile(ctx context.Context, employeeID string) (*models.EmployeeProfile, error)
	// Get top in-demand skills
	GetTopInDemandSkills(ctx context.Context, limit int, result interface{}) error
}

type CandidateAnalyticsRepositoryImpl struct {
	db *gorm.DB
}

func NewCandidateAnalyticsRepository(db *gorm.DB) CandidateAnalyticsRepository {
	return &CandidateAnalyticsRepositoryImpl{db: db}
}

func (r *CandidateAnalyticsRepositoryImpl) GetApplicationStats(ctx context.Context, employeeID string) (*models.ApplicationStats, error) {
	stats := &models.ApplicationStats{
		ByStatus: make(map[string]int),
	}
	
	// Get counts by status
	var statusCounts []struct {
		Status string
		Count  int
	}
	
	query := `
		SELECT status, COUNT(*) as count
		FROM applications
		WHERE employee_id = ?
		GROUP BY status
	`
	
	err := r.db.WithContext(ctx).Raw(query, employeeID).Scan(&statusCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
		stats.TotalApplications += sc.Count
		
		switch sc.Status {
		case "pending", "viewed":
			stats.ActiveApplications += sc.Count
			stats.PendingCount += sc.Count
		case "shortlisted":
			stats.ShortlistedCount += sc.Count
		case "rejected":
			stats.RejectedCount += sc.Count
		case "hired":
			stats.HiredCount += sc.Count
		case "withdrawn":
			stats.WithdrawnCount += sc.Count
		case "interview":
			stats.InterviewCount += sc.Count
		}
	}
	
	// Calculate response rate (applications that got response - viewed/shortlisted/rejected/hired)
	if stats.TotalApplications > 0 {
		responded := stats.ShortlistedCount + stats.RejectedCount + stats.HiredCount + stats.InterviewCount
		stats.ResponseRate = float64(responded) / float64(stats.TotalApplications) * 100
	}
	
	// Calculate average response time
	var avgResponseTime float64
	r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(AVG(EXTRACT(DAY FROM (reviewed_at - applied_at))), 0)
		FROM applications
		WHERE employee_id = ? AND reviewed_at IS NOT NULL
	`, employeeID).Scan(&avgResponseTime)
	stats.AverageResponseTime = avgResponseTime
	
	return stats, nil
}

func (r *CandidateAnalyticsRepositoryImpl) GetMonthlyApplicationStats(ctx context.Context, employeeID string, months int) ([]models.MonthlyStats, error) {
	var stats []models.MonthlyStats
	
	query := `
		SELECT 
			TO_CHAR(DATE_TRUNC('month', applied_at), 'YYYY-MM') as month,
			COUNT(*) as applications,
			SUM(CASE WHEN reviewed_at IS NOT NULL THEN 1 ELSE 0 END) as responses,
			ROUND(CAST(SUM(CASE WHEN reviewed_at IS NOT NULL THEN 1 ELSE 0 END) AS DECIMAL) / COUNT(*) * 100, 2) as response_rate
		FROM applications
		WHERE employee_id = ? AND applied_at >= NOW() - INTERVAL '? months'
		GROUP BY DATE_TRUNC('month', applied_at)
		ORDER BY month ASC
	`
	
	err := r.db.WithContext(ctx).Raw(query, employeeID, months).Scan(&stats).Error
	return stats, err
}

func (r *CandidateAnalyticsRepositoryImpl) GetSuccessRates(ctx context.Context, employeeID string) (*models.SuccessRate, error) {
	rate := &models.SuccessRate{
		ByExperienceLevel: make(map[string]float64),
		ByJobType:         make(map[string]float64),
		BySkills:          make(map[string]float64),
		IndustryAverages:  make(map[string]float64),
	}
	
	// Get conversion rates
	var rates struct {
		ApplicationToShortlist float64
		ShortlistToInterview   float64
		InterviewToHire        float64
	}
	
	r.db.WithContext(ctx).Raw(`
		SELECT 
			ROUND(CAST(COUNT(CASE WHEN status = 'shortlisted' THEN 1 END) AS DECIMAL) / NULLIF(COUNT(*), 0) * 100, 2) as application_to_shortlist,
			ROUND(CAST(COUNT(CASE WHEN status = 'interview' THEN 1 END) AS DECIMAL) / NULLIF(COUNT(CASE WHEN status = 'shortlisted' THEN 1 END), 0) * 100, 2) as shortlist_to_interview,
			ROUND(CAST(COUNT(CASE WHEN status = 'hired' THEN 1 END) AS DECIMAL) / NULLIF(COUNT(CASE WHEN status = 'interview' THEN 1 END), 0) * 100, 2) as interview_to_hire
		FROM applications
		WHERE employee_id = ?
	`, employeeID).Scan(&rates)
	
	rate.ApplicationToShortlistRate = rates.ApplicationToShortlist
	rate.ShortlistToInterviewRate = rates.ShortlistToInterview
	rate.InterviewToHireRate = rates.InterviewToHire
	
	if rates.ApplicationToShortlist > 0 && rates.ShortlistToInterview > 0 && rates.InterviewToHire > 0 {
		rate.OverallRate = (rates.ApplicationToShortlist / 100) * (rates.ShortlistToInterview / 100) * (rates.InterviewToHire / 100) * 100
	}
	
	// Get rates by experience level
	var expRates []struct {
		Level  string
		Rate   float64
	}
	r.db.WithContext(ctx).Raw(`
		SELECT 
			ep.experience_level as level,
			ROUND(CAST(COUNT(CASE WHEN a.status = 'hired' THEN 1 END) AS DECIMAL) / NULLIF(COUNT(*), 0) * 100, 2) as rate
		FROM applications a
		JOIN employee_profiles ep ON a.employee_id = ep.user_id
		WHERE a.employee_id = ?
		GROUP BY ep.experience_level
	`, employeeID).Scan(&expRates)
	
	for _, er := range expRates {
		rate.ByExperienceLevel[er.Level] = er.Rate
	}
	
	// Get percentile rank
	var totalCandidates int
	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT employee_id)
		FROM applications
		WHERE status = 'hired'
	`).Scan(&totalCandidates)
	
	var hiredCount int
	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM applications
		WHERE employee_id = ? AND status = 'hired'
	`, employeeID).Scan(&hiredCount)
	
	if hiredCount > 0 && totalCandidates > 0 {
		r.db.WithContext(ctx).Raw(`
			SELECT ROUND(PERCENT_RANK() OVER (ORDER BY hires DESC) * 100)
			FROM (
				SELECT employee_id, COUNT(*) as hires
				FROM applications
				WHERE status = 'hired'
				GROUP BY employee_id
			) subquery
			WHERE employee_id = ?
		`, employeeID).Scan(&rate.PercentileRank)
	}
	
	return rate, nil
}

func (r *CandidateAnalyticsRepositoryImpl) GetMarketPositioning(ctx context.Context, employeeID string) (*models.MarketPositioning, error) {
	positioning := &models.MarketPositioning{
		SkillPercentile: make(map[string]int),
	}
	
	// Get total candidates
	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT user_id) FROM employee_profiles
	`).Scan(&positioning.TotalCandidates)
	
	// Get candidate's rank based on profile completeness
	var rank int
	r.db.WithContext(ctx).Raw(`
		SELECT rank FROM (
			SELECT 
				user_id,
				ROW_NUMBER() OVER (ORDER BY 
					CASE WHEN full_name IS NOT NULL THEN 1 ELSE 0 END +
					CASE WHEN headline IS NOT NULL THEN 1 ELSE 0 END +
					CASE WHEN bio IS NOT NULL AND LENGTH(bio) > 50 THEN 1 ELSE 0 END +
					CASE WHEN location IS NOT NULL THEN 1 ELSE 0 END +
					CASE WHEN github_connected THEN 1 ELSE 0 END +
					CARDINALITY(skills) / 5 DESC
				) as rank
			FROM employee_profiles
		) ranked
		WHERE user_id = ?
	`, employeeID).Scan(&rank)
	
	positioning.YourRank = rank
	if positioning.TotalCandidates > 0 {
		positioning.Percentile = (positioning.TotalCandidates - rank) * 100 / positioning.TotalCandidates
	}
	
	// Get top skills
	var topSkills []string
	r.db.WithContext(ctx).Raw(`
		SELECT UNNEST(skills) as skill
		FROM employee_profiles
		WHERE user_id = ?
		LIMIT 5
	`, employeeID).Scan(&topSkills)
	positioning.TopSkills = topSkills
	
	// Get unique strengths (skills that are rare)
	var uniqueSkills []string
	r.db.WithContext(ctx).Raw(`
		SELECT skill
		FROM (
			SELECT 
				UNNEST(skills) as skill,
				COUNT(DISTINCT user_id) as candidate_count
			FROM employee_profiles
			GROUP BY skill
		) skill_counts
		WHERE skill = ANY(ARRAY(SELECT UNNEST(skills) FROM employee_profiles WHERE user_id = ?))
		AND candidate_count < 10
		LIMIT 3
	`, employeeID).Scan(&uniqueSkills)
	positioning.UniqueStrengths = uniqueSkills
	
	return positioning, nil
}

func (r *CandidateAnalyticsRepositoryImpl) GetSkillPercentiles(ctx context.Context, skills []string) (map[string]int, error) {
	percentiles := make(map[string]int)
	
	for _, skill := range skills {
		var percentile int
		r.db.WithContext(ctx).Raw(`
			SELECT ROUND(PERCENT_RANK() OVER (ORDER BY salary_min) * 100)
			FROM jobs
			WHERE ? = ANY(required_skills)
			LIMIT 1
		`, skill).Scan(&percentile)
		percentiles[skill] = percentile
	}
	
	return percentiles, nil
}

func (r *CandidateAnalyticsRepositoryImpl) GetSkillDemandTrends(ctx context.Context, skills []string) (map[string]int, error) {
	trends := make(map[string]int)
	
	for _, skill := range skills {
		var count int
		r.db.WithContext(ctx).Raw(`
			SELECT COUNT(*)
			FROM jobs
			WHERE ? = ANY(required_skills) AND posted_at >= NOW() - INTERVAL '30 days'
		`, skill).Scan(&count)
		trends[skill] = count
	}
	
	return trends, nil
}

func (r *CandidateAnalyticsRepositoryImpl) GetRecentActivity(ctx context.Context, employeeID string, limit int) ([]models.RecentActivity, error) {
	var activities []models.RecentActivity
	
	// Get recent applications
	var applications []struct {
		Title     string
		CreatedAt time.Time
	}
	r.db.WithContext(ctx).Raw(`
		SELECT j.title, a.applied_at as created_at
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE a.employee_id = ?
		ORDER BY a.applied_at DESC
		LIMIT ?
	`, employeeID, limit/2).Scan(&applications)
	
	for _, app := range applications {
		activities = append(activities, models.RecentActivity{
			Type:      "application",
			Title:     "Job Application Submitted",
			Details:   "You applied for " + app.Title,
			CreatedAt: app.CreatedAt,
		})
	}
	
	// Get profile updates (simplified - would need audit table in production)
	
	return activities, nil
}

func (r *CandidateAnalyticsRepositoryImpl) GetIndustryAverages(ctx context.Context, industry string) (map[string]float64, error) {
	averages := make(map[string]float64)
	
	var avgResponseRate float64
	r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(AVG(response_rate), 0)
		FROM (
			SELECT 
				COUNT(CASE WHEN reviewed_at IS NOT NULL THEN 1 END)::float / COUNT(*) * 100 as response_rate
			FROM applications a
			JOIN jobs j ON a.job_id = j.id
			WHERE j.industry = ?
			GROUP BY a.job_id
		) subquery
	`, industry).Scan(&avgResponseRate)
	averages["avg_response_rate"] = avgResponseRate
	
	return averages, nil
}

func (r *CandidateAnalyticsRepositoryImpl) GetTotalCandidatesCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.EmployeeProfile{}).Count(&count).Error
	return count, err
}


func (r *CandidateAnalyticsRepositoryImpl) GetEmployeeProfile(ctx context.Context, employeeID string) (*models.EmployeeProfile, error) {
	var profile models.EmployeeProfile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", employeeID).
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *CandidateAnalyticsRepositoryImpl) GetTopInDemandSkills(ctx context.Context, limit int, result interface{}) error {
	query := `
		SELECT 
			UNNEST(required_skills) as skill,
			COUNT(*) as job_count
		FROM jobs
		WHERE is_active = true
		GROUP BY UNNEST(required_skills)
		ORDER BY job_count DESC
		LIMIT ?
	`
	return r.db.WithContext(ctx).Raw(query, limit).Scan(result).Error
}
