package exprgen_test

import (
	"encoding/json"
	"strings"
	"testing"

	"infomunge/internal/evaluator"
	"infomunge/internal/preprocessor"
	"infomunge/internal/testing/exprgen"

	"pgregory.net/rapid"
)

// evalWithContext evaluates an infomunge expression with the given context.
func evalWithContext(expr string, ctx map[string]interface{}) (interface{}, error) {
	prepared, mapping, err := preprocessor.PrepareForParsing(expr, preprocessor.Options{})
	if err != nil {
		return nil, err
	}
	return evaluator.Evaluate(prepared, ctx, mapping, 0, expr)
}

// --- SampleContext properties ---

func TestSampleContext_ProducesValidJSON(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tc := exprgen.SampleContext().Draw(t, "ctx")
		if tc.JSON == "" {
			t.Fatal("SampleContext produced empty JSON")
		}
		var parsed interface{}
		if err := json.Unmarshal([]byte(tc.JSON), &parsed); err != nil {
			t.Fatalf("SampleContext JSON is invalid: %v\nJSON: %s", err, tc.JSON)
		}
	})
}

func TestSampleContext_HasPayload(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tc := exprgen.SampleContext().Draw(t, "ctx")
		if _, ok := tc.Value["payload"]; !ok {
			t.Fatal("SampleContext missing 'payload' key")
		}
	})
}

func TestSampleContext_PayloadIsObject(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tc := exprgen.SampleContext().Draw(t, "ctx")
		payload := tc.Value["payload"]
		if _, ok := payload.(map[string]interface{}); !ok {
			t.Fatalf("SampleContext payload is %T, want map[string]interface{}", payload)
		}
	})
}

func TestSampleContext_FieldsMatchPayload(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tc := exprgen.SampleContext().Draw(t, "ctx")
		payload := tc.Value["payload"].(map[string]interface{})
		if len(tc.Fields) != len(payload) {
			t.Fatalf("Fields count %d != payload keys %d", len(tc.Fields), len(payload))
		}
		for _, f := range tc.Fields {
			if _, ok := payload[f]; !ok {
				t.Fatalf("Field %q not found in payload", f)
			}
		}
	})
}

func TestSampleContext_JSONRoundTrips(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tc := exprgen.SampleContext().Draw(t, "ctx")
		// Re-marshal the payload and compare to stored JSON.
		payload := tc.Value["payload"]
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		if string(jsonBytes) != tc.JSON {
			t.Fatalf("JSON mismatch:\n  stored:  %s\n  marshal: %s", tc.JSON, string(jsonBytes))
		}
	})
}

// --- Variety tests ---

func TestSampleContext_ProducesVariety(t *testing.T) {
	sawString := false
	sawNumber := false
	sawBool := false
	sawNull := false
	sawNested := false
	sawArray := false

	for i := 0; i < 200; i++ {
		rapid.Check(t, func(t *rapid.T) {
			tc := exprgen.SampleContext().Draw(t, "ctx")
			payload := tc.Value["payload"].(map[string]interface{})
			for _, v := range payload {
				switch val := v.(type) {
				case string:
					sawString = true
				case int, float64:
					sawNumber = true
				case bool:
					sawBool = true
				case nil:
					sawNull = true
				case map[string]interface{}:
					if len(val) > 0 {
						sawNested = true
					}
				case []interface{}:
					sawArray = true
				}
			}
		})
		if sawString && sawNumber && sawBool && sawNull && sawNested && sawArray {
			break
		}
	}

	if !sawString {
		t.Error("never saw a string field value")
	}
	if !sawNumber {
		t.Error("never saw a number field value")
	}
	if !sawBool {
		t.Error("never saw a boolean field value")
	}
	if !sawNull {
		t.Error("never saw a null field value")
	}
	if !sawNested {
		t.Error("never saw a nested object field value")
	}
	if !sawArray {
		t.Error("never saw an array field value")
	}
}

// --- IdentGen tests ---

func TestIdentGen_ProducesPayload(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tc := exprgen.SampleContext().Draw(t, "ctx")
		ident := exprgen.IdentGen(tc).Draw(t, "ident")
		if !strings.HasPrefix(ident, "payload") {
			t.Fatalf("IdentGen produced %q, want prefix 'payload'", ident)
		}
	})
}

func TestIdentGen_FieldsInScope(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tc := exprgen.SampleContext().Draw(t, "ctx")
		ident := exprgen.IdentGen(tc).Draw(t, "ident")
		if ident == "payload" {
			return // bare payload is always valid
		}
		field := strings.TrimPrefix(ident, "payload.")
		found := false
		for _, f := range tc.Fields {
			if f == field {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("IdentGen produced field %q not in scope %v", field, tc.Fields)
		}
	})
}

func TestIdentGen_EvaluatesWithContext(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		tc := exprgen.SampleContext().Draw(t, "ctx")
		ident := exprgen.IdentGen(tc).Draw(t, "ident")
		// Should evaluate without panic.
		_, err := evalWithContext(ident, tc.Value)
		if err != nil {
			t.Fatalf("IdentGen %q eval failed: %v", ident, err)
		}
	})
}

func TestIdentGen_ProducesVariety(t *testing.T) {
	sawBare := false
	sawField := false

	for i := 0; i < 200; i++ {
		rapid.Check(t, func(t *rapid.T) {
			tc := exprgen.SampleContext().Draw(t, "ctx")
			ident := exprgen.IdentGen(tc).Draw(t, "ident")
			if ident == "payload" {
				sawBare = true
			} else {
				sawField = true
			}
		})
		if sawBare && sawField {
			break
		}
	}

	if !sawBare {
		t.Error("never saw bare 'payload' identifier")
	}
	if !sawField {
		t.Error("never saw 'payload.field' identifier")
	}
}

// --- Known examples ---

func TestSampleContext_KnownEval(t *testing.T) {
	// Construct a known context and verify identifiers evaluate correctly.
	ctx := map[string]interface{}{
		"payload": map[string]interface{}{
			"name": "test",
			"age":  42,
		},
	}

	tests := []struct {
		expr     string
		expected interface{}
	}{
		{"payload.name", "test"},
		{"payload.age", 42},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := evalWithContext(tt.expr, ctx)
			if err != nil {
				t.Fatalf("eval %q: %v", tt.expr, err)
			}
			if result != tt.expected {
				t.Errorf("eval %q = %v (%T), want %v (%T)", tt.expr, result, result, tt.expected, tt.expected)
			}
		})
	}
}
