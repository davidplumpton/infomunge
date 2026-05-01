package preprocessor

import (
	"fmt"
	"go/parser"
	"strings"
	"testing"

	"infomunge/internal/sourcemap"
)

func TestOperatorProcessingContractsUseExplicitBinaryOperatorPath(t *testing.T) {
	contracts := operatorProcessingContracts()
	wantNames := []string{
		"replaceDefaultOperator",
		"replaceOnNullOperator",
		"replaceThenOperator",
		"replaceToOperator",
	}
	if len(contracts) != len(wantNames) {
		t.Fatalf("expected %d contracts, got %d", len(wantNames), len(contracts))
	}

	previousOrder := 0
	for i, contract := range contracts {
		if contract.Name != wantNames[i] {
			t.Fatalf("contract[%d] name = %q, want %q", i, contract.Name, wantNames[i])
		}
		if contract.Phase != TransformPhaseOperator {
			t.Fatalf("%s phase = %q, want %q", contract.Name, contract.Phase, TransformPhaseOperator)
		}
		if contract.Mapping != TransformMappingExact {
			t.Fatalf("%s mapping = %q, want %q", contract.Name, contract.Mapping, TransformMappingExact)
		}
		if contract.Loop != TransformLoopFixpoint {
			t.Fatalf("%s loop = %q, want %q", contract.Name, contract.Loop, TransformLoopFixpoint)
		}
		if contract.Precedence == TransformPrecedenceNone {
			t.Fatalf("%s should declare precedence", contract.Name)
		}
		if contract.Associativity != TransformAssociativityLeft {
			t.Fatalf("%s associativity = %q, want %q", contract.Name, contract.Associativity, TransformAssociativityLeft)
		}
		if contract.Order <= previousOrder {
			t.Fatalf("%s order %d should be after %d", contract.Name, contract.Order, previousOrder)
		}
		previousOrder = contract.Order
	}
}

func TestPrepareForParsing_TraceTransforms(t *testing.T) {
	var trace []TransformTraceEntry
	result, _, err := PrepareForParsing(`payload.name default "missing"`, Options{
		TraceTransforms: func(entry TransformTraceEntry) {
			trace = append(trace, entry)
		},
	})
	if err != nil {
		t.Fatalf("PrepareForParsing returned error: %v", err)
	}
	if result != `__default(payload["name"], "missing")` {
		t.Fatalf("unexpected result: %q", result)
	}

	foundDefault := false
	foundDot := false
	for _, entry := range trace {
		switch entry.Transform {
		case "replaceDefaultOperator":
			foundDefault = entry.Changed &&
				entry.Phase == TransformPhaseOperator &&
				entry.Mapping == TransformMappingExact &&
				entry.Loop == TransformLoopFixpoint
		case "replaceDotNotationWithMapping":
			foundDot = entry.Changed &&
				entry.Phase == TransformPhaseSyntax &&
				entry.Mapping == TransformMappingExact
		}
	}
	if !foundDefault {
		t.Fatalf("trace did not include exact default-operator transform: %#v", trace)
	}
	if !foundDot {
		t.Fatalf("trace did not include exact dot-notation transform: %#v", trace)
	}
}

func TestPrepareForParsing_TraceTransformsFullPreprocessPath(t *testing.T) {
	var trace []TransformTraceEntry
	input := `outer: [1] map name: /a+/ default payload.name // trailing`

	_, _, err := PrepareForParsing(input, Options{
		TraceTransforms: func(entry TransformTraceEntry) {
			trace = append(trace, entry)
		},
	})
	if err != nil {
		t.Fatalf("PrepareForParsing returned error: %v", err)
	}

	wantChanged := map[string]TransformPhase{
		"stripLineComments":               TransformPhaseComment,
		"replaceRegexLiteralsWithMapping": TransformPhaseRegex,
		"wrapImplicitObjectLiteralBodies": TransformPhaseWrapper,
		"wrapTopLevelObjectLiteral":       TransformPhaseWrapper,
		"rewriteCoreSyntax":               TransformPhaseRewrite,
		"replaceDefaultOperator":          TransformPhaseOperator,
		"replaceDotNotationWithMapping":   TransformPhaseSyntax,
	}
	for transform, phase := range wantChanged {
		entry, ok := changedTraceEntry(trace, transform)
		if !ok {
			t.Fatalf("trace did not include changed transform %q: %#v", transform, trace)
		}
		if entry.Phase != phase {
			t.Fatalf("%s phase = %q, want %q", transform, entry.Phase, phase)
		}
	}

	wantOrder := []string{
		"stripLineComments",
		"replaceRegexLiteralsWithMapping",
		"wrapImplicitObjectLiteralBodies",
		"wrapTopLevelObjectLiteral",
		"rewriteCoreSyntax",
		"replaceDefaultOperator",
		"replaceDotNotationWithMapping",
	}
	previous := -1
	for _, transform := range wantOrder {
		index := traceIndex(trace, transform)
		if index < 0 {
			t.Fatalf("trace missing %q: %#v", transform, trace)
		}
		if index <= previous {
			t.Fatalf("trace order for %q was %d, want after %d: %#v", transform, index, previous, trace)
		}
		previous = index
	}
}

func TestConfiguredBinaryOperatorSourceMapErrorLocation(t *testing.T) {
	input := `null then (1 + )`
	result, mapping, err := PrepareForParsing(input, Options{})
	if err != nil {
		t.Fatalf("PrepareForParsing returned error: %v", err)
	}

	_, parseErr := parser.ParseExpr(result)
	if parseErr == nil {
		t.Fatalf("expected transformed expression to fail parsing: %q", result)
	}

	formatted := sourcemap.New(input, result, mapping).FormatParseError(parseErr)
	if !strings.Contains(formatted.Error(), "1:16:") {
		t.Fatalf("expected source-mapped error at 1:16, got %q from %q", formatted.Error(), result)
	}
}

func TestFullPreprocessSourceMapComposesRegexWrapperRewriterAndPostTransforms(t *testing.T) {
	input := `foo: /a+/ default (1 + )`
	result, mapping, err := PrepareForParsing(input, Options{})
	if err != nil {
		t.Fatalf("PrepareForParsing returned error: %v", err)
	}

	for _, required := range []string{
		`regex("a+")`,
		`map[string]interface{}`,
		`__default(regex("a+"), (1 + ))`,
	} {
		if !strings.Contains(result, required) {
			t.Fatalf("result missing %q: %q", required, result)
		}
	}

	_, parseErr := parser.ParseExpr(result)
	if parseErr == nil {
		t.Fatalf("expected transformed expression to fail parsing: %q", result)
	}

	formatted := sourcemap.New(input, result, mapping).FormatParseError(parseErr)
	wantColumn := strings.LastIndex(input, ")") + 1
	wantPosition := fmt.Sprintf("1:%d:", wantColumn)
	if !strings.Contains(formatted.Error(), wantPosition) {
		t.Fatalf("expected source-mapped error at %s, got %q from %q", wantPosition, formatted.Error(), result)
	}
}

func changedTraceEntry(trace []TransformTraceEntry, transform string) (TransformTraceEntry, bool) {
	for _, entry := range trace {
		if entry.Transform == transform && entry.Changed {
			return entry, true
		}
	}
	return TransformTraceEntry{}, false
}

func traceIndex(trace []TransformTraceEntry, transform string) int {
	for i, entry := range trace {
		if entry.Transform == transform {
			return i
		}
	}
	return -1
}
