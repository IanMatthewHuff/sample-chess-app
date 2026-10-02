import type { Move } from './pieces';

export function renderMoveHistory(moves: Move[]): void {
    const panel = document.getElementById('move-history');
    const list = document.getElementById('move-list');
    if (!panel || !list) return;
    const key = moves.map(move => move.san).join('|');
    if (list.dataset.moves === key) return;
    const followLatest = panel.scrollHeight - panel.scrollTop - panel.clientHeight < 24;
    const rows: HTMLLIElement[] = [];
    for (let i = 0; i < moves.length; i += 2) {
        const row = document.createElement('li');
        const number = document.createElement('span');
        const white = document.createElement('span');
        const black = document.createElement('span');
        number.textContent = `${Math.floor(i / 2) + 1}.`;
        white.textContent = moves[i].san;
        black.textContent = moves[i + 1]?.san ?? '';
        row.append(number, white, black);
        rows.push(row);
    }
    list.replaceChildren(...rows);
    list.dataset.moves = key;
    if (followLatest) panel.scrollTop = panel.scrollHeight;
}
