package handlers

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
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

func (h *JobHandler) CreateJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req services.CreateJobRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	job, err := h.jobService.CreateJob(c.Context(), userID.(string), &req)
	if err != nil {
		if err.Error() == "free plan limit reached. Upgrade to post more jobs" {
			return utils.Error(c, fiber.StatusPaymentRequired, err.Error())
		}
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessCreated(c, "Job posted successfully", job)
}

func (h *JobHandler) GetJob(c *fiber.Ctx) error {
	jobID := c.Params("id")
	if jobID == "" {
		return utils.BadRequest(c, "Job ID is required")
	}

	go h.jobService.IncrementJobViews(c.Context(), jobID)

	job, err := h.jobService.GetJob(c.Context(), jobID)
	if err != nil {
		return utils.NotFound(c, err.Error())
	}

	return utils.Success(c, job)
}

func (h *JobHandler) UpdateJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	jobID := c.Params("id")
	if jobID == "" {
		return utils.BadRequest(c, "Job ID is required")
	}

	var req services.UpdateJobRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	job, err := h.jobService.UpdateJob(c.Context(), jobID, userID.(string), &req)
	if err != nil {
		if err.Error() == "job not found" {
			return utils.NotFound(c, err.Error())
		}
		if err.Error() == "unauthorized: you don't own this job" {
			return utils.Forbidden(c, err.Error())
		}
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Job updated successfully", job)
}

func (h *JobHandler) DeleteJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	jobID := c.Params("id")
	if jobID == "" {
		return utils.BadRequest(c, "Job ID is required")
	}

	err := h.jobService.DeleteJob(c.Context(), jobID, userID.(string))
	if err != nil {
		if err.Error() == "job not found" {
			return utils.NotFound(c, err.Error())
		}
		if err.Error() == "unauthorized: you don't own this job" {
			return utils.Forbidden(c, err.Error())
		}
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Job deleted successfully", nil)
}

func (h *JobHandler) CloseJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	jobID := c.Params("id")
	if jobID == "" {
		return utils.BadRequest(c, "Job ID is required")
	}

	err := h.jobService.CloseJob(c.Context(), jobID, userID.(string))
	if err != nil {
		if err.Error() == "job not found" {
			return utils.NotFound(c, err.Error())
		}
		if err.Error() == "unauthorized: you don't own this job" {
			return utils.Forbidden(c, err.Error())
		}
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Job closed successfully", nil)
}

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

	result, err := h.jobService.ListJobs(c.Context(), filters, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, result)
}

func (h *JobHandler) ListMyJobs(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	result, err := h.jobService.ListMyJobs(c.Context(), userID.(string), page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, result)
}

func (h *JobHandler) GetFeaturedJobs(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	jobs, err := h.jobService.GetFeaturedJobs(c.Context(), limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, jobs)
}

func (h *JobHandler) GetRecentJobs(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	jobs, err := h.jobService.GetRecentJobs(c.Context(), limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, jobs)
}

func (h *JobHandler) GetSimilarJobs(c *fiber.Ctx) error {
	jobID := c.Params("id")
	if jobID == "" {
		return utils.BadRequest(c, "Job ID is required")
	}

	limit, _ := strconv.Atoi(c.Query("limit", "5"))

	jobs, err := h.jobService.GetSimilarJobs(c.Context(), jobID, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, jobs)
}

func (h *JobHandler) SearchJobs(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return utils.BadRequest(c, "Search query is required")
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
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, result)
}

func (h *JobHandler) SaveJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	jobID := c.Params("id")
	if jobID == "" {
		return utils.BadRequest(c, "Job ID is required")
	}

	err := h.jobService.SaveJob(c.Context(), userID.(string), jobID)
	if err != nil {
		if err.Error() == "job not found" {
			return utils.NotFound(c, err.Error())
		}
		if err.Error() == "job already saved" {
			return utils.Conflict(c, err.Error())
		}
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Job saved successfully", nil)
}

func (h *JobHandler) UnsaveJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	jobID := c.Params("id")
	if jobID == "" {
		return utils.BadRequest(c, "Job ID is required")
	}

	if err := h.jobService.UnsaveJob(c.Context(), userID.(string), jobID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Job removed from saved", nil)
}

func (h *JobHandler) GetSavedJobs(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	result, err := h.jobService.GetSavedJobs(c.Context(), userID.(string), page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, result)
}
