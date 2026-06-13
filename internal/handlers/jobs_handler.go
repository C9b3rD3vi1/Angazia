package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type JobHandler struct {
	jobService     services.JobService
	companyService services.CompanyService
	validator      *validator.Validate
}

func NewJobHandler(jobService services.JobService, companyService services.CompanyService) *JobHandler {
	return &JobHandler{
		jobService:     jobService,
		companyService: companyService,
		validator:      validator.New(),
	}
}

// ── Page rendering handlers ──

// JobsPage renders the public job listings page
func (h *JobHandler) JobsPage(c *fiber.Ctx) error {
	return c.Render("public/jobs", fiber.Map{
		"Title":      "Tech Jobs in Kenya - Find Your Next Role",
		"ActivePage": "jobs",
	}, "layouts/base")
}

// JobDetailPage renders the public job details page
func (h *JobHandler) JobDetailPage(c *fiber.Ctx) error {
	jobID := c.Params("id")

	job, err := h.jobService.GetJob(c.Context(), jobID)
	if err != nil {
		return c.Render("public/job-detail", fiber.Map{
			"Title":      "Job Not Found",
			"JobID":      jobID,
			"ActivePage": "jobs",
		}, "layouts/base")
	}

	var requirements []string
	if job.Requirements != "" {
		requirements = strings.Split(job.Requirements, "\n")
		for i := range requirements {
			requirements[i] = strings.TrimSpace(requirements[i])
		}
	}

	var responsibilities []string
	if job.Responsibilities != "" {
		responsibilities = strings.Split(job.Responsibilities, "\n")
		for i := range responsibilities {
			responsibilities[i] = strings.TrimSpace(responsibilities[i])
		}
	}

	company, _ := h.companyService.GetCompanyProfile(c.Context(), job.EmployerID)

	companyName := ""
	if job.Employer != nil {
		companyName = job.Employer.CompanyName
	}

	jobMap := fiber.Map{
		"title":            job.Title,
		"company":          companyName,
		"companyID":        job.EmployerID,
		"location":         job.Location,
		"salary":           formatJobSalary(job),
		"type":             job.EmploymentType,
		"description":      job.Description,
		"requirements":     requirements,
		"responsibilities": responsibilities,
		"postedDate":       job.PostedAt.Format("Jan 2, 2006"),
		"expiresDate":      "",
	}

	if job.ExpiresAt != nil {
		jobMap["expiresDate"] = job.ExpiresAt.Format("Jan 2, 2006")
	}

	companyMap := fiber.Map{}
	if company != nil && company.Profile != nil {
		companyMap = fiber.Map{
			"name":     company.Profile.CompanyName,
			"logo":     company.Profile.CompanyLogo,
			"industry": company.Profile.Industry,
			"size":     company.Profile.CompanySize,
		}
	}

	similarJobs, _ := h.jobService.GetSimilarJobs(c.Context(), jobID, 5)
	var similarJobsMap []fiber.Map
	for _, sj := range similarJobs {
		companyName := ""
		if sj.Employer != nil {
			companyName = sj.Employer.CompanyName
		}
		similarJobsMap = append(similarJobsMap, fiber.Map{
			"id":      sj.ID,
			"title":   sj.Title,
			"company": companyName,
			"salary":  formatJobSalary(sj),
		})
	}

	return c.Render("public/job-detail", fiber.Map{
		"Title":       job.Title + " - Angazia",
		"JobID":       jobID,
		"ActivePage":  "jobs",
		"job":         jobMap,
		"company":     companyMap,
		"similarJobs": similarJobsMap,
	}, "layouts/base")
}

// EmployeeJobDetailPage renders the employee job detail page
func (h *JobHandler) EmployeeJobDetailPage(c *fiber.Ctx) error {
	return c.Render("employee/job-details", mergePageData(c, fiber.Map{
		"Title":      "Job Details - Angazia",
		"ActivePage": "jobs",
	}), "layouts/employee")
}

// EmployeeJobsPage renders the employee job listings page
func (h *JobHandler) EmployeeJobsPage(c *fiber.Ctx) error {
	return c.Render("employee/jobs", mergePageData(c, fiber.Map{
		"Title":      "Find Jobs - Angazia",
		"ActivePage": "jobs",
	}), "layouts/employee")
}

// EmployeeSavedJobsPage renders the employee saved jobs page
func (h *JobHandler) EmployeeSavedJobsPage(c *fiber.Ctx) error {
	return c.Render("employee/saved", mergePageData(c, fiber.Map{
		"Title":      "Saved Jobs - Angazia",
		"ActivePage": "saved",
	}), "layouts/employee")
}

// EmployeeJobAlertsPage renders the employee job alerts page
func (h *JobHandler) EmployeeJobAlertsPage(c *fiber.Ctx) error {
	return c.Render("employee/alerts", mergePageData(c, fiber.Map{
		"Title":      "Job Alerts - Angazia",
		"ActivePage": "alerts",
	}), "layouts/employee")
}

// EmployerJobsPage renders the employer's job listings page
func (h *JobHandler) EmployerJobsPage(c *fiber.Ctx) error {
	return c.Render("employer/jobs", mergePageData(c, fiber.Map{
		"Title":      "My Jobs - Angazia",
		"ActivePage": "jobs",
	}), "layouts/employer")
}

// EmployerJobPostPage renders the employer job creation page
func (h *JobHandler) EmployerJobPostPage(c *fiber.Ctx) error {
	return c.Render("employer/job-post", mergePageData(c, fiber.Map{
		"Title":      "Post a New Job - Angazia",
		"ActivePage": "jobs",
	}), "layouts/employer")
}

// EmployerJobDetailPage renders the employer's single job detail page
func (h *JobHandler) EmployerJobDetailPage(c *fiber.Ctx) error {
	jobID := c.Params("id")
	return c.Render("employer/job-detail", mergePageData(c, fiber.Map{
		"Title":      "Job Details - Angazia",
		"ActivePage": "jobs",
		"JobID":      jobID,
	}), "layouts/employer")
}

// EmployerJobEditPage renders the employer job edit page
func (h *JobHandler) EmployerJobEditPage(c *fiber.Ctx) error {
	return c.Render("employer/job-edit", mergePageData(c, fiber.Map{
		"Title":      "Edit Job - Angazia",
		"ActivePage": "jobs",
	}), "layouts/employer")
}

// EmployerJobApplicationsPage renders the applications for a job
func (h *JobHandler) EmployerJobApplicationsPage(c *fiber.Ctx) error {
	jobID := c.Params("id")
	return c.Render("employer/job-applications", mergePageData(c, fiber.Map{
		"Title":      "Job Applications - Angazia",
		"ActivePage": "jobs",
		"JobID":      jobID,
	}), "layouts/employer")
}

func formatJobSalary(job *models.Job) string {
	if job.SalaryMin == 0 && job.SalaryMax == 0 {
		return ""
	}
	currency := job.SalaryCurrency
	if currency == "" {
		currency = "KES"
	}
	if job.SalaryMin > 0 && job.SalaryMax > 0 {
		return fmt.Sprintf("%s %s - %s", currency, formatNumber(job.SalaryMin), formatNumber(job.SalaryMax))
	}
	if job.SalaryMin > 0 {
		return fmt.Sprintf("%s %s+", currency, formatNumber(job.SalaryMin))
	}
	if job.SalaryMax > 0 {
		return fmt.Sprintf("Up to %s %s", currency, formatNumber(job.SalaryMax))
	}
	return ""
}

func formatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
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
