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
		return utils.Unauthorized(c, "Missing token")
	}

	claims, err := utils.ValidateJWT(token)
	if err != nil {
		return utils.Unauthorized(c, "Invalid token")
	}

	// Capture safe values before upgrade
	userAgent := string(c.Request().Header.Peek("User-Agent"))
	ip := c.IP()

	err = upgrader.Upgrade(c.Context(), func(conn *websocket.Conn) {
		// Defer recover to prevent panic
		defer func() {
			if r := recover(); r != nil {
				log.Printf("WebSocket panic recovered: %v", r)
				conn.Close()
			}
		}()

		client := &services.Client{
			ID:        uuid.New().String(),
			UserID:    claims.UserID,
			Conn:      conn,
			Send:      make(chan []byte, 256),
			LastPing:  time.Now(),
			UserAgent: userAgent,
			IPAddress: ip,
		}

		h.hub.Register <- client

		go h.writePump(client)
		h.readPump(client)
	})

	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return utils.InternalServerError(c, "WebSocket upgrade failed")
	}

	return nil
}

func (h *WebSocketHandler) writePump(client *services.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		if r := recover(); r != nil {
			log.Printf("Write pump panic recovered: %v", r)
		}
		if client.Conn != nil {
			client.Conn.Close()
		}
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				if client.Conn != nil {
					client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				}
				return
			}

			// Check if connection is still alive
			if client.Conn == nil {
				return
			}

			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Write message error: %v", err)
				return
			}

		case <-ticker.C:
			// Check if connection is still alive
			if client.Conn == nil {
				return
			}

			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping error: %v", err)
				return
			}
		}
	}
}

func (h *WebSocketHandler) readPump(client *services.Client) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Read pump panic recovered: %v", r)
		}
		h.hub.Unregister <- client
		if client.Conn != nil {
			client.Conn.Close()
		}
	}()

	if client.Conn == nil {
		return
	}

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