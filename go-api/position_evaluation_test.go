package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	initTables()
	os.Exit(m.Run())
}

// TestPestoEvalPerspectiveFlip verifies that flipping whose turn it is
// always negates the score exactly, for positions with varying material.
func TestPestoEvalPerspectiveFlip(t *testing.T) {
	positions := []struct {
		name string
		fen  string
	}{
		{"kings only", "4k3/8/8/8/8/8/8/4K3"},
		{"white queen advantage", "4k3/8/8/8/8/8/8/4KQ2"},
		{"equal queens", "q3k3/8/8/8/8/8/8/4KQ2"},
		{"starting position", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"},
		{"rook and pawns", "4k3/ppp5/8/8/8/8/5PPP/4K2R"},
	}
	for _, tc := range positions {
		t.Run(tc.name, func(t *testing.T) {
			boardW := initBoard()
			boardW.updateFromFEN(tc.fen + " w - - 0 1")
			boardB := initBoard()
			boardB.updateFromFEN(tc.fen + " b - - 0 1")

			scoreW := pestoEval(boardW)
			scoreB := pestoEval(boardB)

			if scoreW != -scoreB {
				t.Errorf("perspective flip violated: white=%d, black=%d (expected negation)", scoreW, scoreB)
			}
		})
	}
}

// TestPestoEvalMaterialAdvantage checks the exact score for positions with
// a material imbalance from both sides' perspective.
func TestPestoEvalMaterialAdvantage(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want int
	}{
		{
			name: "white extra queen, white to move",
			fen:  "4k3/8/8/8/8/8/8/4KQ2 w - - 0 1",
			want: 1064,
		},
		{
			name: "white extra queen, black to move",
			fen:  "4k3/8/8/8/8/8/8/4KQ2 b - - 0 1",
			want: -1064,
		},
		{
			name: "black extra rook, black to move",
			fen:  "4k2r/8/8/8/8/8/8/4K3 b - - 0 1",
			want: 456,
		},
		{
			name: "black extra rook, white to move",
			fen:  "4k2r/8/8/8/8/8/8/4K3 w - - 0 1",
			want: -456,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			board := initBoard()
			board.updateFromFEN(tc.fen)
			got := pestoEval(board)
			if got != tc.want {
				t.Errorf("pestoEval(%q) = %d, want %d", tc.fen, got, tc.want)
			}
		})
	}
}

// TestPestoEvalKingsOnly verifies the evaluation for a kings-only endgame,
// where midGamePhase is minimal (only king contributions).
func TestPestoEvalKingsOnly(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want int
	}{
		{
			name: "kings only, white to move",
			fen:  "4k3/8/8/8/8/8/8/4K3 w - - 0 1",
			want: 64,
		},
		{
			name: "kings only, black to move",
			fen:  "4k3/8/8/8/8/8/8/4K3 b - - 0 1",
			want: -64,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			board := initBoard()
			board.updateFromFEN(tc.fen)
			got := pestoEval(board)
			if got != tc.want {
				t.Errorf("pestoEval(%q) = %d, want %d", tc.fen, got, tc.want)
			}
		})
	}
}

// TestPestoEvalTablesInitialized verifies that initTables populates the
// lookup tables — if tables are zeroed, all positions would evaluate to 0.
func TestPestoEvalTablesInitialized(t *testing.T) {
	board := initBoard()
	board.updateFromFEN("4k3/8/8/8/8/8/8/4KQ2 w - - 0 1")
	score := pestoEval(board)
	if score == 0 {
		t.Errorf("pestoEval returned 0 for a position with a queen advantage — mg_table may not be initialized")
	}
}
