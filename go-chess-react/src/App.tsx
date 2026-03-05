import { Chessboard } from 'react-chessboard';
import { Chess } from 'chess.js';
import { useEffect, useRef, useState } from 'react';
import { Loader } from 'react-overlay-loader';

import 'react-overlay-loader/styles.css';

const STORED_FEN_KEY = 'go-chess:fen';

type PieceDropHandlerArgs = {
  sourceSquare: string;
  targetSquare: string | null;
};

function App() {
  const getInitialGame = () => {
    const storedFen = localStorage.getItem(STORED_FEN_KEY);
    if (!storedFen) {
      return new Chess();
    }

    try {
      return new Chess(storedFen);
    } catch {
      return new Chess();
    }
  };

  const chessGameRef = useRef(getInitialGame());
  const engineMoveTimeoutRef = useRef<number | null>(null);
  const chessGame = chessGameRef.current;
  const getBoardSize = () => {
    const horizontalPadding = 32;
    const verticalReservedSpace = 180;
    const availableWidth = window.innerWidth - horizontalPadding;
    const availableHeight = window.innerHeight - verticalReservedSpace;

    return Math.max(280, Math.min(availableWidth, availableHeight));
  };

  // track the current position of the chess game in state to trigger a re-render of the chessboard
  const [chessPosition, setChessPosition] = useState(chessGame.fen());
  const [isLoading, setLoading] = useState(false);
  const [boardSize, setBoardSize] = useState(getBoardSize);

  useEffect(() => {
    localStorage.setItem(STORED_FEN_KEY, chessPosition);
  }, [chessPosition]);

  useEffect(() => {
    const handleResize = () => {
      setBoardSize(getBoardSize());
    };

    window.addEventListener('resize', handleResize);
    return () => {
      window.removeEventListener('resize', handleResize);
      if (engineMoveTimeoutRef.current) {
        window.clearTimeout(engineMoveTimeoutRef.current);
      }
    };
  }, []);

  function getGameStatus(): string {
    if (chessGame.isCheckmate()) {
      const winner = chessGame.turn() === 'w' ? 'Black' : 'White';
      return `Checkmate. ${winner} wins.`;
    }
    if (chessGame.isStalemate()) {
      return 'Stalemate. Draw.';
    }
    if (chessGame.isDraw()) {
      return 'Draw.';
    }
    if (chessGame.inCheck()) {
      const sideToMove = chessGame.turn() === 'w' ? 'White' : 'Black';
      return `${sideToMove} is in check.`;
    }
    const sideToMove = chessGame.turn() === 'w' ? 'White' : 'Black';
    return `${sideToMove} to move.`;
  }

  // make a random "CPU" move
  async function makeEngineMove() {
    if (chessGame.isGameOver()) {
      setLoading(false);
      return;
    }

    setLoading(true);
    const currentFen = chessGame.fen();

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

      if (!data.start_square || !data.end_square) {
        setLoading(false);
        return;
      }

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
    if (chessGame.isGameOver()) {
      return false;
    }

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

      // make engine move after a short delay unless game is over
      if (!chessGame.isGameOver()) {
        engineMoveTimeoutRef.current = window.setTimeout(makeEngineMove, 5);
      }

      // return true as the move was successful
      return true;
    } catch {
      // return false as the move was not successful
      return false;
    }
  }

  function onNewGame() {
    if (engineMoveTimeoutRef.current) {
      window.clearTimeout(engineMoveTimeoutRef.current);
      engineMoveTimeoutRef.current = null;
    }
    chessGame.reset();
    setLoading(false);
    setChessPosition(chessGame.fen());
  }

  // set the chessboard options
  const chessboardOptions = {
    position: chessPosition,
    onPieceDrop,
    id: 'play-vs-random',
    allowDragging: !chessGame.isGameOver(),
  };

  return (
    <div
      style={{
        width: '100%',
        minHeight: '100vh',
        margin: 0,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        padding: '1rem',
        boxSizing: 'border-box',
      }}
      className="chessboard-container"
    >
      <h3 style={{ margin: '0 0 1rem' }}>{getGameStatus()}</h3>
      <div
        style={{
          display: 'flex',
          alignItems: 'flex-start',
          gap: '1rem',
          width: '100%',
          justifyContent: 'center',
          flexWrap: 'wrap',
        }}
      >
        <div style={{ width: `${boardSize}px`, height: `${boardSize}px`, maxWidth: '100%' }}>
          <Chessboard options={chessboardOptions} />
        </div>
        <button
          type="button"
          onClick={onNewGame}
          style={{
            padding: '0.5rem 1rem',
            fontSize: '1rem',
            cursor: 'pointer',
            alignSelf: 'center',
            whiteSpace: 'nowrap',
          }}
        >
          New Game
        </button>
      </div>
      <Loader fullPage loading={isLoading} />
    </div>
  )
}


export default App
