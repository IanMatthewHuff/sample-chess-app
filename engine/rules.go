package main

import (
	"encoding/json"
	"math"
)

func attacked(state *BoardState, square Position, by PieceColor) bool {
	for _, entry := range state.Board.GetPiecesByColor(by) {
		p, from := entry.Piece, entry.Position
		dr, dc := square.Row-from.Row, square.Col-from.Col
		switch p.Type {
		case Pawn:
			direction := -1
			if by == Black {
				direction = 1
			}
			if dr == direction && abs(dc) == 1 {
				return true
			}
		case Knight:
			if (abs(dr) == 2 && abs(dc) == 1) || (abs(dr) == 1 && abs(dc) == 2) {
				return true
			}
		case King:
			if abs(dr) <= 1 && abs(dc) <= 1 {
				return true
			}
		default:
			straight := (dr == 0 || dc == 0) && (dr != 0 || dc != 0)
			diagonal := abs(dr) == abs(dc) && dr != 0
			if !((p.Type == Rook && straight) || (p.Type == Bishop && diagonal) || (p.Type == Queen && (straight || diagonal))) {
				continue
			}
			stepR, stepC := sign(dr), sign(dc)
			clear := true
			for r, c := from.Row+stepR, from.Col+stepC; r != square.Row || c != square.Col; r, c = r+stepR, c+stepC {
				if state.Board[r][c] != nil {
					clear = false
					break
				}
			}
			if clear {
				return true
			}
		}
	}
	return false
}

func IsCaptureMove(state *BoardState, from, to Position) bool {
	piece := state.Board.GetPieceAt(from)
	return state.Board.GetPieceAt(to) != nil ||
		(piece != nil && piece.Type == Pawn && from.Col != to.Col &&
			state.EnPassantTarget != nil && to == *state.EnPassantTarget)
}

func sign(n int) int {
	if n > 0 {
		return 1
	}
	if n < 0 {
		return -1
	}
	return 0
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func IsInCheck(state *BoardState, color PieceColor) bool {
	king := state.Board.FindKingPosition(color)
	return king == nil || attacked(state, *king, color.Opposite())
}

func GetValidMoves(state *BoardState, from Position) []Position {
	piece := state.Board.GetPieceAt(from)
	moves := make([]Position, 0)
	if piece == nil || piece.Color != state.CurrentTurn {
		return moves
	}
	add := func(to Position) {
		if !to.IsValid() {
			return
		}
		target := state.Board.GetPieceAt(to)
		if target != nil && (target.Color == piece.Color || target.Type == King) {
			return
		}
		if !IsInCheck(applyMove(state, from, to, Queen), piece.Color) {
			moves = append(moves, to)
		}
	}
	switch piece.Type {
	case Pawn:
		direction, start := -1, 6
		if piece.Color == Black {
			direction, start = 1, 1
		}
		one := Position{from.Row + direction, from.Col}
		if one.IsValid() && state.Board.GetPieceAt(one) == nil {
			add(one)
			two := Position{from.Row + 2*direction, from.Col}
			if from.Row == start && state.Board.GetPieceAt(two) == nil {
				add(two)
			}
		}
		for _, d := range []int{-1, 1} {
			to := Position{from.Row + direction, from.Col + d}
			if !to.IsValid() {
				continue
			}
			target := state.Board.GetPieceAt(to)
			if target != nil && target.Color != piece.Color {
				add(to)
			} else if target == nil && state.EnPassantTarget != nil && to == *state.EnPassantTarget {
				adjacent := state.Board.GetPieceAt(Position{from.Row, to.Col})
				if adjacent != nil && adjacent.Type == Pawn && adjacent.Color != piece.Color {
					add(to)
				}
			}
		}
	case Knight:
		for _, d := range [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}} {
			add(Position{from.Row + d[0], from.Col + d[1]})
		}
	case King:
		for dr := -1; dr <= 1; dr++ {
			for dc := -1; dc <= 1; dc++ {
				if dr != 0 || dc != 0 {
					add(Position{from.Row + dr, from.Col + dc})
				}
			}
		}
		home := 7
		if piece.Color == Black {
			home = 0
		}
		if from == (Position{home, 4}) && !IsInCheck(state, piece.Color) {
			for _, side := range []int{-1, 1} {
				rookCol, targetCol := 0, 2
				allowed := state.CastlingRights.WhiteQueenside
				if piece.Color == Black {
					allowed = state.CastlingRights.BlackQueenside
				}
				if side == 1 {
					rookCol, targetCol = 7, 6
					allowed = state.CastlingRights.WhiteKingside
					if piece.Color == Black {
						allowed = state.CastlingRights.BlackKingside
					}
				}
				rook := state.Board[home][rookCol]
				if !allowed || piece.HasMoved || rook == nil || rook.Color != piece.Color || rook.Type != Rook || rook.HasMoved {
					continue
				}
				clear := true
				for c := 4 + side; c != rookCol; c += side {
					if state.Board[home][c] != nil {
						clear = false
						break
					}
				}
				if clear && !attacked(state, Position{home, 4 + side}, piece.Color.Opposite()) &&
					!attacked(state, Position{home, targetCol}, piece.Color.Opposite()) {
					add(Position{home, targetCol})
				}
			}
		}
	default:
		directions := [][2]int{}
		if piece.Type == Bishop || piece.Type == Queen {
			directions = append(directions, [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}...)
		}
		if piece.Type == Rook || piece.Type == Queen {
			directions = append(directions, [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}...)
		}
		for _, d := range directions {
			for to := (Position{from.Row + d[0], from.Col + d[1]}); to.IsValid(); to = (Position{to.Row + d[0], to.Col + d[1]}) {
				add(to)
				if state.Board.GetPieceAt(to) != nil {
					break
				}
			}
		}
	}
	return moves
}

func IsValidMove(state *BoardState, from, to Position) bool {
	for _, move := range GetValidMoves(state, from) {
		if move == to {
			return true
		}
	}
	return false
}

func applyMove(state *BoardState, from, to Position, promotion PieceType) *BoardState {
	next := *state
	next.Board = state.Board.Clone()
	piece := *next.Board.GetPieceAt(from)
	captured := next.Board.GetPieceAt(to)
	rights := state.CastlingRights
	if piece.Type == King {
		if piece.Color == White {
			rights.WhiteKingside, rights.WhiteQueenside = false, false
		} else {
			rights.BlackKingside, rights.BlackQueenside = false, false
		}
	}
	if piece.Type == Rook {
		disableRookRight(&rights, from)
	}
	if captured != nil && captured.Type == Rook {
		disableRookRight(&rights, to)
	}
	if piece.Type == Pawn && from.Col != to.Col && captured == nil {
		next.Board[from.Row][to.Col] = nil
	}
	if piece.Type == King && abs(from.Col-to.Col) == 2 {
		rookFrom, rookTo := 0, 3
		if to.Col == 6 {
			rookFrom, rookTo = 7, 5
		}
		rook := *next.Board[from.Row][rookFrom]
		rook.HasMoved = true
		next.Board[from.Row][rookFrom] = nil
		next.Board[from.Row][rookTo] = &rook
	}
	piece.HasMoved = true
	if piece.Type == Pawn && (to.Row == 0 || to.Row == 7) {
		piece.Type = promotion
	}
	next.Board[to.Row][to.Col] = &piece
	next.Board[from.Row][from.Col] = nil
	next.CastlingRights = rights
	next.EnPassantTarget = nil
	if piece.Type == Pawn && abs(from.Row-to.Row) == 2 {
		next.EnPassantTarget = &Position{(from.Row + to.Row) / 2, from.Col}
	}
	next.HalfMoveClock++
	if captured != nil || state.Board[from.Row][from.Col].Type == Pawn {
		next.HalfMoveClock = 0
	}
	if state.CurrentTurn == Black {
		next.FullMoveNumber++
	}
	next.CurrentTurn = state.CurrentTurn.Opposite()
	return &next
}

func disableRookRight(rights *CastlingRights, pos Position) {
	switch pos {
	case (Position{7, 0}):
		rights.WhiteQueenside = false
	case (Position{7, 7}):
		rights.WhiteKingside = false
	case (Position{0, 0}):
		rights.BlackQueenside = false
	case (Position{0, 7}):
		rights.BlackKingside = false
	}
}

func MakeMove(state *BoardState, from, to Position, promotion PieceType) *BoardState {
	return applyMove(state, from, to, promotion)
}

type GameStatus string

const (
	Active    GameStatus = "active"
	Check     GameStatus = "check"
	Checkmate GameStatus = "checkmate"
	Stalemate GameStatus = "stalemate"
	Draw      GameStatus = "draw"
)

func GetGameStatus(state *BoardState) GameStatus {
	check := IsInCheck(state, state.CurrentTurn)
	for _, entry := range state.Board.GetPiecesByColor(state.CurrentTurn) {
		if len(GetValidMoves(state, entry.Position)) > 0 {
			if state.HalfMoveClock >= 100 || insufficientMaterial(state) {
				return Draw
			}

			if check {
				return Check
			}
			return Active
		}
	}
	if check {
		return Checkmate
	}
	return Stalemate
}

func RepetitionKey(state *BoardState) (string, error) {
	var board [8][8]struct {
		Type  PieceType
		Color PieceColor
	}
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			if piece := state.Board[row][col]; piece != nil {
				board[row][col].Type = piece.Type
				board[row][col].Color = piece.Color
			}
		}
	}
	var enPassant *Position
	if state.EnPassantTarget != nil {
		target := *state.EnPassantTarget
		fromRow := target.Row + 1
		if state.CurrentTurn == Black {
			fromRow = target.Row - 1
		}
		for _, col := range []int{target.Col - 1, target.Col + 1} {
			from := Position{fromRow, col}
			pawn := state.Board.GetPieceAt(from)
			if pawn != nil && pawn.Type == Pawn && pawn.Color == state.CurrentTurn &&
				IsValidMove(state, from, target) {
				enPassant = &target
				break
			}
		}
	}
	key, err := json.Marshal(struct {
		Board [8][8]struct {
			Type  PieceType
			Color PieceColor
		}
		Turn           PieceColor
		CastlingRights CastlingRights
		EnPassant      *Position
	}{board, state.CurrentTurn, state.CastlingRights, enPassant})
	return string(key), err
}

func insufficientMaterial(state *BoardState) bool {
	var minors []struct {
		piece *Piece
		pos   Position
	}
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			p := state.Board[row][col]
			if p == nil || p.Type == King {
				continue
			}
			if p.Type != Knight && p.Type != Bishop {
				return false
			}
			minors = append(minors, struct {
				piece *Piece
				pos   Position
			}{p, Position{row, col}})
		}
	}
	if len(minors) <= 1 {
		return true
	}
	if len(minors) == 2 && minors[0].piece.Type == Bishop && minors[1].piece.Type == Bishop &&
		minors[0].piece.Color != minors[1].piece.Color {
		return (minors[0].pos.Row+minors[0].pos.Col)%2 == (minors[1].pos.Row+minors[1].pos.Col)%2
	}
	return false
}

type EngineMove struct {
	From      Position  `json:"from"`
	To        Position  `json:"to"`
	Promotion PieceType `json:"promotion,omitempty"`
}

func BestMove(state *BoardState) *EngineMove {
	if GetGameStatus(state) != Active && GetGameStatus(state) != Check {
		return nil
	}
	var best *EngineMove
	bestScore := math.MinInt
	for _, entry := range state.Board.GetPiecesByColor(state.CurrentTurn) {
		for _, to := range GetValidMoves(state, entry.Position) {
			target := state.Board.GetPieceAt(to)
			score := 0
			if target != nil {
				score += PieceValue(target.Type) * 100
			}
			if entry.Piece.Type == Pawn && (to.Row == 0 || to.Row == 7) {
				score += 800
			}
			next := applyMove(state, entry.Position, to, Queen)
			if IsInCheck(next, next.CurrentTurn) {
				score += 15
				if GetGameStatus(next) == Checkmate {
					score += 100000
				}
			}
			// Prefer central squares, then stable board-order tie breaking.
			score += 6 - abs(3-to.Row) - abs(3-to.Col)
			if score > bestScore {
				bestScore = score
				best = &EngineMove{From: entry.Position, To: to, Promotion: Queen}
			}
		}
	}
	return best
}
