package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Move struct {
	startSquare Square
	endSquare   Square
}

type MoveString struct {
	startSquare string
	endSquare   string
}

type MoveGenerator struct {
	board Board
}

func (mg *MoveGenerator) updateBoard(board Board) {
	mg.board = board
}

func (mg *MoveGenerator) generateAttacks(color Color) [8][8]int {
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
		if isPawn(pieceType) {
			moves = append(moves, mg.generatePawnAttacks(p, color)...)
		}
		if isKnight(pieceType) {
			moves = append(moves, mg.generateKnightMoves(p, color, true, fullMask)...)
		}

		if isSlidingPiece(pieceType) {
			moves = append(moves, mg.generateSlidingMoves(p, color, pieceType, true, fullMask)...)
		}

		if isKing(pieceType) {
			moves = append(moves, mg.generateKingMoves(p, color, true, emptyMask)...)
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

func (mg *MoveGenerator) pinnedPieces(kingRow int, kingCol int) [][2]int {
	pinnedRays := [8][8]int{
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
	pinnedRays = mg.slidingRays(kingRow, kingCol, oppositeColor, pinnedRays, true)
	var pinnedPieces [][2]int

	attackedSquares := mg.generateAttacks(oppositeColor)

	for i := range 8 {
		for j := range 8 {
			if pinnedRays[i][j] > 0 && attackedSquares[i][j] > 0 {
				newSquare := [2]int{i, j}
				pinnedPieces = append(pinnedPieces, newSquare)

			}
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

	if kingCol > 0 && mg.board.canCapture(kingRow+directions[0], kingCol-1, color) && isPawn(pieceType(mg.board.board[kingRow+directions[0]][kingCol-1])) {
		checkMask[kingRow+directions[0]][kingCol-1] = 1
		return checkMask
	}
	if kingCol < 7 && mg.board.canCapture(kingRow+directions[0], kingCol+1, color) && isPawn(pieceType(mg.board.board[kingRow+directions[0]][kingCol+1])) {
		checkMask[kingRow+directions[0]][kingCol+1] = 1
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

	var isDiagonal bool
	for slidingIdx, move := range slidingMoves {
		if slidingIdx > 3 {
			isDiagonal = false
		} else {
			isDiagonal = true
		}
		row := kingRow + move[0]
		col := kingCol + move[1]
		for i := range 7 {
			_ = i
			if row < 0 || row > 7 || col < 0 || col > 7 {
				break
			}

			pieceType := pieceType(mg.board.board[row][col])

			if mg.board.canCapture(row, col, color) && (isSlidingPiece(pieceType) || isPinRay) {

				if !isPinRay && isDiagonal && isStriaghtSlidingPiece(pieceType) {
					continue
				}

				if !isPinRay && !isDiagonal && isDiagonalSlidingPiece(pieceType) {
					continue
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
func (mg *MoveGenerator) generateMoves(color Color) []Move {
	moves := []Move{}

	oppositeColor := oppositeColor(color)
	attackedSquares := mg.generateAttacks(oppositeColor)

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
		moves = append(moves, mg.generateKingMoves(piece, color, false, attackedSquares)...)
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
	for _, p := range pieces {
		if getColor(p.piece) != color {
			continue
		}

		pieceType := pieceType(p.piece)
		if isPawn(pieceType) {
			moves = append(moves, mg.generatePawnMoves(p, color, checkMask)...)
		}

		if isKnight(pieceType) {
			moves = append(moves, mg.generateKnightMoves(p, color, false, checkMask)...)
		}

		if isSlidingPiece(pieceType) {
			moves = append(moves, mg.generateSlidingMoves(p, color, pieceType, false, checkMask)...)
		}

		if isKing(pieceType) {
			moves = append(moves, mg.generateKingMoves(p, color, false, attackedSquares)...)
		}
	}
	return moves
}

func (mg *MoveGenerator) generateSlidingMoves(p Square, color Color, pt PieceType, isAttacks bool, checkMask [8][8]int) []Move {
	moves := []Move{}

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
					endSquare := Square{row: row, col: col, piece: p.piece}
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
				}
			} else if mg.board.canCapture(row, col, color) {
				if checkMask[row][col] == 1 {
					startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
					endSquare := Square{row: row, col: col, piece: p.piece}
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
					break
				}
			} else {
				if isAttacks {
					startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
					endSquare := Square{row: row, col: col, piece: p.piece}
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

func (mg *MoveGenerator) generateKingMoves(p Square, color Color, isAttack bool, attackMask [8][8]int) []Move {
	moves := []Move{}

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
					endSquare := Square{row: row, col: col, piece: p.piece}
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
				}
			}
		}
	}
	return moves

}
func (mg *MoveGenerator) generateKnightMoves(p Square, color Color, isAttacks bool, checkMask [8][8]int) []Move {

	moves := []Move{}

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
					endSquare := Square{row: row, col: col, piece: p.piece}
					moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
				}
			}
		}
	}

	return moves
}

func (mg *MoveGenerator) generatePawnMoves(p Square, color Color, checkMask [8][8]int) []Move {
	moves := []Move{}

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
			endSquare := Square{row: p.row + directions[0], col: p.col, piece: p.piece}
			moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
		}
	}

	if p.row == startRow && mg.board.cellEmpty(p.row+directions[1], p.col) && mg.board.cellEmpty(p.row+directions[0], p.col) {
		if checkMask[currentRow+directions[1]][currentCol] == 1 {
			startSquare := Square{row: p.row, col: p.col, piece: p.piece}
			endSquare := Square{row: p.row + directions[1], col: p.col, piece: p.piece}
			moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
		}
	}

	//Capture Moves
	if p.col > 0 && mg.board.canCapture(p.row+directions[0], currentCol-1, color) {
		if checkMask[currentRow+directions[0]][currentCol-1] == 1 {
			startSquare := Square{row: p.row, col: p.col, piece: p.piece}
			endSquare := Square{row: p.row + directions[0], col: p.col - 1, piece: p.piece}
			moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
		}
	}
	if p.col < 7 && mg.board.canCapture(p.row+directions[0], p.col+1, color) {
		if checkMask[currentRow+directions[0]][currentCol+1] == 1 {
			startSquare := Square{row: p.row, col: p.col, piece: p.piece}
			endSquare := Square{row: p.row + directions[0], col: p.col + 1, piece: p.piece}
			moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
		}
	}

	//ENPASSANT
	//TODO: Check rules for enpassant
	if mg.board.enpassant != "-" {
		ep_row, ep_col := fromSquare(mg.board.enpassant)
		if p.row == enpassantRow && (ep_col-p.col == 1 || ep_col-p.col == -1) {
			startSquare := Square{row: p.row, col: p.col, piece: p.piece}
			endSquare := Square{row: ep_row, col: ep_col, piece: p.piece}
			moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
		}
	}

	return moves
}

func (mg *MoveGenerator) generatePawnAttacks(p Square, color Color) []Move {
	moves := []Move{}

	directions := []int{1, 2, -1, -2}

	if color == Color(White) {
		directions = directions[2:]
	} else {
		directions = directions[:2]
	}

	if p.col > 0 {
		startSquare := Square{row: p.row, col: p.col, piece: p.piece}
		endSquare := Square{row: p.row + directions[0], col: p.col - 1, piece: p.piece}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}
	if p.col < 7 {
		startSquare := Square{row: p.row, col: p.col, piece: p.piece}
		endSquare := Square{row: p.row + directions[0], col: p.col + 1, piece: p.piece}
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
	moves := mg.generateMoves(Color(Black))

	seed := rand.NewSource(time.Now().Unix())
	r := rand.New(seed)

	random_index := r.Intn(len(moves))
	random_move := moves[random_index]
	startSquare := toSquare(random_move.startSquare.row, random_move.startSquare.col)
	endSquare := toSquare(random_move.endSquare.row, random_move.endSquare.col)
	return MoveString{startSquare: startSquare, endSquare: endSquare}
}
