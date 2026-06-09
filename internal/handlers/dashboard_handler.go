package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type DashboardHandler struct {
	analyticsService    services.AnalyticsService
	subscriptionService services.SubscriptionService
}

func NewDashboardHandler(
	analyticsService services.AnalyticsService,
	subscriptionService services.SubscriptionService,
) *DashboardHandler {
	return &DashboardHandler{
		analyticsService:    analyticsService,
		subscriptionService: subscriptionService,
	}
}

func (h *DashboardHandler) GetEmployerDashboard(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	days, _ := strconv.Atoi(c.Query("days", "30"))
	if days < 1 {
		days = 30
	}
	if days > 365 {
		days = 365
	}

	dashboard, err := h.analyticsService.GetDashboard(c.Context(), userID.(string), days)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	if sub, subErr := h.subscriptionService.GetCurrentSubscription(c.Context(), userID.(string)); subErr == nil && sub != nil {
		jobsUsed := 0
		if dashboard.Stats != nil {
			jobsUsed = dashboard.Stats.ActiveJobs
		}
		dashboard.Subscription = &models.SubscriptionInfo{
			PlanName:  sub.PlanName,
			Amount:    sub.Amount,
			Currency:  sub.Currency,
			Interval:  sub.Interval,
			JobsUsed:  jobsUsed,
			JobsLimit: sub.JobPostLimit,
			Status:    sub.Status,
			Features:  sub.Features,
		}
	}

	return utils.Success(c, dashboard)
}
