package determinism

import (
	"errors"
	"testing"
)

func TestEqualErrorsIgnoresPointerAddresses(t *testing.T) {
	first := errors.New(":0:0: unsupported operation: <nil> % &{[{x Unknown <nil> false}] 0x2b0fee700f80 map[]}")
	second := errors.New(":0:0: unsupported operation: <nil> % &{[{x Unknown <nil> false}] 0x2b0fee742200 map[]}")

	if !EqualErrors(first, second) {
		t.Fatal("expected errors with only pointer-address differences to compare equal")
	}
}

func TestEqualErrorsDetectsMeaningfulDifferences(t *testing.T) {
	first := errors.New("unsupported operation: 1 + true")
	second := errors.New("unsupported operation: 1 - true")

	if EqualErrors(first, second) {
		t.Fatal("expected meaningfully different errors to not compare equal")
	}
}

func TestEqualErrorsHandlesNil(t *testing.T) {
	var err error
	if !EqualErrors(err, err) {
		t.Fatal("expected nil errors to compare equal")
	}
	if EqualErrors(errors.New("x"), nil) {
		t.Fatal("expected non-nil and nil errors to differ")
	}
}
