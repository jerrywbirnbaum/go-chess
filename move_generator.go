package main

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"time"
)

type Move struct {
	startSquare          Square
	endSquare            Square
	previousEnpassant    string
	previousCastleRights string
	nextEnpassant        string
	nextCastleRights     string
	isEnpassant          bool
	isCastleKingSide     bool
	isCastleQueenSide    bool
	isPromotion          bool
	enpassantCapture     Piece
}

type MoveString struct {
	startSquare string
	endSquare   string
}

type MoveGenerator struct {
	board Board
}

func inBounds(row int, col int) bool {
	return row >= 0 && row <= 7 && col >= 0 && col <= 7
}

func (mg *MoveGenerator) updateBoard(board Board) {
	mg.board = board
}

func (mg *MoveGenerator) generateAttacks(color Color, slidingOnly bool) [8][8]int {
	moves := []Move{}
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

	var kingRune rune
	if color == Color(White) {
		kingRune = 'k'
	} else {
		kingRune = 'K'
	}
	oppositeColor := oppositeColor(color)

	//Remove king
	var kingRow int
	var kingCol int
	kingExists := false
	for i := range 8 {
		for j := range 8 {
			piece := mg.board.board[i][j]
			pieceType := pieceType(piece)
			if sameColor(piece, oppositeColor) && isKing(pieceType) {
				kingExists = true
				kingRow = i
				kingCol = j
			}

		}
	}

	if kingExists {
		mg.board.board[kingRow][kingCol] = EmptyPiece
	}

	emptyMask := [8][8]int{
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}

	fullMask := [8][8]int{
		{1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1},
	}

	pieces := mg.board.piecesGenerator()
	for _, p := range pieces {
		if getColor(p.piece) != color {
			continue
		}

		pieceType := pieceType(p.piece)
		if !slidingOnly && isPawn(pieceType) {
			moves = mg.generatePawnAttacks(p, color, moves)
		}
		if !slidingOnly && isKnight(pieceType) {
			moves = mg.generateKnightMoves(p, color, true, fullMask, false, moves)
		}

		if isSlidingPiece(pieceType) {
			moves = mg.generateSlidingMoves(p, color, pieceType, true, fullMask, false, moves)
		}

		if !slidingOnly && isKing(pieceType) {
			moves = mg.generateKingMoves(p, color, true, emptyMask, false, moves)
		}
	}

	if kingExists {
		mg.board.board[kingRow][kingCol] = newPiece(kingRune)
	}

	for _, move := range moves {
		attacks[move.endSquare.row][move.endSquare.col] += 1
	}
	return attacks
}

func (mg *MoveGenerator) pinnedPieces(kingRow int, kingCol int) []Square {
	var pinnedPieces []Square
	piece := mg.board.board[kingRow][kingCol]
	color := getColor(piece)

	slidingMoves := [][2]int{
		{1, 1},
		{-1, -1},
		{-1, 1},
		{1, -1},
		{1, 0},
		{-1, 0},
		{0, 1},
		{0, -1},
	}

	for slidingIdx, move := range slidingMoves {
		isDiagonal := slidingIdx < 4
		row := kingRow + move[0]
		col := kingCol + move[1]
		foundFriendly := false
		var candidate Square

		for row >= 0 && row <= 7 && col >= 0 && col <= 7 {
			current := mg.board.board[row][col]
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
				pinnedPieces = append(pinnedPieces, candidate)
			}
			break
		}
	}

	return pinnedPieces
}
func (mg *MoveGenerator) checkRays(kingRow int, kingCol int) [8][8]int {
	checkMask := [8][8]int{
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}

	piece := mg.board.board[kingRow][kingCol]
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
	if inBounds(pawnAttackRow, kingCol-1) && mg.board.canCapture(pawnAttackRow, kingCol-1, color) && isPawn(pieceType(mg.board.board[pawnAttackRow][kingCol-1])) {
		checkMask[pawnAttackRow][kingCol-1] = 1
		return checkMask
	}
	if inBounds(pawnAttackRow, kingCol+1) && mg.board.canCapture(pawnAttackRow, kingCol+1, color) && isPawn(pieceType(mg.board.board[pawnAttackRow][kingCol+1])) {
		checkMask[pawnAttackRow][kingCol+1] = 1
		return checkMask
	}

	//checkKnights
	knightMoves := [][2]int{
		{1, 2},
		{2, 1},
		{-1, -2},
		{-2, -1},
		{2, -1},
		{-2, 1},
		{-1, 2},
		{1, -2},
	}

	var row int
	var col int
	for _, move := range knightMoves {
		row = kingRow + move[0]
		col = kingCol + move[1]
		if row >= 0 && row <= 7 && col >= 0 && col <= 7 {
			if mg.board.canCapture(row, col, color) && isKnight(pieceType(mg.board.board[row][col])) {
				checkMask[row][col] = 1
				return checkMask
			}
		}
	}

	// check sliding
	checkMask = mg.slidingRays(kingRow, kingCol, color, checkMask, false)

	return checkMask
}

func (mg *MoveGenerator) slidingRays(kingRow int, kingCol int, color Color, checkMask [8][8]int, isPinRay bool) [8][8]int {
	// check sliding
	slidingMoves := [][2]int{
		{1, 1},
		{-1, -1},
		{-1, 1},
		{1, -1},
		{1, 0},
		{-1, 0},
		{0, 1},
		{0, -1},
	}

	for slidingIdx, move := range slidingMoves {
		isDiagonal := slidingIdx < 4
		row := kingRow + move[0]
		col := kingCol + move[1]
		for i := range 7 {
			_ = i
			if row < 0 || row > 7 || col < 0 || col > 7 {
				break
			}

			pieceType := pieceType(mg.board.board[row][col])

			if mg.board.canCapture(row, col, color) && (isSlidingPiece(pieceType) || isPinRay) {
				if !isPinRay && isDiagonal && isRook(pieceType) {
					break
				}

				if !isPinRay && !isDiagonal && isBishop(pieceType) {
					break
				}
				checkMask[row][col] = 1
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
					checkMask[row][col] = 1
				}
			}

			row += move[0]
			col += move[1]
		}
	}
	return checkMask
}
func (mg *MoveGenerator) generateMoves(onlyCaptures bool) []Move {
	// moves := []Move{}
	moves := make([]Move, 0, 90)
	color := mg.board.currentColor()

	oppositeColor := oppositeColor(color)
	attackedSquares := mg.generateAttacks(oppositeColor, false)

	var kingRow int
	var kingCol int
	kingExists := false
	for i := range 8 {
		for j := range 8 {
			piece := mg.board.board[i][j]
			pieceType := pieceType(piece)
			if sameColor(piece, color) && isKing(pieceType) {
				kingExists = true
				kingRow = i
				kingCol = j
			}

		}
	}

	//Double check
	var checkMask [8][8]int
	if kingExists && attackedSquares[kingRow][kingCol] > 1 {
		piece := Square{row: kingRow, col: kingCol, piece: mg.board.board[kingRow][kingCol]}
		moves = mg.generateKingMoves(piece, color, false, attackedSquares, onlyCaptures, moves)
		return moves
	} else if kingExists && attackedSquares[kingRow][kingCol] == 1 {
		checkMask = mg.checkRays(kingRow, kingCol)
	} else {
		checkMask = [8][8]int{
			{1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1},
		}
	}

	pieces := mg.board.piecesGenerator()
	pinnedPieces := mg.pinnedPieces(kingRow, kingCol)
	for _, p := range pieces {
		if getColor(p.piece) != color {
			continue
		}
		if slices.Contains(pinnedPieces, p) {
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

func (mg *MoveGenerator) generateCastles(color Color, checkMask [8][8]int) []Move {
	moves := []Move{}
	availability := mg.board.castleAvailable

	if color == Color(White) && strings.Contains(availability, "K") {
		if checkMask[7][4] == 0 && checkMask[7][5] == 0 && checkMask[7][6] == 0 {
			if mg.board.cellEmpty(7, 5) && mg.board.cellEmpty(7, 6) && mg.board.board[7][7] == newPiece('R') && mg.board.board[7][4] == newPiece('K') {
				startSquare := Square{row: 7, col: 4, piece: newPiece('K')}
				endSquare := Square{row: 7, col: 6, piece: newPiece('*')}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	if color == Color(White) && strings.Contains(availability, "Q") {
		if checkMask[7][4] == 0 && checkMask[7][3] == 0 && checkMask[7][2] == 0 {
			if mg.board.cellEmpty(7, 1) && mg.board.cellEmpty(7, 2) && mg.board.cellEmpty(7, 3) && mg.board.board[7][0] == newPiece('R') && mg.board.board[7][4] == newPiece('K') {
				startSquare := Square{row: 7, col: 4, piece: newPiece('K')}
				endSquare := Square{row: 7, col: 2, piece: newPiece('*')}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	if color == Color(Black) && strings.Contains(availability, "k") {
		if checkMask[0][4] == 0 && checkMask[0][5] == 0 && checkMask[0][6] == 0 {
			if mg.board.cellEmpty(0, 5) && mg.board.cellEmpty(0, 6) && mg.board.board[0][7] == newPiece('r') && mg.board.board[0][4] == newPiece('k') {
				startSquare := Square{row: 0, col: 4, piece: newPiece('k')}
				endSquare := Square{row: 0, col: 6, piece: newPiece('*')}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	if color == Color(Black) && strings.Contains(availability, "q") {
		if checkMask[0][4] == 0 && checkMask[0][3] == 0 && checkMask[0][2] == 0 {
			if mg.board.cellEmpty(0, 1) && mg.board.cellEmpty(0, 2) && mg.board.cellEmpty(0, 3) && mg.board.board[0][0] == newPiece('r') && mg.board.board[0][4] == newPiece('k') {
				startSquare := Square{row: 0, col: 4, piece: newPiece('k')}
				endSquare := Square{row: 0, col: 2, piece: newPiece('*')}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	return moves
}
func (mg *MoveGenerator) pinDirection(kingRow int, kingCol int, row int, col int) ([2]int, bool) {
	slidingMoves := [][2]int{
		{1, 1},
		{-1, -1},
		{-1, 1},
		{1, -1},
		{1, 0},
		{-1, 0},
		{0, 1},
		{0, -1},
	}
	if kingRow == row {
		if col > kingCol {
			return slidingMoves[6], false
		} else if col < kingCol {
			return slidingMoves[7], false
		}
	} else if kingCol == col {
		if row > kingRow {
			return slidingMoves[4], false
		} else if row < kingRow {
			return slidingMoves[5], false
		}
	} else if kingRow > row {
		if col > kingCol {
			return slidingMoves[2], true
		} else if col < kingCol {
			return slidingMoves[1], true
		}
	} else if kingRow < row {
		if col > kingCol {
			return slidingMoves[0], true
		} else if col < kingCol {
			return slidingMoves[3], true
		}
	}
	return slidingMoves[0], true

}
func (mg *MoveGenerator) generatePinnedMoves(p Square, color Color, kingRow int, kingCol int, checkMask [8][8]int, onlyCaptures bool) []Move {
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
	pinnedMask := [8][8]int{
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}

	for i := range 7 {
		_ = i
		currentRow += direction[0]
		currentCol += direction[1]
		if currentRow < 0 || currentRow > 7 || currentCol < 0 || currentCol > 7 {
			break
		}
		if mg.board.canCapture(currentRow, currentCol, color) {
			pinnedMask[currentRow][currentCol] = 1
			break
		}
		pinnedMask[currentRow][currentCol] = 1

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

		piece := mg.board.board[currentRow][currentCol]
		pieceType := pieceType(piece)
		if isKing(pieceType) {
			break
		}
		pinnedMask[currentRow][currentCol] = 1

	}

	for i := range 8 {
		for j := range 8 {
			if pinnedMask[i][j] > 0 && checkMask[i][j] > 0 {
				pinnedMask[i][j] = 1
			}
		}
	}

	if isPawn(currentPieceType) {
		moves = mg.generatePawnMoves(p, color, pinnedMask, onlyCaptures, moves)
	}

	if isSlidingPiece(currentPieceType) {
		moves = mg.generateSlidingMoves(p, color, currentPieceType, false, pinnedMask, onlyCaptures, moves)
	}

	return moves
}
func (mg *MoveGenerator) generateSlidingMoves(p Square, color Color, pt PieceType, isAttacks bool, checkMask [8][8]int, onlyCaptures bool, moves []Move) []Move {

	slidingMoves := [][2]int{
		{1, 1},
		{-1, -1},
		{-1, 1},
		{1, -1},
		{1, 0},
		{-1, 0},
		{0, 1},
		{0, -1},
	}

	if isRook(pt) {
		slidingMoves = slidingMoves[4:]
	} else if isBishop(pt) {
		slidingMoves = slidingMoves[:4]
	}
	currentRow := p.row
	currentCol := p.col
	for _, move := range slidingMoves {
		row := currentRow + move[0]
		col := currentCol + move[1]
		for i := range 7 {
			_ = i
			if row < 0 || row > 7 || col < 0 || col > 7 {
				break
			}

			if mg.board.cellEmpty(row, col) {
				if checkMask[row][col] == 1 {
					startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
					endSquare := Square{row: row, col: col, piece: mg.board.board[row][col]}
					if !onlyCaptures || !mg.board.cellEmpty(row, col) {
						moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
					}
				}
			} else if mg.board.canCapture(row, col, color) {
				if checkMask[row][col] == 1 {
					startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
					endSquare := Square{row: row, col: col, piece: mg.board.board[row][col]}
					if !onlyCaptures || !mg.board.cellEmpty(row, col) {
						moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
					}
					break
				}
			} else {
				if isAttacks {
					startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
					endSquare := Square{row: row, col: col, piece: mg.board.board[row][col]}
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

func (mg *MoveGenerator) generateKingMoves(p Square, color Color, isAttack bool, attackMask [8][8]int, onlyCaptures bool, moves []Move) []Move {

	kingMoves := [][2]int{
		{1, 1},
		{1, -1},
		{-1, 1},
		{-1, -1},
		{1, 0},
		{-1, 0},
		{0, 1},
		{0, -1},
	}

	currentRow := p.row
	currentCol := p.col
	var row int
	var col int
	for _, move := range kingMoves {
		row = currentRow + move[0]
		col = currentCol + move[1]
		if row >= 0 && row <= 7 && col >= 0 && col <= 7 {
			if isAttack || mg.board.canCapture(row, col, color) || mg.board.cellEmpty(row, col) {
				if attackMask[row][col] == 0 {
					startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
					endSquare := Square{row: row, col: col, piece: mg.board.board[row][col]}

					if !onlyCaptures || !mg.board.cellEmpty(row, col) {
						moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
					}
				}
			}
		}
	}
	return moves

}
func (mg *MoveGenerator) generateKnightMoves(p Square, color Color, isAttacks bool, checkMask [8][8]int, onlyCaptures bool, moves []Move) []Move {

	knightMoves := [][2]int{
		{1, 2},
		{2, 1},
		{-1, -2},
		{-2, -1},
		{2, -1},
		{-2, 1},
		{-1, 2},
		{1, -2},
	}

	currentRow := p.row
	currentCol := p.col
	var row int
	var col int
	for _, move := range knightMoves {
		row = currentRow + move[0]
		col = currentCol + move[1]
		if row >= 0 && row <= 7 && col >= 0 && col <= 7 {
			if isAttacks || mg.board.canCapture(row, col, color) || mg.board.cellEmpty(row, col) {
				if checkMask[row][col] == 1 {
					startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
					endSquare := Square{row: row, col: col, piece: mg.board.board[row][col]}
					if !onlyCaptures || !mg.board.cellEmpty(row, col) {
						moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
					}
				}
			}
		}
	}

	return moves
}

func (mg *MoveGenerator) generatePawnMoves(p Square, color Color, checkMask [8][8]int, onlyCaptures bool, moves []Move) []Move {

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
	if mg.board.cellEmpty(p.row+directions[0], p.col) {
		if checkMask[currentRow+directions[0]][currentCol] == 1 {
			startSquare := Square{row: p.row, col: p.col, piece: p.piece}
			endSquare := Square{row: p.row + directions[0], col: p.col, piece: mg.board.board[p.row+directions[0]][p.col]}
			if !onlyCaptures {
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	if p.row == startRow && mg.board.cellEmpty(p.row+directions[1], p.col) && mg.board.cellEmpty(p.row+directions[0], p.col) {
		if checkMask[currentRow+directions[1]][currentCol] == 1 {
			startSquare := Square{row: p.row, col: p.col, piece: p.piece}
			endSquare := Square{row: p.row + directions[1], col: p.col, piece: mg.board.board[p.row+directions[1]][p.col]}
			if !onlyCaptures {
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	//Capture Moves
	if p.col > 0 && mg.board.canCapture(p.row+directions[0], currentCol-1, color) {
		if checkMask[currentRow+directions[0]][currentCol-1] == 1 {
			startSquare := Square{row: p.row, col: p.col, piece: p.piece}
			endSquare := Square{row: p.row + directions[0], col: p.col - 1, piece: mg.board.board[p.row+directions[0]][p.col-1]}

			if !onlyCaptures || !mg.board.cellEmpty(p.row+directions[0], p.col-1) {
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}
	if p.col < 7 && mg.board.canCapture(p.row+directions[0], p.col+1, color) {
		if checkMask[currentRow+directions[0]][currentCol+1] == 1 {
			startSquare := Square{row: p.row, col: p.col, piece: p.piece}
			endSquare := Square{row: p.row + directions[0], col: p.col + 1, piece: mg.board.board[p.row+directions[0]][p.col+1]}
			if !onlyCaptures || !mg.board.cellEmpty(p.row+directions[0], p.col+1) {
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			}
		}
	}

	//ENPASSANT
	if mg.board.enpassant != "-" {
		ep_row, ep_col := fromSquare(mg.board.enpassant)
		if p.row == enpassantRow && (ep_col-p.col == 1 || ep_col-p.col == -1) {
			startSquare := Square{row: p.row, col: p.col, piece: p.piece}
			endSquare := Square{row: ep_row, col: ep_col, piece: mg.board.board[ep_row][ep_col]}
			enpassantMove := Move{startSquare: startSquare, endSquare: endSquare}
			if !mg.enpassantCheck(enpassantMove, color) && (!onlyCaptures || mg.board.cellEmpty(ep_row, ep_col)) {
				moves = append(moves, enpassantMove)
			}
		}
	}

	return moves
}

func (mg *MoveGenerator) enpassantCheck(move Move, color Color) bool {
	simulatedBoard := mg.board
	simulatedBoard.makeMove(&move)

	simulatedMoveGenerator := MoveGenerator{board: simulatedBoard}
	attacks := simulatedMoveGenerator.generateAttacks(oppositeColor(color), true)

	var kingRow int
	var kingCol int
	kingFound := false
	for i := range 8 {
		for j := range 8 {
			piece := simulatedBoard.board[i][j]
			if sameColor(piece, color) && isKing(pieceType(piece)) {
				kingRow = i
				kingCol = j
				kingFound = true
				break
			}
		}
		if kingFound {
			break
		}
	}

	inCheck := kingFound && attacks[kingRow][kingCol] > 0

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
		endSquare := Square{row: targetRow, col: p.col - 1, piece: mg.board.board[targetRow][p.col-1]}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}
	if inBounds(targetRow, p.col+1) {
		startSquare := Square{row: p.row, col: p.col, piece: p.piece}
		endSquare := Square{row: targetRow, col: p.col + 1, piece: mg.board.board[targetRow][p.col+1]}
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
