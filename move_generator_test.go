package main

import (
	"fmt"
	"testing"
)

func TestMoveGeneration(t *testing.T) {
	fmt.Println("TestMoveGeneration")
	board := Board{
		board: [8][8]Piece{
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
		},
		isWhiteTurn: true,
	}

	board.updateFromFEN("8/8/3p4/4P3/8/8/8/8 b KQkq - 0 1")
	moveGenerator := MoveGenerator{board: board}

	moves := moveGenerator.generateMoves(Color(Black))

	if len(moves) != 2 {
		t.Errorf("Failed TestMoveGen")
	}

	board.updateFromFEN("8/8/8/8/2pP4/8/8/8 b KQkq d3 0 1")
	moveGenerator.updateBoard(board)
	moves = moveGenerator.generateMoves(Color(Black))

	if len(moves) != 2 {
		t.Errorf("Failed TestMoveGen")
	}

	board.updateFromFEN("n7/8/8/8/3n4/8/8/8 b KQkq d3 0 1")
	moveGenerator.updateBoard(board)
	moves = moveGenerator.generateMoves(Color(Black))
	if len(moves) != 10 {
		t.Errorf("Failed TestMoveGen Knight")
	}

	board.updateFromFEN("r3R3/b7/8/8/8/8/8/8 b KQkq d3 0 1")
	moveGenerator.updateBoard(board)
	moves = moveGenerator.generateMoves(Color(Black))
	if len(moves) != 11 {
		t.Errorf("Failed TestMoveGen Knight")
	}

	board.updateFromFEN("8/8/8/8/3k4/8/8/8 b KQkq d3 0 1")
	moveGenerator.updateBoard(board)
	moves = moveGenerator.generateMoves(Color(Black))
	// fmt.Println(len(moves))
	// fmt.Println(moves)
	if len(moves) != 8 {
		t.Errorf("Failed TestMoveGen King")
	}

}

func TestAttackedBoard(t *testing.T) {
	board := Board{
		board: [8][8]Piece{
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
		},
		isWhiteTurn: true,
	}

	board.updateFromFEN("rnbqkbnr/pppppppp/PPP3PP/8/8/8/8/RNBQKBNR w KQkq - 0 1")

	moveGenerator := MoveGenerator{board: board}
	attacks := moveGenerator.generateAttacks(Color(White))
	// attacks := board.attackedBoard(Color(Black))
	fmt.Println(attacks)
	// if !result {
	// 	t.Errorf("Failed TestSameColor")
	// }

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
