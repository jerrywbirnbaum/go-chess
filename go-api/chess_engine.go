package main

import (
	"cmp"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

const maxSearchPly = 64
const maxQSearchPly = 16
const maxHistoryScore = 16000
const historyAgeShift = 4

type MoveEvaluation struct {
	evaluation int
	move       *Move
}

// SearchContext holds all per-thread search state.
type SearchContext struct {
	board     *Board
	searchMG  [maxSearchPly]MoveGenerator
	qsearchMG [maxQSearchPly]MoveGenerator
	killers   [maxSearchPly][2]Move
	history   [2][64][64]int
}

type ScoredMove struct {
	move  Move
	score int
}

func (ctx *SearchContext) initMoveGeneratorPools() {
	for i := range ctx.searchMG {
		ctx.searchMG[i].board = ctx.board
	}
	for i := range ctx.qsearchMG {
		ctx.qsearchMG[i].board = ctx.board
	}
}

func (ctx *SearchContext) clearKillers() {
	ctx.killers = [maxSearchPly][2]Move{}
}

func (ctx *SearchContext) clearHistory() {
	ctx.history = [2][64][64]int{}
}

func (ctx *SearchContext) ageHistory() {
	for side := range 2 {
		for from := range 64 {
			for to := range 64 {
				ctx.history[side][from][to] >>= historyAgeShift
			}
		}
	}
}

// Score tiers: TT move (2M) > promotion (1M) > capture/MVV-LVA (100K+) >
// killer 0 (90K) > killer 1 (80K) > history (0–16K).
func (ctx *SearchContext) scoreMoves(moves []Move, ply int, ttMove uint64, side int) []ScoredMove {
	scored := make([]ScoredMove, len(moves))
	for i, move := range moves {
		var score int
		moveBits := move.getMoveBits()
		switch {
		case ttMove != 0 && comparePackedMoves(moveBits, ttMove):
			score = 2_000_000
		case move.getIsPromotion():
			score = 1_000_000
		case move.getEndSquare().piece != EmptyPiece || move.getIsEnpassant():
			victimVal := getPieceValue(pieceType(move.getEndSquare().piece))
			if move.getIsEnpassant() {
				victimVal = getPieceValue(Pawn)
			}
			attackerVal := getPieceValue(pieceType(move.getStartSquare().piece))
			score = 100_000 + victimVal - attackerVal
		case ctx.killers[ply][0].getMoveBits() != 0 && comparePackedMoves(moveBits, ctx.killers[ply][0].getMoveBits()):
			score = 90_000
		case ctx.killers[ply][1].getMoveBits() != 0 && comparePackedMoves(moveBits, ctx.killers[ply][1].getMoveBits()):
			score = 80_000
		default:
			fromSq := move.getStartSquare().row*8 + move.getStartSquare().col
			toSq := move.getEndSquare().row*8 + move.getEndSquare().col
			score = ctx.history[side][fromSq][toSq]
		}
		scored[i] = ScoredMove{move: move, score: score}
	}
	return scored
}

// ChessEngine holds shared state.
type ChessEngine struct {
	ctx                 SearchContext
	transpositionTable  *TranspositionTable
	searchCancelled     *atomic.Bool
	softSearchCancelled *atomic.Bool
	softTimer           *time.Timer
	hardTimer           *time.Timer
	timer               int
}

func (s *ChessEngine) initSearchTranspositionTable() {
	s.transpositionTable = new(TranspositionTable)
	s.searchCancelled = new(atomic.Bool)
	s.softSearchCancelled = new(atomic.Bool)
}

func newWorker(chessEngine *ChessEngine) ChessEngine {
	ctx := SearchContext{}
	ctx.board = chessEngine.ctx.board.Clone()
	ctx.initMoveGeneratorPools()
	return ChessEngine{ctx: ctx,
		transpositionTable:  chessEngine.transpositionTable,
		searchCancelled:     chessEngine.searchCancelled,
		softSearchCancelled: chessEngine.softSearchCancelled,
		timer:               chessEngine.timer,
	}
}

func (s *ChessEngine) workerSearch(depth int) int {
	board := s.ctx.board
	rootMG := &s.ctx.searchMG[0]
	moves := rootMG.generateMoves(false)
	slices.SortFunc(moves, compareMoves)
	sortedMoves := make([]MoveEvaluation, len(moves))
	nodes_searched := 0

	for searchDepth := depth; searchDepth < 200; searchDepth++ {
		if s.softSearchCancelled.Load() {
			break
		}

		for i := range moves {
			move := &moves[i]
			board.makeMove(move)
			eval, nodes := s.searchBruteForce(searchDepth, 1, -20000, 20000, true)
			eval = -eval
			nodes_searched += nodes
			board.unmakeMove(move)
			if s.searchCancelled.Load() {
				return nodes_searched
			}

			sortedMoves[i] = MoveEvaluation{evaluation: eval, move: move}
		}

		if s.searchCancelled.Load() {
			return nodes_searched
		}

		s.ctx.ageHistory()

		moves = nil
		slices.SortFunc(sortedMoves, compareEvaluationMoves)
		for i := range sortedMoves {
			moves = append(moves, *sortedMoves[i].move)
		}
	}
	return nodes_searched
}

func (s *ChessEngine) startSearchTimer() {
	s.searchCancelled.Store(false)
	s.softSearchCancelled.Store(false)

	timer := s.timer
	if timer == 0 {
		timer = 1000
	}
	softTimerLength := timer * 6 / 10
	s.softTimer = time.NewTimer(time.Duration(softTimerLength) * time.Millisecond)

	go func() {
		<-s.softTimer.C
		s.softSearchCancelled.Store(true)
	}()

	s.hardTimer = time.NewTimer(time.Duration(timer) * time.Millisecond)
	go func() {
		<-s.hardTimer.C
		s.searchCancelled.Store(true)
	}()

}

func (s *ChessEngine) setTimer(timer int) {
	s.timer = timer
}

func (s *ChessEngine) bestMove() (MoveString, int, int, int) {
	s.ctx.initMoveGeneratorPools()
	s.ctx.clearKillers()
	s.ctx.clearHistory()

	board := s.ctx.board
	rootMG := &s.ctx.searchMG[0]
	moves := rootMG.generateMoves(false)
	slices.SortFunc(moves, compareMoves)
	totalEvaluated := 0
	sortedMoves := make([]MoveEvaluation, len(moves))

	threads := runtime.NumCPU()
	workerNodes := make([]int, threads)

	s.startSearchTimer()
	var wg sync.WaitGroup
	if multithreading {
		for i := range threads {
			wg.Add(1)
			worker := newWorker(s)
			go func(w *ChessEngine, startDepth int, nodeSlot *int) {
				defer wg.Done()
				*nodeSlot = w.workerSearch(startDepth)
			}(&worker, i+1, &workerNodes[i])
		}
	}

	var bestMove Move
	var bestMoveCurrentIteration Move
	bestEvalCompleted := -40000
	completedDepth := 1

	for searchDepth := range 200 {
		if searchDepth == 0 {
			continue
		}

		if s.softSearchCancelled.Load() {
			break
		}

		bestEvalCurrentIteration := -40000
		bestMoveCurrentIteration = bestMove
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

		s.ctx.ageHistory()

		moves = nil
		slices.SortFunc(sortedMoves, compareEvaluationMoves)
		for i := range sortedMoves {
			moves = append(moves, *sortedMoves[i].move)
		}
	}

	s.searchCancelled.Store(true)
	s.softTimer.Stop()
	s.hardTimer.Stop()
	wg.Wait()

	for _, n := range workerNodes {
		totalEvaluated += n
	}

	return moveToMovestring(bestMove), totalEvaluated, bestEvalCompleted, completedDepth
}

func (s *ChessEngine) searchBruteForce(depth int, ply int, alpha int, beta int, allowNull bool) (int, int) {
	if s.searchCancelled.Load() {
		return 0, 0
	}

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
			return evaluation, 0
		} else if flag == 1 && evaluation >= beta {
			return evaluation, 0
		} else if flag == 2 && evaluation <= alpha {
			return evaluation, 0
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
	if depth >= 3 && !board.inCheck && !board.isPawnEndgame() && allowNull {
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

	side := 0
	if !board.isWhiteTurn {
		side = 1
	}

	scored := s.ctx.scoreMoves(moves, ply, ttMove, side)
	slices.SortFunc(scored, func(a, b ScoredMove) int {
		return cmp.Compare(b.score, a.score)
	})
	for i := range moves {
		moves[i] = scored[i].move
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
			isQuiet := move.getEndSquare().piece == EmptyPiece && !move.getIsPromotion() && !move.getIsEnpassant()
			if isQuiet {
				ctx := &s.ctx
				if !comparePackedMoves(move.getMoveBits(), ctx.killers[ply][0].getMoveBits()) {
					ctx.killers[ply][1] = ctx.killers[ply][0]
					ctx.killers[ply][0] = *move
				}
				fromSq := move.getStartSquare().row*8 + move.getStartSquare().col
				toSq := move.getEndSquare().row*8 + move.getEndSquare().col
				ctx.history[side][fromSq][toSq] = min(ctx.history[side][fromSq][toSq]+depth*depth, maxHistoryScore)
			}
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
