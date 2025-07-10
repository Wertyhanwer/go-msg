package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Message структура для хранения сообщений в памяти
type Message struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	RecipientID int       `json:"recipient_id"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"created_at"`
	SenderName  string    `json:"sender_name,omitempty"`
}

// In-memory storage
var messages []Message
var messageID = 1

// CORS middleware
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

// API endpoints
func helloHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"message": "Simple Chat API v1.0",
		"status":  "running",
		"time":    time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID      int    `json:"user_id"`
		RecipientID int    `json:"recipient_id"`
		Text        string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	// Устанавливаем значения по умолчанию
	if req.UserID == 0 {
		req.UserID = 1
	}
	if req.RecipientID == 0 {
		req.RecipientID = 1
	}

	// Создаем сообщение
	message := Message{
		ID:          messageID,
		UserID:      req.UserID,
		RecipientID: req.RecipientID,
		Text:        req.Text,
		CreatedAt:   time.Now(),
		SenderName:  fmt.Sprintf("User %d", req.UserID),
	}

	messageID++
	messages = append(messages, message)

	log.Printf("Message created: ID=%d, UserID=%d, RecipientID=%d, Text=%s", 
		message.ID, message.UserID, message.RecipientID, message.Text)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user1 := r.URL.Query().Get("user1")
	user2 := r.URL.Query().Get("user2")

	var filteredMessages []Message

	if user1 != "" && user2 != "" {
		// Фильтруем сообщения между двумя пользователями
		for _, msg := range messages {
			if (fmt.Sprintf("%d", msg.UserID) == user1 && fmt.Sprintf("%d", msg.RecipientID) == user2) ||
				(fmt.Sprintf("%d", msg.UserID) == user2 && fmt.Sprintf("%d", msg.RecipientID) == user1) {
				filteredMessages = append(filteredMessages, msg)
			}
		}
	} else {
		// Возвращаем все сообщения
		filteredMessages = messages
	}

	log.Printf("Retrieved %d messages", len(filteredMessages))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filteredMessages)
}

func clearMessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	messages = []Message{}
	messageID = 1
	
	log.Println("All messages cleared")
	
	response := map[string]string{"status": "cleared"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	log.Println("Starting Simple Chat Server...")

	// Routes
	http.HandleFunc("/api/", corsMiddleware(helloHandler))
	http.HandleFunc("/api/messages/create", corsMiddleware(createMessageHandler))
	http.HandleFunc("/api/messages/get", corsMiddleware(getMessagesHandler))
	http.HandleFunc("/api/messages/clear", corsMiddleware(clearMessagesHandler))

	// Static files
	http.Handle("/", http.FileServer(http.Dir("../")))

	port := "8080"
	log.Printf("Server running on http://localhost:%s", port)
	log.Printf("API test page: http://localhost:%s/test-message.html", port)
	log.Printf("Main chat: http://localhost:%s/index.html", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
} 