# Sample Chess App

A multi-language chess application designed for benchmarking AI code editor features like inline suggestions and Next Edit Suggestions (NES). It is also a fully playable game with a basic AI opponent.

## Architecture

```
Browser ──► frontend (TypeScript/Vite, :18471) ──/api proxy──► backend (Python/Flask, :18472) ──► engine (Go, :18473)
            UX only: rendering, clicks,                      game management: sessions,          all chess rules + AI:
            promotion picker, controls                       history, undo/redo, resign,          move gen, check/mate/draws,
                                                             AI turn orchestration                SAN, FEN, alpha-beta search
```

| Component | Responsibility | Key files |
|-----------|----------------|-----------|
| `engine/` (Go) | **Single source of truth for chess rules**: legal moves (castling, en passant, promotion), check/checkmate/stalemate, draws (50-move, threefold repetition, insufficient material), SAN, FEN, and the AI (alpha-beta + quiescence, material + piece-square evaluation) | `moves.go`, `status.go`, `fen.go`, `ai.go`, `main.go` |
| `backend/` (Python) | Game sessions, full-position history for exact undo/redo, resignation, captured pieces, deciding when the AI moves. Contains no chess rules. | `app.py`, `game_manager.py`, `engine_client.py` |
| `frontend/` (TypeScript) | Renders whatever the backend returns and sends user intents. Contains no chess rules. | `src/index.ts`, `src/board.ts`, `src/api.ts` |

## Quick start (Windows / PowerShell)

Prerequisites: Go 1.21+, Python 3.10+, Node 18+.

```powershell
pip install -r backend\requirements.txt
.\start.ps1          # builds the engine, starts all three services, opens http://127.0.0.1:18471/
.\stop.ps1           # stops them
```

The ports are deliberately not the usual defaults (5173/5000/8080), so the app can run next to other local projects. To use different ports:

```powershell
.\start.ps1 -FrontendPort 19471 -BackendPort 19472 -EnginePort 19473 -NoBrowser
```

Logs are written to `.run\` (gitignored).

## Running services manually

```bash
# Engine: ENGINE_PORT (default 18473)
cd engine && go run .

# Backend: BACKEND_PORT (default 18472), ENGINE_URL (default http://127.0.0.1:18473)
cd backend && pip install -r requirements.txt && python app.py

# Frontend: FRONTEND_PORT (default 18471), BACKEND_URL (default http://127.0.0.1:18472)
cd frontend && npm install && npm run dev
```

## Playing

- Click a piece to see its legal moves: a dot marks a quiet move and a ring marks a capture. Click a highlighted square to move there.
- When a pawn promotes, a picker appears (Queen, Rook, Bishop or Knight).
- **New game** options: vs AI or two players on one board, play as white or black, and AI difficulty:
  - **Easy**: searches 1 ply ahead and picks randomly among moves within 1.2 pawns of the best.
  - **Medium**: searches 2 ply.
  - **Hard**: searches up to 4 ply (about 3 s limit) and always plays its best move.
- **Undo/Redo** in AI games steps over both your move and the AI's reply. **Resign** ends the game.
- The current game is restored if you reload the page.

## Backend API

| Method | Path | Body / query |
|--------|------|--------------|
| GET | `/api/health` | returns 503 if the engine is unreachable |
| POST | `/api/games` | `{mode: "ai"\|"local", humanColor, difficulty, fen?}` |
| GET | `/api/games/:id` | |
| GET | `/api/games/:id/moves?row=&col=` | legal targets for a square |
| POST | `/api/games/:id/move` | `{from: {row, col}, to: {row, col}, promotion?}` |
| POST | `/api/games/:id/ai-move` | makes the AI play when it's its turn |
| POST | `/api/games/:id/undo`, `/redo`, `/resign` | |
| GET | `/api/games/:id/history` | |

Every game endpoint returns the same game view: board state, status, SAN move list, captured pieces, `canUndo`, `canRedo` and `aiToMove`.

Engine endpoints (`/health`, `/new`, `/fen`, `/moves`, `/legal`, `/validate`, `/move`, `/status`, `/ai`) are stateless. Each request carries the board, and `/move` and `/status` can also take a `history` of position keys for repetition detection.

## Tests

```powershell
cd engine;   go test ./...          # perft vs. reference counts, special moves, draws, SAN, AI, HTTP
cd backend;  python -m pytest -q    # session logic with a fake engine + integration tests against the real engine
cd frontend; npm run typecheck
```

## Benchmarking Scenarios

This project supports various coding scenarios (see `EDIT_SCENARIOS.md`):
- **Type definitions**: Adding types to interfaces and functions
- **Function implementation**: Filling in stub functions
- **API integration**: Connecting frontend to backend
- **Algorithm implementation**: Move validation logic
- **Refactoring**: Improving code structure
- **Bug fixes**: Identifying and fixing issues
- **Feature additions**: Adding new game features

> Note: the Python and TypeScript rules code (`backend/board.py`, `backend/game_logic.py`, `backend/pieces.py`, and the move logic in `frontend/src/game.ts`) was removed so that the Go engine is the only place rules are implemented. Scenarios in `EDIT_SCENARIOS.md` that refer to those functions no longer match the code.
