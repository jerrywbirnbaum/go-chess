package main

type TranspositionData struct {
	key        uint64
	depth      int
	flag       int
	evaluation int
	bestMove   uint64
}

const ttSize = 1 << 20

type TranspositionTable struct {
	table [ttSize]TranspositionData
}

func initTranspositionTable() TranspositionTable {
	return TranspositionTable{}
}

func (tt *TranspositionTable) push(key int64, depth int, flag int, evaluation int, moveBits uint64) {
	idx := int(uint64(key) & (ttSize - 1))
	tt.table[idx] = TranspositionData{
		key:        uint64(key) ^ moveBits,
		depth:      depth,
		flag:       flag,
		evaluation: evaluation,
		bestMove:   moveBits,
	}
}

func (tt *TranspositionTable) lookup(key int64) (bool, int, int, int, uint64) {
	idx := int(uint64(key) & (ttSize - 1))
	data := tt.table[idx]
	if data.key != uint64(key)^data.bestMove {
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
