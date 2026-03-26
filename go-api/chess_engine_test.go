package main

import (
	"sort"
	"testing"
)

func TestSearchBruteForceDepthZeroMatchesBasicEval(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("8/8/8/3p4/3P4/8/8/K6k w - - 0 1")

	mg := MoveGenerator{board: &board}

	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()
	got, _ := chessEngine.searchBruteForce(0, 0, -20000, 20000, true)
	want := pestoEval(board)
	if got != want {
		t.Fatalf("depth 0 should return static evaluation: got %v, want %v", got, want)
	}
}

func TestSearchBruteForceStalemate(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("K1kb2Q1/P1p5/2P5/8/8/8/8/8 b - - 0 1")

	mg := MoveGenerator{board: &board}

	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()
	got, _ := chessEngine.searchBruteForce(1, 0, -20000, 20000, true)
	want := 0
	if got != want {
		t.Fatalf("depth 0 should return static evaluation: got %v, want %v", got, want)
	}
}

func TestSearchBruteForceDepthZeroContinuesCaptureSequence(t *testing.T) {
	// Forced line:
	// 1. Rxa8 Qxa8, then no captures remain.
	board := initBoard()
	board.updateFromFEN("rq2k3/8/8/8/8/8/8/R3K3 w - - 0 1")

	mg := MoveGenerator{board: &board}

	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()
	got, _ := chessEngine.searchBruteForce(0, 0, -20000, 20000, true)

	moveGenerator := MoveGenerator{board: &board}
	firstCaptures := moveGenerator.generateMoves(true)
	if len(firstCaptures) != 1 {
		t.Fatalf("expected exactly one root capture, got %d", len(firstCaptures))
	}
	firstMove := firstCaptures[0]
	if toSquare(firstMove.startSquare.row, firstMove.startSquare.col)+toSquare(firstMove.endSquare.row, firstMove.endSquare.col) != "a1a8" {
		t.Fatalf("expected forced capture a1a8")
	}

	afterFirst := board
	afterFirst.makeMove(&firstMove)
	stopAfterOneCaptureEval := -pestoEval(afterFirst)

	replyGenerator := MoveGenerator{board: &afterFirst}
	secondCaptures := replyGenerator.generateMoves(true)
	if len(secondCaptures) != 1 {
		t.Fatalf("expected exactly one reply capture, got %d", len(secondCaptures))
	}
	secondMove := secondCaptures[0]
	if toSquare(secondMove.startSquare.row, secondMove.startSquare.col)+toSquare(secondMove.endSquare.row, secondMove.endSquare.col) != "b8a8" {
		t.Fatalf("expected forced recapture b8a8")
	}

	afterSecond := afterFirst
	afterSecond.makeMove(&secondMove)

	if got != -948 {
		t.Fatalf("depth 0 should evaluate after full capture sequence: got %d, want %d", got, -948)
	}
	if got == stopAfterOneCaptureEval {
		t.Fatalf("depth 0 stopped after one capture (got %d), expected it to continue searching captures", got)
	}
}

func TestSearchBruteForceCheckmateReturnsNegativeInfinity(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("7k/6Q1/6K1/8/8/8/8/8 b - - 0 1")

	mg := MoveGenerator{board: &board}

	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()
	got, _ := chessEngine.searchBruteForce(1, 0, -20000, 20000, true)
	if got != -20000 {
		t.Fatalf("checkmate position should evaluate to -20000 (mate at ply 0), got %v", got)
	}
}

func TestSearchBruteForceStalemateReturnsZero(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("7k/5Q2/6K1/8/8/8/8/8 b - - 0 1")

	mg := MoveGenerator{board: &board}

	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()
	got, _ := chessEngine.searchBruteForce(1, 0, -20000, 20000, true)
	if got != 0 {
		t.Fatalf("stalemate position should evaluate to 0, got %v", got)
	}
}

func TestSearchBruteForceDoesNotMutateBoardState(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	before := board.printBoard()
	beforeCastle := board.castleAvailable
	beforeEnpassant := board.enpassant
	beforeTurn := board.isWhiteTurn

	mg := MoveGenerator{board: &board}

	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()
	_, _ = chessEngine.searchBruteForce(2, 0, -20000, 20000, true)

	if board.printBoard() != before {
		t.Fatalf("search should not mutate board placement")
	}
	if board.castleAvailable != beforeCastle {
		t.Fatalf("search should not mutate castle rights: got %d, want %d", board.castleAvailable, beforeCastle)
	}
	if board.enpassant != beforeEnpassant {
		t.Fatalf("search should not mutate en-passant square: got %q, want %q", board.enpassant, beforeEnpassant)
	}
	if board.isWhiteTurn != beforeTurn {
		t.Fatalf("search should not mutate side to move")
	}
}

func TestSearchBruteForceTranspositionExactHitReturnsCachedEval(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("7k/6Q1/6K1/8/8/8/8/8 b - - 0 1")

	key := board.calculateZobrishHash()

	mg := MoveGenerator{board: &board}

	chessEngine := ChessEngine{moveGenerator: mg}

	chessEngine.initSearchTranspositionTable()
	chessEngine.transpositionTable.push(key, 4, 0, 321)
	gotEval, gotPositions := chessEngine.searchBruteForce(0, 0, -20000, 20000, true)

	if gotEval != 321 {
		t.Fatalf("expected exact TT hit to return cached eval 321, got %d", gotEval)
	}
	if gotPositions != 1 {
		t.Fatalf("expected exact TT hit to count one position, got %d", gotPositions)
	}
}

func TestSearchBruteForceTranspositionLowerBoundHitReturnsCachedEval(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("7k/6Q1/6K1/8/8/8/8/8 b - - 0 1")

	key := board.calculateZobrishHash()

	mg := MoveGenerator{board: &board}

	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()
	chessEngine.transpositionTable.push(key, 4, 1, 50)

	gotEval, gotPositions := chessEngine.searchBruteForce(2, 0, -20000, 40, true)
	if gotEval != 50 {
		t.Fatalf("expected lower-bound TT hit to return cached eval 50, got %d", gotEval)
	}
	if gotPositions != 1 {
		t.Fatalf("expected lower-bound TT hit to count one position, got %d", gotPositions)
	}
}

func TestSearchBruteForceTranspositionUpperBoundHitReturnsCachedEval(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("7k/6Q1/6K1/8/8/8/8/8 b - - 0 1")

	key := board.calculateZobrishHash()

	mg := MoveGenerator{board: &board}

	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()
	chessEngine.transpositionTable.push(key, 4, 2, -50)
	gotEval, gotPositions := chessEngine.searchBruteForce(2, 0, -40, 20000, true)
	if gotEval != -50 {
		t.Fatalf("expected upper-bound TT hit to return cached eval -50, got %d", gotEval)
	}
	if gotPositions != 1 {
		t.Fatalf("expected upper-bound TT hit to count one position, got %d", gotPositions)
	}
}

func TestSearchBruteForceTranspositionIgnoresShallowEntry(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("7k/6Q1/6K1/8/8/8/8/8 b - - 0 1")

	key := board.calculateZobrishHash()

	mg := MoveGenerator{board: &board}

	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()
	chessEngine.transpositionTable.push(key, 0, 0, 123)
	gotEval, _ := chessEngine.searchBruteForce(1, 0, -20000, 20000, true)
	if gotEval != -20000 {
		t.Fatalf("expected shallow TT entry to be ignored; got %d", gotEval)
	}
}

func TestSearchBruteForceTranspositionIgnoresUnmetLowerBound(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("7k/6Q1/6K1/8/8/8/8/8 b - - 0 1")

	key := board.calculateZobrishHash()

	mg := MoveGenerator{board: &board}

	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()
	chessEngine.transpositionTable.push(key, 4, 1, 30)
	gotEval, _ := chessEngine.searchBruteForce(2, 0, -20000, 40, true)
	if gotEval != -20000 {
		t.Fatalf("expected unmet lower-bound TT entry to be ignored; got %d", gotEval)
	}
}

func TestSearchAvoidCheckmate(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("8/p1kr2p1/2p5/b7/6Q1/PP1b4/1R3PPP/3r1BKR b - - 10 29")

	mg := MoveGenerator{board: &board}
	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()
	gotEval, _ := chessEngine.searchBruteForce(1, 0, -20000, 20000, true)
	if gotEval != 19999 {
		t.Fatalf("Failed TestSearchAvoidCheckmate: got %d", gotEval)
	}
}
func TestBestMoveForcedMove(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("8/8/8/8/4k3/8/6b1/7K w - - 0 1")
	mg := MoveGenerator{board: &board}
	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()

	moves := mg.generateMoves(false)
	if len(moves) != 3 {
		t.Fatalf("expected three legal moves, got %d", len(moves))
	}

	got, _, _, _ := chessEngine.bestMove()
	if got.startSquare != "h1" || got.endSquare != "g2" {
		t.Fatalf("bestMove selected %s%s, want h1g2", got.startSquare, got.endSquare)
	}
}

func TestBestMoveForcedCheckmate(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("8/1p6/8/8/8/1k4K1/1Q4P1/2Q5 b - - 10 74")
	mg := MoveGenerator{board: &board}

	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()
	got, _, _, _ := chessEngine.bestMove()
	if got.startSquare != "b3" || got.endSquare != "a4" {
		t.Fatalf("bestMove selected %s%s, want b3a4", got.startSquare, got.endSquare)
	}
}

func TestBestMoveQueenFork(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("4rb2/pp6/1k2r3/4Q1p1/1n1P4/5NP1/PPn3PP/2R2K1R w - - 0 23")
	mg := MoveGenerator{board: &board}
	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.setTimer(1000)
	chessEngine.initSearchTranspositionTable()

	got, _, _, _ := chessEngine.bestMove()
	if got.endSquare == "f5" {
		t.Fatalf("bestMove selected %s%s, can't be f5", got.startSquare, got.endSquare)
	}
}

func TestBestMoveIllegal(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("1r3rk1/ppp1qppp/2Pbb3/6N1/2P2p2/3BB3/P1P2PPP/RQ3RK1 b - - 0 1")
	mg := MoveGenerator{board: &board}
	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.setTimer(100)
	chessEngine.initSearchTranspositionTable()

	got, _, _, _ := chessEngine.bestMove()
	if got.endSquare == "a8" {
		t.Fatalf("bestMove selected %s%s, can't be a8a8", got.startSquare, got.endSquare)
	}
}
func TestMoveEvaluationOrderSortsDescending(t *testing.T) {
	evals := MoveEvaluationOrder{
		{evaluation: 50, move: nil},
		{evaluation: -100, move: nil},
		{evaluation: 200, move: nil},
		{evaluation: 0, move: nil},
		{evaluation: 75, move: nil},
	}
	sort.Sort(evals)

	for i := 1; i < len(evals); i++ {
		if evals[i].evaluation > evals[i-1].evaluation {
			t.Fatalf("not sorted descending at index %d: evals[%d]=%d > evals[%d]=%d",
				i, i, evals[i].evaluation, i-1, evals[i-1].evaluation)
		}
	}

	if evals[0].evaluation != 200 {
		t.Fatalf("expected highest eval 200 first, got %d", evals[0].evaluation)
	}
	if evals[len(evals)-1].evaluation != -100 {
		t.Fatalf("expected lowest eval -100 last, got %d", evals[len(evals)-1].evaluation)
	}
}

func TestMoveEvaluationOrderEmpty(t *testing.T) {
	evals := MoveEvaluationOrder{}
	sort.Sort(evals) // should not panic
}

func TestMoveEvaluationOrderSingleElement(t *testing.T) {
	evals := MoveEvaluationOrder{{evaluation: 42, move: nil}}
	sort.Sort(evals)
	if evals[0].evaluation != 42 {
		t.Fatalf("expected 42, got %d", evals[0].evaluation)
	}
}

func TestNxc6EvalCleanTT(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("r6r/ppp1kppp/2p1bn2/4N1B1/8/2P5/PPP2PPP/R4RK1 w - - 3 11")

	mg := MoveGenerator{board: &board}
	moves := mg.generateMoves(false)

	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()

	for i := range moves {
		move := &moves[i]
		before_hash := board.zobristHash
		board.makeMove(move)
		board.unmakeMove(move)
		after_hash := board.zobristHash
		if before_hash != after_hash {
			t.Fatal("hash")
		}

	}

	var nxc6 *Move
	for i := range moves {
		s := toSquare(moves[i].startSquare.row, moves[i].startSquare.col)
		e := toSquare(moves[i].endSquare.row, moves[i].endSquare.col)
		if s == "e5" && e == "c6" {
			nxc6 = &moves[i]
			break
		}
	}
	board.makeMove(nxc6)
	eval, _ := chessEngine.searchBruteForce(2, 0, -20000, 20000, true)
	board.unmakeMove(nxc6)

	chessEngine.moveGenerator = mg
	board.makeMove(nxc6)
	chessEngine.searchBruteForce(1, 0, -20000, 20000, true)

	eval3, _ := chessEngine.searchBruteForce(2, 0, -20000, 20000, true)

	board.unmakeMove(nxc6)

	if eval != eval3 {
		t.Fatalf("TT pollution: populated TT gives %d, clean TT gives %d", eval, eval3)
	}
}
