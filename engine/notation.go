package main

func MoveSAN(state *BoardState, from, to Position, promotion PieceType) string {
	piece := state.Board.GetPieceAt(from)
	var notation string
	if piece.Type == King && abs(from.Col-to.Col) == 2 {
		notation = "O-O"
		if to.Col == 2 {
			notation = "O-O-O"
		}
	} else {
		letters := map[PieceType]string{
			King: "K", Queen: "Q", Rook: "R", Bishop: "B", Knight: "N",
		}
		notation = letters[piece.Type]
		fromSquare := from.ToAlgebraic()
		capture := IsCaptureMove(state, from, to)
		if piece.Type == Pawn && capture {
			notation += fromSquare[:1]
		} else if piece.Type != Pawn {
			sameFile, sameRank, ambiguous := false, false, false
			for _, other := range state.Board.GetPiecesByColor(piece.Color) {
				if other.Position == from || other.Piece.Type != piece.Type {
					continue
				}
				if IsValidMove(state, other.Position, to) {
					ambiguous = true
					sameFile = sameFile || other.Position.Col == from.Col
					sameRank = sameRank || other.Position.Row == from.Row
				}
			}
			if ambiguous {
				if !sameFile {
					notation += fromSquare[:1]
				} else if !sameRank {
					notation += fromSquare[1:]
				} else {
					notation += fromSquare
				}
			}
		}
		if capture {
			notation += "x"
		}
		notation += to.ToAlgebraic()
		if piece.Type == Pawn && (to.Row == 0 || to.Row == 7) {
			notation += "=" + letters[promotion]
		}
	}
	next := MakeMove(state, from, to, promotion)
	if GetGameStatus(next) == Checkmate {
		notation += "#"
	} else if IsInCheck(next, next.CurrentTurn) {
		notation += "+"
	}
	return notation
}
