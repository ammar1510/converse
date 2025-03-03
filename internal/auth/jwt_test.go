package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ammar1510/converse/internal/models"
)

// TestInitJWTKey tests the JWT initialization 
func TestInitJWTKey(t *testing.T) {
	// Setup
	testKey := []byte("test-secret-key-for-jwt-tests")
	
	// Test
	InitJWTKey(testKey)
	
	// Since jwtKey is private, we can only test indirectly
	// by generating and validating a token
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}
	
	token, _, err := GenerateToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

// TestGenerateToken tests token generation functionality
func TestGenerateToken(t *testing.T) {
	// Setup 
	testKey := []byte("test-secret-key-for-jwt-tests")
	InitJWTKey(testKey)
	
	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name: "valid user",
			user: &models.User{
				ID:       uuid.New(),
				Username: "testuser",
				Email:    "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "missing user ID",
			user: &models.User{
				Username: "testuser",
				Email:    "test@example.com",
			},
			wantErr: true,
		},
		{
			name:    "nil user",
			user:    nil,
			wantErr: true,
		},
	}
	
	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, expiry, err := GenerateToken(tt.user)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				
				// Verify expiration time is in the future
				assert.True(t, expiry.After(time.Now()))
				
				// Verify token can be validated
				claims, err := ValidateToken(token)
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tt.user.ID.String(), claims.UserID)
				assert.Equal(t, tt.user.Username, claims.Username)
			}
		})
	}
}

// TestValidateToken tests token validation
func TestValidateToken(t *testing.T) {
	// Setup
	testKey := []byte("test-secret-key-for-jwt-tests")
	InitJWTKey(testKey)
	
	// Create a valid user and token
	validUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}
	validToken, _, err := GenerateToken(validUser)
	assert.NoError(t, err)
	
	tests := []struct {
		name        string
		tokenString string
		wantErr     bool
	}{
		{
			name:        "valid token",
			tokenString: validToken,
			wantErr:     false,
		},
		{
			name:        "empty token",
			tokenString: "",
			wantErr:     true,
		},
		{
			name:        "invalid token format",
			tokenString: "not.a.valid.jwt.token",
			wantErr:     true,
		},
		{
			name:        "tampered token",
			tokenString: validToken + "tampered",
			wantErr:     true,
		},
	}
	
	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.tokenString)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, validUser.ID.String(), claims.UserID)
				assert.Equal(t, validUser.Username, claims.Username)
			}
		})
	}
}

// TestGetUserIDFromToken tests extracting user ID from claims
func TestGetUserIDFromToken(t *testing.T) {
	// Setup
	testKey := []byte("test-secret-key-for-jwt-tests")
	InitJWTKey(testKey)
	
	// Create a valid user and get claims
	validUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}
	validToken, _, err := GenerateToken(validUser)
	assert.NoError(t, err)
	
	validClaims, err := ValidateToken(validToken)
	assert.NoError(t, err)
	
	// Create invalid claims
	invalidClaims := &JWTClaims{
		UserID:   "not-a-valid-uuid",
		Username: "testuser",
	}
	
	tests := []struct {
		name    string
		claims  *JWTClaims
		wantErr bool
	}{
		{
			name:    "valid claims",
			claims:  validClaims,
			wantErr: false,
		},
		{
			name:    "invalid UUID format",
			claims:  invalidClaims,
			wantErr: true,
		},
		{
			name:    "nil claims",
			claims:  nil,
			wantErr: true,
		},
	}
	
	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := GetUserIDFromToken(tt.claims)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, uuid.Nil, userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, validUser.ID, userID)
			}
		})
	}
}

