package main

type RepititionTable struct {
	table map[int64]int
}

func initRepititionTable() *RepititionTable {
	var rt RepititionTable
	rt.table = make(map[int64]int)
	return &rt
}

func (rt *RepititionTable) increment(key int64) bool {
	value, ok := rt.table[key]
	if ok {
		rt.table[key] = value + 1
	} else {
		rt.table[key] = 1
	}
	return value+1 >= 3
}

func (rt *RepititionTable) decrement(key int64) bool {
	value, ok := rt.table[key]
	if ok {
		rt.table[key] = value - 1
	}

	return value-1 >= 3
}
