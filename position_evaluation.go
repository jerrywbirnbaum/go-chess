package main

func basicEval(b Board) float64 {
	eval := 0.0
	var sign float64

	pieces := b.piecesGenerator()
	for _, p := range pieces {
		if getColor(p.piece) == b.currentColor() {
			sign = 1
		} else {
			sign = -1
		}

		pieceType := pieceType(p.piece)
		if isPawn(pieceType) {
			eval += sign * 1
		}
		if isBishop(pieceType) || isKnight(pieceType) {
			eval += sign * 3
		}
		if isRook(pieceType) {
			eval += sign * 5
		}
		if isQueen(pieceType) {
			eval += sign * 9
		}
	}
	return eval
}
