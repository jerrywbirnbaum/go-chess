package main

type TranspositionData struct {
	depth      int
	flag       int
	evaluation int
	bestMove   uint64
}
type TranspositionTable struct {
	table map[int64]TranspositionData
}

func initTranspositionTable() TranspositionTable {
	var tt TranspositionTable
	tt.table = make(map[int64]TranspositionData)
	return tt
}

func (tt *TranspositionTable) push(key int64, depth int, flag int, evaluation int, moveBits uint64) {
	tt.table[key] = TranspositionData{
		depth:      depth,
		flag:       flag,
		evaluation: evaluation,
		bestMove:   moveBits,
	}
}

func (tt *TranspositionTable) lookup(key int64) (bool, int, int, int, uint64) {
	data, ok := tt.table[key]
	if !ok {
		return false, 0, 0, 0, 0
	}
	return true, data.depth, data.flag, data.evaluation, data.bestMove
}

// TODO: Improve pack function
func packMove(move Move) uint64 {
	return move.getMoveBits()
}

func unpackMove(moveBits uint64) Move {
	move := Move{}
	move.setMoveBits(moveBits)
	return move
}

func comparePackedMoves(move1 uint64, move2 uint64) bool {
	mask := uint64(0b000011101000000000000000000000001111110000111111)
	move1 &= mask
	move2 &= mask
	return move1 == move2
}
