//go:build !tinygo && !pgquery_cgo

package parser

import (
	"context"
	"testing"
)

func TestTrapDoesNotRecycleABI(t *testing.T) {
	abi := getABI()

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic from wasm trap")
			}
		}()
		defer abi.Close()
		// Hash a range past the end of memory to trigger an out of bounds trap.
		abi.hashXXH364.Call3(context.Background(), 0xFFFFFFF0, 16, 0)
	}()

	abiPoolMu.Lock()
	for e := abiPool.Front(); e != nil; e = e.Next() {
		if e.Value == abi {
			t.Error("ABI was returned to the pool after a trap")
		}
	}
	abiPoolMu.Unlock()

	if _, err := ParseToJSON("SELECT 1"); err != nil {
		t.Fatal(err)
	}
}
