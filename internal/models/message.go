package models

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a chat message in the system
type Message struct {
	ID         uuid.UUID  `json:"id"`
	SenderID   uuid.UUID  `json:"sender_id"`
	ReceiverID uuid.UUID  `json:"receiver_id"`
	Content    string     `json:"content"`
	CreatedAt  time.Time  `json:"created_at"`
	IsRead     bool       `json:"is_read"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

// MessageRequest is the structure for message creation requests
type MessageRequest struct {
	ReceiverID uuid.UUID `json:"receiver_id" binding:"required"`
	Content    string    `json:"content" binding:"required,min=1"`
}

// MessageResponse is what we return to clients
type MessageResponse struct {
	ID         uuid.UUID     `json:"id"`
	SenderID   uuid.UUID     `json:"sender_id"`
	ReceiverID uuid.UUID     `json:"receiver_id"`
	Content    string        `json:"content"`
	CreatedAt  time.Time     `json:"created_at"`
	IsRead     bool          `json:"is_read"`
	UpdatedAt  *time.Time    `json:"updated_at,omitempty"`
	Sender     *UserResponse `json:"sender,omitempty"`
}
