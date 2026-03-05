package main

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

func basicEval(b Board) int {
	eval := 0
	var sign int

	pieces := b.piecesGenerator()
	for _, p := range pieces {
		if getColor(p.piece) == b.currentColor() {
			sign = 1
		} else {
			sign = -1
		}

		pieceType := pieceType(p.piece)
		eval += sign * getPieceValue(pieceType)
		// if isPawn(pieceType) {
		// 	eval += sign * 1
		// }
		// if isBishop(pieceType) || isKnight(pieceType) {
		// 	eval += sign * 3
		// }
		// if isRook(pieceType) {
		// 	eval += sign * 5
		// }
		// if isQueen(pieceType) {
		// 	eval += sign * 9
		// }
	}
	return eval
}
