"""
Chess game logic and move validation
"""

from typing import List, Optional, Tuple
from dataclasses import dataclass
from enum import Enum

from pieces import Piece, PieceColor, PieceType, Position, Move
from board import (
    Board, BoardState, create_initial_state, get_piece_at, set_piece_at,
    clone_board, find_king_position, get_pieces_by_color
)


class GameStatus(Enum):
    ACTIVE = "active"
    CHECK = "check"
    CHECKMATE = "checkmate"
    STALEMATE = "stalemate"
    DRAW = "draw"


@dataclass
class GameState:
    """Complete game state including history"""
    board_state: BoardState
    move_history: List[Move]
    status: GameStatus = GameStatus.ACTIVE
    
    def to_dict(self) -> dict:
        return {
            "boardState": self.board_state.to_dict(),
            "moveHistory": [m.to_dict() for m in self.move_history],
            "status": self.status.value
        }


def create_new_game() -> GameState:
    """Create a new game"""
    # TODO: Implement new game creation
    return GameState(
        board_state=create_initial_state(),
        move_history=[],
        status=GameStatus.ACTIVE
    )


def get_valid_moves(state: BoardState, position: Position) -> List[Position]:
    """Get all valid moves for a piece at the given position"""
    # TODO: Implement valid move calculation
    piece = get_piece_at(state.board, position)
    if not piece or piece.color != state.current_turn:
        return []
    
    moves = []
    
    if piece.type == PieceType.PAWN:
        moves = get_pawn_moves(state, position, piece)
    elif piece.type == PieceType.KNIGHT:
        moves = get_knight_moves(state, position, piece)
    elif piece.type == PieceType.BISHOP:
        moves = get_bishop_moves(state, position, piece)
    elif piece.type == PieceType.ROOK:
        moves = get_rook_moves(state, position, piece)
    elif piece.type == PieceType.QUEEN:
        moves = get_queen_moves(state, position, piece)
    elif piece.type == PieceType.KING:
        moves = get_king_moves(state, position, piece)
    
    # Filter moves that would leave king in check
    return [m for m in moves if not would_be_in_check(state, position, m)]


def get_pawn_moves(state: BoardState, position: Position, piece: Piece) -> List[Position]:
    """Get pawn moves"""
    # TODO: Implement pawn move logic
    moves = []
    direction = -1 if piece.color == PieceColor.WHITE else 1
    start_row = 6 if piece.color == PieceColor.WHITE else 1
    
    # Single move forward
    one_forward = Position(position.row + direction, position.col)
    if one_forward.is_valid() and not get_piece_at(state.board, one_forward):
        moves.append(one_forward)
        
        # Double move from starting position
        if position.row == start_row:
            two_forward = Position(position.row + 2 * direction, position.col)
            if not get_piece_at(state.board, two_forward):
                moves.append(two_forward)
    
    # Captures
    for col_offset in [-1, 1]:
        capture_pos = Position(position.row + direction, position.col + col_offset)
        if capture_pos.is_valid():
            target = get_piece_at(state.board, capture_pos)
            if target and target.color != piece.color:
                moves.append(capture_pos)
            
            # En passant
            if state.en_passant_target:
                if (capture_pos.row == state.en_passant_target.row and 
                    capture_pos.col == state.en_passant_target.col):
                    moves.append(capture_pos)
    
    return moves


def get_knight_moves(state: BoardState, position: Position, piece: Piece) -> List[Position]:
    """Get knight moves"""
    # TODO: Implement knight move logic
    moves = []
    offsets = [
        (-2, -1), (-2, 1), (-1, -2), (-1, 2),
        (1, -2), (1, 2), (2, -1), (2, 1)
    ]
    
    for row_off, col_off in offsets:
        new_pos = Position(position.row + row_off, position.col + col_off)
        if new_pos.is_valid():
            target = get_piece_at(state.board, new_pos)
            if not target or target.color != piece.color:
                moves.append(new_pos)
    
    return moves


def get_bishop_moves(state: BoardState, position: Position, piece: Piece) -> List[Position]:
    """Get bishop moves"""
    # TODO: Implement bishop move logic
    return get_sliding_moves(state, position, piece, [(-1, -1), (-1, 1), (1, -1), (1, 1)])


def get_rook_moves(state: BoardState, position: Position, piece: Piece) -> List[Position]:
    """Get rook moves"""
    # TODO: Implement rook move logic
    return get_sliding_moves(state, position, piece, [(-1, 0), (1, 0), (0, -1), (0, 1)])


def get_queen_moves(state: BoardState, position: Position, piece: Piece) -> List[Position]:
    """Get queen moves"""
    # TODO: Implement queen move logic
    return get_sliding_moves(state, position, piece, [
        (-1, -1), (-1, 0), (-1, 1),
        (0, -1), (0, 1),
        (1, -1), (1, 0), (1, 1)
    ])


def get_king_moves(state: BoardState, position: Position, piece: Piece) -> List[Position]:
    """Get king moves (including castling)"""
    # TODO: Implement king move logic
    moves = []
    offsets = [
        (-1, -1), (-1, 0), (-1, 1),
        (0, -1), (0, 1),
        (1, -1), (1, 0), (1, 1)
    ]
    
    for row_off, col_off in offsets:
        new_pos = Position(position.row + row_off, position.col + col_off)
        if new_pos.is_valid():
            target = get_piece_at(state.board, new_pos)
            if not target or target.color != piece.color:
                moves.append(new_pos)
    
    # TODO: Add castling logic
    
    return moves


def get_sliding_moves(state: BoardState, position: Position, piece: Piece, 
                      directions: List[Tuple[int, int]]) -> List[Position]:
    """Get moves for sliding pieces (bishop, rook, queen)"""
    # TODO: Implement sliding move logic
    moves = []
    
    for row_dir, col_dir in directions:
        new_pos = Position(position.row + row_dir, position.col + col_dir)
        
        while new_pos.is_valid():
            target = get_piece_at(state.board, new_pos)
            
            if not target:
                moves.append(Position(new_pos.row, new_pos.col))
            else:
                if target.color != piece.color:
                    moves.append(Position(new_pos.row, new_pos.col))
                break
            
            new_pos = Position(new_pos.row + row_dir, new_pos.col + col_dir)
    
    return moves


def would_be_in_check(state: BoardState, from_pos: Position, to_pos: Position) -> bool:
    """Check if making a move would leave the king in check"""
    # TODO: Implement check detection after move
    test_board = clone_board(state.board)
    piece = get_piece_at(test_board, from_pos)
    
    if not piece:
        return True
    
    set_piece_at(test_board, to_pos, piece)
    set_piece_at(test_board, from_pos, None)
    
    test_state = BoardState(
        board=test_board,
        current_turn=state.current_turn,
        castling_rights=state.castling_rights,
        en_passant_target=state.en_passant_target
    )
    
    return is_in_check(test_state, piece.color)


def is_in_check(state: BoardState, color: PieceColor) -> bool:
    """Check if a player is in check"""
    # TODO: Implement check detection
    king_pos = find_king_position(state.board, color)
    if not king_pos:
        return False
    
    opponent_color = color.opposite()
    
    # Check if any opponent piece can attack the king
    for piece, pos in get_pieces_by_color(state.board, opponent_color):
        attack_squares = get_attack_squares(state, pos, piece)
        if any(sq.row == king_pos.row and sq.col == king_pos.col for sq in attack_squares):
            return True
    
    return False


def get_attack_squares(state: BoardState, position: Position, piece: Piece) -> List[Position]:
    """Get squares that a piece can attack (ignoring check)"""
    # TODO: Implement attack square calculation
    if piece.type == PieceType.PAWN:
        # Pawns attack diagonally
        direction = -1 if piece.color == PieceColor.WHITE else 1
        attacks = []
        for col_offset in [-1, 1]:
            pos = Position(position.row + direction, position.col + col_offset)
            if pos.is_valid():
                attacks.append(pos)
        return attacks
    elif piece.type == PieceType.KNIGHT:
        return get_knight_moves(state, position, piece)
    elif piece.type == PieceType.BISHOP:
        return get_bishop_moves(state, position, piece)
    elif piece.type == PieceType.ROOK:
        return get_rook_moves(state, position, piece)
    elif piece.type == PieceType.QUEEN:
        return get_queen_moves(state, position, piece)
    elif piece.type == PieceType.KING:
        return get_king_moves(state, position, piece)
    return []


def make_move(game_state: GameState, from_pos: Position, to_pos: Position) -> GameState:
    """Make a move and return the new game state"""
    # TODO: Implement move execution
    piece = get_piece_at(game_state.board_state.board, from_pos)
    if not piece:
        return game_state
    
    new_board = clone_board(game_state.board_state.board)
    captured = get_piece_at(new_board, to_pos)
    
    # Update piece
    moved_piece = Piece(piece.type, piece.color, has_moved=True)
    set_piece_at(new_board, to_pos, moved_piece)
    set_piece_at(new_board, from_pos, None)
    
    # Create move record
    move = Move(
        from_pos=from_pos,
        to_pos=to_pos,
        piece=piece,
        captured=captured
    )
    
    # Determine en passant target
    en_passant = None
    if piece.type == PieceType.PAWN and abs(from_pos.row - to_pos.row) == 2:
        en_passant = Position((from_pos.row + to_pos.row) // 2, from_pos.col)
    
    # Update board state
    new_board_state = BoardState(
        board=new_board,
        current_turn=game_state.board_state.current_turn.opposite(),
        castling_rights=game_state.board_state.castling_rights,  # TODO: Update castling rights
        en_passant_target=en_passant,
        half_move_clock=game_state.board_state.half_move_clock + 1,
        full_move_number=game_state.board_state.full_move_number + (1 if game_state.board_state.current_turn == PieceColor.BLACK else 0)
    )
    
    # Determine new game status
    status = get_game_status(new_board_state)
    
    return GameState(
        board_state=new_board_state,
        move_history=game_state.move_history + [move],
        status=status
    )


def get_game_status(state: BoardState) -> GameStatus:
    """Determine the current game status"""
    # TODO: Implement game status detection
    current_in_check = is_in_check(state, state.current_turn)
    
    # Check if current player has any valid moves
    has_valid_moves = False
    for piece, pos in get_pieces_by_color(state.board, state.current_turn):
        if get_valid_moves(state, pos):
            has_valid_moves = True
            break
    
    if not has_valid_moves:
        return GameStatus.CHECKMATE if current_in_check else GameStatus.STALEMATE
    
    return GameStatus.CHECK if current_in_check else GameStatus.ACTIVE


def is_valid_move(state: BoardState, from_pos: Position, to_pos: Position) -> bool:
    """Check if a move is valid"""
    # TODO: Implement move validation
    valid_moves = get_valid_moves(state, from_pos)
    return any(m.row == to_pos.row and m.col == to_pos.col for m in valid_moves)


def undo_move(game_state: GameState) -> GameState:
    """Undo the last move"""
    # TODO: Implement undo
    if not game_state.move_history:
        return game_state
    
    last_move = game_state.move_history[-1]
    new_board = clone_board(game_state.board_state.board)
    
    # Restore piece to original position
    original_piece = Piece(last_move.piece.type, last_move.piece.color, has_moved=False)
    set_piece_at(new_board, last_move.from_pos, original_piece)
    set_piece_at(new_board, last_move.to_pos, last_move.captured)
    
    new_board_state = BoardState(
        board=new_board,
        current_turn=game_state.board_state.current_turn.opposite(),
        castling_rights=game_state.board_state.castling_rights,
        en_passant_target=None  # TODO: Restore proper en passant state
    )
    
    return GameState(
        board_state=new_board_state,
        move_history=game_state.move_history[:-1],
        status=get_game_status(new_board_state)
    )


def move_to_algebraic(move: Move, state: BoardState) -> str:
    """Convert a move to algebraic notation"""
    # TODO: Implement algebraic notation generation
    from pieces import get_piece_notation
    
    piece_notation = get_piece_notation(move.piece)
    dest = move.to_pos.to_algebraic()
    
    capture_str = "x" if move.captured else ""
    
    return f"{piece_notation}{capture_str}{dest}"
