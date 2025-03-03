package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/ammar1510/converse/internal/models"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	// This variable will be initialized either from environment 
	// variables or explicitly via InitJWTKey function
	jwtKey = []byte(os.Getenv("JWT_SECRET"))
)

// InitJWTKey initializes the JWT key with the provided secret
// This allows for explicit initialization after environment variables are loaded
// or for setting a custom key during testing
func InitJWTKey(key []byte) {
	jwtKey = key
}

// JWTClaims represents the claims in the JWT
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(user *models.User) (string, time.Time, error) {
	// Check for nil user
	if user == nil {
		return "", time.Time{}, errors.New("user cannot be nil")
	}
	
	// Check for zero UUID (missing ID)
	if user.ID == uuid.Nil {
		return "", time.Time{}, errors.New("user ID cannot be empty")
	}
	
	expirationTime := time.Now().Add(24 * time.Hour)
	
	claims := &JWTClaims{
		UserID:   user.ID.String(),
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	
	return tokenString, expirationTime, err
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	
	return claims, nil
}

// GetUserIDFromToken extracts the UserID from claims
func GetUserIDFromToken(claims *JWTClaims) (uuid.UUID, error) {
	if claims == nil {
		return uuid.Nil, errors.New("claims cannot be nil")
	}
	return uuid.Parse(claims.UserID)
}
