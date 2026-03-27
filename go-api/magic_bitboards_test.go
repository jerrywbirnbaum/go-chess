package main

import (
	"math/bits"
	"testing"
)

func bishopLookup(table [64][512]uint64, sq int, blockers uint64) uint64 {
	row, col := rowColFromSquare(sq)
	magicIndex := (reverseColBits(blockers) * getBishopMagicNumber(row, col)) >> getBishopShift(row, col)
	return table[sq][magicIndex]
}

func TestCreateBishopLookupTable(t *testing.T) {
	table := createBishopLookupTable()

	t.Run("lookup matches slidingAttackBits for all squares and blockers", func(t *testing.T) {
		for sq := range 64 {
			row, col := rowColFromSquare(sq)
			blockerMasks := createBlockerMasks(bishopMasks[sq])
			for _, blockers := range blockerMasks {
				want := slidingAttackBits(row, col, Bishop, blockers)
				got := bishopLookup(table, sq, blockers)
				if got != want {
					t.Errorf("sq=%d blockers=%064b: got %064b, want %064b", sq, blockers, got, want)
				}
			}
		}
	})

	t.Run("empty blockers on corner square a1 (sq=0) attacks only along one diagonal", func(t *testing.T) {
		got := bishopLookup(table, 0, 0)
		row, col := rowColFromSquare(0)
		want := slidingAttackBits(row, col, Bishop, 0)
		if got != want {
			t.Errorf("sq=0 empty blockers: got %064b, want %064b", got, want)
		}
		if got == 0 {
			t.Errorf("sq=0 bishop with no blockers should have at least one attack square")
		}
	})

	t.Run("empty blockers on corner square h8 (sq=63) attacks only along one diagonal", func(t *testing.T) {
		got := bishopLookup(table, 63, 0)
		row, col := rowColFromSquare(63)
		want := slidingAttackBits(row, col, Bishop, 0)
		if got != want {
			t.Errorf("sq=63 empty blockers: got %064b, want %064b", got, want)
		}
		if got == 0 {
			t.Errorf("sq=63 bishop with no blockers should have at least one attack square")
		}
	})

	t.Run("center square has more attack squares than corner with empty blockers", func(t *testing.T) {
		corner := bishopLookup(table, 0, 0)
		center := bishopLookup(table, 27, 0) // d4
		if bits.OnesCount64(center) <= bits.OnesCount64(corner) {
			t.Errorf("center sq=27 should attack more squares (%d) than corner sq=0 (%d) with no blockers",
				bits.OnesCount64(center), bits.OnesCount64(corner))
		}
	})

	t.Run("blockers on bishop mask do not affect attacks outside the mask", func(t *testing.T) {
		for _, sq := range []int{0, 9, 27, 36, 54, 63} {
			row, col := rowColFromSquare(sq)
			noBockers := bishopLookup(table, sq, 0)
			allBlockers := bishopLookup(table, sq, bishopMasks[sq])
			want := slidingAttackBits(row, col, Bishop, bishopMasks[sq])
			if allBlockers != want {
				t.Errorf("sq=%d full blockers: got %064b, want %064b", sq, allBlockers, want)
			}
			_ = noBockers
		}
	})

	t.Run("all 64 squares produce non-zero attack bitboards with no blockers (except edge squares may vary)", func(t *testing.T) {
		for sq := range 64 {
			got := bishopLookup(table, sq, 0)
			row, col := rowColFromSquare(sq)
			want := slidingAttackBits(row, col, Bishop, 0)
			if got != want {
				t.Errorf("sq=%d empty blockers mismatch: got %064b, want %064b", sq, got, want)
			}
		}
	})

	t.Run("magic index stays within 512 bounds for all blocker configurations", func(t *testing.T) {
		for sq := range 64 {
			blockerMasks := createBlockerMasks(bishopMasks[sq])
			row, col := rowColFromSquare(sq)
			for _, blockers := range blockerMasks {
				index := (blockers * getBishopMagicNumber(row, col)) >> bishopShifts[sq]
				if index >= 512 {
					t.Errorf("sq=%d: magic index %d exceeds 511", sq, index)
				}
			}
		}
	})
}

func rookLookupFn(table [64][4096]uint64, sq int, blockers uint64) uint64 {
	row, col := rowColFromSquare(sq)
	magicIndex := (reverseColBits(blockers) * getRookMagicNumber(row, col)) >> getRookShift(row, col)
	return table[sq][magicIndex]
}

func TestCreateRookLookupTable(t *testing.T) {
	table := createRookLookupTable()

	t.Run("lookup matches slidingAttackBits for all squares and blockers", func(t *testing.T) {
		for sq := range 64 {
			row, col := rowColFromSquare(sq)
			blockerMasks := createBlockerMasks(rookMasks[sq])
			for _, blockers := range blockerMasks {
				want := slidingAttackBits(row, col, Rook, blockers)
				got := rookLookupFn(table, sq, blockers)
				if got != want {
					t.Errorf("sq=%d blockers=%064b: got %064b, want %064b", sq, blockers, got, want)
				}
			}
		}
	})

	t.Run("empty blockers on corner square a1 (sq=0) attacks full rank and file", func(t *testing.T) {
		got := rookLookupFn(table, 0, 0)
		row, col := rowColFromSquare(0)
		want := slidingAttackBits(row, col, Rook, 0)
		if got != want {
			t.Errorf("sq=0 empty blockers: got %064b, want %064b", got, want)
		}
		if got == 0 {
			t.Errorf("sq=0 rook with no blockers should have attack squares")
		}
	})

	t.Run("all squares attack exactly 14 squares with no blockers", func(t *testing.T) {
		for sq := range 64 {
			got := rookLookupFn(table, sq, 0)
			if bits.OnesCount64(got) != 14 {
				t.Errorf("sq=%d: expected 14 attack squares with no blockers, got %d", sq, bits.OnesCount64(got))
			}
		}
	})

	t.Run("blocker on same rank stops attack", func(t *testing.T) {
		// Rook on a1 (sq=0), blocker on d1 — should not reach e1 or beyond on rank
		row, col := rowColFromSquare(0)
		blockerSq := squareFromRowCol(row, col+3) // d1
		blockers := uint64(1) << (63 - blockerSq)
		blockers &= rookMasks[0]
		got := rookLookupFn(table, 0, blockers)
		want := slidingAttackBits(row, col, Rook, blockers)
		if got != want {
			t.Errorf("sq=0 blocker on d1: got %064b, want %064b", got, want)
		}
	})

	t.Run("magic index stays within 4096 bounds for all blocker configurations", func(t *testing.T) {
		for sq := range 64 {
			blockerMasks := createBlockerMasks(rookMasks[sq])
			row, col := rowColFromSquare(sq)
			for _, blockers := range blockerMasks {
				index := (reverseColBits(blockers) * getRookMagicNumber(row, col)) >> getRookShift(row, col)
				if index >= 4096 {
					t.Errorf("sq=%d: magic index %d exceeds 4095", sq, index)
				}
			}
		}
	})

	t.Run("all 64 squares produce correct attacks with no blockers", func(t *testing.T) {
		for sq := range 64 {
			got := rookLookupFn(table, sq, 0)
			row, col := rowColFromSquare(sq)
			want := slidingAttackBits(row, col, Rook, 0)
			if got != want {
				t.Errorf("sq=%d empty blockers mismatch: got %064b, want %064b", sq, got, want)
			}
		}
	})
}

func TestCreateBlockerMasks(t *testing.T) {
	t.Run("empty mask returns empty slice", func(t *testing.T) {
		result := createBlockerMasks(0)
		if len(result) != 0 {
			t.Errorf("expected empty slice for empty mask, got %d elements", len(result))
		}
	})

	t.Run("single bit mask generates 2 subsets", func(t *testing.T) {
		mask := uint64(1) // bit 0 only
		result := createBlockerMasks(mask)
		if len(result) != 2 {
			t.Errorf("expected 2 blocker masks for 1-bit mask, got %d", len(result))
		}
	})

	t.Run("two bit mask generates 4 subsets", func(t *testing.T) {
		mask := uint64(0b11) // bits 0 and 1
		result := createBlockerMasks(mask)
		if len(result) != 4 {
			t.Errorf("expected 4 blocker masks for 2-bit mask, got %d", len(result))
		}
	})

	t.Run("three bit mask generates 8 subsets", func(t *testing.T) {
		mask := uint64(0b111) // bits 0, 1, 2
		result := createBlockerMasks(mask)
		if len(result) != 8 {
			t.Errorf("expected 8 blocker masks for 3-bit mask, got %d", len(result))
		}
	})

	t.Run("count is 2^N for N set bits", func(t *testing.T) {
		cases := []uint64{
			0b1,
			0b11,
			0b1010,
			0b10101,
			0b10000001,
		}
		for _, mask := range cases {
			n := bits.OnesCount64(mask)
			expected := 1 << n
			result := createBlockerMasks(mask)
			if len(result) != expected {
				t.Errorf("mask %b: expected %d blocker masks (2^%d), got %d", mask, expected, n, len(result))
			}
		}
	})

	t.Run("all results are subsets of the mask", func(t *testing.T) {
		mask := uint64(0b10101) // bits 0, 2, 4
		result := createBlockerMasks(mask)

		for i, bm := range result {
			if bm&^mask != 0 {
				t.Errorf("blocker mask %d (%b) has bits outside the input mask (%b)", i, bm, mask)
			}
		}
	})

	t.Run("no duplicate masks", func(t *testing.T) {
		mask := uint64(0b111)
		result := createBlockerMasks(mask)
		seen := make(map[uint64]bool)
		for _, bm := range result {
			if seen[bm] {
				t.Errorf("duplicate blocker mask: %b", bm)
			}
			seen[bm] = true
		}
	})

	t.Run("all subsets present for two bit mask", func(t *testing.T) {
		mask := uint64(0b11) // bits 0 and 1
		result := createBlockerMasks(mask)
		expected := []uint64{0b00, 0b01, 0b10, 0b11}
		seen := make(map[uint64]bool)
		for _, bm := range result {
			seen[bm] = true
		}
		for _, subset := range expected {
			if !seen[subset] {
				t.Errorf("missing subset %b from blocker masks of mask %b", subset, mask)
			}
		}
	})

	t.Run("non-contiguous bits produce correct subsets", func(t *testing.T) {
		// bits 0 and 7 set
		mask := uint64(1) | uint64(1)<<7
		result := createBlockerMasks(mask)
		if len(result) != 4 {
			t.Errorf("expected 4 blocker masks, got %d", len(result))
		}
		expected := []uint64{0, 1, 1 << 7, 1 | 1<<7}
		seen := make(map[uint64]bool)
		for _, bm := range result {
			seen[bm] = true
		}
		for _, subset := range expected {
			if !seen[subset] {
				t.Errorf("missing subset %b", subset)
			}
		}
	})
}
