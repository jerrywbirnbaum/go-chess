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
  const [positionsEvaluated, setPositionsEvaluated] = useState(0);
  const [engineEvaluation, setEngineEvaluation] = useState(0);

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

  async function makeEngineMove() {
    if (chessGame.isGameOver()) {
      setLoading(false);
      return;
    }

    setLoading(true);
    const currentFen = chessGame.fen();

    try {
      const response = await fetch(
        "/api/generate-moves",
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
      setPositionsEvaluated(data.positions_evaluated)
      setEngineEvaluation(data.engine_evaluation)
    } catch (error) {
      console.log(error);
    }

    setChessPosition(chessGame.fen());
    setLoading(false);
  }

  function onPieceDrop({
    sourceSquare,
    targetSquare
  }: PieceDropHandlerArgs) {
    if (chessGame.isGameOver()) {
      return false;
    }

    if (!targetSquare) {
      return false;
    }

    try {
      chessGame.move({
        from: sourceSquare,
        to: targetSquare,
        promotion: 'q' // always promote to a queen for example simplicity
      });

      setChessPosition(chessGame.fen());

      if (!chessGame.isGameOver()) {
        engineMoveTimeoutRef.current = window.setTimeout(makeEngineMove, 1);
      }

      return true;
    } catch {
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
      <h3 style={{ margin: '0 0 1rem' }}>Positions Evaluated: {positionsEvaluated}</h3>
      <h3 style={{ margin: '0 0 1rem' }}>Engine Evaluation: {engineEvaluation}</h3>
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
