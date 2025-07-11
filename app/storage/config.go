package storage

import (
    "fmt"
    "os"
    "strconv"
    "log"
    "context"
    "sync"

    "github.com/jackc/pgx/v5/pgxpool"
    "golang.org/x/crypto/bcrypt"
)

type PostgresConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
}

// DB represents a database instance
type DB struct {
    pool *pgxpool.Pool
}

var (
    db   *DB
    once sync.Once
)

func NewPostgresConfig() *PostgresConfig {
    portStr := os.Getenv("POSTGRES_PORT")
    port, err := strconv.Atoi(portStr)
    if err != nil {
        log.Printf("Invalid POSTGRES_PORT value: %v, defaulting to 5432", err)
        port = 5432
    }

    return &PostgresConfig{
        Host:     os.Getenv("POSTGRES_HOST"),
        Port:     port,
        User:     os.Getenv("POSTGRES_USER"),
        Password: os.Getenv("POSTGRES_PASSWORD"),
        DBName:   os.Getenv("POSTGRES_DB"),
    }
}

func (cfg *PostgresConfig) ConnString() string {
    return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
}

// GetDB returns a singleton database instance
func GetDB() (*DB, error) {
    var initErr error
    once.Do(func() {
        cfg := NewPostgresConfig()
        connStr := cfg.ConnString()
        
        pool, err := pgxpool.New(context.Background(), connStr)
        if err != nil {
            initErr = fmt.Errorf("unable to create connection pool: %v", err)
            return
        }

        db = &DB{pool: pool}
        
        // Initialize the database schema
        if err := db.initSchema(); err != nil {
            initErr = fmt.Errorf("failed to initialize schema: %v", err)
            return
        }
    })

    if initErr != nil {
        return nil, initErr
    }

    return db, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
    if db.pool != nil {
        db.pool.Close()
    }
}

// GetPool returns the underlying connection pool
func (db *DB) GetPool() *pgxpool.Pool {
    return db.pool
}

// initSchema initializes the database schema
func (db *DB) initSchema() error {
    createUsersTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        first_name VARCHAR(100) NOT NULL,
        last_name VARCHAR(100) NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        password VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
    
    createFriendsTableSQL := `
    CREATE TABLE IF NOT EXISTS friends (
        id SERIAL PRIMARY KEY,
        user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
        friend_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
        status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, blocked
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        UNIQUE(user_id, friend_id)
    );`
    
    createMessagesTableSQL := `
    CREATE TABLE IF NOT EXISTS messages (
        id SERIAL PRIMARY KEY,
        user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
        recipient_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
        text TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
    
    // Create users table first
    _, err := db.pool.Exec(context.Background(), createUsersTableSQL)
    if err != nil {
        return err
    }
    
    // Create friends table
    _, err = db.pool.Exec(context.Background(), createFriendsTableSQL)
    if err != nil {
        return err
    }
    
    // Then create messages table with foreign keys
    _, err = db.pool.Exec(context.Background(), createMessagesTableSQL)
    if err != nil {
        return err
    }
    
    // Insert test user if not exists
    err = db.createTestUser()
    return err
}

// GetMessagesCount returns the total number of messages
func (db *DB) GetMessagesCount(ctx context.Context) (int, error) {
    var count int
    err := db.pool.QueryRow(ctx, "SELECT COUNT(*) FROM messages").Scan(&count)
    if err != nil {
        return 0, fmt.Errorf("failed to get messages count: %v", err)
    }
    return count, nil
}

// createTestUser creates a test user for demo purposes
func (db *DB) createTestUser() error {
    ctx := context.Background()
    
    // Check if test user already exists
    var count int
    err := db.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", "test@example.com").Scan(&count)
    if err != nil {
        return fmt.Errorf("failed to check test user existence: %v", err)
    }
    
    if count > 0 {
        return nil // Test user already exists
    }
    
    // Use bcrypt to hash the password for consistency
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("failed to hash test user password: %v", err)
    }
    
    _, err = db.pool.Exec(ctx,
        "INSERT INTO users (first_name, last_name, email, password) VALUES ($1, $2, $3, $4)",
        "Тест", "Юзер", "test@example.com", string(hashedPassword),
    )
    
    if err != nil {
        return fmt.Errorf("failed to create test user: %v", err)
    }
    
    return nil
}

