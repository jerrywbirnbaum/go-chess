package main

import (
	"fmt"
	"testing"
)

func TestBasicEval(t *testing.T) {
	fmt.Println()
	board := initBoard()
	result := basicEval(board)
	if result != 0 {
		t.Errorf("Failed TestBasicEval")
	}

	board.updateBoardFEN("qknbrp2/8/8/8/8/8/8/P7 b KQkq d3 0 1")
	result = basicEval(board)

	if result != -20 {
		t.Errorf("Failed TestBasicEval")
	}
}
