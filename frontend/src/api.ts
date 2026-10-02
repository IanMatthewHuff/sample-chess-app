/**
 * API client for communicating with the Python backend
 */

import { GameState } from './game';
import { Move, Position } from './pieces';
import { BoardState } from './board';

const API_BASE_URL = 'http://localhost:5000/api';

export interface ApiResponse<T> {
    success: boolean;
    data?: T;
    error?: string;
}

export interface GameResponse {
    gameId: string;
    state: BoardState;
    status: string;
}

export interface MoveResponse {
    valid: boolean;
    state: BoardState;
    status: string;
}

export interface ValidationResponse {
    valid: boolean;
    reason?: string;
}

/**
 * Create a new game on the server
 */
export async function createGame(): Promise<ApiResponse<GameResponse>> {
    // TODO: Implement API call to create new game
    try {
        const response = await fetch(`${API_BASE_URL}/games`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        const data = await response.json();
        return { success: true, data };
    } catch (error) {
        return { success: false, error: (error as Error).message };
    }
}

/**
 * Get current game state
 */
export async function getGameState(gameId: string): Promise<ApiResponse<GameResponse>> {
    // TODO: Implement API call to get game state
    try {
        const response = await fetch(`${API_BASE_URL}/games/${gameId}`);
        const data = await response.json();
        return { success: true, data };
    } catch (error) {
        return { success: false, error: (error as Error).message };
    }
}

/**
 * Make a move
 */
export async function makeMove(gameId: string, from: Position, to: Position): Promise<ApiResponse<MoveResponse>> {
    // TODO: Implement API call to make move
    try {
        const response = await fetch(`${API_BASE_URL}/games/${gameId}/move`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ from, to })
        });
        
        const data = await response.json();
        return { success: true, data };
    } catch (error) {
        return { success: false, error: (error as Error).message };
    }
}

/**
 * Validate a move with the Go engine
 */
export async function validateMove(gameId: string, from: Position, to: Position): Promise<ApiResponse<ValidationResponse>> {
    // TODO: Implement API call to validate move
    try {
        const response = await fetch(`${API_BASE_URL}/games/${gameId}/validate`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ from, to })
        });
        
        const data = await response.json();
        return { success: true, data };
    } catch (error) {
        return { success: false, error: (error as Error).message };
    }
}

/**
 * Get valid moves for a position
 */
export async function getValidMoves(gameId: string, position: Position): Promise<ApiResponse<Position[]>> {
    // TODO: Implement API call to get valid moves
    try {
        const response = await fetch(
            `${API_BASE_URL}/games/${gameId}/moves?row=${position.row}&col=${position.col}`
        );
        const data = await response.json();
        return { success: true, data };
    } catch (error) {
        return { success: false, error: (error as Error).message };
    }
}

/**
 * Undo last move
 */
export async function undoMove(gameId: string): Promise<ApiResponse<GameResponse>> {
    // TODO: Implement API call to undo move
    try {
        const response = await fetch(`${API_BASE_URL}/games/${gameId}/undo`, {
            method: 'POST'
        });
        const data = await response.json();
        return { success: true, data };
    } catch (error) {
        return { success: false, error: (error as Error).message };
    }
}

/**
 * Health check for API
 */
export async function healthCheck(): Promise<boolean> {
    // TODO: Implement health check
    try {
        const response = await fetch(`${API_BASE_URL}/health`);
        return response.ok;
    } catch {
        return false;
    }
}
