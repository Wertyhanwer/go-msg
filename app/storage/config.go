package storage


import (
    "fmt"
    "os"
    "strconv"
    "log"
)

type PostgresConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
}

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
