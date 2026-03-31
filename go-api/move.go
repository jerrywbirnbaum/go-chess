package main

// moveBits layout:
//
//	Bits  0– 2   startRow            (3 bits)
//	Bits  3– 5   startCol            (3 bits)
//	Bits  6– 9   startPiece          (4 bits)
//	Bits 10–12   endRow              (3 bits)
//	Bits 13–15   endCol              (3 bits)
//	Bits 16–19   endPiece            (4 bits)
//	Bits 20–23   previousCastleRights(4 bits)
//	Bits 24–27   nextCastleRights    (4 bits)
//	Bits 28–31   previousEnpassant   (4 bits, 0–7 = file, 8 = none)
//	Bits 32–35   nextEnpassant       (4 bits)
//	Bit  36      isEnpassant         (1 bit)
//	Bit  37      isCastleKingSide    (1 bit)
//	Bit  38      isCastleQueenSide   (1 bit)
//	Bit  39      isPromotion         (1 bit)
//	Bit  40      isNull              (1 bit)
//	Bits 41–43   promotionPieceType  (3 bits)
//	Bits 44–47   enpassantCapture    (4 bits)
type Move struct {
	moveBits uint64
}

func newMove(startSq Square, endSq Square, isPromotion bool, promotionPieceType PieceType) Move {
	move := Move{}
	move.setStartSquare(startSq)
	move.setEndSquare(endSq)
	move.setIsPromotion(isPromotion)
	move.setPromotionPieceType(promotionPieceType)
	return move
}

func (move *Move) getMoveBits() uint64 {
	return move.moveBits
}

func (move *Move) setMoveBits(moveBits uint64) {
	move.moveBits = moveBits
}

// --- Start square (bits 0–9) ---

func (move *Move) getStartSquare() Square {
	startRow := int(move.moveBits & 0b111)
	startCol := int(move.moveBits >> 3 & 0b111)
	startPiece := Piece(move.moveBits >> 6 & 0b1111)
	return Square{row: startRow, col: startCol, piece: startPiece}
}

func (move *Move) setStartSquare(sq Square) {
	move.moveBits &= ^uint64(0b1111111111)
	move.moveBits |= uint64(sq.row)
	move.moveBits |= uint64(sq.col) << 3
	move.moveBits |= uint64(sq.piece) << 6
}

// --- End square (bits 10–19) ---

func (move *Move) getEndSquare() Square {
	endRow := int(move.moveBits >> 10 & 0b111)
	endCol := int(move.moveBits >> 13 & 0b111)
	endPiece := Piece(move.moveBits >> 16 & 0b1111)
	return Square{row: endRow, col: endCol, piece: endPiece}
}

func (move *Move) setEndSquare(sq Square) {
	move.moveBits &= ^(uint64(0b1111111111) << 10)
	move.moveBits |= uint64(sq.row) << 10
	move.moveBits |= uint64(sq.col) << 13
	move.moveBits |= uint64(sq.piece) << 16
}

// --- Castle rights (bits 20–27) ---

func (move *Move) getPreviousCastleRights() uint8 {
	return uint8(move.moveBits >> 20 & 0b1111)
}

func (move *Move) setPreviousCastleRights(v uint8) {
	move.moveBits &= ^(uint64(0b1111) << 20)
	move.moveBits |= uint64(v) << 20
}

func (move *Move) getNextCastleRights() uint8 {
	return uint8(move.moveBits >> 24 & 0b1111)
}

func (move *Move) setNextCastleRights(v uint8) {
	move.moveBits &= ^(uint64(0b1111) << 24)
	move.moveBits |= uint64(v) << 24
}

// --- En passant files (bits 28–35) ---

func (move *Move) getPreviousEnpassant() uint8 {
	return uint8(move.moveBits >> 28 & 0b1111)
}

func (move *Move) setPreviousEnpassant(v uint8) {
	move.moveBits &= ^(uint64(0b1111) << 28)
	move.moveBits |= uint64(v) << 28
}

func (move *Move) getNextEnpassant() uint8 {
	return uint8(move.moveBits >> 32 & 0b1111)
}

func (move *Move) setNextEnpassant(v uint8) {
	move.moveBits &= ^(uint64(0b1111) << 32)
	move.moveBits |= uint64(v) << 32
}

// --- Move type flags (bits 36–40) ---

func (move *Move) getIsEnpassant() bool {
	return move.moveBits>>36&1 == 1
}

func (move *Move) setIsEnpassant(v bool) {
	if v {
		move.moveBits |= uint64(1) << 36
	} else {
		move.moveBits &= ^(uint64(1) << 36)
	}
}

func (move *Move) getIsCastleKingSide() bool {
	return move.moveBits>>37&1 == 1
}

func (move *Move) setIsCastleKingSide(v bool) {
	if v {
		move.moveBits |= uint64(1) << 37
	} else {
		move.moveBits &= ^(uint64(1) << 37)
	}
}

func (move *Move) getIsCastleQueenSide() bool {
	return move.moveBits>>38&1 == 1
}

func (move *Move) setIsCastleQueenSide(v bool) {
	if v {
		move.moveBits |= uint64(1) << 38
	} else {
		move.moveBits &= ^(uint64(1) << 38)
	}
}

func (move *Move) getIsPromotion() bool {
	return move.moveBits>>39&1 == 1
}

func (move *Move) setIsPromotion(v bool) {
	if v {
		move.moveBits |= uint64(1) << 39
	} else {
		move.moveBits &= ^(uint64(1) << 39)
	}
}

func (move *Move) getIsNull() bool {
	return move.moveBits>>40&1 == 1
}

func (move *Move) setIsNull(v bool) {
	if v {
		move.moveBits |= uint64(1) << 40
	} else {
		move.moveBits &= ^(uint64(1) << 40)
	}
}

// --- Promotion piece type (bits 41–43) ---

func (move *Move) getPromotionPieceType() PieceType {
	return PieceType(move.moveBits >> 41 & 0b111)
}

func (move *Move) setPromotionPieceType(v PieceType) {
	move.moveBits &= ^(uint64(0b111) << 41)
	move.moveBits |= uint64(v) << 41
}

// --- En passant captured piece (bits 44–47) ---

func (move *Move) getEnpassantCapture() Piece {
	return Piece(move.moveBits >> 44 & 0b1111)
}

func (move *Move) setEnpassantCapture(v Piece) {
	move.moveBits &= ^(uint64(0b1111) << 44)
	move.moveBits |= uint64(v) << 44
}
