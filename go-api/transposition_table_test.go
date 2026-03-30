package main

import (
	"testing"
)

func TestTranspositionTableInitCreatesMap(t *testing.T) {
	var tt TranspositionTable
	tt = initTranspositionTable()

	if tt.table == nil {
		t.Fatal("expected init to allocate table map")
	}
}

func TestTranspositionTablePushAndLookup(t *testing.T) {
	var tt TranspositionTable
	tt = initTranspositionTable()

	const key int64 = 42
	tt.push(key, 5, 1, 320, 0)

	_, depth, flag, evaluation, _ := tt.lookup(key)
	if depth != 5 || flag != 1 || evaluation != 320 {
		t.Fatalf("unexpected lookup values: got (%d, %d, %d), expected (%d, %d, %d)", depth, flag, evaluation, 5, 1, 320)
	}
}

func TestTranspositionTablePushOverwritesExistingKey(t *testing.T) {
	var tt TranspositionTable
	tt = initTranspositionTable()

	const key int64 = 99
	tt.push(key, 2, 0, 100, 0)
	tt.push(key, 7, 2, -50, 0)

	_, depth, flag, evaluation, _ := tt.lookup(key)
	if depth != 7 || flag != 2 || evaluation != -50 {
		t.Fatalf("unexpected lookup after overwrite: got (%d, %d, %d), expected (%d, %d, %d)", depth, flag, evaluation, 7, 2, -50)
	}
}

func TestTranspositionTableLookupMissingKeyReturnsZeroValues(t *testing.T) {
	var tt TranspositionTable
	tt = initTranspositionTable()

	_, depth, flag, evaluation, _ := tt.lookup(123456)
	if depth != 0 || flag != 0 || evaluation != 0 {
		t.Fatalf("unexpected missing-key values: got (%d, %d, %d), expected (0, 0, 0)", depth, flag, evaluation)
	}
}

func TestTranspositionTableStoresBestMove(t *testing.T) {
	tt := initTranspositionTable()
	move := newMove(
		Square{row: 1, col: 3, piece: WhitePawn},
		Square{row: 3, col: 3, piece: EmptyPiece},
		false, EmptyPieceType,
	)
	packed := packMove(move)
	tt.push(1, 3, 0, 100, packed)

	_, _, _, _, got := tt.lookup(1)
	if got != packed {
		t.Fatalf("expected bestMove %d, got %d", packed, got)
	}
}

func TestPackMoveRoundTrip(t *testing.T) {
	move := newMove(
		Square{row: 4, col: 2, piece: WhiteKnight},
		Square{row: 5, col: 4, piece: EmptyPiece},
		false, EmptyPieceType,
	)
	packed := packMove(move)
	if packed == 0 {
		t.Fatal("expected packMove to return non-zero for a real move")
	}
	if !comparePackedMoves(packed, packed) {
		t.Fatal("expected comparePackedMoves(x, x) to be true")
	}
}

func TestComparePackedMovesIgnoresCapturedPiece(t *testing.T) {
	src := Square{row: 3, col: 3, piece: WhiteQueen}
	dst := Square{row: 6, col: 3}

	m1 := newMove(src, Square{row: dst.row, col: dst.col, piece: EmptyPiece}, false, EmptyPieceType)
	m2 := newMove(src, Square{row: dst.row, col: dst.col, piece: BlackRook}, false, EmptyPieceType)

	if !comparePackedMoves(m1.getMoveBits(), m2.getMoveBits()) {
		t.Fatal("expected moves with same squares but different captured piece to compare equal")
	}
}

func TestComparePackedMovesIgnoresMovingPiece(t *testing.T) {
	dst := Square{row: 4, col: 4, piece: EmptyPiece}

	m1 := newMove(Square{row: 3, col: 4, piece: WhitePawn}, dst, false, EmptyPieceType)
	m2 := newMove(Square{row: 3, col: 4, piece: WhiteQueen}, dst, false, EmptyPieceType)

	if !comparePackedMoves(m1.getMoveBits(), m2.getMoveBits()) {
		t.Fatal("expected moves with same squares but different moving piece to compare equal")
	}
}

func TestComparePackedMovesDifferentStartSquare(t *testing.T) {
	dst := Square{row: 4, col: 4, piece: EmptyPiece}

	m1 := newMove(Square{row: 3, col: 4, piece: WhitePawn}, dst, false, EmptyPieceType)
	m2 := newMove(Square{row: 2, col: 4, piece: WhitePawn}, dst, false, EmptyPieceType)

	if comparePackedMoves(m1.getMoveBits(), m2.getMoveBits()) {
		t.Fatal("expected moves with different start rows to compare not equal")
	}
}

func TestComparePackedMovesDifferentEndSquare(t *testing.T) {
	src := Square{row: 1, col: 0, piece: WhitePawn}

	m1 := newMove(src, Square{row: 2, col: 0, piece: EmptyPiece}, false, EmptyPieceType)
	m2 := newMove(src, Square{row: 2, col: 1, piece: EmptyPiece}, false, EmptyPieceType)

	if comparePackedMoves(m1.getMoveBits(), m2.getMoveBits()) {
		t.Fatal("expected moves with different end columns to compare not equal")
	}
}

func TestComparePackedMovesDistinguishesPromotionPieceType(t *testing.T) {
	src := Square{row: 6, col: 0, piece: WhitePawn}
	dst := Square{row: 7, col: 0, piece: EmptyPiece}

	queen := newMove(src, dst, true, Queen)
	knight := newMove(src, dst, true, Knight)

	if comparePackedMoves(queen.getMoveBits(), knight.getMoveBits()) {
		t.Fatal("expected queen and knight promotions to compare not equal")
	}
}

func TestComparePackedMovesDistinguishesPromotionFromNonPromotion(t *testing.T) {
	src := Square{row: 6, col: 0, piece: WhitePawn}
	dst := Square{row: 7, col: 0, piece: EmptyPiece}

	promo := newMove(src, dst, true, Queen)
	nonPromo := newMove(src, dst, false, EmptyPieceType)

	if comparePackedMoves(promo.getMoveBits(), nonPromo.getMoveBits()) {
		t.Fatal("expected promotion and non-promotion to compare not equal")
	}
}
