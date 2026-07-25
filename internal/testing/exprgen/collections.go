package exprgen

import (
	"fmt"
	"strings"

	"pgregory.net/rapid"
)

type lambdaScope struct {
	named         []string
	implicitValue bool
	implicitIndex bool
}

func (s lambdaScope) withNamed(params ...string) lambdaScope {
	next := lambdaScope{
		named:         make([]string, 0, len(s.named)+len(params)),
		implicitValue: s.implicitValue,
		implicitIndex: s.implicitIndex,
	}
	next.named = append(next.named, s.named...)
	next.named = append(next.named, params...)
	return next
}

func (s lambdaScope) withImplicit() lambdaScope {
	s.implicitValue = true
	s.implicitIndex = true
	return s
}

func (s lambdaScope) hasNamed() bool {
	return len(s.named) > 0
}

func (s lambdaScope) hasReferences() bool {
	return len(s.references()) > 0
}

func (s lambdaScope) references() []string {
	refs := make([]string, 0, len(s.named)+2)
	refs = append(refs, s.named...)
	if s.implicitValue {
		refs = append(refs, "$")
	}
	if s.implicitIndex {
		refs = append(refs, "$$")
	}
	return refs
}

func scopeReferenceExpr(scope lambdaScope) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		refs := scope.references()
		if len(refs) == 0 {
			return Literal().Draw(t, "lit")
		}
		base := rapid.SampledFrom(refs).Draw(t, "scopeRef")
		if base != "$$" && rapid.Bool().Draw(t, "scopeDot") {
			field := rapid.SampledFrom(ContextShapeFields()).Draw(t, "scopeField")
			return base + "." + field
		}
		return base
	})
}

var lambdaParamCandidates = []string{
	"x", "y", "item", "value", "elem", "n",
	"idx", "i", "acc", "current", "left", "right",
}

// CollectionOpGen returns a generator for collection operator expressions
// (map, filter, reduce, flatMap) applied to array-like expressions.
func CollectionOpGen(depth int, features Feature) *rapid.Generator[string] {
	return collectionOpGenWithScope(depth, features, lambdaScope{}, exprConfig{})
}

// LambdaGen returns a generator for lambda expressions.
func LambdaGen(depth int, features Feature) *rapid.Generator[string] {
	return lambdaGenWithScope(depth, features, lambdaScope{}, exprConfig{})
}

func lambdaGenWithScope(depth int, features Feature, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	bodyDepth := depth - 1
	if bodyDepth < 0 {
		bodyDepth = 0
	}
	nestedFeatures := expressionNestedFeatures(features, cfg)

	return rapid.Custom(func(t *rapid.T) string {
		withIndex := rapid.Bool().Draw(t, "withIndex")
		paramCount := 1
		if withIndex {
			paramCount = 2
		}
		params := uniqueLambdaParams(t, paramCount, scope)
		lambdaScope := scope.withNamed(params...)
		body := expressionAtDepthWithScope(bodyDepth, nestedFeatures, lambdaScope, cfg).Draw(t, "lambdaBody")
		return fmt.Sprintf("(%s) -> %s", strings.Join(params, ", "), body)
	})
}

func collectionOpGenWithScope(depth int, features Feature, scope lambdaScope, cfg exprConfig) *rapid.Generator[string] {
	bodyDepth := depth - 1
	if bodyDepth < 0 {
		bodyDepth = 0
	}

	return rapid.Custom(func(t *rapid.T) string {
		input := collectionInputExpr(t, depth, features, scope, cfg)
		op := rapid.SampledFrom(cfg.collectionOps()).Draw(t, "collectionOp")
		switch op {
		case "map":
			return mapExpr(t, input, bodyDepth, features, scope, cfg)
		case "filter":
			return filterExpr(t, input, bodyDepth, features, scope, cfg)
		case "flatMap":
			return flatMapExpr(t, input, bodyDepth, features, scope, cfg)
		default:
			return reduceExpr(t, input, bodyDepth, features, scope, cfg)
		}
	})
}

func (c exprConfig) collectionOps() []string {
	return []string{"map", "filter", "flatMap", "reduce"}
}

func collectionInputExpr(t *rapid.T, depth int, features Feature, scope lambdaScope, cfg exprConfig) string {
	nestedFeatures := expressionNestedFeatures(features, cfg)
	if !cfg.DWCompat && depth > 1 && features&FeatureCollections != 0 && rapid.IntRange(0, 3).Draw(t, "nestedCollection") == 0 {
		return collectionOpGenWithScope(depth-1, nestedFeatures, scope, cfg).Draw(t, "nestedSource")
	}
	return filteredArrayExpr(max(depth-1, 0), nestedFeatures, scope, cfg).Draw(t, "arraySource")
}

func mapExpr(t *rapid.T, input string, bodyDepth int, features Feature, scope lambdaScope, cfg exprConfig) string {
	if features&FeatureImplicitLambda != 0 && rapid.Bool().Draw(t, "mapImplicit") {
		implicitScope := scope.withImplicit()
		body := scopeReferenceExpr(implicitScope).Draw(t, "mapImplicitBody")
		return fmt.Sprintf("%s map %s", input, body)
	}
	params := uniqueLambdaParams(t, 1+boolToInt(rapid.Bool().Draw(t, "mapWithIndex")), scope)
	lScope := scope.withNamed(params...)
	body := expressionAtDepthWithScope(bodyDepth, expressionNestedFeatures(features, cfg), lScope, cfg).Draw(t, "mapBody")
	return fmt.Sprintf("%s map (%s) -> %s", input, strings.Join(params, ", "), body)
}

func filterExpr(t *rapid.T, input string, bodyDepth int, features Feature, scope lambdaScope, cfg exprConfig) string {
	if features&FeatureImplicitLambda != 0 && rapid.Bool().Draw(t, "filterImplicit") {
		lScope := scope.withImplicit()
		body := filterPredicateExpr(t, bodyDepth, features, lScope, cfg)
		return fmt.Sprintf("%s filter %s", input, body)
	}
	params := uniqueLambdaParams(t, 1+boolToInt(rapid.Bool().Draw(t, "filterWithIndex")), scope)
	lScope := scope.withNamed(params...)
	body := filterPredicateExpr(t, bodyDepth, features, lScope, cfg)
	return fmt.Sprintf("%s filter (%s) -> %s", input, strings.Join(params, ", "), body)
}

func flatMapExpr(t *rapid.T, input string, bodyDepth int, features Feature, scope lambdaScope, cfg exprConfig) string {
	if features&FeatureImplicitLambda != 0 && rapid.Bool().Draw(t, "flatMapImplicit") {
		implicitScope := scope.withImplicit()
		bodyExpr := scopeReferenceExpr(implicitScope).Draw(t, "flatMapImplicitBody")
		return fmt.Sprintf("%s flatMap [%s]", input, bodyExpr)
	}
	params := uniqueLambdaParams(t, 1+boolToInt(rapid.Bool().Draw(t, "flatMapWithIndex")), scope)
	lScope := scope.withNamed(params...)
	bodyExpr := expressionAtDepthWithScope(bodyDepth, expressionNestedFeatures(features, cfg), lScope, cfg).Draw(t, "flatMapBody")
	return fmt.Sprintf("%s flatMap (%s) -> [%s]", input, strings.Join(params, ", "), bodyExpr)
}

func reduceExpr(t *rapid.T, input string, bodyDepth int, features Feature, scope lambdaScope, cfg exprConfig) string {
	withIndex := !cfg.DWCompat && rapid.Bool().Draw(t, "reduceWithIndex")
	paramCount := 2
	if withIndex {
		paramCount = 3
	}
	params := uniqueLambdaParams(t, paramCount, scope)
	lScope := scope.withNamed(params...)

	if cfg.DWCompat {
		// Returning the accumulator exercises the observable callback order
		// without introducing operator/type mismatches in differential cases.
		return fmt.Sprintf("%s reduce (%s) -> %s", input, strings.Join(params, ", "), params[1])
	}

	var body string
	if withIndex {
		body = fmt.Sprintf("%s + %s + %s", params[0], params[1], params[2])
	} else {
		body = fmt.Sprintf("%s + %s", params[0], params[1])
	}
	// Occasionally emit a generated body to increase structural variety.
	if rapid.IntRange(0, 4).Draw(t, "reduceBodyKind") == 0 {
		body = expressionAtDepthWithScope(bodyDepth, expressionNestedFeatures(features, cfg), lScope, cfg).Draw(t, "reduceBody")
	}
	return fmt.Sprintf("%s reduce (%s) -> %s", input, strings.Join(params, ", "), body)
}

func filterPredicateExpr(t *rapid.T, bodyDepth int, features Feature, scope lambdaScope, cfg exprConfig) string {
	if len(scope.references()) == 0 {
		return expressionAtDepthWithScope(bodyDepth, features, scope, cfg).Draw(t, "filterExpr")
	}
	base := scopeReferenceExpr(scope).Draw(t, "filterBase")
	switch rapid.IntRange(0, 2).Draw(t, "filterPredicateKind") {
	case 0:
		return fmt.Sprintf("%s != null", base)
	case 1:
		return fmt.Sprintf("%s == %s", base, base)
	default:
		return fmt.Sprintf("%s > -1", base)
	}
}

func uniqueLambdaParams(t *rapid.T, count int, scope lambdaScope) []string {
	params := make([]string, 0, count)
	used := make(map[string]struct{}, len(scope.named)+count)
	for _, existing := range scope.named {
		used[existing] = struct{}{}
	}
	for len(params) < count {
		candidate := rapid.SampledFrom(lambdaParamCandidates).Draw(t, fmt.Sprintf("lambdaParam%d", len(params)))
		if _, exists := used[candidate]; exists {
			candidate = fmt.Sprintf("%s%d", candidate, len(params))
		}
		if _, exists := used[candidate]; exists {
			continue
		}
		used[candidate] = struct{}{}
		params = append(params, candidate)
	}
	return params
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
