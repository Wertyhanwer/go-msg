package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Messenger backend is up!"})
	})

	fs := http.FileServer(http.Dir("web"))
	http.Handle("/", fs)

	fmt.Println("Server running at http://localhost:30000")
	err := http.ListenAndServe(":30000", nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
