package main

// Board represents an 8x8 chess board
type Board [8][8]*Piece

// CastlingRights tracks castling availability
type CastlingRights struct {
	WhiteKingside  bool `json:"whiteKingside"`
	WhiteQueenside bool `json:"whiteQueenside"`
	BlackKingside  bool `json:"blackKingside"`
	BlackQueenside bool `json:"blackQueenside"`
}

// BoardState represents the complete state of a chess game
type BoardState struct {
	Board           Board           `json:"board"`
	CurrentTurn     PieceColor      `json:"currentTurn"`
	CastlingRights  CastlingRights  `json:"castlingRights"`
	EnPassantTarget *Position       `json:"enPassantTarget"`
	HalfMoveClock   int             `json:"halfMoveClock"`
	FullMoveNumber  int             `json:"fullMoveNumber"`
}

// NewCastlingRights creates default castling rights
func NewCastlingRights() CastlingRights {
	return CastlingRights{
		WhiteKingside:  true,
		WhiteQueenside: true,
		BlackKingside:  true,
		BlackQueenside: true,
	}
}

// CreateInitialBoard creates the initial chess board setup
func CreateInitialBoard() Board {
	// TODO: Implement initial board setup
	var board Board

	// Setup pawns
	for col := 0; col < 8; col++ {
		board[1][col] = NewPiece(Pawn, Black)
		board[6][col] = NewPiece(Pawn, White)
	}

	// Setup back ranks
	backRankPieces := []PieceType{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}

	for col, pieceType := range backRankPieces {
		board[0][col] = NewPiece(pieceType, Black)
		board[7][col] = NewPiece(pieceType, White)
	}

	return board
}

// CreateInitialState creates the initial board state
func CreateInitialState() *BoardState {
	// TODO: Implement initial state creation
	return &BoardState{
		Board:           CreateInitialBoard(),
		CurrentTurn:     White,
		CastlingRights:  NewCastlingRights(),
		EnPassantTarget: nil,
		HalfMoveClock:   0,
		FullMoveNumber:  1,
	}
}

// GetPieceAt returns the piece at a specific position
func (b *Board) GetPieceAt(pos Position) *Piece {
	// TODO: Implement piece retrieval
	if !pos.IsValid() {
		return nil
	}
	return b[pos.Row][pos.Col]
}

// SetPieceAt sets a piece at a specific position
func (b *Board) SetPieceAt(pos Position, piece *Piece) {
	// TODO: Implement piece placement
	if pos.IsValid() {
		b[pos.Row][pos.Col] = piece
	}
}

// Clone creates a deep copy of the board
func (b *Board) Clone() Board {
	// TODO: Implement board cloning
	var newBoard Board
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			if b[row][col] != nil {
				newBoard[row][col] = &Piece{
					Type:     b[row][col].Type,
					Color:    b[row][col].Color,
					HasMoved: b[row][col].HasMoved,
				}
			}
		}
	}
	return newBoard
}

// FindKingPosition finds the king's position for a given color
func (b *Board) FindKingPosition(color PieceColor) *Position {
	// TODO: Implement king position search
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := b[row][col]
			if piece != nil && piece.Type == King && piece.Color == color {
				return &Position{Row: row, Col: col}
			}
		}
	}
	return nil
}

// GetPiecesByColor returns all pieces of a specific color
func (b *Board) GetPiecesByColor(color PieceColor) []struct {
	Piece    *Piece
	Position Position
} {
	// TODO: Implement piece collection by color
	var pieces []struct {
		Piece    *Piece
		Position Position
	}

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := b[row][col]
			if piece != nil && piece.Color == color {
				pieces = append(pieces, struct {
					Piece    *Piece
					Position Position
				}{
					Piece:    piece,
					Position: Position{Row: row, Col: col},
				})
			}
		}
	}
	return pieces
}

// CalculateMaterial calculates total material value for a color
func (b *Board) CalculateMaterial(color PieceColor) int {
	// TODO: Implement material calculation
	total := 0
	for _, p := range b.GetPiecesByColor(color) {
		total += PieceValue(p.Piece.Type)
	}
	return total
}

// PrintBoard prints the board to console (for debugging)
func (b *Board) PrintBoard() {
	// TODO: Implement board printing
	println("  a b c d e f g h")
	for row := 0; row < 8; row++ {
		print(8-row, " ")
		for col := 0; col < 8; col++ {
			piece := b[row][col]
			if piece != nil {
				print(GetPieceSymbol(piece), " ")
			} else {
				print(". ")
			}
		}
		println(8 - row)
	}
	println("  a b c d e f g h")
}
