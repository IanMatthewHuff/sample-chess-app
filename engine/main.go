package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// ValidationRequest represents a move validation request
type ValidationRequest struct {
	Board BoardState `json:"board"`
	From  Position   `json:"from"`
	To    Position   `json:"to"`
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

// ValidMovesResponse represents the response with valid moves
type ValidMovesResponse struct {
	Moves []Position `json:"moves"`
}

// StatusRequest represents a game status request
type StatusRequest struct {
	Board BoardState `json:"board"`
}

// StatusResponse represents the game status response
type StatusResponse struct {
	Status string `json:"status"`
	InCheck bool   `json:"inCheck"`
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "chess-engine",
	})
}

// handleValidate handles move validation requests
func handleValidate(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement validation handler
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	valid := IsValidMove(&req.Board, req.From, req.To)

	response := ValidationResponse{
		Valid: valid,
	}
	if !valid {
		response.Reason = "Invalid move"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleValidMoves handles requests for valid moves
func handleValidMoves(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement valid moves handler
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidMovesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	moves := GetValidMoves(&req.Board, req.Position)

	response := ValidMovesResponse{
		Moves: moves,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStatus handles game status requests
func handleStatus(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement status handler
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	status := GetGameStatus(&req.Board)
	inCheck := IsInCheck(&req.Board, req.Board.CurrentTurn)

	response := StatusResponse{
		Status:  string(status),
		InCheck: inCheck,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMakeMove handles move execution requests
func handleMakeMove(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement make move handler
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !IsValidMove(&req.Board, req.From, req.To) {
		http.Error(w, "Invalid move", http.StatusBadRequest)
		return
	}

	newState := MakeMove(&req.Board, req.From, req.To)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newState)
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

func main() {
	fmt.Println("Starting Chess Engine...")
	fmt.Println("Engine running at http://localhost:8080")

	// Register routes
	http.HandleFunc("/health", corsMiddleware(handleHealth))
	http.HandleFunc("/validate", corsMiddleware(handleValidate))
	http.HandleFunc("/moves", corsMiddleware(handleValidMoves))
	http.HandleFunc("/status", corsMiddleware(handleStatus))
	http.HandleFunc("/move", corsMiddleware(handleMakeMove))

	// Start server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
