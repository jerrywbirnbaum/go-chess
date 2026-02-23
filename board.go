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
}

func (b Board) cellEmpty(row int, col int) bool {
	return isEmpty(b.board[row][col])
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
	// en_passant := fen_list[3]
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

// func (b *Board) moveAlgebraicNotation(move_string string) int {
// 	// Only works for the white pieces
// 	// var movedPiece MovePiece
//
// 	//Pawn Moves
// 	// TODO: Add en passant
// 	if unicode.IsLower(rune(move_string[0])) {
// 		// movedPiece = Pawn
// 		col := int(move_string[0]) - 'a'
// 		row := 0
// 		if len(move_string) == 2 {
// 			row = int(move_string[1]) - '0'
// 			for i := range 8 {
// 				if b.board[i][col] == 'P' {
// 					b.board[i][col] = '*'
// 					b.board[row][col] = 'P'
// 					break
// 				}
// 			}
// 		} else if len(move_string) == 4 {
// 			row = int(move_string[3]) - '0'
// 			col = int(move_string[2]) - 'a'
// 			from_col := int(move_string[0]) - 'a'
// 			b.board[8-row+1][from_col] = '*'
// 			b.board[8-row][col] = 'P'
// 		}
// 	} else {
// 		strings.Replace(move_string, "x", "A", -1)
// 	}
// 	fmt.Println(legalKnightMoves(4, 4))
//
// 	return -1
// }
//
// func legalKnightMoves(row int, col int) []Square {
// 	result := []Square{}
// 	if row > 2 && col > 0 {
// 		result = append(result, Square{row - 2, col - 1})
// 	}
// 	if row > 2 && col > 0 {
// 		result = append(result, Square{row - 2, col - 1})
// 	}
// 	return result
//
// }
