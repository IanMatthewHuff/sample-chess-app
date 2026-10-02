/**
 * Chess game logic and state management
 */

import { Piece, Position, Move, PieceColor, PieceType } from './pieces';
import { Board, BoardState, createInitialBoardState, getPieceAt, setPieceAt, cloneBoard, findKingPosition, isValidPosition } from './board';

export interface GameState {
    boardState: BoardState;
    moveHistory: Move[];
    redoStack: Move[];
    selectedSquare: Position | null;
    validMoves: Position[];
    gameStatus: GameStatus;
}

export type GameStatus = 'active' | 'check' | 'checkmate' | 'stalemate' | 'draw';

/**
 * Create a new game
 */
export function createNewGame(): GameState {
    // TODO: Implement new game creation
    return {
        boardState: createInitialBoardState(),
        moveHistory: [],
        redoStack: [],
        selectedSquare: null,
        validMoves: [],
        gameStatus: 'active'
    };
}

/**
 * Get all valid moves for a piece at a given position
 */
export function getValidMoves(state: BoardState, position: Position): Position[] {
    // TODO: Implement valid move calculation
    const piece = getPieceAt(state.board, position);
    if (!piece || piece.color !== state.currentTurn) {
        return [];
    }
    
    const moves: Position[] = [];
    
    switch (piece.type) {
        case 'pawn':
            moves.push(...getPawnMoves(state, position, piece));
            break;
        case 'knight':
            moves.push(...getKnightMoves(state, position, piece));
            break;
        case 'bishop':
            moves.push(...getBishopMoves(state, position, piece));
            break;
        case 'rook':
            moves.push(...getRookMoves(state, position, piece));
            break;
        case 'queen':
            moves.push(...getQueenMoves(state, position, piece));
            break;
        case 'king':
            moves.push(...getKingMoves(state, position, piece));
            break;
    }
    
    // Filter out moves that would leave king in check
    return moves.filter(move => !wouldBeInCheck(state, position, move));
}

/**
 * Get pawn moves
 */
function getPawnMoves(state: BoardState, position: Position, piece: Piece): Position[] {
    // TODO: Implement pawn move logic
    const moves: Position[] = [];
    const direction = piece.color === 'white' ? -1 : 1;
    const startRow = piece.color === 'white' ? 6 : 1;
    
    // Single move forward
    const oneForward = { row: position.row + direction, col: position.col };
    if (isValidPosition(oneForward) && !getPieceAt(state.board, oneForward)) {
        moves.push(oneForward);
        
        // Double move from starting position
        if (position.row === startRow) {
            const twoForward = { row: position.row + 2 * direction, col: position.col };
            if (!getPieceAt(state.board, twoForward)) {
                moves.push(twoForward);
            }
        }
    }
    
    // Captures
    for (const colOffset of [-1, 1]) {
        const capturePos = { row: position.row + direction, col: position.col + colOffset };
        if (isValidPosition(capturePos)) {
            const target = getPieceAt(state.board, capturePos);
            if (target && target.color !== piece.color) {
                moves.push(capturePos);
            }
            // En passant
            if (state.enPassantTarget && 
                capturePos.row === state.enPassantTarget.row && 
                capturePos.col === state.enPassantTarget.col) {
                moves.push(capturePos);
            }
        }
    }
    
    return moves;
}

/**
 * Get knight moves
 */
function getKnightMoves(state: BoardState, position: Position, piece: Piece): Position[] {
    // TODO: Implement knight move logic
    const moves: Position[] = [];
    const offsets = [
        [-2, -1], [-2, 1], [-1, -2], [-1, 2],
        [1, -2], [1, 2], [2, -1], [2, 1]
    ];
    
    for (const [rowOffset, colOffset] of offsets) {
        const newPos = { row: position.row + rowOffset, col: position.col + colOffset };
        if (isValidPosition(newPos)) {
            const target = getPieceAt(state.board, newPos);
            if (!target || target.color !== piece.color) {
                moves.push(newPos);
            }
        }
    }
    
    return moves;
}

/**
 * Get bishop moves
 */
function getBishopMoves(state: BoardState, position: Position, piece: Piece): Position[] {
    // TODO: Implement bishop move logic
    return getSlidingMoves(state, position, piece, [[-1, -1], [-1, 1], [1, -1], [1, 1]]);
}

/**
 * Get rook moves
 */
function getRookMoves(state: BoardState, position: Position, piece: Piece): Position[] {
    // TODO: Implement rook move logic
    return getSlidingMoves(state, position, piece, [[-1, 0], [1, 0], [0, -1], [0, 1]]);
}

/**
 * Get queen moves
 */
function getQueenMoves(state: BoardState, position: Position, piece: Piece): Position[] {
    // TODO: Implement queen move logic
    return getSlidingMoves(state, position, piece, [
        [-1, -1], [-1, 0], [-1, 1],
        [0, -1], [0, 1],
        [1, -1], [1, 0], [1, 1]
    ]);
}

/**
 * Get king moves
 */
function getKingMoves(state: BoardState, position: Position, piece: Piece): Position[] {
    // TODO: Implement king move logic (including castling)
    const moves: Position[] = [];
    const offsets = [
        [-1, -1], [-1, 0], [-1, 1],
        [0, -1], [0, 1],
        [1, -1], [1, 0], [1, 1]
    ];
    
    for (const [rowOffset, colOffset] of offsets) {
        const newPos = { row: position.row + rowOffset, col: position.col + colOffset };
        if (isValidPosition(newPos)) {
            const target = getPieceAt(state.board, newPos);
            if (!target || target.color !== piece.color) {
                moves.push(newPos);
            }
        }
    }
    
    // TODO: Add castling logic
    
    return moves;
}

/**
 * Get sliding piece moves (for bishop, rook, queen)
 */
function getSlidingMoves(state: BoardState, position: Position, piece: Piece, directions: number[][]): Position[] {
    // TODO: Implement sliding move logic
    const moves: Position[] = [];
    
    for (const [rowDir, colDir] of directions) {
        let newPos = { row: position.row + rowDir, col: position.col + colDir };
        
        while (isValidPosition(newPos)) {
            const target = getPieceAt(state.board, newPos);
            
            if (!target) {
                moves.push({ ...newPos });
            } else {
                if (target.color !== piece.color) {
                    moves.push({ ...newPos });
                }
                break;
            }
            
            newPos = { row: newPos.row + rowDir, col: newPos.col + colDir };
        }
    }
    
    return moves;
}

/**
 * Check if making a move would leave the king in check
 */
function wouldBeInCheck(state: BoardState, from: Position, to: Position): boolean {
    // TODO: Implement check detection after move
    const testBoard = cloneBoard(state.board);
    const piece = getPieceAt(testBoard, from);
    
    if (!piece) return true;
    
    setPieceAt(testBoard, to, piece);
    setPieceAt(testBoard, from, null);
    
    return isInCheck({ ...state, board: testBoard }, piece.color);
}

/**
 * Check if a player is in check
 */
export function isInCheck(state: BoardState, color: PieceColor): boolean {
    // TODO: Implement check detection
    const kingPos = findKingPosition(state.board, color);
    if (!kingPos) return false;
    
    const opponentColor = color === 'white' ? 'black' : 'white';
    
    // Check all opponent pieces
    for (let row = 0; row < 8; row++) {
        for (let col = 0; col < 8; col++) {
            const piece = getPieceAt(state.board, { row, col });
            if (piece && piece.color === opponentColor) {
                // TODO: Check if this piece can attack the king
            }
        }
    }
    
    return false;
}

/**
 * Make a move
 */
export function makeMove(gameState: GameState, from: Position, to: Position): GameState {
    // TODO: Implement move execution
    const piece = getPieceAt(gameState.boardState.board, from);
    if (!piece) return gameState;
    
    const newBoard = cloneBoard(gameState.boardState.board);
    const captured = getPieceAt(newBoard, to);
    
    setPieceAt(newBoard, to, { ...piece, hasMoved: true });
    setPieceAt(newBoard, from, null);
    
    const move: Move = {
        from,
        to,
        piece,
        captured: captured || undefined
    };
    
    const newBoardState: BoardState = {
        ...gameState.boardState,
        board: newBoard,
        currentTurn: gameState.boardState.currentTurn === 'white' ? 'black' : 'white',
        fullMoveNumber: gameState.boardState.currentTurn === 'black' 
            ? gameState.boardState.fullMoveNumber + 1 
            : gameState.boardState.fullMoveNumber
    };
    
    return {
        ...gameState,
        boardState: newBoardState,
        moveHistory: [...gameState.moveHistory, move],
        redoStack: [],
        selectedSquare: null,
        validMoves: []
    };
}

/**
 * Undo the last move
 */
export function undoMove(gameState: GameState): GameState {
    // TODO: Implement undo
    if (gameState.moveHistory.length === 0) return gameState;
    
    const lastMove = gameState.moveHistory[gameState.moveHistory.length - 1];
    const newBoard = cloneBoard(gameState.boardState.board);
    
    setPieceAt(newBoard, lastMove.from, { ...lastMove.piece, hasMoved: false });
    setPieceAt(newBoard, lastMove.to, lastMove.captured || null);
    
    return {
        ...gameState,
        boardState: {
            ...gameState.boardState,
            board: newBoard,
            currentTurn: gameState.boardState.currentTurn === 'white' ? 'black' : 'white'
        },
        moveHistory: gameState.moveHistory.slice(0, -1),
        redoStack: [...gameState.redoStack, lastMove]
    };
}

/**
 * Redo a previously undone move
 */
export function redoMove(gameState: GameState): GameState {
    // TODO: Implement redo
    if (gameState.redoStack.length === 0) return gameState;
    
    const moveToRedo = gameState.redoStack[gameState.redoStack.length - 1];
    return makeMove(
        { ...gameState, redoStack: gameState.redoStack.slice(0, -1) },
        moveToRedo.from,
        moveToRedo.to
    );
}

/**
 * Get game status (check, checkmate, stalemate, etc.)
 */
export function getGameStatus(state: BoardState): GameStatus {
    // TODO: Implement game status detection
    const currentPlayerInCheck = isInCheck(state, state.currentTurn);
    
    // Check if current player has any valid moves
    let hasValidMoves = false;
    outer: for (let row = 0; row < 8; row++) {
        for (let col = 0; col < 8; col++) {
            const piece = getPieceAt(state.board, { row, col });
            if (piece && piece.color === state.currentTurn) {
                const moves = getValidMoves(state, { row, col });
                if (moves.length > 0) {
                    hasValidMoves = true;
                    break outer;
                }
            }
        }
    }
    
    if (!hasValidMoves) {
        return currentPlayerInCheck ? 'checkmate' : 'stalemate';
    }
    
    return currentPlayerInCheck ? 'check' : 'active';
}

/**
 * Convert move to algebraic notation
 */
export function moveToAlgebraic(move: Move, state: BoardState): string {
    // TODO: Implement algebraic notation generation
    throw new Error('Not implemented');
}
