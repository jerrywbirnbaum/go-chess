package main

var whitePawnMask = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{50, 50, 50, 50, 50, 50, 50, 50},
	{10, 10, 20, 30, 30, 20, 10, 10},
	{5, 5, 10, 25, 25, 10, 5, 5},
	{0, 0, 0, 20, 20, 0, 0, 0},
	{5, -5, -10, 0, 0, -10, -5, 5},
	{5, 10, 10, -20, -20, 10, 10, 5},
	{0, 0, 0, 0, 0, 0, 0, 0},
}

var blackPawnMask = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{5, 10, 10, -20, -20, 10, 10, 5},
	{5, -5, -10, 0, 0, -10, -5, 5},
	{0, 0, 0, 20, 20, 0, 0, 0},
	{5, 5, 10, 25, 25, 10, 5, 5},
	{10, 10, 20, 30, 30, 20, 10, 10},
	{50, 50, 50, 50, 50, 50, 50, 50},
	{0, 0, 0, 0, 0, 0, 0, 0},
}

func getPieceValue(p PieceType) int {
	switch p {
	case EmptyPieceType:
		return 0
	case Knight:
		return 320
	case Bishop:
		return 330
	case Rook:
		return 500
	case Queen:
		return 900
	case King:
		return 20000
	}
	return 0
}

func getPositionFactor(p PieceType, color Color, row int, col int) int {
	if isPawn(p) {
		if color == Color(White) {
			return whitePawnMask[row][col]
		} else {
			return blackPawnMask[row][col]
		}

	}
	return 0
}
func basicEval(b Board) int {
	eval := 0
	var sign int

	pieces := b.piecesGenerator()
	for _, p := range pieces {
		color := getColor(p.piece)
		if color == b.currentColor() {
			sign = 1
		} else {
			sign = -1
		}

		pieceType := pieceType(p.piece)
		eval += sign * getPieceValue(pieceType)
		eval += sign * getPositionFactor(pieceType, color, p.row, p.col)
	}
	return eval
}
