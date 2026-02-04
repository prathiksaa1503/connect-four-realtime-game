package main

import (
	"errors"
	"time"
)

const (
	BoardWidth  = 7
	BoardHeight = 6
)

type Player int

const (
	Empty Player = iota
	Player1
	Player2
)

type GameState int

const (
	Waiting GameState = iota
	InProgress
	Finished
)

type Game struct {
	ID           string
	Player1      string
	Player2      string
	Board        [BoardHeight][BoardWidth]Player
	CurrentTurn  Player
	State        GameState
	Winner       Player
	IsDraw       bool
	CreatedAt    time.Time
	StartedAt    *time.Time
	EndedAt      *time.Time
	LastMoveAt   time.Time
	IsBotGame    bool
	Player1Conn  *Connection
	Player2Conn  *Connection
}

// NewGame creates a new game instance
func NewGame(id, player1 string) *Game {
	return &Game{
		ID:          id,
		Player1:     player1,
		Board:       [BoardHeight][BoardWidth]Player{},
		CurrentTurn: Player1,
		State:       Waiting,
		CreatedAt:   time.Now(),
	}
}

// MakeMove attempts to make a move in the specified column
func (g *Game) MakeMove(column int, player Player) error {
	if g.State != InProgress {
		return errors.New("game is not in progress")
	}

	if player != g.CurrentTurn {
		return errors.New("not your turn")
	}

	if column < 0 || column >= BoardWidth {
		return errors.New("invalid column")
	}

	// Find the lowest empty row in the column
	row := -1
	for r := BoardHeight - 1; r >= 0; r-- {
		if g.Board[r][column] == Empty {
			row = r
			break
		}
	}

	if row == -1 {
		return errors.New("column is full")
	}

	// Place the disc
	g.Board[row][column] = player
	g.LastMoveAt = time.Now()

	// Check for win
	if g.checkWin(row, column, player) {
		g.State = Finished
		g.Winner = player
		now := time.Now()
		g.EndedAt = &now
		return nil
	}

	// Check for draw
	if g.isBoardFull() {
		g.State = Finished
		g.IsDraw = true
		now := time.Now()
		g.EndedAt = &now
		return nil
	}

	// Switch turn
	if g.CurrentTurn == Player1 {
		g.CurrentTurn = Player2
	} else {
		g.CurrentTurn = Player1
	}

	return nil
}

// checkWin checks if the last move resulted in a win
func (g *Game) checkWin(row, col int, player Player) bool {
	// Check horizontal
	count := 1
	for c := col - 1; c >= 0 && g.Board[row][c] == player; c-- {
		count++
	}
	for c := col + 1; c < BoardWidth && g.Board[row][c] == player; c++ {
		count++
	}
	if count >= 4 {
		return true
	}

	// Check vertical
	count = 1
	for r := row - 1; r >= 0 && g.Board[r][col] == player; r-- {
		count++
	}
	for r := row + 1; r < BoardHeight && g.Board[r][col] == player; r++ {
		count++
	}
	if count >= 4 {
		return true
	}

	// Check diagonal (top-left to bottom-right)
	count = 1
	r, c := row-1, col-1
	for r >= 0 && c >= 0 && g.Board[r][c] == player {
		count++
		r--
		c--
	}
	r, c = row+1, col+1
	for r < BoardHeight && c < BoardWidth && g.Board[r][c] == player {
		count++
		r++
		c++
	}
	if count >= 4 {
		return true
	}

	// Check diagonal (top-right to bottom-left)
	count = 1
	r, c = row-1, col+1
	for r >= 0 && c < BoardWidth && g.Board[r][c] == player {
		count++
		r--
		c++
	}
	r, c = row+1, col-1
	for r < BoardHeight && c >= 0 && g.Board[r][c] == player {
		count++
		r++
		c--
	}
	if count >= 4 {
		return true
	}

	return false
}

// isBoardFull checks if the board is completely filled
func (g *Game) isBoardFull() bool {
	for c := 0; c < BoardWidth; c++ {
		if g.Board[0][c] == Empty {
			return false
		}
	}
	return true
}

// GetValidMoves returns a list of columns that have available spaces
func (g *Game) GetValidMoves() []int {
	valid := []int{}
	for c := 0; c < BoardWidth; c++ {
		if g.Board[0][c] == Empty {
			valid = append(valid, c)
		}
	}
	return valid
}

// StartGame starts the game
func (g *Game) StartGame(player2 string) {
	g.Player2 = player2
	g.State = InProgress
	now := time.Now()
	g.StartedAt = &now
	g.LastMoveAt = now
}
