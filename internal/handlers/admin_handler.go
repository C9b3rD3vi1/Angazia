package handlers

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type AdminHandler struct {
	adminService services.AdminService
	validator    *validator.Validate
}

func NewAdminHandler(adminService services.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		validator:    validator.New(),
	}
}

// GetPlatformStats returns overall platform statistics
func (h *AdminHandler) GetPlatformStats(c *fiber.Ctx) error {
	stats, err := h.adminService.GetPlatformStats(c.Context())
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.Success(c, stats)
}

// GetUserStats returns user statistics
func (h *AdminHandler) GetUserStats(c *fiber.Ctx) error {
	stats, err := h.adminService.GetUserStats(c.Context())
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.Success(c, stats)
}

// GetJobStats returns job statistics
func (h *AdminHandler) GetJobStats(c *fiber.Ctx) error {
	stats, err := h.adminService.GetJobStats(c.Context())
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.Success(c, stats)
}

// GetEngagementStats returns engagement statistics
func (h *AdminHandler) GetEngagementStats(c *fiber.Ctx) error {
	stats, err := h.adminService.GetEngagementStats(c.Context())
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.Success(c, stats)
}

// GetAllUsers returns a paginated list of all users
func (h *AdminHandler) GetAllUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	filters := make(map[string]interface{})
	if role := c.Query("role"); role != "" {
		filters["role"] = role
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if isActive := c.Query("is_active"); isActive != "" {
		filters["is_active"] = isActive == "true"
	}
	if isVerified := c.Query("is_verified"); isVerified != "" {
		filters["is_verified"] = isVerified == "true"
	}

	users, total, err := h.adminService.GetAllUsers(c.Context(), filters, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return utils.Success(c, fiber.Map{
		"users":       users,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// GetUserDetails returns detailed information about a specific user
func (h *AdminHandler) GetUserDetails(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return utils.BadRequest(c, "User ID is required")
	}

	user, err := h.adminService.GetUserDetails(c.Context(), userID)
	if err != nil {
		return utils.NotFound(c, "User")
	}
	return utils.Success(c, user)
}

// SuspendUser suspends a user account
func (h *AdminHandler) SuspendUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return utils.BadRequest(c, "User ID is required")
	}

	if err := h.adminService.SuspendUser(c.Context(), userID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "User suspended successfully", nil)
}

// ActivateUser activates a suspended user account
func (h *AdminHandler) ActivateUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return utils.BadRequest(c, "User ID is required")
	}

	if err := h.adminService.ActivateUser(c.Context(), userID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "User activated successfully", nil)
}

// DeleteUser permanently deletes a user account
func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return utils.BadRequest(c, "User ID is required")
	}

	if err := h.adminService.DeleteUser(c.Context(), userID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "User deleted successfully", nil)
}

// VerifyUser marks an employer as verified
func (h *AdminHandler) VerifyUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return utils.BadRequest(c, "User ID is required")
	}

	if err := h.adminService.VerifyUser(c.Context(), userID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "User verified successfully", nil)
}

// GetModerationQueue returns the moderation queue
func (h *AdminHandler) GetModerationQueue(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	entityType := c.Query("entity_type")
	status := c.Query("status", "pending")

	items, total, err := h.adminService.GetModerationQueue(c.Context(), entityType, status, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return utils.Success(c, fiber.Map{
		"items":       items,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// ApproveContent approves a moderation item
func (h *AdminHandler) ApproveContent(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Moderation item ID is required")
	}

	adminID := c.Locals("user_id")
	if adminID == nil {
		return utils.Unauthorized(c, "Admin not authenticated")
	}

	if err := h.adminService.ApproveContent(c.Context(), id, adminID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Content approved successfully", nil)
}

// RejectContent rejects a moderation item
func (h *AdminHandler) RejectContent(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Moderation item ID is required")
	}

	adminID := c.Locals("user_id")
	if adminID == nil {
		return utils.Unauthorized(c, "Admin not authenticated")
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.adminService.RejectContent(c.Context(), id, adminID.(string), req.Reason); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Content rejected successfully", nil)
}

// GetSettings returns system settings
func (h *AdminHandler) GetSettings(c *fiber.Ctx) error {
	category := c.Query("category")

	settings, err := h.adminService.GetSettings(c.Context(), category)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.Success(c, settings)
}

// UpdateSetting updates a system setting
func (h *AdminHandler) UpdateSetting(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return utils.BadRequest(c, "Setting key is required")
	}

	var req struct {
		Value string `json:"value" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.adminService.UpdateSetting(c.Context(), key, req.Value); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Setting updated successfully", nil)
}

// GetReportReasons returns available report reasons
func (h *AdminHandler) GetReportReasons(c *fiber.Ctx) error {
	entityType := c.Query("entity_type")

	reasons, err := h.adminService.GetReportReasons(c.Context(), entityType)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.Success(c, reasons)
}

// GetAuditLogs returns admin action audit logs
func (h *AdminHandler) GetAuditLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	filters := make(map[string]interface{})
	if adminID := c.Query("admin_id"); adminID != "" {
		filters["admin_id"] = adminID
	}
	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}
	if entityType := c.Query("entity_type"); entityType != "" {
		filters["entity_type"] = entityType
	}
	if entityID := c.Query("entity_id"); entityID != "" {
		filters["entity_id"] = entityID
	}

	logs, total, err := h.adminService.GetAuditLogs(c.Context(), filters, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return utils.Success(c, fiber.Map{
		"logs":        logs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// ReportContent allows authenticated users to report content
func (h *AdminHandler) ReportContent(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req services.ReportContentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.adminService.ReportContent(c.Context(), userID.(string), req); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessCreated(c, "Content reported successfully", nil)
}

// ApproveCompanyVerification approves a company verification request
func (h *AdminHandler) ApproveCompanyVerification(c *fiber.Ctx) error {
	adminID, _ := c.Locals("user_id").(string)
	if adminID == "" {
		return utils.Unauthorized(c, "Admin not authenticated")
	}
	companyID := c.Params("id")
	if companyID == "" {
		return utils.BadRequest(c, "Company ID required")
	}
	if err := h.adminService.ApproveCompanyVerification(c.Context(), adminID, companyID); err != nil {
		switch err.Error() {
		case "verification not found":
			return utils.NotFound(c, "Verification")
		case "verification is not pending":
			return utils.Conflict(c, "Verification is not pending")
		default:
			return utils.InternalServerError(c, err.Error())
		}
	}
	return utils.SuccessWithMessage(c, "Company verification approved", nil)
}

// RejectCompanyVerification rejects a company verification request
func (h *AdminHandler) RejectCompanyVerification(c *fiber.Ctx) error {
	adminID, _ := c.Locals("user_id").(string)
	if adminID == "" {
		return utils.Unauthorized(c, "Admin not authenticated")
	}
	companyID := c.Params("id")
	if companyID == "" {
		return utils.BadRequest(c, "Company ID required")
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}
	if err := h.adminService.RejectCompanyVerification(c.Context(), adminID, companyID, req.Reason); err != nil {
		switch err.Error() {
		case "verification not found":
			return utils.NotFound(c, "Verification")
		case "verification is not pending":
			return utils.Conflict(c, "Verification is not pending")
		default:
			return utils.InternalServerError(c, err.Error())
		}
	}
	return utils.SuccessWithMessage(c, "Company verification rejected", nil)
}
