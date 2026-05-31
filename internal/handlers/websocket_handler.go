package handlers

import (
	"log"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	
	"github.com/C9b3rD3vi1/Angazia/internal/services"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
)

var upgrader = websocket.FastHTTPUpgrader{
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
	},
}

type WebSocketHandler struct {
	hub *services.WebSocketHub
}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		hub: services.GetHub(),
	}
}

// UpgradeWebSocket upgrades HTTP connection to WebSocket
func (h *WebSocketHandler) UpgradeWebSocket(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		token = c.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing_token",
		})
	}

	claims, err := utils.ValidateJWT(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid_token",
		})
	}

	err = upgrader.Upgrade(c.Context(), func(conn *websocket.Conn) {
		client := &services.Client{
			ID:        uuid.New().String(),
			UserID:    claims.UserID,
			Conn:      conn,
			Send:      make(chan []byte, 256),
			LastPing:  time.Now(),
			UserAgent: string(c.Request().Header.Peek("User-Agent")),
			IPAddress: c.IP(),
		}

		h.hub.Register <- client

		go h.writePump(client)
		h.readPump(client)
	})
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "websocket_upgrade_failed",
		})
	}

	return nil
}

func (h *WebSocketHandler) writePump(client *services.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WebSocketHandler) readPump(client *services.Client) {
	defer func() {
		h.hub.Unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		client.LastPing = time.Now()
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}
		_ = message
	}
}
