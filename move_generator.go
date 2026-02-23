package main

type Move struct {
	startSquare Square
	endSquare   Square
}

type MoveGenerator struct {
	board Board
}

func (mg *MoveGenerator) generateMoves() []Move {
	moves := []Move{}

	pieces := mg.board.piecesGenerator()
	for _, p := range pieces {
		if isWhite(p.piece) {
			continue
		}

		pieceType := pieceType(p.piece)
		if isPawn(pieceType) {
			moves = append(moves, mg.generatePawnMoves(p)...)
		}
	}
	return moves
}

func (mg *MoveGenerator) generatePawnMoves(p Square) []Move {
	moves := []Move{}

	if mg.board.cellEmpty(p.row+1, p.col) {
		startSquare := Square{row: p.row, col: p.col, piece: p.piece}
		endSquare := Square{row: p.row + 1, col: p.col, piece: p.piece}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}

	if p.row == 1 && mg.board.cellEmpty(p.row+2, p.col) && mg.board.cellEmpty(p.row+1, p.col) {
		startSquare := Square{row: p.row, col: p.col, piece: p.piece}
		endSquare := Square{row: p.row + 2, col: p.col, piece: p.piece}
		moves = append(moves, Move{startSquare: startSquare, endSquare: endSquare})
	}
	return moves
}
