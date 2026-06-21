package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type AdminSubscriptionHandler struct {
	subscriptionService services.SubscriptionService
	adminService        services.AdminService
	validator           *validator.Validate
}

func NewAdminSubscriptionHandler(subscriptionService services.SubscriptionService, adminService services.AdminService) *AdminSubscriptionHandler {
	return &AdminSubscriptionHandler{
		subscriptionService: subscriptionService,
		adminService:        adminService,
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

// GetSubscriptionDetail returns subscription with usage, payment history, and timeline
func (h *AdminSubscriptionHandler) GetSubscriptionDetail(c *fiber.Ctx) error {
	id := c.Params("id")

	sub, err := h.subscriptionService.GetSubscription(c.Context(), id)
	if err != nil {
		return utils.NotFound(c, "Subscription not found")
	}

	usage, _ := h.subscriptionService.GetSubscriptionUsage(c.Context(), id)
	payments, totalPayments, _ := h.subscriptionService.GetSubscriptionPayments(c.Context(), id, 1, 50)
	history, totalHistory, _ := h.subscriptionService.GetSubscriptionHistory(c.Context(), id, 1, 50)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"subscription":   sub,
			"usage":          usage,
			"payments":       payments,
			"total_payments": totalPayments,
			"history":        history,
			"total_history":  totalHistory,
		},
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

	adminID, _ := c.Locals("user_id").(string)
	h.adminService.LogAdminAction(c.Context(), adminID, "cancel", "subscription", id, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"reason": req.Reason, "user_id": sub.UserID})

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

	adminID, _ := c.Locals("user_id").(string)
	h.adminService.LogAdminAction(c.Context(), adminID, "reactivate", "subscription", id, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"user_id": sub.UserID})

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

	adminID, _ := c.Locals("user_id").(string)
	h.adminService.LogAdminAction(c.Context(), adminID, "change_plan", "subscription", id, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"new_plan_id": req.PlanID, "user_id": sub.UserID})

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

	adminID, _ := c.Locals("user_id").(string)
	h.adminService.LogAdminAction(c.Context(), adminID, "assign", "subscription", sub.ID, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"user_id": req.UserID, "plan_id": req.PlanID})

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subscription assigned successfully",
		"data":    sub,
	})
}

type adminUpdateSubReq struct {
	PlanID       *string  `json:"plan_id"`
	PlanName     *string  `json:"plan_name"`
	Amount       *float64 `json:"amount"`
	Currency     *string  `json:"currency"`
	Interval     *string  `json:"interval"`
	Status       *string  `json:"status"`
	JobPostLimit *int     `json:"job_post_limit"`
	AutoRenew    *bool    `json:"auto_renew"`
}

// UpdateSubscription allows admin to update subscription fields
func (h *AdminSubscriptionHandler) UpdateSubscription(c *fiber.Ctx) error {
	id := c.Params("id")

	var req adminUpdateSubReq
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	updates := make(map[string]interface{})
	if req.PlanID != nil {
		updates["plan_id"] = *req.PlanID
	}
	if req.PlanName != nil {
		updates["plan_name"] = *req.PlanName
	}
	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}
	if req.Currency != nil {
		updates["currency"] = *req.Currency
	}
	if req.Interval != nil {
		updates["interval"] = *req.Interval
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.JobPostLimit != nil {
		updates["job_post_limit"] = *req.JobPostLimit
	}
	if req.AutoRenew != nil {
		updates["auto_renew"] = *req.AutoRenew
	}

	if _, err := h.subscriptionService.AdminUpdateSubscription(c.Context(), id, updates); err != nil {
		return utils.InternalServerError(c, "Failed to update subscription: "+err.Error())
	}

	adminID, _ := c.Locals("user_id").(string)
	h.adminService.LogAdminAction(c.Context(), adminID, "update", "subscription", id, c.IP(), c.Get("User-Agent"), nil, updates)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subscription updated successfully",
	})
}

// CompletePendingUpgrade manually completes a pending plan change
func (h *AdminSubscriptionHandler) CompletePendingUpgrade(c *fiber.Ctx) error {
	id := c.Params("id")

	sub, err := h.subscriptionService.GetSubscription(c.Context(), id)
	if err != nil {
		return utils.NotFound(c, "Subscription not found")
	}

	if sub.PendingPlanID == nil || *sub.PendingPlanID == "" {
		return utils.BadRequest(c, "No pending plan change found")
	}

	if err := h.subscriptionService.ApplyPendingPlanChange(c.Context(), id); err != nil {
		return utils.InternalServerError(c, "Failed to apply pending plan change: "+err.Error())
	}

	adminID, _ := c.Locals("user_id").(string)
	h.adminService.LogAdminAction(c.Context(), adminID, "complete_pending", "subscription", id, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"pending_plan_id": sub.PendingPlanID})

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Pending plan change applied successfully",
	})
}

// CancelPendingUpgrade cancels a pending plan change
func (h *AdminSubscriptionHandler) CancelPendingUpgrade(c *fiber.Ctx) error {
	id := c.Params("id")

	sub, err := h.subscriptionService.GetSubscription(c.Context(), id)
	if err != nil {
		return utils.NotFound(c, "Subscription not found")
	}

	if sub.PendingPlanID == nil || *sub.PendingPlanID == "" {
		return utils.BadRequest(c, "No pending plan change found")
	}

	if err := h.subscriptionService.ClearPendingPlanChange(c.Context(), id); err != nil {
		return utils.InternalServerError(c, "Failed to cancel pending plan change: "+err.Error())
	}

	adminID, _ := c.Locals("user_id").(string)
	h.adminService.LogAdminAction(c.Context(), adminID, "cancel_pending", "subscription", id, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"pending_plan_id": sub.PendingPlanID})

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Pending plan change cancelled successfully",
	})
}

// GetSubscriptionStats returns aggregated subscription statistics
func (h *AdminSubscriptionHandler) GetSubscriptionStats(c *fiber.Ctx) error {
	stats, err := h.subscriptionService.GetSubscriptionStats(c.Context())
	if err != nil {
		return utils.InternalServerError(c, "Failed to get subscription stats: "+err.Error())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// ReconcileSubscriptions triggers reconciliation of stale pending subscriptions
func (h *AdminSubscriptionHandler) ReconcileSubscriptions(c *fiber.Ctx) error {
	result, err := h.subscriptionService.ReconcilePendingSubscriptions(c.Context())
	if err != nil {
		return utils.InternalServerError(c, "Failed to reconcile subscriptions: "+err.Error())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Reconciliation completed",
		"data":    result,
	})
}
