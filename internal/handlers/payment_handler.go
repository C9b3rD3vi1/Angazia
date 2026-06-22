package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
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

func (h *SubscriptionHandler) GetPlans(c *fiber.Ctx) error {
	plans, err := h.subscriptionService.GetPlans(c.Context())
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.Success(c, plans)
}

func (h *SubscriptionHandler) GetCurrentSubscription(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	uid, ok := userID.(string)
	if !ok {
		return utils.Unauthorized(c, "User ID not found")
	}

	subscription, err := h.subscriptionService.GetCurrentSubscription(c.Context(), uid)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, subscription)
}

func (h *SubscriptionHandler) CancelSubscription(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req struct {
		SubscriptionID string `json:"subscription_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	uid, ok := userID.(string)
	if !ok {
		return utils.Unauthorized(c, "User ID not found")
	}

	if err := h.subscriptionService.CancelSubscription(c.Context(), uid, req.SubscriptionID, "user_requested"); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Subscription cancelled successfully", nil)
}

func (h *SubscriptionHandler) GetInvoices(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	uid, ok := userID.(string)
	if !ok {
		return utils.Unauthorized(c, "User ID not found")
	}

	invoices, total, err := h.subscriptionService.GetInvoices(c.Context(), uid, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, fiber.Map{
		"invoices": invoices,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *SubscriptionHandler) Subscribe(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req struct {
		PlanID      string `json:"plan_id" validate:"required"`
		PhoneNumber string `json:"phone_number"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	uid, ok := userID.(string)
	if !ok {
		return utils.Unauthorized(c, "User ID not found")
	}

	sub, chargeResp, err := h.subscriptionService.SubscribeWithNewPayment(c.Context(), uid, req.PlanID, req.PhoneNumber)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Payment initiated", fiber.Map{
		"subscription": sub,
		"charge":       chargeResp,
	})
}

func (h *SubscriptionHandler) Webhook(c *fiber.Ctx) error {
	var payload models.IntaSendWebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.subscriptionService.HandleWebhook(c.Context(), &payload); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, nil)
}

func (h *SubscriptionHandler) VerifyPayment(c *fiber.Ctx) error {
	transactionID := c.Query("transaction_id")
	reference := c.Query("reference")

	payment, err := h.subscriptionService.VerifyPayment(c.Context(), transactionID, reference)
	if err != nil {
		return utils.NotFound(c, "Payment")
	}

	return utils.Success(c, payment)
}

func (h *SubscriptionHandler) ReactivateSubscription(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var req struct {
		SubscriptionID string `json:"subscription_id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	uid, ok := userID.(string)
	if !ok {
		return utils.Unauthorized(c, "User ID not found")
	}

	sub, err := h.subscriptionService.ReactivateSubscription(c.Context(), uid, req.SubscriptionID)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Subscription reactivated", sub)
}

func (h *SubscriptionHandler) UpgradeSubscription(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var req struct {
		SubscriptionID string `json:"subscription_id" validate:"required"`
		NewPlanID      string `json:"new_plan_id" validate:"required"`
		PhoneNumber    string `json:"phone_number"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	uid, ok := userID.(string)
	if !ok {
		return utils.Unauthorized(c, "User ID not found")
	}

	phoneNumber := req.PhoneNumber
	if phoneNumber == "" {
		pm, err := h.subscriptionService.GetPaymentMethods(c.Context(), uid)
		if err == nil {
			for _, m := range pm {
				if m.PhoneNumber != "" {
					phoneNumber = m.PhoneNumber
					break
				}
			}
		}
	}

	sub, chargeResp, err := h.subscriptionService.UpgradeWithPayment(c.Context(), uid, req.SubscriptionID, req.NewPlanID, phoneNumber)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Subscription upgraded", fiber.Map{
		"subscription": sub,
		"charge":       chargeResp,
	})
}

func (h *SubscriptionHandler) DowngradeSubscription(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var req struct {
		SubscriptionID string `json:"subscription_id" validate:"required"`
		NewPlanID      string `json:"new_plan_id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	uid, ok := userID.(string)
	if !ok {
		return utils.Unauthorized(c, "User ID not found")
	}

	sub, err := h.subscriptionService.DowngradeSubscription(c.Context(), uid, req.SubscriptionID, req.NewPlanID)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Subscription downgraded", sub)
}

func (h *SubscriptionHandler) GetPaymentMethods(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	uid, ok := userID.(string)
	if !ok {
		return utils.Unauthorized(c, "User ID not found")
	}

	methods, err := h.subscriptionService.GetPaymentMethods(c.Context(), uid)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, methods)
}

func (h *SubscriptionHandler) AddPaymentMethod(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var req services.AddPaymentMethodRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	uid, ok := userID.(string)
	if !ok {
		return utils.Unauthorized(c, "User ID not found")
	}

	pm, err := h.subscriptionService.AddPaymentMethod(c.Context(), uid, &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessCreated(c, "Payment method added", pm)
}

func (h *SubscriptionHandler) RemovePaymentMethod(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	methodID := c.Params("id")

	uid, ok := userID.(string)
	if !ok {
		return utils.Unauthorized(c, "User ID not found")
	}

	if err := h.subscriptionService.RemovePaymentMethod(c.Context(), uid, methodID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Payment method removed", nil)
}

func (h *SubscriptionHandler) SetDefaultPaymentMethod(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	methodID := c.Params("id")

	uid, ok := userID.(string)
	if !ok {
		return utils.Unauthorized(c, "User ID not found")
	}

	if err := h.subscriptionService.SetDefaultPaymentMethod(c.Context(), uid, methodID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Default payment method updated", nil)
}

func (h *SubscriptionHandler) GetProration(c *fiber.Ctx) error {
	var req struct {
		SubscriptionID string `json:"subscription_id" validate:"required"`
		NewPlanID      string `json:"new_plan_id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	proration, err := h.subscriptionService.CalculateProration(c.Context(), req.SubscriptionID, req.NewPlanID)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Success(c, proration)
}

func (h *SubscriptionHandler) RetryPayment(c *fiber.Ctx) error {
	var req struct {
		SubscriptionID string `json:"subscription_id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.subscriptionService.RetryPayment(c.Context(), req.SubscriptionID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Payment retry initiated", nil)
}
