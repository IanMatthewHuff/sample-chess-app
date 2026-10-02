/**
 * Chess App Entry Point
 */

import { renderBoard } from './board';
import { GameState, createNewGame, getValidMoves, makeMove, undoMove, redoMove, getGameStatus } from './game';
import { Position } from './pieces';

// Global game state
let gameState: GameState;
let boardFlipped = false;

/**
 * Initialize the chess application
 */
function init(): void {
    gameState = createNewGame();
    
    setupEventListeners();
    render();
    
    console.log('Chess app initialized');
}

/**
 * Setup event listeners for UI controls
 */
function setupEventListeners(): void {
    // New game button
    const newGameBtn = document.getElementById('new-game');
    newGameBtn?.addEventListener('click', handleNewGame);
    
    // Undo button
    const undoBtn = document.getElementById('undo');
    undoBtn?.addEventListener('click', handleUndo);
    
    // Redo button
    const redoBtn = document.getElementById('redo');
    redoBtn?.addEventListener('click', handleRedo);
    
    // Flip board button
    const flipBtn = document.getElementById('flip-board');
    flipBtn?.addEventListener('click', handleFlipBoard);
    
    // Board click handling
    const board = document.getElementById('board');
    board?.addEventListener('click', handleBoardClick);
}

/**
 * Handle board square click
 */
function handleBoardClick(event: MouseEvent): void {
    const target = event.target as HTMLElement;
    const square = target.closest('.square') as HTMLElement;
    
    if (!square) return;
    
    const row = parseInt(square.dataset.row || '0');
    const col = parseInt(square.dataset.col || '0');
    const position: Position = { row, col };
    
    if (gameState.selectedSquare) {
        // Check if clicking on a valid move destination
        const isValidMove = gameState.validMoves.some(
            move => move.row === row && move.col === col
        );
        
        if (isValidMove) {
            gameState = makeMove(gameState, gameState.selectedSquare, position);
            updateGameStatus();
        } else {
            // Select new square or deselect
            selectSquare(position);
        }
    } else {
        selectSquare(position);
    }
    
    render();
}

/**
 * Select a square and calculate valid moves
 */
function selectSquare(position: Position): void {
    // TODO: Implement square selection
    const piece = gameState.boardState.board[position.row][position.col];
    
    if (piece && piece.color === gameState.boardState.currentTurn) {
        gameState.selectedSquare = position;
        gameState.validMoves = getValidMoves(gameState.boardState, position);
    } else {
        gameState.selectedSquare = null;
        gameState.validMoves = [];
    }
}

/**
 * Handle new game button click
 */
function handleNewGame(): void {
    // TODO: Implement new game handling
    gameState = createNewGame();
    render();
}

/**
 * Handle undo button click
 */
function handleUndo(): void {
    // TODO: Implement undo handling
    gameState = undoMove(gameState);
    render();
}

/**
 * Handle redo button click
 */
function handleRedo(): void {
    // TODO: Implement redo handling
    gameState = redoMove(gameState);
    render();
}

/**
 * Handle flip board button click
 */
function handleFlipBoard(): void {
    // TODO: Implement board flip
    boardFlipped = !boardFlipped;
    render();
}

/**
 * Update game status display
 */
function updateGameStatus(): void {
    // TODO: Implement status update
    gameState.gameStatus = getGameStatus(gameState.boardState);
    
    const statusEl = document.getElementById('game-status');
    if (statusEl) {
        switch (gameState.gameStatus) {
            case 'check':
                statusEl.textContent = 'Check!';
                break;
            case 'checkmate':
                const winner = gameState.boardState.currentTurn === 'white' ? 'Black' : 'White';
                statusEl.textContent = `Checkmate! ${winner} wins!`;
                break;
            case 'stalemate':
                statusEl.textContent = 'Stalemate! Draw.';
                break;
            default:
                statusEl.textContent = '';
        }
    }
}

/**
 * Render the game UI
 */
function render(): void {
    // Render board
    renderBoard(gameState.boardState.board, 'board', boardFlipped);
    
    // Highlight selected square
    if (gameState.selectedSquare) {
        highlightSquare(gameState.selectedSquare, 'selected');
    }
    
    // Highlight valid moves
    for (const move of gameState.validMoves) {
        const piece = gameState.boardState.board[move.row][move.col];
        highlightSquare(move, piece ? 'capture-move' : 'valid-move');
    }
    
    // Update turn indicator
    const turnIndicator = document.getElementById('turn-indicator');
    if (turnIndicator) {
        const color = gameState.boardState.currentTurn;
        turnIndicator.textContent = `${color.charAt(0).toUpperCase() + color.slice(1)} to move`;
    }
    
    // Update move history
    renderMoveHistory();
}

/**
 * Highlight a square on the board
 */
function highlightSquare(position: Position, className: string): void {
    // TODO: Implement square highlighting
    const squares = document.querySelectorAll('.square');
    squares.forEach(square => {
        const el = square as HTMLElement;
        if (el.dataset.row === position.row.toString() && 
            el.dataset.col === position.col.toString()) {
            el.classList.add(className);
        }
    });
}

/**
 * Render move history
 */
function renderMoveHistory(): void {
    // TODO: Implement move history rendering
    const historyEl = document.getElementById('move-history');
    if (!historyEl) return;
    
    historyEl.innerHTML = gameState.moveHistory
        .map((move, index) => {
            const moveNum = Math.floor(index / 2) + 1;
            const piece = move.piece.type;
            return `${index % 2 === 0 ? moveNum + '. ' : ''}${piece}`;
        })
        .join(' ');
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', init);

// Export for potential testing
export { gameState, init };
