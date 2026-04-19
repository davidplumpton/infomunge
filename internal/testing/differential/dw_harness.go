package differential

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"infomunge/internal/evaluator"
	"infomunge/pkg/formats"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	dwEvalTimeout    = 10 * time.Second
	floatTolerance   = 1e-9
	floatAbsFloor    = 1.0
	dwNullString     = "null"
	dwNilString      = "nil"
	dwUndefinedValue = "undefined"
)

var runDWCommand = defaultRunDWCommand

// DWAvailable reports whether the DataWeave CLI is available on PATH.
func DWAvailable() bool {
	_, err := exec.LookPath("dw")
	return err == nil
}

// DWEval executes a DataWeave script with optional named inputs and returns
// a normalized parsed value from stdout.
func DWEval(script string, inputs map[string]string) (evaluator.Value, error) {
	if strings.TrimSpace(script) == "" {
		return nil, errors.New("dw script cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), dwEvalTimeout)
	defer cancel()

	stdout, stderr, err := runDWCommand(ctx, script, inputs)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("dw evaluation timed out after %s", dwEvalTimeout)
		}
		if strings.TrimSpace(stderr) != "" {
			return nil, fmt.Errorf("dw evaluation failed: %s: %w", strings.TrimSpace(stderr), err)
		}
		return nil, fmt.Errorf("dw evaluation failed: %w", err)
	}

	parsed, err := parseDWOutput(stdout)
	if err != nil {
		return nil, fmt.Errorf("parse dw output: %w", err)
	}
	return normalizeValue(parsed), nil
}

func defaultRunDWCommand(ctx context.Context, script string, inputs map[string]string) (string, string, error) {
	args := []string{"run"}
	inputArgs, cleanup, err := prepareDWInputArgs(inputs)
	if err != nil {
		return "", "", err
	}
	defer cleanup()

	args = append(args, inputArgs...)
	args = append(args, script)
	cmd := exec.CommandContext(ctx, "dw", args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	return stdout.String(), stderr.String(), err
}

func prepareDWInputArgs(inputs map[string]string) ([]string, func(), error) {
	if len(inputs) == 0 {
		return nil, func() {}, nil
	}

	dir, err := os.MkdirTemp("", "infomunge-dw-inputs-*")
	if err != nil {
		return nil, nil, fmt.Errorf("create dw harness input dir: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(dir) }

	keys := make([]string, 0, len(inputs))
	for k := range inputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	args := make([]string, 0, len(inputs))
	for idx, name := range keys {
		filename := fmt.Sprintf("%03d_%s.input", idx, name)
		filePath := filepath.Join(dir, filename)
		if err := os.WriteFile(filePath, []byte(inputs[name]), 0o600); err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("write dw harness input %q: %w", name, err)
		}
		args = append(args, fmt.Sprintf("-i=%s=%s", name, filePath))
	}

	return args, cleanup, nil
}

func parseDWOutput(output string) (evaluator.Value, error) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil, nil
	}

	if parsed, err := formats.Read(trimmed, "application/json"); err == nil {
		return parsed, nil
	}

	switch strings.ToLower(trimmed) {
	case dwNullString, dwNilString, dwUndefinedValue:
		return nil, nil
	}

	if b, err := strconv.ParseBool(trimmed); err == nil {
		return b, nil
	}
	if i, err := strconv.Atoi(trimmed); err == nil {
		return i, nil
	}
	if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return f, nil
	}
	if unquoted, err := strconv.Unquote(trimmed); err == nil {
		return unquoted, nil
	}

	return trimmed, nil
}

func normalizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case nil:
		return nil
	case string:
		switch strings.ToLower(strings.TrimSpace(val)) {
		case dwNullString, dwNilString, dwUndefinedValue:
			return nil
		default:
			return val
		}
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case evaluator.Array:
		out := make(evaluator.Array, len(val))
		for i := range val {
			out[i] = normalizeValue(val[i])
		}
		return out
	case evaluator.Object:
		out := make(evaluator.Object, len(val))
		for k, inner := range val {
			out[k] = normalizeValue(inner)
		}
		return out
	default:
		return val
	}
}

// StructuralCompare compares two results after normalization and returns a
// detailed path-aware mismatch if values differ.
func StructuralCompare(infomungeResult, dwResult interface{}) error {
	left := normalizeValue(infomungeResult)
	right := normalizeValue(dwResult)
	return compareValues("$", left, right)
}

func compareValues(path string, left, right interface{}) error {
	if left == nil && right == nil {
		return nil
	}

	if lf, okL := toFloat64(left); okL {
		if rf, okR := toFloat64(right); okR {
			if numbersClose(lf, rf) {
				return nil
			}
			return fmt.Errorf("%s: numeric mismatch infomunge=%v dw=%v", path, lf, rf)
		}
	}

	if lObj, ok := left.(evaluator.Object); ok {
		rObj, ok := right.(evaluator.Object)
		if !ok {
			return fmt.Errorf("%s: type mismatch infomunge=%T dw=%T", path, left, right)
		}
		return compareObjects(path, lObj, rObj)
	}

	if lArr, ok := left.(evaluator.Array); ok {
		rArr, ok := right.(evaluator.Array)
		if !ok {
			return fmt.Errorf("%s: type mismatch infomunge=%T dw=%T", path, left, right)
		}
		if len(lArr) != len(rArr) {
			return fmt.Errorf("%s: array length mismatch infomunge=%d dw=%d", path, len(lArr), len(rArr))
		}
		for i := range lArr {
			if err := compareValues(fmt.Sprintf("%s[%d]", path, i), lArr[i], rArr[i]); err != nil {
				return err
			}
		}
		return nil
	}

	if reflect.DeepEqual(left, right) {
		return nil
	}

	return fmt.Errorf("%s: value mismatch infomunge=%#v dw=%#v", path, left, right)
}

func compareObjects(path string, left, right evaluator.Object) error {
	keys := make(map[string]struct{}, len(left)+len(right))
	for k := range left {
		keys[k] = struct{}{}
	}
	for k := range right {
		keys[k] = struct{}{}
	}

	for key := range keys {
		lv, lok := left[key]
		rv, rok := right[key]

		if !lok {
			lv = nil
		}
		if !rok {
			rv = nil
		}

		if err := compareValues(path+"."+key, lv, rv); err != nil {
			return err
		}
	}

	return nil
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	default:
		return 0, false
	}
}

func numbersClose(a, b float64) bool {
	diff := math.Abs(a - b)
	if diff <= floatTolerance {
		return true
	}
	scale := math.Max(floatAbsFloor, math.Max(math.Abs(a), math.Abs(b)))
	return diff <= floatTolerance*scale
}
