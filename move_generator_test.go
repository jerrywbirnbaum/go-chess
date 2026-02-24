package main

import (
	"testing"
)

func TestMoveGeneration(t *testing.T) {
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
	moves = moveGenerator.generateMoves(Color(Black))

	if len(moves) != 2 {
		t.Errorf("Failed TestMoveGen")
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
