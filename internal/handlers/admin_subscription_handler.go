package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type AdminSubscriptionHandler struct {
	subscriptionService services.SubscriptionService
	validator           *validator.Validate
}

func NewAdminSubscriptionHandler(subscriptionService services.SubscriptionService) *AdminSubscriptionHandler {
	return &AdminSubscriptionHandler{
		subscriptionService: subscriptionService,
		validator:           validator.New(),
	}
}

type adminListSubsQuery struct {
	UserID string `query:"user_id"`
	PlanID string `query:"plan_id"`
	Status string `query:"status"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

// ListSubscriptions returns all subscriptions with optional filters
func (h *AdminSubscriptionHandler) ListSubscriptions(c *fiber.Ctx) error {
	var q adminListSubsQuery
	if err := c.QueryParser(&q); err != nil {
		return utils.BadRequest(c, "Invalid query parameters")
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}

	filters := make(map[string]interface{})
	if q.UserID != "" {
		filters["user_id"] = q.UserID
	}
	if q.PlanID != "" {
		filters["plan_id"] = q.PlanID
	}
	if q.Status != "" {
		filters["status"] = q.Status
	}

	subs, total, err := h.subscriptionService.GetAllSubscriptions(c.Context(), filters, q.Page, q.Limit)
	if err != nil {
		return utils.InternalServerError(c, "Failed to fetch subscriptions")
	}

	totalPages := int(total) / q.Limit
	if int(total)%q.Limit > 0 {
		totalPages++
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"subscriptions": subs,
			"total":         total,
			"page":          q.Page,
			"limit":         q.Limit,
			"total_pages":   totalPages,
		},
	})
}

// GetSubscription returns a single subscription by ID
func (h *AdminSubscriptionHandler) GetSubscription(c *fiber.Ctx) error {
	id := c.Params("id")
	sub, err := h.subscriptionService.GetSubscription(c.Context(), id)
	if err != nil {
		return utils.NotFound(c, "Subscription not found")
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    sub,
	})
}

type adminCancelReq struct {
	Reason string `json:"reason"`
}

// CancelSubscription cancels a subscription
func (h *AdminSubscriptionHandler) CancelSubscription(c *fiber.Ctx) error {
	id := c.Params("id")

	sub, err := h.subscriptionService.GetSubscription(c.Context(), id)
	if err != nil {
		return utils.NotFound(c, "Subscription not found")
	}

	var req adminCancelReq
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	if err := h.subscriptionService.CancelSubscription(c.Context(), sub.UserID, id, req.Reason); err != nil {
		return utils.InternalServerError(c, "Failed to cancel subscription")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subscription cancelled successfully",
	})
}

// ReactivateSubscription reactivates a cancelled or expired subscription
func (h *AdminSubscriptionHandler) ReactivateSubscription(c *fiber.Ctx) error {
	id := c.Params("id")

	sub, err := h.subscriptionService.GetSubscription(c.Context(), id)
	if err != nil {
		return utils.NotFound(c, "Subscription not found")
	}

	result, err := h.subscriptionService.ReactivateSubscription(c.Context(), sub.UserID, id)
	if err != nil {
		return utils.InternalServerError(c, "Failed to reactivate subscription: "+err.Error())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subscription reactivated successfully",
		"data":    result,
	})
}

type adminChangePlanReq struct {
	PlanID string `json:"plan_id" validate:"required"`
}

// ChangePlan changes the plan for a subscription
func (h *AdminSubscriptionHandler) ChangePlan(c *fiber.Ctx) error {
	id := c.Params("id")

	var req adminChangePlanReq
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, "Plan ID is required")
	}

	sub, err := h.subscriptionService.GetSubscription(c.Context(), id)
	if err != nil {
		return utils.NotFound(c, "Subscription not found")
	}

	// Determine if it's an upgrade or downgrade by price comparison
	// For simplicity, always use upgrade
	result, err := h.subscriptionService.UpgradeSubscription(c.Context(), sub.UserID, id, req.PlanID)
	if err != nil {
		// Try downgrade if upgrade fails
		result, err = h.subscriptionService.DowngradeSubscription(c.Context(), sub.UserID, id, req.PlanID)
		if err != nil {
			return utils.InternalServerError(c, "Failed to change plan: "+err.Error())
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Plan changed successfully",
		"data":    result,
	})
}

type adminAssignReq struct {
	UserID string `json:"user_id" validate:"required"`
	PlanID string `json:"plan_id" validate:"required"`
}

// AssignSubscription creates a new subscription for a user (admin bypasses payment)
func (h *AdminSubscriptionHandler) AssignSubscription(c *fiber.Ctx) error {
	var req adminAssignReq
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, "User ID and Plan ID are required")
	}

	sub, err := h.subscriptionService.AdminAssignSubscription(c.Context(), req.UserID, req.PlanID)
	if err != nil {
		return utils.InternalServerError(c, "Failed to assign subscription: "+err.Error())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subscription assigned successfully",
		"data":    sub,
	})
}
