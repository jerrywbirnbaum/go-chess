package main

import "testing"

var benchmarkEvalSink int
var benchmarkMoveSink MoveString

func BenchmarkSearchOnlyCapturesForce_StartPosition(b *testing.B) {
	board := initBoard()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		localBoard := board
		benchmarkEvalSink = searchOnlyCapturesForce(-20000, 20000, &localBoard)
	}
}

func BenchmarkSearchOnlyCapturesForce_TacticalPosition(b *testing.B) {
	board := initBoard()
	board.updateFromFEN("r2q1rk1/pp2bppp/2n1pn2/2bp4/2B5/2NP1NP1/PP2PPBP/R1BQ1RK1 w - - 0 9")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		localBoard := board
		benchmarkEvalSink = searchOnlyCapturesForce(-20000, 20000, &localBoard)
	}
}

func BenchmarkBestMove_StartPosition(b *testing.B) {
	board := initBoard()
	moveGenerator := MoveGenerator{board: &board}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkMoveSink, _, _ = moveGenerator.bestMove()
	}
}

func BenchmarkBestMove_MidgamePosition(b *testing.B) {
	board := initBoard()
	board.updateFromFEN("r3k2r/pp1n1ppp/2p1pn2/2bp4/2B5/2NP1NP1/PPQ1PPBP/R3K2R w KQkq - 0 11")
	moveGenerator := MoveGenerator{board: &board}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkMoveSink, _, _ = moveGenerator.bestMove()
	}
}
