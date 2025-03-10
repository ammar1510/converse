package database

import (
	"testing"
	"time"

	"github.com/ammar1510/converse/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *PostgresDB {
	// Use test database connection string
	connStr := "postgres://ammar3.shaikh@localhost:5432/converse_test?sslmode=disable"
	db, err := NewPostgresDB(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up test data
	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Failed to clean up test data: %v", err)
	}

	return db
}

// TestNewPostgresDB tests database connection creation
func TestNewPostgresDB(t *testing.T) {
	tests := []struct {
		name      string
		connStr   string
		wantError bool
	}{
		{
			name:      "valid connection string",
			connStr:   "postgres://ammar3.shaikh@localhost:5432/converse_test?sslmode=disable",
			wantError: false,
		},
		{
			name:      "invalid connection string",
			connStr:   "invalid connection string",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewPostgresDB(tt.connStr)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
				defer db.Close()
			}
		})
	}
}

// TestCreateUser tests user creation functionality
func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name      string
		username  string
		email     string
		password  string
		wantError bool
	}{
		{
			name:      "valid user",
			username:  "testuser",
			email:     "test@example.com",
			password:  "hashedpassword123",
			wantError: false,
		},
		{
			name:      "duplicate email",
			username:  "testuser2",
			email:     "test@example.com", // Same email as above
			password:  "hashedpassword456",
			wantError: true,
		},
		{
			name:      "duplicate username",
			username:  "testuser", // Same username as first test
			email:     "test2@example.com",
			password:  "hashedpassword789",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := db.CreateUser(tt.username, tt.email, tt.password)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, user)
				if tt.name == "duplicate email" {
					assert.Equal(t, ErrUserAlreadyExists, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.username, user.Username)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.password, user.PasswordHash)
				assert.NotEqual(t, uuid.Nil, user.ID)
				assert.True(t, user.CreatedAt.Before(time.Now()))
			}
		})
	}
}

// TestGetUserByEmail tests user retrieval by email
func TestGetUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user
	testUser := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
		CreatedAt:    time.Now(),
		LastSeen:     time.Now(),
	}

	_, err := db.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, last_seen) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		testUser.ID, testUser.Username, testUser.Email, testUser.PasswordHash,
		testUser.CreatedAt, testUser.LastSeen)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		email     string
		wantError bool
	}{
		{
			name:      "existing user",
			email:     "test@example.com",
			wantError: false,
		},
		{
			name:      "non-existent user",
			email:     "nonexistent@example.com",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := db.GetUserByEmail(tt.email)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, user)
				assert.Equal(t, ErrUserNotFound, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, testUser.ID, user.ID)
				assert.Equal(t, testUser.Username, user.Username)
				assert.Equal(t, testUser.Email, user.Email)
				assert.Equal(t, testUser.PasswordHash, user.PasswordHash)
			}
		})
	}
}

// TestUpdateLastSeen tests updating user's last seen timestamp
func TestUpdateLastSeen(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user
	testUser := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
		CreatedAt:    time.Now(),
		LastSeen:     time.Now(),
	}

	_, err := db.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, last_seen) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		testUser.ID, testUser.Username, testUser.Email, testUser.PasswordHash,
		testUser.CreatedAt, testUser.LastSeen)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		userID    uuid.UUID
		wantError bool
	}{
		{
			name:      "existing user",
			userID:    testUser.ID,
			wantError: false,
		},
		{
			name:      "non-existent user",
			userID:    uuid.New(),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.UpdateLastSeen(tt.userID)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the last_seen timestamp was updated
				var lastSeen time.Time
				err = db.QueryRow("SELECT last_seen FROM users WHERE id = $1", tt.userID).Scan(&lastSeen)
				assert.NoError(t, err)
				assert.True(t, lastSeen.After(testUser.LastSeen))
			}
		})
	}
}
