package main

import (
	"context"
	"fmt"
	"log"
	"api/storage"
)

func main() {
	// Get database connection
	db, err := storage.GetDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test CheckFriendship function
	fmt.Println("Testing CheckFriendship function...")
	
	// Test different user combinations
	testCases := []struct {
		user1, user2 int
		description  string
	}{
		{1, 2, "User 1 and User 2"},
		{2, 1, "User 2 and User 1 (reverse)"},
		{1, 4, "User 1 and User 4"},
		{4, 1, "User 4 and User 1 (reverse)"},
		{1, 3, "User 1 and User 3 (should not be friends)"},
		{2, 4, "User 2 and User 4 (should not be friends)"},
	}
	
	ctx := context.Background()
	
	for _, tc := range testCases {
		areFriends, err := db.CheckFriendship(ctx, tc.user1, tc.user2)
		if err != nil {
			fmt.Printf("ERROR: %s - %v\n", tc.description, err)
		} else {
			fmt.Printf("✓ %s: %v\n", tc.description, areFriends)
		}
	}
} 