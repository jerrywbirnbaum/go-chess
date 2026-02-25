package main

import (
	"fmt"
	"strings"
	"unicode"
)

type Square struct {
	row   int
	col   int
	piece Piece
}

type Board struct {
	board       [8][8]Piece
	isWhiteTurn bool
	enpassant   string
}

func (b Board) cellEmpty(row int, col int) bool {
	return isEmpty(b.board[row][col])
}

func (b Board) canCapture(row int, col int, color Color) bool {
	if isEmpty(b.board[row][col]) {
		return false
	}
	return !sameColor(b.board[row][col], color)
}

func (b Board) printBoard() {
	for i := range 8 {
		for j := range 8 {
			fmt.Printf("%q", printPiece(b.board[i][j]))
		}
		fmt.Println()
	}
	fmt.Println("Is white's turn:", b.isWhiteTurn)
}

func (b *Board) updateFromFEN(fen_string string) {
	fen_list := strings.Split(fen_string, " ")
	board_fen_string := fen_list[0]

	turn := fen_list[1]
	b.updateTurnFEN(turn)

	// TODO: Update castle and enpassant rules
	// castle := fen_list[2]
	en_passant := fen_list[3]
	b.enpassant = en_passant
	// halfmove_clock := fen_list[4]
	// fullmove_number := fen_list[5]

	b.updateBoardFEN(board_fen_string)

}
func (b *Board) updateTurnFEN(turn_fen_string string) {
	if strings.ContainsRune(turn_fen_string, 'w') {
		b.isWhiteTurn = true
	} else {
		b.isWhiteTurn = false
	}
}

func (b *Board) updateBoardFEN(board_fen_string string) {
	board_rows := strings.Split(board_fen_string, "/")
	for i := range 8 {
		j := 0
		for idx, c := range board_rows[i] {
			_ = idx
			if unicode.IsDigit(c) {
				digit := int(c - '0')
				for k := range digit {
					_ = k
					b.board[i][j] = newPiece('*')
					j += 1
				}
			} else {
				b.board[i][j] = newPiece(c)
				j += 1
			}
		}
	}
}

func (b *Board) piecesGenerator() []Square {
	pieces := []Square{}
	for i := range 8 {
		for j := range 8 {
			if !isEmpty(b.board[i][j]) {
				pieces = append(pieces, Square{row: i, col: j, piece: b.board[i][j]})
			}
		}
	}
	return pieces
}

func (b *Board) attackedBoard(color Color) [8][8]int {
	attacks := [8][8]int{
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}

	moveGenerator := MoveGenerator{board: *b}

	moves := moveGenerator.generateMoves(color)
	for _, move := range moves {
		fmt.Printf("row: %dcol: %d\n", move.endSquare.row, move.endSquare.row)
		attacks[move.endSquare.row][move.endSquare.row] += 1
	}

	return attacks
}
