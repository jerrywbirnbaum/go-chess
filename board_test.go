package main

import (
	"testing"
)

func TestSameColor(t *testing.T) {
	piece := newPiece('p')
	color := Color(Black)

	result := sameColor(piece, color)
	if !result {
		t.Errorf("Failed TestSameColor")
	}

}
