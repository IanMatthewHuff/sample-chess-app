import { renderBoard, type Board } from './board';
import type { GameState } from './game';
import type { LegalMove, PieceType, Position } from './pieces';
import { renderCapturedPieces } from './captured';
import { renderMoveHistory } from './history';
import { highlightSquare } from './highlights';
import { createGame, getGameState, makeMove, makeAiMove, getValidMoves, undoMove, redoMove } from './api';

let game: GameState | null = null;
let selected: Position | null = null;
let legal: LegalMove[] = [];
let flipped = false;
let busy = false;
let message = '';
let selectionVersion = 0;
let lastRenderedBoard: Board | null = null;
let lastRenderedFlipped = false;

const same = (a: Position, b: Position) => a.row === b.row && a.col === b.col;
const terminal = (status: string) => ['checkmate', 'stalemate', 'draw'].includes(status);

function render(): void {
    if (!game) return;
    if (lastRenderedBoard !== game.state.board || lastRenderedFlipped !== flipped) {
        renderBoard(game.state.board, 'board', flipped);
        lastRenderedBoard = game.state.board;
        lastRenderedFlipped = flipped;
    }
    const board = document.getElementById('board')!;
    if (board.getAttribute('aria-busy') !== String(busy)) {
        board.setAttribute('aria-busy', String(busy));
    }
    const lastMove = game.history[game.history.length - 1];
    const checkedKing = game.status === 'check' || game.status === 'checkmate'
        ? game.state.currentTurn : null;
    for (const square of document.querySelectorAll<HTMLElement>('.square')) {
        const pos = { row: Number(square.dataset.row), col: Number(square.dataset.col) };
        highlightSquare(square, pos, game.state.board[pos.row][pos.col], selected, legal, lastMove, checkedKing);
        const disabled = String(terminal(game.status));
        if (square.getAttribute('aria-disabled') !== disabled) {
            square.setAttribute('aria-disabled', disabled);
        }
    }
    const turn = document.getElementById('turn-indicator')!;
    const turnText = game.state.currentTurn === 'white' ? 'Your turn (White)' : 'Computer thinking (Black)';
    if (turn.textContent !== turnText) turn.textContent = turnText;
    const status = document.getElementById('game-status')!;
    const result = game.status === 'checkmate'
        ? `${game.state.currentTurn === 'white' ? 'Computer' : 'You'} win by checkmate!`
        : game.status === 'stalemate' ? 'Stalemate — draw'
        : game.status === 'draw' ? 'Draw'
        : game.status === 'check' ? 'Check!' : '';
    const statusText = message || result;
    if (status.textContent !== statusText) status.textContent = statusText;
    const undoButton = document.getElementById('undo') as HTMLButtonElement;
    const redoButton = document.getElementById('redo') as HTMLButtonElement;
    const retryButton = document.getElementById('retry-ai') as HTMLButtonElement;
    if (undoButton.disabled !== (busy || !game.canUndo)) undoButton.disabled = busy || !game.canUndo;
    if (redoButton.disabled !== (busy || !game.canRedo)) redoButton.disabled = busy || !game.canRedo;
    const hideRetry = busy || game.state.currentTurn !== 'black' || terminal(game.status);
    if (retryButton.hidden !== hideRetry) retryButton.hidden = hideRetry;
    renderMoveHistory(game.history);
    renderCapturedPieces(game.history);
}

function choosePromotion(): Promise<PieceType | null> {
    const dialog = document.getElementById('promotion-dialog') as HTMLDialogElement;
    return new Promise(resolve => {
        dialog.addEventListener('close', () => {
            const choice = dialog.returnValue;
            resolve(['queen', 'rook', 'bishop', 'knight'].includes(choice)
                ? choice as PieceType : null);
        }, { once: true });
        dialog.returnValue = '';
        dialog.showModal();
    });
}

function showError(error: unknown): void {
    message = error instanceof Error ? error.message : String(error);
    render();
}

async function computerTurn(): Promise<void> {
    if (!game || game.state.currentTurn !== 'black' || terminal(game.status)) return;
    busy = true;
    message = '';
    render();
    try {
        game = await makeAiMove(game.gameId);
    } catch (error) {
        showError(error);
    } finally {
        busy = false;
        render();
    }
}

async function selectSquare(pos: Position): Promise<void> {
    if (!game || game.state.currentTurn !== 'white' || terminal(game.status)) return;
    const piece = game.state.board[pos.row][pos.col];
    if (!piece || piece.color !== 'white') {
        selected = null;
        legal = [];
        render();
        return;
    }
    const version = ++selectionVersion;
    selected = pos;
    legal = [];
    render();
    try {
        const moves = await getValidMoves(game.gameId, pos);
        if (version === selectionVersion) {
            legal = moves;
            message = '';
        }
    } catch (error) {
        if (version === selectionVersion) showError(error);
    }
    render();
}

async function boardAction(pos: Position): Promise<void> {
    if (!game || busy || game.state.currentTurn !== 'white' || terminal(game.status)) return;
    if (selected && legal.some(move => same(move, pos))) {
        const from = selected;
        const piece = game.state.board[from.row][from.col];
        const promotion = piece?.type === 'pawn' && pos.row === 0
            ? await choosePromotion() : 'queen';
        if (promotion === null) return;
        if (!game || busy || !selected || !same(selected, from)) return;
        busy = true;
        ++selectionVersion;
        message = '';
        render();
        try {
            game = await makeMove(game.gameId, from, pos, promotion);
            selected = null;
            legal = [];
            busy = false;
            render();
            await computerTurn();
        } catch (error) {
            showError(error);
        } finally {
            busy = false;
            render();
        }
    } else if (selected && same(pos, selected)) {
        ++selectionVersion;
        selected = null;
        legal = [];
        render();
    } else {
        await selectSquare(pos);
    }
}

async function changeGame(action: () => Promise<GameState>): Promise<void> {
    if (busy) return;
    busy = true;
    ++selectionVersion;
    selected = null;
    legal = [];
    message = '';
    render();
    try {
        game = await action();
        sessionStorage.setItem('chessGameId', game.gameId);
        if (game.state.currentTurn === 'black') {
            busy = false;
            render();
            await computerTurn();
        }
    } catch (error) {
        showError(error);
    } finally {
        busy = false;
        render();
    }
}

function positionFrom(target: EventTarget | null): Position | null {
    const square = (target as HTMLElement | null)?.closest?.('.square') as HTMLElement | null;
    if (!square) return null;
    return { row: Number(square.dataset.row), col: Number(square.dataset.col) };
}

async function init(): Promise<void> {
    document.getElementById('board')!.addEventListener('click', event => {
        const pos = positionFrom(event.target);
        if (pos) void boardAction(pos);
    });
    document.getElementById('board')!.addEventListener('dragstart', event => {
        const pos = positionFrom(event.target);
        if (pos && game?.state.board[pos.row][pos.col]?.color === 'white') {
            event.dataTransfer?.setData('text/plain', JSON.stringify(pos));
            void selectSquare(pos);
        } else event.preventDefault();
    });
    document.getElementById('board')!.addEventListener('dragover', event => event.preventDefault());
    document.getElementById('board')!.addEventListener('drop', event => {
        event.preventDefault();
        const pos = positionFrom(event.target);
        const source = event.dataTransfer?.getData('text/plain');
        if (pos && source && game && !busy) {
            const from: Position = JSON.parse(source) as Position;
            void (async () => {
                await selectSquare(from);
                await boardAction(pos);
            })();
        }
    });
    document.getElementById('new-game')!.addEventListener('click', () => void changeGame(createGame));
    document.getElementById('undo')!.addEventListener('click', () => {
        if (game) void changeGame(() => undoMove(game!.gameId));
    });
    document.getElementById('redo')!.addEventListener('click', () => {
        if (game) void changeGame(() => redoMove(game!.gameId));
    });
    document.getElementById('retry-ai')!.addEventListener('click', () => void computerTurn());
    document.querySelectorAll<HTMLButtonElement>('#promotion-dialog .promotion-options button')
        .forEach(button => button.addEventListener('click', () => {
            (document.getElementById('promotion-dialog') as HTMLDialogElement).close(button.value);
        }));
    document.getElementById('cancel-promotion')!.addEventListener('click', () => {
        (document.getElementById('promotion-dialog') as HTMLDialogElement).close();
    });
    document.getElementById('flip-board')!.addEventListener('click', () => {
        flipped = !flipped;
        render();
    });
    try {
        const saved = sessionStorage.getItem('chessGameId');
        if (saved) {
            try { game = await getGameState(saved); }
            catch { sessionStorage.removeItem('chessGameId'); }
        }
        if (!game) game = await createGame();
        sessionStorage.setItem('chessGameId', game.gameId);
        render();
        await computerTurn();
    } catch (error) {
        showError(error);
        document.getElementById('game-status')!.textContent = message;
    }
}

document.addEventListener('DOMContentLoaded', () => void init());
