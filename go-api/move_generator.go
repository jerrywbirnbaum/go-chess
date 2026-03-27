package main

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"
)

type Move struct {
	startSquare          Square
	endSquare            Square
	previousEnpassant    string
	previousCastleRights uint8
	nextEnpassant        string
	nextCastleRights     uint8
	isEnpassant          bool
	isCastleKingSide     bool
	isCastleQueenSide    bool
	isPromotion          bool
	promotionPieceType   PieceType
	isNull               bool
	enpassantCapture     Piece
	previousZobristHash  int64
}

var emptyBitboard uint64 = 0
var fullBitboard uint64 = math.MaxUint64

var straightDirs = [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
var diagonalDirs = [][2]int{{1, 1}, {-1, -1}, {-1, 1}, {1, -1}}
var allDirs = [][2]int{{1, 1}, {-1, -1}, {-1, 1}, {1, -1}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}}

var knightOffsets = [8][2]int{{1, 2}, {2, 1}, {-1, -2}, {-2, -1}, {2, -1}, {-2, 1}, {-1, 2}, {1, -2}}
var kingOffsets = [8][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}}

var knightAttacks [64]uint64
var kingAttacks [64]uint64
var bishopMasks [64]uint64
var rookMasks [64]uint64

var row4, _ = strconv.ParseUint("00000000FF000000", 16, 64)
var row5, _ = strconv.ParseUint("000000FF00000000", 16, 64)

var colA, _ = strconv.ParseUint("8080808080808080", 16, 64)
var colH, _ = strconv.ParseUint("0101010101010101", 16, 64)

var bishopMagicLookup [64][512]uint64
var rookMagicLookup [64][4096]uint64

func init() {
	for sq := range 64 {
		r, c := sq/8, sq%8
		knightAttacks[sq] = leaperAttackBits(r, c, knightOffsets[:])
		kingAttacks[sq] = leaperAttackBits(r, c, kingOffsets[:])
		bishopMasks[sq] = sliderMaskBits(r, c, diagonalDirs[:])
		rookMasks[sq] = sliderMaskBits(r, c, straightDirs[:])
	}
	bishopMagicLookup = createBishopLookupTable()
	rookMagicLookup = createRookLookupTable()
}

type MoveString struct {
	startSquare string
	endSquare   string
	promotion   string
	isPromotion bool
}

type MoveGenerator struct {
	board           *Board
	repititionTable *RepititionTable
}

func (mg *MoveGenerator) updateBoard(board *Board) {
	mg.board = board
}

func (mg *MoveGenerator) generateAttacks(color Color, slidingOnly bool) (uint64, int) {
	attacks := emptyBitboard
	checkers := 0

	opp := oppositeColor(color)
	kingRow, kingCol := mg.board.kingPos(opp)
	mg.board.setCell(kingRow, kingCol, EmptyPiece)

	for _, p := range mg.board.piecesGenerator() {
		if getColor(p.piece) != color {
			continue
		}
		pt := pieceType(p.piece)
		var pieceAttacks uint64
		if !slidingOnly {
			switch {
			case isPawn(pt):
				pieceAttacks = pawnAttackBits(p.row, p.col, color)
			case isKnight(pt):
				pieceAttacks = knightAttacks[p.row*8+p.col]
			case isKing(pt):
				pieceAttacks = kingAttacks[p.row*8+p.col]
			}
		}
		if isSlidingPiece(pt) {
			pieceAttacks |= mg.slidingAttackBits(p.row, p.col, pt)
		}
		attacks |= pieceAttacks
		if bitboardCheckOne(pieceAttacks, kingRow, kingCol) {
			checkers++
		}
	}

	mg.board.setCell(kingRow, kingCol, newPieceTypeColor(PieceType(King), opp))
	return attacks, checkers
}

func pawnAttackBits(row int, col int, color Color) uint64 {
	pawnBitboard := bitboardAddOne(emptyBitboard, row, col)
	attacksBitboard := uint64(0)
	if color == White {
		attacksBitboard |= (pawnBitboard << 7) & ^colA
		attacksBitboard |= (pawnBitboard << 9) & ^colH
	} else {
		attacksBitboard |= (pawnBitboard >> 7) & ^colH
		attacksBitboard |= (pawnBitboard >> 9) & ^colA
	}
	return attacksBitboard
}

func sliderMaskBits(row int, col int, dirs [][2]int) uint64 {
	var bits uint64
	for _, dir := range dirs {
		currentRow := row + dir[0]
		currentCol := col + dir[1]
		for inBounds(currentRow+dir[0], currentCol+dir[1]) {
			bits = bitboardAddOne(bits, currentRow, currentCol)
			currentRow += dir[0]
			currentCol += dir[1]

		}
	}
	return bits
}

func leaperAttackBits(row, col int, offsets [][2]int) uint64 {
	var bits uint64
	for _, off := range offsets {
		if inBounds(row+off[0], col+off[1]) {
			bits = bitboardAddOne(bits, row+off[0], col+off[1])
		}
	}
	return bits
}

func (mg *MoveGenerator) slidingAttackBits(row, col int, pt PieceType) uint64 {
	dirs := allDirs
	if isRook(pt) {
		dirs = straightDirs
	} else if isBishop(pt) {
		dirs = diagonalDirs
	}
	var bits uint64
	for _, dir := range dirs {
		r, c := row+dir[0], col+dir[1]
		for inBounds(r, c) {
			bits = bitboardAddOne(bits, r, c)
			if !mg.board.cellEmpty(r, c) {
				break
			}
			r += dir[0]
			c += dir[1]
		}
	}
	return bits
}

func (mg *MoveGenerator) pinnedPieces(kingRow int, kingCol int) uint64 {
	pinnedPieces := emptyBitboard
	piece := mg.board.getCell(kingRow, kingCol)
	color := getColor(piece)

	enemyQueen := newPieceTypeColor(Queen, oppositeColor(color))
	queenBitboard := mg.board.getBitboard(enemyQueen)
	enemyBishop := newPieceTypeColor(Bishop, oppositeColor(color))
	bishopBitboard := mg.board.getBitboard(enemyBishop)
	enemyRook := newPieceTypeColor(Rook, oppositeColor(color))
	rookBitboard := mg.board.getBitboard(enemyRook)

	for slidingIdx, move := range allDirs {
		isDiagonal := slidingIdx < 4
		row := kingRow + move[0]
		col := kingCol + move[1]
		foundFriendly := false
		var candidate Square

		for inBounds(row, col) {
			if !bitboardCheckOne(mg.board.allPieceBitboard(), row, col) {
				row += move[0]
				col += move[1]
				continue
			}

			if bitboardCheckOne(mg.board.getColorBitboard(color), row, col) {
				if foundFriendly {
					break
				}
				foundFriendly = true
				candidate = Square{row: row, col: col}
				row += move[0]
				col += move[1]
				continue
			}

			validSlider := false
			if isDiagonal && (bitboardCheckOne(queenBitboard, row, col) || bitboardCheckOne(bishopBitboard, row, col)) {
				validSlider = true
			}
			if !isDiagonal && (bitboardCheckOne(queenBitboard, row, col) || bitboardCheckOne(rookBitboard, row, col)) {
				validSlider = true
			}
			if foundFriendly && validSlider {
				pinnedPieces = bitboardAddOne(pinnedPieces, candidate.row, candidate.col)
			}
			break
		}
	}

	return pinnedPieces
}
func (mg *MoveGenerator) checkRays(kingRow int, kingCol int) uint64 {
	checkBitboard := emptyBitboard

	color := mg.board.currentColor()
	oppositeColor := oppositeColor(color)

	// check pawns
	pawnBitboard := pawnAttackBits(kingRow, kingCol, color)
	pawnBitboard &= mg.board.getBitboard(newPieceTypeColor(Pawn, oppositeColor))
	checkBitboard |= pawnBitboard

	knightBitboard := knightAttacks[kingRow*8+kingCol]
	knightBitboard &= mg.board.getBitboard(newPieceTypeColor(Knight, oppositeColor))
	checkBitboard |= knightBitboard

	// check sliding
	checkBitboard = mg.slidingRays(kingRow, kingCol, color, checkBitboard, false)

	return checkBitboard
}

func (mg *MoveGenerator) slidingRays(kingRow int, kingCol int, color Color, checkBitboard uint64, isPinRay bool) uint64 {

	for slidingIdx, move := range allDirs {
		isDiagonal := slidingIdx < 4
		row := kingRow + move[0]
		col := kingCol + move[1]
		for i := range 7 {
			_ = i
			if row < 0 || row > 7 || col < 0 || col > 7 {
				break
			}

			pieceType := pieceType(mg.board.getCell(row, col))

			if !mg.board.cellEmpty(row, col) && !mg.board.canCapture(row, col, color) {
				break
			}

			if !mg.board.cellEmpty(row, col) && !isSlidingPiece(pieceType) {
				break
			}
			if mg.board.canCapture(row, col, color) && (isSlidingPiece(pieceType) || isPinRay) {
				if !isPinRay && isDiagonal && isRook(pieceType) {
					break
				}

				if !isPinRay && !isDiagonal && isBishop(pieceType) {
					break
				}

				checkBitboard = bitboardAddOne(checkBitboard, row, col)
				for j := range 7 {
					_ = j
					row -= move[0]
					col -= move[1]
					if row == kingRow && col == kingCol {
						if !isPinRay {
							return checkBitboard
						} else {
							break
						}
					}
					checkBitboard = bitboardAddOne(checkBitboard, row, col)
				}
			}

			row += move[0]
			col += move[1]
		}
	}
	return checkBitboard
}
func (mg *MoveGenerator) generateMoves(onlyCaptures bool) []Move {
	moves := make([]Move, 0, 150)
	color := mg.board.currentColor()

	oppositeColor := oppositeColor(color)
	attackedSquares, checkers := mg.generateAttacks(oppositeColor, false)

	kingRow, kingCol := mg.board.kingPos(color)

	//Double check
	var checkBitboard uint64
	if bitboardCheckOne(attackedSquares, kingRow, kingCol) && checkers > 1 {
		piece := Square{row: kingRow, col: kingCol, piece: mg.board.getCell(kingRow, kingCol)}
		moves = mg.generateKingMoves(piece, color, attackedSquares, onlyCaptures, moves)
		return moves
	} else if bitboardCheckOne(attackedSquares, kingRow, kingCol) {
		checkBitboard = mg.checkRays(kingRow, kingCol)
	} else {
		checkBitboard = fullBitboard
	}

	pieces := mg.board.piecesGenerator()
	pinnedPieces := mg.pinnedPieces(kingRow, kingCol)
	for _, p := range pieces {
		if getColor(p.piece) != color {
			continue
		}
		if bitboardCheckOne(pinnedPieces, p.row, p.col) {
			moves = append(moves, mg.generatePinnedMoves(p, color, kingRow, kingCol, checkBitboard, onlyCaptures)...)
			continue
		}

		pieceType := pieceType(p.piece)
		if isPawn(pieceType) {
			moves = mg.generatePawnMoves(p, color, checkBitboard, onlyCaptures, moves)
		}

		if isKnight(pieceType) {
			moves = mg.generateKnightMoves(p, color, checkBitboard, onlyCaptures, moves)
		}

		if isSlidingPiece(pieceType) {
			moves = mg.generateSlidingMoves(p, color, checkBitboard, onlyCaptures, moves)
		}

		if isKing(pieceType) {
			moves = mg.generateKingMoves(p, color, attackedSquares, onlyCaptures, moves)
		}
	}

	if !onlyCaptures {
		moves = append(moves, mg.generateCastles(color, attackedSquares)...)
	}

	return moves
}

func (mg *MoveGenerator) generateCastles(color Color, checkBitboard uint64) []Move {
	moves := []Move{}
	availability := mg.board.castleAvailable

	if color == White && availability&CastleWK != 0 {
		if !bitboardCheckOne(checkBitboard, 7, 4) && !bitboardCheckOne(checkBitboard, 7, 5) && !bitboardCheckOne(checkBitboard, 7, 6) {
			if mg.board.cellEmpty(7, 5) && mg.board.cellEmpty(7, 6) && mg.board.getCell(7, 7) == newPiece('R') && mg.board.getCell(7, 4) == newPiece('K') {
				startSquare := Square{row: 7, col: 4, piece: newPiece('K')}
				endSquare := Square{row: 7, col: 6, piece: EmptyPiece}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	if color == White && availability&CastleWQ != 0 {

		if !bitboardCheckOne(checkBitboard, 7, 4) && !bitboardCheckOne(checkBitboard, 7, 3) && !bitboardCheckOne(checkBitboard, 7, 2) {
			if mg.board.cellEmpty(7, 1) && mg.board.cellEmpty(7, 2) && mg.board.cellEmpty(7, 3) && mg.board.getCell(7, 0) == newPiece('R') && mg.board.getCell(7, 4) == newPiece('K') {
				startSquare := Square{row: 7, col: 4, piece: newPiece('K')}
				endSquare := Square{row: 7, col: 2, piece: EmptyPiece}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	if color == Black && availability&CastleBK != 0 {

		if !bitboardCheckOne(checkBitboard, 0, 4) && !bitboardCheckOne(checkBitboard, 0, 5) && !bitboardCheckOne(checkBitboard, 0, 6) {
			if mg.board.cellEmpty(0, 5) && mg.board.cellEmpty(0, 6) && mg.board.getCell(0, 7) == newPiece('r') && mg.board.getCell(0, 4) == newPiece('k') {
				startSquare := Square{row: 0, col: 4, piece: newPiece('k')}
				endSquare := Square{row: 0, col: 6, piece: EmptyPiece}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	if color == Black && availability&CastleBQ != 0 {
		if !bitboardCheckOne(checkBitboard, 0, 4) && !bitboardCheckOne(checkBitboard, 0, 3) && !bitboardCheckOne(checkBitboard, 0, 2) {
			if mg.board.cellEmpty(0, 1) && mg.board.cellEmpty(0, 2) && mg.board.cellEmpty(0, 3) && mg.board.getCell(0, 0) == newPiece('r') && mg.board.getCell(0, 4) == newPiece('k') {
				startSquare := Square{row: 0, col: 4, piece: newPiece('k')}
				endSquare := Square{row: 0, col: 2, piece: EmptyPiece}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	return moves
}
func (mg *MoveGenerator) pinDirection(kingRow int, kingCol int, row int, col int) ([2]int, bool) {

	if kingRow == row {
		if col > kingCol {
			return allDirs[6], false
		} else if col < kingCol {
			return allDirs[7], false
		}
	} else if kingCol == col {
		if row > kingRow {
			return allDirs[4], false
		} else if row < kingRow {
			return allDirs[5], false
		}
	} else if kingRow > row {
		if col > kingCol {
			return allDirs[2], true
		} else if col < kingCol {
			return allDirs[1], true
		}
	} else if kingRow < row {
		if col > kingCol {
			return allDirs[0], true
		} else if col < kingCol {
			return allDirs[3], true
		}
	}
	return allDirs[0], true

}
func (mg *MoveGenerator) generatePinnedMoves(p Square, color Color, kingRow int, kingCol int, checkBitboard uint64, onlyCaptures bool) []Move {
	moves := []Move{}
	row := p.row
	col := p.col
	currentPieceType := pieceType(p.piece)
	if isKnight(currentPieceType) {
		return moves
	}
	direction, isDiagonal := mg.pinDirection(kingRow, kingCol, row, col)

	if !isPawn(currentPieceType) && isDiagonal && !isDiagonalSlidingPiece(currentPieceType) {
		return moves
	}

	if !isPawn(currentPieceType) && !isDiagonal && !isStriaghtSlidingPiece(currentPieceType) {
		return moves
	}

	currentRow := row
	currentCol := col
	pinnedBitboard := emptyBitboard

	for i := range 7 {
		_ = i
		currentRow += direction[0]
		currentCol += direction[1]
		if currentRow < 0 || currentRow > 7 || currentCol < 0 || currentCol > 7 {
			break
		}
		if mg.board.canCapture(currentRow, currentCol, color) {
			pinnedBitboard = bitboardAddOne(pinnedBitboard, currentRow, currentCol)
			break
		}
		pinnedBitboard = bitboardAddOne(pinnedBitboard, currentRow, currentCol)
	}

	currentRow = row
	currentCol = col

	for i := range 7 {
		_ = i
		currentRow -= direction[0]
		currentCol -= direction[1]
		if currentRow < 0 || currentRow > 7 || currentCol < 0 || currentCol > 7 {
			break
		}

		piece := mg.board.getCell(currentRow, currentCol)
		pieceType := pieceType(piece)
		if isKing(pieceType) {
			break
		}
		pinnedBitboard = bitboardAddOne(pinnedBitboard, currentRow, currentCol)

	}

	pinnedBitboard &= checkBitboard

	if isPawn(currentPieceType) {
		moves = mg.generatePawnMoves(p, color, pinnedBitboard, onlyCaptures, moves)
	}

	if isSlidingPiece(currentPieceType) {
		moves = mg.generateSlidingMoves(p, color, pinnedBitboard, onlyCaptures, moves)
	}

	return moves
}
func (mg *MoveGenerator) generateSlidingMoves(p Square, color Color, checkBitboard uint64, onlyCaptures bool, moves []Move) []Move {
	sameColorBitboard := mg.board.getColorBitboard(color)
	oppositeColorBitboard := mg.board.getColorBitboard(oppositeColor(color))

	currentRow := p.row
	currentCol := p.col
	startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}

	sq := squareFromRowCol(currentRow, currentCol)
	allPieces := mg.board.allPieceBitboard()
	var attacksBitboard uint64
	pt := pieceType(p.piece)
	if pt == Bishop {
		blockers := bishopMasks[sq] & allPieces
		magicIndex := (reverseColBits(blockers) * getBishopMagicNumber(currentRow, currentCol)) >> getBishopShift(currentRow, currentCol)
		attacksBitboard = bishopMagicLookup[sq][magicIndex]
	} else if pt == Rook {
		blockers := rookMasks[sq] & allPieces
		magicIndex := (reverseColBits(blockers) * getRookMagicNumber(currentRow, currentCol)) >> getRookShift(currentRow, currentCol)
		attacksBitboard = rookMagicLookup[sq][magicIndex]
	} else if pt == Queen {
		bishopBlockers := bishopMasks[sq] & allPieces
		bishopIndex := (reverseColBits(bishopBlockers) * getBishopMagicNumber(currentRow, currentCol)) >> getBishopShift(currentRow, currentCol)
		rookBlockers := rookMasks[sq] & allPieces
		rookIndex := (reverseColBits(rookBlockers) * getRookMagicNumber(currentRow, currentCol)) >> getRookShift(currentRow, currentCol)
		attacksBitboard = bishopMagicLookup[sq][bishopIndex] | rookMagicLookup[sq][rookIndex]
	}

	attacksBitboard &= ^sameColorBitboard
	attacksBitboard &= checkBitboard

	if onlyCaptures {
		attacksBitboard &= oppositeColorBitboard
	}

	for attacksBitboard != 0 {
		attackIdx := bitScanForward(attacksBitboard)
		attacksBitboard ^= 1 << attackIdx
		endRow, endCol := rowColFromSquare(63 - attackIdx)
		endSquare := Square{row: endRow, col: endCol, piece: mg.board.getCell(endRow, endCol)}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}
	return moves
}

func (mg *MoveGenerator) generateKingMoves(p Square, color Color, attackBitboard uint64, onlyCaptures bool, moves []Move) []Move {
	currentRow := p.row
	currentCol := p.col

	startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
	pieceAttacks := kingAttacks[currentRow*8+currentCol]
	pieceAttacks &= ^attackBitboard
	pieceAttacks &= ^mg.board.getColorBitboard(color)

	if onlyCaptures {
		pieceAttacks &= mg.board.getColorBitboard(oppositeColor(color))
	}
	for pieceAttacks != 0 {
		attackIdx := bitScanForward(pieceAttacks)
		pieceAttacks ^= 1 << attackIdx
		endRow, endCol := rowColFromSquare(63 - attackIdx)
		endSquare := Square{row: endRow, col: endCol, piece: mg.board.getCell(endRow, endCol)}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}

	return moves

}
func (mg *MoveGenerator) generateKnightMoves(p Square, color Color, checkBitboard uint64, onlyCaptures bool, moves []Move) []Move {

	currentRow := p.row
	currentCol := p.col

	startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
	pieceAttacks := knightAttacks[currentRow*8+currentCol]
	pieceAttacks &= ^mg.board.getColorBitboard(color)
	pieceAttacks &= checkBitboard
	if onlyCaptures {
		pieceAttacks &= mg.board.getColorBitboard(oppositeColor(color))
	}
	for pieceAttacks != 0 {
		attackIdx := bitScanForward(pieceAttacks)
		pieceAttacks ^= 1 << attackIdx
		endRow, endCol := rowColFromSquare(63 - attackIdx)
		endSquare := Square{row: endRow, col: endCol, piece: mg.board.getCell(endRow, endCol)}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}

	return moves
}

func (mg *MoveGenerator) generatePawnMoves(p Square, color Color, checkBitboard uint64, onlyCaptures bool, moves []Move) []Move {
	currentRow := p.row
	currentCol := p.col
	pawnBitboard := bitboardAddOne(emptyBitboard, currentRow, currentCol)
	startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}

	enpassantRow := 4
	if color == White {
		enpassantRow = 3
	}

	allPieceBitboard := mg.board.allPieceBitboard()
	if !onlyCaptures {
		singlePushBitboard := pawnBitboard
		if color == White {
			singlePushBitboard <<= 8
			singlePushBitboard &= ^(allPieceBitboard)
		} else {
			singlePushBitboard >>= 8
			singlePushBitboard &= ^(allPieceBitboard)
		}
		singlePushBitboard &= checkBitboard
		if singlePushBitboard != 0 {
			attackIdx := bitScanForward(singlePushBitboard)
			endRow, endCol := rowColFromSquare(63 - attackIdx)
			endSquare := Square{row: endRow, col: endCol, piece: EmptyPiece}

			if endRow == 0 || endRow == 7 {
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: Queen})
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: Rook})
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: Bishop})
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: Knight})
			} else {
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	//Double Pawn Push
	var doublePushRow uint64

	if color == White {
		doublePushRow = row4
	} else {
		doublePushRow = row5
	}

	if !onlyCaptures {
		doublePushBitboard := pawnBitboard
		if color == White {
			doublePushBitboard <<= 16
			doublePushBitboard &= doublePushRow
			doublePushBitboard &= ^(allPieceBitboard << 8)
			doublePushBitboard &= ^(allPieceBitboard)
		} else {
			doublePushBitboard >>= 16
			doublePushBitboard &= doublePushRow
			doublePushBitboard &= ^(allPieceBitboard >> 8)
			doublePushBitboard &= ^(allPieceBitboard)
		}
		doublePushBitboard &= checkBitboard
		if doublePushBitboard != 0 {
			attackIdx := bitScanForward(doublePushBitboard)
			endRow, endCol := rowColFromSquare(63 - attackIdx)
			endSquare := Square{row: endRow, col: endCol, piece: EmptyPiece}
			moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
		}
	}

	//Capture Moves
	captureBitboard := emptyBitboard
	if color == White {
		captureBitboard |= (pawnBitboard << 7) & ^colA
		captureBitboard |= (pawnBitboard << 9) & ^colH
	} else {
		captureBitboard |= (pawnBitboard >> 7) & ^colH
		captureBitboard |= (pawnBitboard >> 9) & ^colA
	}

	captureBitboard &= checkBitboard
	captureBitboard &= mg.board.getColorBitboard(oppositeColor(color))

	for captureBitboard != 0 {
		attackIdx := bitScanForward(captureBitboard)
		captureBitboard ^= 1 << attackIdx
		endRow, endCol := rowColFromSquare(63 - attackIdx)
		endSquare := Square{row: endRow, col: endCol, piece: mg.board.getCell(endRow, endCol)}
		if endRow == 0 || endRow == 7 {
			moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: Queen})
			moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: Rook})
			moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: Bishop})
			moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: Knight})
		} else {
			moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
		}

	}

	//ENPASSANT
	if mg.board.enpassant != "-" {
		ep_row, ep_col := fromSquare(mg.board.enpassant)
		if p.row == enpassantRow && (ep_col-p.col == 1 || ep_col-p.col == -1) {
			startSquare := Square{row: p.row, col: p.col, piece: p.piece}
			endSquare := Square{row: ep_row, col: ep_col, piece: mg.board.getCell(ep_row, ep_col)}
			enpassantMove := Move{startSquare: startSquare, endSquare: endSquare}
			if !mg.enpassantCheck(enpassantMove, color) && (mg.board.cellEmpty(ep_row, ep_col)) {
				moves = append(moves, enpassantMove)
			}
		}
	}

	return moves
}

func (mg *MoveGenerator) enpassantCheck(move Move, color Color) bool {
	simulatedBoard := *mg.board
	simulatedBoard.makeMove(&move)

	simulatedMoveGenerator := MoveGenerator{board: &simulatedBoard}
	attacks, _ := simulatedMoveGenerator.generateAttacks(oppositeColor(color), true)

	kingRow, kingCol := mg.board.kingPos(color)

	inCheck := bitboardCheckOne(attacks, kingRow, kingCol)

	simulatedBoard.unmakeMove(&move)
	return inCheck
}

func toSquare(row int, col int) string {
	return fmt.Sprintf("%c%d", 'a'+col, 8-row)
}
func fromSquare(square string) (int, int) {
	row := 8 - int(square[1]-'0')
	col := int(square[0] - 'a')
	return row, col
}

func (mg *MoveGenerator) randomMove() MoveString {
	moves := mg.generateMoves(false)

	seed := rand.NewSource(time.Now().Unix())
	r := rand.New(seed)

	random_index := r.Intn(len(moves))
	random_move := moves[random_index]
	startSquare := toSquare(random_move.startSquare.row, random_move.startSquare.col)
	endSquare := toSquare(random_move.endSquare.row, random_move.endSquare.col)
	return MoveString{startSquare: startSquare, endSquare: endSquare}
}
