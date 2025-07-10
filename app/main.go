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

	// Authentication routes
	http.HandleFunc("/api/auth/register", corsMiddleware(handlers.Register))
	http.HandleFunc("/api/auth/login", corsMiddleware(handlers.Login))
	http.HandleFunc("/api/auth/logout", corsMiddleware(handlers.Logout))
	http.HandleFunc("/api/auth/user", corsMiddleware(handlers.GetCurrentUser))
	http.HandleFunc("/api/users", corsMiddleware(handlers.GetUsers))
	
	// Friends routes
	http.HandleFunc("/api/friends/search", corsMiddleware(handlers.SearchUsers))
	http.HandleFunc("/api/friends/request", corsMiddleware(handlers.SendFriendRequest))
	http.HandleFunc("/api/friends/accept", corsMiddleware(handlers.AcceptFriendRequest))
	http.HandleFunc("/api/friends/reject", corsMiddleware(handlers.RejectFriendRequest))
	http.HandleFunc("/api/friends", corsMiddleware(handlers.GetFriends))
	http.HandleFunc("/api/friends/pending", corsMiddleware(handlers.GetPendingRequests))
	http.HandleFunc("/api/friends/status", corsMiddleware(handlers.GetFriendshipStatus))

	// Message handling routes
	http.HandleFunc("/api/messages", corsMiddleware(handlers.GetMessages))
	http.HandleFunc("/api/messages/create", corsMiddleware(handlers.CreateMessage))
	http.HandleFunc("/api/messages/test", corsMiddleware(handlers.CreateTestMessage))
	http.HandleFunc("/api/messages/test-with-recipient", corsMiddleware(handlers.CreateTestMessageWithRecipient))
	http.HandleFunc("/api/messages/test-get", corsMiddleware(handlers.GetTestMessages))
	http.HandleFunc("/api/messages/get", corsMiddleware(handlers.GetMessage))
	http.HandleFunc("/api/messages/update", corsMiddleware(handlers.UpdateMessage))
	http.HandleFunc("/api/messages/delete", corsMiddleware(handlers.DeleteMessage))
	http.HandleFunc("/api/messages/between", corsMiddleware(handlers.GetMessagesBetweenUsers))

	// WebSocket endpoint for real-time messaging
	http.HandleFunc("/api/ws", corsMiddleware(handlers.HandleWebSocket))

	// Test route for DB connectivity
	http.HandleFunc("/api/testdb", corsMiddleware(handlers.TestDBHandler))

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("API server is running on port " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatal("Server failed to start: " + err.Error())
	}
}
