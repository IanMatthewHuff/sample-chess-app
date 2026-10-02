package main

// GetValidMoves returns all valid moves for a piece at the given position
func GetValidMoves(state *BoardState, pos Position) []Position {
	// TODO: Implement valid move calculation
	piece := state.Board.GetPieceAt(pos)
	if piece == nil || piece.Color != state.CurrentTurn {
		return nil
	}

	var moves []Position

	switch piece.Type {
	case Pawn:
		moves = GetPawnMoves(state, pos, piece)
	case Knight:
		moves = GetKnightMoves(state, pos, piece)
	case Bishop:
		moves = GetBishopMoves(state, pos, piece)
	case Rook:
		moves = GetRookMoves(state, pos, piece)
	case Queen:
		moves = GetQueenMoves(state, pos, piece)
	case King:
		moves = GetKingMoves(state, pos, piece)
	}

	// Filter out moves that would leave king in check
	var validMoves []Position
	for _, move := range moves {
		if !WouldBeInCheck(state, pos, move) {
			validMoves = append(validMoves, move)
		}
	}

	return validMoves
}

// GetPawnMoves returns all possible pawn moves
func GetPawnMoves(state *BoardState, pos Position, piece *Piece) []Position {
	// TODO: Implement pawn move logic
	var moves []Position
	direction := -1
	startRow := 6
	if piece.Color == Black {
		direction = 1
		startRow = 1
	}

	// Single move forward
	oneForward := Position{Row: pos.Row + direction, Col: pos.Col}
	if oneForward.IsValid() && state.Board.GetPieceAt(oneForward) == nil {
		moves = append(moves, oneForward)

		// Double move from starting position
		if pos.Row == startRow {
			twoForward := Position{Row: pos.Row + 2*direction, Col: pos.Col}
			if state.Board.GetPieceAt(twoForward) == nil {
				moves = append(moves, twoForward)
			}
		}
	}

	// Captures
	for _, colOffset := range []int{-1, 1} {
		capturePos := Position{Row: pos.Row + direction, Col: pos.Col + colOffset}
		if capturePos.IsValid() {
			target := state.Board.GetPieceAt(capturePos)
			if target != nil && target.Color != piece.Color {
				moves = append(moves, capturePos)
			}

			// En passant
			if state.EnPassantTarget != nil &&
				capturePos.Row == state.EnPassantTarget.Row &&
				capturePos.Col == state.EnPassantTarget.Col {
				moves = append(moves, capturePos)
			}
		}
	}

	return moves
}

// GetKnightMoves returns all possible knight moves
func GetKnightMoves(state *BoardState, pos Position, piece *Piece) []Position {
	// TODO: Implement knight move logic
	var moves []Position
	offsets := [][2]int{
		{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
		{1, -2}, {1, 2}, {2, -1}, {2, 1},
	}

	for _, offset := range offsets {
		newPos := Position{Row: pos.Row + offset[0], Col: pos.Col + offset[1]}
		if newPos.IsValid() {
			target := state.Board.GetPieceAt(newPos)
			if target == nil || target.Color != piece.Color {
				moves = append(moves, newPos)
			}
		}
	}

	return moves
}

// GetBishopMoves returns all possible bishop moves
func GetBishopMoves(state *BoardState, pos Position, piece *Piece) []Position {
	// TODO: Implement bishop move logic
	directions := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	return GetSlidingMoves(state, pos, piece, directions)
}

// GetRookMoves returns all possible rook moves
func GetRookMoves(state *BoardState, pos Position, piece *Piece) []Position {
	// TODO: Implement rook move logic
	directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	return GetSlidingMoves(state, pos, piece, directions)
}

// GetQueenMoves returns all possible queen moves
func GetQueenMoves(state *BoardState, pos Position, piece *Piece) []Position {
	// TODO: Implement queen move logic
	directions := [][2]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}
	return GetSlidingMoves(state, pos, piece, directions)
}

// GetKingMoves returns all possible king moves (including castling)
func GetKingMoves(state *BoardState, pos Position, piece *Piece) []Position {
	// TODO: Implement king move logic
	var moves []Position
	offsets := [][2]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	for _, offset := range offsets {
		newPos := Position{Row: pos.Row + offset[0], Col: pos.Col + offset[1]}
		if newPos.IsValid() {
			target := state.Board.GetPieceAt(newPos)
			if target == nil || target.Color != piece.Color {
				moves = append(moves, newPos)
			}
		}
	}

	// TODO: Add castling logic

	return moves
}

// GetSlidingMoves returns moves for sliding pieces (bishop, rook, queen)
func GetSlidingMoves(state *BoardState, pos Position, piece *Piece, directions [][2]int) []Position {
	// TODO: Implement sliding move logic
	var moves []Position

	for _, dir := range directions {
		newPos := Position{Row: pos.Row + dir[0], Col: pos.Col + dir[1]}

		for newPos.IsValid() {
			target := state.Board.GetPieceAt(newPos)

			if target == nil {
				moves = append(moves, Position{Row: newPos.Row, Col: newPos.Col})
			} else {
				if target.Color != piece.Color {
					moves = append(moves, Position{Row: newPos.Row, Col: newPos.Col})
				}
				break
			}

			newPos = Position{Row: newPos.Row + dir[0], Col: newPos.Col + dir[1]}
		}
	}

	return moves
}

// WouldBeInCheck checks if making a move would leave the king in check
func WouldBeInCheck(state *BoardState, from Position, to Position) bool {
	// TODO: Implement check detection after move
	testBoard := state.Board.Clone()
	piece := testBoard.GetPieceAt(from)

	if piece == nil {
		return true
	}

	testBoard.SetPieceAt(to, piece)
	testBoard.SetPieceAt(from, nil)

	testState := &BoardState{
		Board:           testBoard,
		CurrentTurn:     state.CurrentTurn,
		CastlingRights:  state.CastlingRights,
		EnPassantTarget: state.EnPassantTarget,
	}

	return IsInCheck(testState, piece.Color)
}

// IsInCheck checks if a player is in check
func IsInCheck(state *BoardState, color PieceColor) bool {
	// TODO: Implement check detection
	kingPos := state.Board.FindKingPosition(color)
	if kingPos == nil {
		return false
	}

	opponentColor := color.Opposite()

	// Check if any opponent piece can attack the king
	for _, p := range state.Board.GetPiecesByColor(opponentColor) {
		attackSquares := GetAttackSquares(state, p.Position, p.Piece)
		for _, sq := range attackSquares {
			if sq.Row == kingPos.Row && sq.Col == kingPos.Col {
				return true
			}
		}
	}

	return false
}

// GetAttackSquares returns squares that a piece can attack (ignoring check)
func GetAttackSquares(state *BoardState, pos Position, piece *Piece) []Position {
	// TODO: Implement attack square calculation
	switch piece.Type {
	case Pawn:
		// Pawns attack diagonally
		direction := -1
		if piece.Color == Black {
			direction = 1
		}
		var attacks []Position
		for _, colOffset := range []int{-1, 1} {
			attackPos := Position{Row: pos.Row + direction, Col: pos.Col + colOffset}
			if attackPos.IsValid() {
				attacks = append(attacks, attackPos)
			}
		}
		return attacks
	case Knight:
		return GetKnightMoves(state, pos, piece)
	case Bishop:
		return GetBishopMoves(state, pos, piece)
	case Rook:
		return GetRookMoves(state, pos, piece)
	case Queen:
		return GetQueenMoves(state, pos, piece)
	case King:
		return GetKingMoves(state, pos, piece)
	}
	return nil
}

// IsValidMove checks if a move is valid
func IsValidMove(state *BoardState, from Position, to Position) bool {
	// TODO: Implement move validation
	validMoves := GetValidMoves(state, from)
	for _, m := range validMoves {
		if m.Row == to.Row && m.Col == to.Col {
			return true
		}
	}
	return false
}

// MakeMove executes a move and returns the new board state
func MakeMove(state *BoardState, from Position, to Position) *BoardState {
	// TODO: Implement move execution
	piece := state.Board.GetPieceAt(from)
	if piece == nil {
		return state
	}

	newBoard := state.Board.Clone()
	movedPiece := &Piece{
		Type:     piece.Type,
		Color:    piece.Color,
		HasMoved: true,
	}
	newBoard.SetPieceAt(to, movedPiece)
	newBoard.SetPieceAt(from, nil)

	// Determine en passant target
	var enPassant *Position
	if piece.Type == Pawn && abs(from.Row-to.Row) == 2 {
		enPassant = &Position{Row: (from.Row + to.Row) / 2, Col: from.Col}
	}

	newTurn := state.CurrentTurn.Opposite()
	newFullMove := state.FullMoveNumber
	if state.CurrentTurn == Black {
		newFullMove++
	}

	return &BoardState{
		Board:           newBoard,
		CurrentTurn:     newTurn,
		CastlingRights:  state.CastlingRights, // TODO: Update castling rights
		EnPassantTarget: enPassant,
		HalfMoveClock:   state.HalfMoveClock + 1,
		FullMoveNumber:  newFullMove,
	}
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// GameStatus represents the current game status
type GameStatus string

const (
	Active    GameStatus = "active"
	Check     GameStatus = "check"
	Checkmate GameStatus = "checkmate"
	Stalemate GameStatus = "stalemate"
	Draw      GameStatus = "draw"
)

// GetGameStatus determines the current game status
func GetGameStatus(state *BoardState) GameStatus {
	// TODO: Implement game status detection
	currentInCheck := IsInCheck(state, state.CurrentTurn)

	// Check if current player has any valid moves
	hasValidMoves := false
	for _, p := range state.Board.GetPiecesByColor(state.CurrentTurn) {
		if len(GetValidMoves(state, p.Position)) > 0 {
			hasValidMoves = true
			break
		}
	}

	if !hasValidMoves {
		if currentInCheck {
			return Checkmate
		}
		return Stalemate
	}

	if currentInCheck {
		return Check
	}
	return Active
}
