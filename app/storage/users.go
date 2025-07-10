package storage

import (
	"context"
	"fmt"
	"time"
)

// User represents a user in the database
type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never include password in JSON responses
	CreatedAt time.Time `json:"created_at"`
}

// CreateUser creates a new user in the database
func (db *DB) CreateUser(ctx context.Context, firstName, lastName, email, hashedPassword string) (*User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var user User
	err := db.conn.QueryRow(ctx,
		"INSERT INTO users (first_name, last_name, email, password) VALUES ($1, $2, $3, $4) RETURNING id, first_name, last_name, email, created_at",
		firstName, lastName, email, hashedPassword,
	).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email (for authentication)
func (db *DB) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var user User
	err := db.conn.QueryRow(ctx,
		"SELECT id, first_name, last_name, email, password, created_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %v", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(ctx context.Context, id int) (*User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var user User
	err := db.conn.QueryRow(ctx,
		"SELECT id, first_name, last_name, email, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %v", err)
	}

	return &user, nil
}

// GetAllUsers retrieves all users (for contacts list)
func (db *DB) GetAllUsers(ctx context.Context, excludeUserID int) ([]User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.conn.Query(ctx,
		"SELECT id, first_name, last_name, email, created_at FROM users WHERE id != $1 ORDER BY first_name, last_name",
		excludeUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
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

// EmailExists checks if an email already exists in the database
func (db *DB) EmailExists(ctx context.Context, email string) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var count int
	err := db.conn.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %v", err)
	}

	return count > 0, nil
} 