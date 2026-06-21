package handlers

import (
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
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
	if limit <= 0 {
		limit = 20
	}

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
		if err.Error() == "user not found" {
			return utils.NotFound(c, "User")
		}
		return utils.InternalServerError(c, err.Error())
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
	adminID, _ := c.Locals("user_id").(string)
	h.adminService.LogAdminAction(c.Context(), adminID, "suspend", "user", userID, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"is_active": false})
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
	adminID, _ := c.Locals("user_id").(string)
	h.adminService.LogAdminAction(c.Context(), adminID, "activate", "user", userID, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"is_active": true})
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
	adminID, _ := c.Locals("user_id").(string)
	h.adminService.LogAdminAction(c.Context(), adminID, "delete", "user", userID, c.IP(), c.Get("User-Agent"), nil, nil)
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
	adminID, _ := c.Locals("user_id").(string)
	h.adminService.LogAdminAction(c.Context(), adminID, "verify", "user", userID, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"is_verified": true})
	return utils.SuccessWithMessage(c, "User verified successfully", nil)
}

// GetModerationQueue returns the moderation queue
func (h *AdminHandler) GetModerationQueue(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit <= 0 {
		limit = 20
	}
	entityType := c.Query("entity_type")
	status := c.Query("status", "pending")

	var dateFrom, dateTo *time.Time
	if df := c.Query("date_from"); df != "" {
		if t, err := time.Parse("2006-01-02", df); err == nil {
			dateFrom = &t
		}
	}
	if dt := c.Query("date_to"); dt != "" {
		if t, err := time.Parse("2006-01-02", dt); err == nil {
			dateTo = &t
		}
	}

	items, total, err := h.adminService.GetModerationQueue(c.Context(), entityType, status, page, limit, dateFrom, dateTo)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	// Collect unique user IDs for enrichment
	idSet := make(map[string]struct{})
	for _, item := range items {
		if item.SubmittedBy != "" {
			idSet[item.SubmittedBy] = struct{}{}
		}
		if item.ReviewedBy != nil && *item.ReviewedBy != "" {
			idSet[*item.ReviewedBy] = struct{}{}
		}
	}
	idList := make([]string, 0, len(idSet))
	for id := range idSet {
		idList = append(idList, id)
	}

	// Batch fetch user info
	usersMap := make(map[string]*models.User)
	if len(idList) > 0 {
		usersMap, _ = h.adminService.GetUsersByIDs(c.Context(), idList)
	}

	userDisplayName := func(u *models.User) string {
		if u == nil {
			return "Unknown"
		}
		if u.EmployeeProfile != nil && u.EmployeeProfile.FullName != "" {
			return u.EmployeeProfile.FullName
		}
		if u.EmployerProfile != nil && u.EmployerProfile.CompanyName != "" {
			return u.EmployerProfile.CompanyName
		}
		return u.Email
	}

	enrichedItems := make([]fiber.Map, len(items))
	for i, item := range items {
		ei := fiber.Map{
			"id":           item.ID,
			"entity_type":  item.EntityType,
			"entity_id":    item.EntityID,
			"status":       item.Status,
			"reason":       item.Reason,
			"submitted_by": item.SubmittedBy,
			"metadata":     item.Metadata,
			"created_at":   item.CreatedAt,
			"updated_at":   item.UpdatedAt,
		}
		if item.ReviewedAt != nil {
			ei["reviewed_at"] = item.ReviewedAt
		}
		if item.ReviewedBy != nil {
			ei["reviewed_by"] = *item.ReviewedBy
		}

		// Enrich submitter
		if u, ok := usersMap[item.SubmittedBy]; ok {
			ei["submitter"] = fiber.Map{
				"name":   userDisplayName(u),
				"avatar": u.AvatarURL,
			}
		} else {
			ei["submitter"] = fiber.Map{
				"name":   "Unknown",
				"avatar": "",
			}
		}

		// Enrich reviewer
		if item.ReviewedBy != nil {
			if u, ok := usersMap[*item.ReviewedBy]; ok {
				ei["reviewer"] = fiber.Map{
					"name":   userDisplayName(u),
					"avatar": u.AvatarURL,
				}
			}
		}

		enrichedItems[i] = ei
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return utils.Success(c, fiber.Map{
		"items":       enrichedItems,
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

	adminID, ok := c.Locals("user_id").(string)
	if !ok || adminID == "" {
		return utils.Unauthorized(c, "Admin not authenticated")
	}

	if err := h.adminService.ApproveContent(c.Context(), id, adminID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	h.adminService.LogAdminAction(c.Context(), adminID, "approve", "moderation", id, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"status": "approved"})
	return utils.SuccessWithMessage(c, "Content approved successfully", nil)
}

// RejectContent rejects a moderation item
func (h *AdminHandler) RejectContent(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Moderation item ID is required")
	}

	adminID, ok := c.Locals("user_id").(string)
	if !ok || adminID == "" {
		return utils.Unauthorized(c, "Admin not authenticated")
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.adminService.RejectContent(c.Context(), id, adminID, req.Reason); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	h.adminService.LogAdminAction(c.Context(), adminID, "reject", "moderation", id, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"status": "rejected", "reason": req.Reason})
	return utils.SuccessWithMessage(c, "Content rejected successfully", nil)
}

// GetPendingModerationCount returns the count of pending moderation items
func (h *AdminHandler) GetPendingModerationCount(c *fiber.Ctx) error {
	count, err := h.adminService.GetPendingReportsCount(c.Context())
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.Success(c, fiber.Map{"count": count})
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
	adminID, _ := c.Locals("user_id").(string)
	h.adminService.LogAdminAction(c.Context(), adminID, "update", "setting", key, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"value": req.Value})
	return utils.SuccessWithMessage(c, "Setting updated successfully", nil)
}

// CreateSetting creates a new system setting
func (h *AdminHandler) CreateSetting(c *fiber.Ctx) error {
	var req struct {
		Key         string `json:"key" validate:"required"`
		Value       string `json:"value" validate:"required"`
		Type        string `json:"type"`
		Category    string `json:"category"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	if req.Type == "" {
		req.Type = "string"
	}
	if req.Category == "" {
		req.Category = "general"
	}
	if err := h.adminService.CreateSetting(c.Context(), req.Key, req.Value, req.Type, req.Category, req.Description, req.IsPublic); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	adminID, _ := c.Locals("user_id").(string)
	h.adminService.LogAdminAction(c.Context(), adminID, "create", "setting", req.Key, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"value": req.Value, "type": req.Type, "category": req.Category})
	return utils.SuccessWithMessage(c, "Setting created successfully", nil)
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

// CreateReportReason creates a new report reason
func (h *AdminHandler) CreateReportReason(c *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name" validate:"required"`
		Description string `json:"description"`
		EntityType  string `json:"entity_type"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	reason := &models.ReportReason{
		Name:        req.Name,
		Description: req.Description,
		EntityType:  req.EntityType,
		IsActive:    true,
		SortOrder:   req.SortOrder,
	}
	if err := h.adminService.CreateReportReason(c.Context(), reason); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessCreated(c, "Report reason created", reason)
}

// UpdateReportReason updates a report reason
func (h *AdminHandler) UpdateReportReason(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Reason ID is required")
	}
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		EntityType  *string `json:"entity_type"`
		IsActive    *bool   `json:"is_active"`
		SortOrder   *int    `json:"sort_order"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.EntityType != nil {
		updates["entity_type"] = *req.EntityType
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if len(updates) == 0 {
		return utils.BadRequest(c, "No fields to update")
	}
	if err := h.adminService.UpdateReportReason(c.Context(), id, updates); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Report reason updated", nil)
}

// DeleteReportReason deletes a report reason
func (h *AdminHandler) DeleteReportReason(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Reason ID is required")
	}
	if err := h.adminService.DeleteReportReason(c.Context(), id); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Report reason deleted", nil)
}

// GetAuditLogs returns admin action audit logs
func (h *AdminHandler) GetAuditLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit <= 0 {
		limit = 20
	}

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
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filters["start_date"] = t
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			filters["end_date"] = t
		}
	}

	logs, total, err := h.adminService.GetAuditLogs(c.Context(), filters, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	// Enrich logs with entity names
	entityIDsByType := make(map[string][]string)
	for _, log := range logs {
		if log.EntityID != "" {
			entityIDsByType[log.EntityType] = append(entityIDsByType[log.EntityType], log.EntityID)
		}
	}
	entityNames := make(map[string]map[string]string)
	for eType, ids := range entityIDsByType {
		names, err := h.adminService.GetEntityNames(c.Context(), eType, ids)
		if err == nil {
			entityNames[eType] = names
		}
	}

	type enrichedLog struct {
		*models.AdminActionLog
		EntityName string `json:"entity_name"`
	}

	enriched := make([]enrichedLog, len(logs))
	for i, log := range logs {
		enriched[i].AdminActionLog = log
		if names, ok := entityNames[log.EntityType]; ok {
			enriched[i].EntityName = names[log.EntityID]
		}
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return utils.Success(c, fiber.Map{
		"logs":        enriched,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// ReportContent allows authenticated users to report content
func (h *AdminHandler) ReportContent(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req services.ReportContentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.adminService.ReportContent(c.Context(), userID, req); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	h.adminService.LogAdminAction(c.Context(), userID, "report", req.EntityType, req.EntityID, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"reason_id": req.ReasonID, "description": req.Description})
	return utils.SuccessCreated(c, "Content reported successfully", nil)
}

// GetCompanies returns all companies (employers) for admin panel
func (h *AdminHandler) GetCompanies(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	if limit <= 0 {
		limit = 100
	}

	// Get all employer users
	filters := map[string]interface{}{
		"role": "employer",
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if status := c.Query("verification_status"); status != "" {
		filters["verification_status"] = status
	}
	if hasDocs := c.Query("has_documents"); hasDocs != "" {
		filters["has_documents"] = hasDocs
	}

	users, total, err := h.adminService.GetAllUsers(c.Context(), filters, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	// Transform to company format and compute stats
	companies := make([]map[string]interface{}, 0, len(users))
	stats := map[string]int{"verified": 0, "pending": 0, "rejected": 0, "unverified": 0}
	for _, user := range users {
		vs := user.VerificationStatus
		if vs == "" {
			vs = "unverified"
		}
		stats[vs]++
		company := map[string]interface{}{
			"id":                  user.ID,
			"name":                user.CompanyName,
			"email":               user.Email,
			"logo":                user.CompanyLogo,
			"verification_status": vs,
			"jobs_count":          user.JobCount,
			"created_at":          user.CreatedAt,
			"document_count":      user.DocumentCount,
		}
		companies = append(companies, company)
	}

	return utils.Success(c, fiber.Map{
		"companies":   companies,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
		"stats":       stats,
	})
}

// GetAdminJobs returns paginated jobs for admin listing
func (h *AdminHandler) GetAdminJobs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit <= 0 {
		limit = 20
	}

	filters := make(map[string]interface{})
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if empType := c.Query("employment_type"); empType != "" {
		filters["employment_type"] = empType
	}
	if expLevel := c.Query("experience_level"); expLevel != "" {
		filters["experience_level"] = expLevel
	}
	if location := c.Query("location"); location != "" {
		filters["location"] = location
	}
	if employerID := c.Query("employer_id"); employerID != "" {
		filters["employer_id"] = employerID
	}

	jobs, total, err := h.adminService.GetAdminJobs(c.Context(), filters, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return utils.Success(c, fiber.Map{
		"jobs":        jobs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// GetJobApplications returns paginated applications for an admin-viewed job
func (h *AdminHandler) GetJobApplications(c *fiber.Ctx) error {
	jobID := c.Params("id")
	if jobID == "" {
		return utils.BadRequest(c, "Job ID is required")
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit <= 0 {
		limit = 20
	}

	status := c.Query("status")

	apps, total, err := h.adminService.GetJobApplications(c.Context(), jobID, status, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return utils.Success(c, fiber.Map{
		"applications": apps,
		"total":        total,
		"page":         page,
		"limit":        limit,
		"total_pages":  totalPages,
	})
}

// GetChartData returns time-series data for dashboard charts
func (h *AdminHandler) GetChartData(c *fiber.Ctx) error {
	period, _ := strconv.Atoi(c.Query("period", "30"))

	data, err := h.adminService.GetChartData(c.Context(), period)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, data)
}

// GetPendingVerifications - Get pending verifications for admin
func (h *AdminHandler) GetPendingVerifications(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit <= 0 {
		limit = 20
	}

	result, err := h.adminService.GetPendingVerifications(c.Context(), page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, result)
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
		return utils.InternalServerError(c, err.Error())
	}
	h.adminService.LogAdminAction(c.Context(), adminID, "approve", "company", companyID, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"verification_status": "approved"})
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
		return utils.InternalServerError(c, err.Error())
	}
	h.adminService.LogAdminAction(c.Context(), adminID, "reject", "company", companyID, c.IP(), c.Get("User-Agent"), nil, map[string]interface{}{"verification_status": "rejected", "reason": req.Reason})
	return utils.SuccessWithMessage(c, "Company verification rejected", nil)
}
