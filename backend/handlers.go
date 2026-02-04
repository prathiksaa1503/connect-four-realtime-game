package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"
)

var (
	gameManager      = NewGameManager()
	matchmakingQueue = NewMatchmakingQueue()
)

func init() {
	rand.Seed(time.Now().UnixNano())
	gameManager.CheckDisconnections()
}

// generateGameID generates a unique game ID
func generateGameID() string {
	return time.Now().Format("20060102150405") + string(rune(rand.Intn(1000)))
}

// handleMessage processes incoming WebSocket messages
func handleMessage(conn *Connection, msg *Message) {
	switch msg.Type {
	case "JOIN":
		handleJoin(conn, msg)
	case "MOVE":
		handleMove(conn, msg)
	case "RECONNECT":
		handleReconnect(conn, msg)
	case "GET_LEADERBOARD":
		handleGetLeaderboard(conn)
	default:
		sendError(conn, "unknown message type")
	}
}

// handleJoin handles a player joining the game
func handleJoin(conn *Connection, msg *Message) {
	if msg.Username == "" {
		sendError(conn, "username is required")
		return
	}

	conn.username = msg.Username
	gameID := matchmakingQueue.AddPlayer(msg.Username, conn)
	conn.gameID = gameID

	response := Message{
		Type:     "JOINED",
		GameID:   gameID,
		Username: msg.Username,
	}
	sendMessage(conn, &response)
}

// handleMove handles a player making a move
func handleMove(conn *Connection, msg *Message) {
	// 🔒 Fallback: use connection gameID if client didn't send it yet
	if msg.GameID == "" {
		msg.GameID = conn.gameID
	}

	if msg.GameID == "" {
		sendError(conn, "game ID is required")
		return
	}

	game, exists := gameManager.GetGame(msg.GameID)

	if !exists {
		sendError(conn, "game not found")
		return
	}

	// Determine which player is making the move
	var player Player
	if game.Player1 == conn.username {
		player = Player1
	} else if game.Player2 == conn.username {
		player = Player2
	} else {
		sendError(conn, "you are not a player in this game")
		return
	}

	// Make the move
	if err := game.MakeMove(msg.Column, player); err != nil {
		sendError(conn, err.Error())
		return
	}

	// Emit move made event
	eventProducer.PublishEvent(Event{
		Type:      "MOVE_MADE",
		GameID:    msg.GameID,
		Player:    conn.username,
		Column:    msg.Column,
		Timestamp: time.Now(),
	})

	// Send updated game state to both players
	if game.Player1Conn != nil {
		sendGameState(game, game.Player1Conn)
	}
	if game.Player2Conn != nil {
		sendGameState(game, game.Player2Conn)
	}

	// If game is finished, move to completed games
	if game.State == Finished {
		gameManager.CompleteGame(game.ID)

		// Emit game ended event
		winner := ""
		if game.Winner == Player1 {
			winner = game.Player1
		} else if game.Winner == Player2 {
			winner = game.Player2
		}

		eventProducer.PublishEvent(Event{
			Type:      "GAME_ENDED",
			GameID:    game.ID,
			Winner:    winner,
			IsDraw:    game.IsDraw,
			Timestamp: time.Now(),
		})
	} else if game.IsBotGame && game.CurrentTurn == Player2 {
		// Bot's turn - make bot move after a short delay
		go func() {
			time.Sleep(500 * time.Millisecond)
			bot := NewBotPlayer()
			botMove := bot.GetMove(game)
			if botMove != -1 {
				game.MakeMove(botMove, Player2)

				// Emit move made event
				eventProducer.PublishEvent(Event{
					Type:      "MOVE_MADE",
					GameID:    game.ID,
					Player:    "Bot",
					Column:    botMove,
					Timestamp: time.Now(),
				})

				// Send updated game state
				if game.Player1Conn != nil {
					sendGameState(game, game.Player1Conn)
				}

				// If game is finished
				if game.State == Finished {
					gameManager.CompleteGame(game.ID)

					winner := ""
					if game.Winner == Player1 {
						winner = game.Player1
					} else if game.Winner == Player2 {
						winner = game.Player2
					}

					eventProducer.PublishEvent(Event{
						Type:      "GAME_ENDED",
						GameID:    game.ID,
						Winner:    winner,
						IsDraw:    game.IsDraw,
						Timestamp: time.Now(),
					})
				}
			}
		}()
	}
}

// handleReconnect handles a player reconnecting
func handleReconnect(conn *Connection, msg *Message) {
	var game *Game
	var exists bool

	if msg.GameID != "" {
		game, exists = gameManager.GetGame(msg.GameID)
	} else if msg.Username != "" {
		game, exists = gameManager.GetGameByUsername(msg.Username)
	} else {
		sendError(conn, "game ID or username is required")
		return
	}

	if !exists {
		sendError(conn, "game not found")
		return
	}

	// Update connection
	if game.Player1 == msg.Username {
		game.Player1Conn = conn
	} else if game.Player2 == msg.Username {
		game.Player2Conn = conn
	} else {
		sendError(conn, "you are not a player in this game")
		return
	}

	conn.username = msg.Username
	conn.gameID = game.ID

	// Send current game state
	sendGameState(game, conn)

	response := Message{
		Type:   "RECONNECTED",
		GameID: game.ID,
	}
	sendMessage(conn, &response)
}

// handleGetLeaderboard handles leaderboard requests
func handleGetLeaderboard(conn *Connection) {
	leaderboard := gameManager.GetLeaderboard()
	response := Message{
		Type: "LEADERBOARD",
		Data: leaderboard,
	}
	sendMessage(conn, &response)
}

// sendGameState sends the current game state to a connection
func sendGameState(game *Game, conn *Connection) {
	if conn == nil {
		return
	}

	// Convert board to int array for JSON
	board := make([][]int, BoardHeight)
	for r := 0; r < BoardHeight; r++ {
		board[r] = make([]int, BoardWidth)
		for c := 0; c < BoardWidth; c++ {
			board[r][c] = int(game.Board[r][c])
		}
	}

	stateStr := "waiting"
	if game.State == InProgress {
		stateStr = "inProgress"
	} else if game.State == Finished {
		stateStr = "finished"
	}

	gameResp := GameResponse{
		GameID:      game.ID,
		Player1:     game.Player1,
		Player2:     game.Player2,
		Board:       [BoardHeight][BoardWidth]int{},
		CurrentTurn: int(game.CurrentTurn),
		State:       stateStr,
		Winner:      int(game.Winner),
		IsDraw:      game.IsDraw,
		IsBotGame:   game.IsBotGame,
	}

	// Copy board
	for r := 0; r < BoardHeight; r++ {
		for c := 0; c < BoardWidth; c++ {
			gameResp.Board[r][c] = int(game.Board[r][c])
		}
	}

	response := Message{
		Type:   "GAME_STATE",
		Data:   gameResp,
		GameID: game.ID,
	}
	sendMessage(conn, &response)
}

// sendMessage sends a message to a connection
func sendMessage(conn *Connection, msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error marshaling message: %v", err)
		return
	}

	select {
	case conn.send <- data:
	default:
		log.Printf("connection send buffer full")
	}
}

// sendError sends an error message to a connection
func sendError(conn *Connection, errorMsg string) {
	response := Message{
		Type:  "ERROR",
		Error: errorMsg,
	}
	sendMessage(conn, &response)
}
