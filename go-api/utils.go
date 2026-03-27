package main

func bitboardAddOne(bitboard uint64, row int, col int) uint64 {
	row = 7 - row
	col = 7 - col
	bitboard |= uint64(1) << uint(row*8+col)
	return bitboard
}

func bitboardRemoveOne(bitboard uint64, row int, col int) uint64 {
	row = 7 - row
	col = 7 - col
	bitboard &^= uint64(1) << uint(row*8+col)
	return bitboard
}

func bitboardCheckOne(bitboard uint64, row int, col int) bool {
	row = 7 - row
	col = 7 - col
	return (bitboard>>uint(row*8+col))&1 == 1
}

func inBounds(row int, col int) bool {
	return row >= 0 && row <= 7 && col >= 0 && col <= 7
}

func rowColFromSquare(sq int) (int, int) {
	return sq / 8, sq % 8
}

func squareFromRowCol(row int, col int) int {
	return row*8 + col
}

func bitScanForward(n uint64) int {
	if n == 0 {
		return -1
	}
	var index int
	for n&1 == 0 {
		n >>= 1
		index++
	}
	return index
}

func bitboardToArray(bb uint64) [8][8]int {
	var arr [8][8]int
	for i := range 8 {
		for j := range 8 {
			if bitboardCheckOne(bb, i, j) {
				arr[i][j] = 1
			}
		}
	}
	return arr
}
