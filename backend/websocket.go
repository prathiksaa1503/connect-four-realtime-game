package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Connection represents a WebSocket connection
type Connection struct {
	conn         *websocket.Conn
	send         chan []byte
	username     string
	gameID       string
	lastActivity time.Time
	mu           sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	GameID    string      `json:"gameId,omitempty"`
	Username  string      `json:"username,omitempty"`
	Column    int         `json:"column,omitempty"`
}

// GameResponse represents the game state sent to clients
type GameResponse struct {
	GameID      string              `json:"gameId"`
	Player1     string              `json:"player1"`
	Player2     string              `json:"player2"`
	Board       [BoardHeight][BoardWidth]int `json:"board"`
	CurrentTurn int                 `json:"currentTurn"`
	State       string              `json:"state"`
	Winner      int                 `json:"winner"`
	IsDraw      bool                `json:"isDraw"`
	IsBotGame   bool                `json:"isBotGame"`
}

// ConnectionManager manages all WebSocket connections
type ConnectionManager struct {
	connections map[*Connection]bool
	register    chan *Connection
	unregister  chan *Connection
	broadcast   chan []byte
	mu          sync.RWMutex
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[*Connection]bool),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		broadcast:   make(chan []byte, 256),
	}
}

func (cm *ConnectionManager) run() {
	for {
		select {
		case conn := <-cm.register:
			cm.mu.Lock()
			cm.connections[conn] = true
			cm.mu.Unlock()

		case conn := <-cm.unregister:
			cm.mu.Lock()
			if _, ok := cm.connections[conn]; ok {
				delete(cm.connections, conn)
				close(conn.send)
			}
			cm.mu.Unlock()

		case message := <-cm.broadcast:
			cm.mu.RLock()
			for conn := range cm.connections {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(cm.connections, conn)
				}
			}
			cm.mu.RUnlock()
		}
	}
}

func (c *Connection) readPump() {
	defer func() {
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.updateActivity()
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.updateActivity()

		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("error unmarshaling message: %v", err)
			continue
		}

		handleMessage(c, &msg)
	}
}

// updateActivity updates the last activity timestamp
func (c *Connection) updateActivity() {
	c.mu.Lock()
	c.lastActivity = time.Now()
	c.mu.Unlock()
}

// GetLastActivity returns the last activity time
func (c *Connection) GetLastActivity() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastActivity
}

func (c *Connection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWS(manager *ConnectionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("error upgrading connection: %v", err)
			return
		}

		connection := &Connection{
			conn:         conn,
			send:         make(chan []byte, 256),
			lastActivity: time.Now(),
		}

		manager.register <- connection

		go connection.writePump()
		go connection.readPump()
	}
}
