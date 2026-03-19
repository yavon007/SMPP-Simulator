package handler

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"smpp-simulator/pkg/jwt"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub         *WebSocketHub
	jwtSecret   string
	allowOrigins []string
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *WebSocketHub, jwtSecret string, allowOrigins []string) *WebSocketHandler {
	return &WebSocketHandler{
		hub:          hub,
		jwtSecret:    jwtSecret,
		allowOrigins: allowOrigins,
	}
}

// getUpgrader returns a websocket upgrader with origin check
func (h *WebSocketHandler) getUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// If allowOrigins contains "*", allow all
			for _, origin := range h.allowOrigins {
				if origin == "*" {
					return true
				}
			}
			// Check if origin is in allowed list
			origin := r.Header.Get("Origin")
			if origin == "" {
				// Allow connections without Origin header (e.g., from same origin)
				return true
			}
			for _, allowed := range h.allowOrigins {
				if strings.EqualFold(origin, allowed) {
					return true
				}
				// Also allow if origin matches allowed pattern (for development)
				if strings.HasSuffix(allowed, "*") {
					prefix := strings.TrimSuffix(allowed, "*")
					if strings.HasPrefix(origin, prefix) {
						return true
					}
				}
			}
			log.Printf("WebSocket connection rejected from origin: %s", origin)
			return false
		},
	}
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	conn *websocket.Conn
	send chan []byte
}

// WebSocketHub manages WebSocket connections
type WebSocketHub struct {
	clients    map[*WebSocketClient]bool
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	broadcast  chan []byte
	mu         sync.RWMutex
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*WebSocketClient]bool),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		broadcast:  make(chan []byte, 256),
	}
}

// Run starts the hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected. Total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket client disconnected. Total: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *WebSocketHub) Broadcast(message []byte) {
	h.broadcast <- message
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *WebSocketClient) writePump() {
	defer c.conn.Close()
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *WebSocketClient) readPump(h *WebSocketHub) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// Handle ping/pong heartbeat
		if string(message) == `{"type":"ping"}` {
			c.send <- []byte(`{"type":"pong"}`)
		}
	}
}

// Handle handles WebSocket upgrade
func (h *WebSocketHandler) Handle(c *gin.Context) {
	// Authenticate via query parameter token (optional)
	token := c.Query("token")
	if token == "" {
		// Also check Authorization header
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// Validate JWT token if provided
	var username string
	if token != "" {
		claims, err := jwt.ValidateToken(token, h.jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		username = claims.Username
	}

	// Proceed with WebSocket upgrade (with or without authentication)
	upgrader := h.getUpgrader()
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	if username != "" {
		log.Printf("WebSocket client authenticated: %s", username)
	} else {
		log.Printf("WebSocket client connected (anonymous)")
	}

	client := &WebSocketClient{
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.hub.register <- client

	go client.writePump()
	go client.readPump(h.hub)
}
