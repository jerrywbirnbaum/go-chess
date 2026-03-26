package main

import (
	"fmt"
	"math"
	"math/rand"
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

var emptyMask uint64 = 0
var fullMask uint64 = math.MaxUint64

var straightDirs = [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
var diagonalDirs = [][2]int{{1, 1}, {-1, -1}, {-1, 1}, {1, -1}}
var allDirs = [][2]int{{1, 1}, {-1, -1}, {-1, 1}, {1, -1}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}}

var knightOffsets = [8][2]int{{1, 2}, {2, 1}, {-1, -2}, {-2, -1}, {2, -1}, {-2, 1}, {-1, 2}, {1, -2}}
var kingOffsets = [8][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}}

var knightAttacks [64]uint64
var kingAttacks [64]uint64

func init() {
	for sq := range 64 {
		r, c := sq/8, sq%8
		knightAttacks[sq] = leaperAttackBits(r, c, knightOffsets[:])
		kingAttacks[sq] = leaperAttackBits(r, c, kingOffsets[:])
	}
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
	attacks := emptyMask
	checkers := 0

	opp := oppositeColor(color)
	kingRow, kingCol := mg.board.kingPos(opp)
	mg.board.setCell(kingRow, kingCol, newPiece('*'))

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

func pawnAttackBits(row, col int, color Color) uint64 {
	dir := 1
	if color == Color(White) {
		dir = -1
	}
	var bits uint64
	if inBounds(row+dir, col-1) {
		bits = bitboardAddOne(bits, row+dir, col-1)
	}
	if inBounds(row+dir, col+1) {
		bits = bitboardAddOne(bits, row+dir, col+1)
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
	pinnedPieces := emptyMask
	piece := mg.board.getCell(kingRow, kingCol)
	color := getColor(piece)

	for slidingIdx, move := range allDirs {
		isDiagonal := slidingIdx < 4
		row := kingRow + move[0]
		col := kingCol + move[1]
		foundFriendly := false
		var candidate Square

		for row >= 0 && row <= 7 && col >= 0 && col <= 7 {
			current := mg.board.getCell(row, col)
			if isEmpty(current) {
				row += move[0]
				col += move[1]
				continue
			}

			if sameColor(current, color) {
				if foundFriendly {
					break
				}
				foundFriendly = true
				candidate = Square{row: row, col: col, piece: current}
				row += move[0]
				col += move[1]
				continue
			}

			enemyType := pieceType(current)
			validSlider := false
			if isDiagonal && (isBishop(enemyType) || isQueen(enemyType)) {
				validSlider = true
			}
			if !isDiagonal && (isRook(enemyType) || isQueen(enemyType)) {
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
	checkMask := emptyMask

	piece := mg.board.getCell(kingRow, kingCol)
	color := getColor(piece)
	oppositeColor := oppositeColor(color)
	// check pawns
	directions := []int{1, -1}
	if oppositeColor == Color(White) {
		directions = directions[:1]
	} else {
		directions = directions[1:]
	}

	pawnAttackRow := kingRow + directions[0]
	if inBounds(pawnAttackRow, kingCol-1) && mg.board.canCapture(pawnAttackRow, kingCol-1, color) && isPawn(pieceType(mg.board.getCell(pawnAttackRow, kingCol-1))) {
		checkMask = bitboardAddOne(checkMask, pawnAttackRow, kingCol-1)
		return checkMask
	}
	if inBounds(pawnAttackRow, kingCol+1) && mg.board.canCapture(pawnAttackRow, kingCol+1, color) && isPawn(pieceType(mg.board.getCell(pawnAttackRow, kingCol+1))) {
		checkMask = bitboardAddOne(checkMask, pawnAttackRow, kingCol+1)
		return checkMask
	}

	var row int
	var col int
	for _, move := range knightOffsets {
		row = kingRow + move[0]
		col = kingCol + move[1]
		if row >= 0 && row <= 7 && col >= 0 && col <= 7 {
			if mg.board.canCapture(row, col, color) && isKnight(pieceType(mg.board.getCell(row, col))) {
				checkMask = bitboardAddOne(checkMask, row, col)
				return checkMask
			}
		}
	}

	// check sliding
	checkMask = mg.slidingRays(kingRow, kingCol, color, checkMask, false)

	return checkMask
}

func (mg *MoveGenerator) slidingRays(kingRow int, kingCol int, color Color, checkMask uint64, isPinRay bool) uint64 {

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

				checkMask = bitboardAddOne(checkMask, row, col)
				for j := range 7 {
					_ = j
					row -= move[0]
					col -= move[1]
					if row == kingRow && col == kingCol {
						if !isPinRay {
							return checkMask
						} else {
							break
						}
					}
					checkMask = bitboardAddOne(checkMask, row, col)
				}
			}

			row += move[0]
			col += move[1]
		}
	}
	return checkMask
}
func (mg *MoveGenerator) generateMoves(onlyCaptures bool) []Move {
	moves := make([]Move, 0, 150)
	color := mg.board.currentColor()

	oppositeColor := oppositeColor(color)
	attackedSquares, checkers := mg.generateAttacks(oppositeColor, false)

	kingRow, kingCol := mg.board.kingPos(color)

	//Double check
	var checkMask uint64
	if bitboardCheckOne(attackedSquares, kingRow, kingCol) && checkers > 1 {
		piece := Square{row: kingRow, col: kingCol, piece: mg.board.getCell(kingRow, kingCol)}
		moves = mg.generateKingMoves(piece, color, false, attackedSquares, onlyCaptures, moves)
		return moves
	} else if bitboardCheckOne(attackedSquares, kingRow, kingCol) {
		checkMask = mg.checkRays(kingRow, kingCol)
	} else {
		checkMask = fullMask
	}

	pieces := mg.board.piecesGenerator()
	pinnedPieces := mg.pinnedPieces(kingRow, kingCol)
	for _, p := range pieces {
		if getColor(p.piece) != color {
			continue
		}
		if bitboardCheckOne(pinnedPieces, p.row, p.col) {
			moves = append(moves, mg.generatePinnedMoves(p, color, kingRow, kingCol, checkMask, onlyCaptures)...)
			continue
		}

		pieceType := pieceType(p.piece)
		if isPawn(pieceType) {
			moves = mg.generatePawnMoves(p, color, checkMask, onlyCaptures, moves)
		}

		if isKnight(pieceType) {
			moves = mg.generateKnightMoves(p, color, false, checkMask, onlyCaptures, moves)
		}

		if isSlidingPiece(pieceType) {
			moves = mg.generateSlidingMoves(p, color, pieceType, false, checkMask, onlyCaptures, moves)
		}

		if isKing(pieceType) {
			moves = mg.generateKingMoves(p, color, false, attackedSquares, onlyCaptures, moves)
		}
	}

	if !onlyCaptures {
		moves = append(moves, mg.generateCastles(color, attackedSquares)...)
	}

	return moves
}

func (mg *MoveGenerator) generateCastles(color Color, checkMask uint64) []Move {
	moves := []Move{}
	availability := mg.board.castleAvailable

	if color == Color(White) && availability&CastleWK != 0 {
		if !bitboardCheckOne(checkMask, 7, 4) && !bitboardCheckOne(checkMask, 7, 5) && !bitboardCheckOne(checkMask, 7, 6) {
			if mg.board.cellEmpty(7, 5) && mg.board.cellEmpty(7, 6) && mg.board.getCell(7, 7) == newPiece('R') && mg.board.getCell(7, 4) == newPiece('K') {
				startSquare := Square{row: 7, col: 4, piece: newPiece('K')}
				endSquare := Square{row: 7, col: 6, piece: newPiece('*')}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	if color == Color(White) && availability&CastleWQ != 0 {

		if !bitboardCheckOne(checkMask, 7, 4) && !bitboardCheckOne(checkMask, 7, 3) && !bitboardCheckOne(checkMask, 7, 2) {
			if mg.board.cellEmpty(7, 1) && mg.board.cellEmpty(7, 2) && mg.board.cellEmpty(7, 3) && mg.board.getCell(7, 0) == newPiece('R') && mg.board.getCell(7, 4) == newPiece('K') {
				startSquare := Square{row: 7, col: 4, piece: newPiece('K')}
				endSquare := Square{row: 7, col: 2, piece: newPiece('*')}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	if color == Color(Black) && availability&CastleBK != 0 {

		if !bitboardCheckOne(checkMask, 0, 4) && !bitboardCheckOne(checkMask, 0, 5) && !bitboardCheckOne(checkMask, 0, 6) {
			if mg.board.cellEmpty(0, 5) && mg.board.cellEmpty(0, 6) && mg.board.getCell(0, 7) == newPiece('r') && mg.board.getCell(0, 4) == newPiece('k') {
				startSquare := Square{row: 0, col: 4, piece: newPiece('k')}
				endSquare := Square{row: 0, col: 6, piece: newPiece('*')}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	if color == Color(Black) && availability&CastleBQ != 0 {
		if !bitboardCheckOne(checkMask, 0, 4) && !bitboardCheckOne(checkMask, 0, 3) && !bitboardCheckOne(checkMask, 0, 2) {
			if mg.board.cellEmpty(0, 1) && mg.board.cellEmpty(0, 2) && mg.board.cellEmpty(0, 3) && mg.board.getCell(0, 0) == newPiece('r') && mg.board.getCell(0, 4) == newPiece('k') {
				startSquare := Square{row: 0, col: 4, piece: newPiece('k')}
				endSquare := Square{row: 0, col: 2, piece: newPiece('*')}
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
func (mg *MoveGenerator) generatePinnedMoves(p Square, color Color, kingRow int, kingCol int, checkMask uint64, onlyCaptures bool) []Move {
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
	pinnedMask := emptyMask

	for i := range 7 {
		_ = i
		currentRow += direction[0]
		currentCol += direction[1]
		if currentRow < 0 || currentRow > 7 || currentCol < 0 || currentCol > 7 {
			break
		}
		if mg.board.canCapture(currentRow, currentCol, color) {
			pinnedMask = bitboardAddOne(pinnedMask, currentRow, currentCol)
			break
		}
		pinnedMask = bitboardAddOne(pinnedMask, currentRow, currentCol)
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
		pinnedMask = bitboardAddOne(pinnedMask, currentRow, currentCol)

	}

	pinnedMask &= checkMask

	if isPawn(currentPieceType) {
		moves = mg.generatePawnMoves(p, color, pinnedMask, onlyCaptures, moves)
	}

	if isSlidingPiece(currentPieceType) {
		moves = mg.generateSlidingMoves(p, color, currentPieceType, false, pinnedMask, onlyCaptures, moves)
	}

	return moves
}
func (mg *MoveGenerator) generateSlidingMoves(p Square, color Color, pt PieceType, isAttacks bool, checkMask uint64, onlyCaptures bool, moves []Move) []Move {
	moveDirs := allDirs

	if isRook(pt) {
		moveDirs = straightDirs
	} else if isBishop(pt) {
		moveDirs = diagonalDirs
	}
	currentRow := p.row
	currentCol := p.col
	for _, move := range moveDirs {
		row := currentRow + move[0]
		col := currentCol + move[1]
		for i := range 7 {
			_ = i
			if row < 0 || row > 7 || col < 0 || col > 7 {
				break
			}

			if mg.board.cellEmpty(row, col) {
				if bitboardCheckOne(checkMask, row, col) {
					startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
					endSquare := Square{row: row, col: col, piece: mg.board.getCell(row, col)}
					if !onlyCaptures || !mg.board.cellEmpty(row, col) {
						moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
					}
				}
			} else if mg.board.canCapture(row, col, color) {
				if bitboardCheckOne(checkMask, row, col) {
					startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
					endSquare := Square{row: row, col: col, piece: mg.board.getCell(row, col)}
					if !onlyCaptures || !mg.board.cellEmpty(row, col) {
						moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
					}
				}

				break
			} else {
				if isAttacks {
					startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
					endSquare := Square{row: row, col: col, piece: mg.board.getCell(row, col)}
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
				}
				break
			}

			row += move[0]
			col += move[1]

		}

	}
	return moves
}

func (mg *MoveGenerator) generateKingMoves(p Square, color Color, isAttack bool, attackMask uint64, onlyCaptures bool, moves []Move) []Move {
	currentRow := p.row
	currentCol := p.col
	var row int
	var col int
	for _, move := range kingOffsets {
		row = currentRow + move[0]
		col = currentCol + move[1]
		if row >= 0 && row <= 7 && col >= 0 && col <= 7 {
			if isAttack || mg.board.canCapture(row, col, color) || mg.board.cellEmpty(row, col) {
				if !bitboardCheckOne(attackMask, row, col) {
					startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
					endSquare := Square{row: row, col: col, piece: mg.board.getCell(row, col)}

					if !onlyCaptures || !mg.board.cellEmpty(row, col) {
						moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
					}
				}
			}
		}
	}
	return moves

}
func (mg *MoveGenerator) generateKnightMoves(p Square, color Color, isAttacks bool, checkMask uint64, onlyCaptures bool, moves []Move) []Move {

	currentRow := p.row
	currentCol := p.col

	startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
	pieceAttacks := knightAttacks[currentRow*8+currentCol]
	if !isAttacks {
		pieceAttacks &= ^mg.board.getColorBitboard(color)
		pieceAttacks &= checkMask
	}
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

func (mg *MoveGenerator) generatePawnMoves(p Square, color Color, checkMask uint64, onlyCaptures bool, moves []Move) []Move {

	directions := []int{1, 2, -1, -2}

	startRow := 1
	enpassantRow := 4
	if color == Color(White) {
		directions = directions[2:]
		startRow = 6
		enpassantRow = 3
	} else {
		directions = directions[:2]
	}

	// Forward Moves
	currentRow := p.row
	currentCol := p.col
	if !onlyCaptures {
		if mg.board.cellEmpty(p.row+directions[0], p.col) {
			if bitboardCheckOne(checkMask, currentRow+directions[0], currentCol) {
				startSquare := Square{row: p.row, col: p.col, piece: p.piece}
				endSquare := Square{row: p.row + directions[0], col: p.col, piece: mg.board.getCell(p.row+directions[0], p.col)}
				if p.row+directions[0] == 0 || p.row+directions[0] == 7 {
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: PieceType(Queen)})
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: PieceType(Rook)})
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: PieceType(Bishop)})
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: PieceType(Knight)})
				} else {
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
				}
			}
		}
	}

	if !onlyCaptures {
		if p.row == startRow && mg.board.cellEmpty(p.row+directions[1], p.col) && mg.board.cellEmpty(p.row+directions[0], p.col) {

			if bitboardCheckOne(checkMask, currentRow+directions[1], currentCol) {
				startSquare := Square{row: p.row, col: p.col, piece: p.piece}
				endSquare := Square{row: p.row + directions[1], col: p.col, piece: mg.board.getCell(p.row+directions[1], p.col)}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	//Capture Moves
	if p.col > 0 && mg.board.canCapture(p.row+directions[0], currentCol-1, color) {
		if bitboardCheckOne(checkMask, currentRow+directions[0], currentCol-1) {
			startSquare := Square{row: p.row, col: p.col, piece: p.piece}
			endSquare := Square{row: p.row + directions[0], col: p.col - 1, piece: mg.board.getCell(p.row+directions[0], p.col-1)}
			if !mg.board.cellEmpty(p.row+directions[0], p.col-1) {
				if p.row+directions[0] == 0 || p.row+directions[0] == 7 {
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: PieceType(Queen)})
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: PieceType(Rook)})
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: PieceType(Bishop)})
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: PieceType(Knight)})
				} else {
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
				}
			}
		}
	}
	if p.col < 7 && mg.board.canCapture(p.row+directions[0], p.col+1, color) {
		if bitboardCheckOne(checkMask, currentRow+directions[0], currentCol+1) {
			startSquare := Square{row: p.row, col: p.col, piece: p.piece}
			endSquare := Square{row: p.row + directions[0], col: p.col + 1, piece: mg.board.getCell(p.row+directions[0], p.col+1)}
			if !mg.board.cellEmpty(p.row+directions[0], p.col+1) {
				if p.row+directions[0] == 0 || p.row+directions[0] == 7 {
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: PieceType(Queen)})
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: PieceType(Rook)})
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: PieceType(Bishop)})
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare, isPromotion: true, promotionPieceType: PieceType(Knight)})
				} else {
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
				}
			}
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
func (mg *MoveGenerator) generatePawnAttacks(p Square, color Color, moves []Move) []Move {

	directions := []int{1, 2, -1, -2}

	if color == Color(White) {
		directions = directions[2:]
	} else {
		directions = directions[:2]
	}

	targetRow := p.row + directions[0]
	if inBounds(targetRow, p.col-1) {
		startSquare := Square{row: p.row, col: p.col, piece: p.piece}
		endSquare := Square{row: targetRow, col: p.col - 1, piece: mg.board.getCell(targetRow, p.col-1)}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}
	if inBounds(targetRow, p.col+1) {
		startSquare := Square{row: p.row, col: p.col, piece: p.piece}
		endSquare := Square{row: targetRow, col: p.col + 1, piece: mg.board.getCell(targetRow, p.col+1)}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}

	return moves

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
