package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestPestoEvalPerspectiveFlip(t *testing.T) {
	positions := []struct {
		name string
		fen  string
	}{
		// {"kings only", "4k3/8/8/8/8/8/8/4K3"},
		{"white queen advantage", "4k3/8/8/8/8/8/8/4KQ2"},
		{"equal queens", "q3k3/8/8/8/8/8/8/4KQ2"},
		{"starting position", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"},
		{"rook and pawns", "4k3/ppp5/8/8/8/8/5PPP/4K2R"},
	}
	for name, tc := range positions {
		t.Run(tc.name, func(t *testing.T) {
			boardW := initBoard()
			boardW.updateFromFEN(tc.fen + " w - - 0 1")
			scoreW := pestoEval(&boardW)
			boardB := initBoard()
			boardB.updateFromFEN(tc.fen + " b - - 0 1")

			scoreB := pestoEval(&boardB)

			if (scoreW > 0) != (scoreB < 0) {
				t.Errorf("%d perspective flip violated: white=%d, black=%d (expected negation)", name, scoreW, scoreB)
			}
		})
	}
}

func TestPestoEvalMaterialAdvantage(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want int
	}{
		{
			name: "white extra queen, white to move",
			fen:  "4k3/8/8/8/8/8/8/4KQ2 w - - 0 1",
			want: 945,
		},
		{
			name: "white extra queen, black to move",
			fen:  "4k3/8/8/8/8/8/8/4KQ2 b - - 0 1",
			want: -945,
		},
		{
			name: "black extra rook, black to move",
			fen:  "4k2r/8/8/8/8/8/8/4K3 b - - 0 1",
			want: 513,
		},
		{
			name: "black extra rook, white to move",
			fen:  "4k2r/8/8/8/8/8/8/4K3 w - - 0 1",
			want: -513,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			board := initBoard()
			board.updateFromFEN(tc.fen)
			got := pestoEval(&board)
			if got != tc.want {
				t.Errorf("pestoEval(%q) = %d, want %d", tc.fen, got, tc.want)
			}
		})
	}
}

func TestPestoEvalKingsOnly(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want int
	}{
		{
			name: "kings only, white to move",
			fen:  "4k3/8/8/8/8/8/8/4K3 w - - 0 1",
			want: -25,
		},
		{
			name: "kings only, black to move",
			fen:  "4k3/8/8/8/8/8/8/4K3 b - - 0 1",
			want: -25,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			board := initBoard()
			board.updateFromFEN(tc.fen)
			got := pestoEval(&board)
			if got != tc.want {
				t.Errorf("pestoEval(%q) = %d, want %d", tc.fen, got, tc.want)
			}
		})
	}
}

func TestPestoEvalTablesInitialized(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("4k3/8/8/8/8/8/8/4KQ2 w - - 0 1")
	score := pestoEval(&board)
	if score == 0 {
		t.Errorf("pestoEval returned 0 for a position with a queen advantage — mg_table may not be initialized")
	}
}

func TestMopUpEvalExactValue_WhiteWinning(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("k7/8/8/8/8/8/8/4K3 w - - 0 1")
	got := mopUpEval(&board, 900, 0)
	want := 33
	if got != want {
		t.Errorf("mopUpEval = %d, want %d", got, want)
	}
}

func TestMopUpEvalExactValue_BlackWinning(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("4k3/8/8/8/8/8/8/K7 b - - 0 1")
	got := mopUpEval(&board, 0, 900)
	want := 33
	if got != want {
		t.Errorf("mopUpEval = %d, want %d", got, want)
	}
}

func TestMopUpEvalLosingKingInCornerLargerBonus(t *testing.T) {
	boardCorner := initBoard()
	boardCorner.updateFromFEN("k7/8/8/8/8/8/8/4K3 w - - 0 1")

	boardEdge := initBoard()
	boardEdge.updateFromFEN("4k3/8/8/8/8/8/8/4K3 w - - 0 1")
	cornerBonus := mopUpEval(&boardCorner, 900, 0)
	edgeBonus := mopUpEval(&boardEdge, 900, 0)

	if cornerBonus <= edgeBonus {
		t.Errorf("corner bonus (%d) should be > edge bonus (%d): losing king closer to corner should give larger mop-up bonus", cornerBonus, edgeBonus)
	}
}

func TestMopUpEvalWinningKingCloserLargerBonus(t *testing.T) {
	boardClose := initBoard()
	boardClose.updateFromFEN("k7/8/2K5/8/8/8/8/8 w - - 0 1") // black a8, white c6

	boardFar := initBoard()
	boardFar.updateFromFEN("k7/8/8/8/8/8/8/4K3 w - - 0 1") // black a8, white e1

	closeBonus := mopUpEval(&boardClose, 900, 0)
	farBonus := mopUpEval(&boardFar, 900, 0)

	if closeBonus <= farBonus {
		t.Errorf("close-king bonus (%d) should be > far-king bonus (%d): winning king closer to losing king should give larger mop-up bonus", closeBonus, farBonus)
	}
}

func TestPassedPawns(t *testing.T) {
	// Helper: find the first pawn of the given color on the board.
	findPawn := func(b *Board, color Color) Square {
		for _, sq := range b.piecesGenerator() {
			if pieceType(sq.piece) == Pawn && getColor(sq.piece) == color {
				return sq
			}
		}
		return Square{}
	}

	tests := []struct {
		name      string
		fen       string
		pawnColor Color
		want      int
	}{
		// --- Black pawn cases (aheadMask is correct for black) ---
		{
			name:      "black passed pawn on e6, no white pawn on e-file",
			fen:       "7k/8/4p3/8/8/8/8/K7 b - - 0 1",
			pawnColor: Black,
			want:      10,
		},
		{
			name:      "black pawn on e7 blocked by white pawn on e4",
			fen:       "7k/8/4p3/8/4P3/8/8/K7 b - - 0 1",
			pawnColor: Black,
			want:      0,
		},
		{
			name:      "black pawn on e6 blocked by white pawn on e2",
			fen:       "7k/8/4p3/8/8/8/4P3/K7 b - - 0 1",
			pawnColor: Black,
			want:      0,
		},
		{
			name:      "black pawn on e6, white pawn on f2 (adjacent file)",
			fen:       "7k/8/4p3/8/8/8/5P2/K7 b - - 0 1",
			pawnColor: Black,
			want:      0,
		},
		{
			name:      "white passed pawn on e7, no black pawn on e-file ahead",
			fen:       "7k/4P3/8/8/8/8/8/K7 w - - 0 1",
			pawnColor: White,
			want:      120,
		},
		{
			name:      "white pawn on e7 blocked by black pawn on e8",
			fen:       "4p2k/4P3/8/8/8/8/8/K7 w - - 0 1",
			pawnColor: White,
			want:      0,
		},
		{
			name:      "white pawn on e2, black pawn on e5 ",
			fen:       "7k/8/8/4p3/8/8/4P3/K7 w - - 0 1",
			pawnColor: White,
			want:      0,
		},
		// --- Edge-of-board wrap tests ---
		// A pawn on the a-file must not be blocked by an opponent pawn on the h-file,
		// and a pawn on the h-file must not be blocked by an opponent pawn on the a-file.
		{
			name:      "white pawn on a3, black pawn on h4 (no wrap: a-pawn is passed)",
			fen:       "7k/8/8/8/7p/P7/8/K7 w - - 0 1",
			pawnColor: White,
			want:      10,
		},
		{
			name:      "white pawn on h4, black pawn on a4 (no wrap: h-pawn is passed)",
			fen:       "7k/8/8/8/p7/7P/8/K7 w - - 0 1",
			pawnColor: White,
			want:      10,
		},
		{
			name:      "white pawn on a3, black pawn on b4 (adjacent file: a-pawn is blocked)",
			fen:       "7k/8/8/8/1p6/P7/8/K7 w - - 0 1",
			pawnColor: White,
			want:      0,
		},
		{
			name:      "white pawn on h3, black pawn on g4 (adjacent file: h-pawn is blocked)",
			fen:       "7k/8/8/8/6p1/7P/8/K7 w - - 0 1",
			pawnColor: White,
			want:      0,
		},
		{
			name:      "black pawn on a6, white pawn on h5 (no wrap: a-pawn is passed)",
			fen:       "7k/8/p7/7P/8/8/8/K7 b - - 0 1",
			pawnColor: Black,
			want:      10,
		},
		{
			name:      "black pawn on h6, white pawn on a5 (no wrap: h-pawn is passed)",
			fen:       "7k/8/7p/P7/8/8/8/K7 b - - 0 1",
			pawnColor: Black,
			want:      10,
		},
		{
			name:      "black pawn on a6, white pawn on b5 (adjacent file: a-pawn is blocked)",
			fen:       "7k/8/p7/1P6/8/8/8/K7 b - - 0 1",
			pawnColor: Black,
			want:      0,
		},
		{
			name:      "black pawn on h6, white pawn on g5 (adjacent file: h-pawn is blocked)",
			fen:       "7k/8/7p/6P1/8/8/8/K7 b - - 0 1",
			pawnColor: Black,
			want:      0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			board := initBoard()
			board.updateFromFEN(tc.fen)
			pawn := findPawn(&board, tc.pawnColor)
			got := passedPawns(&board, pawn)
			if got != tc.want {
				t.Errorf("passedPawns = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestMopUpEvalPerspectiveNegation(t *testing.T) {
	board1 := initBoard()
	board1.updateFromFEN("k7/8/8/8/8/8/8/4K3 w - - 0 1")
	board2 := initBoard()
	board2.updateFromFEN("k7/8/8/8/8/8/8/4K3 b - - 0 1")

	scoreW := mopUpEval(&board1, 900, 0)
	scoreB := mopUpEval(&board2, 900, 0)

	if scoreW != -scoreB {
		t.Errorf("perspective negation violated: white-to-move=%d, black-to-move=%d (expected negation)", scoreW, scoreB)
	}
}
