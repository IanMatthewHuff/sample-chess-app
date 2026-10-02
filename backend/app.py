"""HTTP API for human-vs-computer chess games."""

import os
import uuid

import requests
from flask import Flask, jsonify, request
from flask_cors import CORS

from game_logic import ENGINE_URL, EngineError, GameState, engine


app = Flask(__name__)
CORS(app)
games: dict[str, GameState] = {}


@app.errorhandler(EngineError)
def engine_error(error):
    app.logger.error("%s", error)
    return jsonify({"error": str(error)}), 503


@app.errorhandler(ValueError)
def invalid_move(error):
    return jsonify({"error": str(error)}), 400


def game_for(game_id: str) -> GameState | None:
    return games.get(game_id)


def parse_move(data: dict) -> tuple[dict, dict, str]:
    if not isinstance(data, dict):
        raise ValueError("JSON move required")
    positions = []
    for field in ("from", "to"):
        position = data.get(field)
        if not isinstance(position, dict) or any(
            type(position.get(axis)) is not int or not 0 <= position[axis] < 8
            for axis in ("row", "col")
        ):
            raise ValueError(f"Invalid {field} square")
        positions.append({"row": position["row"], "col": position["col"]})
    promotion = data.get("promotion", "queen")
    if promotion not in ("queen", "rook", "bishop", "knight"):
        raise ValueError("Invalid promotion piece")
    return positions[0], positions[1], promotion


@app.get("/api/health")
def health():
    try:
        engine_response = requests.get(f"{ENGINE_URL}/health", timeout=2)
        engine_response.raise_for_status()
    except requests.RequestException as exc:
        raise EngineError(f"Chess engine unavailable: {exc}") from exc
    return jsonify({"status": "healthy", "service": "chess-backend"})


@app.post("/api/games")
def create_game():
    game_id = str(uuid.uuid4())
    game = GameState()
    result = engine("/status", {"board": game.board_state})
    game.status = result["status"]
    game.positions.append(result["repetitionKey"])
    games[game_id] = game
    return jsonify(game.response(game_id)), 201


@app.get("/api/games/<game_id>")
def get_game(game_id):
    game = game_for(game_id)
    return (jsonify(game.response(game_id)) if game else
            (jsonify({"error": "Game not found"}), 404))


@app.post("/api/games/<game_id>/move")
def move(game_id):
    game = game_for(game_id)
    if game is None:
        return jsonify({"error": "Game not found"}), 404
    if game.board_state["currentTurn"] != "white":
        raise ValueError("Wait for the computer's turn")
    from_pos, to_pos, promotion = parse_move(request.get_json(silent=True))
    game.play(from_pos, to_pos, promotion)
    return jsonify(game.response(game_id))


@app.post("/api/games/<game_id>/ai")
def ai_move(game_id):
    game = game_for(game_id)
    if game is None:
        return jsonify({"error": "Game not found"}), 404
    if game.board_state["currentTurn"] != "black" or game.status not in ("active", "check"):
        raise ValueError("Not the computer's turn")
    choice = engine("/best-move", {"board": game.board_state})
    game.play(choice["from"], choice["to"], choice.get("promotion", "queen"))
    return jsonify(game.response(game_id))


@app.get("/api/games/<game_id>/moves")
def moves(game_id):
    game = game_for(game_id)
    if game is None:
        return jsonify({"error": "Game not found"}), 404
    row, col = request.args.get("row", type=int), request.args.get("col", type=int)
    if row is None or col is None or not (0 <= row < 8 and 0 <= col < 8):
        raise ValueError("Invalid square")
    result = engine("/moves", {"board": game.board_state,
                               "position": {"row": row, "col": col}})
    captures = {(position["row"], position["col"]) for position in result["captures"]}
    return jsonify([{**position, "isCapture": (position["row"], position["col"]) in captures}
                    for position in result["moves"]])


@app.post("/api/games/<game_id>/validate")
def validate(game_id):
    game = game_for(game_id)
    if game is None:
        return jsonify({"error": "Game not found"}), 404
    from_pos, to_pos, _ = parse_move(request.get_json(silent=True))
    return jsonify(engine("/validate", {"board": game.board_state,
                                       "from": from_pos, "to": to_pos}))


@app.post("/api/games/<game_id>/undo")
def undo(game_id):
    game = game_for(game_id)
    if game is None:
        return jsonify({"error": "Game not found"}), 404
    game.undo()
    if game.board_state["currentTurn"] == "black" and game.snapshots:
        game.undo()
    return jsonify(game.response(game_id))


@app.post("/api/games/<game_id>/redo")
def redo(game_id):
    game = game_for(game_id)
    if game is None:
        return jsonify({"error": "Game not found"}), 404
    game.redo()
    if game.board_state["currentTurn"] == "black" and game.redo_stack:
        game.redo()
    return jsonify(game.response(game_id))


@app.get("/api/games/<game_id>/history")
def history(game_id):
    game = game_for(game_id)
    if game is None:
        return jsonify({"error": "Game not found"}), 404
    return jsonify({"gameId": game_id, "moves": game.move_history,
                    "count": len(game.move_history)})


if __name__ == "__main__":
    port = int(os.environ.get("BACKEND_PORT", "5000"))
    app.run(host="127.0.0.1", port=port, debug=False)
