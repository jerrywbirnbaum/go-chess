#!/bin/bash
set -e

cd go-api && go build -o go-chess . && cd ..

fastchess \
      -engine cmd="./go-api/go-chess" args="multithreading uci" name=GoChess \
      -engine cmd="stockfish" name=Stockfish \
      -each tc=30+0.3 -openings file=openings.epd format=epd \
      -rounds 10 -pgnout append=false notation=san nodes=true file=./games/round1.pgn \
      -resign movecount=60 score=1500 -recover &

FASTCHESS_PID=$!

# Kill any leftover pprof server from a previous run
lsof -ti:9090 | xargs kill -9 2>/dev/null || true

# Wait for go-chess pprof server to initialize
sleep 2
go tool pprof -http=:9090 "http://localhost:8080/debug/pprof/profile?seconds=30" &

wait $FASTCHESS_PID
