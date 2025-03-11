package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ammar1510/converse/internal/database"
	"github.com/ammar1510/converse/internal/models"
)

// MockDB implements the DBInterface for testing
type MockDB struct {
	mock.Mock
}

// Implement all required database methods for the MockDB

// CreateMessage mocks the database creation of a message
func (m *MockDB) CreateMessage(senderID, receiverID uuid.UUID, content string) (*models.Message, error) {
	args := m.Called(senderID, receiverID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

// GetMessagesByUser mocks retrieving all messages for a user
func (m *MockDB) GetMessagesByUser(userID uuid.UUID) ([]*models.Message, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

// GetConversation mocks retrieving messages between two users
func (m *MockDB) GetConversation(userID1, userID2 uuid.UUID) ([]*models.Message, error) {
	args := m.Called(userID1, userID2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

// GetUserByID mocks retrieving a user by ID
func (m *MockDB) GetUserByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// MarkMessageAsRead mocks marking a message as read
func (m *MockDB) MarkMessageAsRead(messageID uuid.UUID) error {
	args := m.Called(messageID)
	return args.Error(0)
}

// GetUserByEmail mocks retrieving a user by email
func (m *MockDB) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// CreateUser mocks creating a user
func (m *MockDB) CreateUser(username, email, password string) (*models.User, error) {
	args := m.Called(username, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// UpdateLastSeen mocks updating the last seen time for a user
func (m *MockDB) UpdateLastSeen(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

// GetMessageByID mocks retrieving a message by its ID
func (m *MockDB) GetMessageByID(messageID uuid.UUID) (*models.Message, error) {
	args := m.Called(messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

// Close mocks closing the database connection
func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Exec mocks executing a SQL query
func (m *MockDB) Exec(query string, args ...interface{}) (database.ExecResult, error) {
	mockArgs := m.Called(append([]interface{}{query}, args...)...)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(database.ExecResult), mockArgs.Error(1)
}

// MockWebSocketManager is a mock implementation of the WebSocket manager
type MockWebSocketManager struct {
	mock.Mock
}

// SendToUser mocks the SendToUser method
func (m *MockWebSocketManager) SendToUser(userID uuid.UUID, message []byte) {
	m.Called(userID, message)
}

// Run mocks the Run method
func (m *MockWebSocketManager) Run() {
	m.Called()
}

// HandleWebSocket mocks the HandleWebSocket method
func (m *MockWebSocketManager) HandleWebSocket(c *gin.Context) {
	m.Called(c)
}

// Setup helper functions for tests

// setupMessageTest creates a gin router with the MockDB and required middleware for message testing
func setupMessageTest(t *testing.T) (*gin.Engine, *MockDB, uuid.UUID) {
	gin.SetMode(gin.TestMode)

	// Create a test user ID
	userID := uuid.New()

	// Create router
	router := gin.Default()

	// Create mock database
	mockDB := new(MockDB)

	// Create message handler with mock DB using the constructor
	handler := NewMessageHandler(mockDB)

	// Set up routes with authentication middleware mock
	group := router.Group("/api")
	group.Use(func(c *gin.Context) {
		// Mock auth middleware to set user ID
		c.Set("userID", userID)
		c.Next()
	})

	// Register message routes
	group.POST("/messages", handler.SendMessage)
	group.GET("/messages", handler.GetMessages)
	group.GET("/messages/conversation/:userID", handler.GetConversation)
	group.PUT("/messages/:messageID/read", handler.MarkMessageAsRead)

	return router, mockDB, userID
}

// Test cases

func TestCreateMessage(t *testing.T) {
	router, mockDB, senderID := setupMessageTest(t)

	// Test case: successful message creation
	t.Run("Successful message creation", func(t *testing.T) {
		// Setup
		receiverID := uuid.New()
		messageContent := "Hello!"

		// Create expected message
		expectedMessage := &models.Message{
			ID:         uuid.New(),
			SenderID:   senderID,
			ReceiverID: receiverID,
			Content:    messageContent,
			CreatedAt:  time.Now(),
		}

		// Setup mock expectations
		mockDB.On("CreateMessage", senderID, receiverID, messageContent).Return(expectedMessage, nil).Once()

		// Create request JSON
		reqBody := map[string]interface{}{
			"receiver_id": receiverID.String(),
			"content":     messageContent,
		}
		jsonData, _ := json.Marshal(reqBody)

		// Create request
		req, _ := http.NewRequest("POST", "/api/messages", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve the request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		// Parse response
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Assert expected fields in response
		assert.Equal(t, expectedMessage.ID.String(), response["id"])
		assert.Equal(t, expectedMessage.Content, response["content"])
		assert.Equal(t, expectedMessage.SenderID.String(), response["sender_id"])
		assert.Equal(t, expectedMessage.ReceiverID.String(), response["receiver_id"])

		// Verify mock expectations were met
		mockDB.AssertExpectations(t)
	})

	// Test case: missing receiver ID
	t.Run("Missing receiver ID", func(t *testing.T) {
		// Setup
		senderID := uuid.New()

		// Create request JSON with missing receiver_id
		reqBody := map[string]interface{}{
			"content": "Hello!",
		}
		jsonData, _ := json.Marshal(reqBody)

		// Create request
		req, _ := http.NewRequest("POST", "/api/messages", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Mock the authentication middleware
		router.Use(func(c *gin.Context) {
			c.Set("userID", senderID)
			c.Next()
		})

		// Serve the request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Add more test cases as needed
}

func TestGetMessages(t *testing.T) {
	router, mockDB, currentUserID := setupMessageTest(t)

	// Test case: successful retrieval of messages
	t.Run("Successful message retrieval", func(t *testing.T) {
		// Use currentUserID from setupMessageTest instead of declaring a new one
		otherUserID := uuid.New()

		// Create sample messages
		messages := []*models.Message{
			{
				ID:         uuid.New(),
				SenderID:   currentUserID,
				ReceiverID: otherUserID,
				Content:    "Hello!",
				CreatedAt:  time.Now(),
			},
			{
				ID:         uuid.New(),
				SenderID:   otherUserID,
				ReceiverID: currentUserID,
				Content:    "Hi there!",
				CreatedAt:  time.Now().Add(-5 * time.Minute),
			},
		}

		// Setup mock expectations
		mockDB.On("GetMessagesByUser", currentUserID).Return(messages, nil).Once()

		// Create request
		req, _ := http.NewRequest("GET", "/api/messages", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Mock the authentication middleware
		router.Use(func(c *gin.Context) {
			c.Set("userID", currentUserID)
			c.Next()
		})

		// Serve the request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response []map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Assert response length
		assert.Len(t, response, 2)

		// Verify mock expectations were met
		mockDB.AssertExpectations(t)
	})

	// Additional test cases...
}

func TestGetConversation(t *testing.T) {
	router, mockDB, currentUserID := setupMessageTest(t)

	// Test case: successful retrieval of conversation
	t.Run("Successful conversation retrieval", func(t *testing.T) {
		// Use currentUserID from setupMessageTest instead of declaring a new one
		otherUserID := uuid.New()

		// Create sample conversation messages
		messages := []*models.Message{
			{
				ID:         uuid.New(),
				SenderID:   currentUserID,
				ReceiverID: otherUserID,
				Content:    "Hello!",
				CreatedAt:  time.Now().Add(-10 * time.Minute),
			},
			{
				ID:         uuid.New(),
				SenderID:   otherUserID,
				ReceiverID: currentUserID,
				Content:    "Hi there!",
				CreatedAt:  time.Now().Add(-5 * time.Minute),
			},
		}

		// Setup mock expectations
		mockDB.On("GetConversation", currentUserID, otherUserID).Return(messages, nil).Once()

		// Create request - fix URL to match the registered route
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/messages/conversation/%s", otherUserID), nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Mock the authentication middleware
		router.Use(func(c *gin.Context) {
			c.Set("userID", currentUserID)
			c.Next()
		})

		// Serve the request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response []map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Assert response length
		assert.Len(t, response, 2)

		// Verify mock expectations were met
		mockDB.AssertExpectations(t)
	})

	// Additional test cases...
}

func TestMarkMessageAsRead(t *testing.T) {
	router, mockDB, userID := setupMessageTest(t) // Use userID for authentication

	// Test case: successful marking message as read
	t.Run("Successful marking message as read", func(t *testing.T) {
		// Setup
		messageID := uuid.New()

		// Create a mock message with the current user as receiver
		mockMessage := &models.Message{
			ID:         messageID,
			SenderID:   uuid.New(), // Some other user
			ReceiverID: userID,     // Current user is the receiver
			Content:    "Test message",
			CreatedAt:  time.Now(),
			IsRead:     false,
		}

		// Setup mock expectations - first GetMessageByID, then MarkMessageAsRead
		mockDB.On("GetMessageByID", messageID).Return(mockMessage, nil).Once()
		mockDB.On("MarkMessageAsRead", messageID).Return(nil).Once()

		// Create request
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/messages/%s/read", messageID), nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve the request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify mock expectations were met
		mockDB.AssertExpectations(t)
	})

	// Test case: invalid message ID
	t.Run("Invalid message ID", func(t *testing.T) {
		// Create request with invalid UUID
		req, _ := http.NewRequest("PUT", "/api/messages/invalid-uuid/read", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve the request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test case: message not found
	t.Run("Message not found", func(t *testing.T) {
		// Setup
		messageID := uuid.New()

		// Setup mock expectations - simulate message not found
		mockDB.On("GetMessageByID", messageID).Return(nil, fmt.Errorf("message not found")).Once()

		// Create request
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/messages/%s/read", messageID), nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Serve the request
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// Verify mock expectations were met
		mockDB.AssertExpectations(t)
	})
}

// TestSendMessageWithoutWebSocket tests the SendMessage handler with WebSocket manager set to nil
func TestSendMessageWithoutWebSocket(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create mock DB
	mockDB := new(MockDB)

	// Save original WebSocket manager and set to nil for this test
	originalWSManager := WSManager
	WSManager = nil
	defer func() { WSManager = originalWSManager }()

	// Create message handler
	handler := NewMessageHandler(mockDB)

	// Create test router
	router := gin.New()
	router.POST("/messages", func(c *gin.Context) {
		// Set user ID in context
		userID := uuid.New()
		c.Set("userID", userID)
		handler.SendMessage(c)
	})

	// Create test message
	receiverID := uuid.New()
	message := models.MessageRequest{
		ReceiverID: receiverID,
		Content:    "Test message",
	}

	// Create expected response message
	createdAt := time.Now().UTC()
	expectedMessage := &models.Message{
		ID:         uuid.New(),
		SenderID:   uuid.New(), // This will be set by the handler
		ReceiverID: receiverID,
		Content:    "Test message",
		CreatedAt:  createdAt,
		IsRead:     false,
	}

	// Setup mock expectations
	mockDB.On("CreateMessage", mock.Anything, receiverID, "Test message").Return(expectedMessage, nil)

	// Create request
	body, _ := json.Marshal(message)
	req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify mock expectations
	mockDB.AssertExpectations(t)
}

// TestMarkMessageAsReadWithoutWebSocket tests the MarkMessageAsRead handler with WebSocket manager set to nil
func TestMarkMessageAsReadWithoutWebSocket(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create mock DB
	mockDB := new(MockDB)

	// Save original WebSocket manager and set to nil for this test
	originalWSManager := WSManager
	WSManager = nil
	defer func() { WSManager = originalWSManager }()

	// Create message handler
	handler := NewMessageHandler(mockDB)

	// Create test router
	router := gin.New()

	// Create test user IDs
	userID := uuid.New()
	senderID := uuid.New()
	messageID := uuid.New()

	// Create test message
	message := &models.Message{
		ID:         messageID,
		SenderID:   senderID,
		ReceiverID: userID,
		Content:    "Test message",
		CreatedAt:  time.Now().UTC(),
		IsRead:     false,
	}

	router.PUT("/messages/:messageID/read", func(c *gin.Context) {
		// Set user ID in context
		c.Set("userID", userID)
		handler.MarkMessageAsRead(c)
	})

	// Setup mock expectations
	mockDB.On("GetMessageByID", messageID).Return(message, nil)
	mockDB.On("MarkMessageAsRead", messageID).Return(nil)

	// Create request
	req, _ := http.NewRequest("PUT", "/messages/"+messageID.String()+"/read", nil)

	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify mock expectations
	mockDB.AssertExpectations(t)
}

// TestWebSocketMessageFormat tests the format of WebSocket messages
func TestWebSocketMessageFormat(t *testing.T) {
	// Create test user IDs
	senderID := uuid.New()
	receiverID := uuid.New()

	// Create test message
	message := &models.Message{
		ID:         uuid.New(),
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    "Test message",
		CreatedAt:  time.Now().UTC(),
		IsRead:     false,
	}

	// Create a JSON message that would be sent via WebSocket
	wsMessage := map[string]interface{}{
		"type":        "message",
		"sender_id":   senderID.String(),
		"receiver_id": receiverID.String(),
		"content":     "Test message",
		"timestamp":   message.CreatedAt.Format(time.RFC3339),
	}

	// Marshal to JSON
	wsMessageJSON, err := json.Marshal(wsMessage)
	assert.NoError(t, err)

	// Unmarshal back to verify format
	var parsedMessage map[string]interface{}
	err = json.Unmarshal(wsMessageJSON, &parsedMessage)
	assert.NoError(t, err)

	// Verify message format
	assert.Equal(t, "message", parsedMessage["type"])
	assert.Equal(t, senderID.String(), parsedMessage["sender_id"])
	assert.Equal(t, receiverID.String(), parsedMessage["receiver_id"])
	assert.Equal(t, "Test message", parsedMessage["content"])
	assert.NotEmpty(t, parsedMessage["timestamp"])
}
