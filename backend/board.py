"""
Chess board representation and manipulation
"""

from typing import Optional, List, Tuple
from dataclasses import dataclass, field
from pieces import Piece, PieceColor, PieceType, Position


Board = List[List[Optional[Piece]]]


@dataclass
class CastlingRights:
    """Tracks castling availability"""
    white_kingside: bool = True
    white_queenside: bool = True
    black_kingside: bool = True
    black_queenside: bool = True
    
    def to_dict(self) -> dict:
        return {
            "whiteKingside": self.white_kingside,
            "whiteQueenside": self.white_queenside,
            "blackKingside": self.black_kingside,
            "blackQueenside": self.black_queenside
        }
    
    @classmethod
    def from_dict(cls, data: dict) -> "CastlingRights":
        return cls(
            white_kingside=data.get("whiteKingside", True),
            white_queenside=data.get("whiteQueenside", True),
            black_kingside=data.get("blackKingside", True),
            black_queenside=data.get("blackQueenside", True)
        )


@dataclass
class BoardState:
    """Complete state of a chess game"""
    board: Board
    current_turn: PieceColor = PieceColor.WHITE
    castling_rights: CastlingRights = field(default_factory=CastlingRights)
    en_passant_target: Optional[Position] = None
    half_move_clock: int = 0
    full_move_number: int = 1
    
    def to_dict(self) -> dict:
        """Convert board state to dictionary"""
        board_data = []
        for row in self.board:
            row_data = []
            for piece in row:
                row_data.append(piece.to_dict() if piece else None)
            board_data.append(row_data)
        
        return {
            "board": board_data,
            "currentTurn": self.current_turn.value,
            "castlingRights": self.castling_rights.to_dict(),
            "enPassantTarget": self.en_passant_target.to_dict() if self.en_passant_target else None,
            "halfMoveClock": self.half_move_clock,
            "fullMoveNumber": self.full_move_number
        }


def create_empty_board() -> Board:
    """Create an empty 8x8 board"""
    return [[None for _ in range(8)] for _ in range(8)]


def create_initial_board() -> Board:
    """Create the initial chess board setup"""
    # TODO: Implement initial board setup
    board = create_empty_board()
    
    # Setup pawns
    for col in range(8):
        board[1][col] = Piece(PieceType.PAWN, PieceColor.BLACK)
        board[6][col] = Piece(PieceType.PAWN, PieceColor.WHITE)
    
    # Setup back ranks
    back_rank_pieces = [
        PieceType.ROOK, PieceType.KNIGHT, PieceType.BISHOP, PieceType.QUEEN,
        PieceType.KING, PieceType.BISHOP, PieceType.KNIGHT, PieceType.ROOK
    ]
    
    for col, piece_type in enumerate(back_rank_pieces):
        board[0][col] = Piece(piece_type, PieceColor.BLACK)
        board[7][col] = Piece(piece_type, PieceColor.WHITE)
    
    return board


def create_initial_state() -> BoardState:
    """Create the initial board state"""
    # TODO: Implement initial state creation
    return BoardState(
        board=create_initial_board(),
        current_turn=PieceColor.WHITE,
        castling_rights=CastlingRights(),
        en_passant_target=None,
        half_move_clock=0,
        full_move_number=1
    )


def get_piece_at(board: Board, position: Position) -> Optional[Piece]:
    """Get piece at a specific position"""
    # TODO: Implement piece retrieval
    if not position.is_valid():
        return None
    return board[position.row][position.col]


def set_piece_at(board: Board, position: Position, piece: Optional[Piece]) -> None:
    """Set piece at a specific position"""
    # TODO: Implement piece placement
    if position.is_valid():
        board[position.row][position.col] = piece


def clone_board(board: Board) -> Board:
    """Create a deep copy of the board"""
    # TODO: Implement board cloning
    new_board = create_empty_board()
    for row in range(8):
        for col in range(8):
            piece = board[row][col]
            if piece:
                new_board[row][col] = Piece(
                    type=piece.type,
                    color=piece.color,
                    has_moved=piece.has_moved
                )
    return new_board


def find_king_position(board: Board, color: PieceColor) -> Optional[Position]:
    """Find the king's position for a given color"""
    # TODO: Implement king position search
    for row in range(8):
        for col in range(8):
            piece = board[row][col]
            if piece and piece.type == PieceType.KING and piece.color == color:
                return Position(row, col)
    return None


def get_pieces_by_color(board: Board, color: PieceColor) -> List[Tuple[Piece, Position]]:
    """Get all pieces of a specific color"""
    # TODO: Implement piece collection by color
    pieces = []
    for row in range(8):
        for col in range(8):
            piece = board[row][col]
            if piece and piece.color == color:
                pieces.append((piece, Position(row, col)))
    return pieces


def calculate_material(board: Board, color: PieceColor) -> int:
    """Calculate total material value for a color"""
    # TODO: Implement material calculation
    from pieces import get_piece_value
    total = 0
    for piece, _ in get_pieces_by_color(board, color):
        total += get_piece_value(piece)
    return total


def parse_fen(fen: str) -> BoardState:
    """Parse FEN string to board state"""
    # TODO: Implement FEN parsing
    # FEN format: piece placement, active color, castling, en passant, halfmove, fullmove
    parts = fen.split(" ")
    
    board = create_empty_board()
    rows = parts[0].split("/")
    
    piece_map = {
        'k': (PieceType.KING, PieceColor.BLACK),
        'q': (PieceType.QUEEN, PieceColor.BLACK),
        'r': (PieceType.ROOK, PieceColor.BLACK),
        'b': (PieceType.BISHOP, PieceColor.BLACK),
        'n': (PieceType.KNIGHT, PieceColor.BLACK),
        'p': (PieceType.PAWN, PieceColor.BLACK),
        'K': (PieceType.KING, PieceColor.WHITE),
        'Q': (PieceType.QUEEN, PieceColor.WHITE),
        'R': (PieceType.ROOK, PieceColor.WHITE),
        'B': (PieceType.BISHOP, PieceColor.WHITE),
        'N': (PieceType.KNIGHT, PieceColor.WHITE),
        'P': (PieceType.PAWN, PieceColor.WHITE),
    }
    
    for row_idx, row in enumerate(rows):
        col_idx = 0
        for char in row:
            if char.isdigit():
                col_idx += int(char)
            else:
                piece_type, color = piece_map[char]
                board[row_idx][col_idx] = Piece(piece_type, color)
                col_idx += 1
    
    # Parse active color
    current_turn = PieceColor.WHITE if parts[1] == 'w' else PieceColor.BLACK
    
    # Parse castling rights
    castling_rights = CastlingRights(
        white_kingside='K' in parts[2],
        white_queenside='Q' in parts[2],
        black_kingside='k' in parts[2],
        black_queenside='q' in parts[2]
    )
    
    # Parse en passant
    en_passant = None
    if parts[3] != '-':
        en_passant = Position.from_algebraic(parts[3])
    
    return BoardState(
        board=board,
        current_turn=current_turn,
        castling_rights=castling_rights,
        en_passant_target=en_passant,
        half_move_clock=int(parts[4]),
        full_move_number=int(parts[5])
    )


def to_fen(state: BoardState) -> str:
    """Convert board state to FEN string"""
    # TODO: Implement FEN generation
    raise NotImplementedError("FEN generation not implemented")


def print_board(board: Board) -> None:
    """Print the board to console (for debugging)"""
    from pieces import get_piece_symbol
    
    print("  a b c d e f g h")
    for row in range(8):
        print(f"{8 - row} ", end="")
        for col in range(8):
            piece = board[row][col]
            if piece:
                print(get_piece_symbol(piece) + " ", end="")
            else:
                print(". ", end="")
        print(f"{8 - row}")
    print("  a b c d e f g h")
