package preprocessor

import (
	"fmt"
	"go/parser"
	"reflect"
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

func TestFunctionalContractsUseExactTypedOperatorMappings(t *testing.T) {
	contracts := functionalProcessingContracts()
	for _, name := range []string{"replaceAsOperator", "replaceIsOperator"} {
		contract, ok := findTransformContract(contracts, name)
		if !ok {
			t.Fatalf("missing contract %q", name)
		}
		if contract.Mapping != TransformMappingExact {
			t.Fatalf("%s mapping = %q, want %q", name, contract.Mapping, TransformMappingExact)
		}
		if contract.Loop != TransformLoopFixpoint {
			t.Fatalf("%s loop = %q, want %q", name, contract.Loop, TransformLoopFixpoint)
		}
	}
}

func TestConfiguredBinaryOperatorsPreserveMixedLeftAssociativity(t *testing.T) {
	tests := []struct {
		name     string
		stage    PipelineStage
		input    string
		expected string
	}{
		{
			name:     "onNull before then",
			stage:    createOperatorProcessingStage(nil),
			input:    "a onNull b then c",
			expected: "then(onNull(a, b), c)",
		},
		{
			name:     "then before onNull",
			stage:    createOperatorProcessingStage(nil),
			input:    "a then b onNull c",
			expected: "onNull(then(a, b), c)",
		},
		{
			name:     "splitBy before joinBy",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a splitBy b joinBy c",
			expected: "joinBy(splitBy(a, b), c)",
		},
		{
			name:     "joinBy before splitBy",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a joinBy b splitBy c",
			expected: "splitBy(joinBy(a, b), c)",
		},
		{
			name:     "all collection operators",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a ~ b find c splitBy d joinBy e",
			expected: "joinBy(splitBy(find(__update(a, b), c), d), e)",
		},
		{
			name:     "concatenate before remove",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a ++ b -- c",
			expected: "__remove(__concat(a, b), c)",
		},
		{
			name:     "remove before concatenate",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a -- b ++ c",
			expected: "__concat(__remove(a, b), c)",
		},
		{
			name:     "contains before matches",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a contains b matches c",
			expected: "matches(contains(a, b), c)",
		},
		{
			name:     "matches before contains",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a matches b contains c",
			expected: "contains(matches(a, b), c)",
		},
		{
			name:     "all comparison operators",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a contains b match c matches d",
			expected: "matches(match(contains(a, b), c), d)",
		},
		{
			name:     "mod before repeat",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a mod b repeat c",
			expected: "repeat(mod(a, b), c)",
		},
		{
			name:     "repeat before mod",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a repeat b mod c",
			expected: "mod(repeat(a, b), c)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, mapping, err := tt.stage.Execute(tt.input, identityMapping(len(tt.input)))
			if err != nil {
				t.Fatalf("stage.Execute() error = %v", err)
			}
			if got != tt.expected {
				t.Fatalf("stage.Execute() = %q, want %q", got, tt.expected)
			}
			if len(mapping) != len(got) {
				t.Fatalf("mapping length = %d, want %d", len(mapping), len(got))
			}
		})
	}
}

func TestFullPreprocessingPipelineUsesStandalonePostProcessingSuffix(t *testing.T) {
	fullNames := CreateFullPreprocessingPipelineWithOptions(Options{}).GetStageNames()
	postNames := CreateModularPostProcessingPipelineWithOptions(Options{}).GetStageNames()

	if len(fullNames) < len(postNames) {
		t.Fatalf("full pipeline stages %v shorter than post pipeline stages %v", fullNames, postNames)
	}

	fullSuffix := fullNames[len(fullNames)-len(postNames):]
	if !reflect.DeepEqual(fullSuffix, postNames) {
		t.Fatalf("full pipeline suffix = %v, want post pipeline stages %v", fullSuffix, postNames)
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

func TestTypedOperatorSourceMapErrorLocation(t *testing.T) {
	input := `1 as Number {format: 1 + }`
	result, mapping, err := PrepareForParsing(input, Options{})
	if err != nil {
		t.Fatalf("PrepareForParsing returned error: %v", err)
	}

	if !strings.Contains(result, `__coerce(1, "Number", map[string]interface{}{`) {
		t.Fatalf("result did not use mapped as-operator rewrite: %q", result)
	}

	_, parseErr := parser.ParseExpr(result)
	if parseErr == nil {
		t.Fatalf("expected transformed expression to fail parsing: %q", result)
	}

	formatted := sourcemap.New(input, result, mapping).FormatParseError(parseErr)
	if !strings.Contains(formatted.Error(), "1:26:") {
		t.Fatalf("expected source-mapped error at 1:26, got %q from %q", formatted.Error(), result)
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

func findTransformContract(contracts []TransformContract, name string) (TransformContract, bool) {
	for _, contract := range contracts {
		if contract.Name == name {
			return contract, true
		}
	}
	return TransformContract{}, false
}

func traceIndex(trace []TransformTraceEntry, transform string) int {
	for i, entry := range trace {
		if entry.Transform == transform {
			return i
		}
	}
	return -1
}
