package main

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"unicode"
)

type Square struct {
	row   int
	col   int
	piece Piece
}

const (
	CastleWK uint8 = 1 << iota
	CastleWQ
	CastleBK
	CastleBQ
)

type Board struct {
	board           [8][8]Piece
	pieces          [32]Square
	pieceCount      int
	isWhiteTurn     bool
	enpassant       string
	castleAvailable uint8
	moveCount       int
	zobrishTable    [781]int64
	zobristHash     int64
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

	castle := fen_list[2]
	b.castleAvailable = parseCastleRights(castle)
	en_passant := fen_list[3]
	b.enpassant = en_passant
	// halfmove_clock := fen_list[4]
	// fullmove_number := fen_list[5]

	b.updateBoardFEN(board_fen_string)
	b.rebuildPieceList()
	b.zobristHash = b.calculateZobrishHash()
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

	moveGenerator := MoveGenerator{board: b}

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
	board.zobrishTable = board.zobrishHashTable()
	board.updateFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	return board
}

func parseCastleRights(s string) uint8 {
	var rights uint8
	for _, c := range s {
		switch c {
		case 'K':
			rights |= CastleWK
		case 'Q':
			rights |= CastleWQ
		case 'k':
			rights |= CastleBK
		case 'q':
			rights |= CastleBQ
		}
	}
	return rights
}

func updateCastleRights(castle uint8, move *Move) uint8 {
	startRow := move.startSquare.row
	startCol := move.startSquare.col
	endRow := move.endSquare.row
	endCol := move.endSquare.col
	startPieceType := pieceType(move.startSquare.piece)
	endPieceType := pieceType(move.endSquare.piece)
	movingWhite := isWhite(move.startSquare.piece)

	if isKing(startPieceType) {
		if movingWhite {
			castle &^= CastleWK | CastleWQ
		} else {
			castle &^= CastleBK | CastleBQ
		}
	}

	if isRook(startPieceType) {
		if startRow == 7 && startCol == 7 {
			castle &^= CastleWK
		} else if startRow == 7 && startCol == 0 {
			castle &^= CastleWQ
		} else if startRow == 0 && startCol == 7 {
			castle &^= CastleBK
		} else if startRow == 0 && startCol == 0 {
			castle &^= CastleBQ
		}
	}

	if isRook(endPieceType) {
		if endRow == 7 && endCol == 7 {
			castle &^= CastleWK
		} else if endRow == 7 && endCol == 0 {
			castle &^= CastleWQ
		} else if endRow == 0 && endCol == 7 {
			castle &^= CastleBK
		} else if endRow == 0 && endCol == 0 {
			castle &^= CastleBQ
		}
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

func (b *Board) makeMoveUpdateSide(move *Move) {
	if b.castleAvailable != move.nextCastleRights {
		b.xorCastleKey(b.castleAvailable)
		b.xorCastleKey(move.nextCastleRights)
	}
	if b.enpassant != move.nextEnpassant {
		b.xorEnpassantKey(b.enpassant)
		b.xorEnpassantKey(move.nextEnpassant)
	}
	b.zobristHash ^= b.zobrishTable[768] // toggle side to move
	b.castleAvailable = move.nextCastleRights
	b.enpassant = move.nextEnpassant
	b.moveCount += 1
	b.isWhiteTurn = !b.isWhiteTurn
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

	move.previousZobristHash = b.zobristHash

	//Null move
	if move.isNull {
		b.makeMoveUpdateSide(move)
		return
	}
	//Pawn Promotion
	if move.isPromotion {
		var promotedPiece Piece
		promotedPiece = newPieceTypeColor(move.promotionPieceType, b.currentColor())

		promotedPiece = newPieceTypeColor(move.promotionPieceType, b.currentColor())
		b.xorPieceSquare(move.startSquare.piece, startRow, startCol)
		b.xorPieceSquare(move.endSquare.piece, endRow, endCol)
		b.xorPieceSquare(promotedPiece, endRow, endCol)
		b.board[endRow][endCol] = promotedPiece
		b.board[startRow][startCol] = newPiece('*')
		b.removePieceFromList(startRow, startCol)
		b.setPieceInList(endRow, endCol, promotedPiece)
		b.makeMoveUpdateSide(move)
		return
	}
	//Enpassant
	if move.isEnpassant {
		b.xorPieceSquare(move.startSquare.piece, startRow, startCol)
		b.xorPieceSquare(move.enpassantCapture, startRow, endCol)
		b.xorPieceSquare(move.startSquare.piece, endRow, endCol)
		b.board[endRow][endCol] = move.startSquare.piece
		b.board[startRow][endCol] = newPiece('*')
		b.board[startRow][startCol] = newPiece('*')
		b.removePieceFromList(startRow, startCol)
		b.removePieceFromList(startRow, endCol)
		b.setPieceInList(endRow, endCol, move.startSquare.piece)
		b.makeMoveUpdateSide(move)
		return
	}

	//Castling
	if move.isCastleKingSide {
		rookPiece := b.board[startRow][7]
		b.xorPieceSquare(move.startSquare.piece, startRow, startCol)
		b.xorPieceSquare(rookPiece, startRow, 7)
		b.xorPieceSquare(move.startSquare.piece, startRow, 6)
		b.xorPieceSquare(rookPiece, startRow, 5)
		b.board[startRow][6] = move.startSquare.piece
		b.board[startRow][5] = b.board[startRow][7]
		b.board[startRow][startCol] = newPiece('*')
		b.board[startRow][7] = newPiece('*')
		b.removePieceFromList(startRow, startCol)
		b.removePieceFromList(startRow, 7)
		b.setPieceInList(startRow, 6, move.startSquare.piece)
		b.setPieceInList(startRow, 5, b.board[startRow][5])
		b.makeMoveUpdateSide(move)
		return
	} else if move.isCastleQueenSide {
		rookPiece := b.board[startRow][0]
		b.xorPieceSquare(move.startSquare.piece, startRow, startCol)
		b.xorPieceSquare(rookPiece, startRow, 0)
		b.xorPieceSquare(move.startSquare.piece, startRow, 2)
		b.xorPieceSquare(rookPiece, startRow, 3)
		b.board[startRow][2] = move.startSquare.piece
		b.board[startRow][3] = b.board[startRow][0]
		b.board[startRow][startCol] = newPiece('*')
		b.board[startRow][0] = newPiece('*')
		b.removePieceFromList(startRow, startCol)
		b.removePieceFromList(startRow, 0)
		b.setPieceInList(startRow, 2, move.startSquare.piece)
		b.setPieceInList(startRow, 3, b.board[startRow][3])
		b.makeMoveUpdateSide(move)
		return
	}

	//Normal Move
	b.xorPieceSquare(move.startSquare.piece, startRow, startCol)
	b.xorPieceSquare(move.endSquare.piece, endRow, endCol)
	b.xorPieceSquare(move.startSquare.piece, endRow, endCol)
	b.board[startRow][startCol] = newPiece('*')
	b.board[endRow][endCol] = move.startSquare.piece
	b.removePieceFromList(startRow, startCol)
	b.setPieceInList(endRow, endCol, move.startSquare.piece)
	b.makeMoveUpdateSide(move)
}

func (b *Board) unmakeMoveUpdateSide(move *Move) {
	b.castleAvailable = move.previousCastleRights
	b.enpassant = move.previousEnpassant
	b.moveCount -= 1
	b.isWhiteTurn = !b.isWhiteTurn
}

func (b *Board) unmakeMove(move *Move) {
	b.zobristHash = move.previousZobristHash

	startRow := move.startSquare.row
	startCol := move.startSquare.col
	endRow := move.endSquare.row
	endCol := move.endSquare.col

	//Null move
	if move.isNull {
		b.unmakeMoveUpdateSide(move)
		return
	}
	//Enpassant
	if move.isPromotion {
		b.board[startRow][startCol] = move.startSquare.piece
		b.board[endRow][endCol] = move.endSquare.piece
		b.setPieceInList(endRow, endCol, move.endSquare.piece)
		b.setPieceInList(startRow, startCol, move.startSquare.piece)
		b.unmakeMoveUpdateSide(move)
		return
	}

	if move.isEnpassant {
		b.board[startRow][startCol] = move.startSquare.piece
		b.board[endRow][endCol] = newPiece('*')
		b.board[startRow][endCol] = move.enpassantCapture
		b.setPieceInList(startRow, startCol, move.startSquare.piece)
		b.setPieceInList(endRow, endCol, newPiece('*'))
		b.setPieceInList(startRow, endCol, move.enpassantCapture)
		b.unmakeMoveUpdateSide(move)
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
		b.unmakeMoveUpdateSide(move)
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
		b.unmakeMoveUpdateSide(move)
		return
	}
	b.board[startRow][startCol] = move.startSquare.piece
	b.board[endRow][endCol] = move.endSquare.piece
	b.setPieceInList(endRow, endCol, move.endSquare.piece)
	b.setPieceInList(startRow, startCol, move.startSquare.piece)
	b.unmakeMoveUpdateSide(move)
}

func (b *Board) currentColor() Color {
	if b.isWhiteTurn {
		return Color(White)
	}
	return Color(Black)
}
func (b *Board) playerInCheck() bool {
	color := b.currentColor()
	moveGenerator := MoveGenerator{board: b}
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

func (b *Board) zobrishHashTable() [781]int64 {
	hashTable := [781]int64{}

	seed1 := uint64(123)
	seed2 := uint64(456)
	r := rand.New(rand.NewPCG(seed1, seed2))

	for i := range 781 {
		hashTable[i] = r.Int64()
	}
	return hashTable
}

func calculateZobristIndex(piece Piece, row int, col int) int {
	index := 0
	index += (int(piece) - int(WhitePawn)) * 64
	index += row * 8
	index += col
	return index

}
func (b *Board) xorPieceSquare(piece Piece, row, col int) {
	if !isEmpty(piece) {
		b.zobristHash ^= b.zobrishTable[calculateZobristIndex(piece, row, col)]
	}
}

func (b *Board) xorCastleKey(castle uint8) {
	if castle&CastleWK != 0 {
		b.zobristHash ^= b.zobrishTable[769]
	}
	if castle&CastleWQ != 0 {
		b.zobristHash ^= b.zobrishTable[770]
	}
	if castle&CastleBK != 0 {
		b.zobristHash ^= b.zobrishTable[771]
	}
	if castle&CastleBQ != 0 {
		b.zobristHash ^= b.zobrishTable[772]
	}
}

func (b *Board) xorEnpassantKey(ep string) {
	if ep != "-" {
		b.zobristHash ^= b.zobrishTable[733+int(ep[0]-'a')]
	}
}

func (b *Board) calculateZobrishHash() int64 {
	var hash int64
	hash = 0
	for _, square := range b.piecesGenerator() {
		index := calculateZobristIndex(square.piece, square.row, square.col)
		hash = hash ^ b.zobrishTable[index]
	}
	if !b.isWhiteTurn {
		hash = hash ^ b.zobrishTable[768]
	}

	if b.castleAvailable&CastleWK != 0 {
		hash = hash ^ b.zobrishTable[769]
	}
	if b.castleAvailable&CastleWQ != 0 {
		hash = hash ^ b.zobrishTable[770]
	}
	if b.castleAvailable&CastleBK != 0 {
		hash = hash ^ b.zobrishTable[771]
	}
	if b.castleAvailable&CastleBQ != 0 {
		hash = hash ^ b.zobrishTable[772]
	}

	if b.enpassant != "-" {
		enpassant_idx := 733
		enpassant_idx += int(b.enpassant[0]) - int('a')
		hash = hash ^ b.zobrishTable[enpassant_idx]

	}

	return hash
}
