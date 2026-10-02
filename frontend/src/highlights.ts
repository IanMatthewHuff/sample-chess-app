import type { LegalMove, Move, Piece, PieceColor, Position } from './pieces';

const same = (a: Position, b: Position) => a.row === b.row && a.col === b.col;

export function highlightSquare(
    square: HTMLElement,
    pos: Position,
    piece: Piece | null,
    selected: Position | null,
    legal: LegalMove[],
    lastMove?: Move,
    checkedKing: PieceColor | null = null
): void {
    square.classList.toggle('selected', selected !== null && same(pos, selected));
    square.classList.toggle('last-move-from', lastMove !== undefined && same(pos, lastMove.from));
    square.classList.toggle('last-move-to', lastMove !== undefined && same(pos, lastMove.to));
    const move = legal.find(candidate => same(candidate, pos));
    square.classList.toggle('capture-move', move?.isCapture === true);
    square.classList.toggle('valid-move', move !== undefined && !move.isCapture);
    const inCheck = piece?.type === 'king' && piece.color === checkedKing;
    square.classList.toggle('king-in-check', inCheck);
    const label = `${'abcdefgh'[pos.col]}${8 - pos.row}${piece ? ` ${piece.color} ${piece.type}` : ''}${inCheck ? ', in check' : ''}${move ? move.isCapture ? ', capture available' : ', legal move' : ''}`;
    if (square.getAttribute('aria-label') !== label) square.setAttribute('aria-label', label);
}
