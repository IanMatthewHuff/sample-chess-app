package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpeningNotationAndMate(t *testing.T) {
	s := CreateInitialState()
	opening := []struct {
		from, to Position
		want     string
	}{
		{Position{6, 5}, Position{5, 5}, "f3"},
		{Position{1, 4}, Position{3, 4}, "e5"},
		{Position{6, 6}, Position{4, 6}, "g4"},
		{Position{0, 3}, Position{4, 7}, "Qh4#"},
	}
	for _, test := range opening {
		if got := MoveSAN(s, test.from, test.to, Queen); got != test.want {
			t.Errorf("SAN: got %s, want %s", got, test.want)
		}
		s = MakeMove(s, test.from, test.to, Queen)
	}
}

func TestCaptureCastlePromotionAndCheckNotation(t *testing.T) {
	s := CreateInitialState()
	s = MakeMove(s, Position{6, 4}, Position{4, 4}, Queen)
	s = MakeMove(s, Position{1, 3}, Position{3, 3}, Queen)
	if got := MoveSAN(s, Position{4, 4}, Position{3, 3}, Queen); got != "exd5" {
		t.Errorf("pawn capture: %s", got)
	}

	s = emptyState()
	put(s, 7, 4, King, White)
	put(s, 7, 7, Rook, White)
	put(s, 0, 4, King, Black)
	s.CastlingRights.WhiteKingside = true
	if got := MoveSAN(s, Position{7, 4}, Position{7, 6}, Queen); got != "O-O" {
		t.Errorf("castling: %s", got)
	}
	put(s, 1, 0, Pawn, White)
	if got := MoveSAN(s, Position{1, 0}, Position{0, 0}, Knight); got != "a8=N" {
		t.Errorf("underpromotion: %s", got)
	}

	s = emptyState()
	put(s, 7, 4, King, White)
	put(s, 0, 4, King, Black)
	put(s, 3, 4, Pawn, White)
	put(s, 3, 3, Pawn, Black)
	s.EnPassantTarget = &Position{2, 3}
	if got := MoveSAN(s, Position{3, 4}, Position{2, 3}, Queen); got != "exd6" {
		t.Errorf("en passant: %s", got)
	}
	put(s, 7, 3, Queen, White)
	if got := MoveSAN(s, Position{7, 3}, Position{3, 7}, Queen); got != "Qh5+" {
		t.Errorf("check: %s", got)
	}
}

func TestSANDisambiguation(t *testing.T) {
	s := emptyState()
	put(s, 7, 4, King, White)
	put(s, 0, 4, King, Black)
	put(s, 7, 1, Knight, White)
	put(s, 7, 5, Knight, White)
	if got := MoveSAN(s, Position{7, 1}, Position{6, 3}, Queen); got != "Nbd2" {
		t.Errorf("file disambiguation: %s", got)
	}
	s.Board[7][5] = nil
	put(s, 5, 1, Knight, White)
	if got := MoveSAN(s, Position{7, 1}, Position{6, 3}, Queen); got != "N1d2" {
		t.Errorf("rank disambiguation: %s", got)
	}
	put(s, 7, 5, Knight, White)
	if got := MoveSAN(s, Position{7, 1}, Position{6, 3}, Queen); got != "Nb1d2" {
		t.Errorf("full disambiguation: %s", got)
	}
}

func TestMoveWithNotationEndpoint(t *testing.T) {
	request, err := json.Marshal(ValidationRequest{
		Board: *CreateInitialState(), From: Position{6, 4}, To: Position{4, 4},
	})
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handleMoveWithNotation(recorder, httptest.NewRequest(http.MethodPost, "/move-san", bytes.NewReader(request)))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body.String())
	}
	var response MoveWithNotationResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.SAN != "e4" || response.State.CurrentTurn != Black {
		t.Fatalf("unexpected move response: %+v", response)
	}
}

func TestCaptureMovesEndpoint(t *testing.T) {
	state := emptyState()
	put(state, 7, 4, King, White)
	put(state, 0, 4, King, Black)
	put(state, 3, 4, Pawn, White)
	put(state, 3, 3, Pawn, Black)
	state.EnPassantTarget = &Position{2, 3}
	payload, err := json.Marshal(ValidMovesRequest{Board: *state, Position: Position{3, 4}})
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handleValidMoves(recorder, httptest.NewRequest(http.MethodPost, "/moves", bytes.NewReader(payload)))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body.String())
	}
	var response ValidMovesResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Captures) != 1 || response.Captures[0] != (Position{2, 3}) {
		t.Fatalf("en passant capture missing: %+v", response)
	}
	if len(response.Moves) != 2 {
		t.Fatalf("expected forward and capture moves: %+v", response)
	}
}
