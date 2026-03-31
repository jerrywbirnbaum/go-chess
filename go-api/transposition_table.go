package main

// TTEntry is a 10-byte transposition table entry.
type TTEntry struct {
	key16    uint16 // upper 16 bits of zobrist key (collision check)
	move16   uint16 // packed move: bits 0-5 = startRow+Col, 6-11 = endRow+Col, 12 = isPromo, 13-15 = promoType
	value16  int16  // search value
	depth8   int8   // search depth
	genflags uint8  // bits 7-2: generation (6 bits), bits 1-0: flag (0=exact, 1=lower, 2=upper)
}

// TTCluster holds 3 entries and pads to 32 bytes (two clusters fit in one cache line).
type TTCluster struct {
	entries [3]TTEntry // 30 bytes
	_pad    [2]byte
}

const ttClusterCount = 1 << 22 // 4M clusters × 3 entries = 12M entries, ~128MB

type TranspositionTable struct {
	clusters   [ttClusterCount]TTCluster
	generation uint8
}

func initTranspositionTable() TranspositionTable {
	return TranspositionTable{}
}

func (tt *TranspositionTable) push(key int64, depth int, flag int, evaluation int, moveBits uint16) {
	idx := uint64(key) & (ttClusterCount - 1)
	cluster := &tt.clusters[idx]
	key16 := uint16(key >> 48)

	worstIdx := 0
	worstScore := int(1<<15) - 1
	for i := range cluster.entries {
		e := &cluster.entries[i]
		if e.key16 == key16 {
			cluster.entries[i] = TTEntry{
				key16:    key16,
				move16:   moveBits,
				value16:  int16(evaluation),
				depth8:   int8(depth),
				genflags: (tt.generation << 2) | uint8(flag),
			}
			return
		}
		score := int(e.depth8) - 4*int((tt.generation-e.genflags>>2)&0x3F)
		if score < worstScore {
			worstScore = score
			worstIdx = i
		}
	}
	cluster.entries[worstIdx] = TTEntry{
		key16:    key16,
		move16:   moveBits,
		value16:  int16(evaluation),
		depth8:   int8(depth),
		genflags: (tt.generation << 2) | uint8(flag),
	}
}

func (tt *TranspositionTable) lookup(key int64) (bool, int, int, int, uint16) {
	idx := uint64(key) & (ttClusterCount - 1)
	cluster := &tt.clusters[idx]
	key16 := uint16(key >> 48)

	for i := range cluster.entries {
		e := &cluster.entries[i]
		if e.key16 == key16 {
			cluster.entries[i].genflags = (tt.generation << 2) | (e.genflags & 0x3)
			return true, int(e.depth8), int(e.genflags & 0x3), int(e.value16), e.move16
		}
	}
	return false, 0, 0, 0, 0
}

// packMove compresses a Move to the 16 essential bits needed for TT storage.
func packMove(move Move) uint16 {
	bits := move.getMoveBits()
	return uint16(bits&0b111111) |
		uint16((bits>>10)&0b111111)<<6 |
		uint16((bits>>39)&1)<<12 |
		uint16((bits>>41)&0b111)<<13
}

// unpackTTMove reconstructs the relevant uint64 bits from a packed uint16 move,
// in the same format expected by comparePackedMoves.
func unpackTTMove(packed uint16) uint64 {
	return uint64(packed&0b111111) |
		uint64((packed>>6)&0b111111)<<10 |
		uint64((packed>>12)&1)<<39 |
		uint64((packed>>13)&0b111)<<41
}

func comparePackedMoves(move1 uint64, move2 uint64) bool {
	mask := uint64(0b000011101000000000000000000000001111110000111111)
	move1 &= mask
	move2 &= mask
	return move1 == move2
}
