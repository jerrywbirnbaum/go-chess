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
	board.updateFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	board.printBoard()

	board.updateFromFEN("r2qk3/2pp1p2/3p2b1/8/1P2R3/8/P7/1N2K3 b - - 0 1")
	board.printBoard()
}
