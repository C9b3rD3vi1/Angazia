package handlers

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type ReviewHandler struct {
	companyService services.CompanyService
	validator      *validator.Validate
}

func NewReviewHandler(companyService services.CompanyService) *ReviewHandler {
	return &ReviewHandler{
		companyService: companyService,
		validator:      validator.New(),
	}
}

// SubmitReview submits a review for a company
// @Summary Submit a company review
// @Description Submit a review and rating for a company
// @Tags Reviews
// @Security BearerAuth
// @Param id path string true "Company ID"
// @Param request body services.SubmitReviewRequest true "Review details"
// @Success 201 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Router /companies/{id}/reviews [post]
func (h *ReviewHandler) SubmitReview(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	companyID := c.Params("id")
	if companyID == "" {
		return utils.BadRequest(c, "Company ID is required")
	}

	var req services.SubmitReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	review, err := h.companyService.SubmitReview(c.Context(), companyID, userID.(string), &req)
	if err != nil {
		if err.Error() == "you have already reviewed this company" {
			return utils.Conflict(c, err.Error())
		}
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessCreated(c, "Review submitted successfully", review)
}

// GetCompanyReviews retrieves all reviews for a company
// @Summary Get company reviews
// @Description Get paginated list of reviews for a company
// @Tags Reviews
// @Param id path string true "Company ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param sort query string false "Sort by (newest, highest, lowest)" default(newest)
// @Success 200 {object} APIResponse
// @Router /companies/{id}/reviews [get]
func (h *ReviewHandler) GetCompanyReviews(c *fiber.Ctx) error {
	companyID := c.Params("id")
	if companyID == "" {
		return utils.BadRequest(c, "Company ID is required")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	sort := c.Query("sort", "newest")

	// Validate sort parameter
	if sort != "newest" && sort != "highest" && sort != "lowest" {
		sort = "newest"
	}

	reviews, err := h.companyService.GetCompanyReviews(c.Context(), companyID, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	// Apply sorting after fetching (or implement in repository)
	if sort == "highest" {
		for i := 0; i < len(reviews.Reviews)-1; i++ {
			for j := i + 1; j < len(reviews.Reviews); j++ {
				if reviews.Reviews[j].Rating > reviews.Reviews[i].Rating {
					reviews.Reviews[i], reviews.Reviews[j] = reviews.Reviews[j], reviews.Reviews[i]
				}
			}
		}
	} else if sort == "lowest" {
		for i := 0; i < len(reviews.Reviews)-1; i++ {
			for j := i + 1; j < len(reviews.Reviews); j++ {
				if reviews.Reviews[j].Rating < reviews.Reviews[i].Rating {
					reviews.Reviews[i], reviews.Reviews[j] = reviews.Reviews[j], reviews.Reviews[i]
				}
			}
		}
	}

	return utils.Success(c, reviews)
}

// GetReviewStats retrieves review statistics for a company
// @Summary Get company review statistics
// @Description Get aggregated review statistics including average rating and distribution
// @Tags Reviews
// @Param id path string true "Company ID"
// @Success 200 {object} APIResponse
// @Router /companies/{id}/reviews/stats [get]
func (h *ReviewHandler) GetReviewStats(c *fiber.Ctx) error {
	companyID := c.Params("id")
	if companyID == "" {
		return utils.BadRequest(c, "Company ID is required")
	}

	stats, err := h.companyService.GetReviewStats(c.Context(), companyID)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, stats)
}

// MarkReviewHelpful marks a review as helpful
// @Summary Mark review as helpful
// @Description Increment the helpful count for a review
// @Tags Reviews
// @Security BearerAuth
// @Param id path string true "Review ID"
// @Success 200 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Router /reviews/{id}/helpful [post]
func (h *ReviewHandler) MarkReviewHelpful(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	reviewID := c.Params("id")
	if reviewID == "" {
		return utils.BadRequest(c, "Review ID is required")
	}

	if err := h.companyService.MarkReviewHelpful(c.Context(), reviewID, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Review marked as helpful", nil)
}

// DeleteReview deletes a review (admin or review owner only)
// @Summary Delete a review
// @Description Delete a review (admin or review owner only)
// @Tags Reviews
// @Security BearerAuth
// @Param id path string true "Review ID"
// @Success 200 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 403 {object} APIResponse
// @Router /reviews/{id} [delete]
func (h *ReviewHandler) DeleteReview(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	userRole := c.Locals("user_role")
	if userRole == nil {
		return utils.Unauthorized(c, "User role not found")
	}

	reviewID := c.Params("id")
	if reviewID == "" {
		return utils.BadRequest(c, "Review ID is required")
	}

	// Check if user is admin or review owner
	// For now, only admin can delete
	if userRole != "admin" {
		return utils.Forbidden(c, "Only administrators can delete reviews")
	}

	// Delete review (implement in repository)
	// For now, return not implemented
	return utils.Error(c, fiber.StatusNotImplemented, "Review deletion will be implemented in admin panel")
}

// GetMyReviews retrieves reviews written by the authenticated user
// @Summary Get my reviews
// @Description Get all reviews written by the authenticated user
// @Tags Reviews
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} APIResponse
// @Router /user/reviews [get]
func (h *ReviewHandler) GetMyReviews(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	// Get reviews by user (implement in repository)
	// For now, return empty
	return utils.Success(c, fiber.Map{
		"reviews": []interface{}{},
		"total":   0,
		"page":    page,
		"limit":   limit,
	})
}
