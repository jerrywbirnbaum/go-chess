package main

import "sort"

func (mg *MoveGenerator) bestMove() MoveString {
	board := mg.board
	moves := mg.generateMoves(false)
	sort.Sort(MoveOrder(moves))

	var bestMove Move
	bestEval := -20000
	for i := range moves {
		move := &moves[i]
		board.makeMove(move)

		eval := -searchBruteForce(3, -20000, 20000, board)
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
		return searchOnlyCapturesForce(alpha, beta, board)
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

func searchOnlyCapturesForce(alpha int, beta int, board Board) int {
	standPat := basicEval(board)
	if standPat >= beta {
		return beta
	}
	if standPat > alpha {
		alpha = standPat
	}

	moveGenerator := MoveGenerator{board: board}
	moves := moveGenerator.generateMoves(true)
	if len(moves) == 0 {
		return standPat
	}
	sort.Sort(MoveOrder(moves))

	for i := range moves {
		move := &moves[i]
		board.makeMove(move)
		currentMoveEval := -searchOnlyCapturesForce(-beta, -alpha, board)
		board.unmakeMove(move)

		if currentMoveEval >= beta {
			return beta
		}
		alpha = max(alpha, currentMoveEval)
	}
	return alpha
}

type MoveOrder []Move

func (a MoveOrder) Len() int      { return len(a) }
func (a MoveOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a MoveOrder) Less(i, j int) bool {
	startPieceType := pieceType(a[i].startSquare.piece)
	endPieceType := pieceType(a[i].endSquare.piece)
	iPieceDiff := getPieceValue(endPieceType) - getPieceValue(startPieceType)
	if endPieceType == 0 {
		iPieceDiff = 0
	}

	startPieceType = pieceType(a[j].startSquare.piece)
	endPieceType = pieceType(a[j].endSquare.piece)
	jPieceDiff := getPieceValue(endPieceType) - getPieceValue(startPieceType)
	if endPieceType == 0 {
		jPieceDiff = 0
	}

	return iPieceDiff > jPieceDiff
}
