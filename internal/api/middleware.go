package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ammar1510/converse/internal/auth"
	"github.com/ammar1510/converse/internal/models"
)

// AuthMiddleware validates JWT tokens and sets user info in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		
		// Check if Authorization header exists and has Bearer format
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}
		
		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		// Validate token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		
		// Set user ID in context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		
		c.Next()
	}
} 