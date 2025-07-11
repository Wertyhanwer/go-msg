package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
	"api/storage"
)

// FriendRequest represents a friend request
type FriendRequest struct {
	FriendID int `json:"friend_id"`
}

// SearchUsers handler for searching users
func SearchUsers(w http.ResponseWriter, r *http.Request) {
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

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20 // default value

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	users, err := db.SearchUsers(r.Context(), query, userID, limit)
	if err != nil {
		log.Printf("Failed to search users: %v", err)
		http.Error(w, "Failed to search users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(users)
}

// SendFriendRequest handler for sending friend requests
func SendFriendRequest(w http.ResponseWriter, r *http.Request) {
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

	var req FriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FriendID == 0 {
		http.Error(w, "Friend ID is required", http.StatusBadRequest)
		return
	}

	if req.FriendID == userID {
		http.Error(w, "Cannot send friend request to yourself", http.StatusBadRequest)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	err = db.SendFriendRequest(r.Context(), userID, req.FriendID)
	if err != nil {
		log.Printf("Failed to send friend request: %v", err)
		http.Error(w, "Failed to send friend request", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Заявка в друзья отправлена",
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

// AcceptFriendRequest handler for accepting friend requests
func AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
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

	var req FriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FriendID == 0 {
		http.Error(w, "Friend ID is required", http.StatusBadRequest)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	err = db.AcceptFriendRequest(r.Context(), userID, req.FriendID)
	if err != nil {
		log.Printf("Failed to accept friend request: %v", err)
		http.Error(w, "Failed to accept friend request", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Заявка в друзья принята",
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

// RejectFriendRequest handler for rejecting friend requests
func RejectFriendRequest(w http.ResponseWriter, r *http.Request) {
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

	var req FriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FriendID == 0 {
		http.Error(w, "Friend ID is required", http.StatusBadRequest)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	err = db.RejectFriendRequest(r.Context(), userID, req.FriendID)
	if err != nil {
		log.Printf("Failed to reject friend request: %v", err)
		http.Error(w, "Failed to reject friend request", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Заявка в друзья отклонена",
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

// GetFriends handler for getting user's friends
func GetFriends(w http.ResponseWriter, r *http.Request) {
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

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	friends, err := db.GetFriends(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get friends: %v", err)
		http.Error(w, "Failed to get friends", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(friends)
}

// GetPendingRequests handler for getting pending friend requests
func GetPendingRequests(w http.ResponseWriter, r *http.Request) {
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

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	requests, err := db.GetPendingRequests(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get pending requests: %v", err)
		http.Error(w, "Failed to get pending requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(requests)
}

// GetFriendshipStatus handler for checking friendship status between users
func GetFriendshipStatus(w http.ResponseWriter, r *http.Request) {
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

	friendIDStr := r.URL.Query().Get("friend_id")
	friendID, err := strconv.Atoi(friendIDStr)
	if err != nil {
		http.Error(w, "Invalid friend ID", http.StatusBadRequest)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	status, err := db.GetFriendshipStatus(r.Context(), userID, friendID)
	if err != nil {
		log.Printf("Failed to get friendship status: %v", err)
		http.Error(w, "Failed to get friendship status", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"friend_id": friendID,
		"status":    status,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

// DebugFriends - временная функция для отладки дружбы
func DebugFriends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	// Получить все записи из таблицы friends
	rows, err := db.GetPool().Query(r.Context(),
		"SELECT id, user_id, friend_id, status, created_at FROM friends ORDER BY created_at DESC")
	if err != nil {
		log.Printf("Failed to debug friends: %v", err)
		http.Error(w, "Failed to get friends data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var friendsData []map[string]interface{}
	for rows.Next() {
		var id, userID, friendID int
		var status string
		var createdAt time.Time
		
		if err := rows.Scan(&id, &userID, &friendID, &status, &createdAt); err != nil {
			log.Printf("Failed to scan friend row: %v", err)
			continue
		}
		
		friendsData = append(friendsData, map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"friend_id":  friendID,
			"status":     status,
			"created_at": createdAt,
		})
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"friends_count": len(friendsData),
		"friends":       friendsData,
	})
} 