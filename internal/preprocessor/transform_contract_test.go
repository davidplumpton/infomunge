package preprocessor

import (
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
