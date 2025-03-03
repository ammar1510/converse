package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/ammar1510/converse/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrMessageNotFound   = errors.New("message not found")
)

// DB is the database connection
type DB struct {
	*sql.DB
}

// NewDB creates a new database connection
func NewDB(connStr string) (*DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// CreateUser stores a new user in the database
func (db *DB) CreateUser(username, email, passwordHash string) (*models.User, error) {
	// Check if user already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2",
		username, email).Scan(&count)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, ErrUserAlreadyExists
	}

	// Create new user
	user := &models.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		LastSeen:     time.Now(),
	}

	_, err = db.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, last_seen) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt, user.LastSeen)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}

	err := db.QueryRow(`
		SELECT id, username, email, password_hash, 
		       COALESCE(display_name, ''), COALESCE(avatar_url, ''), 
		       created_at, last_seen 
		FROM users WHERE email = $1`, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.AvatarURL, &user.CreatedAt, &user.LastSeen)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateLastSeen updates the user's last_seen timestamp
func (db *DB) UpdateLastSeen(userID uuid.UUID) error {
	result, err := db.Exec("UPDATE users SET last_seen = $1 WHERE id = $2",
		time.Now(), userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetUserByID retrieves a user by their ID
func (db *DB) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := db.QueryRow(`
		SELECT id, username, email, password_hash, 
		       COALESCE(display_name, ''), COALESCE(avatar_url, ''), 
		       created_at, last_seen 
		FROM users WHERE id = $1`,
		id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.LastSeen,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateMessage stores a new message in the database
func (db *DB) CreateMessage(senderID, receiverID uuid.UUID, content string) (*models.Message, error) {
	// Verify both users exist
	_, err := db.GetUserByID(senderID)
	if err != nil {
		return nil, err
	}

	_, err = db.GetUserByID(receiverID)
	if err != nil {
		return nil, err
	}

	message := &models.Message{
		ID:         uuid.New(),
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		CreatedAt:  time.Now().UTC(),
		IsRead:     false,
	}

	_, err = db.Exec(
		"INSERT INTO messages (id, sender_id, receiver_id, content, created_at, is_read) VALUES ($1, $2, $3, $4, $5, $6)",
		message.ID, message.SenderID, message.ReceiverID, message.Content, message.CreatedAt, message.IsRead,
	)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// GetMessagesByUser retrieves all messages where the user is either sender or receiver
func (db *DB) GetMessagesByUser(userID uuid.UUID) ([]*models.Message, error) {
	rows, err := db.Query(
		"SELECT id, sender_id, receiver_id, content, created_at, is_read, updated_at FROM messages WHERE sender_id = $1 OR receiver_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		var updatedAt sql.NullTime

		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.CreatedAt, &msg.IsRead, &updatedAt)
		if err != nil {
			return nil, err
		}

		if updatedAt.Valid {
			msg.UpdatedAt = &updatedAt.Time
		}

		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// GetConversation retrieves messages between two users
func (db *DB) GetConversation(userID1, userID2 uuid.UUID) ([]*models.Message, error) {
	rows, err := db.Query(
		`SELECT id, sender_id, receiver_id, content, created_at, is_read, updated_at 
		FROM messages 
		WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1) 
		ORDER BY created_at ASC`,
		userID1, userID2,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		var updatedAt sql.NullTime

		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.CreatedAt, &msg.IsRead, &updatedAt)
		if err != nil {
			return nil, err
		}

		if updatedAt.Valid {
			msg.UpdatedAt = &updatedAt.Time
		}

		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// MarkMessageAsRead marks a message as read
func (db *DB) MarkMessageAsRead(messageID uuid.UUID) error {
	now := time.Now().UTC()
	result, err := db.Exec(
		"UPDATE messages SET is_read = true, updated_at = $1 WHERE id = $2",
		now, messageID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrMessageNotFound
	}

	return nil
}
