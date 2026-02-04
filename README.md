# Connect Four – Real-Time Multiplayer Game

This project is a complete implementation of a real-time multiplayer Connect Four (4 in a Row) game built as part of a Backend Engineering assignment. The focus of this project is on real-time communication, backend correctness, and clean system design, with a simple frontend for interaction.

---

## Features

- Real-time Player vs Player gameplay using WebSockets
- Automatic matchmaking between players
- Competitive bot fallback if no opponent joins within 10 seconds
- Deterministic bot logic (non-random, strategic moves)
- Live game state synchronization between players
- Leaderboard tracking number of wins per player
- Event-driven analytics using Kafka-style simulation
- Player reconnection support within 30 seconds
- Graceful handling of disconnections and forfeits

---

## Tech Stack

- Backend: Go (net/http, gorilla/websocket)
- Frontend: HTML, CSS, Vanilla JavaScript
- State Management: In-memory
- Analytics: Go channels simulating Kafka producer/consumer

---

## Project Structure

assessment-emitter/
- backend/
  - main.go – Server entry point
  - game.go – Core game logic and rules
  - bot.go – Bot player logic
  - websocket.go – WebSocket setup
  - matchmaking.go – Player matchmaking
  - gamemanager.go – Game state management
  - handlers.go – WebSocket message handlers
  - kafka_simulator.go – Event producer
  - analytics.go – Event consumer
  - go.mod – Go dependencies
- analytics/
  - kafka_simulator.go – Standalone producer
  - consumer.go – Standalone analytics consumer
- frontend/
  - index.html – UI
  - style.css – Basic styling
  - script.js – Client-side logic
- README.md

---

## Prerequisites

- Go 1.21 or higher
- Modern web browser with WebSocket support

---

## How to Run

1. Navigate to the backend directory:
   cd backend

2. Download dependencies:
   go mod download

3. Run the server:
   go run .

4. Open the game in a browser:
   http://localhost:8080

Open two browser tabs or two different browsers to test multiplayer.

---

## How to Play

1. Enter a username and join the game
2. Wait to be matched with another player
3. If no opponent joins within 10 seconds, a bot starts the game
4. Click a column to drop a disc
5. The first player to connect four discs wins
6. View the leaderboard after the game ends

---

## Game Rules

- Board size: 7 columns × 6 rows
- Players take turns dropping discs
- Discs fall to the lowest available position
- Win conditions:
  - Horizontal
  - Vertical
  - Diagonal
- The game is a draw if the board fills with no winner

---

## Bot Logic

The bot plays deterministically using the following priority:
1. Make a winning move if available
2. Block the opponent’s immediate winning move
3. Choose the first valid column as a fallback

The bot does not play random moves.

---

## Real-Time Architecture

The game uses WebSockets for real-time, bidirectional communication.

Client to Server messages:
- JOIN
- MOVE
- RECONNECT
- GET_LEADERBOARD

Server to Client messages:
- JOINED
- GAME_STATE
- ERROR
- LEADERBOARD
- RECONNECTED

The server maintains the game state and pushes updates to connected clients.

---

## Event-Driven Analytics (Kafka Simulation)

The analytics system follows a Kafka-style architecture using Go channels.

Events emitted:
- GAME_STARTED
- MOVE_MADE
- GAME_ENDED

Analytics tracked:
- Total number of games played
- Wins per player
- Average game duration

Analytics processing is decoupled from the gameplay logic.

---

## Reconnection Handling

- Players can reconnect within 30 seconds using their username or game ID
- If a player fails to reconnect within the timeout, the opponent wins by forfeit

---

## Implementation Notes

- The backend acts as the single source of truth for all game state
- Real-time race conditions between client and server messages were handled using server-side context
- Message contracts were normalized to ensure stable multiplayer behavior
- In-memory storage was chosen to prioritize simplicity and real-time performance
- Kafka was simulated to demonstrate event-driven system design without external dependencies

---

## Future Improvements

- Persistent storage using a database
- User authentication
- Game history and replay
- Spectator mode
- Advanced bot AI
- Containerized deployment

---

## License

This project is provided as-is for technical assessment purposes.
