package main

func (mg *MoveGenerator) bestMove() MoveString {
	board := mg.board
	moves := mg.generateMoves(false)

	var bestMove Move
	bestEval := -20000
	for i := range moves {
		move := &moves[i]
		board.makeMove(move)

		eval := -searchBruteForce(4, -20000, 20000, board)
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

func searchBruteForce(depth int, alpha int, beta int, board Board) int {
	if depth == 0 {
		return basicEval(board)
	}

	moveGenerator := MoveGenerator{board: board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) == 0 {
		if board.playerInCheck() {
			return -20000
		} else {
			return 0
		}
	}

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
