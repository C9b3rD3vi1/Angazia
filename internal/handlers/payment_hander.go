package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type SubscriptionHandler struct {
	subscriptionService services.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

// GetPlans returns available subscription plans
func (h *SubscriptionHandler) GetPlans(c *fiber.Ctx) error {
	plans, err := h.subscriptionService.GetPlans(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    plans,
	})
}

// GetCurrentSubscription returns user's current subscription
func (h *SubscriptionHandler) GetCurrentSubscription(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	subscription, err := h.subscriptionService.GetCurrentSubscription(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    subscription,
	})
}

// CancelSubscription cancels user's subscription
func (h *SubscriptionHandler) CancelSubscription(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	var req struct {
		SubscriptionID string `json:"subscription_id"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	if err := h.subscriptionService.CancelSubscription(c.Context(), userID.(string), req.SubscriptionID, "user_requested"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "cancel_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Subscription cancelled successfully",
	})
}

// GetInvoices returns user's invoices
func (h *SubscriptionHandler) GetInvoices(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	invoices, total, err := h.subscriptionService.GetInvoices(c.Context(), userID.(string), page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"invoices": invoices,
			"total":    total,
			"page":     page,
			"limit":    limit,
		},
	})
}

// Subscribe creates a new subscription via M-Pesa
func (h *SubscriptionHandler) Subscribe(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false, Error: "unauthorized", Message: "User not authenticated",
		})
	}

	var req struct {
		PlanID      string `json:"plan_id" validate:"required"`
		PhoneNumber string `json:"phone_number"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false, Error: "invalid_request", Message: err.Error(),
		})
	}

	sub, chargeResp, err := h.subscriptionService.SubscribeWithNewPayment(c.Context(), userID.(string), req.PlanID, req.PhoneNumber)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false, Error: "subscribe_failed", Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true, Message: "Payment initiated",
		Data: fiber.Map{
			"subscription": sub,
			"charge":       chargeResp,
		},
	})
}

// Webhook handles IntaSend payment callbacks
func (h *SubscriptionHandler) Webhook(c *fiber.Ctx) error {
	var payload models.IntaSendWebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false, Error: "invalid_payload", Message: err.Error(),
		})
	}

	if err := h.subscriptionService.HandleWebhook(c.Context(), &payload); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false, Error: "webhook_failed", Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{Success: true})
}

// VerifyPayment checks the status of a payment
func (h *SubscriptionHandler) VerifyPayment(c *fiber.Ctx) error {
	transactionID := c.Query("transaction_id")
	reference := c.Query("reference")

	payment, err := h.subscriptionService.VerifyPayment(c.Context(), transactionID, reference)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false, Error: "verification_failed", Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{Success: true, Data: payment})
}

// ReactivateSubscription reactivates a cancelled/expired subscription
func (h *SubscriptionHandler) ReactivateSubscription(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var req struct {
		SubscriptionID string `json:"subscription_id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false, Error: "invalid_request", Message: err.Error(),
		})
	}

	sub, err := h.subscriptionService.ReactivateSubscription(c.Context(), userID.(string), req.SubscriptionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false, Error: "reactivation_failed", Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true, Message: "Subscription reactivated", Data: sub,
	})
}

// UpgradeSubscription upgrades a subscription with prorated billing
func (h *SubscriptionHandler) UpgradeSubscription(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var req struct {
		SubscriptionID string `json:"subscription_id" validate:"required"`
		NewPlanID      string `json:"new_plan_id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false, Error: "invalid_request", Message: err.Error(),
		})
	}

	proration, err := h.subscriptionService.CalculateProration(c.Context(), req.SubscriptionID, req.NewPlanID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false, Error: "proration_failed", Message: err.Error(),
		})
	}

	if proration.DueNow > 0 {
		return c.Status(fiber.StatusPaymentRequired).JSON(APIResponse{
			Success: false, Error: "payment_required",
			Message: "Additional payment required for upgrade",
			Data:    proration,
		})
	}

	sub, err := h.subscriptionService.UpgradeSubscription(c.Context(), userID.(string), req.SubscriptionID, req.NewPlanID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false, Error: "upgrade_failed", Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true, Message: "Subscription upgraded", Data: sub,
	})
}

// DowngradeSubscription downgrades a subscription with prorated credit
func (h *SubscriptionHandler) DowngradeSubscription(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var req struct {
		SubscriptionID string `json:"subscription_id" validate:"required"`
		NewPlanID      string `json:"new_plan_id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false, Error: "invalid_request", Message: err.Error(),
		})
	}

	sub, err := h.subscriptionService.DowngradeSubscription(c.Context(), userID.(string), req.SubscriptionID, req.NewPlanID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false, Error: "downgrade_failed", Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true, Message: "Subscription downgraded", Data: sub,
	})
}

// GetPaymentMethods returns saved payment methods
func (h *SubscriptionHandler) GetPaymentMethods(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	methods, err := h.subscriptionService.GetPaymentMethods(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false, Error: "fetch_failed", Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{Success: true, Data: methods})
}

// AddPaymentMethod saves a new payment method
func (h *SubscriptionHandler) AddPaymentMethod(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var req services.AddPaymentMethodRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false, Error: "invalid_request", Message: err.Error(),
		})
	}

	pm, err := h.subscriptionService.AddPaymentMethod(c.Context(), userID.(string), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false, Error: "add_failed", Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true, Message: "Payment method added", Data: pm,
	})
}

// RemovePaymentMethod deletes a payment method
func (h *SubscriptionHandler) RemovePaymentMethod(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	methodID := c.Params("id")

	if err := h.subscriptionService.RemovePaymentMethod(c.Context(), userID.(string), methodID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false, Error: "remove_failed", Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true, Message: "Payment method removed",
	})
}

// SetDefaultPaymentMethod sets a payment method as default
func (h *SubscriptionHandler) SetDefaultPaymentMethod(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	methodID := c.Params("id")

	if err := h.subscriptionService.SetDefaultPaymentMethod(c.Context(), userID.(string), methodID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false, Error: "update_failed", Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true, Message: "Default payment method updated",
	})
}

// GetProration calculates prorated charges for plan change
func (h *SubscriptionHandler) GetProration(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var req struct {
		SubscriptionID string `json:"subscription_id" validate:"required"`
		NewPlanID      string `json:"new_plan_id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false, Error: "invalid_request", Message: err.Error(),
		})
	}

	_ = userID
	proration, err := h.subscriptionService.CalculateProration(c.Context(), req.SubscriptionID, req.NewPlanID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false, Error: "proration_failed", Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{Success: true, Data: proration})
}

// RetryPayment retries a failed payment
func (h *SubscriptionHandler) RetryPayment(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var req struct {
		SubscriptionID string `json:"subscription_id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false, Error: "invalid_request", Message: err.Error(),
		})
	}

	_ = userID
	if err := h.subscriptionService.RetryPayment(c.Context(), req.SubscriptionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false, Error: "retry_failed", Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true, Message: "Payment retry initiated",
	})
}