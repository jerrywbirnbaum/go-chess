package main

import (
	"math"
)

func init() {
	for sq := range 64 {
		r, c := sq/8, sq%8
		knightAttacks[sq] = leaperAttackBits(r, c, knightOffsets[:])
		kingAttacks[sq] = leaperAttackBits(r, c, kingOffsets[:])
		bishopMasks[sq] = sliderMaskBits(r, c, diagonalDirs[:])
		rookMasks[sq] = sliderMaskBits(r, c, straightDirs[:])
	}
	bishopMagicLookup = createBishopLookupTable()
	rookMagicLookup = createRookLookupTable()
	initTables()
}

func getBishopMagicNumber(row int, col int) uint64 {
	row = 7 - row
	return BishopMagics[row*8+col]
}

func getBishopShift(row int, col int) uint64 {
	row = 7 - row
	return bishopShifts[row*8+col]
}

func getRookMagicNumber(row int, col int) uint64 {
	row = 7 - row
	return rookMagics[row*8+col]
}

func getRookShift(row int, col int) uint64 {
	row = 7 - row
	return rookShifts[row*8+col]
}

// reverseColBits reverses bit order within each byte, converting the engine's
// occupancy encoding ((7-r)*8+(7-c)) to the standard encoding ((7-r)*8+c)
func reverseColBits(x uint64) uint64 {
	x = ((x & 0xF0F0F0F0F0F0F0F0) >> 4) | ((x & 0x0F0F0F0F0F0F0F0F) << 4)
	x = ((x & 0xCCCCCCCCCCCCCCCC) >> 2) | ((x & 0x3333333333333333) << 2)
	x = ((x & 0xAAAAAAAAAAAAAAAA) >> 1) | ((x & 0x5555555555555555) << 1)
	return x
}

func createBlockerMasks(mask uint64) []uint64 {
	var blockerMasks []uint64
	var blockerIdxs []int
	for mask != 0 {
		blockerIdx := bitScanForward(mask)
		mask ^= 1 << blockerIdx
		blockerIdxs = append(blockerIdxs, blockerIdx)
	}
	if len(blockerIdxs) == 0 {
		return blockerMasks
	}
	nums := uint64(math.Pow(2, float64(len(blockerIdxs))))
	for i := range nums {
		blockerMask := uint64(0)
		for blockerIdx := range blockerIdxs {
			binaryShift := nthBinaryValue(i, blockerIdx)
			blockerMask |= binaryShift << uint64(blockerIdxs[blockerIdx])
		}
		blockerMasks = append(blockerMasks, blockerMask)
	}
	return blockerMasks
}

func nthBinaryValue(binary uint64, index int) uint64 {
	return binary >> index & 1
}
func createBishopLookupTable() [64][512]uint64 {
	var bishopLookup [64][512]uint64
	for sq := range 64 {
		bishopBlockers := createBlockerMasks(bishopMasks[sq])
		for index, blockers := range bishopBlockers {
			_ = index

			row, col := rowColFromSquare(sq)
			magicIndex := (reverseColBits(blockers) * getBishopMagicNumber(row, col)) >> getBishopShift(row, col)
			precomputedBitboard := slidingAttackBits(row, col, Bishop, blockers)
			bishopLookup[sq][magicIndex] = precomputedBitboard
		}
	}
	return bishopLookup
}

func createRookLookupTable() [64][4096]uint64 {
	var rookLookup [64][4096]uint64
	for sq := range 64 {
		rookBlockers := createBlockerMasks(rookMasks[sq])
		for _, blockers := range rookBlockers {
			row, col := rowColFromSquare(sq)
			magicIndex := (reverseColBits(blockers) * getRookMagicNumber(row, col)) >> getRookShift(row, col)
			precomputedBitboard := slidingAttackBits(row, col, Rook, blockers)
			rookLookup[sq][magicIndex] = precomputedBitboard
		}
	}
	return rookLookup
}

func slidingAttackBits(row, col int, pt PieceType, bitboard uint64) uint64 {
	dirs := allDirs
	if isRook(pt) {
		dirs = straightDirs
	} else if isBishop(pt) {
		dirs = diagonalDirs
	}
	var bits uint64
	for _, dir := range dirs {
		r, c := row+dir[0], col+dir[1]
		for inBounds(r, c) {
			bits = bitboardAddOne(bits, r, c)
			if bitboardCheckOne(bitboard, r, c) {
				break
			}
			r += dir[0]
			c += dir[1]
		}
	}
	return bits
}
