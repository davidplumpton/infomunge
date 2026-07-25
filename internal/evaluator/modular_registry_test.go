package evaluator

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
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
	RegisterBuiltinSpecial("runtimeSpecial", func(_ *ast.CallExpr, _ *Scope, _ int) (Value, error) {
		return true, nil
	})

	if _, ok := GetBuiltinFunction("runtimeFn"); !ok {
		t.Fatalf("expected runtimeFn to be registered")
	}
	if _, ok := GetBuiltinSpecial("runtimeSpecial"); !ok {
		t.Fatalf("expected runtimeSpecial to be registered")
	}

	if spec, ok := GetBuiltinSpec("runtimeFn"); !ok {
		t.Fatalf("expected runtimeFn spec to be registered")
	} else if spec.Mode != BuiltinEvaluationEager {
		t.Fatalf("expected runtimeFn to be eager, got %s", spec.Mode)
	}
	if spec, ok := GetBuiltinSpec("runtimeSpecial"); !ok {
		t.Fatalf("expected runtimeSpecial spec to be registered")
	} else if spec.Mode != BuiltinEvaluationSpecial {
		t.Fatalf("expected runtimeSpecial to be special, got %s", spec.Mode)
	}
}

func TestBuildBuiltinRegistriesRejectsInvalidSpecs(t *testing.T) {
	validRegular := regularBuiltinSpec("validRegular", builtinCategoryCore, exactArity(0), func(_ []Value, _ *ast.CallExpr) (Value, error) {
		return "regular", nil
	}, "valid regular")
	validSpecial := specialBuiltinSpec("validSpecial", builtinCategoryCore, exactArity(0), func(_ *ast.CallExpr, _ *Scope, _ int) (Value, error) {
		return "special", nil
	}, "valid special")

	tests := []struct {
		name  string
		specs []BuiltinSpec
	}{
		{
			name: "duplicate names",
			specs: []BuiltinSpec{
				validRegular,
				regularBuiltinSpec("validRegular", builtinCategoryCore, exactArity(1), func(_ []Value, _ *ast.CallExpr) (Value, error) {
					return nil, nil
				}, "duplicate"),
			},
		},
		{name: "empty name", specs: []BuiltinSpec{{Category: builtinCategoryCore, Module: "internal::core", Arity: exactArity(0), Mode: BuiltinEvaluationEager, Regular: validRegular.Regular, Summary: "missing name"}}},
		{name: "empty category", specs: []BuiltinSpec{{Name: "missingCategory", Module: "internal::core", Arity: exactArity(0), Mode: BuiltinEvaluationEager, Regular: validRegular.Regular, Summary: "missing category"}}},
		{name: "empty module", specs: []BuiltinSpec{{Name: "missingModule", Category: builtinCategoryCore, Arity: exactArity(0), Mode: BuiltinEvaluationEager, Regular: validRegular.Regular, Summary: "missing module"}}},
		{name: "invalid arity", specs: []BuiltinSpec{{Name: "badArity", Category: builtinCategoryCore, Module: "internal::core", Arity: BuiltinArity{Min: 2, Max: 1}, Mode: BuiltinEvaluationEager, Regular: validRegular.Regular, Summary: "bad arity"}}},
		{name: "empty summary", specs: []BuiltinSpec{{Name: "missingSummary", Category: builtinCategoryCore, Module: "internal::core", Arity: exactArity(0), Mode: BuiltinEvaluationEager, Regular: validRegular.Regular}}},
		{name: "nil regular handler", specs: []BuiltinSpec{{Name: "nilRegular", Category: builtinCategoryCore, Module: "internal::core", Arity: exactArity(0), Mode: BuiltinEvaluationEager, Summary: "nil regular"}}},
		{name: "nil special handler", specs: []BuiltinSpec{{Name: "nilSpecial", Category: builtinCategoryCore, Module: "internal::core", Arity: exactArity(0), Mode: BuiltinEvaluationSpecial, Summary: "nil special"}}},
		{name: "unknown mode", specs: []BuiltinSpec{{Name: "badMode", Category: builtinCategoryCore, Module: "internal::core", Arity: exactArity(0), Mode: BuiltinEvaluationMode("unknown"), Regular: validRegular.Regular, Summary: "bad mode"}}},
		{name: "special with regular handler", specs: []BuiltinSpec{{
			Name:     "mixedSpecial",
			Category: builtinCategoryCore,
			Module:   "internal::core",
			Arity:    exactArity(0),
			Mode:     BuiltinEvaluationSpecial,
			Special:  validSpecial.Special,
			Regular:  validRegular.Regular,
			Summary:  "mixed special",
		}}},
		{name: "regular with special handler", specs: []BuiltinSpec{{
			Name:     "mixedRegular",
			Category: builtinCategoryCore,
			Module:   "internal::core",
			Arity:    exactArity(0),
			Mode:     BuiltinEvaluationEager,
			Regular:  validRegular.Regular,
			Special:  validSpecial.Special,
			Summary:  "mixed regular",
		}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, _, err := buildBuiltinRegistries(tt.specs); err == nil {
				t.Fatalf("expected invalid specs to fail")
			}
		})
	}
}

func TestBuildBuiltinRegistriesDispatchesByEvaluationMode(t *testing.T) {
	specialCalled := false
	regularCalled := false

	special, regular, specs, err := buildBuiltinRegistries([]BuiltinSpec{
		regularBuiltinSpec("regularOnly", builtinCategoryCore, exactArity(0), func(_ []Value, _ *ast.CallExpr) (Value, error) {
			regularCalled = true
			return "regular", nil
		}, "regular only"),
		specialBuiltinSpec("specialOnly", builtinCategoryCore, exactArity(0), func(_ *ast.CallExpr, _ *Scope, _ int) (Value, error) {
			specialCalled = true
			return "special", nil
		}, "special only"),
	})
	if err != nil {
		t.Fatalf("unexpected error building registry: %v", err)
	}

	if _, ok := regular["regularOnly"]; !ok {
		t.Fatalf("expected regularOnly in regular registry")
	}
	if _, ok := special["regularOnly"]; ok {
		t.Fatalf("did not expect regularOnly in special registry")
	}
	if _, ok := special["specialOnly"]; !ok {
		t.Fatalf("expected specialOnly in special registry")
	}
	if _, ok := regular["specialOnly"]; ok {
		t.Fatalf("did not expect specialOnly in regular registry")
	}
	if specs["regularOnly"].Mode != BuiltinEvaluationEager {
		t.Fatalf("expected regularOnly spec to be eager")
	}
	if specs["specialOnly"].Mode != BuiltinEvaluationSpecial {
		t.Fatalf("expected specialOnly spec to be special")
	}

	invalidCall := &ast.CallExpr{
		Fun:  &ast.Ident{NamePos: token.Pos(9), Name: "invalid"},
		Args: []ast.Expr{&ast.Ident{NamePos: token.Pos(10), Name: "arg"}},
	}
	if _, err := regular["regularOnly"]([]Value{nil}, invalidCall); err == nil {
		t.Fatal("expected regular dispatch to reject an argument outside its declared arity")
	}
	if _, err := special["specialOnly"](invalidCall, nil, 0); err == nil {
		t.Fatal("expected special dispatch to reject an argument outside its declared arity")
	}
	if regularCalled || specialCalled {
		t.Fatal("arity validation must run before builtin handlers")
	}

	if got, err := regular["regularOnly"](nil, nil); err != nil || got != "regular" {
		t.Fatalf("regular dispatch = %v, %v", got, err)
	}
	if got, err := special["specialOnly"](&ast.CallExpr{}, nil, 0); err != nil || got != "special" {
		t.Fatalf("special dispatch = %v, %v", got, err)
	}
	if !regularCalled || !specialCalled {
		t.Fatalf("expected both handlers to be called")
	}
}

func TestBuildBuiltinRegistriesEnforcesAdvertisedArity(t *testing.T) {
	special, regular, specs, err := buildBuiltinRegistries(defaultBuiltinSpecs())
	if err != nil {
		t.Fatalf("buildBuiltinRegistries() error = %v", err)
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			if spec.Arity.Min > 0 {
				assertBuiltinArityRejected(t, spec, special, regular, spec.Arity.Min-1)
			}
			if spec.Arity.Max != BuiltinArityVariadic {
				assertBuiltinArityRejected(t, spec, special, regular, spec.Arity.Max+1)
			}
		})
	}
}

func TestDefaultBuiltinSpecsMatchPublicArityContracts(t *testing.T) {
	expected := map[string]BuiltinArity{
		"__case":        exactArity(2),
		"__coerce":      rangeArity(2, 3),
		"__do":          exactArity(2),
		"__lazyReduce":  exactArity(3),
		"__modcall":     variadicArity(2),
		"__multival":    exactArity(2),
		"assert":        rangeArity(2, 3),
		"beGreaterThan": rangeArity(1, 2),
		"beLowerThan":   rangeArity(1, 2),
		"date":          exactArity(3),
		"dateTime":      rangeArity(6, 7),
		"hash":          exactArity(2),
		"leftPad":       exactArity(3),
		"localDateTime": exactArity(6),
		"localTime":     exactArity(3),
		"log":           exactArity(1),
		"logDebug":      exactArity(1),
		"logError":      exactArity(1),
		"logInfo":       exactArity(1),
		"logWarn":       exactArity(1),
		"match":         rangeArity(2, 3),
		"matches":       rangeArity(2, 3),
		"rightPad":      exactArity(3),
		"scan":          rangeArity(2, 3),
		"time":          rangeArity(3, 4),
	}

	_, _, specs, err := buildBuiltinRegistries(defaultBuiltinSpecs())
	if err != nil {
		t.Fatalf("buildBuiltinRegistries() error = %v", err)
	}
	for name, want := range expected {
		spec, ok := specs[name]
		if !ok {
			t.Errorf("missing builtin spec %q", name)
			continue
		}
		if spec.Arity != want {
			t.Errorf("%s arity = %+v, want %+v", name, spec.Arity, want)
		}
	}
}

func assertBuiltinArityRejected(
	t *testing.T,
	spec BuiltinSpec,
	special map[string]SpecialBuiltinFunc,
	regular map[string]RegularBuiltinFunc,
	count int,
) {
	t.Helper()

	const callPos = token.Pos(17)
	call := &ast.CallExpr{
		Fun:  &ast.Ident{NamePos: callPos, Name: spec.Name},
		Args: make([]ast.Expr, count),
	}
	for i := range call.Args {
		call.Args[i] = &ast.Ident{NamePos: callPos + token.Pos(i+1), Name: "arg"}
	}

	var err error
	switch spec.Mode {
	case BuiltinEvaluationSpecial:
		_, err = special[spec.Name](call, nil, 0)
	case BuiltinEvaluationEager:
		_, err = regular[spec.Name](make([]Value, count), call)
	default:
		t.Fatalf("unexpected evaluation mode %q", spec.Mode)
	}
	if err == nil {
		t.Fatalf("%s accepted %d arguments outside advertised range %+v", spec.Name, count, spec.Arity)
	}
	expectedMessage := spec.Name
	if count < spec.Arity.Min && spec.tooFewArgumentsMessage != "" {
		expectedMessage = spec.tooFewArgumentsMessage
	}
	if count > spec.Arity.Max && spec.tooManyArgumentsMessage != "" {
		expectedMessage = spec.tooManyArgumentsMessage
	}
	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf("arity error %q does not contain %q", err, expectedMessage)
	}

	positioned, ok := err.(interface {
		GeneratedPosition() (token.Pos, bool)
	})
	if !ok {
		t.Fatalf("arity error %T does not expose its generated position", err)
	}
	if got, ok := positioned.GeneratedPosition(); !ok || got != callPos {
		t.Errorf("arity error position = (%d, %t), want (%d, true)", got, ok, callPos)
	}
}

func TestBuiltinSpecsByCategory(t *testing.T) {
	grouped := BuiltinSpecsByCategory()

	arraySpecs := grouped[builtinCategoryArrays]
	var sizeOfSpec BuiltinSpec
	for _, spec := range arraySpecs {
		if spec.Name == "sizeOf" {
			sizeOfSpec = spec
			break
		}
	}
	if sizeOfSpec.Name == "" {
		t.Fatalf("expected sizeOf to be listed under arrays category")
	}
	if sizeOfSpec.Mode != BuiltinEvaluationEager {
		t.Fatalf("expected sizeOf to be eager, got %s", sizeOfSpec.Mode)
	}
	if sizeOfSpec.Arity != exactArity(1) {
		t.Fatalf("expected sizeOf arity 1, got %+v", sizeOfSpec.Arity)
	}
	if sizeOfSpec.Module == "" || sizeOfSpec.Summary == "" {
		t.Fatalf("expected sizeOf metadata to include module and summary")
	}

	coreSpecs := grouped[builtinCategoryCore]
	for _, spec := range coreSpecs {
		if spec.Name == "__ifelse" && spec.Mode == BuiltinEvaluationSpecial {
			return
		}
	}
	t.Fatalf("expected __ifelse special builtin under core category")
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
