/**
 * Chess App Entry Point — UI controller. All game state comes from the backend.
 */

import * as api from './api';
import { findKing, renderBoard } from './board';
import { createUIState, describeStatus, GameView, isHumanTurn, Difficulty, GameMode } from './game';
import { capitalize, getPieceSymbol, PieceColor, Position, PromotionPiece, sortForDisplay } from './pieces';

const STORAGE_KEY = 'chess.gameId';
const AI_MOVE_DELAY_MS = 300;

const ui = createUIState();

function $(id: string): HTMLElement {
    const el = document.getElementById(id);
    if (!el) throw new Error(`Missing element #${id}`);
    return el;
}

/**
 * Initialize the chess application
 */
async function init(): Promise<void> {
    setupEventListeners();
    render();

    const health = await api.healthCheck();
    if (!health.ok) {
        showError(health.engine === 'unreachable'
            ? 'The chess engine is not running. Start it and reload.'
            : 'Cannot reach the game server. Start the backend and reload.');
        return;
    }

    const savedId = localStorage.getItem(STORAGE_KEY);
    if (savedId) {
        try {
            applyGame(await api.getGame(savedId));
            syncFormToGame();
            await maybeAIMove();
            return;
        } catch {
            localStorage.removeItem(STORAGE_KEY);
        }
    }
    await startNewGame();
}

/**
 * Setup event listeners for UI controls
 */
function setupEventListeners(): void {
    $('new-game').addEventListener('click', () => void startNewGame());
    $('undo').addEventListener('click', () => void handleUndo());
    $('redo').addEventListener('click', () => void handleRedo());
    $('flip-board').addEventListener('click', handleFlipBoard);
    $('resign').addEventListener('click', () => void handleResign());
    $('board').addEventListener('click', event => void handleBoardClick(event));
    $('opt-mode').addEventListener('change', () => {
        $('ai-options').classList.toggle('hidden', (($('opt-mode') as HTMLSelectElement).value as GameMode) !== 'ai');
    });

    document.querySelectorAll<HTMLButtonElement>('#promotion-dialog button[data-piece]').forEach(btn => {
        btn.addEventListener('click', () => void handlePromotionChoice(btn.dataset.piece as PromotionPiece));
    });
    $('promotion-cancel').addEventListener('click', () => {
        ui.pendingPromotion = null;
        render();
    });
    document.addEventListener('keydown', e => {
        if (e.key === 'Escape') {
            ui.pendingPromotion = null;
            clearSelection();
            render();
        }
    });
}

function readOptions(): api.NewGameOptions {
    return {
        mode: ($('opt-mode') as HTMLSelectElement).value as GameMode,
        humanColor: ($('opt-color') as HTMLSelectElement).value as PieceColor,
        difficulty: ($('opt-difficulty') as HTMLSelectElement).value as Difficulty
    };
}

function syncFormToGame(): void {
    if (!ui.game) return;
    ($('opt-mode') as HTMLSelectElement).value = ui.game.mode;
    ($('opt-color') as HTMLSelectElement).value = ui.game.humanColor;
    ($('opt-difficulty') as HTMLSelectElement).value = ui.game.difficulty;
    $('ai-options').classList.toggle('hidden', ui.game.mode !== 'ai');
}

function applyGame(game: GameView): void {
    const isNew = ui.game?.gameId !== game.gameId;
    ui.game = game;
    ui.error = null;
    clearSelection();
    ui.pendingPromotion = null;
    if (isNew) {
        ui.flipped = game.mode === 'ai' && game.humanColor === 'black';
        localStorage.setItem(STORAGE_KEY, game.gameId);
    }
    render();
}

function clearSelection(): void {
    ui.selected = null;
    ui.targets = [];
}

function showError(message: string): void {
    ui.error = message;
    render();
}

/**
 * Run an API action with the busy flag set and errors surfaced in the UI
 */
async function run<T>(action: () => Promise<T>): Promise<T | undefined> {
    ui.busy = true;
    render();
    try {
        return await action();
    } catch (e) {
        showError((e as Error).message);
        return undefined;
    } finally {
        ui.busy = false;
        render();
    }
}

async function startNewGame(): Promise<void> {
    if (ui.busy || ui.aiThinking) return;
    const game = await run(() => api.createGame(readOptions()));
    if (game) {
        applyGame(game);
        await maybeAIMove();
    }
}

/**
 * If it is the AI's turn, ask the backend to play its move
 */
async function maybeAIMove(): Promise<void> {
    const game = ui.game;
    if (!game || !game.aiToMove || ui.aiThinking) return;

    ui.aiThinking = true;
    render();
    try {
        await new Promise(resolve => setTimeout(resolve, AI_MOVE_DELAY_MS));
        const updated = await api.requestAIMove(game.gameId);
        // Ignore results for a game the user has since replaced.
        if (ui.game?.gameId === game.gameId) applyGame(updated);
    } catch (e) {
        if (ui.game?.gameId === game.gameId) {
            // The game may have changed (e.g. undo) while the AI was thinking; resync.
            try {
                applyGame(await api.getGame(game.gameId));
            } catch {
                // keep the original error below
            }
            showError(`AI move failed: ${(e as Error).message}`);
        }
    } finally {
        ui.aiThinking = false;
        render();
    }
    if (ui.game?.aiToMove && ui.game.gameId === game.gameId && !ui.error) {
        await maybeAIMove();
    }
}

function canInteract(): boolean {
    return !!ui.game && !ui.busy && !ui.aiThinking && !ui.pendingPromotion && isHumanTurn(ui.game);
}

/**
 * Handle board square click
 */
async function handleBoardClick(event: MouseEvent): Promise<void> {
    if (!canInteract() || !ui.game) return;
    const square = (event.target as HTMLElement).closest('.square') as HTMLElement | null;
    if (!square) return;

    const position: Position = { row: Number(square.dataset.row), col: Number(square.dataset.col) };
    const target = ui.targets.find(t => t.row === position.row && t.col === position.col);

    if (ui.selected && target) {
        if (target.isPromotion) {
            ui.pendingPromotion = { from: ui.selected, to: position };
            render();
        } else {
            await submitMove(ui.selected, position);
        }
        return;
    }

    await selectSquare(position);
}

/**
 * Select a square and fetch its legal moves from the backend
 */
async function selectSquare(position: Position): Promise<void> {
    const game = ui.game;
    if (!game) return;
    const piece = game.state.board[position.row][position.col];
    const alreadySelected = ui.selected?.row === position.row && ui.selected?.col === position.col;

    if (!piece || piece.color !== game.state.currentTurn || alreadySelected) {
        clearSelection();
        render();
        return;
    }

    ui.selected = position;
    ui.targets = [];
    render();
    try {
        const targets = await api.getValidMoves(game.gameId, position);
        // Only apply if the selection hasn't changed meanwhile
        if (ui.selected === position) {
            ui.targets = targets;
            render();
        }
    } catch (e) {
        showError((e as Error).message);
    }
}

async function submitMove(from: Position, to: Position, promotion?: PromotionPiece): Promise<void> {
    if (!ui.game) return;
    const gameId = ui.game.gameId;
    const updated = await run(() => api.makeMove(gameId, from, to, promotion));
    if (updated) {
        applyGame(updated);
        await maybeAIMove();
    }
}

async function handlePromotionChoice(piece: PromotionPiece): Promise<void> {
    const pending = ui.pendingPromotion;
    if (!pending) return;
    ui.pendingPromotion = null;
    await submitMove(pending.from, pending.to, piece);
}

async function handleUndo(): Promise<void> {
    if (!ui.game || ui.aiThinking) return;
    const gameId = ui.game.gameId;
    const updated = await run(() => api.undoMove(gameId));
    if (updated) {
        applyGame(updated);
        await maybeAIMove();
    }
}

async function handleRedo(): Promise<void> {
    if (!ui.game || ui.aiThinking) return;
    const gameId = ui.game.gameId;
    const updated = await run(() => api.redoMove(gameId));
    if (updated) {
        applyGame(updated);
        await maybeAIMove();
    }
}

async function handleResign(): Promise<void> {
    if (!ui.game || ui.game.status.gameOver) return;
    const who = ui.game.mode === 'ai' ? 'You' : capitalize(ui.game.state.currentTurn);
    if (!confirm(`${who} will resign this game. Continue?`)) return;
    const gameId = ui.game.gameId;
    const updated = await run(() => api.resignGame(gameId));
    if (updated) applyGame(updated);
}

function handleFlipBoard(): void {
    ui.flipped = !ui.flipped;
    render();
}

/**
 * Render the game UI
 */
function render(): void {
    const game = ui.game;

    if (game) {
        const checkedKing = game.status.inCheck ? findKing(game.state.board, game.state.currentTurn) : null;
        renderBoard(game.state.board, 'board', {
            flipped: ui.flipped,
            selected: ui.selected,
            targets: ui.targets,
            lastMove: game.lastMove,
            checkedKing,
            interactive: canInteract()
        });
    }

    renderTurnIndicator(game);
    renderStatus(game);
    renderCaptured(game);
    renderMoveHistory(game);
    renderControls(game);

    const dialog = $('promotion-dialog');
    dialog.classList.toggle('hidden', !ui.pendingPromotion);
    if (ui.pendingPromotion && game) {
        const color = game.state.currentTurn;
        dialog.querySelectorAll<HTMLButtonElement>('button[data-piece]').forEach(btn => {
            btn.textContent = getPieceSymbol({ type: btn.dataset.piece as PromotionPiece, color });
        });
    }

    const errorEl = $('error-banner');
    errorEl.textContent = ui.error ?? '';
    errorEl.classList.toggle('hidden', !ui.error);
}

function renderTurnIndicator(game: GameView | null): void {
    const el = $('turn-indicator');
    if (!game) {
        el.textContent = 'Connecting…';
        return;
    }
    if (game.status.gameOver) {
        el.textContent = 'Game over';
    } else if (ui.aiThinking) {
        el.textContent = 'AI is thinking…';
    } else if (game.mode === 'ai') {
        el.textContent = game.state.currentTurn === game.humanColor ? 'Your move' : 'AI to move';
    } else {
        el.textContent = `${capitalize(game.state.currentTurn)} to move`;
    }
    el.classList.toggle('thinking', ui.aiThinking);
    el.dataset.turn = game.state.currentTurn;
}

function renderStatus(game: GameView | null): void {
    const statusEl = $('game-status');
    const text = game ? describeStatus(game) : '';
    statusEl.textContent = text;
    statusEl.classList.toggle('game-over', !!game?.status.gameOver);

    const info = $('game-meta');
    if (!game) {
        info.textContent = '';
        return;
    }
    info.textContent = game.mode === 'ai'
        ? `vs AI (${game.difficulty}) · you play ${game.humanColor}`
        : 'Two players, one board';
}

function renderCaptured(game: GameView | null): void {
    for (const color of ['white', 'black'] as PieceColor[]) {
        const el = $(`captured-${color}`);
        const victims = game ? sortForDisplay(game.captured[color]) : [];
        const victimColor: PieceColor = color === 'white' ? 'black' : 'white';
        el.replaceChildren(...victims.map(type => {
            const span = document.createElement('span');
            span.className = `piece piece-${victimColor}`;
            span.textContent = getPieceSymbol({ type, color: victimColor });
            return span;
        }));
        el.classList.toggle('empty', victims.length === 0);
        el.title = `Captured by ${color}`;
    }
}

/**
 * Render move history
 */
function renderMoveHistory(game: GameView | null): void {
    const historyEl = $('move-history');
    historyEl.innerHTML = '';
    if (!game || game.moves.length === 0) {
        historyEl.textContent = 'No moves yet';
        return;
    }

    const list = document.createElement('ol');
    for (let i = 0; i < game.moves.length; i += 2) {
        const li = document.createElement('li');
        for (const move of game.moves.slice(i, i + 2)) {
            const span = document.createElement('span');
            span.className = 'san';
            span.textContent = move.san ?? '?';
            li.appendChild(span);
        }
        list.appendChild(li);
    }
    historyEl.appendChild(list);
    historyEl.scrollTop = historyEl.scrollHeight;
}

function renderControls(game: GameView | null): void {
    const locked = ui.busy || ui.aiThinking;
    ($('undo') as HTMLButtonElement).disabled = !game || locked || !game.canUndo;
    ($('redo') as HTMLButtonElement).disabled = !game || locked || !game.canRedo;
    ($('resign') as HTMLButtonElement).disabled = !game || locked || game.status.gameOver;
    ($('new-game') as HTMLButtonElement).disabled = locked;
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => void init());

export { ui, init };
