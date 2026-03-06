package main

import "testing"

func TestTranspositionTableInitCreatesMap(t *testing.T) {
	var tt TranspositionTable
	tt.init()

	if tt.table == nil {
		t.Fatal("expected init to allocate table map")
	}
}

func TestTranspositionTablePushAndLookup(t *testing.T) {
	var tt TranspositionTable
	tt.init()

	const key int64 = 42
	tt.push(key, 5, 1, 320)

	depth, flag, evaluation := tt.lookup(key)
	if depth != 5 || flag != 1 || evaluation != 320 {
		t.Fatalf("unexpected lookup values: got (%d, %d, %d), expected (%d, %d, %d)", depth, flag, evaluation, 5, 1, 320)
	}
}

func TestTranspositionTablePushOverwritesExistingKey(t *testing.T) {
	var tt TranspositionTable
	tt.init()

	const key int64 = 99
	tt.push(key, 2, 0, 100)
	tt.push(key, 7, 2, -50)

	depth, flag, evaluation := tt.lookup(key)
	if depth != 7 || flag != 2 || evaluation != -50 {
		t.Fatalf("unexpected lookup after overwrite: got (%d, %d, %d), expected (%d, %d, %d)", depth, flag, evaluation, 7, 2, -50)
	}
}

func TestTranspositionTableLookupMissingKeyReturnsZeroValues(t *testing.T) {
	var tt TranspositionTable
	tt.init()

	depth, flag, evaluation := tt.lookup(123456)
	if depth != 0 || flag != 0 || evaluation != 0 {
		t.Fatalf("unexpected missing-key values: got (%d, %d, %d), expected (0, 0, 0)", depth, flag, evaluation)
	}
}
