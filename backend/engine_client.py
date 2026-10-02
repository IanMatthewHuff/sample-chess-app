"""
HTTP client for the Go chess engine. All chess rules live in the engine;
this module only transports requests and normalises errors.
"""

import os
from typing import Any, Dict, List, Optional

import requests

ENGINE_URL = os.environ.get("ENGINE_URL", "http://127.0.0.1:18473")
DEFAULT_TIMEOUT = 5.0
AI_TIMEOUT = 15.0


class EngineError(Exception):
    """The engine rejected a request (e.g. an illegal move)."""

    def __init__(self, message: str, status_code: int = 400):
        super().__init__(message)
        self.status_code = status_code


class EngineUnavailable(Exception):
    """The engine could not be reached."""


class EngineClient:
    def __init__(self, base_url: str = ENGINE_URL):
        self.base_url = base_url.rstrip("/")
        self.session = requests.Session()

    def _request(self, method: str, path: str, payload: Optional[dict] = None,
                 timeout: float = DEFAULT_TIMEOUT) -> Any:
        try:
            response = self.session.request(method, f"{self.base_url}{path}", json=payload, timeout=timeout)
        except requests.RequestException as exc:
            raise EngineUnavailable(f"Chess engine unreachable at {self.base_url}: {exc}") from exc

        try:
            data = response.json()
        except ValueError:
            data = {"error": response.text.strip() or "invalid engine response"}

        if response.status_code >= 400:
            raise EngineError(data.get("error", "engine error"), response.status_code)
        return data

    def health(self) -> bool:
        try:
            return self._request("GET", "/health", timeout=2).get("status") == "healthy"
        except (EngineError, EngineUnavailable):
            return False

    def new_game(self) -> Dict[str, Any]:
        return self._request("GET", "/new")

    def from_fen(self, fen: str) -> Dict[str, Any]:
        return self._request("POST", "/fen", {"fen": fen})

    def valid_moves(self, board: dict, position: dict) -> List[dict]:
        return self._request("POST", "/moves", {"board": board, "position": position})["moves"]

    def make_move(self, board: dict, from_pos: dict, to_pos: dict,
                  promotion: Optional[str], history: List[str]) -> Dict[str, Any]:
        payload = {"board": board, "from": from_pos, "to": to_pos, "history": history}
        if promotion:
            payload["promotion"] = promotion
        return self._request("POST", "/move", payload)

    def best_move(self, board: dict, difficulty: str) -> Dict[str, Any]:
        return self._request("POST", "/ai", {"board": board, "difficulty": difficulty}, timeout=AI_TIMEOUT)
