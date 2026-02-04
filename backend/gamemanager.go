package main

import (
	"sync"
	"time"
)

// GameManager manages all active and completed games
type GameManager struct {
	games          map[string]*Game
	completedGames map[string]*Game
	mu             sync.RWMutex
}

func NewGameManager() *GameManager {
	return &GameManager{
		games:          make(map[string]*Game),
		completedGames: make(map[string]*Game),
	}
}

// AddGame adds a game to the manager
func (gm *GameManager) AddGame(game *Game) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.games[game.ID] = game
}

// GetGame retrieves a game by ID
func (gm *GameManager) GetGame(gameID string) (*Game, bool) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	game, exists := gm.games[gameID]
	if !exists {
		game, exists = gm.completedGames[gameID]
	}
	return game, exists
}

// GetGameByUsername retrieves a game by username
func (gm *GameManager) GetGameByUsername(username string) (*Game, bool) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	for _, game := range gm.games {
		if game.Player1 == username || game.Player2 == username {
			return game, true
		}
	}

	for _, game := range gm.completedGames {
		if game.Player1 == username || game.Player2 == username {
			return game, true
		}
	}

	return nil, false
}

// CompleteGame moves a game from active to completed
func (gm *GameManager) CompleteGame(gameID string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	if game, exists := gm.games[gameID]; exists {
		gm.completedGames[gameID] = game
		delete(gm.games, gameID)
	}
}

// CheckDisconnections checks for disconnected players and handles forfeits
func (gm *GameManager) CheckDisconnections() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			gm.mu.RLock()
			gamesToCheck := make([]*Game, 0, len(gm.games))
			for _, game := range gm.games {
				if game.State == InProgress {
					gamesToCheck = append(gamesToCheck, game)
				}
			}
			gm.mu.RUnlock()

			for _, game := range gamesToCheck {
				// Check if player disconnected (no activity for 30 seconds)
				if game.Player1Conn != nil {
					lastActivity := game.Player1Conn.GetLastActivity()
					if time.Since(lastActivity) > 30*time.Second && time.Since(game.LastMoveAt) > 30*time.Second {
						// Player 1 disconnected, player 2 wins
						game.State = Finished
						game.Winner = Player2
						now := time.Now()
						game.EndedAt = &now
						gm.CompleteGame(game.ID)

						// Notify player 2 if connected
						if game.Player2Conn != nil {
							sendGameState(game, game.Player2Conn)
						}

						// Emit game ended event
						eventProducer.PublishEvent(Event{
							Type:      "GAME_ENDED",
							GameID:    game.ID,
							Winner:    game.Player2,
							Timestamp: time.Now(),
						})
						continue
					}
				}

				if game.Player2Conn != nil && !game.IsBotGame {
					lastActivity := game.Player2Conn.GetLastActivity()
					if time.Since(lastActivity) > 30*time.Second && time.Since(game.LastMoveAt) > 30*time.Second {
						// Player 2 disconnected, player 1 wins
						game.State = Finished
						game.Winner = Player1
						now := time.Now()
						game.EndedAt = &now
						gm.CompleteGame(game.ID)

						// Notify player 1 if connected
						if game.Player1Conn != nil {
							sendGameState(game, game.Player1Conn)
						}

						// Emit game ended event
						eventProducer.PublishEvent(Event{
							Type:      "GAME_ENDED",
							GameID:    game.ID,
							Winner:    game.Player1,
							Timestamp: time.Now(),
						})
					}
				}
			}
		}
	}()
}

// GetLeaderboard returns the leaderboard data
func (gm *GameManager) GetLeaderboard() map[string]int {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	wins := make(map[string]int)
	for _, game := range gm.completedGames {
		if game.Winner == Player1 {
			wins[game.Player1]++
		} else if game.Winner == Player2 {
			wins[game.Player2]++
		}
	}

	return wins
}
