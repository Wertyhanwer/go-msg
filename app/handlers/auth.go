package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"api/storage"
	"api/models"
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	User    *storage.User  `json:"user,omitempty"`
}

// Register handler for user registration
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Password == "" {
		response := AuthResponse{
			Success: false,
			Message: "Все поля обязательны для заполнения",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate email format (simple validation)
	if !strings.Contains(req.Email, "@") {
		response := AuthResponse{
			Success: false,
			Message: "Некорректный формат email",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate password length
	if len(req.Password) < 6 {
		response := AuthResponse{
			Success: false,
			Message: "Пароль должен содержать минимум 6 символов",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(response)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	// Check if email already exists
	exists, err := db.EmailExists(r.Context(), req.Email)
	if err != nil {
		log.Printf("Failed to check email existence: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		response := AuthResponse{
			Success: false,
			Message: "Пользователь с таким email уже существует",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Hash password
	hashedPassword, err := models.HashPassword(req.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := db.CreateUser(r.Context(), req.FirstName, req.LastName, req.Email, hashedPassword)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	setSessionCookie(w, user.ID)

	response := AuthResponse{
		Success: true,
		Message: "Пользователь успешно зарегистрирован",
		User:    user,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

// Login handler for user authentication
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		response := AuthResponse{
			Success: false,
			Message: "Email и пароль обязательны",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(response)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	// Get user by email
	user, err := db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		response := AuthResponse{
			Success: false,
			Message: "Неверный email или пароль",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create a User model instance to check password
	userModel := &models.User{Password: user.Password}
	if !userModel.CheckPassword(req.Password) {
		response := AuthResponse{
			Success: false,
			Message: "Неверный email или пароль",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Set session cookie
	setSessionCookie(w, user.ID)

	response := AuthResponse{
		Success: true,
		Message: "Вход выполнен успешно",
		User:    user,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

// Logout handler for user logout
func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user_id",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
	})

	response := AuthResponse{
		Success: true,
		Message: "Выход выполнен успешно",
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

// GetCurrentUser returns the current authenticated user
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromSession(r)
	if userID == 0 {
		response := AuthResponse{
			Success: false,
			Message: "Пользователь не аутентифицирован",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(response)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	user, err := db.GetUserByID(r.Context(), userID)
	if err != nil {
		response := AuthResponse{
			Success: false,
			Message: "Пользователь не найден",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := AuthResponse{
		Success: true,
		Message: "Пользователь найден",
		User:    user,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

// GetUsers returns all users except current user (for contacts)
func GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromSession(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db, err := storage.GetDB()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	users, err := db.GetAllUsers(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get users: %v", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(users)
}

// Helper functions

// setSessionCookie sets a session cookie with user ID
func setSessionCookie(w http.ResponseWriter, userID int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user_id",
		Value:    strconv.Itoa(userID),
		Expires:  time.Now().Add(24 * time.Hour), // 24 hours
		HttpOnly: true,
		Path:     "/",
	})
}



// requireAuth middleware to check if user is authenticated
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := getUserIDFromSession(r)
		if userID == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
} 