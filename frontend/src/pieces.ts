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

export interface LegalMove extends Position {
    isCapture: boolean;
}

export interface Move {
    from: Position;
    to: Position;
    piece: Piece;
    captured: Piece | null;
    promotion: PieceType | null;
    san: string;
}

const symbols: Record<PieceColor, Record<PieceType, string>> = {
    white: { king: '♔', queen: '♕', rook: '♖', bishop: '♗', knight: '♘', pawn: '♙' },
    black: { king: '♚', queen: '♛', rook: '♜', bishop: '♝', knight: '♞', pawn: '♟' }
};

export function getPieceSymbol(piece: Piece): string {
    return symbols[piece.color][piece.type];
}
