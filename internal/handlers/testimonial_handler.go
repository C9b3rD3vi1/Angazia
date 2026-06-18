package handlers

import (
	"strconv"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
	"github.com/gofiber/fiber/v2"
)

type TestimonialHandler struct {
	testimonialService services.TestimonialService
}

func NewTestimonialHandler(testimonialService services.TestimonialService) *TestimonialHandler {
	return &TestimonialHandler{testimonialService: testimonialService}
}

// ── User endpoints (employee/employer) ──

func (h *TestimonialHandler) Create(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "unauthorized")
	}
	role, _ := c.Locals("user_role").(string)
	if role != "employee" && role != "employer" {
		role = "employee"
	}

	var req models.CreateTestimonialRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if req.Content == "" {
		return utils.BadRequest(c, "content is required")
	}

	t, err := h.testimonialService.Create(c.Context(), userID, &req, role)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Testimonial submitted for review", t)
}

func (h *TestimonialHandler) ListMy(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "unauthorized")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	result, err := h.testimonialService.ListMyTestimonials(c.Context(), userID, page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.Success(c, result)
}

func (h *TestimonialHandler) Update(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "unauthorized")
	}
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "id is required")
	}

	var req models.UpdateTestimonialRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	t, err := h.testimonialService.Update(c.Context(), id, userID, &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Testimonial updated", t)
}

func (h *TestimonialHandler) DeleteMy(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "unauthorized")
	}
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "id is required")
	}

	if err := h.testimonialService.DeleteOwn(c.Context(), id, userID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Testimonial deleted", nil)
}

// ── Public endpoint ──

func (h *TestimonialHandler) ListApproved(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	testimonials, err := h.testimonialService.ListApproved(c.Context(), page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.Success(c, fiber.Map{
		"testimonials": testimonials,
	})
}

// ── Admin endpoints ──

func (h *TestimonialHandler) AdminList(c *fiber.Ctx) error {
	params := &models.ListTestimonialsParams{
		Page:   parseInt(c.Query("page", "1")),
		Limit:  parseInt(c.Query("limit", "20")),
		Status: c.Query("status"),
		Role:   c.Query("role"),
		Search: c.Query("search"),
	}

	result, err := h.testimonialService.List(c.Context(), params)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.Success(c, result)
}

func (h *TestimonialHandler) AdminApprove(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "id is required")
	}
	if err := h.testimonialService.Approve(c.Context(), id); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Testimonial approved", nil)
}

func (h *TestimonialHandler) AdminReject(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "id is required")
	}
	if err := h.testimonialService.Reject(c.Context(), id); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Testimonial rejected", nil)
}

func (h *TestimonialHandler) AdminToggleFeatured(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "id is required")
	}
	if err := h.testimonialService.ToggleFeatured(c.Context(), id); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Featured status toggled", nil)
}

func (h *TestimonialHandler) AdminDelete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "id is required")
	}
	if err := h.testimonialService.Delete(c.Context(), id); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Testimonial deleted", nil)
}

func parseInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 1
	}
	return n
}
