/**
 * Chess piece types and display helpers. No rules live here — the Go engine
 * is the single source of truth for legality.
 */

export type PieceColor = 'white' | 'black';

export type PieceType = 'king' | 'queen' | 'rook' | 'bishop' | 'knight' | 'pawn';

export type PromotionPiece = 'queen' | 'rook' | 'bishop' | 'knight';

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
    san?: string;
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
export function getPieceSymbol(piece: Pick<Piece, 'type' | 'color'>): string {
    return PIECE_SYMBOLS[piece.color][piece.type];
}

// Used only to sort captured pieces for display.
const DISPLAY_ORDER: Record<PieceType, number> = {
    queen: 0, rook: 1, bishop: 2, knight: 3, pawn: 4, king: 5
};

export function sortForDisplay(types: PieceType[]): PieceType[] {
    return [...types].sort((a, b) => DISPLAY_ORDER[a] - DISPLAY_ORDER[b]);
}

export function samePosition(a: Position | null | undefined, b: Position | null | undefined): boolean {
    return !!a && !!b && a.row === b.row && a.col === b.col;
}

export function capitalize(s: string): string {
    return s.charAt(0).toUpperCase() + s.slice(1);
}
