package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type AnalyticsHandler struct {
	analyticsService services.AnalyticsService
}

func NewAnalyticsHandler(analyticsService services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetApplicationTrends returns application trends over time
func (h *AnalyticsHandler) GetApplicationTrends(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	period := c.Query("period", "daily")
	duration, _ := strconv.Atoi(c.Query("duration", "30"))
	
	trends, err := h.analyticsService.GetApplicationTrends(c.Context(), userID.(string), period, duration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    trends,
	})
}

// GetConversionFunnel returns the application conversion funnel
func (h *AnalyticsHandler) GetConversionFunnel(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	funnel, err := h.analyticsService.GetConversionFunnel(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    funnel,
	})
}

// GetJobPerformance returns performance metrics for all jobs
func (h *AnalyticsHandler) GetJobPerformance(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	performance, err := h.analyticsService.GetJobPerformance(c.Context(), userID.(string), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    performance,
	})
}

// GetJobPerformanceByID returns performance metrics for a specific job
func (h *AnalyticsHandler) GetJobPerformanceByID(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	jobID := c.Params("jobId")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_job_id",
			Message: "Job ID is required",
		})
	}
	
	performance, err := h.analyticsService.GetJobPerformanceByID(c.Context(), jobID, userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    performance,
	})
}

// GetTimeToHireMetrics returns time-to-hire statistics
func (h *AnalyticsHandler) GetTimeToHireMetrics(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	metrics, err := h.analyticsService.GetTimeToHireMetrics(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    metrics,
	})
}

// GetApplicationQualityScores returns application quality metrics
func (h *AnalyticsHandler) GetApplicationQualityScores(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	scores, err := h.analyticsService.GetApplicationQualityScores(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    scores,
	})
}

// GetSourceAnalytics returns applicant source analytics
func (h *AnalyticsHandler) GetSourceAnalytics(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	sources, err := h.analyticsService.GetSourceAnalytics(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    sources,
	})
}

// ExportAnalytics exports analytics data
func (h *AnalyticsHandler) ExportAnalytics(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	format := models.ExportFormat(c.Query("format", "csv"))
	startDateStr := c.Query("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.Query("end_date", time.Now().Format("2006-01-02"))
	
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_start_date",
			Message: "Invalid start date format. Use YYYY-MM-DD",
		})
	}
	
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_end_date",
			Message: "Invalid end date format. Use YYYY-MM-DD",
		})
	}
	
	data, mimeType, err := h.analyticsService.ExportAnalytics(c.Context(), userID.(string), format, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "export_failed",
			Message: err.Error(),
		})
	}
	
	filename := "analytics_export_" + time.Now().Format("20060102") + "." + string(format)
	
	c.Set("Content-Type", mimeType)
	c.Set("Content-Disposition", "attachment; filename="+filename)
	
	return c.Send(data)
}

// GetDashboardOverview returns a comprehensive dashboard overview
func (h *AnalyticsHandler) GetDashboardOverview(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	// Get all analytics in parallel (simplified - in production use goroutines)
	trends, _ := h.analyticsService.GetApplicationTrends(c.Context(), userID.(string), "daily", 30)
	funnel, _ := h.analyticsService.GetConversionFunnel(c.Context(), userID.(string))
	jobPerformance, _ := h.analyticsService.GetJobPerformance(c.Context(), userID.(string), 5)
	timeToHire, _ := h.analyticsService.GetTimeToHireMetrics(c.Context(), userID.(string))
	qualityScores, _ := h.analyticsService.GetApplicationQualityScores(c.Context(), userID.(string))
	sources, _ := h.analyticsService.GetSourceAnalytics(c.Context(), userID.(string))
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"trends":          trends,
			"funnel":          funnel,
			"top_jobs":        jobPerformance,
			"time_to_hire":    timeToHire,
			"quality_scores":  qualityScores,
			"top_sources":     sources,
		},
	})
}