package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ammar1510/converse/internal/auth"
	"github.com/ammar1510/converse/internal/database"
	"github.com/ammar1510/converse/internal/models"
)

// AuthHandler handles authentication routes
type AuthHandler struct {
	DB database.DBInterface
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db database.DBInterface) *AuthHandler {
	return &AuthHandler{DB: db}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var input models.UserRegistration

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Create user
	user, err := h.DB.CreateUser(input.Username, input.Email, hashedPassword)
	if err == database.ErrUserAlreadyExists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Return user data (without password)
	c.JSON(http.StatusCreated, models.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		CreatedAt:   user.CreatedAt,
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var input models.UserLogin

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user by email
	user, err := h.DB.GetUserByEmail(input.Email)
	if err == database.ErrUserNotFound {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Check password
	if !auth.CheckPasswordHash(input.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update last seen
	if err := h.DB.UpdateLastSeen(user.ID); err != nil {
		// Just log this error, don't return it
		// log.Printf("Failed to update last_seen: %v", err)
	}

	// Generate JWT token
	token, expiry, err := auth.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return user data with token
	c.JSON(http.StatusOK, gin.H{
		"token":  token,
		"expiry": expiry,
		"user": models.UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
			CreatedAt:   user.CreatedAt,
		},
	})
}

// GetMe gets the current user profile
func (h *AuthHandler) GetMe(c *gin.Context) {
	// The user should be added to context by auth middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// The userID is already a UUID from the middleware
	userUUID := userID.(uuid.UUID)

	// Get user from database
	user, err := h.DB.GetUserByID(userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Return user data
	c.JSON(http.StatusOK, models.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		CreatedAt:   user.CreatedAt,
	})
}
