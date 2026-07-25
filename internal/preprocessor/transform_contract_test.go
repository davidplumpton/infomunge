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

func TestModContractDeclaresDataWeavePrecedence(t *testing.T) {
	contract, ok := findTransformContract(functionalProcessingContracts(), "replaceModOperator")
	if !ok {
		t.Fatal("missing replaceModOperator contract")
	}
	if contract.Precedence != TransformPrecedenceModulo {
		t.Fatalf("mod precedence = %d, want %d", contract.Precedence, TransformPrecedenceModulo)
	}
	if contract.Associativity != TransformAssociativityLeft {
		t.Fatalf("mod associativity = %q, want %q", contract.Associativity, TransformAssociativityLeft)
	}
	if !(TransformPrecedenceComparison < contract.Precedence &&
		contract.Precedence < TransformPrecedenceAdditive) {
		t.Fatalf(
			"mod precedence %d should be between comparison %d and additive %d",
			contract.Precedence,
			TransformPrecedenceComparison,
			TransformPrecedenceAdditive,
		)
	}
}

func TestRangeOperatorStopsBeforeDownstreamOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "map",
			input:    `1 to 3 map __lambda("x", x * 2)`,
			expected: `to(1, 3) map __lambda("x", x * 2)`,
		},
		{
			name:     "filter",
			input:    `1 to 3 filter __lambda("x", x > 1)`,
			expected: `to(1, 3) filter __lambda("x", x > 1)`,
		},
		{
			name:     "reduce",
			input:    `1 to 4 reduce __lambda("acc, x", acc + x)`,
			expected: `to(1, 4) reduce __lambda("acc, x", acc + x)`,
		},
		{
			name:     "joinBy",
			input:    `1 to 3 joinBy "-"`,
			expected: `to(1, 3) joinBy "-"`,
		},
		{
			name:     "concatenate",
			input:    `1 to 2 ++ [3]`,
			expected: `to(1, 2) ++ [3]`,
		},
		{
			name:     "remove",
			input:    `1 to 3 -- [2]`,
			expected: `to(1, 3) -- [2]`,
		},
		{
			name:     "find",
			input:    `1 to 3 find 2`,
			expected: `to(1, 3) find 2`,
		},
		{
			name:     "contains",
			input:    `1 to 3 contains 2`,
			expected: `to(1, 3) contains 2`,
		},
		{
			name:     "operator name can begin upper bound",
			input:    `1 to map + 2`,
			expected: `to(1, map + 2)`,
		},
	}

	stage := createOperatorProcessingStage(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, mapping, err := stage.Execute(tt.input, identityMapping(len(tt.input)))
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
	for _, operator := range CollectionOperators {
		t.Run("canonical collection operator/"+operator, func(t *testing.T) {
			input := `1 to 3 ` + operator + ` __lambda("x", x)`
			expected := `to(1, 3) ` + operator + ` __lambda("x", x)`
			got, _, err := stage.Execute(input, identityMapping(len(input)))
			if err != nil {
				t.Fatalf("stage.Execute() error = %v", err)
			}
			if got != expected {
				t.Fatalf("stage.Execute() = %q, want %q", got, expected)
			}
		})
	}
}

func TestRangeOperatorPreservesUpperBoundGroupingAndAssociativity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "parentheses override downstream grouping",
			input:    `1 to (2 ++ [3])`,
			expected: `to(1, (2 ++ [3]))`,
		},
		{
			name:     "negative fractional arithmetic bound feeds map",
			input:    `-2.5 to 1.5 + 1 map __lambda("x", x)`,
			expected: `to(-2.5, 1.5 + 1) map __lambda("x", x)`,
		},
		{
			name:     "chained range remains left associative",
			input:    `1 to 3 to 4 joinBy "-"`,
			expected: `to(to(1, 3), 4) joinBy "-"`,
		},
		{
			name:     "comparison remains in upper bound",
			input:    `1 to 3 == [1, 2, 3]`,
			expected: `to(1, 3 == [1, 2, 3])`,
		},
	}

	stage := createOperatorProcessingStage(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := stage.Execute(tt.input, identityMapping(len(tt.input)))
			if err != nil {
				t.Fatalf("stage.Execute() error = %v", err)
			}
			if got != tt.expected {
				t.Fatalf("stage.Execute() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestRangeSelectorStillFeedsCollectionPipeline(t *testing.T) {
	input := `[10, 20, 30][0 to 1] map $ * 2`
	got, _, err := PrepareForParsing(input, Options{})
	if err != nil {
		t.Fatalf("PrepareForParsing() error = %v", err)
	}
	expected := `__map(__rangeIndex([]interface{}{10, 20, 30,}, 0, 1), __lambda("__arg", __arg * 2))`
	if got != expected {
		t.Fatalf("PrepareForParsing() = %q, want %q", got, expected)
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
			name:     "splitBy before concatenate",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a splitBy b ++ c",
			expected: "__concat(splitBy(a, b), c)",
		},
		{
			name:     "joinBy before concatenate",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a joinBy b ++ c",
			expected: "__concat(joinBy(a, b), c)",
		},
		{
			name:     "parenthesized concatenate remains joinBy delimiter",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a joinBy (b ++ c)",
			expected: "joinBy(a, (__concat(b, c)))",
		},
		{
			name:     "parenthesized concatenate remains splitBy delimiter",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a splitBy (b ++ c)",
			expected: "splitBy(a, (__concat(b, c)))",
		},
		{
			name:     "joinBy chain remains left associative before concatenate",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a joinBy b joinBy c ++ d",
			expected: "__concat(joinBy(joinBy(a, b), c), d)",
		},
		{
			name:     "concatenate remains complete before find",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a ++ b find c",
			expected: "find(__concat(a, b), c)",
		},
		{
			name:     "concatenate remains complete before contains",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a ++ b contains c",
			expected: "contains(__concat(a, b), c)",
		},
		{
			name:     "concatenate remains complete before splitBy",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a ++ b splitBy c",
			expected: "splitBy(__concat(a, b), c)",
		},
		{
			name:     "concatenate remains complete before joinBy",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a ++ b joinBy c",
			expected: "joinBy(__concat(a, b), c)",
		},
		{
			name:     "concatenate remains complete before match",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a ++ b match c",
			expected: "match(__concat(a, b), c)",
		},
		{
			name:     "concatenate remains complete before matches",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a ++ b matches c",
			expected: "matches(__concat(a, b), c)",
		},
		{
			name:     "concatenate remains complete before scan",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a ++ b scan c",
			expected: "scan(__concat(a, b), c)",
		},
		{
			name:     "removal remains complete before joinBy",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a -- b joinBy c",
			expected: "joinBy(__remove(a, b), c)",
		},
		{
			name:     "mixed collection and additive chain",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a splitBy b ++ c joinBy d",
			expected: "joinBy(__concat(splitBy(a, b), c), d)",
		},
		{
			name:     "map receives complete additive source",
			stage:    createFunctionalProcessingStage(nil),
			input:    `a + b map __lambda("x", x + 1)`,
			expected: `__map(a + b, __lambda("x", x + 1))`,
		},
		{
			name:     "map receives complete concatenated source",
			stage:    createFunctionalProcessingStage(nil),
			input:    `a ++ b map __lambda("x", x + 1)`,
			expected: `__map(__concat(a, b), __lambda("x", x + 1))`,
		},
		{
			name:     "map receives complete removal source",
			stage:    createFunctionalProcessingStage(nil),
			input:    `a -- b map __lambda("x", x + 1)`,
			expected: `__map(__remove(a, b), __lambda("x", x + 1))`,
		},
		{
			name:     "map receives left associative mixed additive source",
			stage:    createFunctionalProcessingStage(nil),
			input:    `a ++ b -- c map __lambda("x", x)`,
			expected: `__map(__remove(__concat(a, b), c), __lambda("x", x))`,
		},
		{
			name:     "mapObject receives complete object merge source",
			stage:    createFunctionalProcessingStage(nil),
			input:    `a ++ b mapObject __lambda("v, k", v)`,
			expected: `mapObject(__concat(a, b), __lambda("v, k", v))`,
		},
		{
			name:     "regex splitBy operand before concatenate",
			stage:    createFunctionalProcessingStage(nil),
			input:    `a splitBy regex("b ++ c", "") ++ d`,
			expected: `__concat(splitBy(a, regex("b ++ c", "")), d)`,
		},
		{
			name:     "operator text inside delimiter string",
			stage:    createFunctionalProcessingStage(nil),
			input:    `a joinBy " ++ " ++ c`,
			expected: `__concat(joinBy(a, " ++ "), c)`,
		},
		{
			name:     "nested delimiter call before concatenate",
			stage:    createFunctionalProcessingStage(nil),
			input:    "a joinBy make(b ++ c) ++ d",
			expected: "__concat(joinBy(a, make(__concat(b, c))), d)",
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
			input:    "a contains b match c matches d scan e",
			expected: "scan(matches(match(contains(a, b), c), d), e)",
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
		{
			name:     "default stays inside trailing type coercion",
			stage:    createOperatorProcessingStage(nil),
			input:    `null as String default "x"`,
			expected: `__default(null, "x") as String`,
		},
		{
			name:     "parentheses keep coercion inside default",
			stage:    createOperatorProcessingStage(nil),
			input:    `(null as String) default "x"`,
			expected: `__default((null as String), "x")`,
		},
		{
			name:     "missing default operand preserves typed expression",
			stage:    createOperatorProcessingStage(nil),
			input:    `null as String default `,
			expected: `null as String default `,
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

func TestTypedOperatorsComposeWithConfiguredAndKeywordOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "as before mod",
			input:    `5 as Number mod 2`,
			expected: `mod(__coerce(5, "Number"), 2)`,
		},
		{
			name:     "as before contains",
			input:    `"abc" as String contains "a"`,
			expected: `contains(__coerce("abc", "String"), "a")`,
		},
		{
			name:     "as before concatenation",
			input:    `1 as String ++ "2"`,
			expected: `__concat(__coerce(1, "String"), "2")`,
		},
		{
			name:     "as outside default",
			input:    `null as String default "x"`,
			expected: `__coerce(__default(null, "x"), "String")`,
		},
		{
			name:     "as after mod",
			input:    `5 mod 2 as String`,
			expected: `mod(5, __coerce(2, "String"))`,
		},
		{
			name:     "is before logical operator",
			input:    `1 is Number and true`,
			expected: `__isType(1, "Number") && true`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := PrepareForParsing(tt.input, Options{})
			if err != nil {
				t.Fatalf("PrepareForParsing returned error: %v", err)
			}
			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
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

func TestTypedOperatorPreservesFollowingOperatorSourceMap(t *testing.T) {
	input := `1 as String ++ (2 + )`
	result, mapping, err := PrepareForParsing(input, Options{})
	if err != nil {
		t.Fatalf("PrepareForParsing returned error: %v", err)
	}

	if result != `__concat(__coerce(1, "String"), (2 + ))` {
		t.Fatalf("unexpected transformed expression: %q", result)
	}

	_, parseErr := parser.ParseExpr(result)
	if parseErr == nil {
		t.Fatalf("expected transformed expression to fail parsing: %q", result)
	}

	formatted := sourcemap.New(input, result, mapping).FormatParseError(parseErr)
	if !strings.Contains(formatted.Error(), "1:21:") {
		t.Fatalf("expected source-mapped error at 1:21, got %q from %q", formatted.Error(), result)
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
