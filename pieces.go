package main

import "fmt"

type Piece struct {
	representation rune
	isWhite        bool
	isSlidingPiece bool
}

func (p Piece) printPiece() {
	fmt.Printf("%q", p.representation)
}
