package main

import "math"

func (mg *MoveGenerator) bestMove() MoveString {
	board := mg.board
	moves := mg.generateMoves()

	var bestMove Move
	bestEval := math.Inf(-1)
	for _, move := range moves {
		castle := board.castleAvailable
		enpassant := board.enpassant
		board.makeMove(move)

		eval := -searchBruteForce(3, board)
		if eval > bestEval {
			bestMove = move
			bestEval = eval
		}
		board.unmakeMove(move)
		board.castleAvailable = castle
		board.enpassant = enpassant
	}

	startSquare := toSquare(bestMove.startSquare.row, bestMove.startSquare.col)
	endSquare := toSquare(bestMove.endSquare.row, bestMove.endSquare.col)
	return MoveString{startSquare: startSquare, endSquare: endSquare}
}

func searchBruteForce(depth int, board Board) float64 {
	if depth == 0 {
		return basicEval(board)
	}

	moveGenerator := MoveGenerator{board: board}
	moves := moveGenerator.generateMoves()
	if len(moves) == 0 {
		if board.playerInCheck() {
			return math.Inf(-1)
		} else {
			return 0
		}
	}

	bestMoveEval := math.Inf(-1)
	for _, move := range moves {
		castle := board.castleAvailable
		enpassant := board.enpassant
		board.makeMove(move)
		currentMoveEval := -searchBruteForce(depth-1, board)
		bestMoveEval = max(bestMoveEval, currentMoveEval)
		board.unmakeMove(move)
		board.castleAvailable = castle
		board.enpassant = enpassant
	}
	return bestMoveEval
}
