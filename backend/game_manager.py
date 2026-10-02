"""
Game session management. Positions, legality and status all come from the Go
engine; this module owns sessions, history, undo/redo, resignation and turn
orchestration for the AI opponent.
"""

import threading
import uuid
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional

from engine_client import EngineClient

MODES = ("ai", "local")
COLORS = ("white", "black")
DIFFICULTIES = ("easy", "medium", "hard")
PROMOTIONS = ("queen", "rook", "bishop", "knight")


class GameError(Exception):
    def __init__(self, message: str, status_code: int = 400):
        super().__init__(message)
        self.status_code = status_code


def opposite(color: str) -> str:
    return "black" if color == "white" else "white"


def valid_position(pos: Any) -> bool:
    return (isinstance(pos, dict)
            and all(isinstance(pos.get(k), int) and not isinstance(pos.get(k), bool) for k in ("row", "col"))
            and 0 <= pos["row"] < 8 and 0 <= pos["col"] < 8)


@dataclass
class Snapshot:
    """A position reached in the game plus the move that led to it."""
    state: dict
    status: dict
    fen: str
    position_key: str
    move: Optional[dict] = None

    @classmethod
    def from_engine(cls, data: dict) -> "Snapshot":
        return cls(state=data["state"], status=data["status"], fen=data["fen"],
                   position_key=data["positionKey"], move=data.get("move"))


@dataclass
class Game:
    id: str
    mode: str
    human_color: str
    difficulty: str
    snapshots: List[Snapshot]
    cursor: int = 0
    resigned_by: Optional[str] = None
    version: int = 0
    created_at: str = field(default_factory=lambda: datetime.now(timezone.utc).isoformat())
    lock: threading.RLock = field(default_factory=threading.RLock, repr=False)

    @property
    def current(self) -> Snapshot:
        return self.snapshots[self.cursor]

    @property
    def turn(self) -> str:
        return self.current.state["currentTurn"]

    @property
    def ai_color(self) -> Optional[str]:
        return opposite(self.human_color) if self.mode == "ai" else None

    def status(self) -> dict:
        if self.resigned_by:
            return {"status": "resigned", "inCheck": self.current.status.get("inCheck", False), "gameOver": True,
                    "winner": opposite(self.resigned_by), "reason": f"{self.resigned_by} resigned"}
        return self.current.status

    @property
    def game_over(self) -> bool:
        return bool(self.status().get("gameOver"))

    def history_keys(self) -> List[str]:
        return [s.position_key for s in self.snapshots[: self.cursor + 1]]

    def played_moves(self) -> List[dict]:
        return [s.move for s in self.snapshots[1: self.cursor + 1]]

    def is_ai_turn(self) -> bool:
        return self.mode == "ai" and not self.game_over and self.turn == self.ai_color

    def push(self, snapshot: Snapshot) -> None:
        # A new move discards any redo history.
        del self.snapshots[self.cursor + 1:]
        self.snapshots.append(snapshot)
        self.cursor += 1
        self.version += 1

    def to_view(self) -> dict:
        moves = self.played_moves()
        captured: Dict[str, List[str]] = {"white": [], "black": []}
        for m in moves:
            if m.get("captured"):
                captured[m["piece"]["color"]].append(m["captured"]["type"])
        return {
            "gameId": self.id,
            "mode": self.mode,
            "humanColor": self.human_color,
            "difficulty": self.difficulty,
            "state": self.current.state,
            "status": self.status(),
            "fen": self.current.fen,
            "moves": moves,
            "lastMove": moves[-1] if moves else None,
            "captured": captured,
            "canUndo": self._undo_target() is not None,
            "canRedo": self._redo_target() is not None,
            "aiToMove": self.is_ai_turn(),
            "createdAt": self.created_at,
        }

    def _is_resting_point(self, index: int) -> bool:
        """Undo/redo in AI mode stops only where it is the human's turn (or at the start)."""
        if self.mode != "ai" or index == 0:
            return True
        snap = self.snapshots[index]
        return snap.state["currentTurn"] == self.human_color or bool(snap.status.get("gameOver"))

    def _undo_target(self) -> Optional[int]:
        if self.resigned_by:
            return None
        for i in range(self.cursor - 1, -1, -1):
            if self._is_resting_point(i):
                return i
        return None

    def _redo_target(self) -> Optional[int]:
        if self.resigned_by:
            return None
        for i in range(self.cursor + 1, len(self.snapshots)):
            if self._is_resting_point(i):
                return i
        return None

    def undo(self) -> None:
        target = self._undo_target()
        if target is None:
            raise GameError("Nothing to undo", 409)
        self.cursor = target
        self.version += 1

    def redo(self) -> None:
        target = self._redo_target()
        if target is None:
            raise GameError("Nothing to redo", 409)
        self.cursor = target
        self.version += 1


class GameManager:
    def __init__(self, engine: EngineClient):
        self.engine = engine
        self.games: Dict[str, Game] = {}
        self._lock = threading.Lock()

    def get(self, game_id: str) -> Game:
        with self._lock:
            game = self.games.get(game_id)
        if game is None:
            raise GameError("Game not found", 404)
        return game

    def create(self, mode: str = "ai", human_color: str = "white", difficulty: str = "medium",
               fen: Optional[str] = None) -> Game:
        if mode not in MODES:
            raise GameError(f"mode must be one of {', '.join(MODES)}")
        if human_color not in COLORS:
            raise GameError(f"humanColor must be one of {', '.join(COLORS)}")
        if difficulty not in DIFFICULTIES:
            raise GameError(f"difficulty must be one of {', '.join(DIFFICULTIES)}")

        start = self.engine.from_fen(fen) if fen else self.engine.new_game()
        game = Game(id=str(uuid.uuid4()), mode=mode, human_color=human_color, difficulty=difficulty,
                    snapshots=[Snapshot.from_engine(start)])
        with self._lock:
            self.games[game.id] = game
        return game

    def valid_moves(self, game: Game, position: dict) -> List[dict]:
        if not valid_position(position):
            raise GameError("row and col must be integers between 0 and 7")
        with game.lock:
            if game.game_over or game.is_ai_turn():
                return []
            board = game.current.state
        return self.engine.valid_moves(board, position)

    def human_move(self, game: Game, from_pos: Any, to_pos: Any, promotion: Optional[str] = None) -> Game:
        if not valid_position(from_pos) or not valid_position(to_pos):
            raise GameError("from and to must be positions with row/col between 0 and 7")
        if promotion is not None and promotion not in PROMOTIONS:
            raise GameError(f"promotion must be one of {', '.join(PROMOTIONS)}")
        with game.lock:
            if game.game_over:
                raise GameError("Game is over", 409)
            if game.is_ai_turn():
                raise GameError("It is the AI's turn", 409)
            result = self.engine.make_move(game.current.state, from_pos, to_pos, promotion, game.history_keys())
            game.push(Snapshot.from_engine(result))
        return game

    def ai_move(self, game: Game) -> Game:
        with game.lock:
            if game.mode != "ai":
                raise GameError("AI moves are only available in AI games", 409)
            if game.game_over:
                raise GameError("Game is over", 409)
            if not game.is_ai_turn():
                raise GameError("It is not the AI's turn", 409)
            board, difficulty, version = game.current.state, game.difficulty, game.version

        # Search without holding the lock so reads (and undo/new game) stay responsive.
        choice = self.engine.best_move(board, difficulty)["move"]

        with game.lock:
            if game.version != version:
                raise GameError("Game changed while the AI was thinking", 409)
            result = self.engine.make_move(board, choice["from"], choice["to"],
                                           choice.get("promotedTo") or None, game.history_keys())
            game.push(Snapshot.from_engine(result))
        return game

    def undo(self, game: Game) -> Game:
        with game.lock:
            game.undo()
        return game

    def redo(self, game: Game) -> Game:
        with game.lock:
            game.redo()
        return game

    def resign(self, game: Game) -> Game:
        with game.lock:
            if game.game_over:
                raise GameError("Game is over", 409)
            game.resigned_by = game.human_color if game.mode == "ai" else game.turn
            game.version += 1
        return game
