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
	pieces          [32]Square
	pieceCount      int
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
	b.rebuildPieceList()

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
	return b.pieces[:b.pieceCount]
}

func (b *Board) rebuildPieceList() {
	b.pieceCount = 0
	for i := range 8 {
		for j := range 8 {
			p := b.board[i][j]
			if isEmpty(p) {
				continue
			}
			b.pieces[b.pieceCount] = Square{row: i, col: j, piece: p}
			b.pieceCount += 1
		}
	}
}

func (b *Board) removePieceFromList(row int, col int) {
	for i := 0; i < b.pieceCount; i++ {
		p := b.pieces[i]
		if p.row == row && p.col == col {
			lastIdx := b.pieceCount - 1
			b.pieces[i] = b.pieces[lastIdx]
			b.pieces[lastIdx] = Square{}
			b.pieceCount -= 1
			return
		}
	}
}

func (b *Board) setPieceInList(row int, col int, piece Piece) {
	if isEmpty(piece) {
		b.removePieceFromList(row, col)
		return
	}

	for i := 0; i < b.pieceCount; i++ {
		if b.pieces[i].row == row && b.pieces[i].col == col {
			b.pieces[i].piece = piece
			return
		}
	}

	b.pieces[b.pieceCount] = Square{row: row, col: col, piece: piece}
	b.pieceCount += 1
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

	moves := moveGenerator.generateMoves(false)
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

func removeCastleRight(castle string, right string) string {
	return strings.ReplaceAll(castle, right, "")
}

func updateCastleRights(castle string, move *Move) string {
	startRow := move.startSquare.row
	startCol := move.startSquare.col
	endRow := move.endSquare.row
	endCol := move.endSquare.col
	startPieceType := pieceType(move.startSquare.piece)
	endPieceType := pieceType(move.endSquare.piece)
	movingWhite := isWhite(move.startSquare.piece)

	if isKing(startPieceType) {
		if movingWhite {
			castle = removeCastleRight(castle, "K")
			castle = removeCastleRight(castle, "Q")
		} else {
			castle = removeCastleRight(castle, "k")
			castle = removeCastleRight(castle, "q")
		}
	}

	if isRook(startPieceType) {
		if startRow == 7 && startCol == 7 {
			castle = removeCastleRight(castle, "K")
		} else if startRow == 7 && startCol == 0 {
			castle = removeCastleRight(castle, "Q")
		} else if startRow == 0 && startCol == 7 {
			castle = removeCastleRight(castle, "k")
		} else if startRow == 0 && startCol == 0 {
			castle = removeCastleRight(castle, "q")
		}
	}

	if isRook(endPieceType) {
		if endRow == 7 && endCol == 7 {
			castle = removeCastleRight(castle, "K")
		} else if endRow == 7 && endCol == 0 {
			castle = removeCastleRight(castle, "Q")
		} else if endRow == 0 && endCol == 7 {
			castle = removeCastleRight(castle, "k")
		} else if endRow == 0 && endCol == 0 {
			castle = removeCastleRight(castle, "q")
		}
	}

	if castle == "" {
		return "-"
	}
	return castle
}

func updateEnpassantSquare(move *Move) string {
	startRow := move.startSquare.row
	startCol := move.startSquare.col
	endRow := move.endSquare.row
	startPieceType := pieceType(move.startSquare.piece)

	if isPawn(startPieceType) && (endRow-startRow) == 2 {
		return toSquare(2, startCol)
	}
	if isPawn(startPieceType) && (endRow-startRow) == -2 {
		return toSquare(5, startCol)
	}
	return "-"
}

func (b *Board) makeMove(move *Move) {
	startRow := move.startSquare.row
	startCol := move.startSquare.col
	endRow := move.endSquare.row
	endCol := move.endSquare.col
	pieceType := pieceType(move.startSquare.piece)

	move.previousCastleRights = b.castleAvailable
	move.previousEnpassant = b.enpassant
	move.nextCastleRights = updateCastleRights(b.castleAvailable, move)
	move.nextEnpassant = updateEnpassantSquare(move)
	move.isCastleKingSide = isKing(pieceType) && (endCol-startCol) == 2
	move.isCastleQueenSide = isKing(pieceType) && (endCol-startCol) == -2
	move.isPromotion = isPawn(pieceType) && (endRow == 0 || endRow == 7)
	move.isEnpassant = false
	move.enpassantCapture = newPiece('*')

	if isPawn(pieceType) && move.previousEnpassant != "-" && isEmpty(move.endSquare.piece) {
		enpassantRow, enpassantCol := fromSquare(move.previousEnpassant)
		if endRow == enpassantRow && endCol == enpassantCol {
			move.isEnpassant = true
			move.enpassantCapture = b.board[startRow][enpassantCol]
		}
	}

	//Pawn Promotion
	if move.isPromotion {
		var promotedPiece Piece
		if b.isWhiteTurn {
			promotedPiece = newPiece('Q')
		} else {
			promotedPiece = newPiece('q')
		}
		b.board[endRow][endCol] = promotedPiece
		b.board[startRow][startCol] = newPiece('*')
		b.removePieceFromList(startRow, startCol)
		b.setPieceInList(endRow, endCol, promotedPiece)
		b.castleAvailable = move.nextCastleRights
		b.enpassant = move.nextEnpassant
		b.moveCount += 1
		b.isWhiteTurn = !b.isWhiteTurn
		return
	}
	//Enpassant
	if move.isEnpassant {
		b.board[endRow][endCol] = move.startSquare.piece
		b.board[startRow][endCol] = newPiece('*')
		b.board[startRow][startCol] = newPiece('*')
		b.removePieceFromList(startRow, startCol)
		b.removePieceFromList(startRow, endCol)
		b.setPieceInList(endRow, endCol, move.startSquare.piece)
		b.castleAvailable = move.nextCastleRights
		b.enpassant = move.nextEnpassant
		b.moveCount += 1
		b.isWhiteTurn = !b.isWhiteTurn
		return
	}

	//Castling
	if move.isCastleKingSide {
		b.board[startRow][6] = move.startSquare.piece
		b.board[startRow][5] = b.board[startRow][7]
		b.board[startRow][startCol] = newPiece('*')
		b.board[startRow][7] = newPiece('*')
		b.removePieceFromList(startRow, startCol)
		b.removePieceFromList(startRow, 7)
		b.setPieceInList(startRow, 6, move.startSquare.piece)
		b.setPieceInList(startRow, 5, b.board[startRow][5])
		b.castleAvailable = move.nextCastleRights
		b.enpassant = move.nextEnpassant
		b.moveCount += 1
		b.isWhiteTurn = !b.isWhiteTurn
		return
	} else if move.isCastleQueenSide {
		b.board[startRow][2] = move.startSquare.piece
		b.board[startRow][3] = b.board[startRow][0]
		b.board[startRow][startCol] = newPiece('*')
		b.board[startRow][0] = newPiece('*')
		b.removePieceFromList(startRow, startCol)
		b.removePieceFromList(startRow, 0)
		b.setPieceInList(startRow, 2, move.startSquare.piece)
		b.setPieceInList(startRow, 3, b.board[startRow][3])
		b.castleAvailable = move.nextCastleRights
		b.enpassant = move.nextEnpassant
		b.moveCount += 1
		b.isWhiteTurn = !b.isWhiteTurn
		return
	}

	//Normal Move
	b.board[startRow][startCol] = newPiece('*')
	b.board[endRow][endCol] = move.startSquare.piece
	b.removePieceFromList(startRow, startCol)
	b.setPieceInList(endRow, endCol, move.startSquare.piece)
	b.castleAvailable = move.nextCastleRights
	b.enpassant = move.nextEnpassant
	b.moveCount += 1
	b.isWhiteTurn = !b.isWhiteTurn
}

func (b *Board) unmakeMove(move *Move) {
	startRow := move.startSquare.row
	startCol := move.startSquare.col
	endRow := move.endSquare.row
	endCol := move.endSquare.col

	//Enpassant
	if move.isPromotion {
		b.board[startRow][startCol] = move.startSquare.piece
		b.board[endRow][endCol] = move.endSquare.piece
		b.setPieceInList(endRow, endCol, move.endSquare.piece)
		b.setPieceInList(startRow, startCol, move.startSquare.piece)
		b.castleAvailable = move.previousCastleRights
		b.enpassant = move.previousEnpassant
		b.moveCount -= 1
		b.isWhiteTurn = !b.isWhiteTurn
		return
	}

	if move.isEnpassant {
		b.board[startRow][startCol] = move.startSquare.piece
		b.board[endRow][endCol] = newPiece('*')
		b.board[startRow][endCol] = move.enpassantCapture
		b.setPieceInList(startRow, startCol, move.startSquare.piece)
		b.setPieceInList(endRow, endCol, newPiece('*'))
		b.setPieceInList(startRow, endCol, move.enpassantCapture)
		b.castleAvailable = move.previousCastleRights
		b.enpassant = move.previousEnpassant
		b.moveCount -= 1
		b.isWhiteTurn = !b.isWhiteTurn
		return
	}

	// Castling
	if move.isCastleKingSide {
		b.board[startRow][4] = move.startSquare.piece
		b.board[startRow][7] = b.board[startRow][5]
		b.board[startRow][5] = newPiece('*')
		b.board[startRow][6] = newPiece('*')
		b.setPieceInList(startRow, 5, newPiece('*'))
		b.setPieceInList(startRow, 6, newPiece('*'))
		b.setPieceInList(startRow, 4, move.startSquare.piece)
		b.setPieceInList(startRow, 7, b.board[startRow][7])
		b.castleAvailable = move.previousCastleRights
		b.enpassant = move.previousEnpassant
		b.moveCount -= 1
		b.isWhiteTurn = !b.isWhiteTurn
		return
	} else if move.isCastleQueenSide {
		b.board[startRow][4] = move.startSquare.piece
		b.board[startRow][0] = b.board[startRow][3]
		b.board[startRow][2] = newPiece('*')
		b.board[startRow][3] = newPiece('*')
		b.setPieceInList(startRow, 2, newPiece('*'))
		b.setPieceInList(startRow, 3, newPiece('*'))
		b.setPieceInList(startRow, 4, move.startSquare.piece)
		b.setPieceInList(startRow, 0, b.board[startRow][0])
		b.castleAvailable = move.previousCastleRights
		b.enpassant = move.previousEnpassant
		b.moveCount -= 1
		b.isWhiteTurn = !b.isWhiteTurn
		return
	}
	b.board[startRow][startCol] = move.startSquare.piece
	b.board[endRow][endCol] = move.endSquare.piece
	b.setPieceInList(endRow, endCol, move.endSquare.piece)
	b.setPieceInList(startRow, startCol, move.startSquare.piece)
	b.castleAvailable = move.previousCastleRights
	b.enpassant = move.previousEnpassant
	b.moveCount -= 1
	b.isWhiteTurn = !b.isWhiteTurn
}

func (b *Board) currentColor() Color {
	if b.isWhiteTurn {
		return Color(White)
	}
	return Color(Black)
}
func (b *Board) playerInCheck() bool {
	color := b.currentColor()
	moveGenerator := MoveGenerator{board: *b}
	attacks := moveGenerator.generateAttacks(oppositeColor(color), false)
	for i := range 8 {
		for j := range 8 {
			piece := b.board[i][j]
			pieceType := pieceType(piece)
			if attacks[i][j] > 0 && isKing(pieceType) && getColor(piece) == color {
				return true
			}

		}
	}
	return false

}
