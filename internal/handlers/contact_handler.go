package handlers

import (
	"strconv"

	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
	"github.com/gofiber/fiber/v2"
)

type ContactHandler struct {
	contactService services.ContactService
}

func NewContactHandler(contactService services.ContactService) *ContactHandler {
	return &ContactHandler{contactService: contactService}
}

func (h *ContactHandler) AdminList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	result, err := h.contactService.List(c.Context(), page, limit, search)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.Success(c, result)
}

func (h *ContactHandler) AdminGet(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "id is required")
	}

	sub, err := h.contactService.GetByID(c.Context(), id)
	if err != nil {
		return utils.NotFound(c, "contact submission not found")
	}
	return utils.Success(c, sub)
}

func (h *ContactHandler) AdminMarkRead(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "id is required")
	}

	if err := h.contactService.MarkAsRead(c.Context(), id); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Marked as read", nil)
}

func (h *ContactHandler) AdminDelete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "id is required")
	}

	if err := h.contactService.Delete(c.Context(), id); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	return utils.SuccessWithMessage(c, "Contact submission deleted", nil)
}
