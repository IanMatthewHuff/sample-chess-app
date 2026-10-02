import os
import subprocess
import socket
import time
from pathlib import Path

import pytest

from app import create_app
from engine_client import EngineClient, EngineError


class FakeEngine:
    """Minimal stand-in for the Go engine. It does not know chess: every move
    toggles the side to move; a move with from == to is 'illegal'."""

    def __init__(self):
        self.ai_calls = 0
        self.mate_after = None

    def _snapshot(self, turn, n, move=None, game_over=False):
        status = {"status": "active", "inCheck": False, "gameOver": False}
        if game_over:
            status = {"status": "checkmate", "inCheck": True, "gameOver": True,
                      "winner": "white" if turn == "black" else "black", "reason": "checkmate"}
        data = {"state": {"currentTurn": turn, "n": n}, "status": status,
                "fen": f"fen{n}", "positionKey": f"key{n}"}
        if move:
            data["move"] = move
        return data

    def health(self):
        return True

    def new_game(self):
        return self._snapshot("white", 0)

    def from_fen(self, fen):
        return self._snapshot("black" if " b " in fen else "white", 0)

    def valid_moves(self, board, position):
        return [{"row": 4, "col": 4, "isCapture": False, "isPromotion": False, "isCastling": False}]

    def make_move(self, board, from_pos, to_pos, promotion, history):
        if from_pos == to_pos:
            raise EngineError("Illegal move", 400)
        n = board["n"] + 1
        turn = board["currentTurn"]
        move = {"from": from_pos, "to": to_pos, "piece": {"type": "pawn", "color": turn},
                "san": f"m{n}", "promotedTo": promotion}
        if n % 3 == 0:
            move["captured"] = {"type": "knight", "color": "black" if turn == "white" else "white"}
        return self._snapshot("black" if turn == "white" else "white", n, move,
                              game_over=self.mate_after == n)

    def best_move(self, board, difficulty):
        self.ai_calls += 1
        return {"move": {"from": {"row": 1, "col": 0}, "to": {"row": 2, "col": 0}}}


A = {"row": 6, "col": 4}
B = {"row": 4, "col": 4}


@pytest.fixture
def engine():
    return FakeEngine()


@pytest.fixture
def client(engine):
    return create_app(engine).test_client()


def new_game(client, **kwargs):
    res = client.post("/api/games", json=kwargs)
    assert res.status_code == 201, res.get_json()
    return res.get_json()


def test_create_game_defaults(client):
    game = new_game(client)
    assert game["mode"] == "ai" and game["humanColor"] == "white" and game["difficulty"] == "medium"
    assert game["moves"] == [] and not game["canUndo"] and not game["aiToMove"]


@pytest.mark.parametrize("payload", [{"mode": "x"}, {"humanColor": "red"}, {"difficulty": "insane"}])
def test_create_game_validation(client, payload):
    assert client.post("/api/games", json=payload).status_code == 400


def test_unknown_game_404(client):
    assert client.get("/api/games/nope").status_code == 404


def test_ai_game_flow(client, engine):
    game = new_game(client)
    gid = game["gameId"]

    res = client.post(f"/api/games/{gid}/move", json={"from": A, "to": B})
    view = res.get_json()
    assert res.status_code == 200 and view["aiToMove"] and len(view["moves"]) == 1

    # Human may not move on the AI's turn
    assert client.post(f"/api/games/{gid}/move", json={"from": A, "to": B}).status_code == 409
    assert client.get(f"/api/games/{gid}/moves?row=6&col=4").get_json()["moves"] == []

    view = client.post(f"/api/games/{gid}/ai-move").get_json()
    assert engine.ai_calls == 1 and not view["aiToMove"] and len(view["moves"]) == 2

    # AI move when it's not the AI's turn is rejected
    assert client.post(f"/api/games/{gid}/ai-move").status_code == 409


def test_illegal_and_malformed_moves(client):
    gid = new_game(client)["gameId"]
    assert client.post(f"/api/games/{gid}/move", json={"from": A, "to": A}).status_code == 400
    assert client.post(f"/api/games/{gid}/move", json={"from": A}).status_code == 400
    assert client.post(f"/api/games/{gid}/move", json={"from": {"row": 9, "col": 0}, "to": B}).status_code == 400
    assert client.post(f"/api/games/{gid}/move", json={"from": A, "to": B, "promotion": "king"}).status_code == 400
    assert client.get(f"/api/games/{gid}/moves?row=a&col=1").status_code == 400


def test_undo_redo_in_ai_mode_skips_ai_reply(client):
    gid = new_game(client)["gameId"]
    client.post(f"/api/games/{gid}/move", json={"from": A, "to": B})
    client.post(f"/api/games/{gid}/ai-move")

    view = client.post(f"/api/games/{gid}/undo").get_json()
    assert view["moves"] == [] and view["state"]["currentTurn"] == "white" and view["canRedo"]
    assert client.post(f"/api/games/{gid}/undo").status_code == 409

    view = client.post(f"/api/games/{gid}/redo").get_json()
    assert len(view["moves"]) == 2 and view["state"]["currentTurn"] == "white"


def test_new_move_clears_redo(client):
    gid = new_game(client, mode="local")["gameId"]
    client.post(f"/api/games/{gid}/move", json={"from": A, "to": B})
    client.post(f"/api/games/{gid}/undo")
    view = client.post(f"/api/games/{gid}/move", json={"from": A, "to": {"row": 5, "col": 4}}).get_json()
    assert not view["canRedo"] and len(view["moves"]) == 1


def test_local_mode_single_step_undo_and_no_ai(client):
    gid = new_game(client, mode="local")["gameId"]
    client.post(f"/api/games/{gid}/move", json={"from": A, "to": B})
    view = client.post(f"/api/games/{gid}/move", json={"from": B, "to": A}).get_json()
    assert not view["aiToMove"] and len(view["moves"]) == 2
    assert client.post(f"/api/games/{gid}/ai-move").status_code == 409
    assert len(client.post(f"/api/games/{gid}/undo").get_json()["moves"]) == 1


def test_human_black_ai_moves_first(client):
    view = new_game(client, humanColor="black")
    assert view["aiToMove"]
    view = client.post(f"/api/games/{view['gameId']}/ai-move").get_json()
    assert view["state"]["currentTurn"] == "black" and not view["aiToMove"]


def test_captured_pieces(client):
    gid = new_game(client, mode="local")["gameId"]
    for _ in range(3):
        view = client.post(f"/api/games/{gid}/move", json={"from": A, "to": B}).get_json()
    assert view["captured"] == {"white": ["knight"], "black": []}


def test_resign(client):
    gid = new_game(client)["gameId"]
    view = client.post(f"/api/games/{gid}/resign").get_json()
    assert view["status"]["status"] == "resigned" and view["status"]["winner"] == "black"
    assert client.post(f"/api/games/{gid}/move", json={"from": A, "to": B}).status_code == 409
    assert client.post(f"/api/games/{gid}/resign").status_code == 409


def test_game_over_blocks_moves(client, engine):
    engine.mate_after = 1
    gid = new_game(client)["gameId"]
    view = client.post(f"/api/games/{gid}/move", json={"from": A, "to": B}).get_json()
    assert view["status"]["gameOver"] and not view["aiToMove"]
    assert client.post(f"/api/games/{gid}/ai-move").status_code == 409


def test_history_endpoint(client):
    gid = new_game(client, mode="local")["gameId"]
    client.post(f"/api/games/{gid}/move", json={"from": A, "to": B})
    data = client.get(f"/api/games/{gid}/history").get_json()
    assert data["count"] == 1 and data["moves"][0]["san"] == "m1"


def test_engine_unavailable_returns_503():
    app = create_app(EngineClient("http://127.0.0.1:1"))
    c = app.test_client()
    assert c.post("/api/games", json={}).status_code == 503
    assert c.get("/api/health").status_code == 503


# ---------------------------------------------------------------------------
# Integration tests against the real Go engine
# ---------------------------------------------------------------------------

ENGINE_DIR = Path(__file__).resolve().parent.parent / "engine"


def _free_port():
    with socket.socket() as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


@pytest.fixture(scope="module")
def real_client():
    if subprocess.run(["go", "version"], capture_output=True).returncode != 0:
        pytest.skip("Go toolchain not available")
    port = _free_port()
    binary = ENGINE_DIR / ("chess-engine-test.exe" if os.name == "nt" else "chess-engine-test")
    subprocess.run(["go", "build", "-o", str(binary), "."], cwd=ENGINE_DIR, check=True)
    proc = subprocess.Popen([str(binary)], env={**os.environ, "ENGINE_PORT": str(port)},
                            stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    engine = EngineClient(f"http://127.0.0.1:{port}")
    try:
        for _ in range(50):
            if engine.health():
                break
            time.sleep(0.1)
        else:
            pytest.fail("engine did not start")
        yield create_app(engine).test_client()
    finally:
        proc.terminate()
        proc.wait(timeout=5)
        binary.unlink(missing_ok=True)


def sq(name):
    return {"row": 8 - int(name[1]), "col": "abcdefgh".index(name[0])}


def test_real_engine_fools_mate(real_client):
    gid = new_game(real_client, mode="local")["gameId"]
    for frm, to in [("f2", "f3"), ("e7", "e5"), ("g2", "g4"), ("d8", "h4")]:
        res = real_client.post(f"/api/games/{gid}/move", json={"from": sq(frm), "to": sq(to)})
        assert res.status_code == 200, res.get_json()
    view = res.get_json()
    assert view["status"]["status"] == "checkmate" and view["status"]["winner"] == "black"
    assert [m["san"] for m in view["moves"]] == ["f3", "e5", "g4", "Qh4#"]


def test_real_engine_ai_game_and_illegal_move(real_client):
    view = new_game(real_client, difficulty="easy")
    gid = view["gameId"]
    assert real_client.post(f"/api/games/{gid}/move", json={"from": sq("e2"), "to": sq("e5")}).status_code == 400
    moves = real_client.get(f"/api/games/{gid}/moves?row=6&col=4").get_json()["moves"]
    assert {(m["row"], m["col"]) for m in moves} == {(5, 4), (4, 4)}
    real_client.post(f"/api/games/{gid}/move", json={"from": sq("e2"), "to": sq("e4")})
    view = real_client.post(f"/api/games/{gid}/ai-move").get_json()
    assert len(view["moves"]) == 2 and view["state"]["currentTurn"] == "white"
