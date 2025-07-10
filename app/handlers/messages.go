package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"api/storage"
)

// MessageRequest represents a request for creating/updating a message
type MessageRequest struct {
	RecipientID int    `json:"recipient_id"`
	Text        string `json:"text"`
}

// CreateMessage handler for creating a new message
func CreateMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	userID := getUserIDFromSession(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		http.Error(w, "Message text is required", http.StatusBadRequest)
		return
	}

	if req.RecipientID == 0 {
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	if req.RecipientID == userID {
		http.Error(w, "Cannot send message to yourself", http.StatusBadRequest)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	// Check if users are friends before allowing message creation
	areFriends, err := db.CheckFriendship(r.Context(), userID, req.RecipientID)
	if err != nil {
		log.Printf("Failed to check friendship: %v", err)
		http.Error(w, "Failed to verify friendship", http.StatusInternalServerError)
		return
	}

	if !areFriends {
		http.Error(w, "You can only send messages to your friends", http.StatusForbidden)
		return
	}

	message, err := db.CreateMessage(r.Context(), userID, req.RecipientID, req.Text)
	if err != nil {
		log.Printf("Failed to create message: %v", err)
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(message)
}

// GetMessage handler for getting a message by ID
func GetMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	userID := getUserIDFromSession(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	message, err := db.GetMessage(r.Context(), id)
	if err != nil {
		log.Printf("Failed to get message: %v", err)
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Check if user has access to this message (sender or recipient)
	if message.UserID != userID && message.RecipientID != userID {
		http.Error(w, "Access denied to this message", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(message)
}

// UpdateMessage handler for updating a message
func UpdateMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	userID := getUserIDFromSession(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		http.Error(w, "Message text is required", http.StatusBadRequest)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	// First, get the message to check ownership
	existingMessage, err := db.GetMessage(r.Context(), id)
	if err != nil {
		log.Printf("Failed to get message for update: %v", err)
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Check if user is the author of the message
	if existingMessage.UserID != userID {
		http.Error(w, "Only message author can update the message", http.StatusForbidden)
		return
	}

	message, err := db.UpdateMessage(r.Context(), id, req.Text)
	if err != nil {
		log.Printf("Failed to update message: %v", err)
		http.Error(w, "Failed to update message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(message)
}

// DeleteMessage handler for deleting a message
func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	userID := getUserIDFromSession(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	// First, get the message to check ownership
	existingMessage, err := db.GetMessage(r.Context(), id)
	if err != nil {
		log.Printf("Failed to get message for deletion: %v", err)
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Check if user is the author of the message
	if existingMessage.UserID != userID {
		http.Error(w, "Only message author can delete the message", http.StatusForbidden)
		return
	}

	if err := db.DeleteMessage(r.Context(), id); err != nil {
		log.Printf("Failed to delete message: %v", err)
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetMessages handler for getting list of messages
func GetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	userID := getUserIDFromSession(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default value
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	messages, err := db.GetAllMessages(r.Context(), limit, offset)
	if err != nil {
		log.Printf("Failed to get messages: %v", err)
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(messages)
}

// TestDBHandler checks database connection
func TestDBHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	count, err := db.GetMessagesCount(r.Context())
	if err != nil {
		http.Error(w, "Failed to get messages count", http.StatusInternalServerError)
		return
	}

	response := map[string]int{
		"messages_count": count,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

// GetMessagesBetweenUsers handler for getting messages between two users
func GetMessagesBetweenUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	userID := getUserIDFromSession(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	recipientIDStr := r.URL.Query().Get("recipient_id")
	recipientID, err := strconv.Atoi(recipientIDStr)
	if err != nil {
		http.Error(w, "Invalid recipient ID", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default value
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	messages, err := db.GetMessagesBetweenUsers(r.Context(), userID, recipientID, limit, offset)
	if err != nil {
		log.Printf("Failed to get messages: %v", err)
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(messages)
} 