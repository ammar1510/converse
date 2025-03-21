package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ammar1510/converse/internal/auth"
	"github.com/ammar1510/converse/internal/logger"
)

var mwLog = logger.New("api-middleware")

// AuthMiddleware validates JWT tokens and sets user info in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// Check if Authorization header exists and has Bearer format
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			mwLog.Debug("Missing or invalid Authorization header from %s", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			mwLog.Debug("Invalid token from %s: %v", c.ClientIP(), err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Parse user ID string into UUID
		userUUID, err := uuid.Parse(claims.UserID)
		if err != nil {
			mwLog.Error("Invalid user ID format in token: %s", claims.UserID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format in token"})
			c.Abort()
			return
		}

		// Set user ID (as UUID) and username in context
		c.Set("userID", userUUID)
		c.Set("username", claims.Username)
		mwLog.Debug("User %s (%s) authenticated", claims.Username, userUUID)

		c.Next()
	}
}

// TokenAuthMiddleware validates JWT tokens from either header or URL parameter
// This is especially useful for WebSocket connections where setting headers can be problematic
func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// First try to get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			mwLog.Debug("Token found in Authorization header from %s", c.ClientIP())
		} else {
			// If no valid Authorization header, try to get token from URL parameter
			tokenString = c.Query("token")
			if tokenString == "" {
				mwLog.Debug("No authentication token provided from %s", c.ClientIP())
				c.JSON(http.StatusUnauthorized, gin.H{"error": "No authentication token provided"})
				c.Abort()
				return
			}
			mwLog.Debug("Token found in URL parameter from %s", c.ClientIP())
		}

		// Validate token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			mwLog.Debug("Invalid token from %s: %v", c.ClientIP(), err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Parse user ID string into UUID
		userUUID, err := uuid.Parse(claims.UserID)
		if err != nil {
			mwLog.Error("Invalid user ID format in token: %s", claims.UserID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format in token"})
			c.Abort()
			return
		}

		// Set user ID (as UUID) and username in context
		c.Set("userID", userUUID)
		c.Set("username", claims.Username)
		mwLog.Debug("User %s (%s) authenticated via TokenAuthMiddleware", claims.Username, userUUID)

		c.Next()
	}
}
