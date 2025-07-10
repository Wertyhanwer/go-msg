package storage

import (
	"context"
	"fmt"
	"time"
)

// Friend represents a friend relationship in the database
type Friend struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	FriendID  int       `json:"friend_id"`
	Status    string    `json:"status"` // pending, accepted, blocked
	CreatedAt time.Time `json:"created_at"`
	// Populated fields for API responses
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
}

// SendFriendRequest sends a friend request to another user
func (db *DB) SendFriendRequest(ctx context.Context, userID, friendID int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check if relationship already exists
	var count int
	err := db.conn.QueryRow(ctx,
		"SELECT COUNT(*) FROM friends WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)",
		userID, friendID,
	).Scan(&count)
	
	if err != nil {
		return fmt.Errorf("failed to check existing friendship: %v", err)
	}
	
	if count > 0 {
		return fmt.Errorf("friendship already exists")
	}

	// Create friend request
	_, err = db.conn.Exec(ctx,
		"INSERT INTO friends (user_id, friend_id, status) VALUES ($1, $2, 'pending')",
		userID, friendID,
	)

	if err != nil {
		return fmt.Errorf("failed to send friend request: %v", err)
	}

	return nil
}

// AcceptFriendRequest accepts a friend request
func (db *DB) AcceptFriendRequest(ctx context.Context, userID, friendID int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Update the existing request
	_, err := db.conn.Exec(ctx,
		"UPDATE friends SET status = 'accepted' WHERE user_id = $1 AND friend_id = $2 AND status = 'pending'",
		friendID, userID,
	)

	if err != nil {
		return fmt.Errorf("failed to accept friend request: %v", err)
	}

	// Create reverse relationship for easy querying
	_, err = db.conn.Exec(ctx,
		"INSERT INTO friends (user_id, friend_id, status) VALUES ($1, $2, 'accepted') ON CONFLICT (user_id, friend_id) DO NOTHING",
		userID, friendID,
	)

	return err
}

// GetFriends returns all friends of a user
func (db *DB) GetFriends(ctx context.Context, userID int) ([]Friend, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.conn.Query(ctx,
		`SELECT f.id, f.user_id, f.friend_id, f.status, f.created_at,
		 u.first_name, u.last_name, u.email
		 FROM friends f
		 JOIN users u ON f.friend_id = u.id
		 WHERE f.user_id = $1 AND f.status = 'accepted'
		 ORDER BY u.first_name, u.last_name`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get friends: %v", err)
	}
	defer rows.Close()

	var friends []Friend
	for rows.Next() {
		var friend Friend
		if err := rows.Scan(&friend.ID, &friend.UserID, &friend.FriendID, &friend.Status, &friend.CreatedAt, &friend.FirstName, &friend.LastName, &friend.Email); err != nil {
			return nil, fmt.Errorf("failed to scan friend: %v", err)
		}
		friends = append(friends, friend)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating friends: %v", err)
	}

	return friends, nil
}

// GetPendingRequests returns pending friend requests for a user
func (db *DB) GetPendingRequests(ctx context.Context, userID int) ([]Friend, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.conn.Query(ctx,
		`SELECT f.id, f.user_id, f.friend_id, f.status, f.created_at,
		 u.first_name, u.last_name, u.email
		 FROM friends f
		 JOIN users u ON f.user_id = u.id
		 WHERE f.friend_id = $1 AND f.status = 'pending'
		 ORDER BY f.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending requests: %v", err)
	}
	defer rows.Close()

	var requests []Friend
	for rows.Next() {
		var request Friend
		if err := rows.Scan(&request.ID, &request.UserID, &request.FriendID, &request.Status, &request.CreatedAt, &request.FirstName, &request.LastName, &request.Email); err != nil {
			return nil, fmt.Errorf("failed to scan request: %v", err)
		}
		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating requests: %v", err)
	}

	return requests, nil
}

// SearchUsers searches for users by name or email
func (db *DB) SearchUsers(ctx context.Context, query string, excludeUserID int, limit int) ([]User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	searchQuery := "%" + query + "%"
	
	rows, err := db.conn.Query(ctx,
		`SELECT id, first_name, last_name, email, created_at 
		 FROM users 
		 WHERE id != $1 AND (
		 	first_name ILIKE $2 OR 
		 	last_name ILIKE $2 OR 
		 	email ILIKE $2 OR
		 	(first_name || ' ' || last_name) ILIKE $2
		 )
		 ORDER BY first_name, last_name
		 LIMIT $3`,
		excludeUserID, searchQuery, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %v", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %v", err)
	}

	return users, nil
}

// GetFriendshipStatus returns the friendship status between two users
func (db *DB) GetFriendshipStatus(ctx context.Context, userID, friendID int) (string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var status string
	err := db.conn.QueryRow(ctx,
		"SELECT status FROM friends WHERE user_id = $1 AND friend_id = $2",
		userID, friendID,
	).Scan(&status)

	if err != nil {
		// Check reverse relationship
		err = db.conn.QueryRow(ctx,
			"SELECT status FROM friends WHERE user_id = $1 AND friend_id = $2",
			friendID, userID,
		).Scan(&status)
		
		if err != nil {
			return "none", nil // No relationship exists
		}
		
		// If there's a pending request from the other user
		if status == "pending" {
			return "pending_received", nil
		}
	}

	return status, nil
}

// RejectFriendRequest rejects or removes a friend request
func (db *DB) RejectFriendRequest(ctx context.Context, userID, friendID int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.conn.Exec(ctx,
		"DELETE FROM friends WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)",
		userID, friendID,
	)

	if err != nil {
		return fmt.Errorf("failed to reject friend request: %v", err)
	}

	return nil
} 