package main

import "strings"

// GameStatus represents the current game status
type GameStatus string

const (
	Active    GameStatus = "active"
	Check     GameStatus = "check"
	Checkmate GameStatus = "checkmate"
	Stalemate GameStatus = "stalemate"
	Draw      GameStatus = "draw"
)

// Draw reasons
const (
	ReasonFiftyMove            = "fifty-move rule"
	ReasonThreefoldRepetition  = "threefold repetition"
	ReasonInsufficientMaterial = "insufficient material"
)

// StatusResult is the full status of a position
type StatusResult struct {
	Status   GameStatus `json:"status"`
	InCheck  bool       `json:"inCheck"`
	GameOver bool       `json:"gameOver"`
	Winner   PieceColor `json:"winner,omitempty"`
	Reason   string     `json:"reason,omitempty"`
}

// EvaluateStatus determines the game status. history contains the position keys
// of every position reached in the game (including the current one) and is
// used for threefold repetition detection; it may be empty.
func EvaluateStatus(state *BoardState, history []string) StatusResult {
	inCheck := IsInCheck(state, state.CurrentTurn)

	if !HasLegalMove(state) {
		if inCheck {
			return StatusResult{Status: Checkmate, InCheck: true, GameOver: true, Winner: state.CurrentTurn.Opposite(), Reason: "checkmate"}
		}
		return StatusResult{Status: Stalemate, GameOver: true, Reason: "stalemate"}
	}
	if IsInsufficientMaterial(&state.Board) {
		return StatusResult{Status: Draw, InCheck: inCheck, GameOver: true, Reason: ReasonInsufficientMaterial}
	}
	if state.HalfMoveClock >= 100 {
		return StatusResult{Status: Draw, InCheck: inCheck, GameOver: true, Reason: ReasonFiftyMove}
	}
	if len(history) > 0 {
		key := PositionKey(state)
		count := 0
		for _, h := range history {
			if h == key {
				count++
			}
		}
		if count >= 3 {
			return StatusResult{Status: Draw, InCheck: inCheck, GameOver: true, Reason: ReasonThreefoldRepetition}
		}
	}
	if inCheck {
		return StatusResult{Status: Check, InCheck: true}
	}
	return StatusResult{Status: Active}
}

// GetGameStatus determines the current game status (without repetition history)
func GetGameStatus(state *BoardState) GameStatus {
	return EvaluateStatus(state, nil).Status
}

// IsInsufficientMaterial reports whether neither side can possibly checkmate:
// K v K, K+minor v K, or kings with bishops all on the same square color.
func IsInsufficientMaterial(b *Board) bool {
	var minors []struct {
		t           PieceType
		lightSquare bool
	}
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			p := b[row][col]
			if p == nil || p.Type == King {
				continue
			}
			if p.Type == Pawn || p.Type == Rook || p.Type == Queen {
				return false
			}
			minors = append(minors, struct {
				t           PieceType
				lightSquare bool
			}{p.Type, (row+col)%2 == 0})
		}
	}
	if len(minors) <= 1 {
		return true
	}
	for _, m := range minors {
		if m.t != Bishop || m.lightSquare != minors[0].lightSquare {
			return false
		}
	}
	return true
}

// MoveToSAN converts a legal move to Standard Algebraic Notation
func MoveToSAN(state *BoardState, m Move) string {
	var sb strings.Builder
	if m.IsCastling {
		if m.To.Col == 6 {
			sb.WriteString("O-O")
		} else {
			sb.WriteString("O-O-O")
		}
	} else if m.Piece.Type == Pawn {
		if m.Captured != nil {
			sb.WriteByte("abcdefgh"[m.From.Col])
			sb.WriteByte('x')
		}
		sb.WriteString(m.To.ToAlgebraic())
		if m.IsPromotion {
			sb.WriteString("=" + PieceLetter(m.PromotedTo))
		}
	} else {
		sb.WriteString(PieceLetter(m.Piece.Type))
		sb.WriteString(disambiguation(state, m))
		if m.Captured != nil {
			sb.WriteByte('x')
		}
		sb.WriteString(m.To.ToAlgebraic())
	}

	next := ApplyMove(state, m)
	if IsInCheck(next, next.CurrentTurn) {
		if HasLegalMove(next) {
			sb.WriteByte('+')
		} else {
			sb.WriteByte('#')
		}
	}
	return sb.String()
}

func disambiguation(state *BoardState, m Move) string {
	sameFile, sameRank, ambiguous := false, false, false
	for _, other := range GenerateLegalMoves(state) {
		if other.From.Equals(m.From) || !other.To.Equals(m.To) || other.Piece.Type != m.Piece.Type {
			continue
		}
		ambiguous = true
		if other.From.Col == m.From.Col {
			sameFile = true
		}
		if other.From.Row == m.From.Row {
			sameRank = true
		}
	}
	square := m.From.ToAlgebraic()
	switch {
	case !ambiguous:
		return ""
	case !sameFile:
		return square[:1]
	case !sameRank:
		return square[1:]
	default:
		return square
	}
}
