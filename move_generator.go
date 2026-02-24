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

func (mg *MoveGenerator) generateMoves() []Move {
	moves := []Move{}

	pieces := mg.board.piecesGenerator()
	for _, p := range pieces {
		if isWhite(p.piece) {
			continue
		}

		pieceType := pieceType(p.piece)
		if isPawn(pieceType) {
			moves = append(moves, mg.generatePawnMoves(p)...)
		}
	}
	return moves
}

func (mg *MoveGenerator) generatePawnMoves(p Square) []Move {
	moves := []Move{}

	if mg.board.cellEmpty(p.row+1, p.col) {
		startSquare := Square{row: p.row, col: p.col, piece: p.piece}
		endSquare := Square{row: p.row + 1, col: p.col, piece: p.piece}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}

	if p.row == 1 && mg.board.cellEmpty(p.row+2, p.col) && mg.board.cellEmpty(p.row+1, p.col) {
		startSquare := Square{row: p.row, col: p.col, piece: p.piece}
		endSquare := Square{row: p.row + 2, col: p.col, piece: p.piece}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}
	return moves
}

func toSquare(row int, col int) string {
	return fmt.Sprintf("%c%d", 'a'+col, 8-row)
}

func (mg *MoveGenerator) randomMove() MoveString {
	moves := mg.generateMoves()

	seed := rand.NewSource(time.Now().Unix())
	r := rand.New(seed)

	random_index := r.Intn(len(moves))
	random_move := moves[random_index]
	startSquare := toSquare(random_move.startSquare.row, random_move.startSquare.col)
	endSquare := toSquare(random_move.endSquare.row, random_move.endSquare.col)
	return MoveString{startSquare: startSquare, endSquare: endSquare}
}
