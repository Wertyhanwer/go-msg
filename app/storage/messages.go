package storage

import (
	"context"
	"fmt"
	"time"
)

// Message represents a message in the database
type Message struct {
	ID        int       `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateMessage creates a new message in the database
func (db *DB) CreateMessage(ctx context.Context, text string) (*Message, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var message Message
	err := db.conn.QueryRow(ctx,
		"INSERT INTO messages (text) VALUES ($1) RETURNING id, text, created_at",
		text,
	).Scan(&message.ID, &message.Text, &message.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create message: %v", err)
	}

	return &message, nil
}

// GetMessage retrieves a message by ID
func (db *DB) GetMessage(ctx context.Context, id int) (*Message, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var message Message
	err := db.conn.QueryRow(ctx,
		"SELECT id, text, created_at FROM messages WHERE id = $1",
		id,
	).Scan(&message.ID, &message.Text, &message.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get message: %v", err)
	}

	return &message, nil
}

// UpdateMessage updates a message by ID
func (db *DB) UpdateMessage(ctx context.Context, id int, text string) (*Message, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var message Message
	err := db.conn.QueryRow(ctx,
		"UPDATE messages SET text = $1 WHERE id = $2 RETURNING id, text, created_at",
		text, id,
	).Scan(&message.ID, &message.Text, &message.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update message: %v", err)
	}

	return &message, nil
}

// DeleteMessage deletes a message by ID
func (db *DB) DeleteMessage(ctx context.Context, id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	result, err := db.conn.Exec(ctx,
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
func (db *DB) GetAllMessages(ctx context.Context, limit, offset int) ([]Message, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.conn.Query(ctx,
		"SELECT id, text, created_at FROM messages ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %v", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Text, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan message: %v", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %v", err)
	}

	return messages, nil
} 