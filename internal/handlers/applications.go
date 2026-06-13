package handlers

import (
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type ApplicationHandler struct {
	applicationService services.ApplicationService
	validator          *validator.Validate
}

type ApplyRequest struct {
	JobID       string `json:"job_id" validate:"required"`
	CoverLetter string `json:"cover_letter"`
	ResumeURL   string `json:"resume_url"`
	PortfolioURL string `json:"portfolio_url"`
}

type ScheduleInterviewRequest struct {
	InterviewDate time.Time `json:"interview_date" validate:"required"`
	InterviewType string    `json:"interview_type" validate:"required,oneof=phone technical onsite final"`
	Notes         string    `json:"notes"`
}

func NewApplicationHandler(applicationService services.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{
		applicationService: applicationService,
		validator:          validator.New(),
	}
}

// Apply submits a job application
// @Summary Apply for a job
// @Description Submit application for a job
// @Tags Applications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ApplyRequest true "Application details"
// @Success 201 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Router /applications [post]
func (h *ApplicationHandler) Apply(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	var req ApplyRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	application, err := h.applicationService.Apply(c.Context(), userID.(string), &services.ApplyRequest{
		JobID:       req.JobID,
		CoverLetter: req.CoverLetter,
		ResumeURL:   req.ResumeURL,
		PortfolioURL: req.PortfolioURL,
	})
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "job not found" {
			status = fiber.StatusNotFound
		} else if err.Error() == "you have already applied for this job" {
			status = fiber.StatusConflict
		}
		return utils.Error(c, status, err.Error())
	}
	
	return utils.SuccessCreated(c, "Application submitted successfully", application)
}

// WithdrawApplication withdraws an application
// @Summary Withdraw application
// @Description Withdraw a submitted application
// @Tags Applications
// @Security BearerAuth
// @Param id path string true "Application ID"
// @Success 200 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Router /applications/{id}/withdraw [post]
func (h *ApplicationHandler) WithdrawApplication(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	applicationID := c.Params("id")
	if applicationID == "" {
		return utils.BadRequest(c, "Application ID is required")
	}
	
	if err := h.applicationService.WithdrawApplication(c.Context(), applicationID, userID.(string)); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "application not found" {
			status = fiber.StatusNotFound
		}
		return utils.Error(c, status, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Application withdrawn successfully", nil)
}

// GetApplication retrieves application details
// @Summary Get application details
// @Description Get detailed application information
// @Tags Applications
// @Security BearerAuth
// @Param id path string true "Application ID"
// @Success 200 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Router /applications/{id} [get]
func (h *ApplicationHandler) GetApplication(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	role := c.Locals("user_role").(string)
	
	applicationID := c.Params("id")
	if applicationID == "" {
		return utils.BadRequest(c, "Application ID is required")
	}
	
	application, err := h.applicationService.GetApplication(c.Context(), applicationID, userID.(string), role)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "application not found" {
			status = fiber.StatusNotFound
		} else if err.Error() == "unauthorized" {
			status = fiber.StatusForbidden
		}
		return utils.Error(c, status, err.Error())
	}
	
	return utils.Success(c, application)
}

// ListMyApplications lists applications for the authenticated candidate
// @Summary List my applications
// @Description Get all applications submitted by the candidate
// @Tags Applications
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} APIResponse
// @Router /employee/applications [get]
func (h *ApplicationHandler) ListMyApplications(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	result, err := h.applicationService.ListMyApplications(c.Context(), userID.(string), page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, result)
}

// ListJobApplications lists applications for a specific job (employer only)
// @Summary List job applications
// @Description Get all applications for a specific job
// @Tags Applications
// @Security BearerAuth
// @Param jobId path string true "Job ID"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} APIResponse
// @Router /employer/jobs/{jobId}/applications [get]
func (h *ApplicationHandler) ListJobApplications(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	jobID := c.Params("jobId")
	if jobID == "" {
		return utils.BadRequest(c, "Job ID is required")
	}
	
	status := c.Query("status", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	result, err := h.applicationService.ListJobApplications(c.Context(), jobID, userID.(string), status, page, limit)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "unauthorized" {
			statusCode = fiber.StatusForbidden
		}
		return utils.Error(c, statusCode, err.Error())
	}
	
	return utils.Success(c, result)
}

// ListCompanyApplications lists all applications for the employer's company
// @Summary List company applications
// @Description Get all applications for all jobs posted by the employer
// @Tags Applications
// @Security BearerAuth
// @Param status query string false "Filter by status"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} APIResponse
// @Router /employer/applications [get]
func (h *ApplicationHandler) ListCompanyApplications(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	status := c.Query("status", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	result, err := h.applicationService.ListCompanyApplications(c.Context(), userID.(string), status, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, result)
}

// ShortlistApplication shortlists an application
// @Summary Shortlist application
// @Description Mark an application as shortlisted
// @Tags Applications
// @Security BearerAuth
// @Param id path string true "Application ID"
// @Param notes body string false "Employer notes"
// @Success 200 {object} APIResponse
// @Router /employer/applications/{id}/shortlist [post]
func (h *ApplicationHandler) ShortlistApplication(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	applicationID := c.Params("id")
	if applicationID == "" {
		return utils.BadRequest(c, "Application ID is required")
	}
	
	var req struct {
		Notes string `json:"notes"`
	}
	c.BodyParser(&req)
	
	if err := h.applicationService.ShortlistApplication(c.Context(), applicationID, userID.(string), req.Notes); err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "application not found" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "unauthorized" {
			statusCode = fiber.StatusForbidden
		}
		return utils.Error(c, statusCode, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Application shortlisted successfully", nil)
}

// RejectApplication rejects an application
// @Summary Reject application
// @Description Mark an application as rejected
// @Tags Applications
// @Security BearerAuth
// @Param id path string true "Application ID"
// @Param notes body string false "Rejection reason"
// @Success 200 {object} APIResponse
// @Router /employer/applications/{id}/reject [post]
func (h *ApplicationHandler) RejectApplication(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	applicationID := c.Params("id")
	if applicationID == "" {
		return utils.BadRequest(c, "Application ID is required")
	}
	
	var req struct {
		Notes string `json:"notes"`
	}
	c.BodyParser(&req)
	
	if err := h.applicationService.RejectApplication(c.Context(), applicationID, userID.(string), req.Notes); err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "application not found" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "unauthorized" {
			statusCode = fiber.StatusForbidden
		}
		return utils.Error(c, statusCode, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Application rejected", nil)
}

// SaveNotes saves employer notes on an application
// @Summary Save employer notes
// @Description Save employer notes for an application
// @Tags Applications
// @Security BearerAuth
// @Param id path string true "Application ID"
// @Param notes body string false "Employer notes"
// @Success 200 {object} APIResponse
// @Router /employer/applications/{id}/notes [post]
func (h *ApplicationHandler) SaveNotes(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	applicationID := c.Params("id")
	if applicationID == "" {
		return utils.BadRequest(c, "Application ID is required")
	}

	var req struct {
		Notes string `json:"notes"`
	}
	c.BodyParser(&req)

	if err := h.applicationService.SaveNotes(c.Context(), applicationID, userID.(string), req.Notes); err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "application not found" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "unauthorized: you don't own this job" {
			statusCode = fiber.StatusForbidden
		}
		return utils.Error(c, statusCode, err.Error())
	}

	return utils.SuccessWithMessage(c, "Notes saved", nil)
}

// ScheduleInterview schedules an interview
// @Summary Schedule interview
// @Description Schedule an interview for an application
// @Tags Applications
// @Security BearerAuth
// @Param id path string true "Application ID"
// @Param request body ScheduleInterviewRequest true "Interview details"
// @Success 200 {object} APIResponse
// @Router /employer/applications/{id}/interview [post]
func (h *ApplicationHandler) ScheduleInterview(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	applicationID := c.Params("id")
	if applicationID == "" {
		return utils.BadRequest(c, "Application ID is required")
	}
	
	var req ScheduleInterviewRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.applicationService.ScheduleInterview(c.Context(), applicationID, userID.(string), req.InterviewDate, req.InterviewType); err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "application not found" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "unauthorized" {
			statusCode = fiber.StatusForbidden
		}
		return utils.Error(c, statusCode, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Interview scheduled successfully", nil)
}

// GetApplicationStats returns application statistics
// @Summary Get application statistics
// @Description Get statistics about applications
// @Tags Applications
// @Security BearerAuth
// @Success 200 {object} APIResponse
// @Router /applications/stats [get]
func (h *ApplicationHandler) GetApplicationStats(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	role := c.Locals("user_role").(string)
	
	stats, err := h.applicationService.GetApplicationStats(c.Context(), userID.(string), role)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, stats)
}

// BulkShortlist bulk shortlists applications
// @Summary Bulk shortlist applications
// @Description Shortlist multiple applications at once
// @Tags Applications
// @Security BearerAuth
// @Param ids body []string true "Application IDs"
// @Success 200 {object} APIResponse
// @Router /employer/applications/bulk-shortlist [post]
func (h *ApplicationHandler) BulkShortlist(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	var req struct {
		ApplicationIDs []string `json:"application_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if len(req.ApplicationIDs) == 0 {
		return utils.BadRequest(c, "No application IDs provided")
	}
	
	if err := h.applicationService.BulkShortlist(c.Context(), req.ApplicationIDs, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Applications shortlisted successfully", nil)
}

// MarkAsHired marks an application as hired
// @Summary Mark application as hired
// @Tags Applications
// @Security BearerAuth
// @Param id path string true "Application ID"
// @Success 200 {object} APIResponse
// @Router /employer/applications/{id}/hire [post]
func (h *ApplicationHandler) MarkAsHired(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	applicationID := c.Params("id")
	if applicationID == "" {
		return utils.BadRequest(c, "Application ID is required")
	}
	
	if err := h.applicationService.MarkAsHired(c.Context(), applicationID, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Candidate marked as hired", nil)
}

// BulkReject bulk rejects applications
// @Summary Bulk reject applications
// @Tags Applications
// @Security BearerAuth
// @Param ids body object true "Application IDs"
// @Success 200 {object} APIResponse
// @Router /employer/applications/bulk-reject [post]
func (h *ApplicationHandler) BulkReject(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	var req struct {
		ApplicationIDs []string `json:"application_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if len(req.ApplicationIDs) == 0 {
		return utils.BadRequest(c, "No application IDs provided")
	}
	
	if err := h.applicationService.BulkReject(c.Context(), req.ApplicationIDs, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Applications rejected successfully", nil)
}
