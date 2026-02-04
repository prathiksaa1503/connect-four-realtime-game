let ws = null;
let currentGame = null;
let username = '';
let gameId = '';
let socketReady = false;

// DOM elements
const loginSection = document.getElementById('loginSection');
const gameSection = document.getElementById('gameSection');
const leaderboardSection = document.getElementById('leaderboardSection');
const usernameInput = document.getElementById('usernameInput');
const joinButton = document.getElementById('joinButton');
const newGameButton = document.getElementById('newGameButton');
const leaderboardButton = document.getElementById('leaderboardButton');
const closeLeaderboardButton = document.getElementById('closeLeaderboardButton');
const board = document.getElementById('board');
const gameStatus = document.getElementById('gameStatus');
const player1Name = document.getElementById('player1Name');
const player2Name = document.getElementById('player2Name');
const player1Indicator = document.getElementById('player1Indicator');
const player2Indicator = document.getElementById('player2Indicator');
const messageDiv = document.getElementById('message');

/* ---------------- BOARD ---------------- */
function initializeBoard() {
    board.innerHTML = '';
    for (let r = 0; r < 6; r++) {
        for (let c = 0; c < 7; c++) {
            const cell = document.createElement('div');
            cell.className = 'cell';
            cell.addEventListener('click', () => handleCellClick(c));
            board.appendChild(cell);
        }
    }
}

/* ---------------- SOCKET ---------------- */
function connectWebSocket() {
    if (ws && socketReady) return;

    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    ws = new WebSocket(`${protocol}//${location.host}/ws`);

    ws.onopen = () => {
        socketReady = true;
        sendMessage({ type: 'JOIN', username });
    };

    ws.onmessage = (e) => handleMessage(JSON.parse(e.data));

    ws.onclose = () => {
        socketReady = false;
        setTimeout(connectWebSocket, 2000);
    };
}

/* ---------------- MESSAGES ---------------- */
function handleMessage(message) {
    switch (message.type) {
        case 'JOINED':
            gameId = message.gameId;
            showMessage('Waiting for opponent...');
            break;

        case 'GAME_STATE':
            gameId = message.gameId || gameId;
            updateGameState(message.data);
            break;

        case 'ERROR':
            showMessage(message.error, 'error');
            break;

        case 'LEADERBOARD':
            displayLeaderboard(message.data);
            break;
    }
}

/* ---------------- GAME STATE ---------------- */
function updateGameState(game) {
    currentGame = game;

    player1Name.textContent = game.player1;
    player2Name.textContent = game.player2 || 'Waiting';

    for (let r = 0; r < 6; r++) {
        for (let c = 0; c < 7; c++) {
            const cell = board.children[r * 7 + c];
            cell.className = 'cell';
            if (game.board[r][c] === 1) cell.classList.add('player1');
            if (game.board[r][c] === 2) cell.classList.add('player2');
        }
    }

    if (game.state === 'inProgress') {
        const myTurn =
            (game.currentTurn === 1 && username === game.player1) ||
            (game.currentTurn === 2 && username === game.player2);
        gameStatus.textContent = myTurn ? 'Your turn!' : 'Opponent turn';
    }
}

/* ---------------- MOVE ---------------- */
function handleCellClick(col) {
    if (!currentGame || currentGame.state !== 'inProgress') return;

    sendMessage({
        type: 'MOVE',
        gameId,
        column: col
    });
}

/* ---------------- SEND ---------------- */
function sendMessage(msg) {
    if (socketReady) {
        ws.send(JSON.stringify(msg));
    }
}

/* ---------------- UI ---------------- */
function showMessage(text, type = '') {
    messageDiv.textContent = text;
    messageDiv.className = `message ${type}`;
    messageDiv.classList.remove('hidden');
    setTimeout(() => messageDiv.classList.add('hidden'), 3000);
}

function displayLeaderboard(data) {
    document.getElementById('leaderboardContent').innerHTML =
        Object.entries(data || {}).map(
            ([n, w]) => `<div>${n}: ${w}</div>`
        ).join('');
    leaderboardSection.classList.remove('hidden');
}

/* ---------------- EVENTS ---------------- */
joinButton.onclick = () => {
    username = usernameInput.value.trim();
    if (!username) return;

    loginSection.classList.add('hidden');
    gameSection.classList.remove('hidden');

    initializeBoard();
    connectWebSocket();
};

newGameButton.onclick = () => sendMessage({ type: 'JOIN', username });
leaderboardButton.onclick = () => sendMessage({ type: 'GET_LEADERBOARD' });
closeLeaderboardButton.onclick = () => leaderboardSection.classList.add('hidden');

initializeBoard();
