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

var knightMask = [8][8]int{
	{-50, -40, -30, -30, -30, -30, -40, -50},
	{-40, -20, 0, 0, 0, 0, -20, -40},
	{-30, 0, 10, 15, 15, 10, 0, -30},
	{-30, 5, 15, 20, 20, 15, 5, -30},
	{-30, 0, 15, 20, 20, 15, 0, -30},
	{-30, 5, 10, 15, 15, 10, 5, -30},
	{-40, -20, 0, 5, 5, 0, -20, -40},
	{-50, -40, -30, -30, -30, -30, -40, -50},
}

var whiteBishopMask = [8][8]int{
	{-20, -10, -10, -10, -10, -10, -10, -20},
	{-10, 0, 0, 0, 0, 0, 0, -10},
	{-10, 0, 5, 10, 10, 5, 0, -10},
	{-10, 5, 5, 10, 10, 5, 5, -10},
	{-10, 0, 10, 10, 10, 10, 0, -10},
	{-10, 10, 10, 10, 10, 10, 10, -10},
	{-10, 5, 0, 0, 0, 0, 5, -10},
	{-20, -10, -10, -10, -10, -10, -10, -20},
}

var blackBishopMask = [8][8]int{
	{-20, -10, -10, -10, -10, -10, -10, -20},
	{-10, 5, 0, 0, 0, 0, 5, -10},
	{-10, 10, 10, 10, 10, 10, 10, -10},
	{-10, 0, 10, 10, 10, 10, 0, -10},
	{-10, 5, 5, 10, 10, 5, 5, -10},
	{-10, 0, 5, 10, 10, 5, 0, -10},
	{-10, 0, 0, 0, 0, 0, 0, -10},
	{-20, -10, -10, -10, -10, -10, -10, -20},
}

var whiteRookMask = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{5, 10, 10, 10, 10, 10, 10, 5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{0, 0, 0, 5, 5, 0, 0, 0},
}

var blackRookMask = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{5, 10, 10, 10, 10, 10, 10, 5},
	{0, 0, 0, 5, 5, 0, 0, 0},
}

var queenMask = [8][8]int{
	{-20, -10, -10, -5, -5, -10, -10, -20},
	{-10, 0, 0, 0, 0, 0, 0, -10},
	{-10, 0, 5, 5, 5, 5, 0, -10},
	{-5, 0, 5, 5, 5, 5, 0, -5},
	{0, 0, 5, 5, 5, 5, 0, -5},
	{-10, 5, 5, 5, 5, 5, 0, -10},
	{-10, 0, 5, 0, 0, 0, 0, -10},
	{-20, -10, -10, -5, -5, -10, -10, -20},
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
	if isKnight(p) {
		return knightMask[row][col]

	}
	if isBishop(p) {
		if color == Color(White) {
			return whiteBishopMask[row][col]
		} else {
			return blackBishopMask[row][col]
		}

	}
	if isRook(p) {
		if color == Color(White) {
			return whiteRookMask[row][col]
		} else {
			return blackRookMask[row][col]
		}

	}
	if isQueen(p) {
		return queenMask[row][col]

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
