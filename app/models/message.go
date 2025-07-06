package models

import (
	"time"
)

// Message represents a message in the system
type Message struct {
	ID        int       `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// NewMessage creates a new message with the given text
func NewMessage(text string) *Message {
	return &Message{
		Text:      text,
		CreatedAt: time.Now(),
	}
}
