package evaluator

import (
	"math/big"
	"reflect"
	"strconv"
)

// CoerceEquals implements the ~= coercion equality operator.
func CoerceEquals(left, right Value) bool {
	return newEqualityComparator(true).equal(left, right)
}

// numericEquals compares values recursively, treating exactly equivalent int
// and float64 values as equal at every nesting level.
func numericEquals(left, right Value) bool {
	return newEqualityComparator(false).equal(left, right)
}

type equalityVisit struct {
	kind        uint8
	left        uintptr
	right       uintptr
	leftLength  int
	rightLength int
}

type equalityComparator struct {
	coerce bool
	seen   map[equalityVisit]struct{}
}

func newEqualityComparator(coerce bool) *equalityComparator {
	return &equalityComparator{
		coerce: coerce,
		seen:   make(map[equalityVisit]struct{}),
	}
}

func (c *equalityComparator) equal(left, right Value) bool {
	if leftObject, ok := left.(Object); ok {
		rightObject, ok := right.(Object)
		if !ok {
			return false
		}
		return c.equalObjects(leftObject, rightObject)
	}
	if _, ok := right.(Object); ok {
		return false
	}

	if leftSequence, ok := equalitySequence(left); ok {
		rightSequence, ok := equalitySequence(right)
		if !ok {
			return false
		}
		return c.equalSequences(leftSequence, rightSequence)
	}
	if _, ok := equalitySequence(right); ok {
		return false
	}

	if c.coerce {
		if leftNumber, leftOK := coercionNumericRat(left); leftOK {
			if rightNumber, rightOK := coercionNumericRat(right); rightOK {
				return leftNumber.Cmp(rightNumber) == 0
			}
		}
		if leftBool, ok := left.(bool); ok {
			if rightString, ok := right.(string); ok {
				return rightString == strconv.FormatBool(leftBool)
			}
		}
		if leftString, ok := left.(string); ok {
			if rightBool, ok := right.(bool); ok {
				return leftString == strconv.FormatBool(rightBool)
			}
		}
	} else if leftNumber, leftOK := exactNumericRat(left); leftOK {
		if rightNumber, rightOK := exactNumericRat(right); rightOK {
			return leftNumber.Cmp(rightNumber) == 0
		}
		return false
	}

	if left == nil || right == nil {
		return left == nil && right == nil
	}

	leftType := reflect.TypeOf(left)
	if leftType != reflect.TypeOf(right) || !leftType.Comparable() {
		return false
	}
	return left == right
}

func equalitySequence(value Value) ([]Value, bool) {
	switch sequence := value.(type) {
	case Array:
		return []Value(sequence), true
	case XMLMultiValue:
		return []Value(sequence), true
	default:
		return nil, false
	}
}

func (c *equalityComparator) equalSequences(left, right []Value) bool {
	if len(left) != len(right) {
		return false
	}
	visit := equalityVisit{
		kind:        1,
		left:        uintptr(reflect.ValueOf(left).UnsafePointer()),
		right:       uintptr(reflect.ValueOf(right).UnsafePointer()),
		leftLength:  len(left),
		rightLength: len(right),
	}
	if c.alreadySeen(visit) {
		return true
	}
	for index := range left {
		if !c.equal(left[index], right[index]) {
			return false
		}
	}
	return true
}

func (c *equalityComparator) equalObjects(left, right Object) bool {
	if len(left) != len(right) {
		return false
	}
	visit := equalityVisit{
		kind:        2,
		left:        uintptr(reflect.ValueOf(left).UnsafePointer()),
		right:       uintptr(reflect.ValueOf(right).UnsafePointer()),
		leftLength:  len(left),
		rightLength: len(right),
	}
	if c.alreadySeen(visit) {
		return true
	}
	for key, leftValue := range left {
		rightValue, ok := right[key]
		if !ok || !c.equal(leftValue, rightValue) {
			return false
		}
	}
	return true
}

func (c *equalityComparator) alreadySeen(visit equalityVisit) bool {
	if _, ok := c.seen[visit]; ok {
		return true
	}
	c.seen[visit] = struct{}{}
	return false
}

func coercionNumericRat(value Value) (*big.Rat, bool) {
	if number, ok := exactNumericRat(value); ok {
		return number, true
	}

	text, ok := value.(string)
	if !ok || text == "" {
		return nil, false
	}
	if integer, ok := new(big.Int).SetString(text, 10); ok {
		return new(big.Rat).SetInt(integer), true
	}
	number, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil, false
	}
	return exactNumericRat(number)
}
