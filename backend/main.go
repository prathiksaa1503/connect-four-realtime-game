package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
)

var (
	eventProducer *EventProducer
)

func main() {
	// Initialize event producer (simulated Kafka)
	eventProducer = NewEventProducer(1000)

	// Start analytics consumer
	go startAnalyticsConsumer()

	// Initialize connection manager
	connManager := NewConnectionManager()
	go connManager.run()

	// HTTP routes
	http.HandleFunc("/ws", serveWS(connManager))
	http.HandleFunc("/leaderboard", handleLeaderboardHTTP)
	http.HandleFunc("/health", handleHealth)

	// Serve frontend
	// Get the frontend directory path (relative to backend directory)
	frontendPath := filepath.Join("..", "frontend")
	fs := http.FileServer(http.Dir(frontendPath))
	http.Handle("/", fs)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleLeaderboardHTTP handles HTTP requests for leaderboard
func handleLeaderboardHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	leaderboard := gameManager.GetLeaderboard()
	json.NewEncoder(w).Encode(leaderboard)
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
