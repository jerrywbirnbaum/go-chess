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
	}
	return moves
}

func (mg *MoveGenerator) generatePawnMoves(p Square, color Color) []Move {
	moves := []Move{}

	directions := []int{1, 2, -1, -2}

	startRow := 1
	if color == Color(White) {
		directions = directions[2:]
		startRow = 6
	} else {
		directions = directions[:3]
	}
	fmt.Println(directions)

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
	// left_capture :=

	return moves
}

func toSquare(row int, col int) string {
	return fmt.Sprintf("%c%d", 'a'+col, 8-row)
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
