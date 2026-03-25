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
