/**
 * API client for communicating with the Python backend
 */

import { MoveTarget } from './board';
import { Difficulty, GameMode, GameView } from './game';
import { PieceColor, Position, PromotionPiece } from './pieces';

// Defaults to same-origin `/api`, which the Vite dev server proxies to the backend.
const API_BASE_URL: string = import.meta.env.VITE_API_URL || '/api';

export class ApiError extends Error {
    constructor(message: string, public status: number) {
        super(message);
    }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
    let response: Response;
    try {
        response = await fetch(`${API_BASE_URL}${path}`, {
            ...init,
            headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) }
        });
    } catch {
        throw new ApiError('Cannot reach the game server. Is the backend running?', 0);
    }

    let data: unknown = null;
    try {
        data = await response.json();
    } catch {
        // non-JSON body
    }
    if (!response.ok) {
        const message = (data as { error?: string } | null)?.error ?? `Request failed (${response.status})`;
        throw new ApiError(message, response.status);
    }
    return data as T;
}

const post = <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: body === undefined ? undefined : JSON.stringify(body) });

export interface NewGameOptions {
    mode: GameMode;
    humanColor: PieceColor;
    difficulty: Difficulty;
}

export const createGame = (opts: NewGameOptions) => post<GameView>('/games', opts);

export const getGame = (gameId: string) => request<GameView>(`/games/${gameId}`);

export async function getValidMoves(gameId: string, position: Position): Promise<MoveTarget[]> {
    const data = await request<{ moves: MoveTarget[] }>(
        `/games/${gameId}/moves?row=${position.row}&col=${position.col}`
    );
    return data.moves;
}

export const makeMove = (gameId: string, from: Position, to: Position, promotion?: PromotionPiece) =>
    post<GameView>(`/games/${gameId}/move`, { from, to, promotion });

export const requestAIMove = (gameId: string) => post<GameView>(`/games/${gameId}/ai-move`);

export const undoMove = (gameId: string) => post<GameView>(`/games/${gameId}/undo`);

export const redoMove = (gameId: string) => post<GameView>(`/games/${gameId}/redo`);

export const resignGame = (gameId: string) => post<GameView>(`/games/${gameId}/resign`);

export async function healthCheck(): Promise<{ ok: boolean; engine: string }> {
    try {
        const data = await request<{ engine: string }>('/health');
        return { ok: true, engine: data.engine };
    } catch (e) {
        return { ok: false, engine: e instanceof ApiError && e.status === 503 ? 'unreachable' : 'unknown' };
    }
}
