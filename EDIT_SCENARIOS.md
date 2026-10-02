# Edit Scenarios for AI Code Editor Benchmarking

Specific, targeted edit scenarios for testing inline suggestions and Next Edit Suggestions (NES).

---

## TypeScript (frontend/src/)

### Type Changes

1. **Change `PieceColor` from union type to enum**
   - File: `pieces.ts`
   - Change: `type PieceColor = 'white' | 'black'` → `enum PieceColor { White = 'white', Black = 'black' }`
   - Cascading edits required in multiple files

2. **Change `hasMoved` property type from `boolean` to `number` (move count)**
   - File: `pieces.ts`
   - Change: `hasMoved: boolean` → `moveCount: number`
   - Update all usages in `board.ts`, `game.ts`

3. **Add optional `id` property to `Piece` interface**
   - File: `pieces.ts`
   - Add: `id?: string` to the `Piece` interface

4. **Change `Board` type from 2D array to Map**
   - File: `board.ts`
   - Change: `type Board = (Piece | null)[][]` → `type Board = Map<string, Piece>`

5. **Change `Position` to use single index instead of row/col**
   - File: `pieces.ts`
   - Change: `{ row: number; col: number }` → `{ index: number }` (0-63)

### Function Signature Changes

6. **Add `promotionPiece` parameter to `makeMove` function**
   - File: `game.ts`
   - Change: `makeMove(gameState: GameState, from: Position, to: Position)` → `makeMove(gameState: GameState, from: Position, to: Position, promotionPiece?: PieceType)`

7. **Change `renderBoard` to return HTMLElement instead of void**
   - File: `board.ts`
   - Change: `renderBoard(...): void` → `renderBoard(...): HTMLElement`

8. **Add async/await to `handleBoardClick`**
   - File: `index.ts`
   - Change: `function handleBoardClick(event: MouseEvent): void` → `async function handleBoardClick(event: MouseEvent): Promise<void>`

9. **Change `getValidMoves` return type to include move metadata**
   - File: `game.ts`
   - Change: `getValidMoves(...): Position[]` → `getValidMoves(...): { position: Position; isCapture: boolean; isCheck: boolean }[]`

10. **Add timeout parameter to API functions**
    - File: `api.ts`
    - Add: `timeout?: number` parameter to `createGame`, `makeMove`, `getGameState`

### Variable/Constant Changes

11. **Extract magic number 8 to constant `BOARD_SIZE`**
    - File: `board.ts`
    - Add: `const BOARD_SIZE = 8`
    - Replace all `8` with `BOARD_SIZE`

12. **Change `API_BASE_URL` to be configurable from environment**
    - File: `api.ts`
    - Change: `const API_BASE_URL = 'http://localhost:5000/api'` → `const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:5000/api'`

13. **Rename `gameState` to `currentGame` in index.ts**
    - File: `index.ts`
    - Global rename of the variable

14. **Change piece symbols from Unicode to image paths**
    - File: `pieces.ts`
    - Change: `PIECE_SYMBOLS` values from `'♔'` etc. to `'/images/white-king.svg'` etc.

---

## Python (backend/)

### Type Annotation Changes

15. **Add return type annotation to `create_initial_board`**
    - File: `board.py`
    - Change: `def create_initial_board()` → `def create_initial_board() -> Board`

16. **Change `Board` type alias to use numpy array**
    - File: `board.py`
    - Change: `Board = List[List[Optional[Piece]]]` → `Board = np.ndarray`

17. **Add `@dataclass(frozen=True)` to make `Position` immutable**
    - File: `pieces.py`
    - Change: `@dataclass` → `@dataclass(frozen=True)`

18. **Change `PieceColor` enum to use `auto()` values**
    - File: `pieces.py`
    - Change: `WHITE = "white"` → `WHITE = auto()`

19. **Add generic type parameter to `ApiResponse`**
    - File: `app.py`
    - Add proper typing with `TypeVar` and `Generic`

### Function Changes

20. **Add `depth` parameter to move validation for engine analysis**
    - File: `game_logic.py`
    - Change: `def get_valid_moves(state: BoardState, position: Position)` → `def get_valid_moves(state: BoardState, position: Position, depth: int = 1)`

21. **Convert `get_piece_at` to use `__getitem__` on Board class**
    - File: `board.py`
    - Add `__getitem__` method to a new `Board` class

22. **Add caching decorator to `get_valid_moves`**
    - File: `game_logic.py`
    - Add: `@lru_cache(maxsize=1000)` decorator

23. **Change `to_dict` methods to use Pydantic `model_dump`**
    - Files: `pieces.py`, `board.py`
    - Convert dataclasses to Pydantic models

24. **Add logging to `make_move` function**
    - File: `game_logic.py`
    - Add: `logger.info(f"Move: {from_pos} -> {to_pos}")` at start

### Variable/Structure Changes

25. **Rename `games` dict to `game_sessions` with type annotation**
    - File: `app.py`
    - Change: `games: Dict[str, GameState] = {}` → `game_sessions: Dict[str, GameState] = {}`

26. **Add `created_at` timestamp to `GameState`**
    - File: `game_logic.py`
    - Add: `created_at: datetime = field(default_factory=datetime.now)`

27. **Change `half_move_clock` from int to `HalfMoveClock` class**
    - File: `board.py`
    - Create new class with validation logic

28. **Extract piece movement offsets to constants**
    - File: `game_logic.py`
    - Add: `KNIGHT_OFFSETS`, `BISHOP_DIRECTIONS`, `ROOK_DIRECTIONS` constants

---

## Go (engine/)

### Type Changes

29. **Change `PieceColor` from string to int constant**
    - File: `pieces.go`
    - Change: `type PieceColor string` → `type PieceColor int` with `iota`

30. **Add `LastMove` field to `BoardState` struct**
    - File: `board.go`
    - Add: `LastMove *Move` field

31. **Change `Board` from array to struct with methods**
    - File: `board.go`
    - Change: `type Board [8][8]*Piece` → `type Board struct { squares [64]*Piece }`

32. **Add `Score` field to `ValidationResponse`**
    - File: `main.go`
    - Add: `Score int` for position evaluation

33. **Change `Position` to use named fields with tags**
    - File: `pieces.go`
    - Add: JSON tags `json:"row"` and validation tags

### Function Changes

34. **Add context parameter to HTTP handlers**
    - File: `main.go`
    - Change handlers to accept `context.Context`

35. **Change `GetValidMoves` to return channel for concurrent generation**
    - File: `moves.go`
    - Change: `func GetValidMoves(...) []Position` → `func GetValidMoves(...) <-chan Position`

36. **Add error return to `MakeMove` function**
    - File: `moves.go`
    - Change: `func MakeMove(...) *BoardState` → `func MakeMove(...) (*BoardState, error)`

37. **Extract `corsMiddleware` to separate middleware package**
    - File: `main.go`
    - Create `middleware/cors.go` with the function

38. **Add benchmark function for `GetValidMoves`**
    - File: `moves_test.go` (new)
    - Add: `func BenchmarkGetValidMoves(b *testing.B)`

### Variable/Constant Changes

39. **Change hardcoded port to environment variable**
    - File: `main.go`
    - Change: `":8080"` → `os.Getenv("PORT")` with default

40. **Rename `offsets` to `knightMoveOffsets` with package-level const**
    - File: `moves.go`
    - Extract local variable to package constant

41. **Add `var ErrInvalidMove` error variable**
    - File: `moves.go`
    - Add: `var ErrInvalidMove = errors.New("invalid move")`

42. **Change piece values to configurable map**
    - File: `pieces.go`
    - Change: hardcoded switch to `var PieceValues = map[PieceType]int{...}`

---

## Cross-Language Scenarios

### API Contract Changes (requires edits in all 3 languages)

43. **Add `timestamp` field to move responses**
    - Files: `api.ts`, `app.py`, `main.go`
    - Add timestamp to all move-related responses

44. **Rename `currentTurn` to `activePlayer` everywhere**
    - Files: All files containing turn/player logic
    - Global rename across the codebase

45. **Change position format from `{row, col}` to algebraic notation string**
    - Files: All position-related code
    - Change: `{row: 0, col: 4}` → `"e8"`

46. **Add `gameVersion` field to all API responses**
    - Files: `api.ts`, `app.py`, `main.go`
    - Add version tracking for state sync

---

## New Code Block Generation Scenarios (Progressive Reveal Testing)

These scenarios require generating substantial new code blocks, ideal for testing progressive reveal functionality.

### TypeScript - New Functions

47. **Add `isCheckmate` function to game.ts**
    - File: `game.ts`
    - Generate: Complete function that checks if the current player is in checkmate
    - Expected ~20-30 lines of logic

48. **Implement `moveToAlgebraic` function in game.ts**
    - File: `game.ts`

49. **Add `getMoveHistory` function with algebraic notation**
    - File: `game.ts`
    - Generate: Function that returns move history in standard chess notation (e.g., "e4", "Nf3", "O-O")
    - Expected ~15-25 lines

50. **Add `highlightValidMoves` function to board.ts**
    - File: `board.ts`
    - Generate: Function that adds CSS classes to valid move squares with different colors for captures vs moves
    - Expected ~20 lines

51. **Add keyboard navigation handler**
    - File: `index.ts`
    - Generate: `handleKeyboardNavigation(event: KeyboardEvent)` function for arrow key piece selection
    - Expected ~30-40 lines

### TypeScript - New Classes/Interfaces

52. **Add `MoveValidator` class**
    - File: `game.ts` (or new file)
    - Generate: Complete class with methods for each piece type's move validation
    - Expected ~50-80 lines

53. **Add `GameTimer` class for chess clock functionality**
    - File: new `timer.ts`
    - Generate: Class with `start()`, `pause()`, `switchPlayer()`, `getTimeRemaining()` methods
    - Expected ~40-60 lines

54. **Add `PGNExporter` class for game export**
    - File: new `pgn.ts`
    - Generate: Class that exports game state to PGN (Portable Game Notation) format
    - Expected ~50-70 lines

### Python - New Functions

55. **Add `evaluate_position` function for basic position scoring**
    - File: `game_logic.py`
    - Generate: Function that scores a position based on material, piece positions, king safety
    - Expected ~30-50 lines

56. **Add `find_best_move` function with minimax**
    - File: `game_logic.py`
    - Generate: Basic minimax algorithm with alpha-beta pruning
    - Expected ~40-60 lines

57. **Add `generate_fen` function**
    - File: `board.py`
    - Generate: Function that converts board state to FEN (Forsyth-Edwards Notation) string
    - Expected ~30-40 lines

58. **Add `parse_fen` function**
    - File: `board.py`
    - Generate: Function that creates board state from FEN string
    - Expected ~40-50 lines

### Python - New Classes

59. **Add `MoveHistory` class with undo/redo**
    - File: new `history.py`
    - Generate: Class managing move stack with `push()`, `pop()`, `undo()`, `redo()` methods
    - Expected ~40-50 lines

60. **Add `ChessEngine` class wrapping game logic**
    - File: new `engine.py`
    - Generate: Facade class with `analyze()`, `suggest_move()`, `evaluate()` methods
    - Expected ~60-80 lines

61. **Add `GameSerializer` class for save/load**
    - File: new `serializer.py`
    - Generate: Class with `to_json()`, `from_json()`, `to_pgn()`, `from_pgn()` methods
    - Expected ~50-70 lines

### Go - New Functions

62. **Add `PerftTest` function for move generation testing**
    - File: `moves.go` or new `perft.go`
    - Generate: Performance test function that counts all possible positions at depth N
    - Expected ~30-40 lines

63. **Add `ParseAlgebraic` function**
    - File: `moves.go`
    - Generate: Function that parses algebraic notation ("Nf3", "O-O-O") to Move struct
    - Expected ~40-60 lines

64. **Add `GenerateFEN` function**
    - File: `board.go`
    - Generate: Function that converts BoardState to FEN string
    - Expected ~40-50 lines

### Go - New Types and Methods

65. **Add `TranspositionTable` type with methods**
    - File: new `cache.go`
    - Generate: Hash table for caching position evaluations with `Store()`, `Probe()`, `Clear()`
    - Expected ~50-70 lines

66. **Add `MoveGenerator` struct with iterator pattern**
    - File: `moves.go`
    - Generate: Struct that generates moves lazily with `Next()`, `HasMore()` methods
    - Expected ~40-60 lines

67. **Add comprehensive `String()` methods for all types**
    - Files: `pieces.go`, `board.go`, `moves.go`
    - Generate: Stringer implementations for debugging output
    - Expected ~30-40 lines total

### Multi-File New Code

68. **Add WebSocket support for live game updates**
    - Files: `api.ts` (client), `app.py` (server)
    - Generate: WebSocket connection handling, message types, event handlers
    - Expected ~80-100 lines total

69. **Add unit test suite for move validation**
    - Files: new `game.test.ts`, `test_game_logic.py`, `moves_test.go`
    - Generate: Test cases for each piece type, special moves (castling, en passant)
    - Expected ~100-150 lines per file

70. **Add opening book lookup**
    - Files: new `openings.ts`, `openings.py`, `openings.go`
    - Generate: Functions to look up standard openings and suggest book moves
    - Expected ~40-60 lines per file

71. **Add move animation system**
    - File: new `animation.ts`
    - Generate: Functions for animating piece movement, captures, and promotions
    - Expected ~60-80 lines

---

## Difficulty Ratings

| Difficulty | Description | Example Scenarios |
|------------|-------------|-------------------|
| ⭐ Easy | Single-line change, no cascading | 3, 11, 15, 24, 39 |
| ⭐⭐ Medium | Multi-line or few cascading edits | 6, 10, 20, 26, 32, 36 |
| ⭐⭐⭐ Hard | Multiple files, type system changes | 1, 2, 4, 16, 31, 44 |
| ⭐⭐⭐⭐ Expert | Architectural changes, all languages | 43, 45, 46 |
| 🆕 New Code | Generate substantial code blocks | 47-70 (progressive reveal testing) |

---

## Suggested Test Order

1. Start with single-file, single-line edits (Easy)
2. Progress to multi-line edits in one file (Medium)
3. Test cascading edits across files (Hard)
4. Test cross-language consistency (Expert)
