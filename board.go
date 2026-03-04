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
	board           [8][8]Piece
	isWhiteTurn     bool
	enpassant       string
	castleAvailable string
	moveCount       int
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

func (b Board) printBoard() string {
	s := ""
	for i := range 8 {
		for j := range 8 {
			s += fmt.Sprintf("%q", printPiece(b.board[i][j]))
		}
		s += "\n"
	}
	s += fmt.Sprintf("Is white's turn: %t", b.isWhiteTurn)
	return s
}

func (b *Board) updateFromFEN(fen_string string) {
	fen_list := strings.Split(fen_string, " ")
	board_fen_string := fen_list[0]

	turn := fen_list[1]
	b.updateTurnFEN(turn)

	// TODO: Update castle and enpassant rules
	castle := fen_list[2]
	b.castleAvailable = castle
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

	moves := moveGenerator.generateMoves()
	for _, move := range moves {
		fmt.Printf("row: %dcol: %d\n", move.endSquare.row, move.endSquare.row)
		attacks[move.endSquare.row][move.endSquare.row] += 1
	}

	return attacks
}

func initBoard() Board {
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
	board.updateFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	return board
}

func (b *Board) updateCastle(move Move) {
	startCol := move.startSquare.col
	pieceType := pieceType(move.startSquare.piece)
	isWhite := isWhite(move.startSquare.piece)
	if isKing(pieceType) {
		if isWhite {
			b.castleAvailable = strings.ReplaceAll(b.castleAvailable, "Q", "")
			b.castleAvailable = strings.ReplaceAll(b.castleAvailable, "K", "")
		}
		if !isWhite {
			b.castleAvailable = strings.ReplaceAll(b.castleAvailable, "q", "")
			b.castleAvailable = strings.ReplaceAll(b.castleAvailable, "k", "")
		}
	}
	if isRook(pieceType) {
		if isWhite && startCol == 7 {
			b.castleAvailable = strings.ReplaceAll(b.castleAvailable, "K", "")
		} else if isWhite && startCol == 0 {
			b.castleAvailable = strings.ReplaceAll(b.castleAvailable, "K", "")
		} else if !isWhite && startCol == 7 {
			b.castleAvailable = strings.ReplaceAll(b.castleAvailable, "k", "")
		} else if !isWhite && startCol == 0 {
			b.castleAvailable = strings.ReplaceAll(b.castleAvailable, "q", "")
		}
	}
}
func (b *Board) makeMove(move Move) {
	startRow := move.startSquare.row
	startCol := move.startSquare.col
	endRow := move.endSquare.row
	endCol := move.endSquare.col
	pieceType := pieceType(move.startSquare.piece)
	b.updateCastle(move)
	//Double Pawn Push
	if isPawn(pieceType) && (endRow-startRow) == 2 {
		b.enpassant = toSquare(2, startCol)
	} else if isPawn(pieceType) && (endRow-startRow) == 2 {
		b.enpassant = toSquare(5, startCol)
	}

	//Enpassant
	if isPawn(pieceType) && b.enpassant != "-" {
		enpassantRow, enpassantCol := fromSquare(b.enpassant)
		if endRow == enpassantRow && endCol == enpassantCol {
			b.board[enpassantRow][enpassantCol] = move.startSquare.piece
			b.board[startRow][enpassantCol] = newPiece('*')
			b.board[startRow][startCol] = newPiece('*')
			b.moveCount += 1
			b.isWhiteTurn = !b.isWhiteTurn
			return
		}
	}

	//Castling
	if isKing(pieceType) && (endCol-startCol) == 2 {
		b.board[startRow][6] = move.startSquare.piece
		b.board[startRow][5] = b.board[startRow][7]
		b.board[startRow][startCol] = newPiece('*')
		b.board[startRow][7] = newPiece('*')
		b.moveCount -= 1
		b.isWhiteTurn = !b.isWhiteTurn
		return
	} else if isKing(pieceType) && (endCol-startCol) == -2 {
		b.board[startRow][2] = move.startSquare.piece
		b.board[startRow][3] = b.board[startRow][0]
		b.board[startRow][startCol] = newPiece('*')
		b.board[startRow][0] = newPiece('*')
		b.moveCount -= 1
		b.isWhiteTurn = !b.isWhiteTurn
		return
	}

	//Normal Move
	b.board[startRow][startCol] = newPiece('*')
	b.board[endRow][endCol] = move.startSquare.piece
	b.moveCount += 1
	b.isWhiteTurn = !b.isWhiteTurn
}

func (b *Board) unmakeMove(move Move) {
	startRow := move.startSquare.row
	startCol := move.startSquare.col
	endRow := move.endSquare.row
	endCol := move.endSquare.col
	pieceType := pieceType(move.startSquare.piece)

	//Enpassant
	if isPawn(pieceType) && b.enpassant != "-" {
		enpassantRow, enpassantCol := fromSquare(b.enpassant)
		if endRow == enpassantRow && endCol == enpassantCol {
			b.board[enpassantRow][enpassantCol] = newPiece('*')
			b.board[startRow][startCol] = move.startSquare.piece
			if endRow > startRow {
				b.board[startRow][enpassantCol] = newPiece('P')
			} else {
				b.board[startRow][enpassantCol] = newPiece('p')
			}
			b.moveCount -= 1
			b.isWhiteTurn = !b.isWhiteTurn
			return
		}
	}
	// Castling
	if isKing(pieceType) && (endCol-startCol) == 2 {
		b.board[startRow][4] = move.startSquare.piece
		b.board[startRow][7] = b.board[startRow][5]
		b.board[startRow][5] = newPiece('*')
		b.board[startRow][6] = newPiece('*')
		b.moveCount -= 1
		b.isWhiteTurn = !b.isWhiteTurn
		return
	} else if isKing(pieceType) && (endCol-startCol) == -2 {
		b.board[startRow][4] = move.startSquare.piece
		b.board[startRow][0] = b.board[startRow][3]
		b.board[startRow][2] = newPiece('*')
		b.board[startRow][3] = newPiece('*')
		b.moveCount -= 1
		b.isWhiteTurn = !b.isWhiteTurn
		return
	}
	b.board[startRow][startCol] = move.startSquare.piece
	b.board[endRow][endCol] = move.endSquare.piece

	b.moveCount -= 1
	b.isWhiteTurn = !b.isWhiteTurn
}
