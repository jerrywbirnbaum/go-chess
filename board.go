package main

import (
	"fmt"
	// "slices"
	"strings"
	"unicode"
)

type MovePiece int

const (
	Pawn MovePiece = iota
	Bishop
	Knight
	Rook
	Queen
	King
)

type Board struct {
	board       [8][8]rune
	isWhiteTurn bool
}
type Square struct {
	row int
	col int
}

func (b Board) printBoard() {
	for i := range 8 {
		for j := range 8 {
			fmt.Printf("%q", b.board[i][j])
		}
		fmt.Println()
	}
	fmt.Println("Is white's turn:", b.isWhiteTurn)
}

func (b *Board) moveAlgebraicNotation(move_string string) int {
	// Only works for the white pieces
	// var movedPiece MovePiece

	//Pawn Moves
	// TODO: Add en passant
	if unicode.IsLower(rune(move_string[0])) {
		// movedPiece = Pawn
		col := int(move_string[0]) - 'a'
		row := 0
		if len(move_string) == 2 {
			row = int(move_string[1]) - '0'
			for i := 0; i < 8; i++ {
				if b.board[i][col] == 'P' {
					b.board[i][col] = '*'
					b.board[row][col] = 'P'
					break
				}
			}
		} else if len(move_string) == 4 {
			row = int(move_string[3]) - '0'
			col = int(move_string[2]) - 'a'
			from_col := int(move_string[0]) - 'a'
			b.board[8-row+1][from_col] = '*'
			b.board[8-row][col] = 'P'
		}
	} else {
		strings.Replace(move_string, "x", "A", -1)
	}
	fmt.Println(legalKnightMoves(4, 4))

	return -1
}

func legalKnightMoves(row int, col int) []Square {
	result := []Square{}
	if row > 2 && col > 0 {
		result = append(result, Square{row - 2, col - 1})
	}
	if row > 2 && col > 0 {
		result = append(result, Square{row - 2, col - 1})
	}
	return result

}
