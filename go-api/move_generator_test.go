package main

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func TestMoveGeneration(t *testing.T) {
	fmt.Println()
	board := initBoard()
	board.updateFromFEN("8/8/3p4/4P3/8/8/8/8 b KQkq - 0 1")
	moveGenerator := MoveGenerator{board: &board}

	moves := moveGenerator.generateMoves(false)

	if len(moves) != 2 {
		t.Errorf("Failed TestMoveGen")
	}

	board.updateFromFEN("8/8/8/8/2pP4/8/8/8 b KQkq d3 0 1")
	moveGenerator.updateBoard(&board)
	moves = moveGenerator.generateMoves(false)

	if len(moves) != 2 {
		t.Errorf("Failed TestMoveGen")
	}
}

func TestKnightMoveGeneration(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("n7/8/1p6/8/3n4/8/8/8 b KQkq d3 0 1")
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 10 {
		t.Errorf("Failed TestMoveGen Knight")
	}

	board.updateFromFEN("r3R3/b7/8/8/8/8/8/8 b KQkq d3 0 1")
	moveGenerator.updateBoard(&board)
	moves = moveGenerator.generateMoves(false)
	if len(moves) != 11 {
		t.Errorf("Failed TestMoveGen Sliding")
	}

	board.updateFromFEN("8/8/8/8/3k4/8/8/8 b KQkq d3 0 1")
	moveGenerator.updateBoard(&board)
	moves = moveGenerator.generateMoves(false)
	if len(moves) != 8 {
		t.Errorf("Failed TestMoveGen King")
	}
}

func TestMoveGenerationDoubleCheck(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("1k5r/8/2N5/4Q3/8/8/8/8 b KQkq d3 0 1")

	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	// fmt.Println(moves)
	if len(moves) != 3 {
		t.Errorf("Failed TestMoveGen DoubleCheck")
	}

}

func TestAttackedBoard(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("rnbqkbnr/pppppppp/PPP3PP/8/8/8/8/RNBQKBNR w KQkq - 0 1")
	expected := [8][8]int{
		{0, 1, 1, 1, 1, 1, 1, 0},
		{1, 1, 1, 4, 4, 1, 1, 1},
		{2, 2, 3, 2, 2, 3, 2, 2},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
	moveGenerator := MoveGenerator{board: &board}
	attacks := moveGenerator.generateAttacks(Color(Black), false)
	if !reflect.DeepEqual(attacks, expected) {
		t.Errorf("Failed generate attacks")
	}

	board.updateFromFEN("rK6/8/8/8/8/8/8/8 w KQkq - 0 1")
	expected = [8][8]int{
		{0, 1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
	}
	moveGenerator.updateBoard(&board)
	attacks = moveGenerator.generateAttacks(Color(Black), false)
	if !reflect.DeepEqual(attacks, expected) {
		t.Errorf("Failed generate attacks")
	}

}

func TestFromSquare(t *testing.T) {
	row, col := fromSquare("a1")
	if row != 7 || col != 0 {
		t.Errorf("Failed TestFromSquare")
	}

	row, col = fromSquare("c3")
	if row != 5 || col != 2 {
		t.Errorf("Failed TestFromSquare")
	}
}

func TestMoveGenerationChecks(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("k7/3r4/2n5/8/8/R7/8/8 b KQkq - 0 1")

	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 5 {
		t.Errorf("Failed TestMoveGen Check")
	}
}

func TestCheckRaysPawn(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("k7/1P6/8/8/8/8/8/8 b KQkq - 0 1")

	expected := [8][8]int{
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 1, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
	moveGenerator := MoveGenerator{board: &board}
	result := moveGenerator.checkRays(0, 0)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Failed Check Rays")
	}
}

func TestCheckRaysKnight(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("k7/8/1N6/8/8/8/8/8 b KQkq - 0 1")

	expected := [8][8]int{
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 1, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
	moveGenerator := MoveGenerator{board: &board}
	result := moveGenerator.checkRays(0, 0)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Failed Check Rays")
	}
}

func TestCheckRaysRook(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("k7/8/8/8/8/8/8/R7 b KQkq - 0 1")
	expected := [8][8]int{
		{0, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
	}
	moveGenerator := MoveGenerator{board: &board}
	result := moveGenerator.checkRays(0, 0)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Failed Check Rays")
	}
}

func TestCheckRaysBishop(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("k7/8/8/8/8/8/6B1/8 b KQkq - 0 1")
	expected := [8][8]int{
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 1, 0, 0, 0, 0, 0, 0},
		{0, 0, 1, 0, 0, 0, 0, 0},
		{0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0},
		{0, 0, 0, 0, 0, 1, 0, 0},
		{0, 0, 0, 0, 0, 0, 1, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
	moveGenerator := MoveGenerator{board: &board}
	result := moveGenerator.checkRays(0, 0)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Failed Check Rays")
	}
}

func TestPinnedPieces(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("kn4Q1/1b6/r7/8/8/8/6B1/R7 b KQkq - 0 1")
	moveGenerator := MoveGenerator{board: &board}
	result := moveGenerator.pinnedPieces(0, 0)
	expected := [8][8]int{
		{0, 1, 0, 0, 0, 0, 0, 0},
		{0, 1, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Failed Pinned Pieces")
	}
}

func TestMoveGenerationPinnedKnight(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("k3n2R/8/8/8/8/8/8/8 b KQkq - 0 1")
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 3 {
		t.Errorf("Failed TestMoveGenerationPinned")
	}

}

func TestMoveGenerationPinned(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("k3q2R/7R/8/8/8/8/8/1R6 b KQkq - 0 1")
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 6 {
		t.Errorf("Failed TestMoveGenerationPinned")
	}
}

func TestMoveGenerationCastle(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("r3k2r/8/8/8/8/8/8/7K b KQkq - 0 1")
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 26 {
		t.Errorf("Failed TestMoveGenerationPinned")
	}
}

func TestMoveGenerationChecksStarting(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("rnbqkb1r/pppppppp/7n/1B6/8/4P3/PPPP1PPP/RNBQK1NR b KQkq - 0 1")
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 17 {
		t.Errorf("Failed  TestMoveGenerationChecksStarting")
	}
}

func TestGenerateAttacksPawnOutOfBoundsGuard(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("k6K/P7/8/8/8/8/8/8 w - - 0 1")
	moveGenerator := MoveGenerator{board: &board}

	attacks := moveGenerator.generateAttacks(Color(White), false)
	if attacks[0][0] != 0 {
		t.Errorf("unexpected attack count on a8: got %d", attacks[0][0])
	}
}

func TestMoveGenerationDoublePin(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("k1rr3Q/8/8/8/8/8/8/8 b - - 0 1")
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 22 {
		t.Errorf("Failed  TestMoveGenerationChecksStarting")
	}
}

func TestMoveGenerationEnpassantCheck(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("8/8/8/8/k2Pp1R1/8/8/8 b - d3 0 1")
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 6 {
		t.Errorf("Failed  TestMoveGenerationChecksStarting")
	}
}

func moveGenerationRecursive(depth int, board Board) int {
	if depth == 0 {
		return 1
	}

	numPositions := 0
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	for i := range moves {
		move := &moves[i]
		board.makeMove(move)
		numPositions += moveGenerationRecursive(depth-1, board)
		board.unmakeMove(move)
	}
	return numPositions
}

func TestMultipleMoves(t *testing.T) {
	board := initBoard()
	result := moveGenerationRecursive(4, board)

	if result != 197281 {
		t.Errorf("Failed RecursiveMoveGeneration")
	}
}

func TestMoveSortFunction(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("8/4p2k/1R1Q4/8/2n5/8/1B1P4/7K b - - 0 1")
	mg := MoveGenerator{board: &board}
	moves := mg.generateMoves(false)
	sort.Sort(MoveOrder(moves))

	got := []MoveString{}
	for _, move := range moves {
		got = append(got, MoveString{
			startSquare: fmt.Sprintf("%c%d", byte('a'+move.startSquare.col), 8-move.startSquare.row),
			endSquare:   fmt.Sprintf("%c%d", byte('a'+move.endSquare.col), 8-move.endSquare.row),
		})
	}

	expected := []MoveString{
		{startSquare: "e7", endSquare: "d6"},
		{startSquare: "c4", endSquare: "d6"},
		{startSquare: "c4", endSquare: "b6"},
		{startSquare: "c4", endSquare: "b2"},
		{startSquare: "e7", endSquare: "e6"},
		{startSquare: "e7", endSquare: "e5"},
		{startSquare: "h7", endSquare: "g8"},
		{startSquare: "c4", endSquare: "e3"},
		{startSquare: "c4", endSquare: "a5"},
		{startSquare: "c4", endSquare: "e5"},
		{startSquare: "c4", endSquare: "a3"},
		{startSquare: "c4", endSquare: "d2"},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Failed Move Ordering Test")
	}
}
