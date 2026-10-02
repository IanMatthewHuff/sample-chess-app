/**
 * Chess piece types and related utilities
 */

export type PieceColor = 'white' | 'black';

export type PieceType = 'king' | 'queen' | 'rook' | 'bishop' | 'knight' | 'pawn';

export interface Piece {
    type: PieceType;
    color: PieceColor;
    hasMoved: boolean;
}

export interface Position {
    row: number;
    col: number;
}

export interface Move {
    from: Position;
    to: Position;
    piece: Piece;
    captured?: Piece;
    isPromotion?: boolean;
    promotedTo?: PieceType;
    isCastling?: boolean;
    isEnPassant?: boolean;
}

// Unicode chess piece symbols
const PIECE_SYMBOLS: Record<PieceColor, Record<PieceType, string>> = {
    white: {
        king: '♔',
        queen: '♕',
        rook: '♖',
        bishop: '♗',
        knight: '♘',
        pawn: '♙'
    },
    black: {
        king: '♚',
        queen: '♛',
        rook: '♜',
        bishop: '♝',
        knight: '♞',
        pawn: '♟'
    }
};

/**
 * Get the Unicode symbol for a piece
 */
export function getPieceSymbol(piece: Piece): string {
    // TODO: Implement piece symbol lookup
    return PIECE_SYMBOLS[piece.color][piece.type];
}

/**
 * Create a new piece
 */
export function createPiece(type: PieceType, color: PieceColor): Piece {
    // TODO: Implement piece creation
    return {
        type,
        color,
        hasMoved: false
    };
}

/**
 * Get the point value of a piece (for material calculation)
 */
export function getPieceValue(piece: Piece): number {
    // TODO: Implement piece valuation
    // Standard values: pawn=1, knight=3, bishop=3, rook=5, queen=9, king=0
    const values: Record<PieceType, number> = {
        pawn: 1,
        knight: 3,
        bishop: 3,
        rook: 5,
        queen: 9,
        king: 0
    };
    return values[piece.type];
}

/**
 * Check if two pieces are the same color
 */
export function isSameColor(piece1: Piece, piece2: Piece): boolean {
    // TODO: Implement color comparison
    return piece1.color === piece2.color;
}

/**
 * Check if a piece is a sliding piece (queen, rook, bishop)
 */
export function isSlidingPiece(piece: Piece): boolean {
    // TODO: Implement sliding piece check
    return ['queen', 'rook', 'bishop'].includes(piece.type);
}

/**
 * Get algebraic notation for a piece (K, Q, R, B, N, or empty for pawn)
 */
export function getPieceNotation(piece: Piece): string {
    // TODO: Implement piece notation
    const notation: Record<PieceType, string> = {
        king: 'K',
        queen: 'Q',
        rook: 'R',
        bishop: 'B',
        knight: 'N',
        pawn: ''
    };
    return notation[piece.type];
}
