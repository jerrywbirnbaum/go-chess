package main

import (
	"fmt"
	"slices"
	"testing"
)

func moveToUCI(move Move) string {
	return fmt.Sprintf("%s%s",
		toSquare(move.getStartSquare().row, move.getStartSquare().col),
		toSquare(move.getEndSquare().row, move.getEndSquare().col),
	)
}

func movesToUCISet(moves []Move) map[string]struct{} {
	result := make(map[string]struct{}, len(moves))
	for _, move := range moves {
		result[moveToUCI(move)] = struct{}{}
	}
	return result
}

func assertMovesExactly(t *testing.T, fen string, expected []string) {
	t.Helper()

	board := initBoard()
	board.updateFromFEN(fen)
	mg := MoveGenerator{board: &board}
	moves := mg.generateMoves(false)
	gotSet := movesToUCISet(moves)

	if len(gotSet) != len(expected) {
		t.Fatalf("FEN %q: expected %d unique moves, got %d", fen, len(expected), len(gotSet))
	}

	for _, uci := range expected {
		if _, ok := gotSet[uci]; !ok {
			t.Fatalf("FEN %q: missing move %s", fen, uci)
		}
	}
}

func assertHasMove(t *testing.T, fen string, expectedMove string) {
	t.Helper()

	board := initBoard()
	board.updateFromFEN(fen)
	mg := MoveGenerator{board: &board}
	moves := mg.generateMoves(false)

	uciMoves := make([]string, 0, len(moves))
	for _, move := range moves {
		uciMoves = append(uciMoves, moveToUCI(move))
	}
	if !slices.Contains(uciMoves, expectedMove) {
		t.Fatalf("FEN %q: expected move %s to exist, got moves: %v", fen, expectedMove, uciMoves)
	}
}

func assertMissingMove(t *testing.T, fen string, missingMove string) {
	t.Helper()

	board := initBoard()
	board.updateFromFEN(fen)
	mg := MoveGenerator{board: &board}
	moves := mg.generateMoves(false)

	uciMoves := make([]string, 0, len(moves))
	for _, move := range moves {
		uciMoves = append(uciMoves, moveToUCI(move))
	}
	if slices.Contains(uciMoves, missingMove) {
		t.Fatalf("FEN %q: expected move %s to be illegal, got moves: %v", fen, missingMove, uciMoves)
	}
}

func assertMovesExactlyWithOnlyCaptures(t *testing.T, fen string, onlyCaptures bool, expected []string) {
	t.Helper()

	board := initBoard()
	board.updateFromFEN(fen)
	mg := MoveGenerator{board: &board}
	moves := mg.generateMoves(onlyCaptures)
	gotSet := movesToUCISet(moves)

	if len(gotSet) != len(expected) {
		t.Fatalf("FEN %q (onlyCaptures=%v): expected %d unique moves, got %d", fen, onlyCaptures, len(expected), len(gotSet))
	}

	for _, uci := range expected {
		if _, ok := gotSet[uci]; !ok {
			t.Fatalf("FEN %q (onlyCaptures=%v): missing move %s", fen, onlyCaptures, uci)
		}
	}
}

func TestMoveGeneration_StartPositionExactWhite(t *testing.T) {
	expected := []string{
		"a2a3", "a2a4", "b2b3", "b2b4", "c2c3", "c2c4", "d2d3", "d2d4",
		"e2e3", "e2e4", "f2f3", "f2f4", "g2g3", "g2g4", "h2h3", "h2h4",
		"b1a3", "b1c3", "g1f3", "g1h3",
	}
	assertMovesExactly(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", expected)
}

func TestMoveGeneration_StartPositionExactBlack(t *testing.T) {
	expected := []string{
		"a7a6", "a7a5", "b7b6", "b7b5", "c7c6", "c7c5", "d7d6", "d7d5",
		"e7e6", "e7e5", "f7f6", "f7f5", "g7g6", "g7g5", "h7h6", "h7h5",
		"b8a6", "b8c6", "g8f6", "g8h6",
	}
	assertMovesExactly(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1", expected)
}

func TestMoveGeneration_DoubleCheckOnlyKingMoves(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("1k5r/8/2N5/4Q3/8/8/8/8 b KQkq - 0 1")
	mg := MoveGenerator{board: &board}
	moves := mg.generateMoves(false)

	if len(moves) == 0 {
		t.Fatalf("expected king escapes in double-check position")
	}
	for _, move := range moves {
		if moveToUCI(move)[:2] != "b8" {
			t.Fatalf("double check should only allow king moves, got %s", moveToUCI(move))
		}
	}
}

func TestMoveGeneration_CastlingBlockedByAttack(t *testing.T) {
	// Black rook on f8 attacks f1, so white cannot castle king-side.
	fen := "5r2/8/8/8/8/8/8/R3K2R w KQ - 0 1"
	assertMissingMove(t, fen, "e1g1")
	assertHasMove(t, fen, "e1c1")
}

func TestMoveGeneration_EnPassantIllegalWhenItExposesCheck(t *testing.T) {
	// Capturing en-passant would expose a rook attack on black king.
	fen := "8/8/8/8/k2Pp1R1/8/8/8 b - d3 0 1"
	assertMissingMove(t, fen, "e4d3")
}

func TestMoveGeneration_EnPassantLegalWhenSafe(t *testing.T) {
	fen := "8/8/8/3Pp3/8/8/8/4K2k w - e6 0 1"
	assertHasMove(t, fen, "d5e6")
}

func TestMoveGeneration_PinnedPieceCannotLeavePinLine(t *testing.T) {
	// White bishop on e2 is pinned to king e1 by black rook e8.
	fen := "4r1k1/8/8/8/8/8/4B3/4K3 w - - 0 1"
	assertMissingMove(t, fen, "e2f3")
	assertMissingMove(t, fen, "e2d3")
}

func TestPerft_StartPositionRegression(t *testing.T) {
	board := initBoard()
	for _, tc := range []struct {
		depth int
		nodes int
	}{
		{depth: 1, nodes: 20},
		{depth: 2, nodes: 400},
		{depth: 3, nodes: 8902},
		{depth: 4, nodes: 197281},
		{depth: 5, nodes: 4865609},
		// {depth: 6, nodes: 119060324},
	} {
		got := moveGenerationRecursive(tc.depth, board)
		if got != tc.nodes {
			t.Fatalf("start position perft depth %d: expected %d, got %d", tc.depth, tc.nodes, got)
		}
	}
}

func TestPerft_KiwiPeteRegression(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - ")
	for _, tc := range []struct {
		depth int
		nodes int
	}{
		{depth: 1, nodes: 48},
		{depth: 2, nodes: 2039},
		{depth: 3, nodes: 97862},
		{depth: 4, nodes: 4085603},
		// {depth: 5, nodes: 193690690},
	} {
		got := moveGenerationRecursive(tc.depth, board)
		if got != tc.nodes {
			t.Fatalf("start position perft depth %d: expected %d, got %d", tc.depth, tc.nodes, got)
		}
	}
}

func TestMoveGeneration_QueensideCastleAllowedWhenB1Attacked(t *testing.T) {
	// b1 is attacked by a rook, but that should not prevent white O-O-O.
	fen := "1r6/8/8/8/8/8/8/R3K2R w Q - 0 1"
	assertHasMove(t, fen, "e1c1")
}

func TestBoardUpdateCastle_WhiteQueensideRookMoveClearsQOnly(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("4k3/8/8/8/8/8/8/R3K2R w KQ - 0 1")

	move := newMove(
		Square{row: 7, col: 0, piece: newPiece('R')},
		Square{row: 7, col: 1, piece: newPiece('*')},
		false, 0)
	board.makeMove(&move)

	if board.castleAvailable != CastleWK {
		t.Fatalf("expected castle rights to be CastleWK after moving a1 rook, got %d", board.castleAvailable)
	}
}

func TestBoardMakeMove_BlackDoublePushSetsEnPassant(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("4k3/4p3/8/8/8/8/8/4K3 b - - 0 1")
	move := newMove(
		Square{row: 1, col: 4, piece: newPiece('p')},
		Square{row: 3, col: 4, piece: newPiece('*')},
		false, 0)
	board.makeMove(&move)
	if board.enpassant != 4 {
		t.Fatalf("expected en-passant column e (4) after e7e5, got %q", board.enpassant)
	}
}

func TestMoveGeneration_OnlyCaptures_NoCapturesReturnsEmpty(t *testing.T) {
	fen := "8/8/8/8/8/8/8/4K3 w - - 0 1"
	assertMovesExactlyWithOnlyCaptures(t, fen, true, []string{})
}

func TestMoveGeneration_OnlyCaptures_FiltersQuietMoves(t *testing.T) {
	// White knight has one capture (f5) plus multiple quiet moves; king has quiet moves.
	fen := "8/8/8/5p2/3N4/8/8/4K3 w - - 0 1"
	assertMovesExactlyWithOnlyCaptures(t, fen, true, []string{"d4f5"})
}

func TestMoveGeneration_OnlyCaptures_AllowsPawnDiagonalCaptures(t *testing.T) {
	// White pawn can capture d5 from e4. Quiet pawn pushes and king moves must be excluded.
	fen := "8/8/8/3p4/4P3/8/8/4K3 w - - 0 1"
	assertMovesExactlyWithOnlyCaptures(t, fen, true, []string{"e4d5"})
}

func TestMoveGeneration_OnlyCaptures_AllowsEnPassant(t *testing.T) {
	fen := "8/8/8/3Pp3/8/8/8/4K2k w - e6 0 1"
	assertMovesExactlyWithOnlyCaptures(t, fen, true, []string{"d5e6"})
}
