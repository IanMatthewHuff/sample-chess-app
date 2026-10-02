"""
Flask REST API for chess game management.
Chess rules are delegated to the Go engine (see engine_client.py).
"""

import os

from flask import Flask, jsonify, request
from flask_cors import CORS

from engine_client import EngineClient, EngineError, EngineUnavailable
from game_manager import GameError, GameManager


def create_app(engine: EngineClient = None) -> Flask:
    app = Flask(__name__)
    CORS(app)
    manager = GameManager(engine or EngineClient())
    app.config["GAME_MANAGER"] = manager

    @app.errorhandler(GameError)
    def handle_game_error(err: GameError):
        return jsonify({"error": str(err)}), err.status_code

    @app.errorhandler(EngineError)
    def handle_engine_error(err: EngineError):
        # 4xx from the engine means the request was rejected (e.g. illegal move).
        status = 400 if err.status_code < 500 else 502
        return jsonify({"error": str(err)}), status

    @app.errorhandler(EngineUnavailable)
    def handle_engine_unavailable(err: EngineUnavailable):
        return jsonify({"error": str(err)}), 503

    def body() -> dict:
        data = request.get_json(silent=True)
        return data if isinstance(data, dict) else {}

    @app.route("/api/health", methods=["GET"])
    def health_check():
        engine_ok = manager.engine.health()
        return jsonify({
            "status": "healthy" if engine_ok else "degraded",
            "service": "chess-backend",
            "engine": "healthy" if engine_ok else "unreachable",
        }), 200 if engine_ok else 503

    @app.route("/api/games", methods=["POST"])
    def create_game():
        data = body()
        game = manager.create(
            mode=data.get("mode", "ai"),
            human_color=data.get("humanColor", "white"),
            difficulty=data.get("difficulty", "medium"),
            fen=data.get("fen"),
        )
        return jsonify(game.to_view()), 201

    @app.route("/api/games/<game_id>", methods=["GET"])
    def get_game(game_id: str):
        return jsonify(manager.get(game_id).to_view())

    @app.route("/api/games/<game_id>/moves", methods=["GET"])
    def get_moves(game_id: str):
        game = manager.get(game_id)
        position = {"row": request.args.get("row", type=int), "col": request.args.get("col", type=int)}
        return jsonify({"moves": manager.valid_moves(game, position)})

    @app.route("/api/games/<game_id>/move", methods=["POST"])
    def make_game_move(game_id: str):
        game = manager.get(game_id)
        data = body()
        manager.human_move(game, data.get("from"), data.get("to"), data.get("promotion"))
        return jsonify(game.to_view())

    @app.route("/api/games/<game_id>/ai-move", methods=["POST"])
    def make_ai_move(game_id: str):
        game = manager.get(game_id)
        manager.ai_move(game)
        return jsonify(game.to_view())

    @app.route("/api/games/<game_id>/undo", methods=["POST"])
    def undo_game_move(game_id: str):
        return jsonify(manager.undo(manager.get(game_id)).to_view())

    @app.route("/api/games/<game_id>/redo", methods=["POST"])
    def redo_game_move(game_id: str):
        return jsonify(manager.redo(manager.get(game_id)).to_view())

    @app.route("/api/games/<game_id>/resign", methods=["POST"])
    def resign_game(game_id: str):
        return jsonify(manager.resign(manager.get(game_id)).to_view())

    @app.route("/api/games/<game_id>/history", methods=["GET"])
    def get_move_history(game_id: str):
        game = manager.get(game_id)
        moves = game.played_moves()
        return jsonify({"gameId": game_id, "moves": moves, "count": len(moves)})

    return app


if __name__ == "__main__":
    host = os.environ.get("BACKEND_HOST", "127.0.0.1")
    port = int(os.environ.get("BACKEND_PORT", "18472"))
    print(f"Chess backend API listening on http://{host}:{port} (engine: {EngineClient().base_url})")
    create_app().run(host=host, port=port, debug=False, threaded=True)
