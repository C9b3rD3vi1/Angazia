package handlers

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type AlertHandler struct {
	alertService services.AlertService
	validator    *validator.Validate
}

func NewAlertHandler(alertService services.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
		validator:    validator.New(),
	}
}

func (h *AlertHandler) CreateSavedSearch(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req services.CreateSavedSearchRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	search, err := h.alertService.CreateSavedSearch(c.Context(), userID.(string), &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessCreated(c, "Saved search created successfully", search)
}

func (h *AlertHandler) ListSavedSearches(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	result, err := h.alertService.ListSavedSearches(c.Context(), userID.(string), page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, result)
}

func (h *AlertHandler) GetSavedSearch(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Search ID is required")
	}

	search, err := h.alertService.GetSavedSearch(c.Context(), id, userID.(string))
	if err != nil {
		return utils.NotFound(c, "Saved search")
	}

	return utils.Success(c, search)
}

func (h *AlertHandler) UpdateSavedSearch(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Search ID is required")
	}

	var req services.UpdateSavedSearchRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	search, err := h.alertService.UpdateSavedSearch(c.Context(), id, userID.(string), &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Saved search updated successfully", search)
}

func (h *AlertHandler) DeleteSavedSearch(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Search ID is required")
	}

	if err := h.alertService.DeleteSavedSearch(c.Context(), id, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Saved search deleted successfully", nil)
}

func (h *AlertHandler) TestAlert(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Search ID is required")
	}

	result, err := h.alertService.SendTestAlert(c.Context(), id, userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Test alert sent successfully", result)
}

func (h *AlertHandler) GetAlertSettings(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	settings, err := h.alertService.GetAlertSettings(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, settings)
}

func (h *AlertHandler) UpdateAlertSettings(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req services.UpdateAlertSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	settings, err := h.alertService.UpdateAlertSettings(c.Context(), userID.(string), &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Alert settings updated successfully", settings)
}

func (h *AlertHandler) GetAlertHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	days, _ := strconv.Atoi(c.Query("days", "30"))

	history, err := h.alertService.GetAlertHistory(c.Context(), userID.(string), days)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, history)
}
