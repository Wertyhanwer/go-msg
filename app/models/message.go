package models

import (
	"time"
)

// Message represents a message in the system
type Message struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	RecipientID int       `json:"recipient_id"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"created_at"`
	// Populated fields for API responses
	SenderName    string `json:"sender_name,omitempty"`
	RecipientName string `json:"recipient_name,omitempty"`
}

// NewMessage creates a new message with the given parameters
func NewMessage(userID, recipientID int, text string) *Message {
	return &Message{
		UserID:      userID,
		RecipientID: recipientID,
		Text:        text,
		CreatedAt:   time.Now(),
	}
}
