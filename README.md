# Sample Chess App

A multi-language chess application designed for benchmarking AI code editor features like inline suggestions and Next Edit Suggestions (NES).

## Project Structure

```
sample-chess-app/
├── frontend/           # TypeScript - Web UI
├── backend/            # Python - Game state API
├── engine/             # Go - Move validation engine
└── README.md
```

## Components

### Frontend (TypeScript)
- Chess board visualization
- Piece rendering and drag-drop
- Game controls (new game, undo, etc.)

### Backend (Python)
- REST API for game state management
- Player session handling
- Game history and persistence

### Engine (Go)
- Move validation
- Legal move generation
- Check/checkmate detection

## Getting Started

Start each service in its own terminal, in this order. The example ports are
non-default so they can run alongside another local copy of the app; change
all matching port values together if a port is occupied.

**PowerShell (Windows):**
```powershell
cd engine
$env:ENGINE_PORT = "15387"
go run .
```
```powershell
cd backend
pip install -r requirements.txt
$env:ENGINE_URL = "http://127.0.0.1:15387"
$env:BACKEND_PORT = "15388"
python app.py
```
```powershell
cd frontend
npm ci
$env:BACKEND_PORT = "15388"
npm run dev -- --host 127.0.0.1 --port 15389 --strictPort
```
Open **http://127.0.0.1:15389/**. Play White against a basic Go-powered
computer opponent. Click a piece and a highlighted destination (or drag it);
choose the promotion piece in the popup when a pawn reaches the last rank. The
side panel displays pieces captured by each player. Undo and redo
operate on a human/computer pair when both moves are present. Games live in
Python memory and persist across a browser reload, but not a backend restart.
If the computer move fails, use **Retry computer move**.
The move list uses standard algebraic notation (for example, `e4`, `Nf3`,
`O-O`, `exd6`, and `Qh4#`) in vertically scrolling numbered rows. The last
move's start and end squares are shaded. Selecting a piece shows legal empty
destinations as green dots and legal captures (including en passant) as red
rings around the destination. A king in check (or checkmate) gets a red square
and an exclamation badge on the board.

The frontend proxies `/api` to Python; Python calls Go for all chess rules and
AI moves. Neither Python nor TypeScript validates chess legality locally.

Run the focused checks with `cd engine; go test ./...`,
`cd backend; python -m unittest -v test_app`, and
`cd frontend; npm test; npm run build`.

## Benchmarking Scenarios

This project supports various coding scenarios:
- **Type definitions**: Adding types to interfaces and functions
- **Function implementation**: Filling in stub functions
- **API integration**: Connecting frontend to backend
- **Algorithm implementation**: Move validation logic
- **Refactoring**: Improving code structure
- **Bug fixes**: Identifying and fixing issues
- **Feature additions**: Adding new game features
