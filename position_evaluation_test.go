package main

import "testing"

func TestBasicEval(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		want int
	}{
		{
			name: "starting position is equal",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			want: 0,
		},
		{
			name: "white to move counts own queen positively",
			fen:  "4k3/8/8/8/8/8/8/4KQ2 w - - 0 1",
			want: 890,
		},
		{
			name: "black to move sees same position negatively",
			fen:  "4k3/8/8/8/8/8/8/4KQ2 b - - 0 1",
			want: -890,
		},
		{
			name: "weighted material sum from white perspective",
			fen:  "q3k3/1pp5/8/8/8/8/3P4/RBN1K3 w - - 0 1",
			want: 190,
		},
		{
			name: "weighted material sum flips for black perspective",
			fen:  "q3k3/1pp5/8/8/8/8/3P4/RBN1K3 b - - 0 1",
			want: -190,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			board := initBoard()
			board.updateFromFEN(tc.fen)

			got := basicEval(board)
			if got != tc.want {
				t.Fatalf("basicEval(%q) = %v, want %v", tc.fen, got, tc.want)
			}
		})
	}
}
