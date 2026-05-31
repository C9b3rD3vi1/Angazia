package routes

import (
	"github.com/gofiber/fiber/v2"
	
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
)

func SetupWebSocketRoutes(app *fiber.App, websocketHandler *handlers.WebSocketHandler) {
	// WebSocket endpoint
	app.Get("/ws", websocketHandler.UpgradeWebSocket)
}