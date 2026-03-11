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

type TimerRequest struct {
	TimeSeconds int `json:"timer" binding:"required"`
}

func main() {
	board := initBoard()
	moveGenerator := MoveGenerator{board: &board}
	chessEngine := ChessEngine{moveGenerator: moveGenerator}
	chessEngine.setTimer(1000)

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

		engine_move, positions_evaluated, engine_evauluation := chessEngine.bestMove()
		c.JSON(http.StatusOK, gin.H{
			"status":              "received",
			"start_square":        engine_move.startSquare,
			"end_square":          engine_move.endSquare,
			"promotion":           engine_move.promotion,
			"fen":                 receivedFen,
			"positions_evaluated": positions_evaluated,
			"engine_evaluation":   engine_evauluation,
		})
	})

	r.POST("/set-timer", func(c *gin.Context) {
		var json TimerRequest

		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		timeSeconds := json.TimeSeconds
		chessEngine.setTimer(timeSeconds)

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"timer":  timeSeconds,
		})
	})

	r.Run()
}
