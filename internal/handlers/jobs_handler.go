package handlers

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type JobHandler struct {
	jobService services.JobService
	validator  *validator.Validate
}

func NewJobHandler(jobService services.JobService) *JobHandler {
	return &JobHandler{
		jobService: jobService,
		validator:  validator.New(),
	}
}

// CreateJob handles job posting
// @Summary Post a new job
// @Description Create a new job listing
// @Tags Jobs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body services.CreateJobRequest true "Job details"
// @Success 201 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Router /employer/jobs [post]
func (h *JobHandler) CreateJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	var req services.CreateJobRequest
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
	
	job, err := h.jobService.CreateJob(c.Context(), userID.(string), &req)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "free plan limit reached. Upgrade to post more jobs" {
			status = fiber.StatusPaymentRequired
		}
		return c.Status(status).JSON(APIResponse{
			Success: false,
			Error:   "job_creation_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: "Job posted successfully",
		Data:    job,
	})
}

// GetJob retrieves a job by ID
// @Summary Get job details
// @Description Get detailed job information
// @Tags Jobs
// @Param id path string true "Job ID"
// @Success 200 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Router /jobs/{id} [get]
func (h *JobHandler) GetJob(c *fiber.Ctx) error {
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_id",
			Message: "Job ID is required",
		})
	}
	
	// Increment view count asynchronously
	go h.jobService.IncrementJobViews(c.Context(), jobID)
	
	job, err := h.jobService.GetJob(c.Context(), jobID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Error:   "not_found",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    job,
	})
}

// UpdateJob updates an existing job
// @Summary Update job
// @Description Update job details
// @Tags Jobs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param request body services.UpdateJobRequest true "Job updates"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 403 {object} APIResponse
// @Router /employer/jobs/{id} [put]
func (h *JobHandler) UpdateJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_id",
			Message: "Job ID is required",
		})
	}
	
	var req services.UpdateJobRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	job, err := h.jobService.UpdateJob(c.Context(), jobID, userID.(string), &req)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "job not found" {
			status = fiber.StatusNotFound
		} else if err.Error() == "unauthorized: you don't own this job" {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(APIResponse{
			Success: false,
			Error:   "job_update_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Job updated successfully",
		Data:    job,
	})
}

// DeleteJob deletes a job
// @Summary Delete job
// @Description Delete a job listing
// @Tags Jobs
// @Security BearerAuth
// @Param id path string true "Job ID"
// @Success 200 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 403 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Router /employer/jobs/{id} [delete]
func (h *JobHandler) DeleteJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_id",
			Message: "Job ID is required",
		})
	}
	
	if err := h.jobService.DeleteJob(c.Context(), jobID, userID.(string)); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "job not found" {
			status = fiber.StatusNotFound
		} else if err.Error() == "unauthorized: you don't own this job" {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(APIResponse{
			Success: false,
			Error:   "job_deletion_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Job deleted successfully",
	})
}

// CloseJob closes a job (marks as inactive)
// @Summary Close job
// @Description Close a job listing
// @Tags Jobs
// @Security BearerAuth
// @Param id path string true "Job ID"
// @Success 200 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 403 {object} APIResponse
// @Router /employer/jobs/{id}/close [post]
func (h *JobHandler) CloseJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_id",
			Message: "Job ID is required",
		})
	}
	
	if err := h.jobService.CloseJob(c.Context(), jobID, userID.(string)); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "job not found" {
			status = fiber.StatusNotFound
		} else if err.Error() == "unauthorized: you don't own this job" {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(APIResponse{
			Success: false,
			Error:   "job_close_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Job closed successfully",
	})
}

// ListJobs lists all active jobs with filters
// @Summary List jobs
// @Description Get paginated list of jobs with filters
// @Tags Jobs
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param title query string false "Job title search"
// @Param location query string false "Location filter"
// @Param remote query bool false "Remote jobs only"
// @Param employment_type query string false "Employment type"
// @Param experience_level query string false "Experience level"
// @Param min_salary query int false "Minimum salary"
// @Param max_salary query int false "Maximum salary"
// @Param skills query []string false "Required skills"
// @Success 200 {object} APIResponse
// @Router /jobs [get]
func (h *JobHandler) ListJobs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	remote := c.Query("remote")
	var isRemote *bool
	if remote != "" {
		b := remote == "true"
		isRemote = &b
	}
	
	filters := &services.JobFilters{
		Title:           c.Query("title"),
		Location:        c.Query("location"),
		IsRemote:        isRemote,
		EmploymentType:  c.Query("employment_type"),
		ExperienceLevel: c.Query("experience_level"),
		MinSalary:       c.QueryInt("min_salary"),
		MaxSalary:       c.QueryInt("max_salary"),
	}
	
	// Parse skills
	if skills := c.Query("skills"); skills != "" {
		// Skills can be comma-separated
		// filters.Skills = strings.Split(skills, ",")
	}
	
	result, err := h.jobService.ListJobs(c.Context(), filters, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "list_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    result,
	})
}

// ListMyJobs lists jobs for the authenticated employer
// @Summary List my jobs
// @Description Get all jobs posted by the authenticated employer
// @Tags Jobs
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} APIResponse
// @Router /employer/jobs [get]
func (h *JobHandler) ListMyJobs(c *fiber.Ctx) error {
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
	
	result, err := h.jobService.ListMyJobs(c.Context(), userID.(string), page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "list_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetFeaturedJobs returns featured jobs
// @Summary Get featured jobs
// @Description Get list of featured job listings
// @Tags Jobs
// @Param limit query int false "Number of jobs" default(10)
// @Success 200 {object} APIResponse
// @Router /jobs/featured [get]
func (h *JobHandler) GetFeaturedJobs(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	
	jobs, err := h.jobService.GetFeaturedJobs(c.Context(), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    jobs,
	})
}

// GetRecentJobs returns recent jobs
// @Summary Get recent jobs
// @Description Get list of recent job listings
// @Tags Jobs
// @Param limit query int false "Number of jobs" default(10)
// @Success 200 {object} APIResponse
// @Router /jobs/recent [get]
func (h *JobHandler) GetRecentJobs(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	
	jobs, err := h.jobService.GetRecentJobs(c.Context(), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    jobs,
	})
}

// GetSimilarJobs returns similar jobs based on a job
// @Summary Get similar jobs
// @Description Get jobs similar to the specified job
// @Tags Jobs
// @Param id path string true "Job ID"
// @Param limit query int false "Number of jobs" default(5)
// @Success 200 {object} APIResponse
// @Router /jobs/{id}/similar [get]
func (h *JobHandler) GetSimilarJobs(c *fiber.Ctx) error {
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_id",
			Message: "Job ID is required",
		})
	}
	
	limit, _ := strconv.Atoi(c.Query("limit", "5"))
	
	jobs, err := h.jobService.GetSimilarJobs(c.Context(), jobID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    jobs,
	})
}

// SearchJobs searches jobs with full-text search
// @Summary Search jobs
// @Description Full-text search on job titles and descriptions
// @Tags Jobs
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param location query string false "Location filter"
// @Param remote query bool false "Remote jobs only"
// @Success 200 {object} APIResponse
// @Router /jobs/search [get]
func (h *JobHandler) SearchJobs(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_query",
			Message: "Search query is required",
		})
	}
	
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	remote := c.Query("remote")
	var isRemote *bool
	if remote != "" {
		b := remote == "true"
		isRemote = &b
	}
	
	filters := &services.JobFilters{
		Location: c.Query("location"),
		IsRemote: isRemote,
	}
	
	result, err := h.jobService.SearchJobs(c.Context(), query, filters, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "search_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    result,
	})
}

// SaveJob saves a job for a candidate
// @Summary Save job
// @Description Save a job to user's saved list
// @Tags Jobs
// @Security BearerAuth
// @Param id path string true "Job ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Router /jobs/{id}/save [post]
func (h *JobHandler) SaveJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_id",
			Message: "Job ID is required",
		})
	}
	
	if err := h.jobService.SaveJob(c.Context(), userID.(string), jobID); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "job not found" {
			status = fiber.StatusNotFound
		} else if err.Error() == "job already saved" {
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(APIResponse{
			Success: false,
			Error:   "save_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Job saved successfully",
	})
}

// UnsaveJob removes a saved job
// @Summary Unsave job
// @Description Remove a job from user's saved list
// @Tags Jobs
// @Security BearerAuth
// @Param id path string true "Job ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Router /jobs/{id}/save [delete]
func (h *JobHandler) UnsaveJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "missing_id",
			Message: "Job ID is required",
		})
	}
	
	if err := h.jobService.UnsaveJob(c.Context(), userID.(string), jobID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "unsave_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Job removed from saved",
	})
}

// GetSavedJobs returns candidate's saved jobs
// @Summary Get saved jobs
// @Description Get all jobs saved by the candidate
// @Tags Jobs
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} APIResponse
// @Router /employee/saved-jobs [get]
func (h *JobHandler) GetSavedJobs(c *fiber.Ctx) error {
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
	
	result, err := h.jobService.GetSavedJobs(c.Context(), userID.(string), page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "fetch_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    result,
	})
}