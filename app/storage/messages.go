package storage

import (
	"context"
	"fmt"
	"time"
)

// Message represents a message in the database
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

// CreateMessage creates a new message in the database
func (db *DB) CreateMessage(ctx context.Context, userID, recipientID int, text string) (*Message, error) {
	var message Message
	err := db.pool.QueryRow(ctx,
		"INSERT INTO messages (user_id, recipient_id, text) VALUES ($1, $2, $3) RETURNING id, user_id, recipient_id, text, created_at",
		userID, recipientID, text,
	).Scan(&message.ID, &message.UserID, &message.RecipientID, &message.Text, &message.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create message: %v", err)
	}

	return &message, nil
}

// GetMessage retrieves a message by ID
func (db *DB) GetMessage(ctx context.Context, id int) (*Message, error) {	defer	var message Message
	err := db.pool.QueryRow(ctx,
		`SELECT m.id, m.user_id, m.recipient_id, m.text, m.created_at,
		 u1.first_name || ' ' || u1.last_name as sender_name,
		 u2.first_name || ' ' || u2.last_name as recipient_name
		 FROM messages m
		 JOIN users u1 ON m.user_id = u1.id
		 JOIN users u2 ON m.recipient_id = u2.id
		 WHERE m.id = $1`,
		id,
	).Scan(&message.ID, &message.UserID, &message.RecipientID, &message.Text, &message.CreatedAt, &message.SenderName, &message.RecipientName)

	if err != nil {
		return nil, fmt.Errorf("failed to get message: %v", err)
	}

	return &message, nil
}

// UpdateMessage updates a message by ID
func (db *DB) UpdateMessage(ctx context.Context, id int, text string) (*Message, error) {	defer	var message Message
	err := db.pool.QueryRow(ctx,
		"UPDATE messages SET text = $1 WHERE id = $2 RETURNING id, user_id, recipient_id, text, created_at",
		text, id,
	).Scan(&message.ID, &message.UserID, &message.RecipientID, &message.Text, &message.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update message: %v", err)
	}

	return &message, nil
}

// DeleteMessage deletes a message by ID
func (db *DB) DeleteMessage(ctx context.Context, id int) error {	defer	result, err := db.pool.Exec(ctx,
		"DELETE FROM messages WHERE id = $1",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete message: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// GetAllMessages retrieves a list of messages with pagination
func (db *DB) GetAllMessages(ctx context.Context, limit, offset int) ([]Message, error) {	defer	rows, err := db.pool.Query(ctx,
		`SELECT m.id, m.user_id, m.recipient_id, m.text, m.created_at,
		 u1.first_name || ' ' || u1.last_name as sender_name,
		 u2.first_name || ' ' || u2.last_name as recipient_name
		 FROM messages m
		 JOIN users u1 ON m.user_id = u1.id
		 JOIN users u2 ON m.recipient_id = u2.id
		 ORDER BY m.created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %v", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.RecipientID, &msg.Text, &msg.CreatedAt, &msg.SenderName, &msg.RecipientName); err != nil {
			return nil, fmt.Errorf("failed to scan message: %v", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %v", err)
	}

	return messages, nil
}

// GetMessagesBetweenUsers retrieves messages between two users
func (db *DB) GetMessagesBetweenUsers(ctx context.Context, userID1, userID2 int, limit, offset int) ([]Message, error) {	defer	rows, err := db.pool.Query(ctx,
		`SELECT m.id, m.user_id, m.recipient_id, m.text, m.created_at,
		 u1.first_name || ' ' || u1.last_name as sender_name,
		 u2.first_name || ' ' || u2.last_name as recipient_name
		 FROM messages m
		 JOIN users u1 ON m.user_id = u1.id
		 JOIN users u2 ON m.recipient_id = u2.id
		 WHERE (m.user_id = $1 AND m.recipient_id = $2) OR (m.user_id = $2 AND m.recipient_id = $1)
		 ORDER BY m.created_at ASC LIMIT $3 OFFSET $4`,
		userID1, userID2, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %v", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.RecipientID, &msg.Text, &msg.CreatedAt, &msg.SenderName, &msg.RecipientName); err != nil {
			return nil, fmt.Errorf("failed to scan message: %v", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %v", err)
	}

	return messages, nil
} 
