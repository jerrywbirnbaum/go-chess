package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	initTables()
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

			if scoreW != -scoreB {
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
