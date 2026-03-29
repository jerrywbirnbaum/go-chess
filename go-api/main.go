package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
)

type ChessRequest struct {
	FenString string `json:"fen" binding:"required"`
}

type TimerRequest struct {
	TimeSeconds int `json:"timer" binding:"required"`
}

func main() {
	go http.ListenAndServe(":8080", nil)
	board := initBoard()
	moveGenerator := MoveGenerator{board: &board}
	chessEngine := ChessEngine{moveGenerator: moveGenerator}
	chessEngine.initSearchTranspositionTable()
	chessEngine.setTimer(1000)
	initTables()

	if len(os.Args) > 1 && os.Args[1] == "uci" {
		runUCI(&chessEngine, &board)
		return
	}

	router := NewRouter(&chessEngine, &board)
	router.Run()
}
