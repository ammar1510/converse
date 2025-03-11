package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/ammar1510/converse/internal/auth"
)

// Message types
const (
	MessageTypeAuth    = "auth"
	MessageTypeMessage = "message"
	MessageTypeTyping  = "typing"
)

// Client represents a connected websocket client
type Client struct {
	ID     uuid.UUID
	Socket *websocket.Conn
	Send   chan []byte
}

// Manager maintains the set of active clients
type Manager struct {
	clients    map[uuid.UUID]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.Mutex
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type       string    `json:"type"`
	SenderID   uuid.UUID `json:"sender_id,omitempty"`
	ReceiverID uuid.UUID `json:"receiver_id,omitempty"`
	Content    string    `json:"content,omitempty"`
	IsTyping   bool      `json:"is_typing,omitempty"`
	Token      string    `json:"token,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// NewManager creates a new websocket manager
func NewManager() *Manager {
	return &Manager{
		clients:    make(map[uuid.UUID]*Client),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the websocket manager
func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mutex.Lock()
			m.clients[client.ID] = client
			log.Printf("Client connected: %s", client.ID)
			m.mutex.Unlock()
		case client := <-m.unregister:
			m.mutex.Lock()
			if _, ok := m.clients[client.ID]; ok {
				delete(m.clients, client.ID)
				close(client.Send)
				log.Printf("Client disconnected: %s", client.ID)
			}
			m.mutex.Unlock()
		case message := <-m.broadcast:
			m.mutex.Lock()
			for _, client := range m.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(m.clients, client.ID)
				}
			}
			m.mutex.Unlock()
		}
	}
}

// SendToUser sends a message to a specific user
func (m *Manager) SendToUser(userID uuid.UUID, message []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if client, ok := m.clients[userID]; ok {
		select {
		case client.Send <- message:
			log.Printf("Message sent to user %s", userID)
		default:
			close(client.Send)
			delete(m.clients, userID)
			log.Printf("Failed to send message to user %s, client removed", userID)
		}
	} else {
		log.Printf("User %s not connected", userID)
	}
}

// HandleWebSocket handles websocket requests from clients
func (m *Manager) HandleWebSocket(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		// If userID doesn't exist in context, check if token is provided in WebSocket protocol
		// This is a fallback for clients that can't set Authorization header
		protocols := c.Request.Header.Get("Sec-WebSocket-Protocol")
		if protocols != "" {
			protocolSlice := strings.Split(protocols, ", ")
			if len(protocolSlice) >= 2 && protocolSlice[0] == "jwt" {
				// Extract token from protocol
				tokenString := protocolSlice[1]

				// Validate token
				claims, err := auth.ValidateToken(tokenString)
				if err == nil {
					// Parse user ID string into UUID
					if userUUID, err := uuid.Parse(claims.UserID); err == nil {
						userID = userUUID
						exists = true
					}
				}
			}
		}

		// If still no valid user ID, return unauthorized
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
	}

	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
		Subprotocols: []string{"jwt"},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		ID:     userID.(uuid.UUID),
		Socket: conn,
		Send:   make(chan []byte, 256),
	}

	m.register <- client

	// Start goroutines for reading and writing
	go client.readPump(m)
	go client.writePump()
}

// readPump pumps messages from the websocket connection to the manager
func (c *Client) readPump(m *Manager) {
	defer func() {
		m.unregister <- c
		c.Socket.Close()
	}()

	c.Socket.SetReadLimit(512 * 1024) // 512KB
	c.Socket.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Socket.SetPongHandler(func(string) error {
		c.Socket.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		// Process the message
		var wsMessage WebSocketMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Set sender ID and timestamp
		wsMessage.SenderID = c.ID
		wsMessage.Timestamp = time.Now()

		// Handle different message types
		switch wsMessage.Type {
		case MessageTypeAuth:
			// Handle auth message (for backward compatibility)
			// The primary authentication should happen via the HTTP Authorization header
			// This is just a fallback for clients that send auth after connection
			if wsMessage.Token != "" {
				log.Printf("Received auth message from client %s", c.ID)
				// We don't need to do anything here as authentication is already handled by middleware
			}
		case MessageTypeMessage:
			// Send message to recipient
			if wsMessage.ReceiverID != uuid.Nil {
				messageJSON, _ := json.Marshal(wsMessage)
				m.SendToUser(wsMessage.ReceiverID, messageJSON)
			}
		case MessageTypeTyping:
			// Send typing indicator to recipient
			if wsMessage.ReceiverID != uuid.Nil {
				messageJSON, _ := json.Marshal(wsMessage)
				m.SendToUser(wsMessage.ReceiverID, messageJSON)
			}
		}
	}
}

// writePump pumps messages from the manager to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Socket.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The manager closed the channel
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Socket.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Socket.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
