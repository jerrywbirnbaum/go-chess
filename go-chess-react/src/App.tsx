import { Chessboard } from 'react-chessboard';
import { Chess } from 'chess.js';
import { useRef, useState } from 'react';
import { Loader } from 'react-overlay-loader';

import 'react-overlay-loader/styles.css';
function App() {
  const chessGameRef = useRef(new Chess());
  const chessGame = chessGameRef.current;

  // track the current position of the chess game in state to trigger a re-render of the chessboard
  const [chessPosition, setChessPosition] = useState(chessGame.fen());
  const [isLoading, setLoading] = useState(false);

  // make a random "CPU" move
  async function makeEngineMove() {
    setLoading(true);
    const currentFen = chessGame.fen()
    if (chessGame.isGameOver()) {
      return;
    }

    try {
      const response = await fetch(
        "http://localhost:8080/generate-moves",
        {
          method: "POST",
          headers: {
            "Content-Type":
              "application/json",
          },
          body: JSON.stringify({
            fen: currentFen,
          }),
        },
      );
      const data = await response.json();
      console.log(data);

      chessGame.move({
        from: data.start_square,
        to: data.end_square,
        promotion: "q",
      });
    } catch (error) {
      console.log(error);
    }

    // pick a random move
    // const randomMove = possibleMoves[Math.floor(Math.random() * possibleMoves.length)];

    // make the move
    // chessGame.move(randomMove);

    // update the position state
    setChessPosition(chessGame.fen());
    setLoading(false);
  }

  // handle piece drop
  function onPieceDrop({
    sourceSquare,
    targetSquare
  }: PieceDropHandlerArgs) {
    // type narrow targetSquare potentially being null (e.g. if dropped off board)
    if (!targetSquare) {
      return false;
    }

    // try to make the move according to chess.js logic
    try {
      chessGame.move({
        from: sourceSquare,
        to: targetSquare,
        promotion: 'q' // always promote to a queen for example simplicity
      });

      // update the position state upon successful move to trigger a re-render of the chessboard
      setChessPosition(chessGame.fen());

      // make random cpu move after a short delay
      setTimeout(makeEngineMove, 5);

      // return true as the move was successful
      return true;
    } catch {
      // return false as the move was not successful
      return false;
    }
  }

  // set the chessboard options
  const chessboardOptions = {
    position: chessPosition,
    onPieceDrop,
    id: 'play-vs-random',

  };

  return (
    <div style={{ width: '50vw' }} className="chessboard-container">
      <Chessboard options={chessboardOptions} />);
      <Loader fullPage loading={isLoading} />
    </div>
  )
}


export default App
