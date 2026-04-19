package test

import (
	"infomunge/internal/evaluator"
	"testing"
)

func evaluate(expr string) (evaluator.Value, error) {
	return evaluateWithContext(expr, make(evaluator.Context))
}

func evaluateWithContext(expr string, context evaluator.Context) (evaluator.Value, error) {
	ctx := &evaluator.ErrorContext{
		ExprStr:    expr,
		Mapping:    make([]int, len(expr)),
		BodyOffset: 0,
		FullRaw:    expr,
	}
	// Fill mapping with sequential indices
	for i := range ctx.Mapping {
		ctx.Mapping[i] = i
	}
	return evaluator.EvaluateWithContext(expr, context, ctx)
}

// TestTypeMatchers tests type assertion matchers
func TestTypeMatchers(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		context evaluator.Context
		wantErr bool
	}{
		{
			name:    "beArray matches array",
			expr:    `must(arr, beArray())`,
			context: evaluator.Context{"arr": evaluator.Array{1.0, 2.0, 3.0}},
		},
		{
			name:    "beArray fails for non-array",
			expr:    `must("not an array", beArray())`,
			wantErr: true,
		},
		{
			name:    "beObject matches object",
			expr:    `must(obj, beObject())`,
			context: evaluator.Context{"obj": evaluator.Context{"x": 1.0}},
		},
		{
			name:    "beObject fails for non-object",
			expr:    `must(arr, beObject())`,
			context: evaluator.Context{"arr": evaluator.Array{1.0, 2.0}},
			wantErr: true,
		},
		{
			name: "beString matches string",
			expr: `must("hello", beString())`,
		},
		{
			name:    "beString fails for non-string",
			expr:    `must(42, beString())`,
			wantErr: true,
		},
		{
			name: "beNumber matches number",
			expr: `must(42, beNumber())`,
		},
		{
			name:    "beNumber fails for non-number",
			expr:    `must("42", beNumber())`,
			wantErr: true,
		},
		{
			name: "beBoolean matches boolean",
			expr: `must(true, beBoolean())`,
		},
		{
			name:    "beBoolean fails for non-boolean",
			expr:    `must(1, beBoolean())`,
			wantErr: true,
		},
		{
			name: "beNull matches null",
			expr: `must(nil, beNull())`,
		},
		{
			name:    "beNull fails for non-null",
			expr:    `must(42, beNull())`,
			wantErr: true,
		},
		{
			name: "beEmpty matches empty string",
			expr: `must("", beEmpty())`,
		},
		{
			name:    "beEmpty matches empty array",
			expr:    `must(emptyArr, beEmpty())`,
			context: evaluator.Context{"emptyArr": evaluator.Array{}},
		},
		{
			name:    "beEmpty fails for non-empty",
			expr:    `must("hello", beEmpty())`,
			wantErr: true,
		},
		{
			name: "beBlank matches blank string",
			expr: `must("   ", beBlank())`,
		},
		{
			name:    "beBlank fails for non-blank",
			expr:    `must("hello", beBlank())`,
			wantErr: true,
		},
		{
			name: "notBeNull matches non-null",
			expr: `must(42, notBeNull())`,
		},
		{
			name:    "notBeNull fails for null",
			expr:    `must(nil, notBeNull())`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluateWithContext(tt.expr, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Errorf("evaluateWithContext() expected result, got nil")
			}
		})
	}
}

// TestComparisonMatchers tests comparison assertion matchers
func TestComparisonMatchers(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		context evaluator.Context
		wantErr bool
	}{
		{
			name: "equalTo matches equal values",
			expr: `must(42, equalTo(42))`,
		},
		{
			name:    "equalTo fails for unequal values",
			expr:    `must(42, equalTo(43))`,
			wantErr: true,
		},
		{
			name: "beGreaterThan for greater value",
			expr: `must(50, beGreaterThan(42))`,
		},
		{
			name:    "beGreaterThan fails for equal value",
			expr:    `must(42, beGreaterThan(42))`,
			wantErr: true,
		},
		{
			name: "beGreaterThan inclusive for equal value",
			expr: `must(42, beGreaterThan(42, true))`,
		},
		{
			name: "beLowerThan for lower value",
			expr: `must(42, beLowerThan(50))`,
		},
		{
			name:    "beLowerThan fails for equal value",
			expr:    `must(50, beLowerThan(50))`,
			wantErr: true,
		},
		{
			name: "beLowerThan inclusive for equal value",
			expr: `must(50, beLowerThan(50, true))`,
		},
		{
			name:    "beOneOf with matching value",
			expr:    `must(v, beOneOf(opts))`,
			context: evaluator.Context{"v": 2.0, "opts": evaluator.Array{1.0, 2.0, 3.0}},
		},
		{
			name:    "beOneOf with non-matching value",
			expr:    `must(v, beOneOf(opts))`,
			context: evaluator.Context{"v": 5.0, "opts": evaluator.Array{1.0, 2.0, 3.0}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		ctx := tt.context
		if ctx == nil {
			ctx = make(evaluator.Context)
		}
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluateWithContext(tt.expr, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Errorf("evaluateWithContext() expected result, got nil")
			}
		})
	}
}

// TestStringMatchers tests string-specific assertion matchers
func TestStringMatchers(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{
			name: "containStr matches substring",
			expr: `must("hello world", containStr("world"))`,
		},
		{
			name:    "containStr fails for missing substring",
			expr:    `must("hello world", containStr("foo"))`,
			wantErr: true,
		},
		{
			name: "startWith matches prefix",
			expr: `must("hello world", startWith("hello"))`,
		},
		{
			name:    "startWith fails for missing prefix",
			expr:    `must("hello world", startWith("world"))`,
			wantErr: true,
		},
		{
			name: "endWith matches suffix",
			expr: `must("hello world", endWith("world"))`,
		},
		{
			name:    "endWith fails for missing suffix",
			expr:    `must("hello world", endWith("hello"))`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluate(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Errorf("evaluate() expected result, got nil")
			}
		})
	}
}

// TestCollectionMatchers tests collection-specific assertion matchers
func TestCollectionMatchers(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		context evaluator.Context
		wantErr bool
	}{
		{
			name:    "containVal matches array element",
			expr:    `must(arr, containVal(2))`,
			context: evaluator.Context{"arr": evaluator.Array{1.0, 2.0, 3.0}},
		},
		{
			name:    "containVal fails for missing element",
			expr:    `must(arr, containVal(5))`,
			context: evaluator.Context{"arr": evaluator.Array{1.0, 2.0, 3.0}},
			wantErr: true,
		},
		{
			name:    "haveSize matches array length",
			expr:    `must(arr, haveSize(3))`,
			context: evaluator.Context{"arr": evaluator.Array{1.0, 2.0, 3.0}},
		},
		{
			name:    "haveSize fails for wrong length",
			expr:    `must(arr, haveSize(4))`,
			context: evaluator.Context{"arr": evaluator.Array{1.0, 2.0, 3.0}},
			wantErr: true,
		},
		{
			name: "haveSize matches string length",
			expr: `must("hello", haveSize(5))`,
		},
		{
			name:    "haveKey matches object key",
			expr:    `must(obj, haveKey("x"))`,
			context: evaluator.Context{"obj": evaluator.Context{"x": 1.0, "y": 2.0}},
		},
		{
			name:    "haveKey fails for missing key",
			expr:    `must(obj, haveKey("z"))`,
			context: evaluator.Context{"obj": evaluator.Context{"x": 1.0, "y": 2.0}},
			wantErr: true,
		},
		{
			name:    "haveValue matches object value",
			expr:    `must(obj, haveValue(1))`,
			context: evaluator.Context{"obj": evaluator.Context{"x": 1.0, "y": 2.0}},
		},
		{
			name:    "haveValue fails for missing value",
			expr:    `must(obj, haveValue(3))`,
			context: evaluator.Context{"obj": evaluator.Context{"x": 1.0, "y": 2.0}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluateWithContext(tt.expr, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Errorf("evaluateWithContext() expected result, got nil")
			}
		})
	}
}

// TestMatcherCombinators tests matcher combining functions
func TestMatcherCombinators(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		context evaluator.Context
		wantErr bool
	}{
		{
			name: "notBe negates matcher",
			expr: `must(5, notBe(equalTo(10)))`,
		},
		{
			name:    "notBe fails when matcher succeeds",
			expr:    `must(5, notBe(equalTo(5)))`,
			wantErr: true,
		},
		{
			name:    "eachItem validates all array items",
			expr:    `must(arr, eachItem(beNumber()))`,
			context: evaluator.Context{"arr": evaluator.Array{1.0, 2.0, 3.0}},
		},
		{
			name:    "eachItem fails when one item doesn't match",
			expr:    `must(arr, eachItem(beNumber()))`,
			context: evaluator.Context{"arr": evaluator.Array{1.0, "two", 3.0}},
			wantErr: true,
		},
		{
			name:    "haveItem finds matching item",
			expr:    `must(arr, haveItem(equalTo(2)))`,
			context: evaluator.Context{"arr": evaluator.Array{1.0, 2.0, 3.0}},
		},
		{
			name:    "haveItem fails when no item matches",
			expr:    `must(arr, haveItem(equalTo(5)))`,
			context: evaluator.Context{"arr": evaluator.Array{1.0, 2.0, 3.0}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.context
			if ctx == nil {
				ctx = make(evaluator.Context)
			}
			result, err := evaluateWithContext(tt.expr, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Errorf("evaluateWithContext() expected result, got nil")
			}
		})
	}
}
