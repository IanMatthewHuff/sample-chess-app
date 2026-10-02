"""
Chess piece definitions and utilities
"""

from enum import Enum
from dataclasses import dataclass
from typing import Optional


class PieceColor(Enum):
    WHITE = "white"
    BLACK = "black"
    
    def opposite(self) -> "PieceColor":
        """Get the opposite color"""
        return PieceColor.BLACK if self == PieceColor.WHITE else PieceColor.WHITE


class PieceType(Enum):
    KING = "king"
    QUEEN = "queen"
    ROOK = "rook"
    BISHOP = "bishop"
    KNIGHT = "knight"
    PAWN = "pawn"


@dataclass
class Piece:
    """Represents a chess piece"""
    type: PieceType
    color: PieceColor
    has_moved: bool = False
    
    def to_dict(self) -> dict:
        """Convert piece to dictionary"""
        return {
            "type": self.type.value,
            "color": self.color.value,
            "hasMoved": self.has_moved
        }
    
    @classmethod
    def from_dict(cls, data: dict) -> "Piece":
        """Create piece from dictionary"""
        return cls(
            type=PieceType(data["type"]),
            color=PieceColor(data["color"]),
            has_moved=data.get("hasMoved", False)
        )


@dataclass
class Position:
    """Represents a position on the board"""
    row: int
    col: int
    
    def to_dict(self) -> dict:
        """Convert position to dictionary"""
        return {"row": self.row, "col": self.col}
    
    @classmethod
    def from_dict(cls, data: dict) -> "Position":
        """Create position from dictionary"""
        return cls(row=data["row"], col=data["col"])
    
    def to_algebraic(self) -> str:
        """Convert to algebraic notation (e.g., 'e4')"""
        files = "abcdefgh"
        rank = 8 - self.row
        return f"{files[self.col]}{rank}"
    
    @classmethod
    def from_algebraic(cls, notation: str) -> "Position":
        """Create position from algebraic notation"""
        files = "abcdefgh"
        col = files.index(notation[0])
        row = 8 - int(notation[1])
        return cls(row=row, col=col)
    
    def is_valid(self) -> bool:
        """Check if position is within board bounds"""
        return 0 <= self.row < 8 and 0 <= self.col < 8


@dataclass
class Move:
    """Represents a chess move"""
    from_pos: Position
    to_pos: Position
    piece: Piece
    captured: Optional[Piece] = None
    is_promotion: bool = False
    promoted_to: Optional[PieceType] = None
    is_castling: bool = False
    is_en_passant: bool = False
    
    def to_dict(self) -> dict:
        """Convert move to dictionary"""
        result = {
            "from": self.from_pos.to_dict(),
            "to": self.to_pos.to_dict(),
            "piece": self.piece.to_dict()
        }
        if self.captured:
            result["captured"] = self.captured.to_dict()
        if self.is_promotion:
            result["isPromotion"] = True
            result["promotedTo"] = self.promoted_to.value if self.promoted_to else None
        if self.is_castling:
            result["isCastling"] = True
        if self.is_en_passant:
            result["isEnPassant"] = True
        return result


# Piece values for material calculation
PIECE_VALUES = {
    PieceType.PAWN: 1,
    PieceType.KNIGHT: 3,
    PieceType.BISHOP: 3,
    PieceType.ROOK: 5,
    PieceType.QUEEN: 9,
    PieceType.KING: 0
}


def get_piece_value(piece: Piece) -> int:
    """Get the material value of a piece"""
    # TODO: Implement piece value lookup
    return PIECE_VALUES.get(piece.type, 0)


def get_piece_symbol(piece: Piece) -> str:
    """Get Unicode symbol for a piece"""
    # TODO: Implement symbol lookup
    symbols = {
        PieceColor.WHITE: {
            PieceType.KING: "♔",
            PieceType.QUEEN: "♕",
            PieceType.ROOK: "♖",
            PieceType.BISHOP: "♗",
            PieceType.KNIGHT: "♘",
            PieceType.PAWN: "♙"
        },
        PieceColor.BLACK: {
            PieceType.KING: "♚",
            PieceType.QUEEN: "♛",
            PieceType.ROOK: "♜",
            PieceType.BISHOP: "♝",
            PieceType.KNIGHT: "♞",
            PieceType.PAWN: "♟"
        }
    }
    return symbols[piece.color][piece.type]


def get_piece_notation(piece: Piece) -> str:
    """Get algebraic notation for a piece (K, Q, R, B, N, or empty for pawn)"""
    # TODO: Implement notation lookup
    notation = {
        PieceType.KING: "K",
        PieceType.QUEEN: "Q",
        PieceType.ROOK: "R",
        PieceType.BISHOP: "B",
        PieceType.KNIGHT: "N",
        PieceType.PAWN: ""
    }
    return notation[piece.type]
