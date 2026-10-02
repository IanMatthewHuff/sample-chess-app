package main

import (
	"errors"
)

// ErrInvalidMove is returned when a requested move is not legal
var ErrInvalidMove = errors.New("invalid move")

var (
	knightOffsets    = [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}
	kingOffsets      = [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}
	bishopDirections = [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	rookDirections   = [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	queenDirections  = kingOffsets
	promotionPieces  = []PieceType{Queen, Rook, Bishop, Knight}
)

// pawnDirection returns the row delta a pawn of the given color moves in
func pawnDirection(c PieceColor) int {
	if c == White {
		return -1
	}
	return 1
}

// homeRow returns the back rank row for a color
func homeRow(c PieceColor) int {
	if c == White {
		return 7
	}
	return 0
}

// GeneratePseudoLegalMoves returns moves for the piece at pos, ignoring whether
// they leave the mover's king in check.
func GeneratePseudoLegalMoves(state *BoardState, pos Position) []Move {
	piece := state.Board.GetPieceAt(pos)
	if piece == nil {
		return nil
	}

	var moves []Move
	add := func(to Position) {
		moves = append(moves, Move{From: pos, To: to, Piece: piece, Captured: state.Board.GetPieceAt(to)})
	}

	switch piece.Type {
	case Pawn:
		moves = pawnMoves(state, pos, piece)
	case Knight:
		for _, to := range stepTargets(&state.Board, pos, piece.Color, knightOffsets) {
			add(to)
		}
	case Bishop:
		for _, to := range slideTargets(&state.Board, pos, piece.Color, bishopDirections) {
			add(to)
		}
	case Rook:
		for _, to := range slideTargets(&state.Board, pos, piece.Color, rookDirections) {
			add(to)
		}
	case Queen:
		for _, to := range slideTargets(&state.Board, pos, piece.Color, queenDirections) {
			add(to)
		}
	case King:
		for _, to := range stepTargets(&state.Board, pos, piece.Color, kingOffsets) {
			add(to)
		}
		moves = append(moves, castlingMoves(state, pos, piece)...)
	}
	return moves
}

func pawnMoves(state *BoardState, pos Position, piece *Piece) []Move {
	var moves []Move
	dir := pawnDirection(piece.Color)
	startRow := homeRow(piece.Color) + dir
	lastRow := homeRow(piece.Color.Opposite())

	addPawnMove := func(to Position, captured *Piece, enPassant bool) {
		if to.Row == lastRow {
			for _, pt := range promotionPieces {
				moves = append(moves, Move{From: pos, To: to, Piece: piece, Captured: captured, IsPromotion: true, PromotedTo: pt})
			}
			return
		}
		moves = append(moves, Move{From: pos, To: to, Piece: piece, Captured: captured, IsEnPassant: enPassant})
	}

	one := Position{Row: pos.Row + dir, Col: pos.Col}
	if one.IsValid() && state.Board.GetPieceAt(one) == nil {
		addPawnMove(one, nil, false)
		two := Position{Row: pos.Row + 2*dir, Col: pos.Col}
		if pos.Row == startRow && state.Board.GetPieceAt(two) == nil {
			addPawnMove(two, nil, false)
		}
	}

	for _, dc := range []int{-1, 1} {
		to := Position{Row: pos.Row + dir, Col: pos.Col + dc}
		if !to.IsValid() {
			continue
		}
		if target := state.Board.GetPieceAt(to); target != nil {
			if target.Color != piece.Color {
				addPawnMove(to, target, false)
			}
			continue
		}
		if ep := state.EnPassantTarget; ep != nil && ep.Equals(to) {
			victim := state.Board.GetPieceAt(Position{Row: pos.Row, Col: to.Col})
			if victim != nil && victim.Type == Pawn && victim.Color != piece.Color {
				addPawnMove(to, victim, true)
			}
		}
	}
	return moves
}

func castlingMoves(state *BoardState, pos Position, king *Piece) []Move {
	row := homeRow(king.Color)
	if pos.Row != row || pos.Col != 4 {
		return nil
	}
	opponent := king.Color.Opposite()
	if IsSquareAttacked(&state.Board, pos, opponent) {
		return nil
	}

	kingside, queenside := state.CastlingRights.WhiteKingside, state.CastlingRights.WhiteQueenside
	if king.Color == Black {
		kingside, queenside = state.CastlingRights.BlackKingside, state.CastlingRights.BlackQueenside
	}

	var moves []Move
	try := func(allowed bool, rookCol int, empty []int, safe []int, kingTo int) {
		if !allowed {
			return
		}
		rook := state.Board[row][rookCol]
		if rook == nil || rook.Type != Rook || rook.Color != king.Color {
			return
		}
		for _, c := range empty {
			if state.Board[row][c] != nil {
				return
			}
		}
		for _, c := range safe {
			if IsSquareAttacked(&state.Board, Position{Row: row, Col: c}, opponent) {
				return
			}
		}
		moves = append(moves, Move{From: pos, To: Position{Row: row, Col: kingTo}, Piece: king, IsCastling: true})
	}
	try(kingside, 7, []int{5, 6}, []int{5, 6}, 6)
	try(queenside, 0, []int{1, 2, 3}, []int{2, 3}, 2)
	return moves
}

func stepTargets(board *Board, pos Position, color PieceColor, offsets [][2]int) []Position {
	var targets []Position
	for _, o := range offsets {
		to := Position{Row: pos.Row + o[0], Col: pos.Col + o[1]}
		if !to.IsValid() {
			continue
		}
		if t := board.GetPieceAt(to); t == nil || t.Color != color {
			targets = append(targets, to)
		}
	}
	return targets
}

func slideTargets(board *Board, pos Position, color PieceColor, directions [][2]int) []Position {
	var targets []Position
	for _, d := range directions {
		to := Position{Row: pos.Row + d[0], Col: pos.Col + d[1]}
		for to.IsValid() {
			t := board.GetPieceAt(to)
			if t == nil {
				targets = append(targets, to)
			} else {
				if t.Color != color {
					targets = append(targets, to)
				}
				break
			}
			to = Position{Row: to.Row + d[0], Col: to.Col + d[1]}
		}
	}
	return targets
}

// IsSquareAttacked reports whether any piece of color `by` attacks pos
func IsSquareAttacked(board *Board, pos Position, by PieceColor) bool {
	// Pawns: a pawn of color `by` attacks diagonally forward, so look one row behind.
	pawnRow := pos.Row - pawnDirection(by)
	for _, dc := range []int{-1, 1} {
		p := board.GetPieceAt(Position{Row: pawnRow, Col: pos.Col + dc})
		if p != nil && p.Color == by && p.Type == Pawn {
			return true
		}
	}
	for _, o := range knightOffsets {
		p := board.GetPieceAt(Position{Row: pos.Row + o[0], Col: pos.Col + o[1]})
		if p != nil && p.Color == by && p.Type == Knight {
			return true
		}
	}
	for _, o := range kingOffsets {
		p := board.GetPieceAt(Position{Row: pos.Row + o[0], Col: pos.Col + o[1]})
		if p != nil && p.Color == by && p.Type == King {
			return true
		}
	}
	rayHit := func(dirs [][2]int, a, b PieceType) bool {
		for _, d := range dirs {
			sq := Position{Row: pos.Row + d[0], Col: pos.Col + d[1]}
			for sq.IsValid() {
				if p := board.GetPieceAt(sq); p != nil {
					if p.Color == by && (p.Type == a || p.Type == b) {
						return true
					}
					break
				}
				sq = Position{Row: sq.Row + d[0], Col: sq.Col + d[1]}
			}
		}
		return false
	}
	return rayHit(rookDirections, Rook, Queen) || rayHit(bishopDirections, Bishop, Queen)
}

// IsInCheck checks if a player is in check
func IsInCheck(state *BoardState, color PieceColor) bool {
	kingPos := state.Board.FindKingPosition(color)
	if kingPos == nil {
		return false
	}
	return IsSquareAttacked(&state.Board, *kingPos, color.Opposite())
}

// GetLegalMovesFrom returns all legal moves for the side to move from pos
func GetLegalMovesFrom(state *BoardState, pos Position) []Move {
	piece := state.Board.GetPieceAt(pos)
	if piece == nil || piece.Color != state.CurrentTurn {
		return nil
	}
	var legal []Move
	for _, m := range GeneratePseudoLegalMoves(state, pos) {
		if !IsInCheck(ApplyMove(state, m), piece.Color) {
			legal = append(legal, m)
		}
	}
	return legal
}

// GenerateLegalMoves returns every legal move for the side to move
func GenerateLegalMoves(state *BoardState) []Move {
	var legal []Move
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			p := state.Board[row][col]
			if p != nil && p.Color == state.CurrentTurn {
				legal = append(legal, GetLegalMovesFrom(state, Position{Row: row, Col: col})...)
			}
		}
	}
	return legal
}

// HasLegalMove reports whether the side to move has at least one legal move
func HasLegalMove(state *BoardState) bool {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			p := state.Board[row][col]
			if p != nil && p.Color == state.CurrentTurn && len(GetLegalMovesFrom(state, Position{Row: row, Col: col})) > 0 {
				return true
			}
		}
	}
	return false
}

// GetValidMoves returns the distinct destination squares for the piece at pos
func GetValidMoves(state *BoardState, pos Position) []Position {
	var targets []Position
	seen := map[Position]bool{}
	for _, m := range GetLegalMovesFrom(state, pos) {
		if !seen[m.To] {
			seen[m.To] = true
			targets = append(targets, m.To)
		}
	}
	return targets
}

// FindLegalMove finds the legal move matching from/to. For promotions an empty
// promotion type defaults to a queen.
func FindLegalMove(state *BoardState, from, to Position, promotion PieceType) (Move, error) {
	if promotion == "" {
		promotion = Queen
	}
	for _, m := range GetLegalMovesFrom(state, from) {
		if !m.To.Equals(to) {
			continue
		}
		if m.IsPromotion && m.PromotedTo != promotion {
			continue
		}
		return m, nil
	}
	return Move{}, ErrInvalidMove
}

// IsValidMove checks if a move is legal
func IsValidMove(state *BoardState, from Position, to Position) bool {
	_, err := FindLegalMove(state, from, to, "")
	return err == nil
}

// ApplyMove executes a (pseudo-legal) move and returns the resulting state.
// The input state is not modified.
func ApplyMove(state *BoardState, m Move) *BoardState {
	board := state.Board.Clone()
	piece := board.GetPieceAt(m.From)
	if piece == nil {
		return state
	}

	moved := &Piece{Type: piece.Type, Color: piece.Color, HasMoved: true}
	if m.IsPromotion {
		promo := m.PromotedTo
		if !IsValidPromotion(promo) {
			promo = Queen
		}
		moved.Type = promo
	}
	board.SetPieceAt(m.From, nil)
	board.SetPieceAt(m.To, moved)

	if m.IsEnPassant {
		board.SetPieceAt(Position{Row: m.From.Row, Col: m.To.Col}, nil)
	}
	if m.IsCastling {
		row := m.From.Row
		rookFrom, rookTo := 7, 5
		if m.To.Col == 2 {
			rookFrom, rookTo = 0, 3
		}
		if rook := board[row][rookFrom]; rook != nil {
			board[row][rookTo] = &Piece{Type: Rook, Color: rook.Color, HasMoved: true}
			board[row][rookFrom] = nil
		}
	}

	rights := state.CastlingRights
	if piece.Type == King {
		if piece.Color == White {
			rights.WhiteKingside, rights.WhiteQueenside = false, false
		} else {
			rights.BlackKingside, rights.BlackQueenside = false, false
		}
	}
	for _, sq := range []Position{m.From, m.To} {
		switch {
		case sq.Row == 7 && sq.Col == 7:
			rights.WhiteKingside = false
		case sq.Row == 7 && sq.Col == 0:
			rights.WhiteQueenside = false
		case sq.Row == 0 && sq.Col == 7:
			rights.BlackKingside = false
		case sq.Row == 0 && sq.Col == 0:
			rights.BlackQueenside = false
		}
	}

	var enPassant *Position
	if piece.Type == Pawn && abs(m.From.Row-m.To.Row) == 2 {
		enPassant = &Position{Row: (m.From.Row + m.To.Row) / 2, Col: m.From.Col}
	}

	halfMove := state.HalfMoveClock + 1
	if piece.Type == Pawn || m.Captured != nil || state.Board.GetPieceAt(m.To) != nil {
		halfMove = 0
	}
	fullMove := state.FullMoveNumber
	if state.CurrentTurn == Black {
		fullMove++
	}

	return &BoardState{
		Board:           board,
		CurrentTurn:     state.CurrentTurn.Opposite(),
		CastlingRights:  rights,
		EnPassantTarget: enPassant,
		HalfMoveClock:   halfMove,
		FullMoveNumber:  fullMove,
	}
}

// MakeMove executes a legal move (auto-promoting to a queen) and returns the new state.
// If the move is illegal the original state is returned.
func MakeMove(state *BoardState, from Position, to Position) *BoardState {
	m, err := FindLegalMove(state, from, to, "")
	if err != nil {
		return state
	}
	return ApplyMove(state, m)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
