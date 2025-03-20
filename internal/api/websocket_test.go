package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ammar1510/converse/internal/auth"
	"github.com/ammar1510/converse/internal/models"
	"github.com/ammar1510/converse/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gorilla "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupWsTokenAuthRouter creates a test router with the WebSocket route protected by TokenAuthMiddleware
func setupWsTokenAuthRouter(t *testing.T) (*gin.Engine, *websocket.Manager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create WebSocket manager
	wsManager := websocket.NewManager()
	go wsManager.Run()

	// Save the global WebSocket manager
	originalWSManager := WSManager
	WSManager = wsManager
	t.Cleanup(func() {
		WSManager = originalWSManager
	})

	// Create a group with TokenAuthMiddleware
	wsRoute := router.Group("/api")
	wsRoute.Use(TokenAuthMiddleware())
	wsRoute.GET("/ws", func(c *gin.Context) {
		wsManager.HandleWebSocket(c)
	})

	return router, wsManager
}

// TestWebSocketWithTokenAuthMiddleware tests the WebSocket endpoint with TokenAuthMiddleware
func TestWebSocketWithTokenAuthMiddleware(t *testing.T) {
	router, _ := setupWsTokenAuthRouter(t)

	// Create a test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Create a test user and token
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}

	token, _, err := auth.GenerateToken(testUser)
	require.NoError(t, err)

	// Test cases for different authentication methods
	tests := []struct {
		name          string
		urlPath       string
		headers       map[string]string
		expectedCode  int
		shouldConnect bool
	}{
		{
			name:          "valid token in URL parameter",
			urlPath:       "/api/ws?token=" + token,
			headers:       nil,
			expectedCode:  http.StatusSwitchingProtocols,
			shouldConnect: true,
		},
		{
			name:    "valid token in Authorization header",
			urlPath: "/api/ws",
			headers: map[string]string{
				"Authorization": "Bearer " + token,
			},
			expectedCode:  http.StatusSwitchingProtocols,
			shouldConnect: true,
		},
		{
			name:          "no token provided",
			urlPath:       "/api/ws",
			headers:       nil,
			expectedCode:  http.StatusUnauthorized,
			shouldConnect: false,
		},
		{
			name:          "invalid token in URL parameter",
			urlPath:       "/api/ws?token=invalid.token",
			headers:       nil,
			expectedCode:  http.StatusUnauthorized,
			shouldConnect: false,
		},
		{
			name:    "invalid token in Authorization header",
			urlPath: "/api/ws",
			headers: map[string]string{
				"Authorization": "Bearer invalid.token",
			},
			expectedCode:  http.StatusUnauthorized,
			shouldConnect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert http:// to ws://
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + tt.urlPath

			// Setup headers
			header := http.Header{}
			for k, v := range tt.headers {
				header.Add(k, v)
			}

			// Attempt to connect
			ws, resp, err := gorilla.DefaultDialer.Dial(wsURL, header)

			// Check response code
			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.shouldConnect {
				// Connection should succeed
				assert.NoError(t, err)
				require.NotNil(t, ws)

				// Wait for client to be registered
				time.Sleep(100 * time.Millisecond)

				// Instead of directly accessing unexported fields, we'll use an indirect way
				// to check if the client is connected by trying to send a message
				connected := false
				testMessage := websocket.WebSocketMessage{
					Type:    "test",
					Content: "test message",
				}
				messageJSON, err := json.Marshal(testMessage)
				require.NoError(t, err)

				err = ws.WriteMessage(gorilla.TextMessage, messageJSON)
				if err == nil {
					// If we can send a message, the client is connected
					connected = true
				}
				assert.True(t, connected, "Client should be connected")

				// Close connection
				ws.WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, ""))
				ws.Close()

				// Wait for connection to close
				time.Sleep(100 * time.Millisecond)
			} else {
				// Connection should fail
				assert.Error(t, err)
			}
		})
	}
}
