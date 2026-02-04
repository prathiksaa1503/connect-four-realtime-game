package main

import (
	"sync"
	"time"
)

// MatchmakingQueue manages the queue of waiting players
type MatchmakingQueue struct {
	waitingPlayers map[string]*WaitingPlayer
	mu             sync.RWMutex
}

// WaitingPlayer represents a player waiting for a match
type WaitingPlayer struct {
	Username string
	Conn     *Connection
	GameID   string
	JoinedAt time.Time
}

func NewMatchmakingQueue() *MatchmakingQueue {
	return &MatchmakingQueue{
		waitingPlayers: make(map[string]*WaitingPlayer),
	}
}

// AddPlayer adds a player to the matchmaking queue
func (mq *MatchmakingQueue) AddPlayer(username string, conn *Connection) string {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	// Check if player is already waiting
	if wp, exists := mq.waitingPlayers[username]; exists {
		return wp.GameID
	}

	// Try to find an opponent
	for otherUsername, otherPlayer := range mq.waitingPlayers {
		if otherUsername != username {

			// ✅ USE THE SAME GAME ID (KEY FIX)
			gameID := otherPlayer.GameID

			game := NewGame(gameID, otherPlayer.Username)
			game.Player1Conn = otherPlayer.Conn
			game.Player2Conn = conn
			game.StartGame(username)
			game.State = InProgress

			// Remove waiting player
			delete(mq.waitingPlayers, otherUsername)

			// Add game to game manager
			gameManager.AddGame(game)

			// Notify both players
			sendGameState(game, game.Player1Conn)
			sendGameState(game, game.Player2Conn)

			// Emit game started event
			eventProducer.PublishEvent(Event{
				Type:      "GAME_STARTED",
				GameID:    gameID,
				Player1:   otherPlayer.Username,
				Player2:   username,
				Timestamp: time.Now(),
			})

			return gameID
		}
	}

	// No opponent found, add to queue
	gameID := generateGameID()
	mq.waitingPlayers[username] = &WaitingPlayer{
		Username: username,
		Conn:     conn,
		GameID:   gameID,
		JoinedAt: time.Now(),
	}

	// Start timeout goroutine
	go mq.startMatchmakingTimeout(username, gameID)

	return gameID
}

// startMatchmakingTimeout starts a bot game after 10 seconds if no opponent joins
func (mq *MatchmakingQueue) startMatchmakingTimeout(username, gameID string) {
	time.Sleep(10 * time.Second)

	mq.mu.Lock()
	defer mq.mu.Unlock()

	// Check if player is still waiting
	if wp, exists := mq.waitingPlayers[username]; exists {

		bot := NewBotPlayer()
		game := NewGame(gameID, username)
		game.Player1Conn = wp.Conn
		game.Player2 = bot.name
		game.IsBotGame = true
		game.StartGame(bot.name)
		game.State = InProgress

		delete(mq.waitingPlayers, username)

		gameManager.AddGame(game)

		sendGameState(game, wp.Conn)

		eventProducer.PublishEvent(Event{
			Type:      "GAME_STARTED",
			GameID:    gameID,
			Player1:   username,
			Player2:   bot.name,
			Timestamp: time.Now(),
		})
	}
}

// RemovePlayer removes a player from the matchmaking queue
func (mq *MatchmakingQueue) RemovePlayer(username string) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	delete(mq.waitingPlayers, username)
}
