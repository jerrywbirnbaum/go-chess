package main

const repititionTableSize = 1 << 16

type RepititionEntry struct {
	hash  int64
	count int8
}

type RepititionTable struct {
	table [repititionTableSize]RepititionEntry
}

func initRepititionTable() *RepititionTable {
	return &RepititionTable{}
}

func (rt *RepititionTable) increment(key int64) bool {
	idx := key & (repititionTableSize - 1)
	if rt.table[idx].hash != key {
		rt.table[idx].hash = key
		rt.table[idx].count = 0
	}
	rt.table[idx].count++
	return rt.table[idx].count >= 3
}

func (rt *RepititionTable) decrement(key int64) bool {
	idx := key & (repititionTableSize - 1)
	if rt.table[idx].hash != key {
		return false
	}
	rt.table[idx].count--
	return rt.table[idx].count >= 3
}

func (rt *RepititionTable) isRepeat(key int64) bool {
	idx := key & (repititionTableSize - 1)
	if rt.table[idx].hash != key {
		return false
	}
	return rt.table[idx].count >= 2
}
