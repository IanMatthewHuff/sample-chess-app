"""
Flask REST API for chess game
"""

import uuid
from flask import Flask, request, jsonify
from flask_cors import CORS
from typing import Dict

from pieces import Position
from game_logic import GameState, create_new_game, make_move, undo_move, get_valid_moves, is_valid_move


app = Flask(__name__)
CORS(app)

# In-memory game storage
games: Dict[str, GameState] = {}


@app.route('/api/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    return jsonify({"status": "healthy", "service": "chess-backend"})


@app.route('/api/games', methods=['POST'])
def create_game():
    """Create a new game"""
    # TODO: Implement game creation endpoint
    game_id = str(uuid.uuid4())
    games[game_id] = create_new_game()
    
    return jsonify({
        "gameId": game_id,
        "state": games[game_id].board_state.to_dict(),
        "status": games[game_id].status.value
    }), 201


@app.route('/api/games/<game_id>', methods=['GET'])
def get_game(game_id: str):
    """Get current game state"""
    # TODO: Implement game retrieval endpoint
    if game_id not in games:
        return jsonify({"error": "Game not found"}), 404
    
    game = games[game_id]
    return jsonify({
        "gameId": game_id,
        "state": game.board_state.to_dict(),
        "status": game.status.value,
        "moveCount": len(game.move_history)
    })


@app.route('/api/games/<game_id>/move', methods=['POST'])
def make_game_move(game_id: str):
    """Make a move in the game"""
    # TODO: Implement move endpoint
    if game_id not in games:
        return jsonify({"error": "Game not found"}), 404
    
    data = request.get_json()
    from_pos = Position.from_dict(data['from'])
    to_pos = Position.from_dict(data['to'])
    
    game = games[game_id]
    
    # Validate move
    if not is_valid_move(game.board_state, from_pos, to_pos):
        return jsonify({
            "valid": False,
            "error": "Invalid move"
        }), 400
    
    # Make the move
    games[game_id] = make_move(game, from_pos, to_pos)
    
    return jsonify({
        "valid": True,
        "state": games[game_id].board_state.to_dict(),
        "status": games[game_id].status.value
    })


@app.route('/api/games/<game_id>/moves', methods=['GET'])
def get_moves(game_id: str):
    """Get valid moves for a position"""
    # TODO: Implement valid moves endpoint
    if game_id not in games:
        return jsonify({"error": "Game not found"}), 404
    
    row = request.args.get('row', type=int)
    col = request.args.get('col', type=int)
    
    if row is None or col is None:
        return jsonify({"error": "Position required"}), 400
    
    position = Position(row, col)
    game = games[game_id]
    
    valid_moves = get_valid_moves(game.board_state, position)
    
    return jsonify([m.to_dict() for m in valid_moves])


@app.route('/api/games/<game_id>/validate', methods=['POST'])
def validate_move(game_id: str):
    """Validate a move (calls Go engine)"""
    # TODO: Implement validation endpoint (call Go engine)
    if game_id not in games:
        return jsonify({"error": "Game not found"}), 404
    
    data = request.get_json()
    from_pos = Position.from_dict(data['from'])
    to_pos = Position.from_dict(data['to'])
    
    game = games[game_id]
    valid = is_valid_move(game.board_state, from_pos, to_pos)
    
    # TODO: Call Go engine for validation
    # engine_result = call_go_engine(game.board_state, from_pos, to_pos)
    
    return jsonify({
        "valid": valid,
        "reason": None if valid else "Invalid move"
    })


@app.route('/api/games/<game_id>/undo', methods=['POST'])
def undo_game_move(game_id: str):
    """Undo the last move"""
    # TODO: Implement undo endpoint
    if game_id not in games:
        return jsonify({"error": "Game not found"}), 404
    
    games[game_id] = undo_move(games[game_id])
    
    return jsonify({
        "gameId": game_id,
        "state": games[game_id].board_state.to_dict(),
        "status": games[game_id].status.value
    })


@app.route('/api/games/<game_id>/history', methods=['GET'])
def get_move_history(game_id: str):
    """Get move history for a game"""
    # TODO: Implement history endpoint
    if game_id not in games:
        return jsonify({"error": "Game not found"}), 404
    
    game = games[game_id]
    history = [move.to_dict() for move in game.move_history]
    
    return jsonify({
        "gameId": game_id,
        "moves": history,
        "count": len(history)
    })


def call_go_engine(board_state, from_pos: Position, to_pos: Position) -> dict:
    """Call the Go validation engine"""
    # TODO: Implement Go engine integration
    import requests
    
    try:
        response = requests.post(
            "http://localhost:8080/validate",
            json={
                "board": board_state.to_dict(),
                "from": from_pos.to_dict(),
                "to": to_pos.to_dict()
            },
            timeout=5
        )
        return response.json()
    except requests.RequestException as e:
        return {"valid": False, "error": str(e)}


if __name__ == '__main__':
    print("Starting Chess Backend API...")
    print("API running at http://localhost:5000")
    app.run(host='0.0.0.0', port=5000, debug=True)
