package database

import (
	"database/sql"
	"errors"
	"fmt"
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

type PostgresDB struct {
	*sql.DB
}

func NewPostgresDB(connStr string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresDB{db}, nil
}

func (db *PostgresDB) CreateUser(username, email, passwordHash string) (*models.User, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2",
		username, email).Scan(&count)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, ErrUserAlreadyExists
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		LastSeen:     time.Now(),
	}

	_, err = db.Exec(
		"INSERT INTO users (id, username, email, password_hash, created_at, last_seen) VALUES ($1, $2, $3, $4, $5, $6)",
		user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt, user.LastSeen,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *PostgresDB) GetUserByEmail(email string) (*models.User, error) {
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

func (db *PostgresDB) UpdateLastSeen(userID uuid.UUID) error {
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

func (db *PostgresDB) GetUserByID(id uuid.UUID) (*models.User, error) {
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

func (db *PostgresDB) CreateMessage(senderID, receiverID uuid.UUID, content string) (*models.Message, error) {
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

func (db *PostgresDB) GetMessagesByUser(userID uuid.UUID) ([]*models.Message, error) {
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

func (db *PostgresDB) GetMessageByID(messageID uuid.UUID) (*models.Message, error) {
	var msg models.Message
	var updatedAt sql.NullTime

	err := db.QueryRow(`
		SELECT id, sender_id, receiver_id, content, created_at, is_read, updated_at 
		FROM messages 
		WHERE id = $1`,
		messageID).Scan(
		&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content,
		&msg.CreatedAt, &msg.IsRead, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrMessageNotFound
	}
	if err != nil {
		return nil, err
	}

	if updatedAt.Valid {
		msg.UpdatedAt = &updatedAt.Time
	}

	return &msg, nil
}

func (db *PostgresDB) GetConversation(userID1, userID2 uuid.UUID) ([]*models.Message, error) {
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

func (db *PostgresDB) MarkMessageAsRead(messageID uuid.UUID) error {
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

func (db *PostgresDB) Close() error {
	return db.DB.Close()
}

func (db *PostgresDB) Exec(query string, args ...interface{}) (ExecResult, error) {
	return db.DB.Exec(query, args...)
}

func (db *PostgresDB) GetAllUsers(excludeUserID uuid.UUID) ([]*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, created_at, last_seen
		FROM users
		WHERE id != $1
		ORDER BY username
	`

	rows, err := db.Query(query, excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var displayName, avatarURL sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&displayName,
			&avatarURL,
			&user.CreatedAt,
			&user.LastSeen,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}

		if displayName.Valid {
			user.DisplayName = displayName.String
		}
		if avatarURL.Valid {
			user.AvatarURL = avatarURL.String
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}
