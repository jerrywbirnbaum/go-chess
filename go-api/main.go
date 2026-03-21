package main

type ChessRequest struct {
	FenString string `json:"fen" binding:"required"`
}

type TimerRequest struct {
	TimeSeconds int `json:"timer" binding:"required"`
}

func main() {
	board := initBoard()
	moveGenerator := MoveGenerator{board: &board}
	chessEngine := ChessEngine{moveGenerator: moveGenerator}
	chessEngine.initSearchTranspositionTable()
	chessEngine.setTimer(1000)
	router := NewRouter(&chessEngine, &board)

	router.Run()
}
