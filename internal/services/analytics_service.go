package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type AnalyticsService interface {
	// Dashboard stats
	GetDashboardStats(ctx context.Context, employerID string) (*models.DashboardStats, error)
	
	// Combined dashboard data
	GetDashboard(ctx context.Context, employerID string, days int) (*models.DashboardResponse, error)
	
	// Application trends
	GetApplicationTrends(ctx context.Context, employerID string, period string, duration int) (*models.ApplicationTrendsResponse, error)
	
	// Conversion funnel
	GetConversionFunnel(ctx context.Context, employerID string) (*models.FunnelResponse, error)
	
	// Job performance
	GetJobPerformance(ctx context.Context, employerID string, limit int) ([]models.JobPerformance, error)
	GetJobPerformanceByID(ctx context.Context, jobID, employerID string) (*models.JobPerformance, error)
	
	// Time to hire
	GetTimeToHireMetrics(ctx context.Context, employerID string) (*models.TimeToHireMetric, error)
	
	// Application quality
	GetApplicationQualityScores(ctx context.Context, employerID string) (*models.ApplicationQualityScore, error)
	
	// Source analytics
	GetSourceAnalytics(ctx context.Context, employerID string) ([]models.SourceAnalytics, error)
	
	// Export
	ExportAnalytics(ctx context.Context, employerID string, format models.ExportFormat, startDate, endDate time.Time) ([]byte, string, error)
}

type AnalyticsServiceImpl struct {
	cfg           *config.Config
	analyticsRepo repository.AnalyticsRepository
}

func NewAnalyticsService(
	cfg *config.Config,
	analyticsRepo repository.AnalyticsRepository,
) AnalyticsService {
	return &AnalyticsServiceImpl{
		cfg:           cfg,
		analyticsRepo: analyticsRepo,
	}
}

func (s *AnalyticsServiceImpl) GetDashboardStats(ctx context.Context, employerID string) (*models.DashboardStats, error) {
	return s.analyticsRepo.GetDashboardStats(ctx, employerID)
}

func (s *AnalyticsServiceImpl) GetDashboard(ctx context.Context, employerID string, days int) (*models.DashboardResponse, error) {
	stats, err := s.analyticsRepo.GetDashboardStats(ctx, employerID)
	if err != nil {
		return nil, err
	}

	trends, err := s.GetApplicationTrends(ctx, employerID, "daily", days)
	if err != nil {
		trends = &models.ApplicationTrendsResponse{
			Daily:   []models.ApplicationTrend{},
			Weekly:  []models.ApplicationTrend{},
			Monthly: []models.ApplicationTrend{},
		}
	}

	funnel, err := s.analyticsRepo.GetConversionFunnel(ctx, employerID)
	if err != nil {
		funnel = []models.ConversionFunnel{}
	}
	funnelResp := &models.FunnelResponse{Stages: funnel}
	if len(funnel) > 0 {
		funnelResp.OverallRate = funnel[len(funnel)-1].ConversionRate
	}

	jobs, err := s.analyticsRepo.GetJobPerformance(ctx, employerID, 10)
	if err != nil {
		jobs = []models.JobPerformance{}
	}

	recentApps, err := s.analyticsRepo.GetRecentApplications(ctx, employerID, 5)
	if err != nil {
		recentApps = []models.RecentApplication{}
	}

	return &models.DashboardResponse{
		Stats:      stats,
		Trends:     trends,
		Funnel:     funnelResp,
		Jobs:       jobs,
		RecentApps: recentApps,
	}, nil
}

func (s *AnalyticsServiceImpl) GetApplicationTrends(ctx context.Context, employerID string, period string, duration int) (*models.ApplicationTrendsResponse, error) {
	response := &models.ApplicationTrendsResponse{
		Daily:   []models.ApplicationTrend{},
		Weekly:  []models.ApplicationTrend{},
		Monthly: []models.ApplicationTrend{},
	}
	
	var err error
	
	switch period {
	case "daily":
		response.Daily, err = s.analyticsRepo.GetDailyTrends(ctx, employerID, duration)
	case "weekly":
		response.Weekly, err = s.analyticsRepo.GetWeeklyTrends(ctx, employerID, duration)
	case "monthly":
		response.Monthly, err = s.analyticsRepo.GetMonthlyTrends(ctx, employerID, duration)
	default:
		response.Daily, err = s.analyticsRepo.GetDailyTrends(ctx, employerID, 30)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Calculate summary
	response.Summary = s.calculateTrendSummary(response)
	
	return response, nil
}

func (s *AnalyticsServiceImpl) calculateTrendSummary(response *models.ApplicationTrendsResponse) models.TrendSummary {
	summary := models.TrendSummary{}
	
	trends := response.Daily
	if len(trends) == 0 {
		return summary
	}
	
	var total int
	peakApplications := 0
	
	for _, trend := range trends {
		total += trend.Total
		if trend.Total > peakApplications {
			peakApplications = trend.Total
			summary.PeakDay = trend.Date
			summary.PeakApplications = trend.Total
		}
	}
	
	summary.TotalApplications = total
	summary.AveragePerDay = float64(total) / float64(len(trends))
	
	// Calculate growth rate (compare last 7 days vs previous 7 days)
	if len(trends) >= 14 {
		recentTotal := 0
		previousTotal := 0
		
		for i := len(trends) - 7; i < len(trends); i++ {
			recentTotal += trends[i].Total
		}
		for i := len(trends) - 14; i < len(trends)-7; i++ {
			previousTotal += trends[i].Total
		}
		
		if previousTotal > 0 {
			summary.GrowthRate = float64(recentTotal-previousTotal) / float64(previousTotal) * 100
		}
	}
	
	return summary
}

func (s *AnalyticsServiceImpl) GetConversionFunnel(ctx context.Context, employerID string) (*models.FunnelResponse, error) {
	stages, err := s.analyticsRepo.GetConversionFunnel(ctx, employerID)
	if err != nil {
		return nil, err
	}
	
	response := &models.FunnelResponse{
		Stages: stages,
	}
	
	// Calculate drop-off points
	for i := 0; i < len(stages)-1; i++ {
		dropOff := models.DropOffPoint{
			FromStage:    stages[i].Stage,
			ToStage:      stages[i+1].Stage,
			DropOffCount: stages[i].Count - stages[i+1].Count,
			DropOffRate:  100 - stages[i+1].ConversionRate,
		}
		
		// Add suggestions based on drop-off point
		switch stages[i+1].Stage {
		case "applied":
			if dropOff.DropOffRate > 50 {
				dropOff.Suggestion = "Consider improving job visibility or application process"
			}
		case "shortlisted":
			if dropOff.DropOffRate > 70 {
				dropOff.Suggestion = "Review your shortlisting criteria - many qualified candidates might be missed"
			}
		case "interview":
			if dropOff.DropOffRate > 60 {
				dropOff.Suggestion = "Consider reaching out to shortlisted candidates sooner"
			}
		case "hired":
			if dropOff.DropOffRate > 80 {
				dropOff.Suggestion = "Review your interview process and offer competitiveness"
			}
		}
		
		response.DropOffPoints = append(response.DropOffPoints, dropOff)
	}
	
	if len(stages) > 0 {
		response.OverallRate = stages[len(stages)-1].ConversionRate
	}
	
	return response, nil
}

func (s *AnalyticsServiceImpl) GetJobPerformance(ctx context.Context, employerID string, limit int) ([]models.JobPerformance, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	
	return s.analyticsRepo.GetJobPerformance(ctx, employerID, limit)
}

func (s *AnalyticsServiceImpl) GetJobPerformanceByID(ctx context.Context, jobID, employerID string) (*models.JobPerformance, error) {
	return s.analyticsRepo.GetJobPerformanceByID(ctx, jobID, employerID)
}

func (s *AnalyticsServiceImpl) GetTimeToHireMetrics(ctx context.Context, employerID string) (*models.TimeToHireMetric, error) {
	return s.analyticsRepo.GetTimeToHireMetrics(ctx, employerID)
}

func (s *AnalyticsServiceImpl) GetApplicationQualityScores(ctx context.Context, employerID string) (*models.ApplicationQualityScore, error) {
	return s.analyticsRepo.GetApplicationQualityScores(ctx, employerID)
}

func (s *AnalyticsServiceImpl) GetSourceAnalytics(ctx context.Context, employerID string) ([]models.SourceAnalytics, error) {
	return s.analyticsRepo.GetSourceAnalytics(ctx, employerID)
}

func (s *AnalyticsServiceImpl) ExportAnalytics(ctx context.Context, employerID string, format models.ExportFormat, startDate, endDate time.Time) ([]byte, string, error) {
	data, err := s.analyticsRepo.ExportApplicationsData(ctx, employerID, startDate, endDate)
	if err != nil {
		return nil, "", err
	}
	
	switch format {
	case models.FormatCSV:
		return s.exportToCSV(data), "text/csv", nil
	case models.FormatJSON:
		jsonData, err := json.Marshal(data)
		return jsonData, "application/json", err
	default:
		return s.exportToCSV(data), "text/csv", nil
	}
}

func (s *AnalyticsServiceImpl) exportToCSV(data []map[string]interface{}) []byte {
	var csvData [][]string
	
	// Headers
	if len(data) > 0 {
		headers := make([]string, 0, len(data[0]))
		for key := range data[0] {
			headers = append(headers, key)
		}
		csvData = append(csvData, headers)
	}
	
	// Rows
	for _, row := range data {
		csvRow := make([]string, 0, len(row))
		for _, value := range row {
			csvRow = append(csvRow, fmt.Sprintf("%v", value))
		}
		csvData = append(csvData, csvRow)
	}
	
	var buf []byte
	w := csv.NewWriter(&bytesWriter{&buf})
	w.WriteAll(csvData)
	w.Flush()
	
	return buf
}

// Helper for CSV writing
type bytesWriter struct {
	buf *[]byte
}

func (w *bytesWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}