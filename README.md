# Go Chess project
## About
This is a chess engine project created from scratch.
The UI uses react-chessboard and chess.js as a simple way to play vs the engine.
Currently the player can only play as white.

The backend is written in Go and uses Gin as its HTTP framework.
It generates all legal moves then searches through them using a negamax strategy.
To generate moves quickly the move generator uses check and pin masks to ensure that move do not leave the king in check.
To speedup searching and avoid searching transpositions a transposition table is created with the Zobrish hash of each position.
The search has a set time limit and uses iterative deepening to find the best move given the time limit.
The board is evaluated using piece values and position adjustments.

## Running locally
The project is easiest to run using docker.
To run locally first install Docker Desktop
https://www.docker.com/products/docker-desktop/

Then start the docker containers from the main project directory:
docker compose up --build

The UI then can be accessed at:
http://localhost:5173
