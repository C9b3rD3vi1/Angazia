package services

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
)

// Client represents a WebSocket client
type Client struct {
	ID         string
	UserID     string
	Conn       *websocket.Conn
	Send       chan []byte
	LastPing   time.Time
	UserAgent  string
	IPAddress  string
}

// WebSocketHub manages all WebSocket connections
type WebSocketHub struct {
	// Registered clients
	clients map[*Client]bool
	
	// Register requests from clients
	Register chan *Client
	
	// Unregister requests from clients
	Unregister chan *Client
	
	// Broadcast to all clients
	broadcast chan []byte
	
	// User-specific messages
	userMessages map[string]chan []byte
	
	// Mutex for protecting maps
	mu sync.RWMutex
}

var (
	hub     *WebSocketHub
	hubOnce sync.Once
)

// GetHub returns the singleton WebSocket hub
func GetHub() *WebSocketHub {
	hubOnce.Do(func() {
		hub = &WebSocketHub{
			clients:      make(map[*Client]bool),
			Register:     make(chan *Client),
			Unregister:   make(chan *Client),
			broadcast:    make(chan []byte),
			userMessages: make(map[string]chan []byte),
		}
		go hub.run()
	})
	return hub
}

// run starts the WebSocket hub
func (h *WebSocketHub) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			
			// Create user-specific message channel if not exists
			if _, ok := h.userMessages[client.UserID]; !ok {
				h.userMessages[client.UserID] = make(chan []byte, 100)
				go h.userMessageHandler(client.UserID)
			}
			h.mu.Unlock()
			
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
			
		case <-ticker.C:
			h.pingClients()
		}
	}
}

// userMessageHandler handles user-specific messages
func (h *WebSocketHub) userMessageHandler(userID string) {
	for message := range h.userMessages[userID] {
		h.mu.RLock()
		for client := range h.clients {
			if client.UserID == userID {
				select {
				case client.Send <- message:
				default:
				}
			}
		}
		h.mu.RUnlock()
	}
}

// SendToUser sends a message to a specific user
func (h *WebSocketHub) SendToUser(userID string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	h.mu.RLock()
	ch, ok := h.userMessages[userID]
	h.mu.RUnlock()
	
	if ok {
		select {
		case ch <- data:
		default:
			// Channel full, create new one
			go func() {
				h.mu.Lock()
				h.userMessages[userID] = make(chan []byte, 100)
				h.mu.Unlock()
				h.userMessages[userID] <- data
			}()
		}
	}
	
	return nil
}

// pingClients sends ping messages to all clients
func (h *WebSocketHub) pingClients() {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for client := range h.clients {
		if client.Conn == nil {
			continue
		}
		
		if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			go func(c *Client) {
				h.Unregister <- c
			}(client)
		}
	}
}

// GetConnectedUserCount returns the number of connected users
func (h *WebSocketHub) GetConnectedUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}