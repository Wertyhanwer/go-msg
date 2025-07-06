package main

import (
    "net/http"
    "os"
    "api/handlers"
    "api/utils"
)

func main() {
	logger := utils.NewLogger()
	logger.Info("Starting API server...")

	// Маршруты для работы с сообщениями
	http.HandleFunc("/api/messages", handlers.GetMessages)
	http.HandleFunc("/api/messages/create", handlers.CreateMessage)
	http.HandleFunc("/api/messages/get", handlers.GetMessage)
	http.HandleFunc("/api/messages/update", handlers.UpdateMessage)
	http.HandleFunc("/api/messages/delete", handlers.DeleteMessage)

	// Тестовый маршрут для проверки БД
	http.HandleFunc("/api/testdb", handlers.TestDBHandler)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Server is running on port " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatal("Server failed to start: " + err.Error())
	}
}
