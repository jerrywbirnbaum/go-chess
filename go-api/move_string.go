package main

type MoveString struct {
	startSquare string
	endSquare   string
	promotion   string
	isPromotion bool
}

func moveToMovestring(bestMove Move) MoveString {
	startSquare := toSquare(bestMove.getStartSquare().row, bestMove.getStartSquare().col)
	endSquare := toSquare(bestMove.getEndSquare().row, bestMove.getEndSquare().col)
	promotion := "q"
	if bestMove.getIsPromotion() {
		if bestMove.getPromotionPieceType() == PieceType(Rook) {
			promotion = "r"
		} else if bestMove.getPromotionPieceType() == PieceType(Bishop) {
			promotion = "b"
		} else if bestMove.getPromotionPieceType() == PieceType(Knight) {
			promotion = "n"
		}
	}
	return MoveString{startSquare: startSquare, endSquare: endSquare, promotion: promotion, isPromotion: bestMove.getIsPromotion()}
}
