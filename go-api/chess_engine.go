package main

import (
	"cmp"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

const maxSearchPly = 8
const maxQSearchPly = 8

type MoveEvaluation struct {
	evaluation int
	move       *Move
}

// SearchContext holds all per-thread search state. Cloning this struct
// (with a cloned Board) is sufficient to isolate one search worker from another.
type SearchContext struct {
	board     *Board
	searchMG  [maxSearchPly]MoveGenerator
	qsearchMG [maxQSearchPly]MoveGenerator
}

func (ctx *SearchContext) initMoveGeneratorPools() {
	for i := range ctx.searchMG {
		ctx.searchMG[i].board = ctx.board
	}
	for i := range ctx.qsearchMG {
		ctx.qsearchMG[i].board = ctx.board
	}
}

// ChessEngine holds shared state. transpositionTable and searchCancelled are
// pointers so multiple worker goroutines can share them without accidental copies.
type ChessEngine struct {
	ctx                SearchContext
	transpositionTable *TranspositionTable
	searchCancelled    *atomic.Bool
	timer              int
}

func (s *ChessEngine) newWorker() ChessEngine {
	clonedBoard := s.ctx.board.Clone()
	worker := ChessEngine{
		ctx:                SearchContext{board: clonedBoard},
		transpositionTable: s.transpositionTable,
		searchCancelled:    s.searchCancelled,
		timer:              s.timer,
	}
	worker.ctx.initMoveGeneratorPools()
	return worker
}

func (s *ChessEngine) initSearchTranspositionTable() {
	s.transpositionTable = new(TranspositionTable)
	s.searchCancelled = new(atomic.Bool)
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

func (s *ChessEngine) runWorker(startDepth int) {
	board := s.ctx.board
	rootMG := &s.ctx.searchMG[0]
	moves := rootMG.generateMoves(false)
	sortedMoves := make([]MoveEvaluation, len(moves))
	slices.SortFunc(moves, compareMoves)

	for searchDepth := startDepth; searchDepth < 200; searchDepth++ {
		if s.searchCancelled.Load() {
			return
		}
		for i := range moves {
			move := &moves[i]
			board.makeMove(move)
			eval, _ := s.searchBruteForce(searchDepth, 1, -20000, 20000, true)
			eval = -eval
			sortedMoves[i] = MoveEvaluation{evaluation: eval, move: move}
			board.unmakeMove(move)
			if s.searchCancelled.Load() {
				return
			}
		}
		// Re-sort moves based on scores so next iteration starts with better ordering
		slices.SortFunc(sortedMoves, compareEvaluationMoves)
		moves = moves[:0]
		for i := range sortedMoves {
			moves = append(moves, *sortedMoves[i].move)
		}
	}
}

func (s *ChessEngine) bestMove() (MoveString, int, int, int) {
	s.searchCancelled.Store(false)
	s.ctx.initMoveGeneratorPools()

	board := s.ctx.board
	rootMG := &s.ctx.searchMG[0]
	moves := rootMG.generateMoves(false)
	slices.SortFunc(moves, compareMoves)
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

	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	for i := 1; i < numWorkers; i++ {
		wg.Add(1)
		worker := s.newWorker()
		startDepth := (i % 2) + 1 // alternate starting depths: 2, 1, 2, 1, ...
		go func(w ChessEngine, d int) {
			defer wg.Done()
			w.runWorker(d)
		}(worker, startDepth)
	}

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
		slices.SortFunc(sortedMoves, compareEvaluationMoves)
		for i := range sortedMoves {
			moves = append(moves, *sortedMoves[i].move)
		}
	}

	startSquare := toSquare(bestMove.getStartSquare().row, bestMove.getStartSquare().col)
	endSquare := toSquare(bestMove.getEndSquare().row, bestMove.getEndSquare().col)
	promotion := "q"
	if bestMove.getIsPromotion() {
		if bestMove.getPromotionPieceType() == PieceType(Rook) {
			promotion = "r"
		} else if bestMove.getPromotionPieceType() == PieceType(Bishop) {
			promotion = "b"
		} else if bestMove.getPromotionPieceType() == PieceType(Knight) {
			promotion = "n"
		}
	}
	s.searchCancelled.Store(true)
	wg.Wait()
	return MoveString{startSquare: startSquare, endSquare: endSquare, promotion: promotion, isPromotion: bestMove.getIsPromotion()}, totalEvaluated, bestEvalCompleted, completedDepth
}

func (s *ChessEngine) searchBruteForce(depth int, ply int, alpha int, beta int, allowNull bool) (int, int) {

	board := s.ctx.board

	if board.repititionTable.isRepeat(board.zobristHash) {
		return 0, 1
	}

	if ply >= maxSearchPly {
		return pestoEval(board), 1
	}

	originalAlpha := alpha
	tt := s.transpositionTable

	zHash := board.zobristHash
	isValid, ttDepth, flag, evaluation, ttMove := tt.lookup(zHash)
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

	mg := &s.ctx.searchMG[ply]
	mg.board = board
	moves := mg.generateMoves(false)
	if len(moves) == 0 {
		if board.inCheck {
			return -(20000 - ply), 1
		} else {
			return 0, 1
		}
	}

	//null move pruning
	if depth >= 3 && !board.inCheck && !board.isPawnEndgame() {
		nullMove := Move{}
		nullMove.setIsNull(true)
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

	ttMoveFound := false
	if ttMove != 0 {
		for i := range moves {
			if comparePackedMoves(moves[i].getMoveBits(), ttMove) {
				moves[i], moves[0] = moves[0], moves[i]
				ttMoveFound = true
				break
			}
		}
	}
	if ttMoveFound {

		slices.SortFunc(moves[1:], compareMoves)
	} else {
		slices.SortFunc(moves, compareMoves)
	}

	var currentMoveEval int
	var currentPositionsEvaluated int
	var bestMoveInNode Move

	for i := range moves {
		lateMoveReduction := 0
		move := &moves[i]
		board.makeMove(move)
		isLateMoveReduction := i >= 3 && depth >= 3 && !move.getIsPromotion() && (move.getEndSquare().piece == EmptyPiece)
		if isLateMoveReduction {
			lateMoveReduction = 1
		}

		if isLateMoveReduction {
			currentMoveEval, currentPositionsEvaluated = s.searchBruteForce(depth-lateMoveReduction-1, ply+1, -(alpha + 1), -alpha, true)
		} else {
			currentMoveEval, currentPositionsEvaluated = s.searchBruteForce(depth-lateMoveReduction-1, ply+1, -beta, -alpha, true)
		}
		positionsEvaluated += currentPositionsEvaluated
		currentMoveEval = -currentMoveEval

		if currentMoveEval > alpha && isLateMoveReduction {
			currentMoveEval, currentPositionsEvaluated = s.searchBruteForce(depth-1, ply+1, -beta, -alpha, true)
			positionsEvaluated += currentPositionsEvaluated
			currentMoveEval = -currentMoveEval
		}
		if currentMoveEval >= beta {
			storeBeta := beta
			if storeBeta >= 19900 {
				storeBeta += ply
			} else if storeBeta <= -19900 {
				storeBeta -= ply
			}
			tt.push(zHash, depth, 1, storeBeta, packMove(*move))
			board.unmakeMove(move)
			return beta, positionsEvaluated
		}
		if currentMoveEval > alpha {
			alpha = currentMoveEval
			bestMoveInNode = *move
		}
		board.unmakeMove(move)

		if s.searchCancelled.Load() {
			break
		}
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
	tt.push(zHash, depth, flag, storeAlpha, packMove(bestMoveInNode))

	return alpha, positionsEvaluated
}

func (s *ChessEngine) searchOnlyCapturesForce(ply int, qPly int, alpha int, beta int) (int, int) {
	if s.searchCancelled.Load() {
		return 0, 0
	}
	if qPly >= maxQSearchPly {
		return pestoEval(s.ctx.board), 1
	}

	board := s.ctx.board

	standPat := pestoEval(board)

	if standPat >= beta {
		return beta, 1
	}
	if standPat > alpha {
		alpha = standPat
	}

	mg := &s.ctx.qsearchMG[qPly]
	mg.board = board
	moves := mg.generateMoves(true)

	if len(moves) == 0 {
		if board.inCheck && len(mg.generateMoves(false)) == 0 {
			return -(20000 - ply), 1
		} else {
			return standPat, 1
		}
	}
	slices.SortFunc(moves, compareMoves)

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

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
func compareMoves(i, j Move) int {
	if i.getIsPromotion() != j.getIsPromotion() {
		return -cmp.Compare(boolToInt(i.getIsPromotion()), boolToInt(j.getIsPromotion()))
	}

	iIsCapture := i.getEndSquare().piece != 0
	jIsCapture := j.getEndSquare().piece != 0

	if iIsCapture != jIsCapture {
		return -cmp.Compare(boolToInt(iIsCapture), boolToInt(jIsCapture))
	}
	iStartVal := getPieceValue(pieceType(i.getStartSquare().piece))
	iEndVal := getPieceValue(pieceType(i.getEndSquare().piece))
	jStartVal := getPieceValue(pieceType(j.getStartSquare().piece))
	jEndVal := getPieceValue(pieceType(j.getEndSquare().piece))

	iDiff := iEndVal - iStartVal
	jDiff := jEndVal - jStartVal
	if iDiff != jDiff {
		return -cmp.Compare(iDiff, jDiff)
	}

	if i.getStartSquare().row != j.getStartSquare().row {
		return -cmp.Compare(i.getStartSquare().row, j.getStartSquare().row)
	}
	if i.getStartSquare().col != j.getStartSquare().col {
		return -cmp.Compare(i.getStartSquare().col, j.getStartSquare().col)
	}
	if i.getEndSquare().row != j.getEndSquare().row {
		return -cmp.Compare(i.getEndSquare().row, j.getEndSquare().row)
	}
	return -cmp.Compare(i.getEndSquare().col, j.getEndSquare().col)
}

func compareEvaluationMoves(i, j MoveEvaluation) int {
	return -cmp.Compare(i.evaluation, j.evaluation)
}
