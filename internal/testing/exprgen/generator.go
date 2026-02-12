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
	FeatureArithmetic          Feature = 1 << iota // +, -, *, /
	FeatureComparison                              // ==, !=, <, >
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
	FeatureStringInterpolation                     // "text $(expr) text"
	FeatureBuiltinCalls                            // sizeOf(...), contains(...), ...
	FeatureAll                 Feature = FeatureArithmetic | FeatureComparison | FeatureLogical | FeatureArrays | FeatureObjects | FeatureDotAccess | FeatureIndexAccess | FeatureRangeIndex | FeatureUnary | FeatureParens | FeatureLambdas | FeatureCollections | FeatureImplicitLambda | FeatureConditionals | FeatureDefault | FeatureStringInterpolation | FeatureBuiltinCalls
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

// opsByFeature maps feature flags to the binary operators they enable.
var opsByFeature = map[Feature][]string{
	FeatureArithmetic: {"+", "-", "*", "/"},
	FeatureComparison: {"==", "!=", "<", ">"},
	FeatureLogical:    {"&&", "||"},
}

// filteredOps returns the binary operators enabled by the given feature set.
func filteredOps(features Feature) []string {
	var ops []string
	for feat, featureOps := range opsByFeature {
		if features&feat != 0 {
			ops = append(ops, featureOps...)
		}
	}
	return ops
}

// Expression returns a generator of syntactically valid infomunge expression
// strings with a maximum recursion depth. The features parameter controls
// which expression forms are enabled; start with FeatureAll or a subset like
// FeatureArithmetic|FeatureComparison|FeatureArrays.
func Expression(maxDepth int, features Feature) *rapid.Generator[string] {
	return expressionAtDepthWithScope(maxDepth, features, lambdaScope{})
}

func expressionAtDepth(depth int, features Feature) *rapid.Generator[string] {
	return expressionAtDepthWithScope(depth, features, lambdaScope{})
}

func expressionAtDepthWithScope(depth int, features Feature, scope lambdaScope) *rapid.Generator[string] {
	if depth <= 0 {
		if scope.hasReferences() {
			return rapid.OneOf(
				Literal(),
				scopeReferenceExpr(scope),
			)
		}
		return Literal()
	}
	ops := filteredOps(features)
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
	hasStringInterpolation := features&FeatureStringInterpolation != 0
	hasBuiltinCalls := features&FeatureBuiltinCalls != 0

	// Build a list of enabled form generators with weights.
	type form struct {
		weight int
		name   string
		gen    func(depth int) *rapid.Generator[string]
	}
	forms := []form{{weight: 4, name: "lit", gen: func(_ int) *rapid.Generator[string] { return Literal() }}}
	if scope.hasReferences() {
		forms = append(forms, form{weight: 2, name: "scopeRef", gen: func(_ int) *rapid.Generator[string] {
			return scopeReferenceExpr(scope)
		}})
	}
	if hasBinary {
		captured := ops // capture for closure
		forms = append(forms, form{weight: 2, name: "binary", gen: func(d int) *rapid.Generator[string] {
			return filteredBinaryExpr(d, features, captured, scope)
		}})
	}
	if hasUnary {
		forms = append(forms, form{weight: 1, name: "unary", gen: func(d int) *rapid.Generator[string] {
			return filteredUnaryExpr(d, features, scope)
		}})
	}
	if hasParens {
		forms = append(forms, form{weight: 1, name: "paren", gen: func(d int) *rapid.Generator[string] {
			return filteredParenExpr(d, features, scope)
		}})
	}
	if hasArrays {
		forms = append(forms, form{weight: 2, name: "array", gen: func(d int) *rapid.Generator[string] {
			return filteredArrayExpr(d, features, scope)
		}})
	}
	if hasObjects {
		forms = append(forms, form{weight: 2, name: "object", gen: func(d int) *rapid.Generator[string] {
			return filteredObjectExpr(d, features, scope)
		}})
	}
	if hasDotAccess {
		forms = append(forms, form{weight: 2, name: "dot", gen: func(d int) *rapid.Generator[string] {
			return filteredDotAccessExpr(d, features, scope)
		}})
	}
	if hasIndexAccess {
		forms = append(forms, form{weight: 2, name: "index", gen: func(d int) *rapid.Generator[string] {
			return filteredIndexAccessExpr(d, features, scope)
		}})
	}
	if hasRangeIndex {
		forms = append(forms, form{weight: 1, name: "rangeIndex", gen: func(d int) *rapid.Generator[string] {
			return filteredRangeIndexExpr(d, features, scope)
		}})
	}
	if hasLambdas {
		forms = append(forms, form{weight: 1, name: "lambda", gen: func(d int) *rapid.Generator[string] {
			return lambdaGenWithScope(d, features, scope)
		}})
	}
	if hasCollections {
		forms = append(forms, form{weight: 2, name: "collection", gen: func(d int) *rapid.Generator[string] {
			return collectionOpGenWithScope(d, features, scope)
		}})
	}
	if hasConditionals {
		forms = append(forms, form{weight: 1, name: "conditional", gen: func(d int) *rapid.Generator[string] {
			return filteredConditionalExpr(d, features, scope)
		}})
	}
	if hasDefault {
		forms = append(forms, form{weight: 1, name: "defaultOp", gen: func(d int) *rapid.Generator[string] {
			return filteredDefaultExpr(d, features, scope)
		}})
	}
	if hasStringInterpolation {
		forms = append(forms, form{weight: 1, name: "stringInterpolation", gen: func(d int) *rapid.Generator[string] {
			return filteredStringInterpolationExpr(d, features, scope)
		}})
	}
	if hasBuiltinCalls {
		forms = append(forms, form{weight: 2, name: "builtinCall", gen: func(d int) *rapid.Generator[string] {
			return filteredBuiltinCallExpr(d, features, scope)
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
		return Literal().Draw(t, "lit")
	})
}

func filteredBinaryExpr(depth int, features Feature, ops []string, scope lambdaScope) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features)
	return rapid.Custom(func(t *rapid.T) string {
		left := expressionAtDepthWithScope(depth-1, nestedFeatures, scope).Draw(t, "left")
		op := rapid.SampledFrom(ops).Draw(t, "op")
		right := expressionAtDepthWithScope(depth-1, nestedFeatures, scope).Draw(t, "right")
		return fmt.Sprintf("%s %s %s", left, op, right)
	})
}

func filteredUnaryExpr(depth int, features Feature, scope lambdaScope) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features)
	return rapid.Custom(func(t *rapid.T) string {
		op := UnaryOp().Draw(t, "op")
		operand := expressionAtDepthWithScope(depth-1, nestedFeatures, scope).Draw(t, "operand")
		return fmt.Sprintf("%s(%s)", op, operand)
	})
}

func filteredParenExpr(depth int, features Feature, scope lambdaScope) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features)
	return rapid.Custom(func(t *rapid.T) string {
		inner := expressionAtDepthWithScope(depth-1, nestedFeatures, scope).Draw(t, "inner")
		return fmt.Sprintf("(%s)", inner)
	})
}

func filteredArrayExpr(depth int, features Feature, scope lambdaScope) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features)
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(0, 5).Draw(t, "len")
		if n == 0 {
			return "[]"
		}
		elems := make([]string, n)
		for i := range elems {
			elems[i] = expressionAtDepthWithScope(depth-1, nestedFeatures, scope).Draw(t, fmt.Sprintf("elem%d", i))
		}
		return "[" + strings.Join(elems, ", ") + "]"
	})
}

func filteredObjectExpr(depth int, features Feature, scope lambdaScope) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features)
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(0, 5).Draw(t, "len")
		if n == 0 {
			return "{}"
		}
		fields := make([]string, n)
		used := map[string]struct{}{}
		for i := range fields {
			key := generateIdentifierKey(t, used, i)
			value := expressionAtDepthWithScope(depth-1, nestedFeatures, scope).Draw(t, fmt.Sprintf("value%d", i))
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

func filteredIndexAccessExpr(depth int, features Feature, scope lambdaScope) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
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

func filteredConditionalExpr(depth int, features Feature, scope lambdaScope) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features)
	return rapid.Custom(func(t *rapid.T) string {
		cond := expressionAtDepthWithScope(depth-1, nestedFeatures, scope).Draw(t, "cond")
		thenExpr := expressionAtDepthWithScope(depth-1, nestedFeatures, scope).Draw(t, "then")
		elseExpr := expressionAtDepthWithScope(depth-1, nestedFeatures, scope).Draw(t, "else")
		return fmt.Sprintf("if (%s) %s else %s", cond, thenExpr, elseExpr)
	})
}

func filteredDefaultExpr(depth int, features Feature, scope lambdaScope) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features)
	return rapid.Custom(func(t *rapid.T) string {
		left := expressionAtDepthWithScope(depth-1, nestedFeatures, scope).Draw(t, "left")
		right := expressionAtDepthWithScope(depth-1, nestedFeatures, scope).Draw(t, "right")
		return fmt.Sprintf("%s default %s", left, right)
	})
}

func filteredStringInterpolationExpr(depth int, features Feature, scope lambdaScope) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features)
	return rapid.Custom(func(t *rapid.T) string {
		prefix := interpolationTextPiece(t, "prefix")
		mid := interpolationTextPiece(t, "mid")
		suffix := interpolationTextPiece(t, "suffix")
		expr := expressionAtDepthWithScope(depth-1, nestedFeatures, scope).Draw(t, "interpolatedExpr")
		return strconv.Quote(prefix + "$(" + expr + ")" + mid + "$(" + Literal().Draw(t, "secondInterpolatedExpr") + ")" + suffix)
	})
}

func filteredBuiltinCallExpr(depth int, features Feature, scope lambdaScope) *rapid.Generator[string] {
	nestedFeatures := expressionNestedFeatures(features)
	return rapid.Custom(func(t *rapid.T) string {
		spec := rapid.SampledFrom(builtinSpecs).Draw(t, "builtin")
		args := make([]string, spec.arity)
		for i := range args {
			args[i] = expressionAtDepthWithScope(depth-1, nestedFeatures, scope).Draw(t, fmt.Sprintf("arg%d", i))
		}
		return fmt.Sprintf("%s(%s)", spec.name, strings.Join(args, ", "))
	})
}

func interpolationTextPiece(t *rapid.T, label string) string {
	piece := rapid.StringMatching(`[A-Za-z0-9 ]{0,6}`).Draw(t, label)
	return strings.ReplaceAll(piece, `"`, "")
}

func expressionNestedFeatures(features Feature) Feature {
	return features &^ FeatureConditionals
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
