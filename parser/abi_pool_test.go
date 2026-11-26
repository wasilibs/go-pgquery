//go:build !tinygo && !pgquery_cgo

package parser

import (
	"errors"
	"testing"
)

func TestShouldRecycleABI(t *testing.T) {
	if !shouldRecycleABI(nil) {
		t.Fatal("nil error should recycle")
	}
	if !shouldRecycleABI(&Error{Message: "syntax error"}) {
		t.Fatal("SQL *Error should recycle")
	}
	if shouldRecycleABI(errors.New("wasm boom")) {
		t.Fatal("non-SQL error must not recycle")
	}
}

func drainABIPool() {
	wasmInitOnce.Do(initWASM)
	for abiPoolLen() > 0 {
		getABI().discard()
	}
}

func TestDiscardDoesNotPool(t *testing.T) {
	drainABIPool()

	a := newABI()
	before := abiPoolLen()
	a.discard()
	if abiPoolLen() != before {
		t.Fatalf("discard must not change pool length; before=%d after=%d", before, abiPoolLen())
	}

	a2 := newABI()
	a2.Close()
	if abiPoolLen() != before+1 {
		t.Fatalf("Close should pool; want %d got %d", before+1, abiPoolLen())
	}
	getABI().discard()
}

func TestRunWithABIDiscardsOnNonSQLError(t *testing.T) {
	drainABIPool()

	err := runWithABI(func(*abi) error {
		return errors.New("simulated wasm failure")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if abiPoolLen() != 0 {
		t.Fatalf("non-SQL error must not pool ABI; pool len=%d", abiPoolLen())
	}
}

func TestRunWithABIRecyclesOnNil(t *testing.T) {
	drainABIPool()

	err := runWithABI(func(*abi) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if abiPoolLen() != 1 {
		t.Fatalf("success should pool ABI; pool len=%d", abiPoolLen())
	}
	getABI().discard()
}

func TestSQLErrorStillRecyclesABI(t *testing.T) {
	drainABIPool()

	_, err := ParseToProtobuf("NOT VALID SQL !!!")
	if err == nil {
		t.Fatal("expected SQL parse error")
	}
	var pe *Error
	if !errors.As(err, &pe) {
		t.Fatalf("want *Error, got %T %v", err, err)
	}
	if abiPoolLen() != 1 {
		t.Fatalf("SQL error should recycle ABI; pool len=%d", abiPoolLen())
	}

	_, err = ParseToProtobuf("SELECT 1")
	if err != nil {
		t.Fatalf("recycled ABI should still parse: %v", err)
	}
}
