package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Message types
const (
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
	remoteAddr := c.Request.RemoteAddr
	log.Printf("[WebSocket-Handler] Connection attempt from %s", remoteAddr)

	// Get user ID from context (set by auth middleware or route handler)
	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("[WebSocket-Handler] No userID in context, rejecting connection from %s", remoteAddr)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Validate that userID is actually a uuid.UUID type
	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		log.Printf("[WebSocket-Handler] UserID in context is not a valid UUID, rejecting connection from %s", remoteAddr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user identification"})
		return
	}

	log.Printf("[WebSocket-Handler] User authenticated: %s (IP: %s)", userUUID, remoteAddr)

	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			log.Printf("[WebSocket-Handler] WebSocket origin: %s (IP: %s)", origin, remoteAddr)
			// TODO: In production, implement proper origin checking
			return true // Allow all origins in development
		},
	}

	log.Printf("[WebSocket-Handler] Upgrading connection to WebSocket for %s", remoteAddr)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WebSocket-Handler] Failed to upgrade connection for %s: %v", remoteAddr, err)
		return
	}
	log.Printf("[WebSocket-Handler] Connection upgraded successfully for %s", remoteAddr)

	client := &Client{
		ID:     userUUID,
		Socket: conn,
		Send:   make(chan []byte, 256),
	}
	log.Printf("[WebSocket-Handler] Created client with ID: %s (IP: %s)", client.ID, remoteAddr)

	m.register <- client
	log.Printf("[WebSocket-Handler] Registered client %s with manager", client.ID)

	// Start goroutines for reading and writing
	go client.readPump(m)
	go client.writePump()
	log.Printf("[WebSocket-Handler] Started client read/write pumps for %s", client.ID)
}

// readPump pumps messages from the websocket connection to the manager
func (c *Client) readPump(m *Manager) {
	defer func() {
		log.Printf("[WebSocket-ReadPump] Client %s disconnecting, unregistering from manager", c.ID)
		m.unregister <- c
		c.Socket.Close()
	}()

	c.Socket.SetReadLimit(64 * 1024) // Reduced from 512KB to 64KB for most chat messages
	c.Socket.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Socket.SetPongHandler(func(string) error {
		c.Socket.SetReadDeadline(time.Now().Add(60 * time.Second))
		log.Printf("[WebSocket-ReadPump] Received pong from client %s, extending deadline", c.ID)
		return nil
	})

	log.Printf("[WebSocket-ReadPump] Started read pump for client %s", c.ID)

	// Implement a simple rate limiting mechanism
	messageCount := 0
	lastResetTime := time.Now()
	const maxMessagesPerMinute = 60 // Adjust as needed

	for {
		// Rate limiting check
		if messageCount >= maxMessagesPerMinute {
			if time.Since(lastResetTime) < time.Minute {
				log.Printf("[WebSocket-ReadPump] Rate limit exceeded for client %s", c.ID)
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
				log.Printf("[WebSocket-ReadPump] Error reading message from client %s: %v", c.ID, err)
			} else {
				log.Printf("[WebSocket-ReadPump] Client %s closed connection: %v", c.ID, err)
			}
			break
		}

		messageCount++ // Increment message counter for rate limiting

		// Process the message
		var wsMessage WebSocketMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			log.Printf("[WebSocket-ReadPump] Error unmarshaling message from client %s: %v", c.ID, err)

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

		log.Printf("[WebSocket-ReadPump] Received message type '%s' from client %s", wsMessage.Type, c.ID)

		// Handle different message types
		switch wsMessage.Type {
		case MessageTypeMessage:
			// Validate message
			if wsMessage.Content == "" {
				log.Printf("[WebSocket-ReadPump] Empty message content from client %s", c.ID)
				continue
			}

			// Send message to recipient
			if wsMessage.ReceiverID != uuid.Nil {
				log.Printf("[WebSocket-ReadPump] Forwarding message from client %s to recipient %s", c.ID, wsMessage.ReceiverID)
				messageJSON, _ := json.Marshal(wsMessage)
				m.SendToUser(wsMessage.ReceiverID, messageJSON)
			} else {
				log.Printf("[WebSocket-ReadPump] Invalid receiver ID in message from client %s", c.ID)

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
				log.Printf("[WebSocket-ReadPump] Forwarding typing indicator from client %s to recipient %s (typing: %v)",
					c.ID, wsMessage.ReceiverID, wsMessage.IsTyping)
				messageJSON, _ := json.Marshal(wsMessage)
				m.SendToUser(wsMessage.ReceiverID, messageJSON)
			} else {
				log.Printf("[WebSocket-ReadPump] Invalid receiver ID in typing indicator from client %s", c.ID)
			}
		default:
			log.Printf("[WebSocket-ReadPump] Unknown message type '%s' from client %s", wsMessage.Type, c.ID)

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
