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
	FeatureArithmetic  Feature = 1 << iota // +, -, *, /
	FeatureComparison                      // ==, !=, <, >
	FeatureLogical                         // &&, ||
	FeatureArrays                          // array literals
	FeatureObjects                         // object literals
	FeatureDotAccess                       // expr.field
	FeatureIndexAccess                     // expr[index]
	FeatureRangeIndex                      // expr[start to end]
	FeatureUnary                           // !, - (prefix)
	FeatureParens                          // parenthesized expressions
	FeatureAll         Feature = FeatureArithmetic | FeatureComparison | FeatureLogical | FeatureArrays | FeatureObjects | FeatureDotAccess | FeatureIndexAccess | FeatureRangeIndex | FeatureUnary | FeatureParens
)

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
	return expressionAtDepth(maxDepth, features)
}

func expressionAtDepth(depth int, features Feature) *rapid.Generator[string] {
	if depth <= 0 {
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

	// Build a list of enabled form generators with weights.
	type form struct {
		weight int
		name   string
		gen    func(depth int) *rapid.Generator[string]
	}
	forms := []form{{weight: 4, name: "lit", gen: func(_ int) *rapid.Generator[string] { return Literal() }}}
	if hasBinary {
		captured := ops // capture for closure
		forms = append(forms, form{weight: 2, name: "binary", gen: func(d int) *rapid.Generator[string] {
			return filteredBinaryExpr(d, features, captured)
		}})
	}
	if hasUnary {
		forms = append(forms, form{weight: 1, name: "unary", gen: func(d int) *rapid.Generator[string] {
			return filteredUnaryExpr(d, features)
		}})
	}
	if hasParens {
		forms = append(forms, form{weight: 1, name: "paren", gen: func(d int) *rapid.Generator[string] {
			return filteredParenExpr(d, features)
		}})
	}
	if hasArrays {
		forms = append(forms, form{weight: 2, name: "array", gen: func(d int) *rapid.Generator[string] {
			return filteredArrayExpr(d, features)
		}})
	}
	if hasObjects {
		forms = append(forms, form{weight: 2, name: "object", gen: func(d int) *rapid.Generator[string] {
			return filteredObjectExpr(d, features)
		}})
	}
	if hasDotAccess {
		forms = append(forms, form{weight: 2, name: "dot", gen: func(d int) *rapid.Generator[string] {
			return filteredDotAccessExpr(d, features)
		}})
	}
	if hasIndexAccess {
		forms = append(forms, form{weight: 2, name: "index", gen: func(d int) *rapid.Generator[string] {
			return filteredIndexAccessExpr(d, features)
		}})
	}
	if hasRangeIndex {
		forms = append(forms, form{weight: 1, name: "rangeIndex", gen: func(d int) *rapid.Generator[string] {
			return filteredRangeIndexExpr(d, features)
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

func filteredBinaryExpr(depth int, features Feature, ops []string) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		left := expressionAtDepth(depth-1, features).Draw(t, "left")
		op := rapid.SampledFrom(ops).Draw(t, "op")
		right := expressionAtDepth(depth-1, features).Draw(t, "right")
		return fmt.Sprintf("%s %s %s", left, op, right)
	})
}

func filteredUnaryExpr(depth int, features Feature) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		op := UnaryOp().Draw(t, "op")
		operand := expressionAtDepth(depth-1, features).Draw(t, "operand")
		return fmt.Sprintf("%s(%s)", op, operand)
	})
}

func filteredParenExpr(depth int, features Feature) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		inner := expressionAtDepth(depth-1, features).Draw(t, "inner")
		return fmt.Sprintf("(%s)", inner)
	})
}

func filteredArrayExpr(depth int, features Feature) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(0, 5).Draw(t, "len")
		if n == 0 {
			return "[]"
		}
		elems := make([]string, n)
		for i := range elems {
			elems[i] = expressionAtDepth(depth-1, features).Draw(t, fmt.Sprintf("elem%d", i))
		}
		return "[" + strings.Join(elems, ", ") + "]"
	})
}

func filteredObjectExpr(depth int, features Feature) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		n := rapid.IntRange(0, 5).Draw(t, "len")
		if n == 0 {
			return "{}"
		}
		fields := make([]string, n)
		used := map[string]struct{}{}
		for i := range fields {
			key := generateIdentifierKey(t, used, i)
			value := expressionAtDepth(depth-1, features).Draw(t, fmt.Sprintf("value%d", i))
			fields[i] = fmt.Sprintf("%s: %s", strconv.Quote(key), value)
		}
		return "{" + strings.Join(fields, ", ") + "}"
	})
}

func filteredDotAccessExpr(depth int, features Feature) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		base := payloadBaseExpr(t)
		field := rapid.SampledFrom(ContextShapeFields()).Draw(t, "field")
		return fmt.Sprintf("%s.%s", base, field)
	})
}

func filteredIndexAccessExpr(depth int, features Feature) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		base := payloadBaseExpr(t)
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

func filteredRangeIndexExpr(depth int, features Feature) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		base := payloadBaseExpr(t)
		start := rapid.IntRange(0, 3).Draw(t, "start")
		end := rapid.IntRange(start, start+3).Draw(t, "end")
		return fmt.Sprintf("%s[%d to %d]", base, start, end)
	})
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
