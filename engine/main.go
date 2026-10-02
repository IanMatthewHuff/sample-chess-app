package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

const maxBodyBytes = 1 << 20

// ValidationRequest represents a move request (validate / move)
type ValidationRequest struct {
	Board     BoardState `json:"board"`
	From      Position   `json:"from"`
	To        Position   `json:"to"`
	Promotion PieceType  `json:"promotion,omitempty"`
	// History holds position keys of all positions so far, including the current one.
	History []string `json:"history,omitempty"`
}

// ValidationResponse represents the response to a validation request
type ValidationResponse struct {
	Valid  bool   `json:"valid"`
	Reason string `json:"reason,omitempty"`
}

// ValidMovesRequest represents a request for valid moves
type ValidMovesRequest struct {
	Board    BoardState `json:"board"`
	Position Position   `json:"position"`
}

// MoveTarget is a legal destination square for a selected piece
type MoveTarget struct {
	Row         int  `json:"row"`
	Col         int  `json:"col"`
	IsCapture   bool `json:"isCapture"`
	IsPromotion bool `json:"isPromotion"`
	IsCastling  bool `json:"isCastling"`
}

// ValidMovesResponse represents the response with valid moves
type ValidMovesResponse struct {
	Moves []MoveTarget `json:"moves"`
}

// StatusRequest represents a game status request
type StatusRequest struct {
	Board   BoardState `json:"board"`
	History []string   `json:"history,omitempty"`
}

// StateResponse describes a position along with derived data
type StateResponse struct {
	State       *BoardState  `json:"state"`
	Status      StatusResult `json:"status"`
	FEN         string       `json:"fen"`
	PositionKey string       `json:"positionKey"`
}

// MoveResponse is returned after a move is applied
type MoveResponse struct {
	StateResponse
	Move Move `json:"move"`
}

// AIRequest asks the engine to choose a move
type AIRequest struct {
	Board      BoardState `json:"board"`
	Difficulty Difficulty `json:"difficulty"`
}

// FENRequest asks the engine to parse a FEN string
type FENRequest struct {
	FEN string `json:"fen"`
}

var (
	rngMu sync.Mutex
	rng   = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// decodePost enforces POST and decodes the JSON body into dst
func decodePost(w http.ResponseWriter, r *http.Request, dst any) bool {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return false
	}
	return true
}

// validateState rejects boards that cannot be analysed safely
func validateState(s *BoardState) error {
	if s.CurrentTurn != White && s.CurrentTurn != Black {
		return errors.New("currentTurn must be 'white' or 'black'")
	}
	kings := map[PieceColor]int{}
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			if p := s.Board[row][col]; p != nil {
				if p.Color != White && p.Color != Black {
					return fmt.Errorf("invalid piece color %q", p.Color)
				}
				if _, ok := centipawns[p.Type]; !ok {
					return fmt.Errorf("invalid piece type %q", p.Type)
				}
				if p.Type == King {
					kings[p.Color]++
				}
			}
		}
	}
	if kings[White] != 1 || kings[Black] != 1 {
		return errors.New("each side must have exactly one king")
	}
	return nil
}

func describe(state *BoardState, history []string) StateResponse {
	return StateResponse{
		State:       state,
		Status:      EvaluateStatus(state, history),
		FEN:         ToFEN(state),
		PositionKey: PositionKey(state),
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "chess-engine"})
}

func handleNew(w http.ResponseWriter, r *http.Request) {
	state := CreateInitialState()
	resp := describe(state, nil)
	writeJSON(w, http.StatusOK, resp)
}

func handleFEN(w http.ResponseWriter, r *http.Request) {
	var req FENRequest
	if !decodePost(w, r, &req) {
		return
	}
	state, err := ParseFEN(req.FEN)
	if err == nil {
		err = validateState(state)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, describe(state, nil))
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	var req ValidationRequest
	if !decodePost(w, r, &req) {
		return
	}
	if err := validateState(&req.Board); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	resp := ValidationResponse{Valid: true}
	if _, err := FindLegalMove(&req.Board, req.From, req.To, req.Promotion); err != nil {
		resp = ValidationResponse{Valid: false, Reason: "Illegal move"}
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleValidMoves(w http.ResponseWriter, r *http.Request) {
	var req ValidMovesRequest
	if !decodePost(w, r, &req) {
		return
	}
	if err := validateState(&req.Board); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	targets := []MoveTarget{}
	seen := map[Position]bool{}
	for _, m := range GetLegalMovesFrom(&req.Board, req.Position) {
		if seen[m.To] {
			continue
		}
		seen[m.To] = true
		targets = append(targets, MoveTarget{
			Row: m.To.Row, Col: m.To.Col,
			IsCapture: m.Captured != nil, IsPromotion: m.IsPromotion, IsCastling: m.IsCastling,
		})
	}
	writeJSON(w, http.StatusOK, ValidMovesResponse{Moves: targets})
}

func handleLegal(w http.ResponseWriter, r *http.Request) {
	var req StatusRequest
	if !decodePost(w, r, &req) {
		return
	}
	if err := validateState(&req.Board); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	moves := GenerateLegalMoves(&req.Board)
	for i := range moves {
		moves[i].SAN = MoveToSAN(&req.Board, moves[i])
	}
	if moves == nil {
		moves = []Move{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"moves": moves})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	var req StatusRequest
	if !decodePost(w, r, &req) {
		return
	}
	if err := validateState(&req.Board); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, describe(&req.Board, req.History))
}

// applyAndDescribe applies a legal move and builds the response
func applyAndDescribe(state *BoardState, m Move, history []string) MoveResponse {
	m.SAN = MoveToSAN(state, m)
	next := ApplyMove(state, m)
	key := PositionKey(next)
	fullHistory := append(append([]string{}, history...), key)
	return MoveResponse{
		StateResponse: StateResponse{
			State:       next,
			Status:      EvaluateStatus(next, fullHistory),
			FEN:         ToFEN(next),
			PositionKey: key,
		},
		Move: m,
	}
}

func handleMakeMove(w http.ResponseWriter, r *http.Request) {
	var req ValidationRequest
	if !decodePost(w, r, &req) {
		return
	}
	if err := validateState(&req.Board); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Promotion != "" && !IsValidPromotion(req.Promotion) {
		writeError(w, http.StatusBadRequest, "promotion must be queen, rook, bishop or knight")
		return
	}
	m, err := FindLegalMove(&req.Board, req.From, req.To, req.Promotion)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Illegal move")
		return
	}
	writeJSON(w, http.StatusOK, applyAndDescribe(&req.Board, m, req.History))
}

func handleAI(w http.ResponseWriter, r *http.Request) {
	var req AIRequest
	if !decodePost(w, r, &req) {
		return
	}
	if err := validateState(&req.Board); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Difficulty == "" {
		req.Difficulty = Medium
	}
	if _, ok := aiConfigs[req.Difficulty]; !ok {
		writeError(w, http.StatusBadRequest, "difficulty must be easy, medium or hard")
		return
	}

	rngMu.Lock()
	seed := rng.Int63()
	rngMu.Unlock()

	result, ok := FindBestMove(&req.Board, req.Difficulty, rand.New(rand.NewSource(seed)))
	if !ok {
		writeError(w, http.StatusConflict, "no legal moves available")
		return
	}
	result.Move.SAN = MoveToSAN(&req.Board, result.Move)
	writeJSON(w, http.StatusOK, result)
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", corsMiddleware(handleHealth))
	mux.HandleFunc("/new", corsMiddleware(handleNew))
	mux.HandleFunc("/fen", corsMiddleware(handleFEN))
	mux.HandleFunc("/validate", corsMiddleware(handleValidate))
	mux.HandleFunc("/moves", corsMiddleware(handleValidMoves))
	mux.HandleFunc("/legal", corsMiddleware(handleLegal))
	mux.HandleFunc("/status", corsMiddleware(handleStatus))
	mux.HandleFunc("/move", corsMiddleware(handleMakeMove))
	mux.HandleFunc("/ai", corsMiddleware(handleAI))
	return mux
}

func main() {
	host := os.Getenv("ENGINE_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("ENGINE_PORT")
	if port == "" {
		port = "18473"
	}
	addr := host + ":" + port

	fmt.Printf("Chess engine listening on http://%s\n", addr)
	server := &http.Server{
		Addr:              addr,
		Handler:           newMux(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
