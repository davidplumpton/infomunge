package evaluator

import (
	"go/parser"
	"reflect"
	"strings"
	"testing"

	"infomunge/pkg/values"
)

func TestInputAbsenceFlowsThroughComposedExpressionsToDefault(t *testing.T) {
	payload := Object{"name": -635}
	values.MarkInputValue(payload)
	lambdaBody, err := parser.ParseExpr("x + 1")
	if err != nil {
		t.Fatalf("parse lambda body: %v", err)
	}
	context := Context{
		"payload": payload,
		"plusOne": &Lambda{
			Params:  []ParamDef{{Name: "x"}},
			Body:    "x + 1",
			BodyAST: lambdaBody,
			Env:     Context{},
		},
	}

	tests := []struct {
		name string
		expr string
		want Value
	}{
		{name: "direct selector", expr: `__default(payload["label"], 5)`, want: 5},
		{name: "nested selector", expr: `__default(payload["user"]["name"], "missing")`, want: "missing"},
		{name: "unary expression", expr: `__default(-(payload["label"]), 6)`, want: 6},
		{name: "binary expression", expr: `__default(payload["label"] + 1, 7)`, want: 7},
		{name: "builtin argument", expr: `__default(abs(payload["label"]), 8)`, want: 8},
		{name: "user function argument", expr: `__default(plusOne(payload["label"]), 9)`, want: 9},
		{name: "equality handles materialized null", expr: `payload["label"] == nil`, want: true},
		{name: "inequality handles materialized null", expr: `payload["label"] != nil`, want: false},
		{
			name: "callback local default",
			expr: `__map([]interface{}{1, 2}, __lambda("x", __default(x + payload["label"], 10)))`,
			want: Array{10, 10},
		},
		{
			name: "direct selector materializes in array",
			expr: `[]interface{}{payload["label"]}`,
			want: Array{nil},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Evaluate(test.expr, context, nil, 0, test.expr)
			if err != nil {
				t.Fatalf("Evaluate(%q) error: %v", test.expr, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Evaluate(%q) = %#v, want %#v", test.expr, got, test.want)
			}
		})
	}
}

func TestInputAbsenceDoesNotHideUnrelatedOrUnconsumedFailures(t *testing.T) {
	payload := Object{"name": -635}
	values.MarkInputValue(payload)
	context := Context{"payload": payload}

	tests := []struct {
		name    string
		expr    string
		wantErr string
	}{
		{name: "explicit null arithmetic", expr: `__default(nil + 1, 5)`, wantErr: "cannot add <nil> and int"},
		{name: "division by zero", expr: `__default(payload["label"] + (1 / 0), 5)`, wantErr: "division by zero"},
		{name: "unhandled absent arithmetic", expr: `payload["label"] + 1`, wantErr: "cannot add <nil> and int"},
		{
			name:    "collection callback without local default",
			expr:    `__map([]interface{}{1}, __lambda("x", x + payload["label"]))`,
			wantErr: "cannot add int and <nil>",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Evaluate(test.expr, context, nil, 0, test.expr)
			if err == nil {
				t.Fatalf("Evaluate(%q) succeeded, want error containing %q", test.expr, test.wantErr)
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("Evaluate(%q) error = %q, want %q", test.expr, err, test.wantErr)
			}
		})
	}
}

func TestScriptObjectMissingFieldRemainsOrdinaryNull(t *testing.T) {
	payload := Object{"name": -635}
	expr := `__default(payload["label"] + 1, 5)`

	_, err := Evaluate(expr, Context{"payload": payload}, nil, 0, expr)
	if err == nil {
		t.Fatal("Evaluate succeeded for unmarked object, want null arithmetic error")
	}
	if !strings.Contains(err.Error(), "cannot add <nil> and int") {
		t.Fatalf("Evaluate error = %q, want null arithmetic context", err)
	}
}
