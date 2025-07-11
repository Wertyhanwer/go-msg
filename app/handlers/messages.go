package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"api/storage"
	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все origins для разработки
	},
}

// Структура для WebSocket соединений
type WebSocketConnection struct {
	conn   *websocket.Conn
	userID int
	send   chan []byte
}

// Хаб для управления WebSocket соединениями
type Hub struct {
	connections map[int]*WebSocketConnection
	broadcast   chan []byte
	register    chan *WebSocketConnection
	unregister  chan *WebSocketConnection
	mutex       sync.RWMutex
}

var hub = &Hub{
	connections: make(map[int]*WebSocketConnection),
	broadcast:   make(chan []byte),
	register:    make(chan *WebSocketConnection),
	unregister:  make(chan *WebSocketConnection),
}

// Запуск хаба
func init() {
	go hub.run()
}

func (h *Hub) run() {
	for {
		select {
		case conn := <-h.register:
			h.mutex.Lock()
			h.connections[conn.userID] = conn
			h.mutex.Unlock()
			log.Printf("WebSocket: User %d connected", conn.userID)

		case conn := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.connections[conn.userID]; ok {
				delete(h.connections, conn.userID)
				close(conn.send)
			}
			h.mutex.Unlock()
			log.Printf("WebSocket: User %d disconnected", conn.userID)

		case message := <-h.broadcast:
			h.mutex.RLock()
			for userID, conn := range h.connections {
				select {
				case conn.send <- message:
				default:
					delete(h.connections, userID)
					close(conn.send)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// Отправка сообщения конкретному пользователю
func (h *Hub) sendToUser(userID int, message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	if conn, ok := h.connections[userID]; ok {
		select {
		case conn.send <- message:
		default:
			delete(h.connections, userID)
			close(conn.send)
		}
	}
}

// WebSocket handler
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Проверяем авторизацию
	userID := getUserIDFromSession(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	wsConn := &WebSocketConnection{
		conn:   conn,
		userID: userID,
		send:   make(chan []byte, 256),
	}

	hub.register <- wsConn

	go wsConn.writePump()
	go wsConn.readPump()
}

func (c *WebSocketConnection) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func (c *WebSocketConnection) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

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
		log.Printf("Failed to check friendship between %d and %d: %v", userID, req.RecipientID, err)
		http.Error(w, "Failed to verify friendship", http.StatusInternalServerError)
		return
	}

	if !areFriends {
		log.Printf("Users %d and %d are not friends, blocking message", userID, req.RecipientID)
		http.Error(w, "You can only send messages to your friends", http.StatusForbidden)
		return
	}

	message, err := db.CreateMessage(r.Context(), userID, req.RecipientID, req.Text)
	if err != nil {
		log.Printf("Failed to create message: %v", err)
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	// Отправляем уведомление получателю через WebSocket
	wsMessage := map[string]interface{}{
		"type":    "new_message",
		"message": message,
	}
	if wsData, err := json.Marshal(wsMessage); err == nil {
		hub.sendToUser(req.RecipientID, wsData)
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

// CreateTestMessage handler for creating a test message (simplified version)
func CreateTestMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Text string `json:"text"`
	}
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

	// For testing: create a message from user 1 to user 1 (or use any default user)
	userID := 1
	recipientID := 1

	message, err := db.CreateMessage(r.Context(), userID, recipientID, req.Text)
	if err != nil {
		log.Printf("Failed to create test message: %v", err)
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(message)
}

// CreateTestMessageWithRecipient - временный endpoint для тестирования с указанием получателя
func CreateTestMessageWithRecipient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Text        string `json:"text"`
		RecipientID int    `json:"recipient_id"`
		UserID      int    `json:"user_id,omitempty"`
	}
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

	db, err := storage.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	// Используем переданный user_id или 1 по умолчанию
	userID := req.UserID
	if userID == 0 {
		userID = 1
	}

	message, err := db.CreateMessage(r.Context(), userID, req.RecipientID, req.Text)
	if err != nil {
		log.Printf("Failed to create test message: %v", err)
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	log.Printf("Test message created: ID=%d, UserID=%d, RecipientID=%d, Text=%s", 
		message.ID, message.UserID, message.RecipientID, message.Text)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(message)
}

// GetTestMessages - временный endpoint для получения сообщений без авторизации
func GetTestMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID1Str := r.URL.Query().Get("user1")
	userID2Str := r.URL.Query().Get("user2")
	
	userID1, err1 := strconv.Atoi(userID1Str)
	userID2, err2 := strconv.Atoi(userID2Str)
	
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid user IDs", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
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

	messages, err := db.GetMessagesBetweenUsers(r.Context(), userID1, userID2, limit, offset)
	if err != nil {
		log.Printf("Failed to get test messages: %v", err)
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved %d messages between users %d and %d", len(messages), userID1, userID2)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(messages)
} 