package main

type Move struct {
	previousCastleRights uint8
	enpassant            uint8
	nextCastleRights     uint8
	isEnpassant          bool
	isCastleKingSide     bool
	isCastleQueenSide    bool
	isPromotion          bool
	promotionPieceType   PieceType
	isNull               bool
	enpassantCapture     Piece
	previousZobristHash  int64
	moveBits             uint64
}

func (move *Move) setPreviousEnpassant(next uint8) {
	move.enpassant &= uint8(0b11110000)
	move.enpassant |= next
}

func (move *Move) getPreviousEnpassant() uint8 {
	return move.enpassant & uint8(0b00001111)
}

func (move *Move) setNextEnpassant(next uint8) {
	move.enpassant &= uint8(0b00001111)
	move.enpassant |= next << 4
}

func (move *Move) getNextEnpassant() uint8 {
	return move.enpassant >> 4
}

func newMove(startSq Square, endSq Square, isPromotion bool, promotionPieceType PieceType) Move {
	move := Move{}
	move.setStartSquare(startSq)
	move.setEndSquare(endSq)
	move.isPromotion = isPromotion
	move.promotionPieceType = promotionPieceType
	return move
}
func (move *Move) getStartSquare() Square {
	startRow := int(move.moveBits & uint64(0b111))
	startCol := int(move.moveBits >> 3 & uint64(0b111))
	startPiece := Piece(move.moveBits >> 6 & uint64(0b1111))
	return Square{row: startRow, col: startCol, piece: startPiece}
}

func (move *Move) setStartSquare(sq Square) {
	move.moveBits &= ^uint64(0b1111111111)
	move.moveBits |= uint64(sq.row)
	move.moveBits |= uint64(sq.col) << 3
	move.moveBits |= uint64(sq.piece) << 6
}

func (move *Move) getEndSquare() Square {
	endRow := int(move.moveBits >> 10 & uint64(0b111))
	endCol := int(move.moveBits >> 13 & uint64(0b111))
	endPiece := Piece(move.moveBits >> 16 & uint64(0b1111))
	return Square{row: endRow, col: endCol, piece: endPiece}
}

func (move *Move) setEndSquare(sq Square) {
	move.moveBits &= ^(uint64(0b1111111111) << 10)
	move.moveBits |= uint64(sq.row) << 10
	move.moveBits |= uint64(sq.col) << 13
	move.moveBits |= uint64(sq.piece) << 16
}
