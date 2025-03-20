package auth

import (
	"testing"
	"time"

	"github.com/ammar1510/converse/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInitJWTKey(t *testing.T) {
	testKey := []byte("test-secret-key-for-jwt-tests")

	InitJWTKey(testKey)

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}

	token, _, err := GenerateToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateToken(t *testing.T) {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, expiry, err := GenerateToken(tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				assert.True(t, expiry.After(time.Now()))

				claims, err := ValidateToken(token)
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tt.user.ID.String(), claims.UserID)
				assert.Equal(t, tt.user.Username, claims.Username)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	testKey := []byte("test-secret-key-for-jwt-tests")
	InitJWTKey(testKey)

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

func TestGetUserIDFromToken(t *testing.T) {
	testKey := []byte("test-secret-key-for-jwt-tests")
	InitJWTKey(testKey)

	validUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}
	validToken, _, err := GenerateToken(validUser)
	assert.NoError(t, err)

	validClaims, err := ValidateToken(validToken)
	assert.NoError(t, err)

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
