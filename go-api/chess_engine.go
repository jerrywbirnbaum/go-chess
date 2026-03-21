package main

import (
	"sort"
	"time"
)

type MoveEvaluation struct {
	evaluation int
	move       *Move
}
type ChessEngine struct {
	moveGenerator      MoveGenerator
	transpositionTable TranspositionTable
	searchCancelled    bool
	timer              int
}

func (s *ChessEngine) initSearchTranspositionTable() {
	s.transpositionTable = initTranspositionTable()
}

func (s *ChessEngine) startSearchTimer() {
	if s.timer == 0 {
		s.timer = 1000
	}
	time.Sleep(time.Duration(s.timer) * time.Millisecond)
	s.searchCancelled = true
}

func (s *ChessEngine) setTimer(timer int) {
	s.timer = timer
}

func (s *ChessEngine) bestMove() (MoveString, int, int) {
	s.searchCancelled = false
	board := s.moveGenerator.board
	localMoveGenerator := MoveGenerator{board: board}
	moves := localMoveGenerator.generateMoves(false)
	sort.Sort(MoveOrder(moves))
	bestEval := -40000
	totalEvaluated := 0
	sortedMoves := make([]MoveEvaluation, len(moves))

	var bestMove Move
	var bestMoveCurrentIteration Move
	go s.startSearchTimer()
	for searchDepth := range 200 {
		if searchDepth == 0 {
			continue
		}
		if s.searchCancelled {
			break
		}

		for i := range moves {
			move := &moves[i]
			board.makeMove(move)

			eval, positionsEvaluated := s.searchBruteForce(searchDepth, -20000, 20000)
			totalEvaluated += positionsEvaluated
			eval = -eval
			sortedMoves[i] = MoveEvaluation{evaluation: eval, move: move}
			if eval > bestEval {
				bestMoveCurrentIteration = *move
				bestEval = eval
			}
			board.unmakeMove(move)
		}

		//Sort moves for one iteration deeper in the order of the previous iteration
		moves = nil
		sort.Sort(MoveEvaluationOrder(sortedMoves))
		for i := range sortedMoves {
			moves = append(moves, *sortedMoves[i].move)
		}
	}
	bestMove = bestMoveCurrentIteration

	startSquare := toSquare(bestMove.startSquare.row, bestMove.startSquare.col)
	endSquare := toSquare(bestMove.endSquare.row, bestMove.endSquare.col)
	promotion := "q"
	if bestMove.isPromotion {
		if bestMove.promotionPieceType == PieceType(Rook) {
			promotion = "r"
		} else if bestMove.promotionPieceType == PieceType(Bishop) {
			promotion = "b"
		} else if bestMove.promotionPieceType == PieceType(Knight) {
			promotion = "n"
		}

	}
	return MoveString{startSquare: startSquare, endSquare: endSquare, promotion: promotion, isPromotion: bestMove.isPromotion}, totalEvaluated, bestEval
}

func (s *ChessEngine) searchBruteForce(depth int, alpha int, beta int) (int, int) {
	originalAlpha := alpha
	board := s.moveGenerator.board
	tt := s.transpositionTable

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

	if depth <= 0 || s.searchCancelled {
		return s.searchOnlyCapturesForce(alpha, beta), 1
	}
	positionsEvaluated := 0

	moveGenerator := MoveGenerator{board: board}
	moves := moveGenerator.generateMoves(false)
	sort.Sort(MoveOrder(moves))
	if len(moves) == 0 {
		if board.playerInCheck() {
			return -20000 - depth, 1
		} else {
			return 0, 1
		}
	}

	var currentMoveEval int
	var currentPositionsEvaluated int

	for i := range moves {
		move := &moves[i]
		board.makeMove(move)
		currentMoveEval, currentPositionsEvaluated = s.searchBruteForce(depth-1, -beta, -alpha)
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

func (s *ChessEngine) searchOnlyCapturesForce(alpha int, beta int) int {
	board := s.moveGenerator.board
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
		if board.playerInCheck() {
			return -20000
		} else {
			return standPat
		}
	}
	sort.Sort(MoveOrder(moves))

	for i := range moves {
		move := &moves[i]
		board.makeMove(move)
		currentMoveEval := -s.searchOnlyCapturesForce(-beta, -alpha)
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

type MoveEvaluationOrder []MoveEvaluation

func (a MoveEvaluationOrder) Len() int      { return len(a) }
func (a MoveEvaluationOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a MoveEvaluationOrder) Less(i, j int) bool {
	return a[i].evaluation > a[j].evaluation
}
