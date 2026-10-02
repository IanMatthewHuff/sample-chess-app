"""Game sessions and history; the Go service is the only chess rules authority."""

from dataclasses import dataclass, field
import os

import requests

from board import create_initial_state


ENGINE_URL = os.environ.get("ENGINE_URL", "http://127.0.0.1:8080")


class EngineError(Exception):
    pass


def engine(path: str, payload: dict) -> dict | list:
    try:
        response = requests.post(f"{ENGINE_URL}{path}", json=payload, timeout=10)
        if response.status_code == 400:
            raise ValueError(response.text.strip())
        response.raise_for_status()
        return response.json()
    except (requests.RequestException, requests.exceptions.JSONDecodeError) as exc:
        raise EngineError(f"Chess engine unavailable: {exc}") from exc


@dataclass
class GameState:
    board_state: dict = field(default_factory=lambda: create_initial_state().to_dict())
    status: str = "active"
    move_history: list[dict] = field(default_factory=list)
    snapshots: list[dict] = field(default_factory=list)
    redo_stack: list[tuple[dict, dict, str, str]] = field(default_factory=list)
    positions: list[str] = field(default_factory=list)

    def response(self, game_id: str) -> dict:
        return {"gameId": game_id, "state": self.board_state, "status": self.status,
                "history": self.move_history, "canUndo": bool(self.move_history),
                "canRedo": bool(self.redo_stack)}

    def play(self, from_pos: dict, to_pos: dict, promotion: str = "queen") -> None:
        if self.status in ("checkmate", "stalemate", "draw"):
            raise ValueError("Game is over")
        payload = {"board": self.board_state, "from": from_pos, "to": to_pos,
                   "promotion": promotion}
        move_result = engine("/move-san", payload)
        new_state = move_result["state"]
        result = engine("/status", {"board": new_state})
        new_status = result["status"]
        piece = self.board_state["board"][from_pos["row"]][from_pos["col"]]
        captured = self.board_state["board"][to_pos["row"]][to_pos["col"]]
        if piece["type"] == "pawn" and from_pos["col"] != to_pos["col"] and captured is None:
            captured = self.board_state["board"][from_pos["row"]][to_pos["col"]]
        move = {"from": from_pos, "to": to_pos, "piece": piece,
                "captured": captured, "promotion": promotion if piece["type"] == "pawn"
                and to_pos["row"] in (0, 7) else None, "san": move_result["san"]}
        self.snapshots.append(self.board_state)
        self.board_state = new_state
        self.positions.append(result["repetitionKey"])
        self.status = "draw" if self.positions.count(self.positions[-1]) >= 3 else new_status
        self.move_history.append(move)
        self.redo_stack.clear()

    def undo(self) -> None:
        if not self.snapshots:
            raise ValueError("No move to undo")
        old_status = self.status
        previous = self.snapshots[-1]
        new_status = engine("/status", {"board": previous})["status"]
        self.snapshots.pop()
        move = self.move_history.pop()
        key = self.positions.pop()
        self.redo_stack.append((self.board_state, move, old_status, key))
        self.board_state = previous
        self.status = "draw" if self.positions.count(self.positions[-1]) >= 3 else new_status

    def redo(self) -> None:
        if not self.redo_stack:
            raise ValueError("No move to redo")
        state, move, status, key = self.redo_stack.pop()
        self.snapshots.append(self.board_state)
        self.board_state = state
        self.move_history.append(move)
        self.positions.append(key)
        self.status = status
