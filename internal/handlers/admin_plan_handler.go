// internal/handlers/admin_plan_handler.go
package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	
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
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	var req services.CreatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "validation_failed",
			Message: err.Error(),
		})
	}
	
	plan, err := h.subscriptionService.CreatePlan(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: "Plan created successfully",
		Data:    plan,
	})
}

// GetPlans returns all subscription plans
func (h *AdminPlanHandler) GetPlans(c *fiber.Ctx) error {
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

// GetPlan returns a specific plan
func (h *AdminPlanHandler) GetPlan(c *fiber.Ctx) error {
	planID := c.Params("id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_id",
			Message: "Plan ID is required",
		})
	}
	
	plan, err := h.subscriptionService.GetPlanByID(c.Context(), planID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Error:   "not_found",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    plan,
	})
}

// UpdatePlan updates a subscription plan
func (h *AdminPlanHandler) UpdatePlan(c *fiber.Ctx) error {
	adminID := c.Locals("user_id")
	if adminID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	planID := c.Params("id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_id",
			Message: "Plan ID is required",
		})
	}
	
	var req services.UpdatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	plan, err := h.subscriptionService.UpdatePlan(c.Context(), planID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "update_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Plan updated successfully",
		Data:    plan,
	})
}

// DeletePlan deletes a subscription plan
func (h *AdminPlanHandler) DeletePlan(c *fiber.Ctx) error {
	adminID := c.Locals("user_id")
	if adminID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	planID := c.Params("id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_id",
			Message: "Plan ID is required",
		})
	}
	
	if err := h.subscriptionService.DeletePlan(c.Context(), planID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "deletion_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Plan deleted successfully",
	})
}

// TogglePlanActive enables/disables a plan
func (h *AdminPlanHandler) TogglePlanActive(c *fiber.Ctx) error {
	adminID := c.Locals("user_id")
	if adminID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	planID := c.Params("id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_id",
			Message: "Plan ID is required",
		})
	}
	
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	_, err := h.subscriptionService.UpdatePlan(c.Context(), planID, &services.UpdatePlanRequest{
		IsActive: &req.IsActive,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "update_failed",
			Message: err.Error(),
		})
	}
	
	status := "disabled"
	if req.IsActive {
		status = "enabled"
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Plan " + status + " successfully",
	})
}