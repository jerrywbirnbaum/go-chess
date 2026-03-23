package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func runUCI(engine *ChessEngine, board *Board) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			continue
		}
		switch tokens[0] {
		case "uci":
			fmt.Println("id name Go Chess")
			fmt.Println("id author Jerry")
			fmt.Println("uciok")
		case "isready":
			fmt.Println("readyok")
		case "ucinewgame":
			engine.initSearchTranspositionTable()
			board.updateFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
		case "position":
			handlePosition(board, tokens[1:])
		case "go":
			handleGo(engine, board, tokens[1:])
		case "quit":
			return
		}
	}
}

func handlePosition(board *Board, args []string) {
	if len(args) == 0 {
		return
	}

	movesIdx := -1
	if args[0] == "startpos" {
		board.updateFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
		for i, arg := range args {
			if arg == "moves" {
				movesIdx = i + 1
				break
			}
		}
	} else if args[0] == "fen" {
		fenParts := []string{}
		i := 1
		for i < len(args) && args[i] != "moves" {
			fenParts = append(fenParts, args[i])
			i++
		}
		board.updateFromFEN(strings.Join(fenParts, " "))
		if i < len(args) && args[i] == "moves" {
			movesIdx = i + 1
		}
	}

	if movesIdx >= 0 && movesIdx <= len(args) {
		for _, uciMove := range args[movesIdx:] {
			move := parseMoveFromUCI(board, uciMove)
			board.makeMove(&move)
		}
	}
}

func handleGo(engine *ChessEngine, board *Board, args []string) {
	wtime, btime, winc, binc, movetime := 0, 0, 0, 0, 0
	for i := 0; i < len(args)-1; i++ {
		v, err := strconv.Atoi(args[i+1])
		if err != nil {
			continue
		}
		switch args[i] {
		case "movetime":
			movetime = v
		case "wtime":
			wtime = v
		case "btime":
			btime = v
		case "winc":
			winc = v
		case "binc":
			binc = v
		}
	}

	timeMs := 1000
	if board.isWhiteTurn && wtime > 0 {
		timeMs = wtime/20 + winc
	} else if !board.isWhiteTurn && btime > 0 {
		timeMs = btime/20 + binc
	}
	timeMs += movetime / 2

	engine.setTimer(timeMs)
	move, nodes, eval := engine.bestMove()

	fmt.Printf("info nodes %d score cp %d\n", nodes, eval)

	moveStr := move.startSquare + move.endSquare
	if move.isPromotion {
		moveStr += move.promotion
	}
	fmt.Println("bestmove " + moveStr)
}

func parseMoveFromUCI(board *Board, uciMove string) Move {
	if len(uciMove) < 4 {
		return Move{}
	}

	fromRow, fromCol := fromSquare(uciMove[0:2])
	toRow, toCol := fromSquare(uciMove[2:4])

	startPiece := board.board[fromRow][fromCol]
	endPiece := board.board[toRow][toCol]

	move := Move{
		startSquare: Square{row: fromRow, col: fromCol, piece: startPiece},
		endSquare:   Square{row: toRow, col: toCol, piece: endPiece},
	}

	if len(uciMove) == 5 {
		switch uciMove[4] {
		case 'q':
			move.promotionPieceType = Queen
		case 'r':
			move.promotionPieceType = Rook
		case 'b':
			move.promotionPieceType = Bishop
		case 'n':
			move.promotionPieceType = Knight
		}
	}

	return move
}
