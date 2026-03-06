package main

type TranspositionData struct {
	depth      int
	flag       int
	evaluation int
}
type TranspositionTable struct {
	table map[int64]TranspositionData
}

func initTranspositionTable() TranspositionTable {
	var tt TranspositionTable
	tt.table = make(map[int64]TranspositionData)
	return tt
}

func (tt *TranspositionTable) push(key int64, depth int, flag int, evaluation int) {
	tt.table[key] = TranspositionData{
		depth:      depth,
		flag:       flag,
		evaluation: evaluation,
	}
}

func (tt *TranspositionTable) lookup(key int64) (bool, int, int, int) {
	data, ok := tt.table[key]
	if !ok {
		return false, 0, 0, 0
	}
	return true, data.depth, data.flag, data.evaluation
}
