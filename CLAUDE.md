# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Run the full stack
```bash
docker compose up --build   # API on :8080, UI on :5173
```

### Go API (local, no Docker)
```bash
cd go-api
go run *.go                 # starts Gin server on :8080
go test ./...               # run all tests
go test -run TestName       # run a single test by name
go test -v                  # verbose output
```

### React frontend (local, no Docker)
```bash
cd go-chess-react
npm run dev                 # dev server on :5173
npm run build
npm run lint
```

## Architecture

This is a full-stack chess app where the human player (white) plays against the Go engine (black).

### Backend (`go-api/`)

All Go files are in a single `main` package. The modules are:

| File | Role |
|------|------|
| `main.go` | Gin HTTP server. Two endpoints: `POST /generate-moves` (FEN string → best move + centipawn eval) and `POST /set-timer` (set engine search time in ms) |
| `board.go` | Board state representation, FEN parsing/generation, Zobrist hashing, piece tracking |
| `pieces.go` | Piece type and color constants/utilities |
| `move_generator.go` | Legal move generation using check masks and pin masks; attack board calculation |
| `chess_engine.go` | Negamax with alpha-beta pruning and iterative deepening; search is time-limited via a goroutine timer |
| `position_evaluation.go` | Piece values, phase-aware piece-square tables (opening vs. endgame), quiescence search (captures only) |
| `transposition_table.go` | Zobrist-keyed cache to avoid re-searching seen positions |

**Engine search pipeline:** FEN in → board init → iterative deepening negamax → alpha-beta with null-move pruning → move ordering by SEE (Static Exchange Evaluation) → transposition table lookup/store → quiescence search at leaf nodes → best move + eval out.

### Frontend (`go-chess-react/`)

Single-component app (`App.tsx`) managing all game state, API calls to the Go backend, and the board UI built on `react-chessboard` (rendering) + `chess.js` (move validation and FEN handling).
