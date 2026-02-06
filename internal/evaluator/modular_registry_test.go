package evaluator

import (
	"fmt"
	"go/ast"
	"sync"
	"testing"
)

func TestSetBuiltinRegistriesForTesting_Restore(t *testing.T) {
	_, hasDefault := GetBuiltinFunction("sizeOf")
	if !hasDefault {
		t.Fatalf("expected default builtin sizeOf to exist")
	}

	restore := SetBuiltinRegistriesForTesting(
		map[string]SpecialBuiltinFunc{},
		map[string]RegularBuiltinFunc{
			"customFn": func(_ []Value, _ *ast.CallExpr) (Value, error) {
				return "ok", nil
			},
		},
	)

	if _, ok := GetBuiltinFunction("sizeOf"); ok {
		t.Fatalf("expected default registry to be replaced in test")
	}

	fn, ok := GetBuiltinFunction("customFn")
	if !ok {
		t.Fatalf("expected customFn to be registered in test registry")
	}
	value, err := fn(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error calling customFn: %v", err)
	}
	if value != "ok" {
		t.Fatalf("expected customFn to return ok, got %v", value)
	}

	restore()

	if _, ok := GetBuiltinFunction("sizeOf"); !ok {
		t.Fatalf("expected restore to bring back default registry")
	}
	if _, ok := GetBuiltinFunction("customFn"); ok {
		t.Fatalf("expected customFn to be removed after restore")
	}
}

func TestRegisterBuiltinFunctions(t *testing.T) {
	restore := SetBuiltinRegistriesForTesting(map[string]SpecialBuiltinFunc{}, map[string]RegularBuiltinFunc{})
	defer restore()

	RegisterBuiltinFunction("runtimeFn", func(_ []Value, _ *ast.CallExpr) (Value, error) {
		return 42, nil
	})
	RegisterBuiltinSpecial("runtimeSpecial", func(_ *ast.CallExpr, _ Context, _ int) (Value, error) {
		return true, nil
	})

	if _, ok := GetBuiltinFunction("runtimeFn"); !ok {
		t.Fatalf("expected runtimeFn to be registered")
	}
	if _, ok := GetBuiltinSpecial("runtimeSpecial"); !ok {
		t.Fatalf("expected runtimeSpecial to be registered")
	}
}

func TestBuiltinRegistryConcurrentAccess(t *testing.T) {
	restore := SetBuiltinRegistriesForTesting(map[string]SpecialBuiltinFunc{}, map[string]RegularBuiltinFunc{})
	defer restore()

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			name := fmt.Sprintf("fn_%d", i)
			RegisterBuiltinFunction(name, func(_ []Value, _ *ast.CallExpr) (Value, error) {
				return i, nil
			})
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = i
			for j := 0; j < 20; j++ {
				_, _ = GetBuiltinFunction("fn_0")
			}
		}()
	}

	wg.Wait()

	for i := 0; i < 50; i++ {
		name := fmt.Sprintf("fn_%d", i)
		if _, ok := GetBuiltinFunction(name); !ok {
			t.Fatalf("expected %s to be registered", name)
		}
	}
}
