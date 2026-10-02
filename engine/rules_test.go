package main

import (
	"strings"
	"testing"
)

func perft(state *BoardState, depth int) int {
	if depth == 0 {
		return 1
	}
	count := 0
	for _, entry := range state.Board.GetPiecesByColor(state.CurrentTurn) {
		for _, to := range GetValidMoves(state, entry.Position) {
			if entry.Piece.Type == Pawn && (to.Row == 0 || to.Row == 7) {
				for _, promotion := range []PieceType{Queen, Rook, Bishop, Knight} {
					count += perft(MakeMove(state, entry.Position, to, promotion), depth-1)
				}
			} else {
				count += perft(MakeMove(state, entry.Position, to, Queen), depth-1)
			}
		}
	}
	return count
}

func TestInitialPerft(t *testing.T) {
	for _, test := range []struct{ depth, want int }{{1, 20}, {2, 400}, {3, 8902}, {4, 197281}} {
		if got := perft(CreateInitialState(), test.depth); got != test.want {
			t.Errorf("depth %d: got %d, want %d", test.depth, got, test.want)
		}
	}
}

func TestEndgamePerft(t *testing.T) {
	state := emptyState()
	rows := strings.Split("8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8", "/")
	pieces := map[rune]PieceType{'p': Pawn, 'r': Rook, 'k': King}
	for row, rank := range rows {
		col := 0
		for _, symbol := range rank {
			if symbol >= '1' && symbol <= '8' {
				col += int(symbol - '0')
				continue
			}
			color := Black
			if symbol >= 'A' && symbol <= 'Z' {
				color = White
				symbol += 'a' - 'A'
			}
			put(state, row, col, pieces[symbol], color)
			col++
		}
	}
	for _, test := range []struct{ depth, want int }{{1, 14}, {2, 191}, {3, 2812}, {4, 43238}} {
		if got := perft(state, test.depth); got != test.want {
			t.Errorf("endgame depth %d: got %d, want %d", test.depth, got, test.want)
		}
	}
}

func emptyState() *BoardState {
	return &BoardState{CurrentTurn: White, FullMoveNumber: 1}
}

func put(s *BoardState, row, col int, kind PieceType, color PieceColor) {
	s.Board[row][col] = NewPiece(kind, color)
}

func TestInitialMovesAndOpening(t *testing.T) {
	s := CreateInitialState()
	count := 0
	for _, p := range s.Board.GetPiecesByColor(White) {
		count += len(GetValidMoves(s, p.Position))
	}
	if count != 20 {
		t.Fatalf("opening moves: %d, want 20", count)
	}
	if IsValidMove(s, Position{6, 4}, Position{3, 4}) {
		t.Fatal("pawn cannot jump three squares")
	}
	s = MakeMove(s, Position{6, 4}, Position{4, 4}, Queen)
	if s.CurrentTurn != Black || s.EnPassantTarget == nil || *s.EnPassantTarget != (Position{5, 4}) {
		t.Fatalf("incorrect state after e4: %+v", s)
	}
}

func TestCastleThroughCheckAndMoveRook(t *testing.T) {
	s := emptyState()
	put(s, 7, 4, King, White)
	put(s, 7, 7, Rook, White)
	put(s, 0, 4, King, Black)
	s.CastlingRights.WhiteKingside = true
	if !IsValidMove(s, Position{7, 4}, Position{7, 6}) {
		t.Fatal("castling should be legal")
	}
	next := MakeMove(s, Position{7, 4}, Position{7, 6}, Queen)
	if next.Board[7][5] == nil || next.Board[7][5].Type != Rook || next.Board[7][7] != nil ||
		next.CastlingRights.WhiteKingside {
		t.Fatal("castling did not move rook and revoke right")
	}
	put(s, 0, 5, Rook, Black)
	if IsValidMove(s, Position{7, 4}, Position{7, 6}) {
		t.Fatal("castling through attacked f1")
	}
}

func TestEnPassantRemovesPawnAndDetectsDiscoveredCheck(t *testing.T) {
	s := emptyState()
	put(s, 7, 4, King, White)
	put(s, 0, 4, King, Black)
	put(s, 3, 4, Pawn, White)
	put(s, 3, 3, Pawn, Black)
	s.EnPassantTarget = &Position{2, 3}
	if !IsValidMove(s, Position{3, 4}, Position{2, 3}) {
		t.Fatal("en passant should be legal")
	}
	next := MakeMove(s, Position{3, 4}, Position{2, 3}, Queen)
	if next.Board[3][3] != nil || next.Board[2][3] == nil {
		t.Fatal("en passant capture incorrect")
	}
	// The captured pawn was shielding the king from a rook on the rank.
	s.Board[7][4] = nil
	put(s, 3, 7, King, White)
	put(s, 3, 0, Rook, Black)
	if IsValidMove(s, Position{3, 4}, Position{2, 3}) {
		t.Fatal("en passant exposes own king")
	}
}

func TestPromotionAndFoolsMate(t *testing.T) {
	s := emptyState()
	put(s, 7, 4, King, White)
	put(s, 0, 4, King, Black)
	put(s, 1, 0, Pawn, White)
	p := MakeMove(s, Position{1, 0}, Position{0, 0}, Knight)
	if p.Board[0][0].Type != Knight || p.HalfMoveClock != 0 {
		t.Fatal("underpromotion failed")
	}
	s = CreateInitialState()
	for _, m := range [][2]Position{
		{{6, 5}, {5, 5}}, {{1, 4}, {3, 4}},
		{{6, 6}, {4, 6}}, {{0, 3}, {4, 7}},
	} {
		if !IsValidMove(s, m[0], m[1]) {
			t.Fatalf("opening move %+v not legal", m)
		}
		s = MakeMove(s, m[0], m[1], Queen)
	}
	if GetGameStatus(s) != Checkmate {
		t.Fatalf("fool's mate: %s", GetGameStatus(s))
	}
}

func TestDrawAndAI(t *testing.T) {
	s := emptyState()
	put(s, 7, 4, King, White)
	put(s, 0, 4, King, Black)
	if GetGameStatus(s) != Draw {
		t.Fatal("king vs king is draw")
	}

	put(s, 7, 0, Rook, White)
	s.HalfMoveClock = 100
	if GetGameStatus(s) != Draw {
		t.Fatal("fifty-move draw missing")
	}
	s = CreateInitialState()
	m := BestMove(s)
	if m == nil || !IsValidMove(s, m.From, m.To) {
		t.Fatalf("AI returned illegal move: %+v", m)
	}
}

func TestRepetitionKey(t *testing.T) {
	s := CreateInitialState()
	after := MakeMove(s, Position{6, 4}, Position{4, 4}, Queen)
	key, err := RepetitionKey(after)
	if err != nil {
		t.Fatal(err)
	}
	after.EnPassantTarget = nil
	without, err := RepetitionKey(after)
	if err != nil || key != without {
		t.Fatal("uncapturable en passant changes position key", err)
	}
	put(after, 4, 3, Pawn, Black)
	withoutPawnTarget, err := RepetitionKey(after)
	if err != nil {
		t.Fatal(err)
	}
	after.EnPassantTarget = &Position{5, 4}
	capturable, err := RepetitionKey(after)
	if err != nil || capturable == withoutPawnTarget {
		t.Fatal("capturable en passant missing from key", err)
	}
	after.Board[4][3].HasMoved = true
	again, err := RepetitionKey(after)
	if err != nil || capturable != again {
		t.Fatal("hasMoved changed position key", err)
	}
}
