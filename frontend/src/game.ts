import type { BoardState } from './board';
import type { Move } from './pieces';

export type GameStatus = 'active' | 'check' | 'checkmate' | 'stalemate' | 'draw';

export interface GameState {
    gameId: string;
    state: BoardState;
    status: GameStatus;
    history: Move[];
    canUndo: boolean;
    canRedo: boolean;
}
