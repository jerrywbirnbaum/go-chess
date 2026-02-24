package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ChessRequest struct {
	FenString string `json:"fen_string" binding:"required"`
}

func main() {
	board := Board{
		board: [8][8]Piece{
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
			{newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*'), newPiece('*')},
		},
		isWhiteTurn: true,
	}
	// board.moveAlgebraicNotation("fxe5")
	board.updateFromFEN("rnbqkbnr/pppppppp/PPP3PP/8/8/8/8/RNBQKBNR w KQkq - 0 1")
	// board.printBoard()

	// board.updateFromFEN("r2qk3/2pp1p2/3p2b1/8/1P2R3/8/P7/1N2K3 b - - 0 1")
	// board.printBoard()
	// moveGenerator := MoveGenerator{board: board}
	// for _, move := range moveGenerator.generateMoves() {
	// 	fmt.Println("Move")
	// 	fmt.Printf("%d%d\n", move.startSquare.row, move.startSquare.col)
	// 	fmt.Printf("%d%d\n", move.endSquare.row, move.endSquare.col)
	// }

	r := gin.Default()

	r.POST("/generate-moves", func(c *gin.Context) {
		var json ChessRequest

		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		receivedFen := json.FenString
		fmt.Println("FEN:")
		fmt.Println(receivedFen)

		board.updateFromFEN(receivedFen)

		moveGenerator := MoveGenerator{board: board}
		moves := moveGenerator.generateMoves()
		for _, move := range moves {
			fmt.Println("Move")
			fmt.Printf("%d%d\n", move.startSquare.row, move.startSquare.col)
			fmt.Printf("%d%d\n", move.endSquare.row, move.endSquare.col)
		}
		random_move := moveGenerator.randomMove()
		fmt.Println("randommove")
		fmt.Printf("%d%d\n", random_move.startSquare.col)
		c.JSON(http.StatusOK, gin.H{
			"status":           "received",
			"start_square_col": random_move.startSquare.col,
			"start_square_row": random_move.startSquare.row,
			"end_square_col":   random_move.endSquare.col,
			"end_square_row":   random_move.endSquare.row,
			"promotion":        "q",
			"fen":              receivedFen,
		})
	})

	r.Run()
}
