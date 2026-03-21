package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func NewRouter(chessEngine *ChessEngine, board *Board) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://0.0.0.0:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/new-game", func(c *gin.Context) {
		chessEngine.initSearchTranspositionTable()
		board.updateFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

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
	return r
}
