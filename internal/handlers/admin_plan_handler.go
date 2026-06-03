package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type AdminPlanHandler struct {
	subscriptionService services.SubscriptionService
	validator           *validator.Validate
}

func NewAdminPlanHandler(subscriptionService services.SubscriptionService) *AdminPlanHandler {
	return &AdminPlanHandler{
		subscriptionService: subscriptionService,
		validator:           validator.New(),
	}
}

// CreatePlan creates a new subscription plan
func (h *AdminPlanHandler) CreatePlan(c *fiber.Ctx) error {
	adminID := c.Locals("user_id")
	if adminID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req services.CreatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	plan, err := h.subscriptionService.CreatePlan(c.Context(), &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessCreated(c, "Plan created successfully", plan)
}

// GetPlans returns all subscription plans
func (h *AdminPlanHandler) GetPlans(c *fiber.Ctx) error {
	plans, err := h.subscriptionService.GetPlans(c.Context())
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, plans)
}

// GetPlan returns a specific plan
func (h *AdminPlanHandler) GetPlan(c *fiber.Ctx) error {
	planID := c.Params("id")
	if planID == "" {
		return utils.BadRequest(c, "Plan ID is required")
	}

	plan, err := h.subscriptionService.GetPlanByID(c.Context(), planID)
	if err != nil {
		return utils.NotFound(c, "Plan")
	}

	return utils.Success(c, plan)
}

// UpdatePlan updates a subscription plan
func (h *AdminPlanHandler) UpdatePlan(c *fiber.Ctx) error {
	adminID := c.Locals("user_id")
	if adminID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	planID := c.Params("id")
	if planID == "" {
		return utils.BadRequest(c, "Plan ID is required")
	}

	var req services.UpdatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	plan, err := h.subscriptionService.UpdatePlan(c.Context(), planID, &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Plan updated successfully", plan)
}

// DeletePlan deletes a subscription plan
func (h *AdminPlanHandler) DeletePlan(c *fiber.Ctx) error {
	adminID := c.Locals("user_id")
	if adminID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	planID := c.Params("id")
	if planID == "" {
		return utils.BadRequest(c, "Plan ID is required")
	}

	if err := h.subscriptionService.DeletePlan(c.Context(), planID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Plan deleted successfully", nil)
}

// TogglePlanActive enables/disables a plan
func (h *AdminPlanHandler) TogglePlanActive(c *fiber.Ctx) error {
	adminID := c.Locals("user_id")
	if adminID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	planID := c.Params("id")
	if planID == "" {
		return utils.BadRequest(c, "Plan ID is required")
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	_, err := h.subscriptionService.UpdatePlan(c.Context(), planID, &services.UpdatePlanRequest{
		IsActive: &req.IsActive,
	})
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	status := "disabled"
	if req.IsActive {
		status = "enabled"
	}

	return utils.SuccessWithMessage(c, "Plan "+status+" successfully", nil)
}
