package main

import "math"

func (mg *MoveGenerator) bestMove() MoveString {
	board := mg.board
	moves := mg.generateMoves()

	var bestMove Move
	bestEval := math.Inf(-1)
	for i := range moves {
		move := &moves[i]
		board.makeMove(move)

		eval := -searchBruteForce(5, math.Inf(-1), math.Inf(1), board)
		if eval > bestEval {
			bestMove = *move
			bestEval = eval
		}
		board.unmakeMove(move)
	}

	startSquare := toSquare(bestMove.startSquare.row, bestMove.startSquare.col)
	endSquare := toSquare(bestMove.endSquare.row, bestMove.endSquare.col)
	return MoveString{startSquare: startSquare, endSquare: endSquare}
}

func searchBruteForce(depth int, alpha float64, beta float64, board Board) float64 {
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

	// bestMoveEval := math.Inf(-1)
	for i := range moves {
		move := &moves[i]
		board.makeMove(move)
		currentMoveEval := -searchBruteForce(depth-1, -beta, -alpha, board)
		alpha = max(alpha, currentMoveEval)
		if currentMoveEval >= beta {
			return beta
		}
		board.unmakeMove(move)
	}
	return alpha
}
