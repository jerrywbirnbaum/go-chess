package main

import (
	"fmt"
	"testing"
)

func TestSameColor(t *testing.T) {
	fmt.Println()
	piece := newPiece('p')
	color := Color(Black)

	result := sameColor(piece, color)
	if !result {
		t.Errorf("Failed TestSameColor")
	}

}

func TestMakeMove(t *testing.T) {
	board := initBoard()
	piece := newPiece('P')
	startSquare := Square{row: 6, col: 0, piece: piece}
	endSquare := Square{row: 5, col: 0, piece: newPiece('*')}
	move := Move{startSquare: startSquare, endSquare: endSquare}
	board.makeMove(&move)

	expected := `'r''n''b''q''k''b''n''r'
'p''p''p''p''p''p''p''p'
'*''*''*''*''*''*''*''*'
'*''*''*''*''*''*''*''*'
'*''*''*''*''*''*''*''*'
'P''*''*''*''*''*''*''*'
'*''P''P''P''P''P''P''P'
'R''N''B''Q''K''B''N''R'
Is white's turn: false`
	result := board.printBoard()

	if result != expected {
		t.Errorf("Failed MakeMove")
	}
	board.unmakeMove(&move)

	expected = `'r''n''b''q''k''b''n''r'
'p''p''p''p''p''p''p''p'
'*''*''*''*''*''*''*''*'
'*''*''*''*''*''*''*''*'
'*''*''*''*''*''*''*''*'
'*''*''*''*''*''*''*''*'
'P''P''P''P''P''P''P''P'
'R''N''B''Q''K''B''N''R'
Is white's turn: true`
	result = board.printBoard()

	if result != expected {
		t.Errorf("Failed UnmakeMove")
	}
}

func TestCalculateZobristIndexWhitePieces(t *testing.T) {
	tests := []struct {
		name     string
		piece    Piece
		row      int
		col      int
		expected int
	}{
		{name: "white pawn a8", piece: WhitePawn, row: 0, col: 0, expected: 0},
		{name: "white bishop h8", piece: WhiteBishop, row: 0, col: 7, expected: 71},
		{name: "white knight a7", piece: WhiteKnight, row: 1, col: 0, expected: 136},
		{name: "white rook h1", piece: WhiteRook, row: 7, col: 7, expected: 255},
		{name: "white queen d4", piece: WhiteQueen, row: 4, col: 3, expected: 291},
		{name: "white king h1", piece: WhiteKing, row: 7, col: 7, expected: 383},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := calculateZobristIndex(tc.piece, tc.row, tc.col)
			if got != tc.expected {
				t.Fatalf("calculateZobristIndex(%v, %d, %d) = %d, expected %d", tc.piece, tc.row, tc.col, got, tc.expected)
			}
		})
	}
}

func TestCalculateZobristIndexBlackPieces(t *testing.T) {
	tests := []struct {
		name     string
		piece    Piece
		row      int
		col      int
		expected int
	}{
		{name: "black pawn a8", piece: BlackPawn, row: 0, col: 0, expected: 384},
		{name: "black bishop h8", piece: BlackBishop, row: 0, col: 7, expected: 455},
		{name: "black knight a7", piece: BlackKnight, row: 1, col: 0, expected: 520},
		{name: "black rook h1", piece: BlackRook, row: 7, col: 7, expected: 639},
		{name: "black queen d4", piece: BlackQueen, row: 4, col: 3, expected: 675},
		{name: "black king h1", piece: BlackKing, row: 7, col: 7, expected: 767},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := calculateZobristIndex(tc.piece, tc.row, tc.col)
			if got != tc.expected {
				t.Fatalf("calculateZobristIndex(%v, %d, %d) = %d, expected %d", tc.piece, tc.row, tc.col, got, tc.expected)
			}
		})
	}
}

func TestCalculateZobristIndexRangeAndUniqueness(t *testing.T) {
	pieces := []Piece{
		WhitePawn, WhiteBishop, WhiteKnight, WhiteRook, WhiteQueen, WhiteKing,
		BlackPawn, BlackBishop, BlackKnight, BlackRook, BlackQueen, BlackKing,
	}

	seen := make(map[int]Square, 768)
	for _, piece := range pieces {
		for row := range 8 {
			for col := range 8 {
				index := calculateZobristIndex(piece, row, col)
				if index < 0 || index >= 768 {
					t.Fatalf("index out of range for piece=%v row=%d col=%d: %d", piece, row, col, index)
				}
				if prev, exists := seen[index]; exists {
					t.Fatalf("duplicate index=%d for piece=%v row=%d col=%d; first seen at piece=%v row=%d col=%d",
						index, piece, row, col, prev.piece, prev.row, prev.col)
				}
				seen[index] = Square{row: row, col: col, piece: piece}
			}
		}
	}

	if len(seen) != 768 {
		t.Fatalf("expected 768 unique indices, got %d", len(seen))
	}
}

func testZobrishTable() [781]int64 {
	table := [781]int64{}
	for i := range table {
		table[i] = int64(i + 1)
	}
	return table
}

func TestCalculateZobrishHashPiecesOnly(t *testing.T) {
	table := testZobrishTable()
	board := Board{
		isWhiteTurn:     true,
		castleAvailable: 0,
		enpassant:       "-",
		zobrishTable:    table,
		pieceCount:      3,
	}
	board.pieces[0] = Square{row: 7, col: 4, piece: WhiteKing}
	board.pieces[1] = Square{row: 0, col: 4, piece: BlackKing}
	board.pieces[2] = Square{row: 6, col: 3, piece: WhitePawn}

	expected := table[calculateZobristIndex(WhiteKing, 7, 4)] ^
		table[calculateZobristIndex(BlackKing, 0, 4)] ^
		table[calculateZobristIndex(WhitePawn, 6, 3)]

	got := board.calculateZobrishHash()
	if got != expected {
		t.Fatalf("calculateZobrishHash() = %d, expected %d", got, expected)
	}
}

func TestCalculateZobrishHashIncludesState(t *testing.T) {
	table := testZobrishTable()
	board := Board{
		isWhiteTurn:     false,
		castleAvailable: CastleWK | CastleBQ,
		enpassant:       "e3",
		zobrishTable:    table,
		pieceCount:      2,
	}
	board.pieces[0] = Square{row: 7, col: 4, piece: WhiteKing}
	board.pieces[1] = Square{row: 0, col: 4, piece: BlackKing}

	expected := table[calculateZobristIndex(WhiteKing, 7, 4)] ^
		table[calculateZobristIndex(BlackKing, 0, 4)] ^
		table[768] ^
		table[769] ^
		table[772] ^
		table[733+4]

	got := board.calculateZobrishHash()
	if got != expected {
		t.Fatalf("calculateZobrishHash() = %d, expected %d", got, expected)
	}
}

func TestCalculateZobrishHashRestoresAfterUnmakeMove(t *testing.T) {
	board := initBoard()
	initialHash := board.calculateZobrishHash()

	move := Move{
		startSquare: Square{row: 6, col: 4, piece: board.board[6][4]},
		endSquare:   Square{row: 4, col: 4, piece: board.board[4][4]},
	}
	board.makeMove(&move)
	board.unmakeMove(&move)

	got := board.calculateZobrishHash()
	if got != initialHash {
		t.Fatalf("hash mismatch after make/unmake: got %d, expected %d", got, initialHash)
	}
}
