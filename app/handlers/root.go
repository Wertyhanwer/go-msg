package handlers

import (
    "encoding/json"
    "net/http"
    "log"
    "api/storage"
)

func Hello(w http.ResponseWriter, r *http.Request) {
    log.Printf("Received request to %s", r.URL.Path)
    response := map[string]string{"message": "Hello, golang!"}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func TestDB(w http.ResponseWriter, r *http.Request) {
    log.Printf("Received request to TestDB handler: %s", r.URL.Path)
    
    db, err := storage.GetDB()
    if err != nil {
        log.Printf("Failed to get DB connection: %v", err)
        http.Error(w, "Database connection error", http.StatusInternalServerError)
        return
    }

    count, err := db.GetMessagesCount(r.Context())
    if err != nil {
        log.Printf("Failed to get messages count: %v", err)
        http.Error(w, "Database query error", http.StatusInternalServerError)
        return
    }

    log.Printf("Successfully got messages count: %d", count)
    response := map[string]interface{}{"messages_count": count}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
