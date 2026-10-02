/**
 * Game view types. The backend owns game state; the frontend only keeps
 * transient UI state (selection, highlighted targets, pending promotion).
 */

import { BoardState, MoveTarget } from './board';
import { Move, PieceColor, PieceType, Position } from './pieces';

export type GameMode = 'ai' | 'local';
export type Difficulty = 'easy' | 'medium' | 'hard';
export type GameStatusName = 'active' | 'check' | 'checkmate' | 'stalemate' | 'draw' | 'resigned';

export interface GameStatus {
    status: GameStatusName;
    inCheck: boolean;
    gameOver: boolean;
    winner?: PieceColor;
    reason?: string;
}

export interface GameView {
    gameId: string;
    mode: GameMode;
    humanColor: PieceColor;
    difficulty: Difficulty;
    state: BoardState;
    status: GameStatus;
    fen: string;
    moves: Move[];
    lastMove: Move | null;
    captured: Record<PieceColor, PieceType[]>;
    canUndo: boolean;
    canRedo: boolean;
    aiToMove: boolean;
}

export interface UIState {
    game: GameView | null;
    selected: Position | null;
    targets: MoveTarget[];
    pendingPromotion: { from: Position; to: Position } | null;
    busy: boolean;
    aiThinking: boolean;
    flipped: boolean;
    error: string | null;
}

export function createUIState(): UIState {
    return {
        game: null,
        selected: null,
        targets: [],
        pendingPromotion: null,
        busy: false,
        aiThinking: false,
        flipped: false,
        error: null
    };
}

/**
 * Human-readable description of the game status
 */
export function describeStatus(game: GameView): string {
    const s = game.status;
    const winner = s.winner ? s.winner.charAt(0).toUpperCase() + s.winner.slice(1) : '';
    switch (s.status) {
        case 'checkmate':
            return `Checkmate — ${winner} wins`;
        case 'resigned':
            return `${s.reason ? s.reason.charAt(0).toUpperCase() + s.reason.slice(1) : 'Resigned'} — ${winner} wins`;
        case 'stalemate':
            return 'Stalemate — draw';
        case 'draw':
            return `Draw by ${s.reason ?? 'agreement'}`;
        case 'check':
            return 'Check!';
        default:
            return '';
    }
}

/**
 * Whether the human can interact with the board right now
 */
export function isHumanTurn(game: GameView): boolean {
    if (game.status.gameOver) return false;
    return game.mode === 'local' || game.state.currentTurn === game.humanColor;
}
