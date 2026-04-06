package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/joho/godotenv"
)

type ChessRequest struct {
	FenString string `json:"fen" binding:"required"`
}

type TimerRequest struct {
	TimeSeconds int `json:"timer" binding:"required"`
}

func init() {
	for sq := range 64 {
		r, c := sq/8, sq%8
		knightAttacks[sq] = leaperAttackBits(r, c, knightOffsets[:])
		kingAttacks[sq] = leaperAttackBits(r, c, kingOffsets[:])
		bishopMasks[sq] = sliderMaskBits(r, c, diagonalDirs[:])
		rookMasks[sq] = sliderMaskBits(r, c, straightDirs[:])
	}
	bishopMagicLookup = createBishopLookupTable()
	rookMagicLookup = createRookLookupTable()
	initTables()
}

var multithreading bool = false

func main() {
	godotenv.Load()
	go http.ListenAndServe(":8080", nil)
	board := initBoard()
	chessEngine := ChessEngine{ctx: SearchContext{board: &board}}
	chessEngine.initSearchTranspositionTable()
	chessEngine.setTimer(1000)

	if len(os.Args) > 2 && os.Args[2] == "multithreading" {
		multithreading = true
	}

	if len(os.Args) > 1 && os.Args[1] == "uci" {
		runUCI(&chessEngine, &board)
		return
	}

	router := NewRouter(&chessEngine, &board)
	router.Run()
}
