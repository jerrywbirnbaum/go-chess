package main

import (
	"sort"
	"sync/atomic"
	"time"
)

const maxSearchPly = 16
const maxQSearchPly = 16

type MoveEvaluation struct {
	evaluation int
	move       *Move
}

type ChessEngine struct {
	moveGenerator      MoveGenerator
	transpositionTable TranspositionTable
	searchCancelled    atomic.Bool
	timer              int
	searchMG           [maxSearchPly]MoveGenerator
	qsearchMG          [maxQSearchPly]MoveGenerator
}

func (s *ChessEngine) initMoveGeneratorPools() {
	board := s.moveGenerator.board
	for i := range s.searchMG {
		s.searchMG[i].board = board
	}
	for i := range s.qsearchMG {
		s.qsearchMG[i].board = board
	}
}

func (s *ChessEngine) initSearchTranspositionTable() {
	s.transpositionTable = initTranspositionTable()
}

func (s *ChessEngine) startSearchTimer(done <-chan struct{}) {
	timer := s.timer
	if timer == 0 {
		timer = 1000
	}
	select {
	case <-time.After(time.Duration(timer) * time.Millisecond):
		s.searchCancelled.Store(true)
	case <-done:
	}
}

func (s *ChessEngine) setTimer(timer int) {
	s.timer = timer
}

func (s *ChessEngine) bestMove() (MoveString, int, int, int) {
	s.searchCancelled.Store(false)
	s.initMoveGeneratorPools()

	board := s.moveGenerator.board
	rootMG := &s.searchMG[0]
	moves := rootMG.generateMoves(false)
	sort.Sort(MoveOrder(moves))
	totalEvaluated := 0
	sortedMoves := make([]MoveEvaluation, len(moves))

	var bestMove Move
	var bestMoveCurrentIteration Move
	bestEvalCompleted := -40000
	completedDepth := 1

	done := make(chan struct{})
	defer close(done)
	go s.startSearchTimer(done)

	timer := s.timer
	if timer == 0 {
		timer = 1000
	}
	softLimit := time.Duration(timer) * time.Millisecond * 6 / 10
	searchStart := time.Now()

	for searchDepth := range 200 {
		if searchDepth == 0 {
			continue
		}

		if time.Since(searchStart) > softLimit {
			break
		}

		bestEvalCurrentIteration := -40000
		bestMoveCurrentIteration = bestMove // inherit previous best as fallback
		movesEvaluatedInIteration := 0

		for i := range moves {
			move := &moves[i]
			board.makeMove(move)
			eval, positionsEvaluated := s.searchBruteForce(searchDepth, 1, -20000, 20000, true)
			totalEvaluated += positionsEvaluated
			eval = -eval
			board.unmakeMove(move)
			if s.searchCancelled.Load() {
				break
			}

			movesEvaluatedInIteration++
			sortedMoves[i] = MoveEvaluation{evaluation: eval, move: move}
			if eval > bestEvalCurrentIteration {
				bestMoveCurrentIteration = *move
				bestEvalCurrentIteration = eval
			}
		}

		if s.searchCancelled.Load() {
			if movesEvaluatedInIteration > 0 {
				bestMove = bestMoveCurrentIteration
				bestEvalCompleted = bestEvalCurrentIteration
			}
			break
		}

		bestMove = bestMoveCurrentIteration
		bestEvalCompleted = bestEvalCurrentIteration
		completedDepth = searchDepth

		moves = nil
		sort.Sort(MoveEvaluationOrder(sortedMoves))
		for i := range sortedMoves {
			moves = append(moves, *sortedMoves[i].move)
		}
	}

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
	return MoveString{startSquare: startSquare, endSquare: endSquare, promotion: promotion, isPromotion: bestMove.isPromotion}, totalEvaluated, bestEvalCompleted, completedDepth
}

func (s *ChessEngine) searchBruteForce(depth int, ply int, alpha int, beta int, allowNull bool) (int, int) {
	if s.searchCancelled.Load() {
		return 0, 0
	}

	if ply >= maxSearchPly {
		return pestoEval(*s.moveGenerator.board), 1
	}

	originalAlpha := alpha
	board := s.moveGenerator.board
	tt := s.transpositionTable
	inCheck := board.playerInCheck()

	zHash := board.zobristHash
	isValid, ttDepth, flag, evaluation := tt.lookup(zHash)
	if isValid && ttDepth >= depth {
		if evaluation >= 19900 {
			evaluation -= ply
		} else if evaluation <= -19900 {
			evaluation += ply
		}
		if flag == 0 {
			return evaluation, 1
		} else if flag == 1 && evaluation >= beta {
			return evaluation, 1
		} else if flag == 2 && evaluation <= alpha {
			return evaluation, 1
		}
	}

	if depth <= 0 {
		return s.searchOnlyCapturesForce(ply, 0, alpha, beta)
	}

	positionsEvaluated := 0

	//null move pruning
	if depth >= 3 && !inCheck && !board.isPawnEndgame() {
		nullMove := Move{isNull: true}
		board.makeMove(&nullMove)
		nullMoveReduction := 2
		nullMoveEval, nullNodes := s.searchBruteForce(depth-nullMoveReduction, ply+1, -beta, -beta+1, false)
		nullMoveEval = -nullMoveEval
		board.unmakeMove(&nullMove)
		positionsEvaluated += nullNodes
		if nullMoveEval >= beta {
			return beta, nullNodes
		}
	}

	mg := &s.searchMG[ply]
	mg.board = board
	moves := mg.generateMoves(false)
	if len(moves) == 0 {
		if inCheck {
			return -(20000 - ply), 1
		} else {
			return 0, 1
		}
	}
	sort.Sort(MoveOrder(moves))

	var currentMoveEval int
	var currentPositionsEvaluated int

	for i := range moves {
		move := &moves[i]
		board.makeMove(move)
		currentMoveEval, currentPositionsEvaluated = s.searchBruteForce(depth-1, ply+1, -beta, -alpha, true)
		positionsEvaluated += currentPositionsEvaluated
		currentMoveEval = -currentMoveEval
		if currentMoveEval >= beta {
			storeBeta := beta
			if storeBeta >= 19900 {
				storeBeta += ply
			} else if storeBeta <= -19900 {
				storeBeta -= ply
			}
			tt.push(zHash, depth, 1, storeBeta)
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

	storeAlpha := alpha
	if storeAlpha >= 19900 {
		storeAlpha += ply
	} else if storeAlpha <= -19900 {
		storeAlpha -= ply
	}
	tt.push(zHash, depth, flag, storeAlpha)

	return alpha, positionsEvaluated
}

func (s *ChessEngine) searchOnlyCapturesForce(ply int, qPly int, alpha int, beta int) (int, int) {
	if qPly >= maxQSearchPly {
		return pestoEval(*s.moveGenerator.board), 1
	}

	board := s.moveGenerator.board

	standPat := pestoEval(*board)

	if standPat >= beta {
		return beta, 1
	}
	if standPat > alpha {
		alpha = standPat
	}

	mg := &s.qsearchMG[qPly]
	mg.board = board
	moves := mg.generateMoves(true)

	if len(moves) == 0 {
		if board.playerInCheck() && len(mg.generateMoves(false)) == 0 {
			return -(20000 - ply), 1
		} else {
			return standPat, 1
		}
	}
	sort.Sort(MoveOrder(moves))

	var currentMoveEval int
	var currentNodes int
	for i := range moves {
		move := &moves[i]
		board.makeMove(move)
		currentMoveEval, currentNodes = s.searchOnlyCapturesForce(ply+1, qPly+1, -beta, -alpha)
		currentMoveEval = -currentMoveEval
		board.unmakeMove(move)

		if currentMoveEval >= beta {
			return beta, currentNodes
		}
		alpha = max(alpha, currentMoveEval)
	}
	return alpha, currentNodes
}

type MoveOrder []Move

func (a MoveOrder) Len() int      { return len(a) }
func (a MoveOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a MoveOrder) Less(i, j int) bool {

	if a[i].isPromotion != a[j].isPromotion {
		return a[i].isPromotion
	}

	iIsCapture := a[i].endSquare.piece != 0
	jIsCapture := a[j].endSquare.piece != 0

	if iIsCapture != jIsCapture {
		return iIsCapture
	}
	iStartVal := getPieceValue(pieceType(a[i].startSquare.piece))
	iEndVal := getPieceValue(pieceType(a[i].endSquare.piece))
	jStartVal := getPieceValue(pieceType(a[j].startSquare.piece))
	jEndVal := getPieceValue(pieceType(a[j].endSquare.piece))

	iDiff := iEndVal - iStartVal
	jDiff := jEndVal - jStartVal
	if iDiff != jDiff {
		return iDiff > jDiff
	}

	if a[i].startSquare.row != a[j].startSquare.row {
		return a[i].startSquare.row < a[j].startSquare.row
	}
	if a[i].startSquare.col != a[j].startSquare.col {
		return a[i].startSquare.col < a[j].startSquare.col
	}
	if a[i].endSquare.row != a[j].endSquare.row {
		return a[i].endSquare.row < a[j].endSquare.row
	}
	return a[i].endSquare.col < a[j].endSquare.col
}

type MoveEvaluationOrder []MoveEvaluation

func (a MoveEvaluationOrder) Len() int      { return len(a) }
func (a MoveEvaluationOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a MoveEvaluationOrder) Less(i, j int) bool {
	return a[i].evaluation > a[j].evaluation
}
