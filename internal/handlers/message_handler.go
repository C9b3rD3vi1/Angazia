package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type MessageHandler struct {
	messageService services.MessageService
}

func NewMessageHandler(messageService services.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

func (h *MessageHandler) ListConversations(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	resp, err := h.messageService.ListConversations(c.Context(), userID.(string), page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, resp)
}

func (h *MessageHandler) GetConversation(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Conversation ID is required")
	}

	conv, err := h.messageService.GetConversation(c.Context(), id, userID.(string))
	if err != nil {
		return utils.NotFound(c, err.Error())
	}

	return utils.Success(c, conv)
}

func (h *MessageHandler) CreateConversation(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req struct {
		RecipientID string `json:"recipient_id"`
		Subject     string `json:"subject"`
		JobID       string `json:"job_id"`
		Content     string `json:"content"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	if req.RecipientID == "" {
		return utils.BadRequest(c, "recipient_id is required")
	}

	conv, err := h.messageService.CreateConversation(c.Context(), userID.(string), req.RecipientID, req.Subject, req.JobID)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	if req.Content != "" {
		_, err = h.messageService.SendMessage(c.Context(), userID.(string), conv.ID, req.Content)
		if err != nil {
			return utils.InternalServerError(c, err.Error())
		}
	}

	return utils.SuccessWithMessage(c, "Conversation created", conv)
}

func (h *MessageHandler) SendMessage(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	conversationID := c.Params("id")
	if conversationID == "" {
		return utils.BadRequest(c, "Conversation ID is required")
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	if req.Content == "" {
		return utils.BadRequest(c, "content is required")
	}

	msg, err := h.messageService.SendMessage(c.Context(), userID.(string), conversationID, req.Content)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Message sent", msg)
}

func (h *MessageHandler) ListMessages(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	conversationID := c.Params("id")
	if conversationID == "" {
		return utils.BadRequest(c, "Conversation ID is required")
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)

	resp, err := h.messageService.ListMessages(c.Context(), conversationID, userID.(string), page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, resp)
}

func (h *MessageHandler) MarkConversationRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	conversationID := c.Params("id")
	if conversationID == "" {
		return utils.BadRequest(c, "Conversation ID is required")
	}

	if err := h.messageService.MarkConversationRead(c.Context(), conversationID, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Conversation marked as read", nil)
}

func (h *MessageHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	count, err := h.messageService.GetUnreadCount(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, map[string]int{"unread_count": count})
}

func (h *MessageHandler) MessagesPage(c *fiber.Ctx) error {
	role, _ := c.Locals("user_role").(string)
	layout := "layouts/employee"
	tmpl := "employee/messages"
	if role == "employer" {
		layout = "layouts/employer"
		tmpl = "employer/messages"
	}
	return c.Render(tmpl, mergePageData(c, fiber.Map{
		"Title":      "Messages",
		"ActivePage": "messages",
	}), layout)
}
