"""API tests with the engine boundary mocked; rules are exercised in Go tests."""

import unittest
from unittest.mock import patch

import app
from game_logic import GameState


class GameApiTest(unittest.TestCase):
    def setUp(self):
        app.games.clear()
        self.client = app.app.test_client()
        self.engine_patch = patch("game_logic.engine", side_effect=self.fake_engine)
        self.app_engine_patch = patch("app.engine", side_effect=self.fake_engine)
        self.engine_patch.start()
        self.app_engine_patch.start()
        self.addCleanup(self.engine_patch.stop)
        self.addCleanup(self.app_engine_patch.stop)

    @staticmethod
    def fake_engine(path, payload):
        board = payload["board"]
        if path == "/status":
            placement = [[None if piece is None else (piece["type"], piece["color"])
                          for piece in row] for row in board["board"]]
            return {"status": "active", "repetitionKey": str(placement) + board["currentTurn"]}
        if path == "/moves":
            return {"moves": [{"row": 4, "col": 4}], "captures": []}
        if path == "/validate":
            return {"valid": True}
        if path == "/best-move":
            return {"from": {"row": 1, "col": 4}, "to": {"row": 3, "col": 4}}
        if path == "/move-san":
            from_pos, to_pos = payload["from"], payload["to"]
            state = {**board, "board": [row.copy() for row in board["board"]]}
            state["board"][to_pos["row"]][to_pos["col"]] = board["board"][from_pos["row"]][from_pos["col"]]
            state["board"][from_pos["row"]][from_pos["col"]] = None
            state["currentTurn"] = "black" if board["currentTurn"] == "white" else "white"
            return {"state": state, "san": "e4"}
        raise AssertionError(path)

    def test_game_turn_undo_redo_and_validation(self):
        response = self.client.post("/api/games")
        self.assertEqual(response.status_code, 201)
        id = response.json["gameId"]
        self.assertEqual(len(self.client.get(f"/api/games/{id}/moves?row=6&col=4").json), 1)
        self.assertFalse(self.client.get(f"/api/games/{id}/moves?row=6&col=4").json[0]["isCapture"])
        self.assertEqual(self.client.get(f"/api/games/{id}/moves?row=9&col=4").status_code, 400)
        self.assertEqual(self.client.post(f"/api/games/{id}/move", json={
            "from": {"row": -1, "col": 4}, "to": {"row": 4, "col": 4}}).status_code, 400)
        moved = self.client.post(f"/api/games/{id}/move", json={
            "from": {"row": 6, "col": 4}, "to": {"row": 4, "col": 4}})
        self.assertEqual(moved.json["state"]["currentTurn"], "black")
        self.assertEqual(moved.json["history"][0]["san"], "e4")
        self.assertEqual(self.client.post(f"/api/games/{id}/move", json={
            "from": {"row": 6, "col": 3}, "to": {"row": 4, "col": 3}}).status_code, 400)
        ai = self.client.post(f"/api/games/{id}/ai")
        self.assertEqual(ai.json["state"]["currentTurn"], "white")
        self.assertEqual(len(ai.json["history"]), 2)
        undone = self.client.post(f"/api/games/{id}/undo")
        self.assertEqual(len(undone.json["history"]), 0)
        self.assertEqual(undone.json["state"]["board"][6][4]["type"], "pawn")
        redone = self.client.post(f"/api/games/{id}/redo")
        self.assertEqual(len(redone.json["history"]), 2)
        self.assertEqual(redone.json["state"]["currentTurn"], "white")

    def test_missing_game_and_engine_failure(self):
        self.assertEqual(self.client.get("/api/games/no-such-game").status_code, 404)
        with patch("app.engine", side_effect=app.EngineError("offline")):
            response = self.client.post("/api/games")
        self.assertEqual(response.status_code, 503)
        self.assertEqual(app.games, {})

    def test_third_repetition_is_draw(self):
        game = self.client.post("/api/games").json
        root = f"/api/games/{game['gameId']}"
        cycle = [
            ((7, 6), (5, 5)),
            ((0, 6), (2, 5)),
            ((5, 5), (7, 6)),
            ((2, 5), (0, 6)),
        ]
        for _ in range(2):
            for from_square, to_square in cycle:
                state = app.games[game["gameId"]]
                state.play({"row": from_square[0], "col": from_square[1]},
                           {"row": to_square[0], "col": to_square[1]})
        self.assertEqual(self.client.get(root).json["status"], "draw")
        self.assertEqual(self.client.post(root + "/ai").status_code, 400)

    def test_en_passant_capture_is_recorded(self):
        game = GameState()
        board = game.board_state["board"]
        board[3][4], board[6][4] = board[6][4], None
        board[3][3], board[1][3] = board[1][3], None
        game.board_state["enPassantTarget"] = {"row": 2, "col": 3}
        game.positions.append("before")
        game.play({"row": 3, "col": 4}, {"row": 2, "col": 3})
        self.assertEqual(game.move_history[-1]["captured"]["type"], "pawn")
        self.assertEqual(game.move_history[-1]["captured"]["color"], "black")

    def test_capture_annotation_from_engine(self):
        game_id = self.client.post("/api/games").json["gameId"]
        original = self.fake_engine

        def with_capture(path, payload):
            if path == "/moves":
                return {"moves": [{"row": 2, "col": 3}, {"row": 2, "col": 4}],
                        "captures": [{"row": 2, "col": 3}]}
            return original(path, payload)

        with patch("app.engine", side_effect=with_capture):
            moves = self.client.get(f"/api/games/{game_id}/moves?row=3&col=4").json
        self.assertEqual(moves, [{"row": 2, "col": 3, "isCapture": True},
                                 {"row": 2, "col": 4, "isCapture": False}])


if __name__ == "__main__":
    unittest.main()
