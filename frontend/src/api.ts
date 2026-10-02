import type { GameState } from './game';
import type { LegalMove, PieceType, Position } from './pieces';

const BASE = '/api';

async function request<T>(path: string, body?: object, method = 'GET'): Promise<T> {
    const response = await fetch(`${BASE}${path}`, {
        method,
        headers: body ? { 'Content-Type': 'application/json' } : undefined,
        body: body ? JSON.stringify(body) : undefined
    });
    const data: unknown = await response.json();
    if (!response.ok) {
        const message = typeof data === 'object' && data !== null && 'error' in data
            ? String(data.error) : `Request failed (${response.status})`;
        throw new Error(message);
    }
    return data as T;
}

export const createGame = () => request<GameState>('/games', undefined, 'POST');
export const getGameState = (id: string) => request<GameState>(`/games/${encodeURIComponent(id)}`);
export const makeMove = (id: string, from: Position, to: Position, promotion: PieceType) =>
    request<GameState>(`/games/${encodeURIComponent(id)}/move`, { from, to, promotion }, 'POST');
export const makeAiMove = (id: string) =>
    request<GameState>(`/games/${encodeURIComponent(id)}/ai`, undefined, 'POST');
export const getValidMoves = (id: string, position: Position) =>
    request<LegalMove[]>(`/games/${encodeURIComponent(id)}/moves?row=${position.row}&col=${position.col}`);
export const undoMove = (id: string) =>
    request<GameState>(`/games/${encodeURIComponent(id)}/undo`, undefined, 'POST');
export const redoMove = (id: string) =>
    request<GameState>(`/games/${encodeURIComponent(id)}/redo`, undefined, 'POST');
