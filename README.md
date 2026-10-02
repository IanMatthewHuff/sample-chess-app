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

### Frontend
```bash
cd frontend
npm install
npm run dev
```

### Backend
```bash
cd backend
pip install -r requirements.txt
python app.py
```

### Engine
```bash
cd engine
go run .
```

## Benchmarking Scenarios

This project supports various coding scenarios:
- **Type definitions**: Adding types to interfaces and functions
- **Function implementation**: Filling in stub functions
- **API integration**: Connecting frontend to backend
- **Algorithm implementation**: Move validation logic
- **Refactoring**: Improving code structure
- **Bug fixes**: Identifying and fixing issues
- **Feature additions**: Adding new game features
