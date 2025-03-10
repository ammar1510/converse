package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ammar1510/converse/internal/database"
	"github.com/ammar1510/converse/internal/models"
	"github.com/ammar1510/converse/internal/websocket"
)

// Global WebSocket manager instance
var WSManager *websocket.Manager

// MessageHandler handles message-related routes
type MessageHandler struct {
	DB database.DBInterface
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(db database.DBInterface) *MessageHandler {
	return &MessageHandler{DB: db}
}

// SendMessage handles the creation of a new message
func (h *MessageHandler) SendMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// The userID from context is now a UUID object
	senderID := userID.(uuid.UUID)

	// Create the message
	message, err := h.DB.CreateMessage(senderID, req.ReceiverID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Notify the receiver via WebSocket if they're connected
	if WSManager != nil {
		// Create WebSocket message
		wsMessage := websocket.WebSocketMessage{
			Type:       "message",
			SenderID:   senderID,
			ReceiverID: req.ReceiverID,
			Content:    req.Content,
			Timestamp:  message.CreatedAt,
		}

		// Convert to JSON
		messageJSON, err := json.Marshal(wsMessage)
		if err == nil {
			// Send to receiver
			WSManager.SendToUser(req.ReceiverID, messageJSON)
		}
	}

	c.JSON(http.StatusCreated, message)
}

// GetMessages returns all messages for the authenticated user
func (h *MessageHandler) GetMessages(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// The userID from context is now a UUID object
	userUUID := userID.(uuid.UUID)

	messages, err := h.DB.GetMessagesByUser(userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// GetConversation returns all messages between the authenticated user and another user
func (h *MessageHandler) GetConversation(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// The userID from context is now a UUID object
	userUUID := userID.(uuid.UUID)

	otherUserIDStr := c.Param("userID")
	otherUserID, err := uuid.Parse(otherUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	messages, err := h.DB.GetConversation(userUUID, otherUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// MarkMessageAsRead marks a message as read
func (h *MessageHandler) MarkMessageAsRead(c *gin.Context) {
	// Get the user ID from the context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// The userID from context is now a UUID object
	userUUID := userID.(uuid.UUID)

	messageIDStr := c.Param("messageID")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	// Verify the user is the receiver of this message for security
	message, err := h.DB.GetMessageByID(messageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve message"})
		return
	}

	// Check if the authenticated user is the intended recipient
	if message.ReceiverID != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to mark this message as read"})
		return
	}

	// Continue with marking the message as read
	err = h.DB.MarkMessageAsRead(messageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Notify the sender via WebSocket that their message was read
	if WSManager != nil {
		// Create WebSocket message
		wsMessage := websocket.WebSocketMessage{
			Type:       "read_receipt",
			SenderID:   userUUID,
			ReceiverID: message.SenderID,
			Timestamp:  message.CreatedAt,
		}

		// Convert to JSON
		messageJSON, err := json.Marshal(wsMessage)
		if err == nil {
			// Send to original sender
			WSManager.SendToUser(message.SenderID, messageJSON)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message marked as read"})
}
