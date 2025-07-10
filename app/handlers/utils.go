package handlers

import (
	"net/http"
	"strconv"
)

// getUserIDFromSession extracts user ID from session cookie
func getUserIDFromSession(r *http.Request) int {
	cookie, err := r.Cookie("session_user_id")
	if err != nil {
		return 0
	}

	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return 0
	}

	return userID
} 