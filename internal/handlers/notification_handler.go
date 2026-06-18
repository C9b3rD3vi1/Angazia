package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type NotificationHandler struct {
	notificationService services.NotificationService
}

func NewNotificationHandler(notificationService services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotifications returns paginated notifications for the user
func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "User not authenticated")
	}

	params := &models.NotificationListParams{
		Page:       c.QueryInt("page", 1),
		Limit:      c.QueryInt("limit", 20),
		Type:       c.Query("type"),
		Search:     c.Query("search"),
		Sort:       c.Query("sort"),
		UnreadOnly: c.Query("unread_only") == "true",
	}

	switch params.Sort {
	case "oldest":
		params.Sort = "created_at ASC"
	case "newest", "":
		params.Sort = "created_at DESC"
	default:
		params.Sort = "created_at DESC"
	}

	notifications, err := h.notificationService.GetNotifications(c.Context(), userID, params)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, notifications)
}

// GetUnreadNotifications returns unread notifications
func (h *NotificationHandler) GetUnreadNotifications(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "User not authenticated")
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	notifications, err := h.notificationService.GetUnreadNotifications(c.Context(), userID, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, notifications)
}

// GetNotificationCounts returns unread notification counts
func (h *NotificationHandler) GetNotificationCounts(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "User not authenticated")
	}

	counts, err := h.notificationService.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, counts)
}

// GetNotification returns a single notification
func (h *NotificationHandler) GetNotification(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Notification ID is required")
	}

	notification, err := h.notificationService.GetNotification(c.Context(), id, userID)
	if err != nil {
		return utils.NotFound(c, err.Error())
	}

	return utils.Success(c, notification)
}

// MarkAsRead marks a notification as read
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Notification ID is required")
	}

	if err := h.notificationService.MarkAsRead(c.Context(), id, userID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Notification marked as read", nil)
}

// MarkAllAsRead marks all notifications as read
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "User not authenticated")
	}

	if err := h.notificationService.MarkAllAsRead(c.Context(), userID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "All notifications marked as read", nil)
}

// MarkMultipleAsRead marks multiple notifications as read
func (h *NotificationHandler) MarkMultipleAsRead(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req models.MarkReadRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if req.MarkAll {
		if err := h.notificationService.MarkAllAsRead(c.Context(), userID); err != nil {
			return utils.InternalServerError(c, err.Error())
		}
	} else if len(req.NotificationIDs) > 0 {
		if err := h.notificationService.MarkMultipleAsRead(c.Context(), req.NotificationIDs, userID); err != nil {
			return utils.InternalServerError(c, err.Error())
		}
	}

	return utils.SuccessWithMessage(c, "Notifications marked as read", nil)
}

// Archive archives a notification
func (h *NotificationHandler) Archive(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Notification ID is required")
	}

	if err := h.notificationService.Archive(c.Context(), id, userID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Notification archived", nil)
}

// Delete deletes a notification
func (h *NotificationHandler) Delete(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Notification ID is required")
	}

	if err := h.notificationService.Delete(c.Context(), id, userID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Notification deleted", nil)
}

// DeleteAll deletes all notifications
func (h *NotificationHandler) DeleteAll(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "User not authenticated")
	}

	if err := h.notificationService.DeleteAll(c.Context(), userID); err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "All notifications deleted", nil)
}

// GetPreferences returns notification preferences
func (h *NotificationHandler) GetPreferences(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "User not authenticated")
	}

	prefs, err := h.notificationService.GetPreferences(c.Context(), userID)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, prefs)
}

// UpdatePreferences updates notification preferences
func (h *NotificationHandler) UpdatePreferences(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req services.UpdatePreferencesRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	prefs, err := h.notificationService.UpdatePreferences(c.Context(), userID, &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Preferences updated successfully", prefs)
}
