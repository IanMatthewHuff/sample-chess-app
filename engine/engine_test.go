package main

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

func perft(state *BoardState, depth int) int {
	if depth == 0 {
		return 1
	}
	moves := GenerateLegalMoves(state)
	if depth == 1 {
		return len(moves)
	}
	total := 0
	for _, m := range moves {
		total += perft(ApplyMove(state, m), depth-1)
	}
	return total
}

func mustFEN(t *testing.T, fen string) *BoardState {
	t.Helper()
	s, err := ParseFEN(fen)
	if err != nil {
		t.Fatalf("ParseFEN(%q): %v", fen, err)
	}
	return s
}

func sq(t *testing.T, s string) Position {
	t.Helper()
	p, ok := PositionFromAlgebraic(s)
	if !ok {
		t.Fatalf("bad square %q", s)
	}
	return p
}

// Reference counts from https://www.chessprogramming.org/Perft_Results
func TestPerft(t *testing.T) {
	cases := []struct {
		name  string
		fen   string
		depth int
		nodes int
	}{
		{"initial d3", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 3, 8902},
		{"initial d4", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 4, 197281},
		{"kiwipete d3", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 3, 97862},
		{"position3 d4", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 4, 43238},
		{"position4 d3", "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 3, 9467},
		{"position5 d3", "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 3, 62379},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			if testing.Short() && c.nodes > 100000 {
				t.Skip("slow perft")
			}
			if got := perft(mustFEN(t, c.fen), c.depth); got != c.nodes {
				t.Fatalf("perft = %d, want %d", got, c.nodes)
			}
		})
	}
}

func TestFENRoundTrip(t *testing.T) {
	fens := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2",
	}
	for _, f := range fens {
		if got := ToFEN(mustFEN(t, f)); got != f {
			t.Errorf("round trip: got %q want %q", got, f)
		}
	}
	if ToFEN(CreateInitialState()) != fens[0] {
		t.Errorf("initial state FEN mismatch: %s", ToFEN(CreateInitialState()))
	}
}

func TestCastlingMovesRookAndClearsRights(t *testing.T) {
	s := mustFEN(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	m, err := FindLegalMove(s, sq(t, "e1"), sq(t, "g1"), "")
	if err != nil || !m.IsCastling {
		t.Fatalf("expected legal kingside castle, got %+v %v", m, err)
	}
	next := ApplyMove(s, m)
	if p := next.Board.GetPieceAt(sq(t, "f1")); p == nil || p.Type != Rook {
		t.Fatalf("rook not moved to f1")
	}
	if next.Board.GetPieceAt(sq(t, "h1")) != nil {
		t.Fatalf("h1 should be empty")
	}
	if next.CastlingRights.WhiteKingside || next.CastlingRights.WhiteQueenside {
		t.Fatalf("white castling rights should be cleared")
	}
	if MoveToSAN(s, m) != "O-O" {
		t.Fatalf("SAN = %q", MoveToSAN(s, m))
	}
}

func TestCannotCastleThroughCheck(t *testing.T) {
	s := mustFEN(t, "4k3/8/8/8/8/8/5r2/R3K2R w KQ - 0 1")
	if IsValidMove(s, sq(t, "e1"), sq(t, "g1")) {
		t.Fatal("castling through attacked f1 must be illegal")
	}
	if !IsValidMove(s, sq(t, "e1"), sq(t, "c1")) {
		t.Fatal("queenside castling should be legal")
	}
}

func TestEnPassantRemovesPawn(t *testing.T) {
	s := mustFEN(t, "4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 1")
	m, err := FindLegalMove(s, sq(t, "e5"), sq(t, "d6"), "")
	if err != nil || !m.IsEnPassant {
		t.Fatalf("expected en passant, got %+v %v", m, err)
	}
	next := ApplyMove(s, m)
	if next.Board.GetPieceAt(sq(t, "d5")) != nil {
		t.Fatal("captured pawn should be removed")
	}
	if MoveToSAN(s, m) != "exd6" {
		t.Fatalf("SAN = %q", MoveToSAN(s, m))
	}
}

func TestPromotion(t *testing.T) {
	s := mustFEN(t, "3r3k/4P3/8/8/8/8/8/4K3 w - - 0 1")
	m, err := FindLegalMove(s, sq(t, "e7"), sq(t, "d8"), Knight)
	if err != nil {
		t.Fatal(err)
	}
	next := ApplyMove(s, m)
	if p := next.Board.GetPieceAt(sq(t, "d8")); p == nil || p.Type != Knight || p.Color != White {
		t.Fatalf("expected white knight on d8, got %+v", p)
	}
	if san := MoveToSAN(s, m); san != "exd8=N" {
		t.Fatalf("SAN = %q", san)
	}
	if q, _ := FindLegalMove(s, sq(t, "e7"), sq(t, "e8"), ""); MoveToSAN(s, q) != "e8=Q+" {
		t.Fatalf("SAN = %q", MoveToSAN(s, q))
	}
}

func TestStatuses(t *testing.T) {
	cases := []struct {
		fen    string
		status GameStatus
		reason string
	}{
		{"rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3", Checkmate, "checkmate"},
		{"7k/5Q2/6K1/8/8/8/8/8 b - - 0 1", Stalemate, "stalemate"},
		{"8/8/4k3/8/8/3K4/8/8 w - - 0 1", Draw, ReasonInsufficientMaterial},
		{"8/8/4k3/8/8/3KN3/8/8 w - - 0 1", Draw, ReasonInsufficientMaterial},
		{"8/8/4k3/8/8/R2K4/8/8 w - - 100 80", Draw, ReasonFiftyMove},
		{"4k3/8/8/8/8/8/4r3/4K3 w - - 0 1", Check, ""},
		{"8/8/4k3/8/8/R2K4/8/8 w - - 0 1", Active, ""},
	}
	for _, c := range cases {
		got := EvaluateStatus(mustFEN(t, c.fen), nil)
		if got.Status != c.status || got.Reason != c.reason {
			t.Errorf("%s: got %+v, want %s/%s", c.fen, got, c.status, c.reason)
		}
	}
	mate := EvaluateStatus(mustFEN(t, cases[0].fen), nil)
	if mate.Winner != Black || !mate.GameOver {
		t.Errorf("fool's mate winner = %q gameOver=%v", mate.Winner, mate.GameOver)
	}
}

func TestThreefoldRepetition(t *testing.T) {
	s := CreateInitialState()
	history := []string{PositionKey(s)}
	shuffle := [][2]string{{"g1", "f3"}, {"g8", "f6"}, {"f3", "g1"}, {"f6", "g8"}}
	var status StatusResult
	for i := 0; i < 2; i++ {
		for _, mv := range shuffle {
			m, err := FindLegalMove(s, sq(t, mv[0]), sq(t, mv[1]), "")
			if err != nil {
				t.Fatal(err)
			}
			s = ApplyMove(s, m)
			history = append(history, PositionKey(s))
			status = EvaluateStatus(s, history)
		}
	}
	if status.Status != Draw || status.Reason != ReasonThreefoldRepetition {
		t.Fatalf("expected threefold draw, got %+v", status)
	}
}

func TestSANDisambiguation(t *testing.T) {
	s := mustFEN(t, "4k3/8/8/8/8/8/4K3/R6R w - - 0 1")
	m, _ := FindLegalMove(s, sq(t, "a1"), sq(t, "d1"), "")
	if san := MoveToSAN(s, m); san != "Rad1" {
		t.Fatalf("SAN = %q", san)
	}
}

func TestAIFindsMateInOne(t *testing.T) {
	s := mustFEN(t, "6k1/5ppp/8/8/8/8/5PPP/3R2K1 w - - 0 1")
	for _, d := range []Difficulty{Easy, Medium, Hard} {
		res, ok := FindBestMove(s, d, rand.New(rand.NewSource(1)))
		if !ok || MoveToSAN(s, res.Move) != "Rd8#" {
			t.Errorf("%s: expected Rd8#, got %s", d, MoveToSAN(s, res.Move))
		}
	}
}

func TestAIWinsHangingQueen(t *testing.T) {
	s := mustFEN(t, "4k3/8/8/3q4/8/8/8/3RK3 w - - 0 1")
	res, ok := FindBestMove(s, Medium, rand.New(rand.NewSource(1)))
	if !ok || !res.Move.To.Equals(sq(t, "d5")) {
		t.Fatalf("expected Rxd5, got %+v", res.Move)
	}
}

func postJSON(t *testing.T, h http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHTTPMoveFlow(t *testing.T) {
	mux := newMux()
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/new", nil))
	var start StateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &start); err != nil {
		t.Fatal(err)
	}

	rec = postJSON(t, mux, "/move", map[string]any{
		"board": start.State, "from": sq(t, "e2"), "to": sq(t, "e4"), "history": []string{start.PositionKey},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("move status %d: %s", rec.Code, rec.Body.String())
	}
	var moved MoveResponse
	json.Unmarshal(rec.Body.Bytes(), &moved)
	if moved.Move.SAN != "e4" || moved.State.CurrentTurn != Black {
		t.Fatalf("unexpected move response: %+v", moved.Move)
	}

	rec = postJSON(t, mux, "/move", map[string]any{"board": moved.State, "from": sq(t, "e7"), "to": sq(t, "e4")})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("illegal move should be 400, got %d", rec.Code)
	}

	rec = postJSON(t, mux, "/ai", map[string]any{"board": moved.State, "difficulty": "easy"})
	if rec.Code != http.StatusOK {
		t.Fatalf("ai status %d: %s", rec.Code, rec.Body.String())
	}

	rec = postJSON(t, mux, "/moves", map[string]any{"board": start.State, "position": sq(t, "g1")})
	var vm ValidMovesResponse
	json.Unmarshal(rec.Body.Bytes(), &vm)
	if len(vm.Moves) != 2 {
		t.Fatalf("knight should have 2 moves, got %+v", vm.Moves)
	}

	bad := *start.State
	bad.Board[7][4] = nil
	if rec = postJSON(t, mux, "/status", map[string]any{"board": bad}); rec.Code != http.StatusBadRequest {
		t.Fatalf("board without king should be rejected, got %d", rec.Code)
	}
}
