package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type CandidateAnalyticsHandler struct {
	candidateAnalyticsService services.CandidateAnalyticsService
}

func NewCandidateAnalyticsHandler(candidateAnalyticsService services.CandidateAnalyticsService) *CandidateAnalyticsHandler {
	return &CandidateAnalyticsHandler{
		candidateAnalyticsService: candidateAnalyticsService,
	}
}

// GetDashboard returns complete candidate dashboard
func (h *CandidateAnalyticsHandler) GetDashboard(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	dashboard, err := h.candidateAnalyticsService.GetDashboard(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, dashboard)
}

// GetProfileStrength returns profile strength analysis
func (h *CandidateAnalyticsHandler) GetProfileStrength(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	strength, err := h.candidateAnalyticsService.GetProfileStrength(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, strength)
}

// GetApplicationStats returns application statistics
func (h *CandidateAnalyticsHandler) GetApplicationStats(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	stats, err := h.candidateAnalyticsService.GetApplicationStats(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, stats)
}

// GetMonthlyStats returns monthly application statistics
func (h *CandidateAnalyticsHandler) GetMonthlyStats(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	months, _ := strconv.Atoi(c.Query("months", "6"))
	
	stats, err := h.candidateAnalyticsService.GetMonthlyStats(c.Context(), userID.(string), months)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, stats)
}

// GetSuccessRates returns success rate analysis
func (h *CandidateAnalyticsHandler) GetSuccessRates(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	rates, err := h.candidateAnalyticsService.GetSuccessRates(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, rates)
}

// GetSkillGapAnalysis returns skill gap analysis
func (h *CandidateAnalyticsHandler) GetSkillGapAnalysis(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	analysis, err := h.candidateAnalyticsService.GetSkillGapAnalysis(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, analysis)
}

// GetMarketPositioning returns market positioning data
func (h *CandidateAnalyticsHandler) GetMarketPositioning(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	positioning, err := h.candidateAnalyticsService.GetMarketPositioning(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, positioning)
}

// GetRecommendations returns personalized recommendations
func (h *CandidateAnalyticsHandler) GetRecommendations(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	recommendations, err := h.candidateAnalyticsService.GetRecommendations(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, recommendations)
}

// GetRecentActivity returns recent user activity
func (h *CandidateAnalyticsHandler) GetRecentActivity(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	
	activities, err := h.candidateAnalyticsService.GetRecentActivity(c.Context(), userID.(string), limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, activities)
}
