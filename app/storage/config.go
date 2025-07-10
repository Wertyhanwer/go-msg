package storage

import (
    "fmt"
    "os"
    "strconv"
    "log"
    "context"
    "sync"

    "github.com/jackc/pgx/v5"
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
    conn *pgx.Conn
    mu   sync.RWMutex
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
        
        conn, err := pgx.Connect(context.Background(), connStr)
        if err != nil {
            initErr = fmt.Errorf("unable to connect to database: %v", err)
            return
        }

        db = &DB{conn: conn}
        
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

// Close closes the database connection
func (db *DB) Close(ctx context.Context) error {
    db.mu.Lock()
    defer db.mu.Unlock()
    
    if db.conn != nil {
        return db.conn.Close(ctx)
    }
    return nil
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
    
    createMessagesTableSQL := `
    CREATE TABLE IF NOT EXISTS messages (
        id SERIAL PRIMARY KEY,
        user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
        recipient_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
        text TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
    
    // Create users table first
    _, err := db.conn.Exec(context.Background(), createUsersTableSQL)
    if err != nil {
        return err
    }
    
    // Then create messages table with foreign keys
    _, err = db.conn.Exec(context.Background(), createMessagesTableSQL)
    return err
}

// GetMessagesCount returns the total number of messages
func (db *DB) GetMessagesCount(ctx context.Context) (int, error) {
    db.mu.RLock()
    defer db.mu.RUnlock()

    var count int
    err := db.conn.QueryRow(ctx, "SELECT COUNT(*) FROM messages").Scan(&count)
    if err != nil {
        return 0, fmt.Errorf("failed to get messages count: %v", err)
    }
    return count, nil
}
