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

func (mg *MoveGenerator) generateMoves(color Color) []Move {
	moves := []Move{}

	pieces := mg.board.piecesGenerator()
	for _, p := range pieces {
		if isWhite(p.piece) && color != Color(White) {
			continue
		}

		pieceType := pieceType(p.piece)
		if isPawn(pieceType) {
			moves = append(moves, mg.generatePawnMoves(p, color)...)
		}

		if isKnight(pieceType) {
			moves = append(moves, mg.generateKnightMoves(p, color)...)
		}

		if isSlidingPiece(pieceType) {
			moves = append(moves, mg.generateSlidingMoves(p, color, pieceType)...)
		}

		if isKing(pieceType) {
			moves = append(moves, mg.generateKingMoves(p, color)...)
		}
	}
	return moves
}

func (mg *MoveGenerator) generateSlidingMoves(p Square, color Color, pt PieceType) []Move {
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
				startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
				endSquare := Square{row: row, col: col, piece: p.piece}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
			} else if mg.board.canCapture(row, col, color) {
				startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
				endSquare := Square{row: row, col: col, piece: p.piece}
				moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
				break
			} else {
				break
			}

			row += move[0]
			col += move[1]

		}

	}
	return moves
}

func (mg *MoveGenerator) generateKingMoves(p Square, color Color) []Move {

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
		if row >= 0 && row <= 7 && col >= 0 && col <= 7 && mg.board.cellEmpty(row, col) {
			startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
			endSquare := Square{row: row, col: col, piece: p.piece}
			moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
		}
	}
	return moves

}
func (mg *MoveGenerator) generateKnightMoves(p Square, color Color) []Move {

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
		if row >= 0 && row <= 7 && col >= 0 && col <= 7 && mg.board.cellEmpty(row, col) {
			startSquare := Square{row: currentRow, col: currentCol, piece: p.piece}
			endSquare := Square{row: row, col: col, piece: p.piece}
			moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
		}
	}
	return moves

}
func (mg *MoveGenerator) generatePawnMoves(p Square, color Color) []Move {
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
	if mg.board.cellEmpty(p.row+directions[0], p.col) {
		startSquare := Square{row: p.row, col: p.col, piece: p.piece}
		endSquare := Square{row: p.row + directions[0], col: p.col, piece: p.piece}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}

	if p.row == startRow && mg.board.cellEmpty(p.row+directions[1], p.col) && mg.board.cellEmpty(p.row+1, p.col) {
		startSquare := Square{row: p.row, col: p.col, piece: p.piece}
		endSquare := Square{row: p.row + directions[1], col: p.col, piece: p.piece}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}

	//Capture Moves
	if p.col > 0 && mg.board.canCapture(p.row+directions[0], p.col-1, color) {
		startSquare := Square{row: p.row, col: p.col, piece: p.piece}
		endSquare := Square{row: p.row + directions[0], col: p.col - 1, piece: p.piece}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}
	if p.col < 7 && mg.board.canCapture(p.row+directions[0], p.col+1, color) {
		startSquare := Square{row: p.row, col: p.col, piece: p.piece}
		endSquare := Square{row: p.row + directions[0], col: p.col + 1, piece: p.piece}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}

	//ENPASSANT
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
