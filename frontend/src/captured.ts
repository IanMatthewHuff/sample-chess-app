import type { Move, Piece, PieceColor } from './pieces';
import { getPieceSymbol } from './pieces';

function renderCaptures(containerId: string, color: PieceColor, history: Move[]): void {
    const container = document.getElementById(containerId);
    if (!container) return;
    const pieces = history
        .map(move => move.captured)
        .filter((piece): piece is Piece => piece !== null && piece.color === color);
    const key = pieces.map(piece => piece.type).join(',');
    if (container.dataset.captureKey === key) return;
    container.dataset.captureKey = key;
    if (pieces.length === 0) {
        container.textContent = 'None';
        return;
    }
    container.replaceChildren(...pieces.map(piece => {
        const chip = document.createElement('span');
        chip.className = `captured-piece ${piece.color}`;
        chip.textContent = getPieceSymbol(piece);
        chip.setAttribute('aria-label', `${piece.color} ${piece.type}`);
        chip.title = `${piece.color} ${piece.type}`;
        return chip;
    }));
}

export function renderCapturedPieces(history: Move[]): void {
    renderCaptures('captured-by-white', 'black', history);
    renderCaptures('captured-by-black', 'white', history);
}
