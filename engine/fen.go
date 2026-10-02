package main

import (
	"fmt"
	"strconv"
	"strings"
)

var fenPieces = map[byte]PieceType{
	'k': King, 'q': Queen, 'r': Rook, 'b': Bishop, 'n': Knight, 'p': Pawn,
}

// ParseFEN builds a board state from a FEN string
func ParseFEN(fen string) (*BoardState, error) {
	parts := strings.Fields(fen)
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid FEN: expected at least 4 fields")
	}

	var board Board
	rows := strings.Split(parts[0], "/")
	if len(rows) != 8 {
		return nil, fmt.Errorf("invalid FEN: expected 8 ranks")
	}
	for r, rowStr := range rows {
		col := 0
		for i := 0; i < len(rowStr); i++ {
			ch := rowStr[i]
			if ch >= '1' && ch <= '8' {
				col += int(ch - '0')
				continue
			}
			lower := ch | 0x20
			pt, ok := fenPieces[lower]
			if !ok || col > 7 {
				return nil, fmt.Errorf("invalid FEN: bad rank %q", rowStr)
			}
			color := Black
			if ch != lower {
				color = White
			}
			hasMoved := pt == Pawn && r != homeRow(color)+pawnDirection(color)
			board[r][col] = &Piece{Type: pt, Color: color, HasMoved: hasMoved}
			col++
		}
		if col != 8 {
			return nil, fmt.Errorf("invalid FEN: rank %q does not have 8 squares", rowStr)
		}
	}

	state := &BoardState{Board: board, FullMoveNumber: 1}
	switch parts[1] {
	case "w":
		state.CurrentTurn = White
	case "b":
		state.CurrentTurn = Black
	default:
		return nil, fmt.Errorf("invalid FEN: bad active color %q", parts[1])
	}

	state.CastlingRights = CastlingRights{
		WhiteKingside:  strings.Contains(parts[2], "K"),
		WhiteQueenside: strings.Contains(parts[2], "Q"),
		BlackKingside:  strings.Contains(parts[2], "k"),
		BlackQueenside: strings.Contains(parts[2], "q"),
	}

	if parts[3] != "-" {
		ep, ok := PositionFromAlgebraic(parts[3])
		if !ok {
			return nil, fmt.Errorf("invalid FEN: bad en passant square %q", parts[3])
		}
		state.EnPassantTarget = &ep
	}

	if len(parts) >= 6 {
		hm, err1 := strconv.Atoi(parts[4])
		fm, err2 := strconv.Atoi(parts[5])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("invalid FEN: bad move counters")
		}
		state.HalfMoveClock, state.FullMoveNumber = hm, fm
	}
	return state, nil
}

func placementFEN(b *Board) string {
	var sb strings.Builder
	for row := 0; row < 8; row++ {
		empty := 0
		for col := 0; col < 8; col++ {
			p := b[row][col]
			if p == nil {
				empty++
				continue
			}
			if empty > 0 {
				sb.WriteByte(byte('0' + empty))
				empty = 0
			}
			letter := PieceLetter(p.Type)
			if p.Type == Pawn {
				letter = "P"
			}
			if p.Color == Black {
				letter = strings.ToLower(letter)
			}
			sb.WriteString(letter)
		}
		if empty > 0 {
			sb.WriteByte(byte('0' + empty))
		}
		if row < 7 {
			sb.WriteByte('/')
		}
	}
	return sb.String()
}

func castlingFEN(c CastlingRights) string {
	s := ""
	if c.WhiteKingside {
		s += "K"
	}
	if c.WhiteQueenside {
		s += "Q"
	}
	if c.BlackKingside {
		s += "k"
	}
	if c.BlackQueenside {
		s += "q"
	}
	if s == "" {
		return "-"
	}
	return s
}

func turnFEN(c PieceColor) string {
	if c == Black {
		return "b"
	}
	return "w"
}

// ToFEN converts a board state to a FEN string
func ToFEN(state *BoardState) string {
	ep := "-"
	if state.EnPassantTarget != nil {
		ep = state.EnPassantTarget.ToAlgebraic()
	}
	return fmt.Sprintf("%s %s %s %s %d %d",
		placementFEN(&state.Board), turnFEN(state.CurrentTurn), castlingFEN(state.CastlingRights),
		ep, state.HalfMoveClock, state.FullMoveNumber)
}

// PositionKey identifies a position for repetition detection. The en passant
// square is only included when an en passant capture is actually legal.
func PositionKey(state *BoardState) string {
	ep := "-"
	if state.EnPassantTarget != nil {
		for _, m := range GenerateLegalMoves(state) {
			if m.IsEnPassant {
				ep = state.EnPassantTarget.ToAlgebraic()
				break
			}
		}
	}
	return placementFEN(&state.Board) + " " + turnFEN(state.CurrentTurn) + " " + castlingFEN(state.CastlingRights) + " " + ep
}
