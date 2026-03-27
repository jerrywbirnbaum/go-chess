package main

import (
	"fmt"
	"math/rand/v2"
	"strconv"
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
	board                 [64]Piece
	pieces                [32]Square
	pieceCount            int
	isWhiteTurn           bool
	enpassant             string
	castleAvailable       uint8
	moveCount             int
	zobrishTable          [781]int64
	zobristHash           int64
	repititionTable       *RepititionTable
	isThreeFoldRepitition bool
	whiteKingRow          int
	whiteKingCol          int
	blackKingRow          int
	blackKingCol          int
	fullMoveClock         int
	halfMoveClock         int
	bitboards             [12]uint64
	colorBitboards        [2]uint64
}

func (b *Board) isNoPawnEndGame() bool {
	pieces := [2]Piece{WhitePawn, BlackPawn}
	for i := range pieces {
		if b.getBitboard(pieces[i]) != 0 {
			return false
		}
	}

	return true
}

func (b *Board) isPawnEndgame() bool {
	pieces := [8]Piece{WhiteKnight, WhiteBishop, WhiteRook, WhiteQueen, BlackKnight, BlackBishop, BlackRook, BlackQueen}
	for i := range pieces {
		if b.getBitboard(pieces[i]) != 0 {
			return false
		}
	}

	return true
}

func (b *Board) getBitboard(piece Piece) uint64 {
	return b.bitboards[piece-1]
}

func (b *Board) getColorBitboard(color Color) uint64 {
	return b.colorBitboards[color]
}
func (b *Board) setBitboardPiece(piece Piece, row int, col int) {
	if piece == EmptyPiece {
		return
	}
	b.bitboards[piece-1] = bitboardAddOne(b.bitboards[piece-1], row, col)
	color := getColor(piece)
	b.colorBitboards[color] = bitboardAddOne(b.colorBitboards[color], row, col)

}

func (b *Board) removeBitboardPiece(piece Piece, row int, col int) {
	if piece == EmptyPiece {
		return
	}
	b.bitboards[piece-1] = bitboardRemoveOne(b.bitboards[piece-1], row, col)
	color := getColor(piece)
	b.colorBitboards[color] = bitboardRemoveOne(b.colorBitboards[color], row, col)
}

func (b *Board) getCell(row int, col int) Piece {
	return b.board[row*8+col]
}

func (b *Board) setCell(row int, col int, piece Piece) {
	b.board[row*8+col] = piece
}

func (b *Board) cellEmpty(row int, col int) bool {
	return isEmpty(b.getCell(row, col))
}

func (b *Board) canCapture(row int, col int, color Color) bool {
	if isEmpty(b.getCell(row, col)) {
		return false
	}
	return !sameColor(b.getCell(row, col), color)
}

func (b *Board) printBoard() string {
	s := ""
	for i := range 8 {
		for j := range 8 {
			s += fmt.Sprintf("%q", printPiece(b.getCell(i, j)))
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
	fullMoveClock, _ := strconv.Atoi(fen_list[4])
	b.fullMoveClock = fullMoveClock
	halfMoveClock, _ := strconv.Atoi(fen_list[4])
	b.halfMoveClock = halfMoveClock

	b.updateBoardFEN(board_fen_string)
	b.rebuildPieceList()
	b.zobristHash = b.calculateZobrishHash()
	b.repititionTable = initRepititionTable()
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
					b.setCell(i, j, newPiece('*'))
					j += 1
				}
			} else {
				b.setCell(i, j, newPiece(c))
				j += 1
			}
		}
	}
}

func (b *Board) piecesGenerator() []Square {
	return b.pieces[:b.pieceCount]
}

func (b *Board) clearBitboards() {
	for i := range 12 {
		b.bitboards[i] = 0
	}
	for i := range 2 {
		b.colorBitboards[i] = 0
	}
}
func (b *Board) rebuildPieceList() {
	b.pieceCount = 0
	b.clearBitboards()
	for i := range 8 {
		for j := range 8 {
			p := b.getCell(i, j)
			if isEmpty(p) {
				continue
			}
			b.setBitboardPiece(p, i, j)
			b.pieces[b.pieceCount] = Square{row: i, col: j, piece: p}
			b.pieceCount += 1
			if isKing(pieceType(p)) {
				if isWhite(p) {
					b.whiteKingRow = i
					b.whiteKingCol = j
				} else {
					b.blackKingRow = i
					b.blackKingCol = j
				}
			}
		}
	}
}

func (b *Board) kingPos(color Color) (int, int) {
	if color == White {
		return b.whiteKingRow, b.whiteKingCol
	}
	return b.blackKingRow, b.blackKingCol
}

func (b *Board) updateKingPos(king Piece, row, col int) {
	if isWhite(king) {
		b.whiteKingRow = row
		b.whiteKingCol = col
	} else {
		b.blackKingRow = row
		b.blackKingCol = col
	}
}

func (b *Board) removePieceFromList(piece Piece, row int, col int) {
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
		b.removePieceFromList(piece, row, col)
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
		board: [64]Piece{
			newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'),
			newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'),
			newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'),
			newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'),
			newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'),
			newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'),
			newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'),
			newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'),
		},
		isWhiteTurn: true,
	}
	board.zobrishTable = board.zobrishHashTable()
	board.updateFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	board.repititionTable = initRepititionTable()
	board.isThreeFoldRepitition = false
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
	startPiece := move.startSquare.piece
	endRow := move.endSquare.row
	endCol := move.endSquare.col
	endPiece := move.endSquare.piece

	pieceType := pieceType(startPiece)

	move.previousCastleRights = b.castleAvailable
	move.previousEnpassant = b.enpassant
	move.nextCastleRights = updateCastleRights(b.castleAvailable, move)
	move.nextEnpassant = updateEnpassantSquare(move)
	move.isCastleKingSide = isKing(pieceType) && (endCol-startCol) == 2
	move.isCastleQueenSide = isKing(pieceType) && (endCol-startCol) == -2
	move.isPromotion = isPawn(pieceType) && (endRow == 0 || endRow == 7)
	move.isEnpassant = false
	move.enpassantCapture = newPiece('*')

	if isPawn(pieceType) && move.previousEnpassant != "-" && isEmpty(endPiece) {
		enpassantRow, enpassantCol := fromSquare(move.previousEnpassant)
		if endRow == enpassantRow && endCol == enpassantCol {
			move.isEnpassant = true
			move.enpassantCapture = b.getCell(startRow, enpassantCol)
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
		promotedPiece = newPieceTypeColor(move.promotionPieceType, getColor(startPiece))
		b.xorPieceSquare(startPiece, startRow, startCol)
		b.xorPieceSquare(endPiece, endRow, endCol)
		b.xorPieceSquare(promotedPiece, endRow, endCol)
		b.setCell(endRow, endCol, promotedPiece)
		b.setCell(startRow, startCol, newPiece('*'))

		b.setBitboardPiece(promotedPiece, endRow, endCol)
		b.removeBitboardPiece(startPiece, startRow, startCol)
		b.removeBitboardPiece(endPiece, endRow, endCol)

		b.removePieceFromList(startPiece, startRow, startCol)
		b.setPieceInList(endRow, endCol, promotedPiece)
	} else if move.isEnpassant {
		//enpassant
		b.xorPieceSquare(startPiece, startRow, startCol)
		b.xorPieceSquare(move.enpassantCapture, startRow, endCol)
		b.xorPieceSquare(startPiece, endRow, endCol)

		b.setBitboardPiece(startPiece, endRow, endCol)
		b.removeBitboardPiece(startPiece, startRow, startCol)
		b.removeBitboardPiece(b.getCell(startRow, endCol), startRow, endCol)

		b.setCell(endRow, endCol, startPiece)
		b.setCell(startRow, endCol, newPiece('*'))
		b.setCell(startRow, startCol, newPiece('*'))

		b.removePieceFromList(startPiece, startRow, startCol)
		b.removePieceFromList(endPiece, startRow, endCol)
		b.setPieceInList(endRow, endCol, startPiece)
	} else if move.isCastleKingSide {
		//Castling
		rookPiece := b.getCell(startRow, 7)
		b.xorPieceSquare(startPiece, startRow, startCol)
		b.xorPieceSquare(rookPiece, startRow, 7)
		b.xorPieceSquare(startPiece, startRow, 6)
		b.xorPieceSquare(rookPiece, startRow, 5)
		b.setCell(startRow, 6, startPiece)
		b.setCell(startRow, 5, b.getCell(startRow, 7))
		b.setCell(startRow, startCol, newPiece('*'))
		b.setCell(startRow, 7, newPiece('*'))

		b.setBitboardPiece(startPiece, startRow, 6)
		b.setBitboardPiece(rookPiece, startRow, 5)
		b.removeBitboardPiece(startPiece, startRow, startCol)
		b.removeBitboardPiece(rookPiece, startRow, 7)

		b.removePieceFromList(startPiece, startRow, startCol)
		b.removePieceFromList(newPieceTypeColor(Rook, b.currentColor()), startRow, 7)
		b.setPieceInList(startRow, 6, startPiece)
		b.setPieceInList(startRow, 5, b.getCell(startRow, 5))
		b.updateKingPos(startPiece, startRow, 6)
	} else if move.isCastleQueenSide {
		rookPiece := b.getCell(startRow, 0)
		b.xorPieceSquare(startPiece, startRow, startCol)
		b.xorPieceSquare(rookPiece, startRow, 0)
		b.xorPieceSquare(startPiece, startRow, 2)
		b.xorPieceSquare(rookPiece, startRow, 3)
		b.setCell(startRow, 2, startPiece)
		b.setCell(startRow, 3, b.getCell(startRow, 0))
		b.setCell(startRow, startCol, newPiece('*'))
		b.setCell(startRow, 0, newPiece('*'))

		b.setBitboardPiece(startPiece, startRow, 2)
		b.setBitboardPiece(rookPiece, startRow, 3)
		b.removeBitboardPiece(startPiece, startRow, startCol)
		b.removeBitboardPiece(rookPiece, startRow, 0)

		b.removePieceFromList(startPiece, startRow, startCol)
		b.removePieceFromList(newPieceTypeColor(Rook, b.currentColor()), startRow, 0)
		b.setPieceInList(startRow, 2, startPiece)
		b.setPieceInList(startRow, 3, b.getCell(startRow, 3))
		b.updateKingPos(startPiece, startRow, 2)
	} else {
		//Normal Move
		b.xorPieceSquare(startPiece, startRow, startCol)
		b.xorPieceSquare(endPiece, endRow, endCol)
		b.xorPieceSquare(startPiece, endRow, endCol)
		b.setCell(startRow, startCol, newPiece('*'))
		b.setCell(endRow, endCol, startPiece)

		b.removeBitboardPiece(startPiece, startRow, startCol)
		b.removeBitboardPiece(endPiece, endRow, endCol)
		b.setBitboardPiece(startPiece, endRow, endCol)

		b.removePieceFromList(startPiece, startRow, startCol)
		b.setPieceInList(endRow, endCol, startPiece)
		if isKing(pieceType) {
			b.updateKingPos(startPiece, endRow, endCol)
		}
	}

	b.makeMoveUpdateSide(move)
	b.isThreeFoldRepitition = b.repititionTable.increment(b.zobristHash)

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
	startPiece := move.startSquare.piece
	endRow := move.endSquare.row
	endCol := move.endSquare.col
	endPiece := move.endSquare.piece

	//Null move
	if move.isNull {
		b.unmakeMoveUpdateSide(move)
		return
	}
	//Enpassant
	if move.isPromotion {
		promotionPiece := newPieceTypeColor(move.promotionPieceType, getColor(startPiece))

		b.setBitboardPiece(startPiece, startRow, startCol)
		b.setBitboardPiece(endPiece, endRow, endCol)
		b.removeBitboardPiece(promotionPiece, endRow, endCol)

		b.setCell(startRow, startCol, startPiece)
		b.setCell(endRow, endCol, endPiece)
		b.setPieceInList(endRow, endCol, endPiece)
		b.setPieceInList(startRow, startCol, startPiece)
	} else if move.isEnpassant {
		b.setCell(startRow, startCol, startPiece)
		b.setCell(endRow, endCol, newPiece('*'))
		b.setCell(startRow, endCol, move.enpassantCapture)
		b.setPieceInList(startRow, startCol, startPiece)
		b.setPieceInList(endRow, endCol, newPiece('*'))
		b.setPieceInList(startRow, endCol, move.enpassantCapture)

		b.setBitboardPiece(startPiece, startRow, startCol)
		b.setBitboardPiece(move.enpassantCapture, startRow, endCol)
		b.removeBitboardPiece(startPiece, endRow, endCol)

	} else if move.isCastleKingSide {
		rookPiece := newPieceTypeColor(Rook, oppositeColor(getColor(startPiece)))

		b.setBitboardPiece(startPiece, startRow, startCol)
		b.setBitboardPiece(rookPiece, startRow, 7)
		b.removeBitboardPiece(startPiece, startRow, 6)
		b.removeBitboardPiece(rookPiece, startRow, 5)

		b.setCell(startRow, 4, startPiece)
		b.setCell(startRow, 7, b.getCell(startRow, 5))
		b.setCell(startRow, 5, newPiece('*'))
		b.setCell(startRow, 6, newPiece('*'))
		b.setPieceInList(startRow, 5, newPiece('*'))
		b.setPieceInList(startRow, 6, newPiece('*'))
		b.setPieceInList(startRow, 4, startPiece)
		b.setPieceInList(startRow, 7, b.getCell(startRow, 7))
		b.updateKingPos(startPiece, startRow, 4)
	} else if move.isCastleQueenSide {
		rookPiece := newPieceTypeColor(Rook, oppositeColor(getColor(startPiece)))

		b.setBitboardPiece(startPiece, startRow, startCol)
		b.setBitboardPiece(rookPiece, startRow, 0)
		b.removeBitboardPiece(startPiece, startRow, 3)
		b.removeBitboardPiece(rookPiece, startRow, 2)

		b.setCell(startRow, 4, startPiece)
		b.setCell(startRow, 0, b.getCell(startRow, 3))
		b.setCell(startRow, 2, newPiece('*'))
		b.setCell(startRow, 3, newPiece('*'))
		b.setPieceInList(startRow, 2, newPiece('*'))
		b.setPieceInList(startRow, 3, newPiece('*'))
		b.setPieceInList(startRow, 4, startPiece)
		b.setPieceInList(startRow, 0, b.getCell(startRow, 0))
		b.updateKingPos(startPiece, startRow, 4)
	} else {
		b.setCell(startRow, startCol, startPiece)
		b.setCell(endRow, endCol, endPiece)

		b.setBitboardPiece(endPiece, endRow, endCol)
		b.setBitboardPiece(startPiece, startRow, startCol)
		b.removeBitboardPiece(startPiece, endRow, endCol)

		b.setPieceInList(endRow, endCol, endPiece)
		b.setPieceInList(startRow, startCol, startPiece)
		if isKing(pieceType(startPiece)) {
			b.updateKingPos(startPiece, startRow, startCol)
		}
	}

	b.unmakeMoveUpdateSide(move)
	b.isThreeFoldRepitition = b.repititionTable.decrement(move.previousZobristHash)
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
	attacks, _ := moveGenerator.generateAttacks(oppositeColor(color), false)
	for i := range 8 {
		for j := range 8 {
			piece := b.getCell(i, j)
			pieceType := pieceType(piece)
			if bitboardCheckOne(attacks, i, j) && isKing(pieceType) && getColor(piece) == color {
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
