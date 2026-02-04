package analytics

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

// AnalyticsConsumer processes game events and calculates analytics
type AnalyticsConsumer struct {
	data          *AnalyticsData
	gameStartTimes map[string]time.Time
}

// NewAnalyticsConsumer creates a new analytics consumer
func NewAnalyticsConsumer() *AnalyticsConsumer {
	return &AnalyticsConsumer{
		data: &AnalyticsData{
			WinsPerPlayer: make(map[string]int),
		},
		gameStartTimes: make(map[string]time.Time),
	}
}

// Start starts consuming events from the producer
func (ac *AnalyticsConsumer) Start(producer *EventProducer) {
	log.Println("Analytics consumer started")

	for event := range producer.GetEventChannel() {
		ac.processEvent(event)
	}
}

// processEvent processes a single event
func (ac *AnalyticsConsumer) processEvent(event Event) {
	switch event.Type {
	case "GAME_STARTED":
		ac.data.mu.Lock()
		ac.data.TotalGames++
		ac.data.GameCount++
		ac.gameStartTimes[event.GameID] = event.Timestamp
		ac.data.mu.Unlock()
		log.Printf("Analytics: Game %s started between %s and %s", event.GameID, event.Player1, event.Player2)

	case "MOVE_MADE":
		log.Printf("Analytics: Move made in game %s by %s in column %d", event.GameID, event.Player, event.Column)

	case "GAME_ENDED":
		ac.data.mu.Lock()
		if startTime, exists := ac.gameStartTimes[event.GameID]; exists {
			duration := event.Timestamp.Sub(startTime)
			ac.data.TotalGameDuration += duration
			delete(ac.gameStartTimes, event.GameID)
		}

		if event.Winner != "" {
			ac.data.WinsPerPlayer[event.Winner]++
		}
		ac.data.mu.Unlock()

		log.Printf("Analytics: Game %s ended. Winner: %s, Draw: %v", event.GameID, event.Winner, event.IsDraw)
	}
}

// GetAnalytics returns the current analytics data
func (ac *AnalyticsConsumer) GetAnalytics() *AnalyticsData {
	ac.data.mu.RLock()
	defer ac.data.mu.RUnlock()

	// Create a copy to avoid race conditions
	copy := &AnalyticsData{
		TotalGames:        ac.data.TotalGames,
		WinsPerPlayer:     make(map[string]int),
		TotalGameDuration: ac.data.TotalGameDuration,
		GameCount:         ac.data.GameCount,
	}

	for k, v := range ac.data.WinsPerPlayer {
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
