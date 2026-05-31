package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupNotificationRoutes(router fiber.Router, notificationHandler *handlers.NotificationHandler) {
	// All routes require authentication
	protected := router.Group("/notifications", middleware.AuthMiddleware())
	
	// Get notifications
	protected.Get("/", notificationHandler.GetNotifications)
	protected.Get("/unread", notificationHandler.GetUnreadNotifications)
	protected.Get("/counts", notificationHandler.GetNotificationCounts)
	protected.Get("/:id", notificationHandler.GetNotification)
	
	// Notification actions
	protected.Post("/:id/read", notificationHandler.MarkAsRead)
	protected.Post("/read-all", notificationHandler.MarkAllAsRead)
	protected.Post("/read-multiple", notificationHandler.MarkMultipleAsRead)
	protected.Post("/:id/archive", notificationHandler.Archive)
	protected.Delete("/:id", notificationHandler.Delete)
	protected.Delete("/", notificationHandler.DeleteAll)
	
	// Preferences
	protected.Get("/preferences", notificationHandler.GetPreferences)
	protected.Put("/preferences", notificationHandler.UpdatePreferences)
}