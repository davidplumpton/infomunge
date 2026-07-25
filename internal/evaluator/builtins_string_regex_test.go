package evaluator

import (
	"go/ast"
	"testing"
)

func TestBuiltinMatchRequiresWholeInputAndIncludesFullMatch(t *testing.T) {
	tests := []struct {
		name     string
		args     []Value
		expected Array
	}{
		{
			name:     "partial match",
			args:     []Value{"foo123bar", `([a-z]+)([0-9]+)`},
			expected: Array{},
		},
		{
			name:     "full match with captures",
			args:     []Value{"foo123", `([a-z]+)([0-9]+)`},
			expected: Array{"foo123", "foo", "123"},
		},
		{
			name:     "anchored suffix is still partial",
			args:     []Value{"ba", `a$`},
			expected: Array{},
		},
		{
			name:     "full match without captures",
			args:     []Value{"hello", `[a-z]+`},
			expected: Array{"hello"},
		},
		{
			name:     "explicit flags are preserved",
			args:     []Value{"HELLO", `hello`, "i"},
			expected: Array{"HELLO"},
		},
		{
			name:     "full alternative can follow prefix alternative",
			args:     []Value{"abc", `a|abc`},
			expected: Array{"abc"},
		},
	}

	call := &ast.CallExpr{Fun: &ast.Ident{Name: "match"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := callBuiltinMatch(tt.args, call)
			if err != nil {
				t.Fatalf("callBuiltinMatch() error = %v", err)
			}
			if !isEqual(got, tt.expected) {
				t.Fatalf("callBuiltinMatch() = %#v, want %#v", got, tt.expected)
			}
		})
	}
}

func TestBuiltinMatchesRequiresWholeInput(t *testing.T) {
	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name:     "substring",
			args:     []Value{"hello world", "l"},
			expected: false,
		},
		{
			name:     "anchored suffix",
			args:     []Value{"ba", `a$`},
			expected: false,
		},
		{
			name:     "whole input",
			args:     []Value{"hello", `[a-z]+`},
			expected: true,
		},
		{
			name:     "explicit flags",
			args:     []Value{"HELLO", "hello", "i"},
			expected: true,
		},
		{
			name:     "full alternative after prefix alternative",
			args:     []Value{"abc", `a|abc`},
			expected: true,
		},
	}

	call := &ast.CallExpr{Fun: &ast.Ident{Name: "matches"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := callBuiltinMatches(tt.args, call)
			if err != nil {
				t.Fatalf("callBuiltinMatches() error = %v", err)
			}
			if got != tt.expected {
				t.Fatalf("callBuiltinMatches() = %#v, want %t", got, tt.expected)
			}
		})
	}
}

func TestBuiltinScanAlwaysIncludesFullMatch(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pattern  string
		expected Array
	}{
		{
			name:    "zero capture groups",
			text:    "abc123def",
			pattern: `[0-9]+`,
			expected: Array{
				Array{"123"},
			},
		},
		{
			name:    "multiple capture groups",
			text:    "foo1 bar2",
			pattern: `([a-z]+)([0-9])`,
			expected: Array{
				Array{"foo1", "foo", "1"},
				Array{"bar2", "bar", "2"},
			},
		},
		{
			name:     "no matches",
			text:     "hello",
			pattern:  `[0-9]+`,
			expected: Array{},
		},
	}

	call := &ast.CallExpr{Fun: &ast.Ident{Name: "scan"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := callBuiltinScan([]Value{tt.text, tt.pattern}, call)
			if err != nil {
				t.Fatalf("callBuiltinScan() error = %v", err)
			}
			if !isEqual(got, tt.expected) {
				t.Fatalf("callBuiltinScan() = %#v, want %#v", got, tt.expected)
			}
		})
	}
}
