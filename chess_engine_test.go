package main

import (
	"testing"
)

func TestSearchBruteForceDepthZeroMatchesBasicEval(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("8/8/8/3p4/3P4/8/8/K6k w - - 0 1")
	got := searchBruteForce(0, -20000, 20000, &board)
	want := basicEval(board)
	if got != want {
		t.Fatalf("depth 0 should return static evaluation: got %v, want %v", got, want)
	}
}

func TestSearchBruteForceDepthZeroContinuesCaptureSequence(t *testing.T) {
	// Forced line:
	// 1. Rxa8 Qxa8, then no captures remain.
	board := initBoard()
	board.updateFromFEN("rq2k3/8/8/8/8/8/8/R3K3 w - - 0 1")

	got := searchBruteForce(0, -20000, 20000, &board)

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
	stopAfterOneCaptureEval := -basicEval(afterFirst)

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
	want := basicEval(afterSecond)

	if got != want {
		t.Fatalf("depth 0 should evaluate after full capture sequence: got %d, want %d", got, want)
	}
	if got == stopAfterOneCaptureEval {
		t.Fatalf("depth 0 stopped after one capture (got %d), expected it to continue searching captures", got)
	}
}

func TestSearchBruteForceCheckmateReturnsNegativeInfinity(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("7k/6Q1/6K1/8/8/8/8/8 b - - 0 1")

	got := searchBruteForce(1, -20000, 20000, &board)
	if got != -20000 {
		t.Fatalf("checkmate position should evaluate to -20000, got %v", got)
	}
}

func TestSearchBruteForceStalemateReturnsZero(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("7k/5Q2/6K1/8/8/8/8/8 b - - 0 1")

	got := searchBruteForce(1, -20000, 20000, &board)
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

	_ = searchBruteForce(2, -20000, 20000, &board)

	if board.printBoard() != before {
		t.Fatalf("search should not mutate board placement")
	}
	if board.castleAvailable != beforeCastle {
		t.Fatalf("search should not mutate castle rights: got %q, want %q", board.castleAvailable, beforeCastle)
	}
	if board.enpassant != beforeEnpassant {
		t.Fatalf("search should not mutate en-passant square: got %q, want %q", board.enpassant, beforeEnpassant)
	}
	if board.isWhiteTurn != beforeTurn {
		t.Fatalf("search should not mutate side to move")
	}
}

func TestBestMoveForcedMove(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("8/8/8/8/4k3/8/6b1/7K w - - 0 1")
	mg := MoveGenerator{board: &board}

	moves := mg.generateMoves(false)
	if len(moves) != 3 {
		t.Fatalf("expected three legal moves, got %d", len(moves))
	}

	got := mg.bestMove()
	if got.startSquare != "h1" || got.endSquare != "g2" {
		t.Fatalf("bestMove selected %s%s, want h1g2", got.startSquare, got.endSquare)
	}
}
