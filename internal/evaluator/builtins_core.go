package evaluator

import (
	"context"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"strconv"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/stringutils"
)

// callBuiltinLazyEval implements the lazy_eval(expr) function.
func callBuiltinLazyEval(e *ast.CallExpr, evalCtx map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 1 {
		return nil, newPosError("lazy_eval requires exactly 1 argument: expression", e.Pos())
	}

	return NewLazyValue(func(ctx context.Context) (Value, error) {
		return evalASTWithDepth(e.Args[0], evalCtx, depth)
	}, GetGoContext(evalCtx)), nil
}

// callBuiltinForceEval implements the force_eval(lazyValue) function.
func callBuiltinForceEval(values []Value, e *ast.CallExpr) (Value, error) {
	if len(values) != 1 {
		return nil, newPosError("force_eval requires exactly 1 argument: lazy value", e.Pos())
	}

	lazy, ok := values[0].(*LazyValue)
	if !ok {
		return nil, newPosError("force_eval argument must be a lazy value", e.Pos())
	}

	return ResolveValue(lazy)
}

// callBuiltinDefault implements the __default(value, defaultValue) function.
func callBuiltinDefault(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("default operator requires exactly 2 arguments: value, defaultValue", e.Pos())
	}
	// Evaluate the left side (the value)
	left, err := evalASTWithDepth(e.Args[0], context, depth)
	if err != nil {
		return nil, err
	}
	// If left is not nil, return it; otherwise evaluate and return right side
	if left != nil {
		return left, nil
	}
	return evalASTWithDepth(e.Args[1], context, depth)
}

// callBuiltinLambdaAST implements the __lambda(paramNames, body) function.
func callBuiltinLambdaAST(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("lambda expression requires exactly 2 arguments: paramNames, body", e.Pos())
	}

	// Evaluate the first argument (parameter names) - should be a string literal
	paramNamesVal, err := evalASTWithDepth(e.Args[0], context, depth)
	if err != nil {
		return nil, err
	}

	paramNamesStr, ok := paramNamesVal.(string)
	if !ok {
		return nil, newPosError("lambda parameter names must be a string", e.Pos())
	}

	// Parse parameter definitions (names, optional types, and defaults)
	var params []ParamDef
	if paramNamesStr != "" {
		for _, p := range splitRespectingDepth(paramNamesStr, ',') {
			p = strings.TrimSpace(p)
			param := ParamDef{ExpectedKind: KindUnknown}

			// Check for default value: "paramName = defaultValue"
			// Use depth-aware search for '=' to skip '=' inside braces/brackets
			if eqIdx := indexOfAtDepthZero(p, '='); eqIdx > 0 {
				param.Name = strings.TrimSpace(p[:eqIdx])
				defaultStr := strings.TrimSpace(p[eqIdx+1:])
				// Parse the default value
				defaultVal, err := parseDefaultValue(defaultStr)
				if err != nil {
					return nil, newPosError(fmt.Sprintf("invalid default value for parameter %s: %v", param.Name, err), e.Pos())
				}
				param.DefaultValue = defaultVal
				param.HasDefault = true
				param.ExpectedKind = KindOf(defaultVal)
			} else {
				param.Name = p
			}
			params = append(params, param)
		}
	}

	// Store the body as an unevaluated AST expression
	bodyExpr := e.Args[1]
	bodyStr := exprToString(bodyExpr)

	return &Lambda{
		Params:  params,
		Body:    bodyStr,
		BodyAST: bodyExpr, // Store the AST for later evaluation
	}, nil
}

// parseDefaultValue parses a string representation of a default value.
func parseDefaultValue(s string) (interface{}, error) {
	s = strings.TrimSpace(s)

	// Check for quoted string
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") && len(s) >= 2 {
		return s[1 : len(s)-1], nil
	}

	// Check for boolean
	if b, ok := ParseBoolLiteral(s); ok {
		return b, nil
	}

	// Check for null
	if s == "null" || s == "nil" {
		return nil, nil
	}

	// Check for number
	if num, ok := ParseNumericLiteral(s); ok {
		return num, nil
	}

	// Check for empty array
	if s == "[]" {
		return []interface{}{}, nil
	}

	// Check for empty object
	if s == "{}" {
		return map[string]interface{}{}, nil
	}

	// Object literal: {key: value, ...}
	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		return parseDefaultObjectLiteral(s)
	}

	// Array literal: [value, ...]
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		return parseDefaultArrayLiteral(s)
	}

	// Go-syntax object literal: map[string]interface{}{"key": value, ...}
	const goObjPrefix = "map[string]interface{}{"
	if strings.HasPrefix(s, goObjPrefix) && strings.HasSuffix(s, "}") {
		return parseGoSyntaxObjectLiteral(s, goObjPrefix)
	}

	// Go-syntax array literal: []interface{}{value, ...}
	const goArrPrefix = "[]interface{}{"
	if strings.HasPrefix(s, goArrPrefix) && strings.HasSuffix(s, "}") {
		return parseGoSyntaxArrayLiteral(s, goArrPrefix)
	}

	return nil, unifiederrors.EvalErrorf("cannot parse value: %s", s)
}

// parseDefaultObjectLiteral parses an object literal like {total: 0, count: 0}.
func parseDefaultObjectLiteral(s string) (map[string]interface{}, error) {
	inner := strings.TrimSpace(s[1 : len(s)-1])
	if inner == "" {
		return map[string]interface{}{}, nil
	}

	result := make(map[string]interface{})
	pairs := splitRespectingDepth(inner, ',')

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		colonIdx := strings.Index(pair, ":")
		if colonIdx < 0 {
			return nil, unifiederrors.EvalErrorf("invalid object literal: missing ':' in pair: %s", pair)
		}
		key := strings.TrimSpace(pair[:colonIdx])
		valueStr := strings.TrimSpace(pair[colonIdx+1:])

		val, err := parseDefaultValue(valueStr)
		if err != nil {
			return nil, err
		}
		result[key] = val
	}

	return result, nil
}

// parseDefaultArrayLiteral parses an array literal like [1, 2, 3].
func parseDefaultArrayLiteral(s string) ([]interface{}, error) {
	inner := strings.TrimSpace(s[1 : len(s)-1])
	if inner == "" {
		return []interface{}{}, nil
	}

	var result []interface{}
	items := splitRespectingDepth(inner, ',')

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		val, err := parseDefaultValue(item)
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}

	return result, nil
}

// parseGoSyntaxObjectLiteral parses a Go-syntax object literal like
// map[string]interface{}{"sum": 0, "count": 1,}.
func parseGoSyntaxObjectLiteral(s string, prefix string) (map[string]interface{}, error) {
	inner := strings.TrimSpace(s[len(prefix) : len(s)-1])
	inner = strings.TrimSuffix(strings.TrimSpace(inner), ",")
	inner = strings.TrimSpace(inner)
	if inner == "" {
		return map[string]interface{}{}, nil
	}

	result := make(map[string]interface{})
	pairs := splitRespectingDepth(inner, ',')

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		colonIdx := strings.Index(pair, ":")
		if colonIdx < 0 {
			return nil, unifiederrors.EvalErrorf("invalid Go object literal: missing ':' in pair: %s", pair)
		}
		keyStr := strings.TrimSpace(pair[:colonIdx])
		valStr := strings.TrimSpace(pair[colonIdx+1:])

		// Unquote key if quoted
		if len(keyStr) >= 2 && strings.HasPrefix(keyStr, "\"") && strings.HasSuffix(keyStr, "\"") {
			keyStr = keyStr[1 : len(keyStr)-1]
		}

		val, err := parseDefaultValue(valStr)
		if err != nil {
			return nil, err
		}
		result[keyStr] = val
	}

	return result, nil
}

// parseGoSyntaxArrayLiteral parses a Go-syntax array literal like []interface{}{1, 2, 3,}.
func parseGoSyntaxArrayLiteral(s string, prefix string) ([]interface{}, error) {
	inner := strings.TrimSpace(s[len(prefix) : len(s)-1])
	inner = strings.TrimSuffix(strings.TrimSpace(inner), ",")
	inner = strings.TrimSpace(inner)
	if inner == "" {
		return []interface{}{}, nil
	}

	var result []interface{}
	items := splitRespectingDepth(inner, ',')

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		val, err := parseDefaultValue(item)
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}

	return result, nil
}

// splitRespectingDepth splits a string by a separator character, respecting
// nested braces, brackets, parentheses, and string literals.
func splitRespectingDepth(s string, sep rune) []string {
	var parts []string
	var current strings.Builder
	var sc stringutils.ScanState

	for i := 0; i < len(s); i++ {
		ch := s[i]
		sc.Advance(ch)
		if sc.AtTopLevel() && rune(ch) == sep {
			parts = append(parts, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(ch)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// indexOfAtDepthZero finds the first occurrence of a character at depth 0
// (not inside braces, brackets, or parentheses), respecting string literals.
// Returns -1 if not found.
func indexOfAtDepthZero(s string, ch byte) int {
	var sc stringutils.ScanState

	for i := 0; i < len(s); i++ {
		c := s[i]
		sc.Advance(c)
		if sc.AtTopLevel() && c == ch {
			return i
		}
	}
	return -1
}

// exprToString converts an AST expression back to a string representation.
func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Value
	case *ast.Ident:
		return e.Name
	case *ast.BinaryExpr:
		left := exprToString(e.X)
		right := exprToString(e.Y)
		return fmt.Sprintf("%s %s %s", left, e.Op, right)
	case *ast.ParenExpr:
		return fmt.Sprintf("(%s)", exprToString(e.X))
	case *ast.CallExpr:
		fun := exprToString(e.Fun)
		var args []string
		for _, arg := range e.Args {
			args = append(args, exprToString(arg))
		}
		return fmt.Sprintf("%s(%s)", fun, strings.Join(args, ", "))
	case *ast.IndexExpr:
		x := exprToString(e.X)
		index := exprToString(e.Index)
		return fmt.Sprintf("%s[%s]", x, index)
	case *ast.CompositeLit:
		return fmt.Sprintf("{...}")
	default:
		return fmt.Sprintf("%v", expr)
	}
}

// callBuiltinModCall implements __modcall("Module", "func", args...) for module function calls.
func callBuiltinModCall(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) < 2 {
		return nil, newPosError("__modcall requires at least 2 arguments: module name and function name", e.Pos())
	}

	modNameVal, err := evalASTWithDepth(e.Args[0], context, depth+1)
	if err != nil {
		return nil, err
	}
	funcNameVal, err := evalASTWithDepth(e.Args[1], context, depth+1)
	if err != nil {
		return nil, err
	}

	modName, ok1 := modNameVal.(string)
	funcName, ok2 := funcNameVal.(string)
	if !ok1 || !ok2 {
		return nil, newPosError("__modcall requires string module and function names", e.Pos())
	}

	modVal, ok := context[modName]
	if !ok {
		return nil, newPosError(fmt.Sprintf("module %q not found", modName), e.Pos())
	}

	ns, ok := modVal.(map[string]interface{})
	if !ok {
		return nil, newPosError(fmt.Sprintf("%q is not a module namespace", modName), e.Pos())
	}

	fnVal, ok := ns[funcName]
	if !ok {
		return nil, newPosError(fmt.Sprintf("module %q has no function %q", modName, funcName), e.Pos())
	}

	l, ok := fnVal.(*Lambda)
	if !ok {
		return nil, newPosError(fmt.Sprintf("%s::%s is not a function", modName, funcName), e.Pos())
	}

	args := make([]interface{}, 0, len(e.Args)-2)
	for _, argExpr := range e.Args[2:] {
		v, err := evalASTWithDepth(argExpr, context, depth+1)
		if err != nil {
			return nil, err
		}
		args = append(args, v)
	}

	return callUserLambda(l, args, context, depth+1)
}

// callUserLambda invokes a user-defined Lambda with the given arguments.
func callUserLambda(l *Lambda, args []interface{}, callingContext map[string]interface{}, depth int) (interface{}, error) {
	base := make(map[string]interface{})

	if l.Env != nil {
		for k, v := range l.Env {
			base[k] = v
		}
	}

	for k, v := range callingContext {
		base[k] = v
	}

	for i, param := range l.Params {
		if i < len(args) {
			base[param.Name] = args[i]
		}
	}

	return evalASTWithDepth(l.BodyAST, base, depth)
}

func callBuiltinCoerce(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) < 2 || len(e.Args) > 3 {
		return nil, newPosError("as operator requires 2 or 3 arguments: value, type[, config]", e.Pos())
	}

	value, err := evalASTWithDepth(e.Args[0], context, depth)
	if err != nil {
		return nil, err
	}

	var typeName string

	switch t := e.Args[1].(type) {
	case *ast.Ident:
		typeName = t.Name

	case *ast.BasicLit:
		if t.Kind != token.STRING {
			return nil, newPosError("as operator type argument must be a string literal or identifier", t.Pos())
		}
		unquoted, err := strconv.Unquote(t.Value)
		if err != nil {
			return nil, newPosError(fmt.Sprintf("invalid type string literal: %v", err), t.Pos())
		}
		typeName = unquoted

	default:
		v, err := evalASTWithDepth(e.Args[1], context, depth)
		if err != nil {
			return nil, err
		}
		s, ok := v.(string)
		if !ok {
			return nil, newPosError("as operator type argument must evaluate to a string", e.Args[1].Pos())
		}
		typeName = s
	}

	var properties Object
	if len(e.Args) == 3 {
		propsVal, err := evalASTWithDepth(e.Args[2], context, depth)
		if err != nil {
			return nil, err
		}

		// If it's nil, treat as empty properties
		if propsVal == nil {
			properties = nil
		} else if p, ok := propsVal.(map[string]interface{}); ok {
			properties = p
		} else if p, ok := propsVal.(Object); ok {
			properties = p
		} else {
			return nil, newPosError(fmt.Sprintf("as operator config argument must be an object, got %T", propsVal), e.Args[2].Pos())
		}
	}

	return coerceToTypeWithVisited(value, typeName, context, e.Args[1].Pos(), nil, properties)
}

func callBuiltinCase(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("__case requires exactly 2 arguments: value, cases", e.Pos())
	}

	val, err := evalASTWithDepth(e.Args[0], context, depth)
	if err != nil {
		return nil, err
	}

	casesAST, ok := e.Args[1].(*ast.CompositeLit)
	if !ok {
		return nil, newPosError("__case expects an array literal for its second argument", e.Args[1].Pos())
	}

	for _, elt := range casesAST.Elts {
		// Each element is map[string]interface{}{ "pattern": "...", "result": ... }
		caseMapAST, ok := elt.(*ast.CompositeLit)
		if !ok {
			continue
		}

		var patternStr string
		var resultAST ast.Expr

		for _, kvElt := range caseMapAST.Elts {
			kv, ok := kvElt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			keyIdent, ok := kv.Key.(*ast.BasicLit)
			if !ok || keyIdent.Kind != token.STRING {
				continue
			}
			key, err := strconv.Unquote(keyIdent.Value)
			if err != nil {
				continue // skip malformed key
			}

			if key == "pattern" {
				if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					unquoted, err := strconv.Unquote(lit.Value)
					if err != nil {
						return nil, newPosError(fmt.Sprintf("invalid case pattern string literal: %v", err), lit.Pos())
					}
					patternStr = unquoted
				}
			} else if key == "result" {
				resultAST = kv.Value
			}
		}

		if patternStr == "" || resultAST == nil {
			continue
		}

		// Check if pattern matches
		matchContext := copyContext(context)
		matched, matchFound, err := matchPattern(val, patternStr, matchContext, e.Pos())
		if err != nil {
			return nil, err
		}

		if matchFound {
			return evalASTWithDepth(resultAST, matched, depth+1)
		}
	}

	return nil, newPosError("no case matched the value", e.Pos())
}

// matchLiteral attempts to match a value against a literal pattern.
func matchLiteral(val interface{}, pattern string) bool {
	if strings.Contains(pattern, " is ") || pattern == "else" {
		return false
	}
	litAST, err := goparser.ParseExpr(pattern)
	if err != nil {
		return false
	}
	litVal, err := evalAST(litAST, nil)
	if err != nil {
		return false
	}
	return litVal == val
}

// matchTypePattern matches a type pattern like "s is String" or "is String".
// Returns the updated context and whether it matched.
func matchTypePattern(val interface{}, pattern string, context map[string]interface{}) (map[string]interface{}, bool) {
	isIdx := strings.Index(pattern, " is ")
	startsIs := strings.HasPrefix(pattern, "is ")
	if isIdx == -1 && !startsIs {
		return nil, false
	}

	var varName, typeExpr string
	if startsIs {
		typeExpr = strings.TrimSpace(pattern[3:])
	} else {
		parts := strings.SplitN(pattern, " is ", 2)
		varName = strings.TrimSpace(parts[0])
		typeExpr = strings.TrimSpace(parts[1])
	}

	if !matchesTypeExactly(val, typeExpr, context) && getTypeName(val) != typeExpr {
		return nil, false
	}

	if varName == "" {
		return context, true
	}
	matchContext := copyContext(context)
	matchContext[varName] = val
	return matchContext, true
}

// evaluateIfCondition evaluates the guard condition in a pattern match.
func evaluateIfCondition(condStr string, matchContext map[string]interface{}, pos token.Pos) (bool, error) {
	condAST, err := goparser.ParseExpr(condStr)
	if err != nil {
		return false, newPosError(fmt.Sprintf("invalid case condition: %v", err), pos)
	}
	condVal, err := evalASTWithDepth(condAST, matchContext, 0)
	if err != nil {
		return false, err
	}
	condBool, ok := condVal.(bool)
	if !ok {
		return false, newPosError(fmt.Sprintf("case condition must be boolean, got %T", condVal), pos)
	}
	return condBool, nil
}

func matchPattern(val interface{}, patternStr string, context map[string]interface{}, pos token.Pos) (map[string]interface{}, bool, error) {
	if patternStr == "else" {
		return context, true, nil
	}

	p := strings.TrimSpace(patternStr)
	ifParts := strings.SplitN(p, " if ", 2)
	p = strings.TrimSpace(ifParts[0])

	var matchContext map[string]interface{}
	matched := false

	// Try literal match first
	if matchLiteral(val, p) {
		matched = true
		matchContext = context
	}

	// Try type match if literal didn't match
	if !matched {
		if ctx, ok := matchTypePattern(val, p, context); ok {
			matched = true
			matchContext = ctx
		}
	}

	// Try simple binding pattern (just an identifier like "x")
	// This allows patterns like "case x if x > 10" to work
	if !matched && isSimpleIdentifier(p) {
		matched = true
		matchContext = copyContext(context)
		matchContext[p] = val
	}

	if !matched {
		return nil, false, nil
	}

	// Evaluate 'if' condition if present
	if len(ifParts) > 1 {
		condOk, err := evaluateIfCondition(strings.TrimSpace(ifParts[1]), matchContext, pos)
		if err != nil {
			return nil, false, err
		}
		if !condOk {
			return nil, false, nil
		}
	}

	return matchContext, true, nil
}

// isSimpleIdentifier checks if a pattern is a simple identifier (variable binding).
// Simple identifiers start with a letter or underscore and contain only letters, digits, or underscores.
func isSimpleIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Check first character is letter or underscore
	first := rune(s[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	// Check rest are letters, digits, or underscores
	for _, ch := range s[1:] {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
			return false
		}
	}
	// Exclude keywords that shouldn't be treated as binding patterns
	switch s {
	case "true", "false", "nil", "null":
		return false
	}
	return true
}

func callBuiltinOnNull(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("onNull requires exactly 2 arguments: value, lambda", e.Pos())
	}

	// Evaluate the value argument
	value, err := evalASTWithDepth(e.Args[0], context, depth)
	if err != nil {
		return nil, err
	}

	// If value is not null, return it
	if value != nil {
		return value, nil
	}

	// Evaluate the lambda argument to get the lambda function
	lambdaVal, err := evalASTWithDepth(e.Args[1], context, depth)
	if err != nil {
		return nil, err
	}

	lambda, ok := lambdaVal.(*Lambda)
	if !ok {
		return nil, newPosError(fmt.Sprintf("onNull expects a lambda function, got %T", lambdaVal), e.Pos())
	}

	// The lambda should have 0 parameters
	if lambda.ParamCount() != 0 {
		return nil, newPosError(fmt.Sprintf("onNull lambda must have 0 parameters, got %d", lambda.ParamCount()), e.Pos())
	}

	// Execute the lambda with a clean context
	lambdaContext := copyContext(context)
	return evalASTWithDepth(lambda.BodyAST, lambdaContext, depth+1)
}

func callBuiltinThen(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("then requires exactly 2 arguments: value, lambda", e.Pos())
	}

	// Evaluate the value argument
	value, err := evalASTWithDepth(e.Args[0], context, depth)
	if err != nil {
		return nil, err
	}

	// If value is null, return null
	if value == nil {
		return nil, nil
	}

	// Evaluate the lambda argument to get the lambda function
	lambdaVal, err := evalASTWithDepth(e.Args[1], context, depth)
	if err != nil {
		return nil, err
	}

	lambda, ok := lambdaVal.(*Lambda)
	if !ok {
		return nil, newPosError(fmt.Sprintf("then expects a lambda function, got %T", lambdaVal), e.Pos())
	}

	// The lambda should have 1 parameter
	if lambda.ParamCount() != 1 {
		return nil, newPosError(fmt.Sprintf("then lambda must have exactly 1 parameter, got %d", lambda.ParamCount()), e.Pos())
	}

	// Execute the lambda with the value as parameter
	lambdaContext := copyContext(context)
	lambdaContext[lambda.ParamName(0)] = value
	return evalASTWithDepth(lambda.BodyAST, lambdaContext, depth+1)
}
