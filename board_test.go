package main

import (
	"fmt"
	"testing"
)

func TestSameColor(t *testing.T) {
	fmt.Println()
	piece := newPiece('p')
	color := Color(Black)

	result := sameColor(piece, color)
	if !result {
		t.Errorf("Failed TestSameColor")
	}

}

func TestMakeMove(t *testing.T) {
	board := initBoard()
	piece := newPiece('P')
	startSquare := Square{row: 6, col: 0, piece: piece}
	endSquare := Square{row: 5, col: 0, piece: piece}
	move := Move{startSquare: startSquare, endSquare: endSquare}
	board.makeMove(move)

	expected := `'r''n''b''q''k''b''n''r'
'p''p''p''p''p''p''p''p'
'*''*''*''*''*''*''*''*'
'*''*''*''*''*''*''*''*'
'*''*''*''*''*''*''*''*'
'P''*''*''*''*''*''*''*'
'*''P''P''P''P''P''P''P'
'R''N''B''Q''K''B''N''R'
Is white's turn: true`
	result := board.printBoard()

	if result != expected {
		t.Errorf("Failed MakeMove")
	}
	board.unmakeMove()

	expected = `'r''n''b''q''k''b''n''r'
'p''p''p''p''p''p''p''p'
'*''*''*''*''*''*''*''*'
'*''*''*''*''*''*''*''*'
'*''*''*''*''*''*''*''*'
'*''*''*''*''*''*''*''*'
'P''P''P''P''P''P''P''P'
'R''N''B''Q''K''B''N''R'
Is white's turn: true`
	result = board.printBoard()

	if result != expected {
		t.Errorf("Failed MakeMove")
	}
}
