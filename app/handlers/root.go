package handlers

import (
    "encoding/json"
    "net/http"
    "log"
)

func Hello(w http.ResponseWriter, r *http.Request) {
    log.Printf("Received request to %s", r.URL.Path)
    response := map[string]string{"message": "Hello, golang!"}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
