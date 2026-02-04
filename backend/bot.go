package main

// BotPlayer represents a bot player
type BotPlayer struct {
	name string
}

// NewBotPlayer creates a new bot player
func NewBotPlayer() *BotPlayer {
	return &BotPlayer{
		name: "Bot",
	}
}

// GetMove returns the bot's move based on deterministic logic
func (b *BotPlayer) GetMove(game *Game) int {
	// Priority 1: Play winning move if available
	if move := b.findWinningMove(game, Player2); move != -1 {
		return move
	}

	// Priority 2: Block opponent's immediate winning move
	if move := b.findWinningMove(game, Player1); move != -1 {
		return move
	}

	// Priority 3: Choose the first valid column
	validMoves := game.GetValidMoves()
	if len(validMoves) > 0 {
		return validMoves[0]
	}

	return -1
}

// findWinningMove checks if there's a winning move for the given player
func (b *BotPlayer) findWinningMove(game *Game, player Player) int {
	validMoves := game.GetValidMoves()

	for _, col := range validMoves {
		// Create a temporary game state to test the move
		testGame := game.clone()
		row := -1
		for r := BoardHeight - 1; r >= 0; r-- {
			if testGame.Board[r][col] == Empty {
				row = r
				break
			}
		}

		if row == -1 {
			continue
		}

		// Test the move
		testGame.Board[row][col] = player
		if testGame.checkWin(row, col, player) {
			return col
		}
	}

	return -1
}

// clone creates a deep copy of the game for testing moves
func (g *Game) clone() *Game {
	clone := &Game{
		ID:          g.ID,
		Player1:     g.Player1,
		Player2:     g.Player2,
		CurrentTurn: g.CurrentTurn,
		State:       g.State,
		Winner:      g.Winner,
		IsDraw:      g.IsDraw,
		IsBotGame:   g.IsBotGame,
	}
	// Deep copy the board
	for r := 0; r < BoardHeight; r++ {
		for c := 0; c < BoardWidth; c++ {
			clone.Board[r][c] = g.Board[r][c]
		}
	}
	return clone
}
