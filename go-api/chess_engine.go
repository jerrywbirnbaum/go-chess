package main

import "sort"

func (mg *MoveGenerator) bestMove() (MoveString, int, int) {
	board := *mg.board
	localMoveGenerator := MoveGenerator{board: &board}
	moves := localMoveGenerator.generateMoves(false)
	sort.Sort(MoveOrder(moves))
	tt := initTranspositionTable()

	var bestMove Move
	bestEval := -40000
	totalEvaluated := 0
	for i := range moves {
		move := &moves[i]
		board.makeMove(move)

		eval, positionsEvaluated := searchBruteForce(3, -20000, 20000, &board, &tt)
		totalEvaluated += positionsEvaluated
		eval = -eval
		if eval > bestEval {
			bestMove = *move
			bestEval = eval
		}
		board.unmakeMove(move)
	}

	startSquare := toSquare(bestMove.startSquare.row, bestMove.startSquare.col)
	endSquare := toSquare(bestMove.endSquare.row, bestMove.endSquare.col)
	return MoveString{startSquare: startSquare, endSquare: endSquare}, totalEvaluated, bestEval
}

func searchBruteForce(depth int, alpha int, beta int, board *Board, tt *TranspositionTable) (int, int) {
	originalAlpha := alpha

	zHash := board.calculateZobrishHash()
	isValid, ttDepth, flag, evaluation := tt.lookup(zHash)
	if isValid && ttDepth >= depth {
		if flag == 0 {
			return evaluation, 1
		} else if flag == 1 && evaluation >= beta {
			return evaluation, 1
		} else if flag == 2 && evaluation <= alpha {
			return evaluation, 1
		}
	}

	if depth == 0 {
		return searchOnlyCapturesForce(alpha, beta, board), 1
	}
	positionsEvaluated := 0

	moveGenerator := MoveGenerator{board: board}
	moves := moveGenerator.generateMoves(false)
	sort.Sort(MoveOrder(moves))
	if len(moves) == 0 {
		if board.playerInCheck() {
			return -20000, 1
		} else {
			return 0, 1
		}
	}

	var currentMoveEval int
	var currentPositionsEvaluated int
	for i := range moves {
		move := &moves[i]
		board.makeMove(move)
		currentMoveEval, currentPositionsEvaluated = searchBruteForce(depth-1, -beta, -alpha, board, tt)
		positionsEvaluated += currentPositionsEvaluated
		currentMoveEval = -currentMoveEval
		if currentMoveEval >= beta {

			tt.push(zHash, depth, 1, beta)
			board.unmakeMove(move)
			return beta, positionsEvaluated
		}
		alpha = max(alpha, currentMoveEval)
		board.unmakeMove(move)
	}

	if alpha <= originalAlpha {
		flag = 2
	} else {
		flag = 0
	}
	tt.push(zHash, depth, flag, alpha)

	return alpha, positionsEvaluated
}

func searchOnlyCapturesForce(alpha int, beta int, board *Board) int {
	standPat := basicEval(*board)
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
