package main

import (
    "net/http"
    "os"
    "api/handlers"
    "api/utils"
)

// corsMiddleware adds CORS headers
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next(w, r)
    }
}

func main() {
	logger := utils.NewLogger()
	logger.Info("Starting API server...")

	// Root API endpoint for frontend
	http.HandleFunc("/api/", corsMiddleware(handlers.Hello))

	// Message handling routes
	http.HandleFunc("/api/messages", corsMiddleware(handlers.GetMessages))
	http.HandleFunc("/api/messages/create", corsMiddleware(handlers.CreateMessage))
	http.HandleFunc("/api/messages/get", corsMiddleware(handlers.GetMessage))
	http.HandleFunc("/api/messages/update", corsMiddleware(handlers.UpdateMessage))
	http.HandleFunc("/api/messages/delete", corsMiddleware(handlers.DeleteMessage))

	// Test route for DB connectivity
	http.HandleFunc("/api/testdb", corsMiddleware(handlers.TestDBHandler))

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Server is running on port " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatal("Server failed to start: " + err.Error())
	}
}
