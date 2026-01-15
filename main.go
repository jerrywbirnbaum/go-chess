package main

func main() {
	board := Board{
		board: [8][8]rune{
			{'r', 'n', 'b', 'q', 'k', 'b', 'n', 'r'},
			{'p', 'p', 'p', 'p', 'p', 'p', 'p', 'p'},
			{'*', '*', '*', '*', '*', '*', '*', '*'},
			{'*', '*', '*', '*', '*', '*', '*', '*'},
			{'*', '*', '*', '*', '*', '*', '*', '*'},
			{'*', '*', '*', '*', '*', '*', '*', '*'},
			{'P', 'P', 'P', 'P', 'P', 'P', 'P', 'P'},
			{'R', 'N', 'B', 'Q', 'K', 'B', 'N', 'R'},
		},
		isWhiteTurn: true,
	}
	board.moveAlgebraicNotation("fxe5")
	board.printBoard()
}
