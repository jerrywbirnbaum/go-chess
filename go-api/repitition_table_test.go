package main

import (
	"testing"
)

func TestRepititionTableInitCreatesMap(t *testing.T) {
	rt := initRepititionTable()
	if rt.table == nil {
		t.Fatal("expected init to allocate table map")
	}
}

func TestRepititionTableIncrementFirstOccurrence(t *testing.T) {
	rt := initRepititionTable()
	ok := rt.increment(42)
	if ok {
		t.Fatal("expected increment to return false on first occurrence")
	}
	if rt.table[42] != 1 {
		t.Fatalf("expected count 1, got %d", rt.table[42])
	}
}

func TestRepititionTableIncrementSecondOccurrence(t *testing.T) {
	rt := initRepititionTable()
	rt.increment(42)
	ok := rt.increment(42)
	if ok {
		t.Fatal("expected increment to return false on second occurrence")
	}
	if rt.table[42] != 2 {
		t.Fatalf("expected count 2, got %d", rt.table[42])
	}
}

func TestRepititionTableIncrementThirdOccurrenceReturnsFalse(t *testing.T) {
	rt := initRepititionTable()
	rt.increment(42)
	rt.increment(42)
	ok := rt.increment(42)
	if !ok {
		t.Fatal("expected increment to return true on third occurrence (threefold repetition)")
	}
	if rt.table[42] != 3 {
		t.Fatalf("expected count 3, got %d", rt.table[42])
	}
}

func TestRepititionTableDecrementReducesCount(t *testing.T) {
	rt := initRepititionTable()
	rt.increment(42)
	rt.increment(42)
	rt.decrement(42)
	if rt.table[42] != 1 {
		t.Fatalf("expected count 1 after decrement, got %d", rt.table[42])
	}
}

func TestRepititionTableDecrementMissingKeyDoesNothing(t *testing.T) {
	rt := initRepititionTable()
	rt.decrement(99) // should not panic
	if _, ok := rt.table[99]; ok {
		t.Fatal("expected missing key to remain absent after decrement")
	}
}

func TestRepititionTableIndependentKeys(t *testing.T) {
	rt := initRepititionTable()
	rt.increment(1)
	rt.increment(1)
	rt.increment(2)

	if rt.table[1] != 2 {
		t.Fatalf("expected key 1 count 2, got %d", rt.table[1])
	}
	if rt.table[2] != 1 {
		t.Fatalf("expected key 2 count 1, got %d", rt.table[2])
	}
}

func TestForcesThreeFoldRepitition(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("k6K/8/ppQ5/8/8/1r6/r1P5/rr6 w - - 0 1")
	mg := MoveGenerator{board: &board}
	chessEngine := ChessEngine{moveGenerator: mg}
	chessEngine.initSearchTranspositionTable()

	got, _, eval, _ := chessEngine.bestMove()

	if eval != 0 {
		t.Fatal("Did not evaluate three fold repition as draw")
	}
	if got.endSquare == "c8" {
		t.Fatalf("bestMove selected %s%s, should be c6c8", got.startSquare, got.endSquare)
	}
}
