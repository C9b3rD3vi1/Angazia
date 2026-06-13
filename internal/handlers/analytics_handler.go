package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
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

func (h *AnalyticsHandler) GetDashboardStats(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	stats, err := h.analyticsService.GetDashboardStats(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, stats)
}

func (h *AnalyticsHandler) GetApplicationTrends(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	period := c.Query("period", "daily")
	duration, _ := strconv.Atoi(c.Query("duration", "30"))
	if d := c.QueryInt("days", 0); d > 0 {
		duration = d
	}

	trends, err := h.analyticsService.GetApplicationTrends(c.Context(), userID.(string), period, duration)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, trends)
}

func (h *AnalyticsHandler) GetConversionFunnel(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	funnel, err := h.analyticsService.GetConversionFunnel(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, funnel)
}

func (h *AnalyticsHandler) GetJobPerformance(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	limit := 10
	if l := c.QueryInt("limit", 10); l > 0 {
		limit = l
	}

	performance, err := h.analyticsService.GetJobPerformance(c.Context(), userID.(string), limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, performance)
}

func (h *AnalyticsHandler) GetJobPerformanceByID(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	jobID := c.Params("jobId")
	if jobID == "" {
		return utils.BadRequest(c, "Job ID is required")
	}

	performance, err := h.analyticsService.GetJobPerformanceByID(c.Context(), jobID, userID.(string))
	if err != nil {
		return utils.NotFound(c, "Job")
	}

	return utils.Success(c, performance)
}

func (h *AnalyticsHandler) GetTimeToHireMetrics(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	metrics, err := h.analyticsService.GetTimeToHireMetrics(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, metrics)
}

func (h *AnalyticsHandler) GetApplicationQualityScores(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	scores, err := h.analyticsService.GetApplicationQualityScores(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, scores)
}

func (h *AnalyticsHandler) GetDemographics(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	demo, err := h.analyticsService.GetDemographics(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, demo)
}

func (h *AnalyticsHandler) GetStageDurations(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	durations, err := h.analyticsService.GetStageDurations(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, durations)
}

func (h *AnalyticsHandler) GetSourceAnalytics(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	sources, err := h.analyticsService.GetSourceAnalytics(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, sources)
}

func (h *AnalyticsHandler) ExportAnalytics(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	format := models.ExportFormat(c.Query("format", "csv"))
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return utils.BadRequest(c, "Invalid start_date format, use YYYY-MM-DD")
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return utils.BadRequest(c, "Invalid end_date format, use YYYY-MM-DD")
		}
	} else {
		endDate = time.Now()
	}

	data, contentType, err := h.analyticsService.ExportAnalytics(c.Context(), userID.(string), format, startDate, endDate)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename=analytics."+string(format))
	return c.Send(data)
}
