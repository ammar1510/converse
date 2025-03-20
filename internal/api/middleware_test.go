package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ammar1510/converse/internal/auth"
	"github.com/ammar1510/converse/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// setupAuthTestRouter creates a test router with the auth middleware
func setupAuthTestRouter(t *testing.T) *gin.Engine {
	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add auth middleware
	router.Use(AuthMiddleware())

	// Add test endpoint
	router.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("userID")
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{
			"userID":   userID,
			"username": username,
		})
	})

	return router
}

// setupTokenAuthTestRouter creates a test router with the token auth middleware
func setupTokenAuthTestRouter(t *testing.T) *gin.Engine {
	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add token auth middleware
	router.Use(TokenAuthMiddleware())

	// Add test endpoint
	router.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("userID")
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{
			"userID":   userID,
			"username": username,
		})
	})

	return router
}

// TestAuthMiddleware tests the authentication middleware
func TestAuthMiddleware(t *testing.T) {
	router := setupAuthTestRouter(t)

	// Create a test user and token
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}

	token, _, err := auth.GenerateToken(testUser)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		token      string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "valid token",
			token:      token,
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "no token",
			token:      "",
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
		{
			name:       "invalid token format",
			token:      "invalid.token.string",
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
		{
			name:       "missing Bearer prefix",
			token:      token,
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/test", nil)

			// Set authorization header based on test case
			if tt.token != "" {
				if tt.name == "missing Bearer prefix" {
					req.Header.Set("Authorization", tt.token)
				} else {
					req.Header.Set("Authorization", "Bearer "+tt.token)
				}
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantError {
				// Parse response
				var response struct {
					UserID   string `json:"userID"`
					Username string `json:"username"`
				}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Verify response
				assert.Equal(t, testUser.ID.String(), response.UserID)
				assert.Equal(t, testUser.Username, response.Username)
			}
		})
	}
}

// TestTokenAuthMiddleware tests the token authentication middleware
// which accepts tokens from both Authorization header and URL parameters
func TestTokenAuthMiddleware(t *testing.T) {
	router := setupTokenAuthTestRouter(t)

	// Create a test user and token
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}

	token, _, err := auth.GenerateToken(testUser)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		headerToken string
		urlToken    string
		wantStatus  int
		wantError   bool
	}{
		{
			name:        "valid token in header",
			headerToken: token,
			urlToken:    "",
			wantStatus:  http.StatusOK,
			wantError:   false,
		},
		{
			name:        "valid token in URL",
			headerToken: "",
			urlToken:    token,
			wantStatus:  http.StatusOK,
			wantError:   false,
		},
		{
			name:        "valid token in both header and URL (header takes precedence)",
			headerToken: token,
			urlToken:    "invalid-token",
			wantStatus:  http.StatusOK,
			wantError:   false,
		},
		{
			name:        "no token",
			headerToken: "",
			urlToken:    "",
			wantStatus:  http.StatusUnauthorized,
			wantError:   true,
		},
		{
			name:        "invalid token in header",
			headerToken: "invalid.token.string",
			urlToken:    "",
			wantStatus:  http.StatusUnauthorized,
			wantError:   true,
		},
		{
			name:        "invalid token in URL",
			headerToken: "",
			urlToken:    "invalid.token.string",
			wantStatus:  http.StatusUnauthorized,
			wantError:   true,
		},
		{
			name:        "missing Bearer prefix in header",
			headerToken: "no-bearer-prefix",
			urlToken:    "",
			wantStatus:  http.StatusUnauthorized,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with appropriate URL parameter if needed
			var req *http.Request
			if tt.urlToken != "" {
				req = httptest.NewRequest("GET", "/test?token="+tt.urlToken, nil)
			} else {
				req = httptest.NewRequest("GET", "/test", nil)
			}

			// Set authorization header if needed
			if tt.headerToken != "" {
				if tt.name == "missing Bearer prefix in header" {
					req.Header.Set("Authorization", tt.headerToken)
				} else {
					req.Header.Set("Authorization", "Bearer "+tt.headerToken)
				}
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantError {
				// Parse response
				var response struct {
					UserID   string `json:"userID"`
					Username string `json:"username"`
				}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Verify response
				assert.Equal(t, testUser.ID.String(), response.UserID)
				assert.Equal(t, testUser.Username, response.Username)
			}
		})
	}
}
