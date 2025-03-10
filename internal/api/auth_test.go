package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ammar1510/converse/internal/auth"
	"github.com/ammar1510/converse/internal/database"
	"github.com/ammar1510/converse/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// setupTestRouter creates a test router with the auth handler
func setupTestRouter(t *testing.T) (*gin.Engine, *AuthHandler) {
	// Create test database connection
	connStr := "postgres://ammar3.shaikh@localhost:5432/converse_test?sslmode=disable"
	db, err := database.NewDatabase(database.PostgreSQL, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up test data
	_, err = db.Exec("DELETE FROM messages")
	if err != nil {
		t.Fatalf("Failed to clean up test data (messages): %v", err)
	}

	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Failed to clean up test data (users): %v", err)
	}

	// Create auth handler
	handler := NewAuthHandler(db)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup routes
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)
	router.GET("/me", AuthMiddleware(), handler.GetMe)

	return router, handler
}

// TestRegister tests user registration endpoint
func TestRegister(t *testing.T) {
	router, _ := setupTestRouter(t)

	tests := []struct {
		name       string
		input      models.UserRegistration
		wantStatus int
		wantError  bool
	}{
		{
			name: "valid registration",
			input: models.UserRegistration{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name: "duplicate email",
			input: models.UserRegistration{
				Username: "testuser2",
				Email:    "test@example.com", // Same email as above
				Password: "password456",
			},
			wantStatus: http.StatusConflict,
			wantError:  true,
		},
		{
			name: "invalid input",
			input: models.UserRegistration{
				Username: "", // Empty username
				Email:    "invalid-email",
				Password: "", // Empty password
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantError {
				// Parse response
				var response models.UserResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Verify response
				assert.Equal(t, tt.input.Username, response.Username)
				assert.Equal(t, tt.input.Email, response.Email)
				assert.NotEqual(t, uuid.Nil, response.ID)
			}
		})
	}
}

// TestLogin tests user login endpoint
func TestLogin(t *testing.T) {
	router, _ := setupTestRouter(t)

	// Create a test user
	hashedPassword, err := auth.HashPassword("password123")
	assert.NoError(t, err)

	connStr := "postgres://ammar3.shaikh@localhost:5432/converse_test?sslmode=disable"
	db, err := database.NewDatabase(database.PostgreSQL, connStr)
	assert.NoError(t, err)
	defer db.Close()

	_, err = db.CreateUser("testuser", "test@example.com", hashedPassword)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		input      models.UserLogin
		wantStatus int
		wantError  bool
	}{
		{
			name: "valid login",
			input: models.UserLogin{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "invalid password",
			input: models.UserLogin{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
		{
			name: "non-existent user",
			input: models.UserLogin{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
		{
			name: "invalid input",
			input: models.UserLogin{
				Email:    "invalid-email",
				Password: "", // Empty password
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantError {
				// Parse response
				var response struct {
					Token  string              `json:"token"`
					Expiry string              `json:"expiry"`
					User   models.UserResponse `json:"user"`
				}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Verify response
				assert.NotEmpty(t, response.Token)
				assert.NotEmpty(t, response.Expiry)
				assert.Equal(t, tt.input.Email, response.User.Email)

				// Verify token
				claims, err := auth.ValidateToken(response.Token)
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, "testuser", claims.Username)
			}
		})
	}
}

// TestGetMe tests the get current user profile endpoint
func TestGetMe(t *testing.T) {
	router, _ := setupTestRouter(t)

	// Create a test user and token
	hashedPassword, err := auth.HashPassword("password123")
	assert.NoError(t, err)

	connStr := "postgres://ammar3.shaikh@localhost:5432/converse_test?sslmode=disable"
	db, err := database.NewDatabase(database.PostgreSQL, connStr)
	assert.NoError(t, err)
	defer db.Close()

	user, err := db.CreateUser("testuser", "test@example.com", hashedPassword)
	assert.NoError(t, err)

	token, _, err := auth.GenerateToken(user)
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
			name:       "invalid token",
			token:      "invalid.token.string",
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/me", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantError {
				// Parse response
				var response models.UserResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Verify response
				assert.Equal(t, user.ID, response.ID)
				assert.Equal(t, user.Username, response.Username)
				assert.Equal(t, user.Email, response.Email)
			}
		})
	}
}
