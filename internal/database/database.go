package database

import (
	"github.com/ammar1510/converse/internal/models"
	"github.com/google/uuid"
)

// DBInterface defines the methods that any database implementation must provide
type DBInterface interface {
	// User methods
	CreateUser(username, email, passwordHash string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	UpdateLastSeen(userID uuid.UUID) error

	// Message methods
	CreateMessage(senderID, receiverID uuid.UUID, content string) (*models.Message, error)
	GetMessagesByUser(userID uuid.UUID) ([]*models.Message, error)
	GetConversation(userID1, userID2 uuid.UUID) ([]*models.Message, error)
	MarkMessageAsRead(messageID uuid.UUID) error
}
