import { Chessboard } from 'react-chessboard';
import { Chess } from 'chess.js';
import { useEffect, useRef, useState } from 'react';
import type { FormEvent } from 'react';
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
  const [fenInput, setFenInput] = useState(chessGame.fen());
  const [fenError, setFenError] = useState('');
  const [timerInput, setTimerInput] = useState('1');
  const [timerError, setTimerError] = useState('');
  const [timerStatus, setTimerStatus] = useState('Engine timer is 1 second.');
  const [isSubmittingTimer, setIsSubmittingTimer] = useState(false);

  useEffect(() => {
    localStorage.setItem(STORED_FEN_KEY, chessPosition);
    setFenInput(chessPosition);
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

  async function postTimer(event?: FormEvent<HTMLFormElement>) {
    event?.preventDefault();

    const parsedTimer = Number(timerInput);
    if (!Number.isInteger(parsedTimer) || parsedTimer < 1) {
      setTimerError('Enter a whole number of seconds greater than 0.');
      setTimerStatus('');
      return;
    }

    setIsSubmittingTimer(true);
    setTimerError('');
    setTimerStatus('');

    try {
      const response = await fetch(
        "/api/set-timer",
        {
          method: "POST",
          headers: {
            "Content-Type":
              "application/json",
          },
          body: JSON.stringify({
            timer: parsedTimer,
          }),
        },
      );

      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }

      const data = await response.json();
      setTimerInput(String(data.timer));
      setTimerStatus(`Engine timer updated to ${data.timer} second${data.timer === 1 ? '' : 's'}.`);
    } catch (error) {
      console.log(error);
      setTimerError('Unable to update the engine timer.');
    } finally {
      setIsSubmittingTimer(false);
    }
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
    setFenError('');
    setChessPosition(chessGame.fen());
  }

  function onImportFen() {
    if (engineMoveTimeoutRef.current) {
      window.clearTimeout(engineMoveTimeoutRef.current);
      engineMoveTimeoutRef.current = null;
    }

    try {
      chessGame.load(fenInput.trim());
      setLoading(false);
      setFenError('');
      setPositionsEvaluated(0);
      setEngineEvaluation(0);
      setChessPosition(chessGame.fen());
    } catch {
      setFenError('Invalid FEN');
    }
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
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'stretch',
            gap: '1rem',
            width: '320px',
            maxWidth: '100%',
          }}
        >
          <h3 style={{ margin: 0 }}>{getGameStatus()}</h3>
          <h3 style={{ margin: 0 }}>Positions Evaluated: {positionsEvaluated}</h3>
          <h3 style={{ margin: 0 }}>Engine Evaluation: {engineEvaluation}</h3>

          <button
            type="button"
            onClick={onNewGame}
            style={{
              padding: '0.5rem 1rem',
              fontSize: '1rem',
              cursor: 'pointer',
              whiteSpace: 'nowrap',
            }}
          >
            New Game
          </button>
          <button
            type="button"
            onClick={makeEngineMove}
            style={{
              padding: '0.5rem 1rem',
              fontSize: '1rem',
              cursor: 'pointer',
              whiteSpace: 'nowrap',
            }}
          >
            Engine Move
          </button>

          <form
            onSubmit={postTimer}
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'stretch',
              gap: '0.5rem',
            }}
          >
            <label
              htmlFor="engine-timer"
              style={{
                fontSize: '0.95rem',
              }}
            >
              Engine timer
            </label>
            <input
              id="engine-timer"
              type="number"
              min="1"
              step="1"
              inputMode="numeric"
              value={timerInput}
              onChange={(event) => {
                setTimerInput(event.target.value);
                if (timerError) {
                  setTimerError('');
                }
                if (timerStatus) {
                  setTimerStatus('');
                }
              }}
              placeholder="1"
              style={{
                padding: '0.5rem 0.75rem',
                fontSize: '0.95rem',
                boxSizing: 'border-box',
              }}
            />
            <button
              type="submit"
              disabled={isSubmittingTimer}
              style={{
                padding: '0.5rem 1rem',
                fontSize: '1rem',
                cursor: isSubmittingTimer ? 'wait' : 'pointer',
                whiteSpace: 'nowrap',
              }}
            >
              {isSubmittingTimer ? 'Updating...' : 'Set Engine Time'}
            </button>
            {timerError ? (
              <p style={{ margin: 0, color: '#b00020', fontSize: '0.9rem' }}>{timerError}</p>
            ) : null}
            {timerStatus ? (
              <p style={{ margin: 0, color: '#1f5d2f', fontSize: '0.9rem' }}>{timerStatus}</p>
            ) : null}
          </form>

          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              gap: '0.5rem',
            }}
          >
            <input
              type="text"
              value={fenInput}
              onChange={(event) => {
                setFenInput(event.target.value);
                if (fenError) {
                  setFenError('');
                }
              }}
              placeholder="Paste a FEN string"
              style={{
                padding: '0.5rem 0.75rem',
                fontSize: '0.95rem',
                boxSizing: 'border-box',
              }}
            />
            <button
              type="button"
              onClick={onImportFen}
              style={{
                padding: '0.5rem 1rem',
                fontSize: '1rem',
                cursor: 'pointer',
                whiteSpace: 'nowrap',
              }}
            >
              Import FEN
            </button>
          </div>
          {fenError ? (
            <p style={{ margin: '0 0 1rem', color: '#b00020' }}>{fenError}</p>
          ) : null}
        </div>
      </div>
      <Loader fullPage loading={isLoading} />
    </div>
  )
}


export default App
