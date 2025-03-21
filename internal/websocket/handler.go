package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/ammar1510/converse/internal/logger"
)

// Message types
const (
	MessageTypeMessage = "message"
	MessageTypeTyping  = "typing"
)

var log = logger.New("websocket")

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
			log.Info("Client connected: %s", client.ID)
			m.mutex.Unlock()
		case client := <-m.unregister:
			m.mutex.Lock()
			if _, ok := m.clients[client.ID]; ok {
				delete(m.clients, client.ID)
				close(client.Send)
				log.Info("Client disconnected: %s", client.ID)
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
			log.Debug("Message sent to user %s", userID)
		default:
			close(client.Send)
			delete(m.clients, client.ID)
			log.Warn("Failed to send message to user %s, removing client", userID)
		}
	} else {
		log.Debug("User %s not connected", userID)
	}
}

// HandleWebSocket handles websocket requests from clients
func (m *Manager) HandleWebSocket(c *gin.Context) {
	// Get user ID from context (set by auth middleware or route handler)
	userID, exists := c.Get("userID")
	if !exists {
		log.Warn("No userID in context, rejecting connection from %s", c.Request.RemoteAddr)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Validate that userID is actually a uuid.UUID type
	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		log.Error("Invalid UUID in context from %s", c.Request.RemoteAddr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user identification"})
		return
	}

	log.Debug("User authenticated: %s (IP: %s)", userUUID, c.Request.RemoteAddr)

	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			log.Debug("WebSocket origin: %s", origin)
			// TODO: In production, implement proper origin checking
			return true // Allow all origins in development
		},
	}

	log.Debug("Upgrading connection to WebSocket for %s", c.Request.RemoteAddr)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		ID:     userUUID,
		Socket: conn,
		Send:   make(chan []byte, 256),
	}

	m.register <- client
	log.Debug("Registered client %s with manager", client.ID)

	// Start goroutines for reading and writing
	go client.readPump(m)
	go client.writePump()
	log.Info("Client %s connected and ready", client.ID)
}

// readPump pumps messages from the websocket connection to the manager
func (c *Client) readPump(m *Manager) {
	defer func() {
		log.Debug("Client %s disconnecting, unregistering from manager", c.ID)
		m.unregister <- c
		c.Socket.Close()
	}()

	c.Socket.SetReadLimit(64 * 1024) // Reduced from 512KB to 64KB for most chat messages
	c.Socket.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Socket.SetPongHandler(func(string) error {
		c.Socket.SetReadDeadline(time.Now().Add(60 * time.Second))
		log.Debug("Received pong from client %s, extending deadline", c.ID)
		return nil
	})

	log.Debug("Started read pump for client %s", c.ID)

	// Implement a simple rate limiting mechanism
	messageCount := 0
	lastResetTime := time.Now()
	const maxMessagesPerMinute = 60 // Adjust as needed

	for {
		// Rate limiting check
		if messageCount >= maxMessagesPerMinute {
			if time.Since(lastResetTime) < time.Minute {
				log.Warn("Rate limit exceeded for client %s", c.ID)
				time.Sleep(time.Second) // Sleep briefly before checking again
				continue
			}
			// Reset counter after a minute
			messageCount = 0
			lastResetTime = time.Now()
		}

		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("Error reading from client %s: %v", c.ID, err)
			} else {
				log.Info("Client %s closed connection: %v", c.ID, err)
			}
			break
		}

		messageCount++ // Increment message counter for rate limiting

		// Process the message
		var wsMessage WebSocketMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			log.Error("Error unmarshaling message: %v", err)

			// Send error message to client
			errMsg := WebSocketMessage{
				Type:      "error",
				Content:   "Invalid message format",
				Timestamp: time.Now(),
			}
			errJSON, _ := json.Marshal(errMsg)
			c.Send <- errJSON

			continue
		}

		// Set sender ID and timestamp
		wsMessage.SenderID = c.ID
		wsMessage.Timestamp = time.Now()

		log.Debug("Received message type '%s' from client %s", wsMessage.Type, c.ID)

		// Handle different message types
		switch wsMessage.Type {
		case MessageTypeMessage:
			// Validate message
			if wsMessage.Content == "" {
				log.Debug("Empty message content from client %s", c.ID)
				continue
			}

			// Send message to recipient
			if wsMessage.ReceiverID != uuid.Nil {
				log.Debug("Forwarding message from client %s to recipient %s", c.ID, wsMessage.ReceiverID)
				messageJSON, _ := json.Marshal(wsMessage)
				m.SendToUser(wsMessage.ReceiverID, messageJSON)
			} else {
				log.Warn("Invalid receiver ID from client %s", c.ID)

				// Send error message to client
				errMsg := WebSocketMessage{
					Type:      "error",
					Content:   "Invalid receiver ID",
					Timestamp: time.Now(),
				}
				errJSON, _ := json.Marshal(errMsg)
				c.Send <- errJSON
			}
		case MessageTypeTyping:
			// Send typing indicator to recipient
			if wsMessage.ReceiverID != uuid.Nil {
				log.Debug("Forwarding typing indicator from client %s to recipient %s (typing: %v)",
					c.ID, wsMessage.ReceiverID, wsMessage.IsTyping)
				messageJSON, _ := json.Marshal(wsMessage)
				m.SendToUser(wsMessage.ReceiverID, messageJSON)
			} else {
				log.Debug("Invalid receiver ID in typing indicator from client %s", c.ID)
			}
		default:
			log.Warn("Unknown message type '%s' from client %s", wsMessage.Type, c.ID)

			// Send error message to client
			errMsg := WebSocketMessage{
				Type:      "error",
				Content:   "Unknown message type",
				Timestamp: time.Now(),
			}
			errJSON, _ := json.Marshal(errMsg)
			c.Send <- errJSON
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
