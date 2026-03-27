package main

import (
	// "fmt"
	"testing"
)

func TestKiwiPeteDivide(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - ")
	mg := MoveGenerator{board: &board}
	moves := mg.generateMoves(false)
	total := 0
	for i := range moves {
		move := &moves[i]
		board.makeMove(move)
		count := moveGenerationRecursive(2, board)
		board.unmakeMove(move)
		// fmt.Printf("%s: %d\n", moveToUCI(*move), count)
		total += count
	}
	// fmt.Printf("Total: %d\n", total)
}
