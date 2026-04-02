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
	pieceListIndex        [64]int8
	pieceCount            int
	isWhiteTurn           bool
	enpassant             uint8
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
	zobristHashTable      []int64
	inCheck               bool
	whiteMGEval           int
	blackMGEval           int
	whiteEGEval           int
	blackEGEval           int
	midGamePhase          int
	// passedPawnScore       int
}

// Clone returns a deep copy of the board suitable for use in a separate search goroutine.
func (b *Board) Clone() *Board {
	clone := *b
	clone.zobristHashTable = append([]int64{}, b.zobristHashTable...)
	repTableCopy := *b.repititionTable
	clone.repititionTable = &repTableCopy
	return &clone
}

func (b *Board) getPrevioiusZHash() int64 {
	return b.zobristHashTable[len(b.zobristHashTable)-1]
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

func (b *Board) allPieceBitboard() uint64 {
	return b.colorBitboards[0] | b.colorBitboards[1]
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
	if en_passant == "-" {
		b.enpassant = 8
	} else {
		b.enpassant = uint8(en_passant[0]) - uint8('a')
	}
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
					b.setCell(i, j, EmptyPiece)
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

func (b *Board) clearEval() {
	b.blackMGEval = 0
	b.blackEGEval = 0
	b.whiteMGEval = 0
	b.whiteEGEval = 0
	b.midGamePhase = 0
}
func (b *Board) rebuildPieceList() {
	b.pieceCount = 0
	b.clearBitboards()
	b.clearEval()
	for i := range b.pieceListIndex {
		b.pieceListIndex[i] = -1
	}
	for i := range 8 {
		for j := range 8 {
			p := b.getCell(i, j)
			if isEmpty(p) {
				continue
			}
			b.setBitboardPiece(p, i, j)
			b.pieceListIndex[i*8+j] = int8(b.pieceCount)
			b.pieces[b.pieceCount] = Square{row: i, col: j, piece: p}
			b.pieceCount++
			b.updateEval(p, i, j, 1)
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
	sq := row*8 + col
	idx := int(b.pieceListIndex[sq])
	if idx < 0 {
		return
	}
	b.pieceListIndex[sq] = -1
	b.pieceCount--
	if idx != b.pieceCount {
		last := b.pieces[b.pieceCount]
		b.pieces[idx] = last
		b.pieceListIndex[last.row*8+last.col] = int8(idx)
	}
	b.pieces[b.pieceCount] = Square{}
}

func (b *Board) setPieceInList(row int, col int, piece Piece) {
	if isEmpty(piece) {
		b.removePieceFromList(piece, row, col)
		return
	}

	sq := row*8 + col
	idx := int(b.pieceListIndex[sq])
	if idx >= 0 {
		b.pieces[idx].piece = piece
		return
	}

	b.pieceListIndex[sq] = int8(b.pieceCount)
	b.pieces[b.pieceCount] = Square{row: row, col: col, piece: piece}
	b.pieceCount++
}

func initBoard() Board {
	board := Board{
		board: [64]Piece{
			EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece,
			EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece,
			EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece,
			EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece,
			EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece,
			EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece,
			EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece,
			EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece, EmptyPiece,
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
	startRow := move.getStartSquare().row
	startCol := move.getStartSquare().col
	endRow := move.getEndSquare().row
	endCol := move.getEndSquare().col
	startPieceType := pieceType(move.getStartSquare().piece)
	endPieceType := pieceType(move.getEndSquare().piece)
	movingWhite := isWhite(move.getStartSquare().piece)

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

func updateEnpassantSquare(move *Move) int {
	startRow := move.getStartSquare().row
	startCol := move.getStartSquare().col
	endRow := move.getEndSquare().row
	startPieceType := pieceType(move.getStartSquare().piece)

	if isPawn(startPieceType) && (endRow-startRow) == 2 {
		return startCol
	}
	if isPawn(startPieceType) && (endRow-startRow) == -2 {
		return startCol
	}
	return 8
}

func (b *Board) makeMoveUpdateSide(move *Move) {
	if b.castleAvailable != move.getNextCastleRights() {
		b.xorCastleKey(b.castleAvailable)
		b.xorCastleKey(move.getNextCastleRights())
	}
	if b.enpassant != move.getNextEnpassant() {
		b.xorEnpassantKey(b.enpassant)
		b.xorEnpassantKey(move.getNextEnpassant())
	}
	b.zobristHash ^= b.zobrishTable[768] // toggle side to move
	b.castleAvailable = move.getNextCastleRights()
	b.enpassant = move.getNextEnpassant()
	b.moveCount += 1
	b.isWhiteTurn = !b.isWhiteTurn

}

func (b *Board) makeMove(move *Move) {
	startRow := move.getStartSquare().row
	startCol := move.getStartSquare().col
	startPiece := move.getStartSquare().piece
	endRow := move.getEndSquare().row
	endCol := move.getEndSquare().col
	endPiece := move.getEndSquare().piece

	pieceType := pieceType(startPiece)

	move.setPreviousCastleRights(b.castleAvailable)
	move.setPreviousEnpassant(b.enpassant)
	move.setNextCastleRights(updateCastleRights(b.castleAvailable, move))
	move.setNextEnpassant(uint8(updateEnpassantSquare(move)))
	move.setIsCastleKingSide(isKing(pieceType) && (endCol-startCol) == 2)
	move.setIsCastleQueenSide(isKing(pieceType) && (endCol-startCol) == -2)
	move.setIsPromotion(isPawn(pieceType) && (endRow == 0 || endRow == 7))
	move.setIsEnpassant(false)
	move.setEnpassantCapture(EmptyPiece)

	if isPawn(pieceType) && move.getPreviousEnpassant() != 8 && isEmpty(endPiece) {
		epCol := int(move.getPreviousEnpassant())
		var epRow int
		if b.isWhiteTurn {
			epRow = 2
		} else {
			epRow = 5
		}
		if endRow == epRow && endCol == epCol {
			move.setIsEnpassant(true)
			move.setEnpassantCapture(b.getCell(startRow, epCol))
		}
	}

	b.zobristHashTable = append(b.zobristHashTable, b.zobristHash)

	//Null move
	if move.getIsNull() {
		b.makeMoveUpdateSide(move)
		return
	}
	//Pawn Promotion
	if move.getIsPromotion() {
		var promotedPiece Piece
		promotedPiece = newPieceTypeColor(move.getPromotionPieceType(), getColor(startPiece))
		b.decrementPiece(endPiece, endRow, endCol, true)
		b.decrementPiece(startPiece, startRow, startCol, true)
		b.incrementPiece(promotedPiece, endRow, endCol, true)
	} else if move.getIsEnpassant() {
		//enpassant
		b.decrementPiece(b.getCell(startRow, endCol), startRow, endCol, true)
		b.decrementPiece(startPiece, startRow, startCol, true)
		b.incrementPiece(startPiece, endRow, endCol, true)
	} else if move.getIsCastleKingSide() {
		//Castling
		rookPiece := b.getCell(startRow, 7)

		b.decrementPiece(rookPiece, startRow, 7, true)
		b.decrementPiece(startPiece, startRow, startCol, true)
		b.incrementPiece(rookPiece, startRow, 5, true)
		b.incrementPiece(startPiece, startRow, 6, true)
		b.updateKingPos(startPiece, startRow, 6)
	} else if move.getIsCastleQueenSide() {
		rookPiece := b.getCell(startRow, 0)

		b.decrementPiece(rookPiece, startRow, 0, true)
		b.decrementPiece(startPiece, startRow, startCol, true)
		b.incrementPiece(rookPiece, startRow, 3, true)
		b.incrementPiece(startPiece, startRow, 2, true)
		b.updateKingPos(startPiece, startRow, 2)
	} else {
		//Normal Move
		b.decrementPiece(endPiece, endRow, endCol, true)
		b.decrementPiece(startPiece, startRow, startCol, true)
		b.incrementPiece(startPiece, endRow, endCol, true)

		if isKing(pieceType) {
			b.updateKingPos(startPiece, endRow, endCol)
		}
	}

	b.makeMoveUpdateSide(move)
	b.isThreeFoldRepitition = b.repititionTable.increment(b.zobristHash)

}

func (b *Board) unmakeMoveUpdateSide(move *Move) {
	b.castleAvailable = move.getPreviousCastleRights()
	b.enpassant = move.getPreviousEnpassant()
	b.moveCount -= 1
	b.isWhiteTurn = !b.isWhiteTurn
}

func (b *Board) unmakeMove(move *Move) {
	startRow := move.getStartSquare().row
	startCol := move.getStartSquare().col
	startPiece := move.getStartSquare().piece
	endRow := move.getEndSquare().row
	endCol := move.getEndSquare().col
	endPiece := move.getEndSquare().piece

	b.isThreeFoldRepitition = b.repititionTable.decrement(b.zobristHash)
	b.zobristHash = b.getPrevioiusZHash()
	b.zobristHashTable = b.zobristHashTable[:len(b.zobristHashTable)-1]

	//Null move
	if move.getIsNull() {
		b.unmakeMoveUpdateSide(move)
		return
	}
	//Enpassant
	if move.getIsPromotion() {
		promotedPiece := newPieceTypeColor(move.getPromotionPieceType(), getColor(startPiece))

		b.decrementPiece(promotedPiece, endRow, endCol, false)
		b.incrementPiece(endPiece, endRow, endCol, false)
		b.incrementPiece(startPiece, startRow, startCol, false)
	} else if move.getIsEnpassant() {

		b.decrementPiece(startPiece, endRow, endCol, false)
		b.incrementPiece(move.getEnpassantCapture(), startRow, endCol, false)
		b.incrementPiece(startPiece, startRow, startCol, false)

	} else if move.getIsCastleKingSide() {
		rookPiece := newPieceTypeColor(Rook, getColor(startPiece))

		b.decrementPiece(rookPiece, startRow, 5, false)
		b.decrementPiece(startPiece, startRow, 6, false)
		b.incrementPiece(rookPiece, startRow, 7, false)
		b.incrementPiece(startPiece, startRow, startCol, false)

		b.updateKingPos(startPiece, startRow, 4)
	} else if move.getIsCastleQueenSide() {
		rookPiece := newPieceTypeColor(Rook, getColor(startPiece))

		b.decrementPiece(rookPiece, startRow, 3, false)
		b.decrementPiece(startPiece, startRow, 2, false)
		b.incrementPiece(rookPiece, startRow, 0, false)
		b.incrementPiece(startPiece, startRow, startCol, false)

		b.updateKingPos(startPiece, startRow, 4)
	} else {

		b.decrementPiece(startPiece, endRow, endCol, false)
		b.incrementPiece(endPiece, endRow, endCol, false)
		b.incrementPiece(startPiece, startRow, startCol, false)

		if isKing(pieceType(startPiece)) {
			b.updateKingPos(startPiece, startRow, startCol)
		}
	}

	b.unmakeMoveUpdateSide(move)
}

func (b *Board) currentColor() Color {
	if b.isWhiteTurn {
		return White
	}
	return Black
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

func (b *Board) xorEnpassantKey(ep uint8) {
	if ep < 8 {
		b.zobristHash ^= b.zobrishTable[733+int(ep)]
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

	if b.enpassant != 8 {
		enpassant_idx := 733
		enpassant_idx += int(b.enpassant)
		hash = hash ^ b.zobrishTable[enpassant_idx]

	}

	return hash
}

func (b *Board) updateEval(piece Piece, row int, col int, sign int) {
	if piece == EmptyPiece {
		return
	}

	color := getColor(piece)
	pieceType := pieceType(piece)
	b.midGamePhase += gamephaseInc[pieceType-1] * sign

	if color == White {
		b.whiteMGEval += getEvalCell(mg_table[piece-1], row, col, false) * sign
		b.whiteEGEval += getEvalCell(eg_table[piece-1], row, col, false) * sign
	} else {
		b.blackMGEval += getEvalCell(mg_table[piece-1], row, col, false) * sign
		b.blackEGEval += getEvalCell(eg_table[piece-1], row, col, false) * sign
	}
}

func (b *Board) incrementPiece(piece Piece, row int, col int, updateHash bool) {
	if updateHash {
		b.xorPieceSquare(piece, row, col)
	}
	b.setCell(row, col, piece)
	b.setBitboardPiece(piece, row, col)
	b.setPieceInList(row, col, piece)
	b.updateEval(piece, row, col, 1)
}

func (b *Board) decrementPiece(piece Piece, row int, col int, updateHash bool) {
	if updateHash {
		b.xorPieceSquare(piece, row, col)
	}
	b.setCell(row, col, EmptyPiece)
	b.removeBitboardPiece(piece, row, col)
	b.removePieceFromList(piece, row, col)
	b.updateEval(piece, row, col, -1)
}
