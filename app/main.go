package main

import (
    "net/http"
    "context"
    "log"

    "api/handlers"
    "api/storage"

    "github.com/jackc/pgx/v5"
)


func main() {
	log.Println("= = = = = = = = = = = = = = = = = = = = = = = ")
	log.Println("Start main.go::main")

	// - - - - - - - - - root - - - - - - - - -
   	http.HandleFunc("/", handlers.Hello)

	// - - - - - get connection string - - - - -
   	cfg := storage.NewPostgresConfig()
   	connStr := cfg.ConnString()

	// - - - - - - test connect bd - - - - - - -
    conn, err := pgx.Connect(context.Background(), connStr)
    if err != nil {
        log.Fatalf("UNABLE to connect to database '%s': %v\n", connStr, err)
    } else {
    	log.Printf("SUCCESSFUL to connect to database: %s", connStr)
    }
    defer conn.Close(context.Background())
    // - - - - - - - - - - - - - - - - - - - - -

    http.ListenAndServe("0.0.0.0:8080", nil)
    log.Println("= = = = = = = = = = = = = = = = = = = = = = = ")
}
