package main

// PieceColor represents the color of a chess piece
type PieceColor string

const (
	White PieceColor = "white"
	Black PieceColor = "black"
)

// Opposite returns the opposite color
func (c PieceColor) Opposite() PieceColor {
	if c == White {
		return Black
	}
	return White
}

// PieceType represents the type of a chess piece
type PieceType string

const (
	King   PieceType = "king"
	Queen  PieceType = "queen"
	Rook   PieceType = "rook"
	Bishop PieceType = "bishop"
	Knight PieceType = "knight"
	Pawn   PieceType = "pawn"
)

// Piece represents a chess piece
type Piece struct {
	Type     PieceType  `json:"type"`
	Color    PieceColor `json:"color"`
	HasMoved bool       `json:"hasMoved"`
}

// Position represents a position on the board
type Position struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// IsValid checks if the position is within board bounds
func (p Position) IsValid() bool {
	return p.Row >= 0 && p.Row < 8 && p.Col >= 0 && p.Col < 8
}

// ToAlgebraic converts position to algebraic notation
func (p Position) ToAlgebraic() string {
	files := "abcdefgh"
	rank := 8 - p.Row
	return string(files[p.Col]) + string(rune('0'+rank))
}

// Move represents a chess move
type Move struct {
	From        Position  `json:"from"`
	To          Position  `json:"to"`
	Piece       *Piece    `json:"piece"`
	Captured    *Piece    `json:"captured,omitempty"`
	IsPromotion bool      `json:"isPromotion,omitempty"`
	PromotedTo  PieceType `json:"promotedTo,omitempty"`
	IsCastling  bool      `json:"isCastling,omitempty"`
	IsEnPassant bool      `json:"isEnPassant,omitempty"`
	SAN         string    `json:"san,omitempty"`
}

// Equals reports whether two positions refer to the same square
func (p Position) Equals(o Position) bool {
	return p.Row == o.Row && p.Col == o.Col
}

// PositionFromAlgebraic parses a square such as "e4"
func PositionFromAlgebraic(s string) (Position, bool) {
	if len(s) != 2 || s[0] < 'a' || s[0] > 'h' || s[1] < '1' || s[1] > '8' {
		return Position{}, false
	}
	return Position{Row: 8 - int(s[1]-'0'), Col: int(s[0] - 'a')}, true
}

// IsValidPromotion reports whether a pawn may promote to the given type
func IsValidPromotion(t PieceType) bool {
	return t == Queen || t == Rook || t == Bishop || t == Knight
}

// PieceLetter returns the SAN letter for a piece type (empty for pawns)
func PieceLetter(t PieceType) string {
	switch t {
	case King:
		return "K"
	case Queen:
		return "Q"
	case Rook:
		return "R"
	case Bishop:
		return "B"
	case Knight:
		return "N"
	}
	return ""
}

// PieceValue returns the material value of a piece
func PieceValue(pieceType PieceType) int {
	// TODO: Implement piece valuation
	values := map[PieceType]int{
		Pawn:   1,
		Knight: 3,
		Bishop: 3,
		Rook:   5,
		Queen:  9,
		King:   0,
	}
	return values[pieceType]
}

// NewPiece creates a new piece
func NewPiece(pieceType PieceType, color PieceColor) *Piece {
	return &Piece{
		Type:     pieceType,
		Color:    color,
		HasMoved: false,
	}
}

// IsSlidingPiece checks if a piece is a sliding piece (queen, rook, bishop)
func IsSlidingPiece(piece *Piece) bool {
	// TODO: Implement sliding piece check
	return piece.Type == Queen || piece.Type == Rook || piece.Type == Bishop
}

// GetPieceSymbol returns the Unicode symbol for a piece
func GetPieceSymbol(piece *Piece) string {
	// TODO: Implement piece symbol lookup
	symbols := map[PieceColor]map[PieceType]string{
		White: {
			King:   "♔",
			Queen:  "♕",
			Rook:   "♖",
			Bishop: "♗",
			Knight: "♘",
			Pawn:   "♙",
		},
		Black: {
			King:   "♚",
			Queen:  "♛",
			Rook:   "♜",
			Bishop: "♝",
			Knight: "♞",
			Pawn:   "♟",
		},
	}
	return symbols[piece.Color][piece.Type]
}
