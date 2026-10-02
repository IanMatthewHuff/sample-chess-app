import type { Piece, PieceColor, Position } from './pieces';
import { getPieceSymbol } from './pieces';

export type Board = (Piece | null)[][];

export interface BoardState {
    board: Board;
    currentTurn: PieceColor;
    castlingRights: {
        whiteKingside: boolean;
        whiteQueenside: boolean;
        blackKingside: boolean;
        blackQueenside: boolean;
    };
    enPassantTarget: Position | null;
    halfMoveClock: number;
    fullMoveNumber: number;
}

export function renderBoard(board: Board, containerId: string, flipped = false): void {
    const container = document.getElementById(containerId);
    if (!container) return;
    if (container.children.length !== 64) {
        const squares = document.createDocumentFragment();
        for (let i = 0; i < 64; i++) {
            const square = document.createElement('button');
            square.type = 'button';
            squares.appendChild(square);
        }
        container.replaceChildren(squares);
    }

    for (let row = 0; row < 8; row++) {
        for (let col = 0; col < 8; col++) {
            const actualRow = flipped ? 7 - row : row;
            const actualCol = flipped ? 7 - col : col;
            const piece = board[actualRow][actualCol];
            const square = container.children[row * 8 + col] as HTMLButtonElement;
            const color = (row + col) % 2 === 0 ? 'light' : 'dark';
            const className = `square ${color}${piece ? ' piece' : ''}`;
            if (square.className !== className) square.className = className;
            square.dataset.row = String(actualRow);
            square.dataset.col = String(actualCol);
            square.setAttribute('aria-label', `${'abcdefgh'[actualCol]}${8 - actualRow}${piece ? ` ${piece.color} ${piece.type}` : ''}`);
            const symbol = piece ? getPieceSymbol(piece) : '';
            if (square.textContent !== symbol) square.textContent = symbol;
            square.draggable = piece?.color === 'white';
        }
    }
}
