package database

import (
	"fmt"

	"github.com/ammar1510/converse/internal/models"
	"github.com/google/uuid"
)

type DBInterface interface {
	// User methods
	CreateUser(username, email, passwordHash string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	UpdateLastSeen(userID uuid.UUID) error
	GetAllUsers(excludeUserID uuid.UUID) ([]*models.User, error)

	// Message methods
	CreateMessage(senderID, receiverID uuid.UUID, content string) (*models.Message, error)
	GetMessagesByUser(userID uuid.UUID) ([]*models.Message, error)
	GetMessageByID(messageID uuid.UUID) (*models.Message, error)
	GetConversation(userID1, userID2 uuid.UUID) ([]*models.Message, error)
	MarkMessageAsRead(messageID uuid.UUID) error

	// Common methods
	Exec(query string, args ...interface{}) (ExecResult, error)
	Close() error
}

type ExecResult interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgres"
	MySQL      DatabaseType = "mysql"
)

func NewDatabase(dbType DatabaseType, connStr string) (DBInterface, error) {
	switch dbType {
	case PostgreSQL:
		return NewPostgresDB(connStr)
	case MySQL:
		return nil, fmt.Errorf("MySQL implementation not available yet")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
