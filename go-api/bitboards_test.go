package main

import (
	"fmt"
	"strconv"
	"testing"
)

func TestGetBitBoards(t *testing.T) {
	board := initBoard()
	board.bitboards = [12]uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}

	whitePawnBitBoard := board.getBitboard(Piece(WhitePawn))
	if whitePawnBitBoard != 0 {
		t.Fatal("Failed White Pawn Bitboard")
	}

	whiteKingBitBoard := board.getBitboard(Piece(WhiteKing))
	if whiteKingBitBoard != 5 {
		t.Fatal("Failed White King Bitboard")
	}

	blackRookBitBoard := board.getBitboard(Piece(BlackRook))
	if blackRookBitBoard != 9 {
		t.Fatal("Failed Black Rook Bitboard")
	}
}

func TestSetRemoveBitBoards(t *testing.T) {
	board := initBoard()
	board.bitboards = [12]uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	board.setBitboardPiece(Piece(WhitePawn), 7, 0)
	board.setBitboardPiece(Piece(WhitePawn), 7, 7)
	whitePawnBitBoard := board.getBitboard(Piece(WhitePawn))
	if whitePawnBitBoard != 129 {
		t.Fatalf("Failed Set White Pawn Bitboard, got: %d", whitePawnBitBoard)
	}

	board.removeBitboardPiece(Piece(WhitePawn), 7, 0)
	board.removeBitboardPiece(Piece(WhitePawn), 7, 7)
	whitePawnBitBoard = board.getBitboard(Piece(WhitePawn))
	if whitePawnBitBoard != 0 {
		t.Fatalf("Failed Remove White Pawn Bitboard, got: %d", whitePawnBitBoard)
	}
}
func TestBitBoardsOpening(t *testing.T) {
	board := initBoard()

	whitePawnBitBoard := board.getBitboard(Piece(WhitePawn))
	expected, _ := strconv.ParseUint("000000000000FF00", 16, 64)
	if whitePawnBitBoard != expected {
		t.Fatalf("Failed Set White Pawn Bitboard, got: %d, expected: %d", whitePawnBitBoard, expected)
	}

	blackKnightBitBoard := board.getBitboard(Piece(BlackKnight))
	expected, _ = strconv.ParseUint("4200000000000000", 16, 64)
	if blackKnightBitBoard != expected {
		t.Fatalf("Failed Set Black Knight Bitboard, got: %d, expected: %d", blackKnightBitBoard, expected)
	}

	e2 := Square{row: 6, col: 4, piece: Piece(WhitePawn)}
	e4 := Square{row: 4, col: 4, piece: Piece(EmptyPiece)}
	move := newMove(e2, e4, false, 0)
	board.makeMove(&move)

	whitePawnBitBoard = board.getBitboard(Piece(WhitePawn))
	expected, _ = strconv.ParseUint("000000000800F700", 16, 64)
	if whitePawnBitBoard != expected {
		fmt.Println(bitboardToArray(whitePawnBitBoard))
		t.Fatalf("Failed Set White Pawn Bitboard after move, got: %d, expected: %d", whitePawnBitBoard, expected)
	}
}

func TestBitScanForward(t *testing.T) {
	bitboard, _ := strconv.ParseUint("1000000000000001", 16, 64)
	got := bitScanForward(bitboard)
	if got != 0 {
		t.Fatalf("Failed bitscan forward expected 0 got: %d", got)
	}

	bitboard, _ = strconv.ParseUint("0000000000000080", 16, 64)
	got = bitScanForward(bitboard)
	if got != 7 {
		t.Fatalf("Failed bitscan forward expected 7 got: %d", got)
	}
	bitboard, _ = strconv.ParseUint("7F000000000000080", 16, 64)
	for bitboard != 0 {
		got = bitScanForward(bitboard)
		bitboard &= 0 << got
	}
	if bitboard != 0 {
		t.Fatal("Failed to update bitboard ")
	}
}
