package main

import (
    "net/http"
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "api/handlers"
    "api/storage"
)

func main() {
	log.Println("= = = = = = = = = = = = = = = = = = = = = = = ")
	log.Println("Start main.go::main")

	// Initialize database connection
	db, err := storage.GetDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup handlers - более специфичные маршруты должны быть первыми
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testdb", handlers.TestDB)
	mux.HandleFunc("/", handlers.Hello)

	// Create server with timeout configurations
	server := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Close database connection
	if err := db.Close(ctx); err != nil {
		log.Printf("Database connection close error: %v", err)
	}

	log.Println("Server stopped gracefully")
	log.Println("= = = = = = = = = = = = = = = = = = = = = = = ")
}
