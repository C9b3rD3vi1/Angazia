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
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	
	notifications, err := h.notificationService.GetNotifications(c.Context(), userID.(string), page, limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, notifications)
}

// GetUnreadNotifications returns unread notifications
func (h *NotificationHandler) GetUnreadNotifications(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	
	notifications, err := h.notificationService.GetUnreadNotifications(c.Context(), userID.(string), limit)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, notifications)
}

// GetNotificationCounts returns unread notification counts
func (h *NotificationHandler) GetNotificationCounts(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	counts, err := h.notificationService.GetUnreadCount(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, counts)
}

// GetNotification returns a single notification
func (h *NotificationHandler) GetNotification(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Notification ID is required")
	}
	
	notification, err := h.notificationService.GetNotification(c.Context(), id, userID.(string))
	if err != nil {
		return utils.NotFound(c, err.Error())
	}
	
	return utils.Success(c, notification)
}

// MarkAsRead marks a notification as read
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Notification ID is required")
	}
	
	if err := h.notificationService.MarkAsRead(c.Context(), id, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Notification marked as read", nil)
}

// MarkAllAsRead marks all notifications as read
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	if err := h.notificationService.MarkAllAsRead(c.Context(), userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "All notifications marked as read", nil)
}

// MarkMultipleAsRead marks multiple notifications as read
func (h *NotificationHandler) MarkMultipleAsRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	var req models.MarkReadRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if req.MarkAll {
		if err := h.notificationService.MarkAllAsRead(c.Context(), userID.(string)); err != nil {
			return utils.InternalServerError(c, err.Error())
		}
	} else if len(req.NotificationIDs) > 0 {
		if err := h.notificationService.MarkMultipleAsRead(c.Context(), req.NotificationIDs, userID.(string)); err != nil {
			return utils.InternalServerError(c, err.Error())
		}
	}
	
	return utils.SuccessWithMessage(c, "Notifications marked as read", nil)
}

// Archive archives a notification
func (h *NotificationHandler) Archive(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Notification ID is required")
	}
	
	if err := h.notificationService.Archive(c.Context(), id, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Notification archived", nil)
}

// Delete deletes a notification
func (h *NotificationHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	id := c.Params("id")
	if id == "" {
		return utils.BadRequest(c, "Notification ID is required")
	}
	
	if err := h.notificationService.Delete(c.Context(), id, userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Notification deleted", nil)
}

// DeleteAll deletes all notifications
func (h *NotificationHandler) DeleteAll(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	if err := h.notificationService.DeleteAll(c.Context(), userID.(string)); err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "All notifications deleted", nil)
}

// GetPreferences returns notification preferences
func (h *NotificationHandler) GetPreferences(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	prefs, err := h.notificationService.GetPreferences(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, prefs)
}

// UpdatePreferences updates notification preferences
func (h *NotificationHandler) UpdatePreferences(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	var req services.UpdatePreferencesRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	prefs, err := h.notificationService.UpdatePreferences(c.Context(), userID.(string), &req)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "Preferences updated successfully", prefs)
}
