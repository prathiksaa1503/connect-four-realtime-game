package analytics

import (
	"time"
)

// Event represents a game event
type Event struct {
	Type      string    `json:"type"`
	GameID    string    `json:"gameId,omitempty"`
	Player1   string    `json:"player1,omitempty"`
	Player2   string    `json:"player2,omitempty"`
	Player    string    `json:"player,omitempty"`
	Column    int       `json:"column,omitempty"`
	Winner    string    `json:"winner,omitempty"`
	IsDraw    bool      `json:"isDraw,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// EventProducer simulates a Kafka producer using Go channels
type EventProducer struct {
	eventChannel chan Event
}

// NewEventProducer creates a new event producer
func NewEventProducer(bufferSize int) *EventProducer {
	return &EventProducer{
		eventChannel: make(chan Event, bufferSize),
	}
}

// PublishEvent publishes an event to the channel (simulating Kafka)
func (ep *EventProducer) PublishEvent(event Event) {
	select {
	case ep.eventChannel <- event:
		// Event published successfully
	default:
		// Channel buffer full, drop event (in real Kafka, this would be handled differently)
	}
}

// GetEventChannel returns the event channel for consumers
func (ep *EventProducer) GetEventChannel() <-chan Event {
	return ep.eventChannel
}
