package main

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func bitboardToArray(bb uint64) [8][8]int {
	var arr [8][8]int
	for i := range 8 {
		for j := range 8 {
			if (bb>>uint(i*8+j))&1 == 1 {
				arr[i][j] = 1
			}
		}
	}
	return arr
}

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
	board.updateFromFEN("n6k/8/1p6/8/3n4/8/8/7K b KQkq d3 0 1")
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 13 {
		t.Errorf("Failed TestMoveGen Knight")
	}

	board.updateFromFEN("r3R3/b7/8/8/8/8/8/k6K b KQkq d3 0 1")
	moveGenerator.updateBoard(&board)
	moves = moveGenerator.generateMoves(false)
	if len(moves) != 14 {
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
		{1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
	moveGenerator := MoveGenerator{board: &board}
	attacks, _ := moveGenerator.generateAttacks(Color(Black), false)
	if !reflect.DeepEqual(bitboardToArray(attacks), expected) {
		t.Errorf("Failed generate attacks")
	}

	board.updateFromFEN("rK6/8/8/8/8/8/8/7k w KQkq - 0 1")
	expected = [8][8]int{
		{0, 1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 1, 1},
		{1, 0, 0, 0, 0, 0, 1, 0},
	}
	moveGenerator.updateBoard(&board)
	attacks, _ = moveGenerator.generateAttacks(Color(Black), false)
	if !reflect.DeepEqual(bitboardToArray(attacks), expected) {
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

	if !reflect.DeepEqual(bitboardToArray(result), expected) {
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

	if !reflect.DeepEqual(bitboardToArray(result), expected) {
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
	if !reflect.DeepEqual(bitboardToArray(result), expected) {
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
	if !reflect.DeepEqual(bitboardToArray(result), expected) {
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

	if !reflect.DeepEqual(bitboardToArray(result), expected) {
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

func TestMoveGenerationPromotion(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("7k/3P4/8/8/8/8/8/8 w - - 0 1")
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 4 {
		t.Errorf("Failed TestMoveGenerationPromotion")
	}
}

func TestGenerateAttacksPawnOutOfBoundsGuard(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("k6K/P7/8/8/8/8/8/8 w - - 0 1")
	moveGenerator := MoveGenerator{board: &board}

	attacks, _ := moveGenerator.generateAttacks(Color(White), false)
	if bitboardCheckOne(attacks, 0, 0) {
		t.Errorf("unexpected attack count on a8: got %d", attacks)
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

func TestMoveSortFunctionSpecial(t *testing.T) {
	e4 := Square{row: 4, col: 5, piece: Piece(WhitePawn)}
	f5 := Square{row: 3, col: 5, piece: Piece(BlackKnight)}
	d5 := Square{row: 3, col: 3, piece: Piece(BlackQueen)}
	e5 := Square{row: 3, col: 4, piece: Piece(EmptyPiece)}
	a4 := Square{row: 4, col: 0, piece: Piece(WhitePawn)}
	b5 := Square{row: 3, col: 1, piece: Piece(BlackKnight)}
	a1 := Square{row: 0, col: 0, piece: Piece(WhiteKnight)}
	a2 := Square{row: 0, col: 1, piece: Piece(BlackPawn)}

	moves := []Move{
		{startSquare: a4, endSquare: b5},
		{startSquare: e4, endSquare: e5},
		{startSquare: a1, endSquare: a2},
		{startSquare: e4, endSquare: d5},
		{startSquare: e4, endSquare: f5},
	}
	sort.Sort(MoveOrder(moves))

	expected := []Move{
		{startSquare: e4, endSquare: d5},
		{startSquare: a4, endSquare: b5},
		{startSquare: e4, endSquare: f5},
		{startSquare: a1, endSquare: a2},
		{startSquare: e4, endSquare: e5},
	}
	if !reflect.DeepEqual(moves, expected) {
		t.Errorf("Failed Move Ordering Test")
		fmt.Println(moves)
		fmt.Println(expected)
	}
}

func TestMoveGenerationFailingPosition(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("1r1R1k2/b4ppp/p7/p1p5/4p1q1/2P5/5P1P/5Q1K b - - 1 22")
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 2 {
		t.Errorf("Failed TestMoveGenerationFailingPosition")
	}
}

func TestMoveGenerationPinnedFriendlyPawn(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("1r1R1k2/b4ppp/p7/p1p5/4p1q1/2P5/7P/5Q1K b - - 1 22")
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 2 {
		t.Errorf("Failed TestMoveGenerationPinnedFriendlyPawn")
	}
}

func TestMoveGenerationPinnedEnemyPawn(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("1r1R1k2/b5pp/p7/p1p5/4p1q1/2P5/5P1P/5Q1K b - - 1 22")
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 3 {
		t.Errorf("Failed TestMoveGenerationPinnedEnemyPawn")
	}
}

func TestMoveGenerationCheckBlocked(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("5rk1/5pp1/1Q2p2p/r1n5/1p1bP3/8/PP3qPP/RR4K1 w - - 0 18")
	moveGenerator := MoveGenerator{board: &board}
	moves := moveGenerator.generateMoves(false)
	if len(moves) != 1 {
		t.Errorf("Failed TestMoveGenerationCheck")
	}
}
