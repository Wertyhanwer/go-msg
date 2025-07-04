package handlers

import (
    "encoding/json"
    "net/http"
)


func Hello(w http.ResponseWriter, r *http.Request) {
    response := map[string]string{"message": "Hello, golang!"}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
