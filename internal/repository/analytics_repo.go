package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type AnalyticsRepository interface {
	// Dashboard stats
	GetDashboardStats(ctx context.Context, employerID string) (*models.DashboardStats, error)
	
	// Recent applications for dashboard
	GetRecentApplications(ctx context.Context, employerID string, limit int) ([]models.RecentApplication, error)
	
	// Active job count for subscription usage
	GetActiveJobCount(ctx context.Context, employerID string) (int, error)
	
	// Application trends
	GetDailyTrends(ctx context.Context, employerID string, days int) ([]models.ApplicationTrend, error)
	GetWeeklyTrends(ctx context.Context, employerID string, weeks int) ([]models.ApplicationTrend, error)
	GetMonthlyTrends(ctx context.Context, employerID string, months int) ([]models.ApplicationTrend, error)
	
	// Conversion funnel
	GetConversionFunnel(ctx context.Context, employerID string) ([]models.ConversionFunnel, error)
	
	// Job performance
	GetJobPerformance(ctx context.Context, employerID string, limit int) ([]models.JobPerformance, error)
	GetJobPerformanceByID(ctx context.Context, jobID, employerID string) (*models.JobPerformance, error)
	
	// Time to hire
	GetTimeToHireMetrics(ctx context.Context, employerID string) (*models.TimeToHireMetric, error)
	
	// Application quality
	GetApplicationQualityScores(ctx context.Context, employerID string) (*models.ApplicationQualityScore, error)
	
	// Source analytics
	GetSourceAnalytics(ctx context.Context, employerID string) ([]models.SourceAnalytics, error)
	
	// Export data
	ExportApplicationsData(ctx context.Context, employerID string, startDate, endDate time.Time) ([]map[string]interface{}, error)
}

type AnalyticsRepositoryImpl struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &AnalyticsRepositoryImpl{db: db}
}

func (r *AnalyticsRepositoryImpl) GetDashboardStats(ctx context.Context, employerID string) (*models.DashboardStats, error) {
	var stats models.DashboardStats
	query := `
		SELECT
			COUNT(DISTINCT j.id) FILTER (WHERE j.is_active = true) as active_jobs,
			COUNT(DISTINCT a.id) as total_applicants,
			COUNT(DISTINCT a.id) FILTER (WHERE a.applied_at >= NOW() - INTERVAL '30 days') as new_applications,
			COALESCE(SUM(j.views_count), 0) as profile_views,
			COUNT(DISTINCT a.id) FILTER (WHERE a.status = 'shortlisted') as shortlisted_count,
			COUNT(DISTINCT a.id) FILTER (WHERE a.status = 'hired') as hired_count
		FROM jobs j
		LEFT JOIN applications a ON a.job_id = j.id
		WHERE j.employer_id = ?
	`
	err := r.db.WithContext(ctx).Raw(query, employerID).Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *AnalyticsRepositoryImpl) GetRecentApplications(ctx context.Context, employerID string, limit int) ([]models.RecentApplication, error) {
	var apps []models.RecentApplication
	query := `
		SELECT 
			a.id,
			COALESCE(ep.full_name, 'Unknown') as candidate_name,
			COALESCE(u.email, '') as candidate_email,
			COALESCE(u.avatar_url, '') as candidate_avatar,
			j.title as job_title,
			j.id as job_id,
			a.status,
			COALESCE(a.match_score, 0) as match_score,
			to_char(a.applied_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as applied_at
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		JOIN users u ON a.employee_id = u.id
		LEFT JOIN employee_profiles ep ON a.employee_id = ep.user_id
		WHERE j.employer_id = ?
		ORDER BY a.applied_at DESC
		LIMIT ?
	`
	err := r.db.WithContext(ctx).Raw(query, employerID, limit).Scan(&apps).Error
	return apps, err
}

func (r *AnalyticsRepositoryImpl) GetActiveJobCount(ctx context.Context, employerID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Job{}).
		Where("employer_id = ? AND is_active = ?", employerID, true).
		Count(&count).Error
	return int(count), err
}

func (r *AnalyticsRepositoryImpl) GetDailyTrends(ctx context.Context, employerID string, days int) ([]models.ApplicationTrend, error) {
	var trends []models.ApplicationTrend
	
	query := `
		SELECT 
			DATE(a.applied_at) as date,
			COUNT(*) as total,
			SUM(CASE WHEN a.status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN a.status = 'viewed' THEN 1 ELSE 0 END) as viewed,
			SUM(CASE WHEN a.status = 'shortlisted' THEN 1 ELSE 0 END) as shortlisted,
			SUM(CASE WHEN a.status = 'rejected' THEN 1 ELSE 0 END) as rejected,
			SUM(CASE WHEN a.status = 'hired' THEN 1 ELSE 0 END) as hired,
			SUM(CASE WHEN a.status = 'withdrawn' THEN 1 ELSE 0 END) as withdrawn
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE j.employer_id = ? 
			AND a.applied_at >= NOW() - (INTERVAL '1 day' * ?)
		GROUP BY DATE(a.applied_at)
		ORDER BY date ASC
	`
	
	err := r.db.WithContext(ctx).Raw(query, employerID, days).Scan(&trends).Error
	return trends, err
}

func (r *AnalyticsRepositoryImpl) GetWeeklyTrends(ctx context.Context, employerID string, weeks int) ([]models.ApplicationTrend, error) {
	var trends []models.ApplicationTrend
	
	query := `
		SELECT 
			DATE_TRUNC('week', a.applied_at) as date,
			COUNT(*) as total,
			SUM(CASE WHEN a.status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN a.status = 'viewed' THEN 1 ELSE 0 END) as viewed,
			SUM(CASE WHEN a.status = 'shortlisted' THEN 1 ELSE 0 END) as shortlisted,
			SUM(CASE WHEN a.status = 'rejected' THEN 1 ELSE 0 END) as rejected,
			SUM(CASE WHEN a.status = 'hired' THEN 1 ELSE 0 END) as hired,
			SUM(CASE WHEN a.status = 'withdrawn' THEN 1 ELSE 0 END) as withdrawn
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE j.employer_id = ? 
			AND a.applied_at >= NOW() - (INTERVAL '1 week' * ?)
		GROUP BY DATE_TRUNC('week', a.applied_at)
		ORDER BY date ASC
	`
	
	err := r.db.WithContext(ctx).Raw(query, employerID, weeks).Scan(&trends).Error
	return trends, err
}

func (r *AnalyticsRepositoryImpl) GetMonthlyTrends(ctx context.Context, employerID string, months int) ([]models.ApplicationTrend, error) {
	var trends []models.ApplicationTrend
	
	query := `
		SELECT 
			DATE_TRUNC('month', a.applied_at) as date,
			COUNT(*) as total,
			SUM(CASE WHEN a.status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN a.status = 'viewed' THEN 1 ELSE 0 END) as viewed,
			SUM(CASE WHEN a.status = 'shortlisted' THEN 1 ELSE 0 END) as shortlisted,
			SUM(CASE WHEN a.status = 'rejected' THEN 1 ELSE 0 END) as rejected,
			SUM(CASE WHEN a.status = 'hired' THEN 1 ELSE 0 END) as hired,
			SUM(CASE WHEN a.status = 'withdrawn' THEN 1 ELSE 0 END) as withdrawn
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE j.employer_id = ? 
			AND a.applied_at >= NOW() - (INTERVAL '1 month' * ?)
		GROUP BY DATE_TRUNC('month', a.applied_at)
		ORDER BY date ASC
	`
	
	err := r.db.WithContext(ctx).Raw(query, employerID, months).Scan(&trends).Error
	return trends, err
}

func (r *AnalyticsRepositoryImpl) GetConversionFunnel(ctx context.Context, employerID string) ([]models.ConversionFunnel, error) {
	var funnel []models.ConversionFunnel
	
	query := `
		WITH stage_counts AS (
			SELECT 
				'viewed' as stage,
				COUNT(DISTINCT a.id) as count
			FROM applications a
			JOIN jobs j ON a.job_id = j.id
			WHERE j.employer_id = ? AND a.viewed_at IS NOT NULL
			
			UNION ALL
			
			SELECT 
				'applied',
				COUNT(*) as count
			FROM applications a
			JOIN jobs j ON a.job_id = j.id
			WHERE j.employer_id = ?
			
			UNION ALL
			
			SELECT 
				'shortlisted',
				COUNT(*) as count
			FROM applications a
			JOIN jobs j ON a.job_id = j.id
			WHERE j.employer_id = ? AND a.status = 'shortlisted'
			
			UNION ALL
			
			SELECT 
				'interview',
				COUNT(*) as count
			FROM applications a
			JOIN jobs j ON a.job_id = j.id
			WHERE j.employer_id = ? AND a.interview_date IS NOT NULL
			
			UNION ALL
			
			SELECT 
				'hired',
				COUNT(*) as count
			FROM applications a
			JOIN jobs j ON a.job_id = j.id
			WHERE j.employer_id = ? AND a.status = 'hired'
		)
		SELECT stage, count FROM stage_counts
	`
	
	err := r.db.WithContext(ctx).Raw(query, employerID, employerID, employerID, employerID, employerID).Scan(&funnel).Error
	if err != nil {
		return nil, err
	}
	
	// Calculate conversion rates
	var previousCount int
	for i := range funnel {
		if i == 0 {
			funnel[i].ConversionRate = 100.0
		} else {
			if previousCount > 0 {
				funnel[i].ConversionRate = float64(funnel[i].Count) / float64(previousCount) * 100
				funnel[i].DropOffRate = 100 - funnel[i].ConversionRate
			}
		}
		previousCount = funnel[i].Count
	}
	
	return funnel, nil
}

func (r *AnalyticsRepositoryImpl) GetJobPerformance(ctx context.Context, employerID string, limit int) ([]models.JobPerformance, error) {
	var performances []models.JobPerformance
	
	query := `
		SELECT 
			j.id as job_id,
			j.title,
			j.views_count as views,
			j.applications_count as applications,
			COUNT(CASE WHEN a.status = 'shortlisted' THEN 1 END) as shortlisted,
			COUNT(CASE WHEN a.status = 'hired' THEN 1 END) as hired,
			j.is_active,
			j.posted_at,
			CASE 
				WHEN j.views_count > 0 THEN ROUND(CAST(j.applications_count AS DECIMAL) / j.views_count * 100, 2)
				ELSE 0
			END as view_to_app_rate,
			CASE 
				WHEN j.applications_count > 0 THEN ROUND(CAST(COUNT(CASE WHEN a.status = 'shortlisted' THEN 1 END) AS DECIMAL) / j.applications_count * 100, 2)
				ELSE 0
			END as app_to_shortlist_rate,
			CASE 
				WHEN COUNT(CASE WHEN a.status = 'shortlisted' THEN 1 END) > 0 THEN ROUND(CAST(COUNT(CASE WHEN a.status = 'hired' THEN 1 END) AS DECIMAL) / COUNT(CASE WHEN a.status = 'shortlisted' THEN 1 END) * 100, 2)
				ELSE 0
			END as shortlist_to_hire_rate
		FROM jobs j
		LEFT JOIN applications a ON a.job_id = j.id
		WHERE j.employer_id = ?
		GROUP BY j.id, j.title, j.views_count, j.applications_count, j.posted_at, j.is_active
		ORDER BY j.posted_at DESC
		LIMIT ?
	`
	
	err := r.db.WithContext(ctx).Raw(query, employerID, limit).Scan(&performances).Error
	if err != nil {
		return nil, err
	}
	
	// Calculate overall success rate
	for i := range performances {
		if performances[i].Applications > 0 {
			performances[i].OverallSuccessRate = float64(performances[i].Hired) / float64(performances[i].Applications) * 100
		}
		
		// Calculate days to hire (simplified)
		if performances[i].Hired > 0 {
			var avgDays float64
			r.db.WithContext(ctx).Raw(`
				SELECT COALESCE(AVG(EXTRACT(DAY FROM (a.responded_at - a.applied_at))), 0)
				FROM applications a
				WHERE a.job_id = ? AND a.status = 'hired' AND a.responded_at IS NOT NULL
			`, performances[i].JobID).Scan(&avgDays)
			days := int(avgDays)
			performances[i].DaysToHire = &days
		}
	}
	
	return performances, nil
}

func (r *AnalyticsRepositoryImpl) GetJobPerformanceByID(ctx context.Context, jobID, employerID string) (*models.JobPerformance, error) {
	var performance models.JobPerformance
	
	query := `
		SELECT 
			j.id as job_id,
			j.title,
			j.views_count as views,
			j.applications_count as applications,
			COUNT(CASE WHEN a.status = 'shortlisted' THEN 1 END) as shortlisted,
			COUNT(CASE WHEN a.status = 'hired' THEN 1 END) as hired,
			j.is_active,
			j.posted_at,
			CASE 
				WHEN j.views_count > 0 THEN ROUND(CAST(j.applications_count AS DECIMAL) / j.views_count * 100, 2)
				ELSE 0
			END as view_to_app_rate,
			CASE 
				WHEN j.applications_count > 0 THEN ROUND(CAST(COUNT(CASE WHEN a.status = 'shortlisted' THEN 1 END) AS DECIMAL) / j.applications_count * 100, 2)
				ELSE 0
			END as app_to_shortlist_rate,
			CASE 
				WHEN COUNT(CASE WHEN a.status = 'shortlisted' THEN 1 END) > 0 THEN ROUND(CAST(COUNT(CASE WHEN a.status = 'hired' THEN 1 END) AS DECIMAL) / COUNT(CASE WHEN a.status = 'shortlisted' THEN 1 END) * 100, 2)
				ELSE 0
			END as shortlist_to_hire_rate
		FROM jobs j
		LEFT JOIN applications a ON a.job_id = j.id
		WHERE j.id = ? AND j.employer_id = ?
		GROUP BY j.id, j.title, j.views_count, j.applications_count, j.posted_at, j.is_active
	`
	
	err := r.db.WithContext(ctx).Raw(query, jobID, employerID).Scan(&performance).Error
	if err != nil {
		return nil, err
	}
	
	if performance.Applications > 0 {
		performance.OverallSuccessRate = float64(performance.Hired) / float64(performance.Applications) * 100
	}
	
	return &performance, nil
}

func (r *AnalyticsRepositoryImpl) GetTimeToHireMetrics(ctx context.Context, employerID string) (*models.TimeToHireMetric, error) {
	metric := &models.TimeToHireMetric{
		ByJobTitle:        make(map[string]int),
		ByExperienceLevel: make(map[string]int),
	}
	
	var result struct {
		AvgDays float64
		MinDays float64
		MaxDays float64
	}
	
	query := `
		SELECT 
			COALESCE(AVG(EXTRACT(DAY FROM (a.responded_at - a.applied_at))), 0) as avg_days,
			COALESCE(MIN(EXTRACT(DAY FROM (a.responded_at - a.applied_at))), 0) as min_days,
			COALESCE(MAX(EXTRACT(DAY FROM (a.responded_at - a.applied_at))), 0) as max_days
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE j.employer_id = ? AND a.status = 'hired' AND a.responded_at IS NOT NULL
	`
	
	err := r.db.WithContext(ctx).Raw(query, employerID).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	
	metric.AverageDays = int(result.AvgDays)
	metric.MinDays = int(result.MinDays)
	metric.MaxDays = int(result.MaxDays)
	metric.MedianDays = metric.AverageDays
	
	// Get time to hire by job title
	type titleResult struct {
		Title string
		Days  float64
	}
	var titleResults []titleResult
	r.db.WithContext(ctx).Raw(`
		SELECT j.title, COALESCE(AVG(EXTRACT(DAY FROM (a.responded_at - a.applied_at))), 0) as days
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE j.employer_id = ? AND a.status = 'hired' AND a.responded_at IS NOT NULL
		GROUP BY j.title
	`, employerID).Scan(&titleResults)
	
	for _, tr := range titleResults {
		metric.ByJobTitle[tr.Title] = int(tr.Days)
	}
	
	return metric, nil
}

func (r *AnalyticsRepositoryImpl) GetApplicationQualityScores(ctx context.Context, employerID string) (*models.ApplicationQualityScore, error) {
	var score models.ApplicationQualityScore
	
	query := `
		SELECT 
			COALESCE(AVG(a.match_score), 0) as average_match_score,
			COUNT(CASE WHEN a.match_score >= 80 THEN 1 END) as high_quality_count,
			COUNT(CASE WHEN a.match_score BETWEEN 50 AND 79 THEN 1 END) as medium_quality_count,
			COUNT(CASE WHEN a.match_score < 50 THEN 1 END) as low_quality_count,
			COALESCE(AVG(EXTRACT(HOUR FROM (a.reviewed_at - a.applied_at))), 0) as average_response_time
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE j.employer_id = ? AND a.match_score IS NOT NULL
	`
	
	err := r.db.WithContext(ctx).Raw(query, employerID).Scan(&score).Error
	if err != nil {
		return nil, err
	}
	
	return &score, nil
}

func (r *AnalyticsRepositoryImpl) GetSourceAnalytics(ctx context.Context, employerID string) ([]models.SourceAnalytics, error) {
	var sources []models.SourceAnalytics
	
	query := `
		SELECT 
			'direct' as source,
			COUNT(*) as count,
			100.00 as percentage,
			ROUND(CAST(SUM(CASE WHEN a.status = 'hired' THEN 1 ELSE 0 END) AS DECIMAL) / COUNT(*) * 100, 2) as conversion_rate
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE j.employer_id = ?
		ORDER BY count DESC
	`
	
	err := r.db.WithContext(ctx).Raw(query, employerID).Scan(&sources).Error
	return sources, err
}

func (r *AnalyticsRepositoryImpl) ExportApplicationsData(ctx context.Context, employerID string, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	query := `
		SELECT 
			j.title as job_title,
			ep.full_name as candidate_name,
			a.applied_at,
			a.status,
			a.match_score,
			LEFT(a.cover_letter, 200) as cover_letter_preview,
			CASE WHEN a.viewed_at IS NOT NULL THEN 'Yes' ELSE 'No' END as viewed,
			CASE WHEN a.status = 'shortlisted' THEN 'Yes' ELSE 'No' END as shortlisted,
			CASE WHEN a.interview_date IS NOT NULL THEN to_char(a.interview_date, 'YYYY-MM-DD') ELSE 'Not Scheduled' END as interview_date,
			CASE WHEN a.status = 'hired' THEN 'Yes' ELSE 'No' END as hired
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		JOIN employee_profiles ep ON a.employee_id = ep.user_id
		WHERE j.employer_id = ? 
			AND a.applied_at BETWEEN ? AND ?
		ORDER BY a.applied_at DESC
	`
	
	err := r.db.WithContext(ctx).Raw(query, employerID, startDate, endDate).Scan(&results).Error
	return results, err
}