package main

type TranspositionData struct {
	depth      int
	flag       int
	evaluation int
}
type TranspositionTable struct {
	table map[int64]TranspositionData
}

func (tt *TranspositionTable) init() {
	tt.table = make(map[int64]TranspositionData)
}

func (tt *TranspositionTable) push(key int64, depth int, flag int, evaluation int) {
	tt.table[key] = TranspositionData{
		depth:      depth,
		flag:       flag,
		evaluation: evaluation,
	}
}

func (tt *TranspositionTable) lookup(key int64) (int, int, int) {
	data := tt.table[key]
	return data.depth, data.flag, data.evaluation
}
