package main

import (
	"log"
	"sync"
	"time"
)

// AnalyticsData holds analytics information
type AnalyticsData struct {
	TotalGames        int
	WinsPerPlayer     map[string]int
	TotalGameDuration time.Duration
	GameCount         int
	mu                sync.RWMutex
}

var analyticsData = &AnalyticsData{
	WinsPerPlayer: make(map[string]int),
}

// startAnalyticsConsumer starts the analytics consumer that processes events
func startAnalyticsConsumer() {
	log.Println("Analytics consumer started")

	gameStartTimes := make(map[string]time.Time)

	for event := range eventProducer.GetEventChannel() {
		switch event.Type {
		case "GAME_STARTED":
			analyticsData.mu.Lock()
			analyticsData.TotalGames++
			analyticsData.GameCount++
			gameStartTimes[event.GameID] = event.Timestamp
			analyticsData.mu.Unlock()
			log.Printf("Analytics: Game %s started between %s and %s", event.GameID, event.Player1, event.Player2)

		case "MOVE_MADE":
			log.Printf("Analytics: Move made in game %s by %s in column %d", event.GameID, event.Player, event.Column)

		case "GAME_ENDED":
			analyticsData.mu.Lock()
			if startTime, exists := gameStartTimes[event.GameID]; exists {
				duration := event.Timestamp.Sub(startTime)
				analyticsData.TotalGameDuration += duration
				delete(gameStartTimes, event.GameID)
			}

			if event.Winner != "" {
				analyticsData.WinsPerPlayer[event.Winner]++
			}
			analyticsData.mu.Unlock()

			log.Printf("Analytics: Game %s ended. Winner: %s, Draw: %v", event.GameID, event.Winner, event.IsDraw)
		}
	}
}

// GetAnalytics returns the current analytics data
func GetAnalytics() *AnalyticsData {
	analyticsData.mu.RLock()
	defer analyticsData.mu.RUnlock()

	// Create a copy to avoid race conditions
	copy := &AnalyticsData{
		TotalGames:        analyticsData.TotalGames,
		WinsPerPlayer:     make(map[string]int),
		TotalGameDuration: analyticsData.TotalGameDuration,
		GameCount:         analyticsData.GameCount,
	}

	for k, v := range analyticsData.WinsPerPlayer {
		copy.WinsPerPlayer[k] = v
	}

	return copy
}

// GetAverageGameDuration returns the average game duration
func (ad *AnalyticsData) GetAverageGameDuration() time.Duration {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	if ad.GameCount == 0 {
		return 0
	}

	return ad.TotalGameDuration / time.Duration(ad.GameCount)
}
