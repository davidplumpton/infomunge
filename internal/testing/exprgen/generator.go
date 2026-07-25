package exprgen

import (
	"fmt"
	"strconv"
	"strings"

	"infomunge/internal/preprocessor"

	"pgregory.net/rapid"
)

// Feature is a bitmask controlling which expression forms are enabled.
type Feature uint

const (
	FeatureArithmetic          Feature = 1 << iota // +, -, *, /, **, %, ++
	FeatureComparison                              // ==, !=, <, >, <=, >=, ~=
	FeatureLogical                                 // &&, ||
	FeatureArrays                                  // array literals
	FeatureObjects                                 // object literals
	FeatureDotAccess                               // expr.field
	FeatureIndexAccess                             // expr[index]
	FeatureRangeIndex                              // expr[start to end]
	FeatureUnary                                   // !, - (prefix)
	FeatureParens                                  // parenthesized expressions
	FeatureLambdas                                 // (x) -> expr
	FeatureCollections                             // arr map/filter/reduce/flatMap ...
	FeatureImplicitLambda                          // $ and $$ in collection bodies
	FeatureConditionals                            // if (cond) expr else expr
	FeatureDefault                                 // expr default expr
	FeatureCaseExpressions                         // expr case { pattern -> expr, ... }
	FeatureStringInterpolation                     // "text $(expr) text"
	FeatureBuiltinCalls                            // sizeOf(...), contains(...), ...
	FeatureAll                 Feature = FeatureArithmetic | FeatureComparison | FeatureLogical | FeatureArrays | FeatureObjects | FeatureDotAccess | FeatureIndexAccess | FeatureRangeIndex | FeatureUnary | FeatureParens | FeatureLambdas | FeatureCollections | FeatureImplicitLambda | FeatureConditionals | FeatureDefault | FeatureCaseExpressions | FeatureStringInterpolation | FeatureBuiltinCalls

	// FeatureDWCompat is the subset of features compatible with DataWeave CLI.
	// Excludes: FeatureRangeIndex, FeatureImplicitLambda, FeatureStringInterpolation.
	FeatureDWCompat Feature = FeatureArithmetic | FeatureComparison | FeatureLogical |
		FeatureArrays | FeatureObjects | FeatureDotAccess | FeatureIndexAccess |
		FeatureUnary | FeatureParens | FeatureLambdas | FeatureCollections |
		FeatureConditionals | FeatureDefault | FeatureBuiltinCalls
)

type builtinSpec struct {
	name  string
	arity int
}

var builtinSpecs = []builtinSpec{
	{name: "sizeOf", arity: 1},
	{name: "typeOf", arity: 1},
	{name: "upper", arity: 1},
	{name: "lower", arity: 1},
	{name: "flatten", arity: 1},
	{name: "keys", arity: 1},
	{name: "values", arity: 1},
	{name: "trim", arity: 1},
	{name: "isEmpty", arity: 1},
	{name: "contains", arity: 2},
	{name: "startsWith", arity: 2},
	{name: "endsWith", arity: 2},
}

type weightedOperator struct {
	op     string
	weight int
}

var operatorFeatureOrder = []Feature{
	FeatureArithmetic,
	FeatureComparison,
	FeatureLogical,
}

// weightedOpsByFeature maps feature flags to weighted binary operators.
// Common operators get larger weights so they appear more often.
var weightedOpsByFeature = map[Feature][]weightedOperator{
	FeatureArithmetic: {
		{op: "+", weight: 8},
		{op: "-", weight: 7},
		{op: "*", weight: 6},
		{op: "/", weight: 5},
		{op: "%", weight: 3},
		{op: "++", weight: 3},
		{op: "**", weight: 1},
	},
	FeatureComparison: {
		{op: "==", weight: 6},
		{op: "!=", weight: 5},
		{op: "<", weight: 4},
		{op: ">", weight: 4},
		{op: "<=", weight: 4},
		{op: ">=", weight: 4},
		{op: "~=", weight: 1},
	},
	FeatureLogical: {
		{op: "&&", weight: 5},
		{op: "||", weight: 5},
	},
}

func expandWeightedOperators(weightedOps []weightedOperator) []string {
	total := 0
	for _, op := range weightedOps {
		if op.weight > 0 {
			total += op.weight
		}
	}
	expanded := make([]string, 0, total)
	for _, op := range weightedOps {
		for i := 0; i < op.weight; i++ {
			expanded = append(expanded, op.op)
		}
	}
	return expanded
}

func allWeightedBinaryOps() []string {
	var all []string
	for _, feature := range operatorFeatureOrder {
		all = append(all, expandWeightedOperators(weightedOpsByFeature[feature])...)
	}
	return all
}

// dwIncompatibleOps lists binary operators that have different syntax or
// semantics in DataWeave and should be excluded from DW-compatible generation.
var dwIncompatibleOps = map[string]bool{
	"+":  true, // mixed string/number coercion differs (bd-qv20)
	"++": true, // concatenation semantics may differ
	"**": true, // DW uses pow() function
	"~=": true, // infomunge-specific regex match
}

// dwOperatorReplacements keeps equivalent operators in differential
// generation when DataWeave spells them differently.
var dwOperatorReplacements = map[string]string{
	"%": "mod",
}

// exprConfig controls expression generation behavior. The zero value produces
// standard infomunge expressions. Set DWCompat to true for the DataWeave-
// compatible subset.
type exprConfig struct {
	DWCompat bool
}

// literalGen returns the appropriate literal generator for the config.
func (c exprConfig) literalGen() *rapid.Generator[string] {
	if c.DWCompat {
		return DWLiteral()
	}
	return Literal()
}

// filterOps applies DW-incompatible operator exclusion when in DW-compat mode.
func (c exprConfig) filterOps(ops []string) []string {
	if !c.DWCompat {
		return ops
	}
	var result []string
	for _, op := range ops {
		if replacement, ok := dwOperatorReplacements[op]; ok {
			result = append(result, replacement)
			continue
		}
		if !dwIncompatibleOps[op] {
			result = append(result, op)
		}
	}
	return result
}

// filteredOps returns weighted binary operators enabled by the given feature set.
func filteredOps(features Feature) []string {
	var ops []string
	for _, feature := range operatorFeatureOrder {
		if features&feature == 0 {
			continue
		}
		ops = append(ops, expandWeightedOperators(weightedOpsByFeature[feature])...)
	}
	return ops
}

// Expression returns a generator of syntactically valid infomunge expression
// strings with a maximum recursion depth. The features parameter controls
// which expression forms are enabled; start with FeatureAll or a subset like
// FeatureArithmetic|FeatureComparison|FeatureArrays.
func Expression(maxDepth int, features Feature) *rapid.Generator[string] {
	return expressionAtDepthWithScope(maxDepth, features, lambdaScope{}, exprConfig{})
}

// DWCompatExpression returns a generator of expressions compatible with both
// infomunge and the DataWeave CLI. It uses the DW-compatible feature set and
// excludes operators that have different syntax in DataWeave.
func DWCompatExpression(maxDepth int) *rapid.Generator[string] {
	return expressionAtDepthWithScope(maxDepth, FeatureDWCompat, lambdaScope{}, exprConfig{DWCompat: true})
}

// WrapDWScript wraps an expression in a script header accepted by both
// infomunge and DataWeave CLI.
func WrapDWScript(expr string) string {
	return "%dw 2.0\noutput application/json\n---\n" + expr
}

func expressionAtDepthWithScope(depth int, features Feature, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	if depth <= 0 {
		if scope.hasReferences() {
			return rapid.OneOf(
				cfg.literalGen(),
				scopeReferenceExpr(scope),
			)
		}
		return cfg.literalGen()
	}
	ops := cfg.filterOps(filteredOps(features))
	hasBinary := len(ops) > 0
	hasUnary := features&FeatureUnary != 0
	hasParens := features&FeatureParens != 0
	hasArrays := features&FeatureArrays != 0
	hasObjects := features&FeatureObjects != 0
	hasDotAccess := features&FeatureDotAccess != 0
	hasIndexAccess := features&FeatureIndexAccess != 0
	hasRangeIndex := features&FeatureRangeIndex != 0
	hasLambdas := features&FeatureLambdas != 0
	hasCollections := features&FeatureCollections != 0
	hasConditionals := features&FeatureConditionals != 0
	hasDefault := features&FeatureDefault != 0
	hasCaseExpressions := features&FeatureCaseExpressions != 0
	hasStringInterpolation := features&FeatureStringInterpolation != 0
	hasBuiltinCalls := features&FeatureBuiltinCalls != 0

	// Build a list of enabled form generators with weights.
	type form struct {
		weight int
		name   string
		gen    func(depth int) *rapid.Generator[string]
	}
	forms := []form{{weight: 4, name: "lit", gen: func(_ int) *rapid.Generator[string] { return cfg.literalGen() }}}
	if scope.hasReferences() {
		forms = append(forms, form{weight: 2, name: "scopeRef", gen: func(_ int) *rapid.Generator[string] {
			return scopeReferenceExpr(scope)
		}})
	}
	if hasBinary {
		captured := ops // capture for closure
		forms = append(forms, form{weight: 2, name: "binary", gen: func(d int) *rapid.Generator[string] {
			return filteredBinaryExpr(d, features, captured, scope, cfg)
		}})
	}
	if hasUnary {
		forms = append(forms, form{weight: 1, name: "unary", gen: func(d int) *rapid.Generator[string] {
			return filteredUnaryExpr(d, features, scope, cfg)
		}})
	}
	if hasParens {
		forms = append(forms, form{weight: 1, name: "paren", gen: func(d int) *rapid.Generator[string] {
			return filteredParenExpr(d, features, scope, cfg)
		}})
	}
	if hasArrays {
		forms = append(forms, form{weight: 2, name: "array", gen: func(d int) *rapid.Generator[string] {
			return filteredArrayExpr(d, features, scope, cfg)
		}})
	}
	if hasObjects {
		forms = append(forms, form{weight: 2, name: "object", gen: func(d int) *rapid.Generator[string] {
			return filteredObjectExpr(d, features, scope, cfg)
		}})
	}
	if hasDotAccess {
		forms = append(forms, form{weight: 2, name: "dot", gen: func(d int) *rapid.Generator[string] {
			return filteredDotAccessExpr(d, features, scope)
		}})
	}
	if hasIndexAccess {
		forms = append(forms, form{weight: 2, name: "index", gen: func(d int) *rapid.Generator[string] {
			return filteredIndexAccessExpr(d, features, scope, cfg)
		}})
	}
	if hasRangeIndex {
		forms = append(forms, form{weight: 1, name: "rangeIndex", gen: func(d int) *rapid.Generator[string] {
			return filteredRangeIndexExpr(d, features, scope)
		}})
	}
	if hasLambdas {
		forms = append(forms, form{weight: 1, name: "lambda", gen: func(d int) *rapid.Generator[string] {
			return lambdaGenWithScope(d, features, scope, cfg)
		}})
	}
	if hasCollections {
		forms = append(forms, form{weight: 2, name: "collection", gen: func(d int) *rapid.Generator[string] {
			return collectionOpGenWithScope(d, features, scope, cfg)
		}})
	}
	if hasConditionals {
		forms = append(forms, form{weight: 1, name: "conditional", gen: func(d int) *rapid.Generator[string] {
			return filteredConditionalExpr(d, features, scope, cfg)
		}})
	}
	if hasDefault {
		forms = append(forms, form{weight: 1, name: "defaultOp", gen: func(d int) *rapid.Generator[string] {
			return filteredDefaultExpr(d, features, scope, cfg)
		}})
	}
	if hasCaseExpressions {
		forms = append(forms, form{weight: 1, name: "caseExpr", gen: func(d int) *rapid.Generator[string] {
			return filteredCaseExpr(d, features, scope, cfg)
		}})
	}
	if hasStringInterpolation {
		forms = append(forms, form{weight: 1, name: "stringInterpolation", gen: func(d int) *rapid.Generator[string] {
			return filteredStringInterpolationExpr(d, features, scope, cfg)
		}})
	}
	if hasBuiltinCalls {
		forms = append(forms, form{weight: 2, name: "builtinCall", gen: func(d int) *rapid.Generator[string] {
			return filteredBuiltinCallExpr(d, features, scope, cfg)
		}})
	}

	totalWeight := 0
	for _, f := range forms {
		totalWeight += f.weight
	}

	return rapid.Custom(func(t *rapid.T) string {
		pick := rapid.IntRange(0, totalWeight-1).Draw(t, "exprPick")
		cumulative := 0
		for _, f := range forms {
			cumulative += f.weight
			if pick < cumulative {
				return f.gen(depth).Draw(t, f.name)
			}
		}
		// Fallback (should not happen).
		return cfg.literalGen().Draw(t, "lit")
	})
}

func filteredBinaryExpr(depth int, features Feature, ops []string, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features, cfg)
	return rapid.Custom(func(t *rapid.T) string {
		left := expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, "left")
		op := rapid.SampledFrom(ops).Draw(t, "op")
		right := expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, "right")
		return fmt.Sprintf("%s %s %s", left, op, right)
	})
}

func filteredUnaryExpr(depth int, features Feature, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features, cfg)
	return rapid.Custom(func(t *rapid.T) string {
		op := UnaryOp().Draw(t, "op")
		operand := expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, "operand")
		return fmt.Sprintf("%s(%s)", op, operand)
	})
}

func filteredParenExpr(depth int, features Feature, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features, cfg)
	return rapid.Custom(func(t *rapid.T) string {
		inner := expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, "inner")
		return fmt.Sprintf("(%s)", inner)
	})
}

func filteredArrayExpr(depth int, features Feature, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features, cfg)
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(0, 5).Draw(t, "len")
		if n == 0 {
			return "[]"
		}
		elems := make([]string, n)
		for i := range elems {
			elems[i] = expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, fmt.Sprintf("elem%d", i))
		}
		return "[" + strings.Join(elems, ", ") + "]"
	})
}

func filteredObjectExpr(depth int, features Feature, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features, cfg)
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(0, 5).Draw(t, "len")
		if n == 0 {
			return "{}"
		}
		fields := make([]string, n)
		used := map[string]struct{}{}
		for i := range fields {
			key := generateIdentifierKey(t, used, i)
			value := expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, fmt.Sprintf("value%d", i))
			fields[i] = fmt.Sprintf("%s: %s", strconv.Quote(key), value)
		}
		return "{" + strings.Join(fields, ", ") + "}"
	})
}

func filteredDotAccessExpr(depth int, features Feature, scope lambdaScope) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		base := payloadOrScopeBaseExpr(t, scope)
		field := rapid.SampledFrom(ContextShapeFields()).Draw(t, "field")
		return fmt.Sprintf("%s.%s", base, field)
	})
}

func filteredIndexAccessExpr(depth int, features Feature, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		if cfg.DWCompat {
			base := payloadBaseExpr(t)
			field := rapid.SampledFrom(ContextShapeFields()).Draw(t, "idxField")
			return fmt.Sprintf("%s[%s]", base, strconv.Quote(field))
		}

		base := payloadOrScopeBaseExpr(t, scope)
		index := rapid.OneOf(
			rapid.Custom(func(t *rapid.T) string {
				return strconv.Itoa(rapid.IntRange(0, 5).Draw(t, "idxInt"))
			}),
			rapid.Custom(func(t *rapid.T) string {
				field := rapid.SampledFrom(ContextShapeFields()).Draw(t, "idxField")
				return strconv.Quote(field)
			}),
		).Draw(t, "index")
		return fmt.Sprintf("%s[%s]", base, index)
	})
}

func filteredRangeIndexExpr(depth int, features Feature, scope lambdaScope) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		base := payloadOrScopeBaseExpr(t, scope)
		start := rapid.IntRange(0, 3).Draw(t, "start")
		end := rapid.IntRange(start, start+3).Draw(t, "end")
		return fmt.Sprintf("%s[%d to %d]", base, start, end)
	})
}

func filteredConditionalExpr(depth int, features Feature, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features, cfg)
	return rapid.Custom(func(t *rapid.T) string {
		cond := expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, "cond")
		thenExpr := expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, "then")
		elseExpr := expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, "else")
		return fmt.Sprintf("if (%s) %s else %s", cond, thenExpr, elseExpr)
	})
}

func filteredDefaultExpr(depth int, features Feature, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features, cfg)
	return rapid.Custom(func(t *rapid.T) string {
		left := expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, "left")
		right := expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, "right")
		return fmt.Sprintf("%s default %s", left, right)
	})
}

func filteredCaseExpr(depth int, features Feature, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features, cfg)
	return rapid.Custom(func(t *rapid.T) string {
		target := expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, "caseTarget")
		caseCount := rapid.IntRange(1, 3).Draw(t, "caseCount")
		clauses := make([]string, 0, caseCount+1)
		for i := 0; i < caseCount; i++ {
			pattern, caseScope := generateCasePattern(t, i, scope, cfg)
			resultExpr := expressionAtDepthWithScope(depth-1, nestedFeatures, caseScope, cfg).Draw(t, fmt.Sprintf("caseResult%d", i))
			clauses = append(clauses, fmt.Sprintf("%s -> %s", pattern, resultExpr))
		}
		elseExpr := expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, "caseElse")
		clauses = append(clauses, fmt.Sprintf("else -> %s", elseExpr))
		return fmt.Sprintf("%s case { %s }", target, strings.Join(clauses, ", "))
	})
}

func generateCasePattern(t *rapid.T, idx int, scope lambdaScope, cfg exprConfig) (string, lambdaScope) {
	switch rapid.IntRange(0, 4).Draw(t, fmt.Sprintf("casePatternKind%d", idx)) {
	case 0:
		return cfg.literalGen().Draw(t, fmt.Sprintf("caseLiteral%d", idx)), scope
	case 1:
		return "is " + caseTypeName(t, idx), scope
	case 2:
		varName := caseVarName(t, idx)
		typeName := caseTypeName(t, idx)
		if rapid.Bool().Draw(t, fmt.Sprintf("caseGuard%d", idx)) {
			return fmt.Sprintf("%s is %s if %s != null", varName, typeName, varName), scope.withNamed(varName)
		}
		return fmt.Sprintf("%s is %s", varName, typeName), scope.withNamed(varName)
	case 3:
		varName := caseVarName(t, idx)
		return varName, scope.withNamed(varName)
	default:
		if rapid.Bool().Draw(t, fmt.Sprintf("caseKeyword%d", idx)) {
			return "case " + cfg.literalGen().Draw(t, fmt.Sprintf("caseKeywordLiteral%d", idx)), scope
		}
		return "is " + caseTypeName(t, idx), scope
	}
}

func caseTypeName(t *rapid.T, idx int) string {
	return rapid.SampledFrom([]string{"String", "Number", "Boolean", "Array", "Object", "Null"}).Draw(t, fmt.Sprintf("caseType%d", idx))
}

func caseVarName(t *rapid.T, idx int) string {
	return rapid.SampledFrom([]string{"v", "n", "s", "item", "value"}).Draw(t, fmt.Sprintf("caseVar%d", idx))
}

func filteredStringInterpolationExpr(depth int, features Feature, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features, cfg)
	return rapid.Custom(func(t *rapid.T) string {
		prefix := interpolationTextPiece(t, "prefix")
		mid := interpolationTextPiece(t, "mid")
		suffix := interpolationTextPiece(t, "suffix")
		expr := expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, "interpolatedExpr")
		return strconv.Quote(prefix + "$(" + expr + ")" + mid + "$(" + cfg.literalGen().Draw(t, "secondInterpolatedExpr") + ")" + suffix)
	})
}

func filteredBuiltinCallExpr(depth int, features Feature, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features, cfg)
	return rapid.Custom(func(t *rapid.T) string {
		spec := rapid.SampledFrom(builtinSpecs).Draw(t, "builtin")
		args := make([]string, spec.arity)
		for i := range args {
			args[i] = expressionAtDepthWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, fmt.Sprintf("arg%d", i))
		}
		return fmt.Sprintf("%s(%s)", spec.name, strings.Join(args, ", "))
	})
}

func interpolationTextPiece(t *rapid.T, label string) string {
	piece := rapid.StringMatching(`[A-Za-z0-9 ]{0,6}`).Draw(t, label)
	return strings.ReplaceAll(piece, `"`, "")
}

func expressionNestedFeatures(features Feature, cfg exprConfig) Feature {
	nested := features &^ FeatureConditionals
	if cfg.DWCompat {
		nested &^= FeatureCollections
	}
	return nested
}

func payloadOrScopeBaseExpr(t *rapid.T, scope lambdaScope) string {
	if scope.hasNamed() && rapid.Bool().Draw(t, "useScopeBase") {
		return rapid.SampledFrom(scope.named).Draw(t, "scopeBase")
	}
	return payloadBaseExpr(t)
}

func payloadBaseExpr(t *rapid.T) string {
	if rapid.Bool().Draw(t, "usePayloadField") {
		field := rapid.SampledFrom(ContextShapeFields()).Draw(t, "baseField")
		return "payload." + field
	}
	return "payload"
}

func generateIdentifierKey(t *rapid.T, used map[string]struct{}, fallbackIdx int) string {
	for attempt := 0; attempt < 5; attempt++ {
		key := rapid.StringMatching(`[A-Za-z_][A-Za-z0-9_]{0,11}`).Draw(t, fmt.Sprintf("key%d", attempt))
		if _, exists := used[key]; exists {
			continue
		}
		used[key] = struct{}{}
		return key
	}
	fallback := fmt.Sprintf("field%d", fallbackIdx)
	if _, exists := used[fallback]; !exists {
		used[fallback] = struct{}{}
		return fallback
	}
	for i := 0; ; i++ {
		candidate := fmt.Sprintf("field%d_%d", fallbackIdx, i)
		if _, exists := used[candidate]; exists {
			continue
		}
		used[candidate] = struct{}{}
		return candidate
	}
}

// WrapScript wraps an expression in a valid infomunge script with header.
func WrapScript(expr string) string {
	return "%im 0.1\noutput application/json\n---\n" + expr
}

// IsValid returns true if the expression passes preprocessing without error.
// Use this as a lightweight validity guard to skip generated expressions that
// would fail with a syntax error.
func IsValid(expr string) bool {
	_, _, err := preprocessor.PrepareForParsing(expr, preprocessor.Options{})
	return err == nil
}
