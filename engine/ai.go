package main

import (
	"math/rand"
	"sort"
	"time"
)

// Difficulty controls AI search depth and randomness
type Difficulty string

const (
	Easy   Difficulty = "easy"
	Medium Difficulty = "medium"
	Hard   Difficulty = "hard"
)

type aiConfig struct {
	maxDepth  int
	timeLimit time.Duration
	// Root moves scoring within tolerance (centipawns) of the best are picked at random.
	tolerance int
}

var aiConfigs = map[Difficulty]aiConfig{
	Easy:   {maxDepth: 1, timeLimit: 500 * time.Millisecond, tolerance: 120},
	Medium: {maxDepth: 2, timeLimit: 1500 * time.Millisecond, tolerance: 25},
	Hard:   {maxDepth: 4, timeLimit: 3 * time.Second, tolerance: 0},
}

const (
	mateScore      = 100000
	infinity       = 1 << 30
	quiescenceMax  = 6
	nodeCheckEvery = 1024
)

var centipawns = map[PieceType]int{Pawn: 100, Knight: 320, Bishop: 330, Rook: 500, Queen: 900, King: 0}

// Piece-square tables from White's point of view; row 0 is rank 8.
var pieceSquareTables = map[PieceType][8][8]int{
	Pawn: {
		{0, 0, 0, 0, 0, 0, 0, 0},
		{50, 50, 50, 50, 50, 50, 50, 50},
		{10, 10, 20, 30, 30, 20, 10, 10},
		{5, 5, 10, 25, 25, 10, 5, 5},
		{0, 0, 0, 20, 20, 0, 0, 0},
		{5, -5, -10, 0, 0, -10, -5, 5},
		{5, 10, 10, -20, -20, 10, 10, 5},
		{0, 0, 0, 0, 0, 0, 0, 0},
	},
	Knight: {
		{-50, -40, -30, -30, -30, -30, -40, -50},
		{-40, -20, 0, 0, 0, 0, -20, -40},
		{-30, 0, 10, 15, 15, 10, 0, -30},
		{-30, 5, 15, 20, 20, 15, 5, -30},
		{-30, 0, 15, 20, 20, 15, 0, -30},
		{-30, 5, 10, 15, 15, 10, 5, -30},
		{-40, -20, 0, 5, 5, 0, -20, -40},
		{-50, -40, -30, -30, -30, -30, -40, -50},
	},
	Bishop: {
		{-20, -10, -10, -10, -10, -10, -10, -20},
		{-10, 0, 0, 0, 0, 0, 0, -10},
		{-10, 0, 5, 10, 10, 5, 0, -10},
		{-10, 5, 5, 10, 10, 5, 5, -10},
		{-10, 0, 10, 10, 10, 10, 0, -10},
		{-10, 10, 10, 10, 10, 10, 10, -10},
		{-10, 5, 0, 0, 0, 0, 5, -10},
		{-20, -10, -10, -10, -10, -10, -10, -20},
	},
	Rook: {
		{0, 0, 0, 0, 0, 0, 0, 0},
		{5, 10, 10, 10, 10, 10, 10, 5},
		{-5, 0, 0, 0, 0, 0, 0, -5},
		{-5, 0, 0, 0, 0, 0, 0, -5},
		{-5, 0, 0, 0, 0, 0, 0, -5},
		{-5, 0, 0, 0, 0, 0, 0, -5},
		{-5, 0, 0, 0, 0, 0, 0, -5},
		{0, 0, 0, 5, 5, 0, 0, 0},
	},
	Queen: {
		{-20, -10, -10, -5, -5, -10, -10, -20},
		{-10, 0, 0, 0, 0, 0, 0, -10},
		{-10, 0, 5, 5, 5, 5, 0, -10},
		{-5, 0, 5, 5, 5, 5, 0, -5},
		{0, 0, 5, 5, 5, 5, 0, -5},
		{-10, 5, 5, 5, 5, 5, 0, -10},
		{-10, 0, 5, 0, 0, 0, 0, -10},
		{-20, -10, -10, -5, -5, -10, -10, -20},
	},
	King: {
		{-30, -40, -40, -50, -50, -40, -40, -30},
		{-30, -40, -40, -50, -50, -40, -40, -30},
		{-30, -40, -40, -50, -50, -40, -40, -30},
		{-30, -40, -40, -50, -50, -40, -40, -30},
		{-20, -30, -30, -40, -40, -30, -30, -20},
		{-10, -20, -20, -20, -20, -20, -20, -10},
		{20, 20, 0, 0, 0, 0, 20, 20},
		{20, 30, 10, 0, 0, 10, 30, 20},
	},
}

// Evaluate scores the position in centipawns from the side to move's perspective
func Evaluate(state *BoardState) int {
	score := 0
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			p := state.Board[row][col]
			if p == nil {
				continue
			}
			tableRow := row
			if p.Color == Black {
				tableRow = 7 - row
			}
			v := centipawns[p.Type] + pieceSquareTables[p.Type][tableRow][col]
			if p.Color == White {
				score += v
			} else {
				score -= v
			}
		}
	}
	if state.CurrentTurn == Black {
		return -score
	}
	return score
}

// SearchResult is the outcome of an AI search
type SearchResult struct {
	Move  Move `json:"move"`
	Score int  `json:"score"`
	Depth int  `json:"depth"`
	Nodes int  `json:"nodes"`
}

type searcher struct {
	nodes    int
	deadline time.Time
	stopped  bool
}

func moveOrderScore(m Move) int {
	s := 0
	if m.Captured != nil {
		s += 10*centipawns[m.Captured.Type] - centipawns[m.Piece.Type]/10 + 1000
	}
	if m.IsPromotion {
		s += centipawns[m.PromotedTo] + 800
	}
	return s
}

func orderMoves(moves []Move) {
	sort.SliceStable(moves, func(i, j int) bool { return moveOrderScore(moves[i]) > moveOrderScore(moves[j]) })
}

func (s *searcher) timeUp() bool {
	if s.stopped {
		return true
	}
	if s.nodes%nodeCheckEvery == 0 && !s.deadline.IsZero() && time.Now().After(s.deadline) {
		s.stopped = true
	}
	return s.stopped
}

func (s *searcher) negamax(state *BoardState, depth, alpha, beta, ply int) int {
	s.nodes++
	if s.timeUp() {
		return 0
	}
	moves := GenerateLegalMoves(state)
	if len(moves) == 0 {
		if IsInCheck(state, state.CurrentTurn) {
			return -mateScore + ply
		}
		return 0
	}
	if state.HalfMoveClock >= 100 || IsInsufficientMaterial(&state.Board) {
		return 0
	}
	if depth <= 0 {
		return s.quiescence(state, alpha, beta, 0)
	}

	orderMoves(moves)
	for _, m := range moves {
		score := -s.negamax(ApplyMove(state, m), depth-1, -beta, -alpha, ply+1)
		if s.stopped {
			return 0
		}
		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}
	return alpha
}

func (s *searcher) quiescence(state *BoardState, alpha, beta, qdepth int) int {
	s.nodes++
	standPat := Evaluate(state)
	if standPat >= beta {
		return beta
	}
	if standPat > alpha {
		alpha = standPat
	}
	if qdepth >= quiescenceMax || s.timeUp() {
		return alpha
	}

	var tactical []Move
	for _, m := range GenerateLegalMoves(state) {
		if m.Captured != nil || m.IsPromotion {
			tactical = append(tactical, m)
		}
	}
	orderMoves(tactical)
	for _, m := range tactical {
		score := -s.quiescence(ApplyMove(state, m), -beta, -alpha, qdepth+1)
		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}
	return alpha
}

// FindBestMove searches for a move at the given difficulty. ok is false when
// the side to move has no legal moves.
func FindBestMove(state *BoardState, difficulty Difficulty, rng *rand.Rand) (SearchResult, bool) {
	cfg, found := aiConfigs[difficulty]
	if !found {
		cfg = aiConfigs[Medium]
	}
	rootMoves := GenerateLegalMoves(state)
	if len(rootMoves) == 0 {
		return SearchResult{}, false
	}
	orderMoves(rootMoves)

	s := &searcher{deadline: time.Now().Add(cfg.timeLimit)}
	type scored struct {
		move  Move
		score int
	}
	var best []scored
	completedDepth := 0

	for depth := 1; depth <= cfg.maxDepth; depth++ {
		// Depth 1 always completes so there is always a move to play.
		if depth == 1 {
			s.deadline = time.Time{}
		} else {
			s.deadline = time.Now().Add(cfg.timeLimit)
		}
		results := make([]scored, 0, len(rootMoves))
		bestScore := -infinity
		for _, m := range rootMoves {
			alpha := -infinity
			if bestScore > -infinity {
				alpha = bestScore - cfg.tolerance - 1
			}
			score := -s.negamax(ApplyMove(state, m), depth-1, -infinity, -alpha, 1)
			if s.stopped {
				break
			}
			results = append(results, scored{m, score})
			if score > bestScore {
				bestScore = score
			}
		}
		if s.stopped {
			break
		}
		sort.SliceStable(results, func(i, j int) bool { return results[i].score > results[j].score })
		best = results
		completedDepth = depth

		// Search the principal move first on the next iteration.
		for i := range results {
			rootMoves[i] = results[i].move
		}
		if best[0].score >= mateScore-100 {
			break
		}
	}

	candidates := []scored{best[0]}
	for _, r := range best[1:] {
		if r.score >= best[0].score-cfg.tolerance {
			candidates = append(candidates, r)
		}
	}
	pick := candidates[0]
	if len(candidates) > 1 && rng != nil {
		pick = candidates[rng.Intn(len(candidates))]
	}
	return SearchResult{Move: pick.move, Score: pick.score, Depth: completedDepth, Nodes: s.nodes}, true
}
