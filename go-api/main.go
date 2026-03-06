package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type ChessRequest struct {
	FenString string `json:"fen" binding:"required"`
}

func main() {
	board := initBoard()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://0.0.0.0:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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

		moveGenerator := MoveGenerator{board: &board}
		random_move, positions_evaluated, engine_evauluation := moveGenerator.bestMove()
		fmt.Println("engine_evauluation")
		fmt.Println(engine_evauluation)
		c.JSON(http.StatusOK, gin.H{
			"status":              "received",
			"start_square":        random_move.startSquare,
			"end_square":          random_move.endSquare,
			"promotion":           "q",
			"fen":                 receivedFen,
			"positions_evaluated": positions_evaluated,
			"engine_evaluation":   engine_evauluation,
		})
	})

	r.Run()
}
