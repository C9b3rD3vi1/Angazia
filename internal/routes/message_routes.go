package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupMessageRoutes(api fiber.Router, messageHandler *handlers.MessageHandler) {
	messages := api.Group("/messages", middleware.AuthMiddleware())

	messages.Get("/", messageHandler.ListConversations)
	messages.Post("/", messageHandler.CreateConversation)
	messages.Get("/unread-count", messageHandler.GetUnreadCount)
	messages.Get("/:id", messageHandler.GetConversation)
	messages.Post("/:id/messages", messageHandler.SendMessage)
	messages.Get("/:id/messages", messageHandler.ListMessages)
	messages.Post("/:id/read", messageHandler.MarkConversationRead)
}
