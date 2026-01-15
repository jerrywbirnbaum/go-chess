package main

import (
	"fmt"
	// "strings"
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

func (b Board) printBoard() {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
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
			println(row, col, from_col)
			b.board[8-row+1][from_col] = '*'
			b.board[8-row][col] = 'P'
		}
	}

	return -1
}

func pawnMove() {
	return
}
