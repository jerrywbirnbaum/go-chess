package main

import (
	"fmt"
	"testing"
)

func TestRepititionTableIncrementFirstOccurrence(t *testing.T) {
	rt := initRepititionTable()
	ok := rt.increment(42)
	if ok {
		t.Fatal("expected increment to return false on first occurrence")
	}
	if rt.table[42].count != 1 {
		t.Fatalf("expected count 1, got %d", rt.table[42].count)
	}
}

func TestRepititionTableIncrementSecondOccurrence(t *testing.T) {
	rt := initRepititionTable()
	rt.increment(42)
	ok := rt.increment(42)
	if ok {
		t.Fatal("expected increment to return false on second occurrence")
	}
	if rt.table[42].count != 2 {
		t.Fatalf("expected count 2, got %d", rt.table[42].count)
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
	if rt.table[42].count != 3 {
		t.Fatalf("expected count 3, got %d", rt.table[42].count)
	}
}

func TestRepititionTableDecrementReducesCount(t *testing.T) {
	rt := initRepititionTable()
	rt.increment(42)
	rt.increment(42)
	rt.decrement(42)
	if rt.table[42].count != 1 {
		t.Fatalf("expected count 1 after decrement, got %d", rt.table[42].count)
	}
}

func TestRepititionTableIndependentKeys(t *testing.T) {
	rt := initRepititionTable()
	rt.increment(1)
	rt.increment(1)
	rt.increment(2)

	if rt.table[1].count != 2 {
		t.Fatalf("expected key 1 count 2, got %d", rt.table[1].count)
	}
	if rt.table[2].count != 1 {
		t.Fatalf("expected key 2 count 1, got %d", rt.table[2].count)
	}
}

func TestForcesThreeFoldRepitition(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("k6K/8/ppQ5/8/8/1r6/r1P5/rr6 w - - 0 1")
	chessEngine := ChessEngine{ctx: SearchContext{board: &board}}
	chessEngine.initSearchTranspositionTable()

	got, _, eval, depth := chessEngine.bestMove()

	if eval != 0 {
		fmt.Println(eval)
		fmt.Println(got)
		fmt.Println(depth)
		t.Fatal("Did not evaluate three fold repition as draw")
	}
}
