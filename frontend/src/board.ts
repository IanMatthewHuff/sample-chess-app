/**
 * Chess board representation and manipulation
 */

import { Piece, Position, PieceColor, PieceType, createPiece, getPieceSymbol } from './pieces';

export type Board = (Piece | null)[][];

export interface BoardState {
    board: Board;
    currentTurn: PieceColor;
    castlingRights: CastlingRights;
    enPassantTarget: Position | null;
    halfMoveClock: number;
    fullMoveNumber: number;
}

export interface CastlingRights {
    whiteKingside: boolean;
    whiteQueenside: boolean;
    blackKingside: boolean;
    blackQueenside: boolean;
}

/**
 * Create the initial chess board setup
 */
export function createInitialBoard(): Board {
    // TODO: Implement initial board setup
    const board: Board = Array(8).fill(null).map(() => Array(8).fill(null));
    
    // Setup pawns
    for (let col = 0; col < 8; col++) {
        board[1][col] = createPiece('pawn', 'black');
        board[6][col] = createPiece('pawn', 'white');
    }
    
    // Setup back ranks
    const backRankPieces: PieceType[] = ['rook', 'knight', 'bishop', 'queen', 'king', 'bishop', 'knight', 'rook'];
    
    for (let col = 0; col < 8; col++) {
        board[0][col] = createPiece(backRankPieces[col], 'black');
        board[7][col] = createPiece(backRankPieces[col], 'white');
    }
    
    return board;
}

/**
 * Create initial board state
 */
export function createInitialBoardState(): BoardState {
    // TODO: Implement initial state creation
    return {
        board: createInitialBoard(),
        currentTurn: 'white',
        castlingRights: {
            whiteKingside: true,
            whiteQueenside: true,
            blackKingside: true,
            blackQueenside: true
        },
        enPassantTarget: null,
        halfMoveClock: 0,
        fullMoveNumber: 1
    };
}

/**
 * Get piece at a specific position
 */
export function getPieceAt(board: Board, position: Position): Piece | null {
    // TODO: Implement piece retrieval
    if (!isValidPosition(position)) {
        return null;
    }
    return board[position.row][position.col];
}

/**
 * Set piece at a specific position
 */
export function setPieceAt(board: Board, position: Position, piece: Piece | null): void {
    // TODO: Implement piece placement
    if (isValidPosition(position)) {
        board[position.row][position.col] = piece;
    }
}

/**
 * Check if a position is valid (within board bounds)
 */
export function isValidPosition(position: Position): boolean {
    // TODO: Implement position validation
    return position.row >= 0 && position.row < 8 && position.col >= 0 && position.col < 8;
}

/**
 * Convert position to algebraic notation (e.g., {row: 0, col: 0} -> "a8")
 */
export function positionToAlgebraic(position: Position): string {
    // TODO: Implement algebraic notation conversion
    const files = 'abcdefgh';
    const rank = 8 - position.row;
    return files[position.col] + rank;
}

/**
 * Convert algebraic notation to position (e.g., "e4" -> {row: 4, col: 4})
 */
export function algebraicToPosition(algebraic: string): Position {
    // TODO: Implement algebraic to position conversion
    const files = 'abcdefgh';
    const col = files.indexOf(algebraic[0]);
    const row = 8 - parseInt(algebraic[1]);
    return { row, col };
}

/**
 * Clone a board (deep copy)
 */
export function cloneBoard(board: Board): Board {
    // TODO: Implement board cloning
    return board.map(row => row.map(piece => piece ? { ...piece } : null));
}

/**
 * Find the king's position for a given color
 */
export function findKingPosition(board: Board, color: PieceColor): Position | null {
    // TODO: Implement king position search
    for (let row = 0; row < 8; row++) {
        for (let col = 0; col < 8; col++) {
            const piece = board[row][col];
            if (piece && piece.type === 'king' && piece.color === color) {
                return { row, col };
            }
        }
    }
    return null;
}

/**
 * Get all pieces of a specific color
 */
export function getPiecesByColor(board: Board, color: PieceColor): { piece: Piece; position: Position }[] {
    // TODO: Implement piece collection by color
    const pieces: { piece: Piece; position: Position }[] = [];
    for (let row = 0; row < 8; row++) {
        for (let col = 0; col < 8; col++) {
            const piece = board[row][col];
            if (piece && piece.color === color) {
                pieces.push({ piece, position: { row, col } });
            }
        }
    }
    return pieces;
}

/**
 * Render the board to the DOM
 */
export function renderBoard(board: Board, containerId: string, flipped: boolean = false): void {
    // TODO: Implement board rendering
    const container = document.getElementById(containerId);
    if (!container) return;
    
    container.innerHTML = '';
    
    for (let row = 0; row < 8; row++) {
        for (let col = 0; col < 8; col++) {
            const actualRow = flipped ? 7 - row : row;
            const actualCol = flipped ? 7 - col : col;
            
            const square = document.createElement('div');
            square.className = `square ${(row + col) % 2 === 0 ? 'light' : 'dark'}`;
            square.dataset.row = actualRow.toString();
            square.dataset.col = actualCol.toString();
            
            const piece = board[actualRow][actualCol];
            if (piece) {
                square.textContent = getPieceSymbol(piece);
                square.classList.add('piece');
            }
            
            container.appendChild(square);
        }
    }
}

/**
 * Parse FEN string to board state
 */
export function parseFEN(fen: string): BoardState {
    // TODO: Implement FEN parsing
    // FEN format: piece placement, active color, castling, en passant, halfmove, fullmove
    throw new Error('Not implemented');
}

/**
 * Convert board state to FEN string
 */
export function toFEN(state: BoardState): string {
    // TODO: Implement FEN generation
    throw new Error('Not implemented');
}
