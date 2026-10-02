/**
 * Board types (as sent by the backend) and DOM rendering.
 */

import { Piece, Position, PieceColor, getPieceSymbol, samePosition } from './pieces';

export type Board = (Piece | null)[][];

export interface CastlingRights {
    whiteKingside: boolean;
    whiteQueenside: boolean;
    blackKingside: boolean;
    blackQueenside: boolean;
}

export interface BoardState {
    board: Board;
    currentTurn: PieceColor;
    castlingRights: CastlingRights;
    enPassantTarget: Position | null;
    halfMoveClock: number;
    fullMoveNumber: number;
}

export interface MoveTarget extends Position {
    isCapture: boolean;
    isPromotion: boolean;
    isCastling: boolean;
}

export interface RenderOptions {
    flipped: boolean;
    selected: Position | null;
    targets: MoveTarget[];
    lastMove: { from: Position; to: Position } | null;
    checkedKing: Position | null;
    interactive: boolean;
}

const FILES = 'abcdefgh';

/**
 * Convert position to algebraic notation (e.g., {row: 0, col: 0} -> "a8")
 */
export function positionToAlgebraic(position: Position): string {
    return FILES[position.col] + (8 - position.row);
}

/**
 * Find the king of the given color on the board
 */
export function findKing(board: Board, color: PieceColor): Position | null {
    for (let row = 0; row < 8; row++) {
        for (let col = 0; col < 8; col++) {
            const p = board[row][col];
            if (p && p.type === 'king' && p.color === color) return { row, col };
        }
    }
    return null;
}

/**
 * Render the board to the DOM
 */
export function renderBoard(board: Board, containerId: string, opts: RenderOptions): void {
    const container = document.getElementById(containerId);
    if (!container) return;

    container.innerHTML = '';
    container.classList.toggle('locked', !opts.interactive);

    for (let displayRow = 0; displayRow < 8; displayRow++) {
        for (let displayCol = 0; displayCol < 8; displayCol++) {
            const row = opts.flipped ? 7 - displayRow : displayRow;
            const col = opts.flipped ? 7 - displayCol : displayCol;
            const pos = { row, col };

            const square = document.createElement('div');
            square.className = `square ${(row + col) % 2 === 0 ? 'light' : 'dark'}`;
            square.dataset.row = String(row);
            square.dataset.col = String(col);
            square.setAttribute('aria-label', positionToAlgebraic(pos));

            const piece = board[row][col];
            if (piece) {
                const span = document.createElement('span');
                span.className = `piece piece-${piece.color}`;
                span.textContent = getPieceSymbol(piece);
                square.appendChild(span);
            }

            if (opts.lastMove && (samePosition(opts.lastMove.from, pos) || samePosition(opts.lastMove.to, pos))) {
                square.classList.add('last-move');
            }
            if (samePosition(opts.selected, pos)) square.classList.add('selected');
            if (samePosition(opts.checkedKing, pos)) square.classList.add('in-check');

            const target = opts.targets.find(t => samePosition(t, pos));
            if (target) square.classList.add(target.isCapture ? 'capture-move' : 'valid-move');

            if (displayCol === 0) {
                const rank = document.createElement('span');
                rank.className = 'coord rank';
                rank.textContent = String(8 - row);
                square.appendChild(rank);
            }
            if (displayRow === 7) {
                const file = document.createElement('span');
                file.className = 'coord file';
                file.textContent = FILES[col];
                square.appendChild(file);
            }

            container.appendChild(square);
        }
    }
}
