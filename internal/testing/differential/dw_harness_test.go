package differential

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestDWAvailable_DoesNotPanic(t *testing.T) {
	_ = DWAvailable()
}

func TestDWEval_ParsesJSONOutput(t *testing.T) {
	orig := runDWCommand
	t.Cleanup(func() { runDWCommand = orig })

	runDWCommand = func(_ context.Context, script string, inputs map[string]string) (string, string, error) {
		if script == "" {
			t.Fatal("script should not be empty")
		}
		if inputs["payload"] != `{"x":1}` {
			t.Fatalf("unexpected payload input: %#v", inputs)
		}
		return `{"a":1,"b":[true,null]}`, "", nil
	}

	got, err := DWEval("%dw 2.0\noutput application/json\n---\npayload", map[string]string{"payload": `{"x":1}`})
	if err != nil {
		t.Fatalf("DWEval failed: %v", err)
	}

	if err := StructuralCompare(got, map[string]interface{}{
		"a": float64(1),
		"b": []interface{}{true, nil},
	}); err != nil {
		t.Fatalf("unexpected parsed result: %v", err)
	}
}

func TestDWEval_Timeout(t *testing.T) {
	orig := runDWCommand
	t.Cleanup(func() { runDWCommand = orig })

	runDWCommand = func(ctx context.Context, _ string, _ map[string]string) (string, string, error) {
		<-ctx.Done()
		return "", "", ctx.Err()
	}

	start := time.Now()
	_, err := DWEval("%dw 2.0\noutput application/json\n---\n1", nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout message, got: %v", err)
	}
	// Guard against accidental no-timeout behavior.
	if time.Since(start) > 11*time.Second {
		t.Fatalf("evaluation exceeded expected timeout window: %v", time.Since(start))
	}
}

func TestDWEval_CommandErrorIncludesStderr(t *testing.T) {
	orig := runDWCommand
	t.Cleanup(func() { runDWCommand = orig })

	runDWCommand = func(_ context.Context, _ string, _ map[string]string) (string, string, error) {
		return "", "syntax error near line 1", errors.New("exit status 1")
	}

	_, err := DWEval("%dw 2.0\noutput application/json\n---\n(", nil)
	if err == nil {
		t.Fatal("expected evaluation error")
	}
	if !strings.Contains(err.Error(), "syntax error near line 1") {
		t.Fatalf("expected stderr in error, got: %v", err)
	}
}

func TestStructuralCompare_NumberTolerance(t *testing.T) {
	err := StructuralCompare(float64(1.0000000001), float64(1.0000000002))
	if err != nil {
		t.Fatalf("expected numbers within tolerance to match: %v", err)
	}
}

func TestStructuralCompare_ObjectKeyOrderAndMissingNull(t *testing.T) {
	left := map[string]interface{}{
		"a": float64(1),
		"b": nil,
	}
	right := map[string]interface{}{
		"b": nil,
		"a": float64(1),
	}
	if err := StructuralCompare(left, right); err != nil {
		t.Fatalf("key order should be ignored: %v", err)
	}

	missingLeft := map[string]interface{}{"a": float64(1)}
	missingRight := map[string]interface{}{"a": float64(1), "b": nil}
	if err := StructuralCompare(missingLeft, missingRight); err != nil {
		t.Fatalf("missing key should match explicit null: %v", err)
	}
}

func TestStructuralCompare_MismatchHasPath(t *testing.T) {
	left := map[string]interface{}{
		"a": []interface{}{
			map[string]interface{}{"x": float64(1)},
		},
	}
	right := map[string]interface{}{
		"a": []interface{}{
			map[string]interface{}{"x": float64(2)},
		},
	}

	err := StructuralCompare(left, right)
	if err == nil {
		t.Fatal("expected mismatch")
	}
	if !strings.Contains(err.Error(), "$.a[0].x") {
		t.Fatalf("expected path-aware diff, got: %v", err)
	}
}

func TestParseDWOutput_NormalizesNullLikeValues(t *testing.T) {
	tests := []string{"null", "nil", "undefined"}
	for _, input := range tests {
		got, err := parseDWOutput(input)
		if err != nil {
			t.Fatalf("parseDWOutput(%q): %v", input, err)
		}
		if got != nil {
			t.Fatalf("expected nil for %q, got %#v", input, got)
		}
	}
}
